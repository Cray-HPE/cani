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

	remapDeviceReferences(inventory, result, map[uuid.UUID]uuid.UUID{
		ephemeralA: existingA,
		ephemeralB: existingB,
		unmapped:   uuid.New(),
	})

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

	remapDeviceReferences(devicetypes.NewInventory(), result, nil)

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

	remapDeviceParents(devices, nil, map[uuid.UUID]uuid.UUID{ephemeralRack: existingRack})

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
