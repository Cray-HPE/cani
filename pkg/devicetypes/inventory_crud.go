/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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
package devicetypes

import (
	"fmt"
	"log"

	"github.com/google/uuid"
)

// --- Location & parent helpers ---

// EnsureLocation guarantees at least one location exists in the
// inventory and returns its UUID.  If no location exists a default
// one is created.
func (inv *Inventory) EnsureLocation() uuid.UUID {
	if inv.Locations == nil {
		inv.Locations = make(map[uuid.UUID]*CaniLocationType)
	}
	for _, loc := range inv.Locations {
		if loc != nil {
			return loc.ID
		}
	}

	loc := NewDefaultLocation()
	inv.Locations[loc.ID] = loc
	log.Printf("Created default location %s", loc.ID)
	return loc.ID
}

// AssignRacksToLocation sets the location of every rack that has no
// parent to the given locationID and records them in the location's
// Racks list.
func (inv *Inventory) AssignRacksToLocation(locationID uuid.UUID) {
	loc := inv.Locations[locationID]
	for id, rack := range inv.Racks {
		if rack != nil && rack.Location == uuid.Nil {
			rack.Location = locationID
			if loc != nil {
				loc.AddRack(id)
			}
			log.Printf("Assigned rack %s to location %s", rack.Name, locationID)
		}
	}
}

//	Location → Racks            (Rack.Location   ↔ Location.Racks)
//	Rack     → Devices          (Device.Parent   ↔ Rack.Devices)
//	Device   → Child Devices    (Device.Parent   ↔ Device.Children)
//
// It also rebuilds the derived Device.Rack/Location/ParentDevice FK fields,
// validates (without mutating) module, FRU, and cable references, and detects
// circular parent chains.
//
// The reverse lists are cleared and rebuilt from scratch so that only setting a
// Parent field is required; all other references are derived. Unlike
// VerifyParentChildRelationships it performs no logging, making it safe to run on
// every datastore load so persisted derived values are never trusted.
func (inv *Inventory) RebuildDerivedState() *RelationshipResult {
	result := &RelationshipResult{}

	result.merge(inv.rebuildLocationRelationships())
	result.merge(inv.rebuildRackRelationships())
	result.merge(inv.rebuildDeviceRelationships())
	result.merge(inv.validateModuleRelationships())
	result.merge(inv.rebuildInterfaceRelationships())
	result.merge(inv.rebuildFruRelationships())
	result.merge(inv.rebuildCableRelationships())
	result.merge(inv.validateCableRelationships())
	result.merge(inv.detectCircularLocationRefs())
	result.merge(inv.detectCircularDeviceRefs())

	return result
}

// VerifyParentChildRelationships rebuilds derived state (see RebuildDerivedState)
// and logs a summary of fixes, warnings, and errors. Mutation commands call this
// so operators see what changed; load paths call RebuildDerivedState directly to
// stay silent.
func (inv *Inventory) VerifyParentChildRelationships() *RelationshipResult {
	result := inv.RebuildDerivedState()
	result.logSummary()
	return result
}

// --- Cascading remove ---

// unlinkDeviceFromParent removes a device from its parent's Children list
// and clears rack slot occupancy.
func (inv *Inventory) unlinkDeviceFromParent(device *CaniDeviceType, id uuid.UUID) {
	if device.Parent == uuid.Nil {
		return
	}
	if parent, ok := inv.Devices[device.Parent]; ok {
		parent.Children = removeUUID(parent.Children, id)
	}
	if rack, ok := inv.Racks[device.Parent]; ok {
		rack.RemoveDevice(id)
	}
}

// removeCablesForDevice deletes all cables that terminate at the given device.
func (inv *Inventory) removeCablesForDevice(id uuid.UUID) {
	for cableID, cable := range inv.Cables {
		if cable == nil {
			continue
		}
		if cable.TerminationADevice == id || cable.TerminationBDevice == id {
			delete(inv.Cables, cableID)
		}
	}
}

// removeModulesForDevice deletes all modules whose parent is the given device.
func (inv *Inventory) removeModulesForDevice(id uuid.UUID) {
	for modID, mod := range inv.Modules {
		if mod != nil && mod.ParentDevice == id {
			inv.removeCablesForDevice(modID)
			delete(inv.Modules, modID)
		}
	}
}

// RemoveDevice deletes a device and cleans up all references:
//   - removes from parent's Children list
//   - removes rack slot occupancy
//   - deletes cables referencing the device
//   - deletes child modules belonging to the device
func (inv *Inventory) RemoveDevice(id uuid.UUID) error {
	device, exists := inv.Devices[id]
	if !exists {
		return fmt.Errorf("device %s not found", id)
	}

	inv.unlinkDeviceFromParent(device, id)
	inv.removeCablesForDevice(id)
	inv.removeModulesForDevice(id)

	for _, childID := range device.Children {
		_ = inv.RemoveDevice(childID) // best-effort
	}

	delete(inv.Devices, id)
	return nil
}

// --- helpers ---

func containsUUID(slice []uuid.UUID, target uuid.UUID) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}

func removeUUID(slice []uuid.UUID, target uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(slice))
	for _, v := range slice {
		if v != target {
			result = append(result, v)
		}
	}
	return result
}
