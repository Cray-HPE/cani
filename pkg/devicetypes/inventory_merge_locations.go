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
package devicetypes

import "github.com/google/uuid"

// MergeRacks merges incoming racks by UUID, source identity, or name.
func (inv *Inventory) MergeRacks(incoming map[uuid.UUID]*CaniRackType) map[uuid.UUID]uuid.UUID {
	remap := make(map[uuid.UUID]uuid.UUID)
	if inv.Racks == nil {
		inv.Racks = make(map[uuid.UUID]*CaniRackType)
	}
	for incomingID, rack := range incoming {
		if rack == nil || rack.Name == "" {
			continue
		}
		if existing, ok := inv.Racks[incomingID]; ok {
			mergeRackProperties(existing, rack)
			continue
		}
		matched := false
		for existingID, existing := range inv.Racks {
			if existing != nil && (sharedExternalID(existing.ExternalIDs, rack.ExternalIDs) ||
				(!conflictingExternalID(existing.ExternalIDs, rack.ExternalIDs) && existing.Name == rack.Name)) {
				rack.ID = existingID
				mergeRackProperties(existing, rack)
				remap[incomingID] = existingID
				matched = true
				break
			}
		}
		if !matched {
			inv.Racks[incomingID] = rack
		}
	}
	return remap
}

func mergeRackProperties(existing, incoming *CaniRackType) {
	if incoming.UHeight > 0 {
		existing.UHeight = incoming.UHeight
	}
	if incoming.Slug != "" {
		existing.Slug = incoming.Slug
	}
	if incoming.Manufacturer != "" {
		existing.Manufacturer = incoming.Manufacturer
	}
	if incoming.Model != "" {
		existing.Model = incoming.Model
	}
	if incoming.Serial != "" {
		existing.Serial = incoming.Serial
	}
	if incoming.Type != "" {
		existing.Type = incoming.Type
	}
	if incoming.Location != uuid.Nil {
		existing.Location = incoming.Location
	}
	mergeObjectMeta(&existing.ObjectMeta, incoming.ObjectMeta)
}

// MergeLocations merges incoming locations by UUID, source identity, or name.
func (inv *Inventory) MergeLocations(incoming map[uuid.UUID]*CaniLocationType) map[uuid.UUID]uuid.UUID {
	remap := make(map[uuid.UUID]uuid.UUID)
	if inv.Locations == nil {
		inv.Locations = make(map[uuid.UUID]*CaniLocationType)
	}
	for incomingID, location := range incoming {
		if location == nil || location.Name == "" {
			continue
		}
		if existing, ok := inv.Locations[incomingID]; ok {
			mergeLocationProperties(existing, location)
			continue
		}
		matched := false
		for existingID, existing := range inv.Locations {
			if existing != nil && (sharedExternalID(existing.ExternalIDs, location.ExternalIDs) ||
				(!conflictingExternalID(existing.ExternalIDs, location.ExternalIDs) && existing.Name == location.Name)) {
				location.ID = existingID
				mergeLocationProperties(existing, location)
				remap[incomingID] = existingID
				matched = true
				break
			}
		}
		if !matched {
			inv.Locations[incomingID] = location
		}
	}
	return remap
}

func mergeLocationProperties(existing, incoming *CaniLocationType) {
	if incoming.LocationType != "" {
		existing.LocationType = incoming.LocationType
	}
	if incoming.Description != "" {
		existing.Description = incoming.Description
	}
	if incoming.Facility != "" {
		existing.Facility = incoming.Facility
	}
	mergeObjectMeta(&existing.ObjectMeta, incoming.ObjectMeta)
}
