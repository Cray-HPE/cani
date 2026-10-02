/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package export

import (
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// modulePortInventory returns a device "node-01" with its own port "1" and a
// module in it whose port is also named "1", rebuilt so the index records the
// module as the port's owner, plus the module and its port's ID.
func modulePortInventory(t *testing.T) (*devicetypes.Inventory, *devicetypes.CaniDeviceType, *devicetypes.CaniModuleType, uuid.UUID) {
	t.Helper()
	inv := devicetypes.NewInventory()
	device := &devicetypes.CaniDeviceType{ID: uuid.New(), Name: "node-01", Interfaces: []devicetypes.InterfaceSpec{{Name: "1"}}}
	module := &devicetypes.CaniModuleType{ID: uuid.New(), Name: "nic-1", ParentDevice: device.ID,
		Interfaces: []devicetypes.InterfaceSpec{{Name: "1"}}}
	inv.Devices[device.ID], inv.Modules[module.ID] = device, module
	if err := inv.RebuildDerivedState().Err(); err != nil {
		t.Fatalf("fixture rebuild: %v", err)
	}
	return inv, device, module, module.Interfaces[0].ID
}

// TestResolveEndInterfaceRefusesUnconfirmedModule verifies a cable end on a
// module-owned port is refused when this run did not create or find the
// module, without looking the port up among the device's own ports, and
// resolves on the module once the run has confirmed it.
//
// Why it matters: the device lookup is fuzzy enough to match a module port
// "1" to the parent's "mgmt1" or "eth1", so a module whose export failed
// would have its cable created on the wrong interface and counted as created
// while only the module error was reported.
// Inputs: an end owned by a module with no confirmed ID against a server
// whose device port list would match; then the module recorded as found.
// Outputs: an error naming the module and the port with zero device-filtered
// interface calls; then the module's port, still with zero device calls.
// Data choice: the device port shares the module port's name so a fallback
// would succeed silently rather than miss.
func TestResolveEndInterfaceRefusesUnconfirmedModule(t *testing.T) {
	var deviceCalls int
	modulePortID := uuid.New()
	e, cleanup := newExporterWithServer(t, ownerListServer(&deviceCalls, "",
		[]string{ifaceResultJSON(modulePortID, "1")}, []string{ifaceResultJSON(uuid.New(), "1")}))
	defer cleanup()
	module := &devicetypes.CaniModuleType{ID: uuid.New(), Name: "nic-1"}
	end := &cableEnd{device: &devicetypes.CaniDeviceType{Name: "node-01"}, module: module, port: "1", deviceID: uuid.New()}

	err := e.resolveEndInterface(end)
	if err == nil || !strings.Contains(err.Error(), "nic-1") || !strings.Contains(err.Error(), "port 1 on node-01") {
		t.Fatalf("unconfirmed module: error = %v, want a refusal naming the module and port", err)
	}
	if deviceCalls != 0 || end.iface != nil {
		t.Fatalf("unconfirmed module: device lookups = %d, iface = %v; want no device lookup and no interface", deviceCalls, end.iface)
	}

	nautobotModuleID := uuid.New()
	e.Cache.RecordModule(module.ID, nautobotModuleID)
	end.moduleID = nautobotModuleID
	if err := e.resolveEndInterface(end); err != nil || end.iface == nil || end.iface.ID != modulePortID {
		t.Fatalf("confirmed module: iface = %v, err = %v; want the module's port %v", end.iface, err, modulePortID)
	}
	if deviceCalls != 0 {
		t.Errorf("confirmed module: device lookups = %d, want none", deviceCalls)
	}
}

// TestResolveIPAssignmentRefusesUnconfirmedModule verifies an address on a
// module-owned port is refused when this run did not confirm the module, is
// resolved on the module once it has, and under a dry run is previewed for a
// module the run only planned.
//
// Why it matters: the IP phase shares the cable phase's lookup, so the same
// fuzzy device fallback would attach the address to the parent's same-named
// port when the module's export failed.
// Inputs: the module-port inventory, whose device port and module port are
// both named "1", and a device map holding the device; the module first
// unconfirmed, then recorded, then planned under a dry run.
// Outputs: an error with no device-filtered lookup; then a target that is
// the module's port; then an assignment with no target and no error.
// Data choice: the identical port names make a silent fallback observable
// only through the device-call count and the resolved ID.
func TestResolveIPAssignmentRefusesUnconfirmedModule(t *testing.T) {
	var deviceCalls int
	modulePortID := uuid.New()
	e, cleanup := newExporterWithServer(t, ownerListServer(&deviceCalls, "",
		[]string{ifaceResultJSON(modulePortID, "1")}, []string{ifaceResultJSON(uuid.New(), "1")}))
	defer cleanup()
	inv, device, module, portID := modulePortInventory(t)
	deviceMap := map[string]uuid.UUID{device.Name: uuid.New()}

	assignment, err := e.resolveIPAssignment(inv, portID, "10.0.0.1/24", deviceMap)
	if err == nil || !strings.Contains(err.Error(), "nic-1") || assignment != nil || deviceCalls != 0 {
		t.Fatalf("unconfirmed module: assignment = %v, err = %v, device lookups = %d; want a refusal naming the module and no device lookup",
			assignment, err, deviceCalls)
	}

	e.Cache.RecordModule(module.ID, uuid.New())
	assignment, err = e.resolveIPAssignment(inv, portID, "10.0.0.1/24", deviceMap)
	if err != nil || assignment == nil || assignment.target == nil || assignment.target.ID != modulePortID || deviceCalls != 0 {
		t.Fatalf("confirmed module: assignment = %+v, err = %v, device lookups = %d; want the module's port %v", assignment, err, deviceCalls, modulePortID)
	}

	e.Cache = NewLookupCache(e.Client)
	e.Cache.PlanModule(module.ID)
	e.Options.DryRun = true
	assignment, err = e.resolveIPAssignment(inv, portID, "10.0.0.1/24", deviceMap)
	if err != nil || assignment == nil || assignment.target != nil {
		t.Fatalf("planned module: assignment = %+v, err = %v; want a preview assignment with no target", assignment, err)
	}
}
