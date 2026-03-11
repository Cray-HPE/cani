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

// AddDevices inserts new devices into the inventory.  Returns an error
// if any UUID already exists.
func (inv *Inventory) AddDevices(batch map[uuid.UUID]*CaniDeviceType) error {
	for id, device := range batch {
		if _, exists := inv.Devices[id]; exists {
			return fmt.Errorf("device with ID %s already exists", id)
		}
		inv.Devices[id] = device
	}
	result := inv.VerifyParentChildRelationships()
	if result.HasErrors() {
		return fmt.Errorf("relationship errors after adding devices: %v", result.Errors)
	}
	return nil
}

// --- Device merge (moved from inventory.go, renamed from Merge) ---

// MergeDevices merges new devices into the inventory by UUID match, then name
// match, then insert.  After changes it verifies parent-child relationships.
func (inv *Inventory) MergeDevices(incoming map[uuid.UUID]*CaniDeviceType) {
	inv.MergeDevicesStrict(incoming, false)
}

// MergeDevicesStrict behaves like MergeDevices but when strict is true it
// skips any device whose Slug and Model are both empty (unclassified). Skipped
// devices are collected and returned so callers can report or interactively
// resolve them.
func (inv *Inventory) MergeDevicesStrict(incoming map[uuid.UUID]*CaniDeviceType, strict bool) []UnclassifiedDevice {
	if inv.Devices == nil {
		inv.Devices = make(map[uuid.UUID]*CaniDeviceType)
	}

	changesDetected := false
	var skipped []UnclassifiedDevice

	for id, device := range incoming {
		if device == nil || device.Name == "" {
			continue
		}

		// In strict mode, reject devices without a slug or model.
		if strict && device.Slug == "" && device.Model == "" {
			skipped = append(skipped, UnclassifiedDevice{
				ID:           id,
				Name:         device.Name,
				HardwareType: device.HardwareType,
				Model:        device.Model,
				Manufacturer: device.Manufacturer,
			})
			continue
		}

		// Case 1: Same UUID already exists
		if existing, ok := inv.Devices[id]; ok {
			inv.unindexDevice(id, existing)
			if existing.MergeProperties(device) {
				changesDetected = true
			}
			inv.indexDevice(id, existing)
			continue
		}

		// Case 2: Existing device with same name
		if inv.mergeByName(device) {
			changesDetected = true
			continue
		}

		// Case 3: New device
		inv.Devices[id] = device
		inv.indexDevice(id, device)
		changesDetected = true
	}

	if changesDetected {
		log.Printf("Changes detected during merge")
	}
	return skipped
}

// mergeByName finds an existing device with the same name and merges
// properties. If the incoming device carries provider metadata with a
// distinguishing key (e.g. bmc_fqdn, bmc_hostname) it must also match
// the existing device; otherwise two servers of the same model but with
// different BMC identities would be collapsed into one.
func (inv *Inventory) mergeByName(device *CaniDeviceType) bool {
	for id, existing := range inv.Devices {
		if existing == nil || existing.Name != device.Name {
			continue
		}
		if !providerIdentityCompatible(existing, device) {
			continue
		}
		inv.unindexDevice(id, existing)
		changed := existing.MergeProperties(device)
		inv.indexDevice(id, existing)
		return changed
	}
	return false
}

// providerIdentityKeys lists metadata keys that uniquely identify a
// physical endpoint. If both devices define any of these keys the values
// must match for the devices to be considered the same.
var providerIdentityKeys = []string{"bmc_fqdn", "bmc_hostname", "xname"}

// providerIdentityCompatible returns true when two devices can be safely
// merged. If neither device has distinguishing provider metadata, they are
// compatible (legacy behaviour). If both have an identity key, the values
// must agree.
func providerIdentityCompatible(a, b *CaniDeviceType) bool {
	for _, key := range providerIdentityKeys {
		va, okA := a.GetProviderMeta(key)
		vb, okB := b.GetProviderMeta(key)
		if okA && okB {
			// Both define this key — values must match.
			return fmt.Sprintf("%v", va) == fmt.Sprintf("%v", vb)
		}
	}
	// No shared identity keys — fall back to name-only merge.
	return true
}

// FindDeviceByProviderKey looks up a device whose
// ProviderMetadata[provider][key] matches value. Uses the O(1) provider-key
// index when available, falling back to a linear scan otherwise.
func (inv *Inventory) FindDeviceByProviderKey(provider, key string, value any) *CaniDeviceType {
	if inv == nil || inv.Devices == nil || value == nil || value == "" {
		return nil
	}

	// Fast path: use the index.
	if id := inv.lookupProviderKey(provider, key, value); id != uuid.Nil {
		if dev, ok := inv.Devices[id]; ok {
			return dev
		}
	}

	// Slow path: linear scan (index may not be built yet).
	valStr := toIndexValue(value)
	if valStr == "" {
		return nil
	}
	for _, dev := range inv.Devices {
		if dev == nil || dev.ProviderMetadata == nil {
			continue
		}
		sub, ok := dev.ProviderMetadata[provider].(map[string]any)
		if !ok {
			continue
		}
		if toIndexValue(sub[key]) == valStr {
			return dev
		}
	}
	return nil
}

// FindDeviceByProviderKeys scans inventory devices for one whose
// ProviderMetadata[provider] matches any of the given key/value pairs.
// Checks are tried in order; returns on first match.
func (inv *Inventory) FindDeviceByProviderKeys(provider string, checks []ProviderKeyCheck) *CaniDeviceType {
	for _, chk := range checks {
		if chk.Value == nil || chk.Value == "" {
			continue
		}
		if dev := inv.FindDeviceByProviderKey(provider, chk.Key, chk.Value); dev != nil {
			return dev
		}
	}
	return nil
}

// ProviderKeyCheck pairs a metadata key with a value for lookup.
type ProviderKeyCheck struct {
	Key   string
	Value any
}

// --- Rack merge ---

// MergeRacks merges incoming racks by UUID match, then name match, then insert.
func (inv *Inventory) MergeRacks(incoming map[uuid.UUID]*CaniRackType) {
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
		for _, existing := range inv.Racks {
			if existing != nil && existing.Name == rack.Name {
				mergeRackProperties(existing, rack)
				found = true
				break
			}
		}

		if !found {
			inv.Racks[id] = rack
		}
	}
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
	if incoming.HardwareType != "" {
		existing.HardwareType = incoming.HardwareType
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

// --- Location merge ---

// MergeLocations merges incoming locations by UUID match, then name match, then insert.
func (inv *Inventory) MergeLocations(incoming map[uuid.UUID]*CaniLocationType) {
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
		for _, existing := range inv.Locations {
			if existing != nil && existing.Name == loc.Name {
				mergeLocationProperties(existing, loc)
				found = true
				break
			}
		}

		if !found {
			inv.Locations[id] = loc
		}
	}
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
	if incoming.HardwareType != "" {
		existing.HardwareType = incoming.HardwareType
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

// --- Relationship verification ---

// VerifyParentChildRelationships rebuilds all bidirectional parent-child
// references across the full inventory hierarchy:
//
//	Location → Child Locations  (Location.Parent ↔ Location.Children)
//	Location → Racks            (Rack.Location   ↔ Location.Racks)
//	Rack     → Devices          (Device.Parent   ↔ Rack.Devices)
//	Device   → Child Devices    (Device.Parent   ↔ Device.Children)
//
// It also validates (without mutating) module, FRU, and cable references
// and detects circular parent chains.
//
// The function clears and rebuilds all reverse lists from scratch so that
// only setting a Parent field is required; all other references are derived.
func (inv *Inventory) VerifyParentChildRelationships() *RelationshipResult {
	result := &RelationshipResult{}

	result.merge(inv.rebuildLocationRelationships())
	result.merge(inv.rebuildRackRelationships())
	result.merge(inv.rebuildDeviceRelationships())
	result.merge(inv.validateModuleRelationships())
	result.merge(inv.rebuildInterfaceRelationships())
	result.merge(inv.rebuildFruRelationships())
	result.merge(inv.validateCableRelationships())
	result.merge(inv.detectCircularLocationRefs())
	result.merge(inv.detectCircularDeviceRefs())

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
