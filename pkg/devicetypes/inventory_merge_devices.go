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

	"github.com/google/uuid"
)

// MergeDevices merges new devices into the inventory by UUID match, then name
// match, then insert.
func (inv *Inventory) MergeDevices(incoming map[uuid.UUID]*CaniDeviceType) {
	inv.MergeDevicesStrict(incoming, false)
}

// MergeDevicesStrict behaves like MergeDevices but when strict is true it
// skips any device whose Slug and Model are both empty.
func (inv *Inventory) MergeDevicesStrict(incoming map[uuid.UUID]*CaniDeviceType, strict bool) (map[uuid.UUID]uuid.UUID, []UnclassifiedDevice) {
	if inv.Devices == nil {
		inv.Devices = make(map[uuid.UUID]*CaniDeviceType)
	}

	remap := make(map[uuid.UUID]uuid.UUID)
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
		if strict && device.Slug == "" && device.Model == "" {
			skipped = append(skipped, UnclassifiedDevice{
				ID: id, Name: device.Name, DeviceType: string(device.Type),
				Model: device.Model, Manufacturer: device.Manufacturer,
			})
			continue
		}
		if existing, ok := inv.Devices[id]; ok {
			inv.unindexDevice(id, existing)
			if existing.MergeProperties(device) {
				changesDetected = true
			}
			inv.indexDevice(id, existing)
			remap[id] = id
			continue
		}
		if matchedID, changed, matched := inv.mergeByName(device); matched {
			changesDetected = changesDetected || changed
			remap[id] = matchedID
			continue
		}
		inv.Devices[id] = device
		inv.indexDevice(id, device)
		remap[id] = id
		changesDetected = true
	}

	if changesDetected {
		log.Printf("Changes detected during merge")
	}
	return remap, skipped
}

func (inv *Inventory) mergeByName(device *CaniDeviceType) (uuid.UUID, bool, bool) {
	for id, existing := range inv.Devices {
		if existing == nil || existing.Name != device.Name || !providerIdentityCompatible(existing, device) {
			continue
		}
		inv.unindexDevice(id, existing)
		changed := existing.MergeProperties(device)
		inv.indexDevice(id, existing)
		return id, changed, true
	}
	return uuid.Nil, false, false
}

var providerIdentityKeys = []string{"bmc_fqdn", "bmc_hostname", "xname"}

func providerIdentityCompatible(left, right *CaniDeviceType) bool {
	for _, key := range providerIdentityKeys {
		leftValue, leftOK := left.GetProviderMeta(key)
		rightValue, rightOK := right.GetProviderMeta(key)
		if leftOK && rightOK {
			return fmt.Sprintf("%v", leftValue) == fmt.Sprintf("%v", rightValue)
		}
	}
	return true
}

// FindDeviceByProviderKey finds a device by one provider metadata value.
func (inv *Inventory) FindDeviceByProviderKey(provider, key string, value any) *CaniDeviceType {
	if inv == nil || inv.Devices == nil || value == nil || value == "" {
		return nil
	}
	if id := inv.lookupProviderKey(provider, key, value); id != uuid.Nil {
		if device, ok := inv.Devices[id]; ok {
			return device
		}
	}
	for _, device := range inv.Devices {
		if device == nil {
			continue
		}
		metadata, ok := device.GetProviderSubMap(provider)
		if !ok {
			continue
		}
		if candidate, ok := metadata[key]; ok && fmt.Sprintf("%v", candidate) == fmt.Sprintf("%v", value) {
			return device
		}
	}
	return nil
}

// ProviderKeyCheck pairs a metadata key with a value for lookup.
type ProviderKeyCheck struct {
	Key   string
	Value any
}

// FindDeviceByProviderKeys returns the first device matching any key/value pair.
func (inv *Inventory) FindDeviceByProviderKeys(provider string, checks []ProviderKeyCheck) *CaniDeviceType {
	for _, check := range checks {
		if device := inv.FindDeviceByProviderKey(provider, check.Key, check.Value); device != nil {
			return device
		}
	}
	return nil
}
