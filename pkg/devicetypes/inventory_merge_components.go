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
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
)

// MergeModules merges modules by UUID, external ID, identity (serial on the
// same type, or asset tag), or name and parent. An incoming module whose
// identity names two existing modules is a conflict the merge refuses rather
// than resolves: folding it into either would leave two records claiming one
// identity, which Nautobot's uniqueness rules forbid.
func (inv *Inventory) MergeModules(incoming map[uuid.UUID]*CaniModuleType) (map[uuid.UUID]uuid.UUID, error) {
	remap := make(map[uuid.UUID]uuid.UUID, len(incoming))
	if inv.Modules == nil {
		inv.Modules = make(map[uuid.UUID]*CaniModuleType)
	}
	for _, incomingID := range sortedIDs(incoming) {
		module := incoming[incomingID]
		if module == nil {
			continue
		}
		if err := module.Validate(); err != nil {
			log.Printf("Skipping invalid module %q: %v", module.Label(), err)
			continue
		}
		if existing, ok := inv.Modules[incomingID]; ok {
			mergeModuleProperties(existing, module)
			remap[incomingID] = incomingID
			continue
		}
		existingID, matched, err := inv.matchModule(module)
		if err != nil {
			return nil, err
		}
		if matched {
			mergeModuleProperties(inv.Modules[existingID], module)
			remap[incomingID] = existingID
			continue
		}
		inv.Modules[incomingID] = module
		remap[incomingID] = incomingID
	}
	return remap, nil
}

// matchModule finds the existing module an incoming one is, visiting existing
// modules in a stable order: a shared external ID settles it; otherwise the
// one module that shares its identity, or failing that the first that shares
// its name. Two existing modules sharing the incoming identity (one its
// serial, another its asset tag, for example) is an error naming all three.
func (inv *Inventory) matchModule(module *CaniModuleType) (uuid.UUID, bool, error) {
	var byIdentity []uuid.UUID
	byName := uuid.Nil
	for _, id := range sortedIDs(inv.Modules) {
		existing := inv.Modules[id]
		switch {
		case existing == nil:
		case sharedExternalID(existing.ExternalIDs, module.ExternalIDs):
			return id, true, nil
		case conflictingExternalID(existing.ExternalIDs, module.ExternalIDs):
		case sameModuleIdentity(existing, module):
			byIdentity = append(byIdentity, id)
		case byName == uuid.Nil && sameModuleName(existing, module):
			byName = id
		}
	}
	switch {
	case len(byIdentity) > 1:
		return uuid.Nil, false, fmt.Errorf("module %s matches %d existing modules by identity (%s); correct the serial or asset tag in the source or the datastore",
			module.Label(), len(byIdentity), inv.moduleLabels(byIdentity))
	case len(byIdentity) == 1:
		return byIdentity[0], true, nil
	case byName != uuid.Nil:
		return byName, true, nil
	}
	return uuid.Nil, false, nil
}

func (inv *Inventory) moduleLabels(ids []uuid.UUID) string {
	labels := make([]string, 0, len(ids))
	for _, id := range ids {
		labels = append(labels, fmt.Sprintf("%s %s", inv.Modules[id].Label(), id))
	}
	return strings.Join(labels, ", ")
}

func mergeModuleProperties(existing, incoming *CaniModuleType) {
	if incoming.Slug != "" {
		existing.Slug = incoming.Slug
	}
	if incoming.Serial != "" {
		existing.Serial = incoming.Serial
	}
	if incoming.AssetTag != "" {
		existing.AssetTag = incoming.AssetTag
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
	mergeModulePlacement(existing, incoming)
	mergeObjectMeta(&existing.ObjectMeta, incoming.ObjectMeta)
}

// mergeModulePlacement takes the incoming module's placement as one unit, so
// a module the provider reports in another bay or device, newly installed, or
// stored as a spare is placed where the provider reports it; field-by-field
// merging could leave a module both installed and stored. An incoming module
// with no placement leaves the existing one untouched, and an incoming module
// on the same parent with no bay name keeps the existing bay, because a
// provider that does not know the library's bay names cannot clear one.
func mergeModulePlacement(existing, incoming *CaniModuleType) {
	switch {
	case incoming.ParentDevice != uuid.Nil:
		if incoming.ParentDevice != existing.ParentDevice || incoming.ModuleBayName != "" {
			existing.ModuleBayName = incoming.ModuleBayName
		}
		existing.ParentDevice = incoming.ParentDevice
		existing.Location = incoming.Location
	case incoming.Location != uuid.Nil:
		existing.ParentDevice, existing.ModuleBayName, existing.Location = uuid.Nil, "", incoming.Location
	}
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

// mergeSet copies src over *dst when src is set and differs, reporting
// whether it did; a zero incoming value never clears an existing one.
func mergeSet[T comparable](dst *T, src T) bool {
	var zero T
	if src == zero || *dst == src {
		return false
	}
	*dst = src
	return true
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

// sameModuleIdentity reports whether two modules share Nautobot's module
// identity: the same asset tag, or the same serial on the same hardware
// type. The type is compared only when both modules record one, as the swap
// rule compares it, so a module imported without a type (Nautobot's module
// type reference carries no model) still converges on its serial with the
// authored record of the same card.
func sameModuleIdentity(a, b *CaniModuleType) bool {
	if a.AssetTag != "" && strings.EqualFold(a.AssetTag, b.AssetTag) {
		return true
	}
	if a.Serial == "" || !strings.EqualFold(a.Serial, b.Serial) {
		return false
	}
	typeA, typeB := moduleTypeKey(a), moduleTypeKey(b)
	return typeA == "" || typeB == "" || strings.EqualFold(typeA, typeB)
}

func sameModuleName(a, b *CaniModuleType) bool {
	return a.Name != "" && a.Name == b.Name &&
		(a.ParentDevice == b.ParentDevice || a.ParentDevice == uuid.Nil || b.ParentDevice == uuid.Nil)
}
