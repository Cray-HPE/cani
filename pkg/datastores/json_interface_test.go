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
package datastores

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// TestJSONStorePreservesAuthoredInterfaceFields verifies interface instance
// data survives conversion to embedded specs and repeated save/load cycles.
//
// Why it matters: loading must not discard management or custom-field data.
// Inputs: an added interface with all authored metadata and switchport fields.
// Outputs: the complete interface is unchanged after each round trip.
// Data choice: explicit nonzero fields expose omissions in either conversion.
func TestJSONStorePreservesAuthoredInterfaceFields(t *testing.T) {
	inventory := devicetypes.NewInventory()
	deviceID := uuid.New()
	inventory.Devices[deviceID] = &devicetypes.CaniDeviceType{ID: deviceID, Name: "switch"}
	iface := &devicetypes.CaniInterface{
		ID: uuid.New(), Name: "mgmt0", DeviceID: deviceID, InterfaceType: "1000base-t",
		MgmtOnly: true, Label: "Management", MacAddress: "02:00:00:00:00:01",
		Lag: "bond0", Mode: "tagged", UntaggedVLAN: 10, TaggedVLANs: []int{20, 30},
		VRF: "operations", Description: "Out-of-band management", ContentType: "dcim.interface",
		ObjectMeta: devicetypes.ObjectMeta{
			Status: string(devicetypes.StatusActive), Role: "management", Tags: []string{"critical"},
			Tenant: "operations", CustomFields: map[string]any{"owner": "network-team"},
			ExternalIDs:      map[string]uuid.UUID{"source": uuid.New()},
			ProviderMetadata: map[string]any{"source": map[string]any{"port": "mgmt0"}},
		},
	}
	if err := inventory.AddInterface(iface); err != nil {
		t.Fatalf("AddInterface() error = %v", err)
	}
	store := &JSONStore{Path: filepath.Join(t.TempDir(), "inventory.json")}

	for range 2 {
		if err := store.Save(inventory); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		loaded, err := store.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if got := loaded.Interfaces[iface.ID]; !reflect.DeepEqual(got, iface) {
			t.Fatalf("loaded interface = %#v, want %#v", got, iface)
		}
		inventory = loaded
	}
}

// TestJSONStoreRebuildsInterfaceIPAssignments verifies interface IP lists are
// derived from address assignments rather than stale serialized reverse links.
//
// Why it matters: assignment identity must survive rebuilding interface objects.
// Inputs: one assigned IP and a stale reverse list on its interface.
// Outputs: repeated rebuilds produce only the actual address UUID, once.
// Data choice: a different cached UUID exposes trusting or dropping reverse data.
func TestJSONStoreRebuildsInterfaceIPAssignments(t *testing.T) {
	inventory := devicetypes.NewInventory()
	deviceID, interfaceID, addressID := uuid.New(), uuid.New(), uuid.New()
	inventory.Devices[deviceID] = &devicetypes.CaniDeviceType{ID: deviceID, Name: "host"}
	iface := &devicetypes.CaniInterface{
		ID: interfaceID, DeviceID: deviceID, Name: "eth0", IPAddresses: []uuid.UUID{uuid.New()},
	}
	if err := inventory.AddInterface(iface); err != nil {
		t.Fatalf("AddInterface() error = %v", err)
	}
	inventory.IPAddresses[addressID] = &devicetypes.CaniIPAddress{
		ID: addressID, Address: "10.0.0.1/24", Interfaces: []uuid.UUID{interfaceID},
	}
	store := &JSONStore{Path: filepath.Join(t.TempDir(), "inventory.json")}
	if err := store.Save(inventory); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if err := loaded.RebuildDerivedState().Err(); err != nil {
		t.Fatalf("rebuild error = %v", err)
	}

	if got := loaded.Interfaces[interfaceID].IPAddresses; !reflect.DeepEqual(got, []uuid.UUID{addressID}) {
		t.Errorf("interface IPs = %v, want [%s]", got, addressID)
	}
}

func TestJSONStorePreservesCustomModulePortTemplates(t *testing.T) {
	inventory := devicetypes.NewInventory()
	deviceID, moduleID := uuid.New(), uuid.New()
	inventory.Devices[deviceID] = &devicetypes.CaniDeviceType{
		ID: deviceID, Name: "node", ModuleBays: []devicetypes.ModuleBaySpec{
			{Name: "PCIe1", Position: "PCIe1"}, {Name: "Slot 1", Position: "1"},
		},
	}
	module := &devicetypes.CaniModuleType{
		ID: moduleID, Name: "custom-nic", ParentDevice: deviceID, ModuleBayName: "PCIe1",
		Interfaces: []devicetypes.InterfaceSpec{
			{Name: "custom-{module}/0", Label: "port {module}"},
			{Name: "custom-{module}/1"},
		},
	}
	if err := inventory.AddModule(module); err != nil {
		t.Fatal(err)
	}
	cable := &devicetypes.CaniCableType{
		ID: uuid.New(), Slug: "test-cable",
		TerminationA: module.Interfaces[0].ID, TerminationADevice: deviceID, TerminationAPort: module.Interfaces[0].Name,
		TerminationB: module.Interfaces[1].ID, TerminationBDevice: moduleID, TerminationBPort: module.Interfaces[1].Name,
	}
	if err := inventory.AddCable(cable); err != nil {
		t.Fatal(err)
	}
	portID := module.Interfaces[0].ID
	store := &JSONStore{Path: filepath.Join(t.TempDir(), "inventory.json")}
	for _, move := range []struct{ bay, position string }{{"Slot 1", "1"}, {"PCIe1", "PCIe1"}} {
		inventory = moduleTemplateStoreRoundTrip(t, store, inventory)
		assertStoredModuleCable(t, inventory, moduleID, cable.ID)
		module = inventory.Modules[moduleID]
		if module == nil {
			t.Fatal("module missing after datastore load")
		}
		module.ModuleBayName = move.bay
		inventory.ResolveModuleNames(module)
		if module.Interfaces[0].Name != "custom-"+move.position+"/0" || module.Interfaces[0].Label != "port "+move.position {
			t.Fatalf("templates lost across save/load: %+v", module.Interfaces[0])
		}
	}
	inventory = moduleTemplateStoreRoundTrip(t, store, inventory)
	assertStoredModuleCable(t, inventory, moduleID, cable.ID)
	if port := inventory.Interfaces[portID]; port == nil || port.Name != "custom-PCIe1/0" {
		t.Fatalf("normalized interface lost identity or resolved name: %+v", port)
	}
}

func assertStoredModuleCable(t *testing.T, inventory *devicetypes.Inventory, moduleID, cableID uuid.UUID) {
	t.Helper()
	module, cable := inventory.Modules[moduleID], inventory.Cables[cableID]
	if module == nil || cable == nil {
		t.Fatal("module or cable missing after datastore load")
	}
	if cable.TerminationA != module.Interfaces[0].ID || cable.TerminationAPort != module.Interfaces[0].Name ||
		cable.TerminationB != module.Interfaces[1].ID || cable.TerminationBPort != module.Interfaces[1].Name {
		t.Fatalf("stored cable no longer matches module ports: %+v", cable)
	}
	if cable.TerminationADevice != module.ParentDevice || cable.TerminationBDevice != module.ID {
		t.Fatal("stored cable endpoint ownership changed")
	}
	if err := inventory.RebuildDerivedState().Err(); err != nil {
		t.Fatalf("stored module move left inconsistent relationships: %v", err)
	}
}

func moduleTemplateStoreRoundTrip(t *testing.T, store *JSONStore, inventory *devicetypes.Inventory) *devicetypes.Inventory {
	t.Helper()
	if err := store.Save(inventory); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	return loaded
}
