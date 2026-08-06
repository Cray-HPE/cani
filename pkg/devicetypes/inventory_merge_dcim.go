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
	"github.com/google/uuid"
)

// --- Rack merge ---

// MergeRacks merges incoming racks by UUID match, then name match, then insert.
// Returns a remap map from incoming UUID to existing UUID for name-matched racks.
func (inv *Inventory) MergeRacks(incoming map[uuid.UUID]*CaniRackType) map[uuid.UUID]uuid.UUID {
	remap := make(map[uuid.UUID]uuid.UUID)
	if inv.Racks == nil {
		inv.Racks = make(map[uuid.UUID]*CaniRackType)
	}

	for id, rack := range incoming {
		if rack == nil || rack.Name == "" {
			continue
		}

		// UUID match — field-level merge
		if existing, ok := inv.Racks[id]; ok {
			mergeRackProperties(existing, rack)
			continue
		}

		// Name match — keep UUID, update fields
		found := false
		for existingID, existing := range inv.Racks {
			if existing != nil && existing.Name == rack.Name {
				mergeRackProperties(existing, rack)
				remap[id] = existingID
				found = true
				break
			}
		}

		if !found {
			inv.Racks[id] = rack
		}
	}
	return remap
}

// mergeRackProperties copies non-empty fields from incoming into existing.
func mergeRackProperties(existing, incoming *CaniRackType) {
	if incoming.UHeight > 0 {
		existing.UHeight = incoming.UHeight
	}
	if incoming.Slug != "" {
		existing.Slug = incoming.Slug
	}
	if incoming.Status != "" {
		existing.Status = incoming.Status
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
	if len(incoming.ProviderMetadata) > 0 {
		if existing.ProviderMetadata == nil {
			existing.ProviderMetadata = make(map[string]any)
		}
		for k, v := range incoming.ProviderMetadata {
			existing.ProviderMetadata[k] = v
		}
	}
}

// --- Cable merge ---

// mergeCableByLabel finds an existing cable with the same label and overwrites it.
// Returns true if a match was found.
func (inv *Inventory) mergeCableByLabel(cable *CaniCableType) bool {
	if cable.Label == "" {
		return false
	}
	for _, existing := range inv.Cables {
		if existing != nil && existing.Label == cable.Label {
			*existing = *cable
			return true
		}
	}
	return false
}

// MergeCables merges incoming cables by UUID match, then label match, then insert.
func (inv *Inventory) MergeCables(incoming map[uuid.UUID]*CaniCableType) {
	if inv.Cables == nil {
		inv.Cables = make(map[uuid.UUID]*CaniCableType)
	}

	for id, cable := range incoming {
		if cable == nil {
			continue
		}

		// UUID match — overwrite
		if _, ok := inv.Cables[id]; ok {
			inv.Cables[id] = cable
			continue
		}

		// Label match — overwrite existing
		if inv.mergeCableByLabel(cable) {
			continue
		}

		inv.Cables[id] = cable
	}
}
