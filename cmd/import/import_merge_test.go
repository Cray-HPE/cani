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
package imprt

import (
	"testing"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// TestRemapDeviceReferences verifies every TransformResult foreign key that
// targets a device is re-pointed after a name-matched merge.
func TestRemapDeviceReferences(t *testing.T) {
	ephemeralA, existingA := uuid.New(), uuid.New()
	ephemeralB, existingB := uuid.New(), uuid.New()
	ifaceA, ifaceB := uuid.New(), uuid.New()
	unmapped := uuid.New()

	cableID := uuid.New()
	moduleID, fruID := uuid.New(), uuid.New()
	childID := uuid.New()
	result := &devicetypes.TransformResult{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			ephemeralA: {ID: ephemeralA, Parent: ephemeralB},
			childID:    {ID: childID, Parent: ephemeralA},
		},
		Modules: map[uuid.UUID]*devicetypes.CaniModuleType{
			moduleID: {ID: moduleID, ParentDevice: ephemeralA},
		},
		Frus: map[uuid.UUID]*devicetypes.CaniFruType{
			fruID: {ID: fruID, Device: ephemeralB},
		},
		Cables: map[uuid.UUID]*devicetypes.CaniCableType{
			cableID: {
				ID:                 cableID,
				TerminationADevice: ephemeralA,
				TerminationBDevice: ephemeralB,
				TerminationA:       ifaceA,
				TerminationB:       ifaceB,
			},
		},
	}
	inventory := devicetypes.NewInventory()
	inventory.Devices[existingA] = &devicetypes.CaniDeviceType{ID: existingA, Parent: ephemeralB}
	inventory.Devices[childID] = result.Devices[childID]

	result.RemapReferences(inventory, devicetypes.ReferenceRemaps{Devices: map[uuid.UUID]uuid.UUID{
		ephemeralA: existingA,
		ephemeralB: existingB,
		unmapped:   uuid.New(),
	}})

	if result.Devices[childID].Parent != existingA {
		t.Errorf("child Parent = %s, want %s", result.Devices[childID].Parent, existingA)
	}
	if inventory.Devices[existingA].Parent != existingB {
		t.Errorf("resolved device Parent = %s, want %s", inventory.Devices[existingA].Parent, existingB)
	}
	if result.Modules[moduleID].ParentDevice != existingA {
		t.Errorf("module ParentDevice = %s, want %s", result.Modules[moduleID].ParentDevice, existingA)
	}
	if result.Frus[fruID].Device != existingB {
		t.Errorf("fru Device = %s, want %s", result.Frus[fruID].Device, existingB)
	}
	got := result.Cables[cableID]
	if got.TerminationADevice != existingA {
		t.Errorf("TerminationADevice = %s, want %s", got.TerminationADevice, existingA)
	}
	if got.TerminationBDevice != existingB {
		t.Errorf("TerminationBDevice = %s, want %s", got.TerminationBDevice, existingB)
	}
	// Interface UUIDs are stable and must not be rewritten.
	if got.TerminationA != ifaceA || got.TerminationB != ifaceB {
		t.Errorf("interface terminations were modified: A=%s B=%s", got.TerminationA, got.TerminationB)
	}
}

// TestRemapDeviceReferencesNoRemap verifies an empty remap is a no-op.
func TestRemapDeviceReferencesNoRemap(t *testing.T) {
	devA, devB := uuid.New(), uuid.New()
	cableID := uuid.New()
	result := &devicetypes.TransformResult{
		Cables: map[uuid.UUID]*devicetypes.CaniCableType{
			cableID: {ID: cableID, TerminationADevice: devA, TerminationBDevice: devB},
		},
	}

	result.RemapReferences(devicetypes.NewInventory(), devicetypes.ReferenceRemaps{})

	if result.Cables[cableID].TerminationADevice != devA || result.Cables[cableID].TerminationBDevice != devB {
		t.Error("expected terminations unchanged with an empty remap")
	}
}

func TestRemapDeviceParentsRemapsExplicitRack(t *testing.T) {
	ephemeralRack, existingRack := uuid.New(), uuid.New()
	deviceID := uuid.New()
	devices := map[uuid.UUID]*devicetypes.CaniDeviceType{
		deviceID: {ID: deviceID, Parent: ephemeralRack, Rack: ephemeralRack},
	}

	result := &devicetypes.TransformResult{Devices: devices}
	result.RemapReferences(nil, devicetypes.ReferenceRemaps{
		Racks: map[uuid.UUID]uuid.UUID{ephemeralRack: existingRack},
	})

	if devices[deviceID].Parent != existingRack {
		t.Errorf("Parent = %s, want %s", devices[deviceID].Parent, existingRack)
	}
	if devices[deviceID].Rack != existingRack {
		t.Errorf("Rack = %s, want %s", devices[deviceID].Rack, existingRack)
	}
}

func TestMergeTransformResultRemapsDeviceChildren(t *testing.T) {
	cfg := config.Cfg
	config.Cfg = &config.Config{}
	t.Cleanup(func() { config.Cfg = cfg })

	existingDeviceID, incomingDeviceID := uuid.New(), uuid.New()
	moduleID, fruID := uuid.New(), uuid.New()
	inventory := devicetypes.NewInventory()
	inventory.Devices[existingDeviceID] = &devicetypes.CaniDeviceType{
		ID:   existingDeviceID,
		Name: "compute-001",
	}
	result := &devicetypes.TransformResult{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			incomingDeviceID: {ID: incomingDeviceID, Name: "compute-001"},
		},
		Modules: map[uuid.UUID]*devicetypes.CaniModuleType{
			moduleID: {ID: moduleID, Name: "nic-1", ParentDevice: incomingDeviceID},
		},
		Frus: map[uuid.UUID]*devicetypes.CaniFruType{
			fruID: {ID: fruID, Name: "dimm-1", Device: incomingDeviceID},
		},
	}

	if err := mergeTransformResult(&etlContext{inventory: inventory}, result); err != nil {
		t.Fatalf("mergeTransformResult() error = %v", err)
	}
	if inventory.Modules[moduleID].ParentDevice != existingDeviceID {
		t.Errorf("module ParentDevice = %s, want %s", inventory.Modules[moduleID].ParentDevice, existingDeviceID)
	}
	if inventory.Frus[fruID].Device != existingDeviceID {
		t.Errorf("fru Device = %s, want %s", inventory.Frus[fruID].Device, existingDeviceID)
	}
}

func TestMergeTransformResultRemapsNestedFrusOnRepeatImport(t *testing.T) {
	deviceID := uuid.New()
	existingParentID, existingChildID := uuid.New(), uuid.New()
	incomingParentID, incomingChildID := uuid.New(), uuid.New()
	inventory := devicetypes.NewInventory()
	inventory.Devices[deviceID] = &devicetypes.CaniDeviceType{ID: deviceID, Name: "server-1"}
	inventory.Frus[existingParentID] = &devicetypes.CaniFruType{
		ID: existingParentID, Name: "system-board", Device: deviceID,
	}
	inventory.Frus[existingChildID] = &devicetypes.CaniFruType{
		ID: existingChildID, Name: "dimm-1", Device: deviceID, Parent: existingParentID,
	}
	result := &devicetypes.TransformResult{
		Frus: map[uuid.UUID]*devicetypes.CaniFruType{
			incomingParentID: {ID: incomingParentID, Name: "system-board", Device: deviceID},
			incomingChildID:  {ID: incomingChildID, Name: "dimm-1", Device: deviceID, Parent: incomingParentID},
		},
	}

	if err := mergeTransformResult(&etlContext{inventory: inventory}, result); err != nil {
		t.Fatalf("mergeTransformResult() error = %v", err)
	}
	if len(inventory.Frus) != 2 {
		t.Fatalf("FRU count = %d, want 2", len(inventory.Frus))
	}
	if inventory.Frus[existingChildID].Parent != existingParentID {
		t.Errorf("child FRU parent = %s, want %s", inventory.Frus[existingChildID].Parent, existingParentID)
	}
}

func TestMergeTransformResultRemapsNestedLocationsOnRepeatImport(t *testing.T) {
	existingRootID, existingChildID := uuid.New(), uuid.New()
	incomingRootID, incomingChildID := uuid.New(), uuid.New()
	rootSourceID, childSourceID := uuid.New(), uuid.New()
	inventory := devicetypes.NewInventory()
	inventory.Locations[existingRootID] = &devicetypes.CaniLocationType{
		ID: existingRootID, Name: "campus", LocationType: "site",
		ObjectMeta: devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": rootSourceID}},
	}
	inventory.Locations[existingChildID] = &devicetypes.CaniLocationType{
		ID: existingChildID, Name: "room-1", LocationType: "room", Parent: existingRootID,
		ObjectMeta: devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": childSourceID}},
	}
	result := &devicetypes.TransformResult{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			incomingRootID: {
				ID: incomingRootID, Name: "campus", LocationType: "site",
				ObjectMeta: devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": rootSourceID}},
			},
			incomingChildID: {
				ID: incomingChildID, Name: "room-1", LocationType: "room", Parent: incomingRootID,
				ObjectMeta: devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": childSourceID}},
			},
		},
	}

	if err := mergeTransformResult(&etlContext{inventory: inventory}, result); err != nil {
		t.Fatalf("mergeTransformResult() error = %v", err)
	}
	if len(inventory.Locations) != 2 {
		t.Fatalf("location count = %d, want 2", len(inventory.Locations))
	}
	if inventory.Locations[existingChildID].Parent != existingRootID {
		t.Errorf("child location parent = %s, want %s", inventory.Locations[existingChildID].Parent, existingRootID)
	}
}

func TestMergeTransformResultRemapsVRFDevicesOnRepeatImport(t *testing.T) {
	existingDeviceID, incomingDeviceID := uuid.New(), uuid.New()
	vrfID := uuid.New()
	inventory := devicetypes.NewInventory()
	inventory.Devices[existingDeviceID] = &devicetypes.CaniDeviceType{
		ID: existingDeviceID, Name: "leaf-1",
	}
	result := &devicetypes.TransformResult{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			incomingDeviceID: {ID: incomingDeviceID, Name: "leaf-1"},
		},
		VRFs: map[uuid.UUID]*devicetypes.CaniVRF{
			vrfID: {ID: vrfID, Name: "BLUE", Devices: []uuid.UUID{incomingDeviceID}},
		},
	}

	if err := mergeTransformResult(&etlContext{inventory: inventory}, result); err != nil {
		t.Fatalf("mergeTransformResult() error = %v", err)
	}
	if got := inventory.VRFs[vrfID].Devices; len(got) != 1 || got[0] != existingDeviceID {
		t.Fatalf("VRF devices = %v, want [%s]", got, existingDeviceID)
	}
}

func TestMergeTransformResultDeduplicatesUnlabeledCableOnRepeatImport(t *testing.T) {
	existingDeviceA, existingDeviceB := uuid.New(), uuid.New()
	incomingDeviceA, incomingDeviceB := uuid.New(), uuid.New()
	interfaceA, interfaceB := uuid.New(), uuid.New()
	existingCableID, incomingCableID := uuid.New(), uuid.New()
	inventory := devicetypes.NewInventory()
	inventory.Devices[existingDeviceA] = &devicetypes.CaniDeviceType{
		ID: existingDeviceA, Name: "leaf-a", Interfaces: []devicetypes.InterfaceSpec{{ID: interfaceA, Name: "eth0"}},
	}
	inventory.Devices[existingDeviceB] = &devicetypes.CaniDeviceType{
		ID: existingDeviceB, Name: "leaf-b", Interfaces: []devicetypes.InterfaceSpec{{ID: interfaceB, Name: "eth0"}},
	}
	inventory.Cables[existingCableID] = &devicetypes.CaniCableType{
		ID: existingCableID, TerminationA: interfaceA, TerminationB: interfaceB,
		TerminationADevice: existingDeviceA, TerminationBDevice: existingDeviceB,
		TerminationAPort: "eth0", TerminationBPort: "eth0",
	}
	result := &devicetypes.TransformResult{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			incomingDeviceA: {
				ID: incomingDeviceA, Name: "leaf-a", Interfaces: []devicetypes.InterfaceSpec{{ID: interfaceA, Name: "eth0"}},
			},
			incomingDeviceB: {
				ID: incomingDeviceB, Name: "leaf-b", Interfaces: []devicetypes.InterfaceSpec{{ID: interfaceB, Name: "eth0"}},
			},
		},
		Cables: map[uuid.UUID]*devicetypes.CaniCableType{
			incomingCableID: {
				ID: incomingCableID, TerminationA: interfaceB, TerminationB: interfaceA,
				TerminationADevice: incomingDeviceB, TerminationBDevice: incomingDeviceA,
				TerminationAPort: "eth0", TerminationBPort: "eth0",
			},
		},
	}

	if err := mergeTransformResult(&etlContext{inventory: inventory}, result); err != nil {
		t.Fatalf("mergeTransformResult() error = %v", err)
	}
	if len(inventory.Cables) != 1 {
		t.Fatalf("cable count = %d, want 1", len(inventory.Cables))
	}
	if inventory.Cables[existingCableID].ID != existingCableID {
		t.Errorf("cable ID = %s, want retained ID %s", inventory.Cables[existingCableID].ID, existingCableID)
	}
}

func TestMergeTransformResultRemapsStrictUnclassifiedDeviceChildren(t *testing.T) {
	originalConfig := config.Cfg
	originalStep := stepFlag
	config.Cfg = &config.Config{Strict: true}
	stepFlag = false
	t.Cleanup(func() {
		config.Cfg = originalConfig
		stepFlag = originalStep
	})

	existingDeviceID, incomingDeviceID := uuid.New(), uuid.New()
	moduleID := uuid.New()
	inventory := devicetypes.NewInventory()
	inventory.Devices[existingDeviceID] = &devicetypes.CaniDeviceType{
		ID:   existingDeviceID,
		Name: "unclassified-001",
	}
	result := &devicetypes.TransformResult{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			incomingDeviceID: {ID: incomingDeviceID, Name: "unclassified-001"},
		},
		Modules: map[uuid.UUID]*devicetypes.CaniModuleType{
			moduleID: {ID: moduleID, Name: "nic-1", ParentDevice: incomingDeviceID},
		},
	}

	if err := mergeTransformResult(&etlContext{inventory: inventory}, result); err != nil {
		t.Fatalf("mergeTransformResult() error = %v", err)
	}
	if inventory.Modules[moduleID].ParentDevice != existingDeviceID {
		t.Errorf("module ParentDevice = %s, want %s", inventory.Modules[moduleID].ParentDevice, existingDeviceID)
	}
}

func TestMergeTransformResultRemapsIPAMReferences(t *testing.T) {
	existingLocationID, incomingLocationID := uuid.New(), uuid.New()
	existingVLANID, incomingVLANID := uuid.New(), uuid.New()
	existingPrefixID, incomingPrefixID := uuid.New(), uuid.New()
	inventory := devicetypes.NewInventory()
	inventory.Locations[existingLocationID] = &devicetypes.CaniLocationType{
		ID: existingLocationID, Name: "test-dc",
	}
	inventory.VLANs[existingVLANID] = &devicetypes.CaniVLAN{
		ID: existingVLANID, VID: 3101, Name: "cani-exported-vlan", Location: existingLocationID,
	}
	inventory.Prefixes[existingPrefixID] = &devicetypes.CaniPrefix{
		ID: existingPrefixID, Prefix: "10.31.1.0/24", Location: existingLocationID, VLAN: existingVLANID,
	}
	result := &devicetypes.TransformResult{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			incomingLocationID: {ID: incomingLocationID, Name: "test-dc"},
		},
		VLANs: map[uuid.UUID]*devicetypes.CaniVLAN{
			incomingVLANID: {
				ID: incomingVLANID, VID: 3101, Name: "cani-exported-vlan", Location: incomingLocationID,
			},
		},
		Prefixes: map[uuid.UUID]*devicetypes.CaniPrefix{
			incomingPrefixID: {
				ID: incomingPrefixID, Prefix: "10.31.1.0/24", Location: incomingLocationID, VLAN: incomingVLANID,
			},
		},
	}

	if err := mergeTransformResult(&etlContext{inventory: inventory}, result); err != nil {
		t.Fatalf("mergeTransformResult() error = %v", err)
	}
	if len(inventory.VLANs) != 1 {
		t.Fatalf("VLAN count = %d, want 1", len(inventory.VLANs))
	}
	if got := inventory.VLANs[existingVLANID]; got == nil || got.Location != existingLocationID {
		t.Fatalf("merged VLAN = %+v, want existing location %s", got, existingLocationID)
	}
	if len(inventory.Prefixes) != 1 {
		t.Fatalf("prefix count = %d, want 1", len(inventory.Prefixes))
	}
	if got := inventory.Prefixes[existingPrefixID]; got == nil ||
		got.Location != existingLocationID || got.VLAN != existingVLANID {
		t.Fatalf("merged prefix = %+v, want location %s and VLAN %s", got, existingLocationID, existingVLANID)
	}
}
