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
package transform

import (
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// TestAttachModuleInterfacesKeepsOwnership verifies module-owned interfaces
// land on their module's specs with the Nautobot UUID as the spec ID, while
// device-owned interfaces are left to the device grouping.
//
// Why it matters: the device grouping skips every interface with a module
// reference, so this pass is the only owner of module ports; two modules with
// the same port name must each keep their own port.
// Inputs: two modules each owning an interface named "p1", both installed in
// one device that Nautobot 3.2 names on their ports, plus a device interface.
// Outputs: each module has exactly one "p1" spec carrying its own Nautobot
// interface ID; the device interface is attached to neither.
// Data choice: identical port names and a shared device make the module
// reference the only thing that can route each interface correctly.
func TestAttachModuleInterfacesKeepsOwnership(t *testing.T) {
	modANB, modBNB, portANB, portBNB := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	modACani, modBCani := uuid.New(), uuid.New()
	modules := map[uuid.UUID]*devicetypes.CaniModuleType{
		modACani: {ID: modACani, Name: "nic-a", ObjectMeta: devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": modANB}}},
		modBCani: {ID: modBCani, Name: "nic-b", ObjectMeta: devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": modBNB}}},
	}
	deviceNB := uuid.New()
	portA := nautobotapi.Interface{Id: oaPtr(portANB), Name: "p1"}
	setNBRef(&portA.Device, deviceNB)
	setNBRef(&portA.Module, modANB)
	portB := nautobotapi.Interface{Id: oaPtr(portBNB), Name: "p1"}
	setNBRef(&portB.Device, deviceNB)
	setNBRef(&portB.Module, modBNB)
	devicePort := nautobotapi.Interface{Id: oaPtr(uuid.New()), Name: "eth0"}
	setNBRef(&devicePort.Device, deviceNB)

	warnings := AttachModuleInterfaces(modules, []nautobotapi.Interface{portA, devicePort, portB})

	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none when every module is present", warnings)
	}
	for cani, want := range map[uuid.UUID]uuid.UUID{modACani: portANB, modBCani: portBNB} {
		specs := modules[cani].Interfaces
		if len(specs) != 1 || specs[0].Name != "p1" || specs[0].ID != want {
			t.Errorf("module %s specs = %+v, want one p1 with ID %s", modules[cani].Name, specs, want)
		}
	}
}

// TestAttachModuleInterfacesReportsUnimportedModule verifies an interface
// whose owning module is not among the imported modules is left out and
// reported, while the imported module still receives its own port.
//
// Why it matters: dropping the interface silently would make a Nautobot
// module port vanish from the inventory with no trace; the other records the
// transform cannot place are reported through the same warning list.
// Inputs: one imported module with port "p1", and a port "p2" referencing a
// module ID that was not imported. Outputs: one warning naming p2 and the
// missing module ID; the imported module has exactly one spec.
// Data choice: both ports carry module references so only the index miss
// explains the different outcome.
func TestAttachModuleInterfacesReportsUnimportedModule(t *testing.T) {
	modNB, missingNB, modCani := uuid.New(), uuid.New(), uuid.New()
	modules := map[uuid.UUID]*devicetypes.CaniModuleType{
		modCani: {ID: modCani, Name: "nic-a", ObjectMeta: devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": modNB}}},
	}
	owned := nautobotapi.Interface{Id: oaPtr(uuid.New()), Name: "p1"}
	setNBRef(&owned.Module, modNB)
	orphan := nautobotapi.Interface{Id: oaPtr(uuid.New()), Name: "p2"}
	setNBRef(&orphan.Module, missingNB)

	warnings := AttachModuleInterfaces(modules, []nautobotapi.Interface{owned, orphan})

	if len(warnings) != 1 || !strings.Contains(warnings[0], "interface p2") || !strings.Contains(warnings[0], missingNB.String()) {
		t.Errorf("warnings = %v, want one naming p2 and module %s", warnings, missingNB)
	}
	if specs := modules[modCani].Interfaces; len(specs) != 1 || specs[0].Name != "p1" {
		t.Errorf("nic-a specs = %+v, want only p1", specs)
	}
}

// TestGroupInterfacesByDeviceLeavesModulePortsToTheirModule verifies only
// device-owned interfaces are grouped onto a device.
//
// Why it matters: Nautobot 3.2 sets the root device on a port of an installed
// module; grouping that port onto the device as well as attaching it to its
// module would import it twice and give the device a port its module owns.
// Inputs: a device-owned port, an installed-module port named like it that
// carries both device and module references, and a spare-module port carrying
// only a module reference. Outputs: the device's group holds only its own port.
// Data choice: these are the three interface shapes Nautobot 3.2 returns.
func TestGroupInterfacesByDeviceLeavesModulePortsToTheirModule(t *testing.T) {
	devNB, devPortNB := uuid.New(), uuid.New()
	devicePort := nautobotapi.Interface{Id: oaPtr(devPortNB), Name: "HSN 0"}
	setNBRef(&devicePort.Device, devNB)
	installedPort := nautobotapi.Interface{Id: oaPtr(uuid.New()), Name: "HSN 0"}
	setNBRef(&installedPort.Device, devNB)
	setNBRef(&installedPort.Module, uuid.New())
	sparePort := nautobotapi.Interface{Id: oaPtr(uuid.New()), Name: "p1"}
	setNBRef(&sparePort.Module, uuid.New())

	groups := GroupInterfacesByDevice([]nautobotapi.Interface{devicePort, installedPort, sparePort})

	if got := groups[devNB]; len(groups) != 1 || len(got) != 1 || directUUID(got[0].Id) != devPortNB {
		t.Fatalf("groups = %+v, want only port %s under device %s", groups, devPortNB, devNB)
	}
}
