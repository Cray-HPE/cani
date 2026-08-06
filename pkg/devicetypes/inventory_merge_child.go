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
	"log"

	"github.com/google/uuid"
)

// --- Location merge ---

// MergeLocations merges incoming locations by UUID match, then name match, then insert.
// Returns a remap map from incoming UUID to existing UUID for name-matched locations.
func (inv *Inventory) MergeLocations(incoming map[uuid.UUID]*CaniLocationType) map[uuid.UUID]uuid.UUID {
	remap := make(map[uuid.UUID]uuid.UUID)
	if inv.Locations == nil {
		inv.Locations = make(map[uuid.UUID]*CaniLocationType)
	}

	for id, loc := range incoming {
		if loc == nil || loc.Name == "" {
			continue
		}

		// UUID match — field-level merge
		if existing, ok := inv.Locations[id]; ok {
			mergeLocationProperties(existing, loc)
			continue
		}

		// Name match — keep UUID, update fields
		found := false
		for existingID, existing := range inv.Locations {
			if existing != nil && existing.Name == loc.Name {
				mergeLocationProperties(existing, loc)
				remap[id] = existingID
				found = true
				break
			}
		}

		if !found {
			inv.Locations[id] = loc
		}
	}
	return remap
}

// mergeLocationProperties copies non-empty fields from incoming into existing.
func mergeLocationProperties(existing, incoming *CaniLocationType) {
	if incoming.LocationType != "" {
		existing.LocationType = incoming.LocationType
	}
	if incoming.Status != "" {
		existing.Status = incoming.Status
	}
	if incoming.Description != "" {
		existing.Description = incoming.Description
	}
	if incoming.Facility != "" {
		existing.Facility = incoming.Facility
	}
}

// --- Module merge ---

// MergeModules merges incoming modules by UUID match, then name match, then insert.
func (inv *Inventory) MergeModules(incoming map[uuid.UUID]*CaniModuleType) {
	if inv.Modules == nil {
		inv.Modules = make(map[uuid.UUID]*CaniModuleType)
	}

	for id, mod := range incoming {
		if mod == nil || mod.Name == "" {
			continue
		}
		if err := mod.Validate(); err != nil {
			log.Printf("Skipping invalid module %q: %v", mod.Name, err)
			continue
		}

		// UUID match — field-level merge
		if existing, ok := inv.Modules[id]; ok {
			mergeModuleProperties(existing, mod)
			continue
		}

		// Name match — keep UUID, update fields
		found := false
		for _, existing := range inv.Modules {
			if existing != nil && existing.Name == mod.Name {
				mergeModuleProperties(existing, mod)
				found = true
				break
			}
		}

		if !found {
			inv.Modules[id] = mod
		}
	}
}

// mergeModuleProperties copies non-empty fields from incoming into existing.
func mergeModuleProperties(existing, incoming *CaniModuleType) {
	if incoming.Slug != "" {
		existing.Slug = incoming.Slug
	}
	if incoming.Status != "" {
		existing.Status = incoming.Status
	}
	if incoming.Serial != "" {
		existing.Serial = incoming.Serial
	}
	if incoming.Manufacturer != "" {
		existing.Manufacturer = incoming.Manufacturer
	}
	if incoming.Model != "" {
		existing.Model = incoming.Model
	}
	if incoming.Type != "" {
		existing.Type = incoming.Type
	}
}

// --- FRU merge ---

// MergeFrus merges incoming FRUs by UUID match, then name match, then insert.
func (inv *Inventory) MergeFrus(incoming map[uuid.UUID]*CaniFruType) {
	if inv.Frus == nil {
		inv.Frus = make(map[uuid.UUID]*CaniFruType)
	}

	for id, fru := range incoming {
		if fru == nil || fru.Name == "" {
			continue
		}

		// UUID match — field-level merge
		if existing, ok := inv.Frus[id]; ok {
			mergeFruProperties(existing, fru)
			continue
		}

		// Name match — keep UUID, update fields
		found := false
		for _, existing := range inv.Frus {
			if existing != nil && existing.Name == fru.Name {
				mergeFruProperties(existing, fru)
				found = true
				break
			}
		}

		if !found {
			inv.Frus[id] = fru
		}
	}
}

// mergeFruProperties copies non-empty fields from incoming into existing.
func mergeFruProperties(existing, incoming *CaniFruType) {
	if incoming.PartNumber != "" {
		existing.PartNumber = incoming.PartNumber
	}
	if incoming.Serial != "" {
		existing.Serial = incoming.Serial
	}
	if incoming.Status != "" {
		existing.Status = incoming.Status
	}
	if incoming.Manufacturer != "" {
		existing.Manufacturer = incoming.Manufacturer
	}
}
