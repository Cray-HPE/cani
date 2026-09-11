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

import (
	"log"

	"github.com/google/uuid"
)

// MergeModules merges modules by UUID, source identity, or name and parent.
func (inv *Inventory) MergeModules(incoming map[uuid.UUID]*CaniModuleType) map[uuid.UUID]uuid.UUID {
	remap := make(map[uuid.UUID]uuid.UUID, len(incoming))
	if inv.Modules == nil {
		inv.Modules = make(map[uuid.UUID]*CaniModuleType)
	}
	for incomingID, module := range incoming {
		if module == nil || module.Name == "" {
			continue
		}
		if err := module.Validate(); err != nil {
			log.Printf("Skipping invalid module %q: %v", module.Name, err)
			continue
		}
		if existing, ok := inv.Modules[incomingID]; ok {
			mergeModuleProperties(existing, module)
			remap[incomingID] = incomingID
			continue
		}
		matched := false
		for existingID, existing := range inv.Modules {
			if existing != nil && (sharedExternalID(existing.ExternalIDs, module.ExternalIDs) ||
				(!conflictingExternalID(existing.ExternalIDs, module.ExternalIDs) && existing.Name == module.Name &&
					(existing.ParentDevice == module.ParentDevice || existing.ParentDevice == uuid.Nil || module.ParentDevice == uuid.Nil))) {
				mergeModuleProperties(existing, module)
				remap[incomingID] = existingID
				matched = true
				break
			}
		}
		if !matched {
			inv.Modules[incomingID] = module
			remap[incomingID] = incomingID
		}
	}
	return remap
}

func mergeModuleProperties(existing, incoming *CaniModuleType) {
	if incoming.Slug != "" {
		existing.Slug = incoming.Slug
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
	mergeObjectMeta(&existing.ObjectMeta, incoming.ObjectMeta)
}

// MergeFrus merges FRUs by UUID, source identity, or name and owner.
func (inv *Inventory) MergeFrus(incoming map[uuid.UUID]*CaniFruType) map[uuid.UUID]uuid.UUID {
	remap := make(map[uuid.UUID]uuid.UUID, len(incoming))
	if inv.Frus == nil {
		inv.Frus = make(map[uuid.UUID]*CaniFruType)
	}
	for incomingID, fru := range incoming {
		if fru == nil || fru.Name == "" {
			continue
		}
		if existing, ok := inv.Frus[incomingID]; ok {
			mergeFruProperties(existing, fru)
			remap[incomingID] = incomingID
			continue
		}
		matched := false
		for existingID, existing := range inv.Frus {
			if existing != nil && (sharedExternalID(existing.ExternalIDs, fru.ExternalIDs) ||
				(!conflictingExternalID(existing.ExternalIDs, fru.ExternalIDs) &&
					existing.Name == fru.Name && existing.Device == fru.Device)) {
				mergeFruProperties(existing, fru)
				remap[incomingID] = existingID
				matched = true
				break
			}
		}
		if !matched {
			inv.Frus[incomingID] = fru
			remap[incomingID] = incomingID
		}
	}
	return remap
}

func mergeFruProperties(existing, incoming *CaniFruType) {
	if incoming.PartNumber != "" {
		existing.PartNumber = incoming.PartNumber
	}
	if incoming.Serial != "" {
		existing.Serial = incoming.Serial
	}
	if incoming.Manufacturer != "" {
		existing.Manufacturer = incoming.Manufacturer
	}
	mergeObjectMeta(&existing.ObjectMeta, incoming.ObjectMeta)
}

func mergeObjectMeta(existing *ObjectMeta, incoming ObjectMeta) {
	if incoming.Status != "" {
		existing.Status = incoming.Status
	}
	if incoming.Role != "" {
		existing.Role = incoming.Role
	}
	if incoming.Tenant != "" {
		existing.Tenant = incoming.Tenant
	}
	if incoming.Tags != nil {
		existing.Tags = append([]string(nil), incoming.Tags...)
	}
	if incoming.CustomFields != nil {
		if existing.CustomFields == nil {
			existing.CustomFields = make(map[string]any)
		}
		for key, value := range incoming.CustomFields {
			existing.CustomFields[key] = value
		}
	}
	if incoming.ExternalIDs != nil {
		if existing.ExternalIDs == nil {
			existing.ExternalIDs = make(map[string]uuid.UUID)
		}
		for source, id := range incoming.ExternalIDs {
			existing.ExternalIDs[source] = id
		}
	}
	if incoming.ProviderMetadata != nil {
		mergeProviderMetadata(existing, incoming.ProviderMetadata)
	}
}

func mergeProviderMetadata(existing *ObjectMeta, incoming map[string]any) {
	if existing.ProviderMetadata == nil {
		existing.ProviderMetadata = make(map[string]any)
	}
	for provider, value := range incoming {
		incomingMap, incomingOK := value.(map[string]any)
		existingMap, existingOK := existing.ProviderMetadata[provider].(map[string]any)
		if !incomingOK || !existingOK {
			existing.ProviderMetadata[provider] = value
			continue
		}
		for key, item := range incomingMap {
			existingMap[key] = item
		}
	}
}
