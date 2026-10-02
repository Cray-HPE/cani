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
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// ownerListServer answers /dcim/interfaces/ with modulePorts when the query
// filters by module and devicePorts when it filters by device, answers
// /dcim/devices/ with device, and counts the device-filtered interface calls.
func ownerListServer(deviceCalls *int, device string, modulePorts, devicePorts []string) http.HandlerFunc {
	list := func(entries []string) string {
		return fmt.Sprintf(`{"count":%d,"results":[%s]}`, len(entries), strings.Join(entries, ","))
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		query := r.URL.Query()
		switch {
		case strings.Contains(r.URL.Path, "dcim/interfaces") && query.Get("module") != "":
			_, _ = io.WriteString(w, list(modulePorts))
		case strings.Contains(r.URL.Path, "dcim/interfaces"):
			*deviceCalls++
			_, _ = io.WriteString(w, list(devicePorts))
		case strings.Contains(r.URL.Path, "dcim/devices"):
			_, _ = io.WriteString(w, list([]string{device}))
		default:
			_, _ = io.WriteString(w, `{}`)
		}
	}
}

// cabledIfaceJSON builds an interface list entry with a nested cable.
func cabledIfaceJSON(id uuid.UUID, name string, cableID uuid.UUID) string {
	return fmt.Sprintf(`{"id":%q,"name":%q,"display":%q,"cable":{"id":%q}}`, id, name, name, cableID)
}

// TestGetInterfaceByModuleAndName_RecordsAttachedCable verifies a module
// port found through the module lookup carries the cable attached to it.
//
// Why it matters: module ports are found only through their module, and the
// cable phase skips a cable whose port is already cabled by reading CableID;
// without it a second export would try to cable the port again.
// Inputs: a module whose one port "Port 1" carries a cable. Outputs: the
// cached item has that port's ID and the cable's ID.
// Data choice: the nested cable object is the shape Nautobot 3.2 returns.
func TestGetInterfaceByModuleAndName_RecordsAttachedCable(t *testing.T) {
	portID, cableID := uuid.New(), uuid.New()
	var deviceCalls int
	e, cleanup := newExporterWithServer(t, ownerListServer(&deviceCalls, "",
		[]string{cabledIfaceJSON(portID, "Port 1", cableID)}, nil))
	defer cleanup()

	item, err := e.Cache.GetInterfaceByModuleAndName(uuid.New(), "Port 1")
	if err != nil {
		t.Fatalf("GetInterfaceByModuleAndName() error = %v", err)
	}
	if item == nil || item.ID != portID || item.CableID != cableID {
		t.Fatalf("item = %+v, want ID %s with CableID %s", item, portID, cableID)
	}
}

// TestGetInterfaceByModuleAndNameFuzzy_MatchesWithinModule verifies a
// spelling difference is absorbed among the module's own ports.
//
// Why it matters: module ports no longer fall back to the device lookup, so
// the module lookup must absorb the naming differences the device lookup did.
// Inputs: a module port "Port 1" and a lookup for "port1". Outputs: that port.
// Data choice: "port1" misses the exact cache key but normalizes to the same
// name, which exercises only the fuzzy step.
func TestGetInterfaceByModuleAndNameFuzzy_MatchesWithinModule(t *testing.T) {
	portID := uuid.New()
	var deviceCalls int
	e, cleanup := newExporterWithServer(t, ownerListServer(&deviceCalls, "",
		[]string{ifaceResultJSON(portID, "Port 1")}, nil))
	defer cleanup()

	item, err := e.Cache.GetInterfaceByModuleAndNameFuzzy(uuid.New(), "port1")
	if err != nil {
		t.Fatalf("GetInterfaceByModuleAndNameFuzzy() error = %v", err)
	}
	if item == nil || item.ID != portID {
		t.Fatalf("item = %+v, want port %s", item, portID)
	}
}

// TestResolveOwnedInterface_ModulePortIgnoresDevicePort verifies a module
// port missing from its module is not matched against the parent device.
//
// Why it matters: a device port and a module port may share a name, so a
// device fallback would attach a cable or address to the wrong port.
// Inputs: a module with no ports and a device whose own port "HSN 0" shares
// the requested name. Outputs: no interface and no device lookup.
// Data choice: the same-named device port is the one a fallback would find.
func TestResolveOwnedInterface_ModulePortIgnoresDevicePort(t *testing.T) {
	var deviceCalls int
	e, cleanup := newExporterWithServer(t, ownerListServer(&deviceCalls, "",
		nil, []string{ifaceResultJSON(uuid.New(), "HSN 0")}))
	defer cleanup()

	iface, err := e.resolveOwnedInterface(uuid.New(), uuid.New(), "HSN 0")
	if err != nil {
		t.Fatalf("resolveOwnedInterface() error = %v", err)
	}
	if iface != nil || deviceCalls != 0 {
		t.Fatalf("iface = %+v after %d device lookups, want nil after none", iface, deviceCalls)
	}
}

// TestInterfaceCabled_ChecksModulePortOnItsModule verifies the dry-run cable
// preview finds a cable already on a module port by looking at the module.
//
// Why it matters: the preview must report the conflicts the real run reports,
// and device lookups no longer list module ports.
// Inputs: an existing device whose module port "HSN 0" is cabled and whose
// own "HSN 0" is not. Outputs: the module end is reported as cabled and the
// device end is not.
// Data choice: one shared port name proves the owner, not the name, decides.
func TestInterfaceCabled_ChecksModulePortOnItsModule(t *testing.T) {
	deviceJSON := fmt.Sprintf(`{"id":%q,"name":"node-a","display":"node-a"}`, uuid.New())
	var deviceCalls int
	e, cleanup := newExporterWithServer(t, ownerListServer(&deviceCalls, deviceJSON,
		[]string{cabledIfaceJSON(uuid.New(), "HSN 0", uuid.New())},
		[]string{ifaceResultJSON(uuid.New(), "HSN 0")}))
	defer cleanup()
	device := &devicetypes.CaniDeviceType{Name: "node-a"}

	if !e.interfaceCabled(cableEnd{device: device, port: "HSN 0", moduleID: uuid.New()}) {
		t.Error("module end: want cabled")
	}
	if e.interfaceCabled(cableEnd{device: device, port: "HSN 0"}) {
		t.Error("device end: want not cabled")
	}
}
