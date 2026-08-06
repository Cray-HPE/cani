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

// AddDevices inserts new devices into the inventory.  Returns an error
// if any UUID already exists.
func (inv *Inventory) AddDevices(batch map[uuid.UUID]*CaniDeviceType) error {
	for id, device := range batch {
		if device == nil {
			return fmt.Errorf("device %s must not be nil", id)
		}
		if err := device.Validate(); err != nil {
			return err
		}
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
		if err := device.Validate(); err != nil {
			log.Printf("Skipping invalid device %q: %v", device.Name, err)
			continue
		}

		// In strict mode, reject devices without a slug or model.
		if strict && device.Slug == "" && device.Model == "" {
			skipped = append(skipped, UnclassifiedDevice{
				ID:           id,
				Name:         device.Name,
				DeviceType:   string(device.Type),
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
		if matched, changed := inv.mergeByName(device); matched {
			if changed {
				changesDetected = true
			}
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
// properties. It returns whether a compatible device matched and whether that
// merge changed fields. If the incoming device carries provider metadata with a
// distinguishing key (e.g. bmc_fqdn, bmc_hostname) it must also match
// the existing device; otherwise two servers of the same model but with
// different BMC identities would be collapsed into one.
func (inv *Inventory) mergeByName(device *CaniDeviceType) (bool, bool) {
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
		return true, changed
	}
	return false, false
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
