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

type Inventory struct {
	Devices map[uuid.UUID]*CaniDeviceType `yaml:"devices"`
}

func (inv *Inventory) Merge(new map[uuid.UUID]*CaniDeviceType) {
	if inv.Devices == nil {
		inv.Devices = make(map[uuid.UUID]*CaniDeviceType)
	}
	// Track which devices we've already processed
	processed := make(map[uuid.UUID]bool)
	// Track if any actual changes happened
	changesDetected := false

	for id, device := range new {
		// Skip nil devices or devices without names
		if device == nil || device.Name == "" {
			continue
		}

		// Case 1: Same UUID already exists - just merge properties
		if existing, ok := inv.Devices[id]; ok {
			existing.MergeProperties(device)
			processed[id] = true
			changesDetected = true
			continue
		}

		// Case 2: Look for existing device with same name
		found := false
		for _, existingDevice := range inv.Devices {
			if existingDevice != nil && existingDevice.Name == device.Name {
				// Found device with same name - merge properties but keep UUID
				changed := existingDevice.MergeProperties(device)
				processed[id] = true
				found = true
				changesDetected = changed
				break
			}
		}

		// Case 3: No match - add new device
		if !found {
			log.Printf("Adding new device %s", device.Name)
			inv.Devices[id] = device
			processed[id] = true
			changesDetected = true
		}
	}

	// After all devices have been processed, verify parent-child relationships
	// If we made no changes at all, let the user know
	if !changesDetected {
		log.Printf("No changes detected during merge")
	} else {
		// After all devices have been processed, verify parent-child relationships
		log.Printf("Changes detected during merge")
		log.Printf("Verifying parent-child relationships")
		inv.VerifyParentChildRelationships()
	}
}

// MergeProperties merges only properties from another device, preserving identity
func (d *CaniDeviceType) MergeProperties(other *CaniDeviceType) bool {
	// Don't change ID, Name, Parent or Children

	changesMade := false
	// Update basic properties
	if other.Type != d.Type {
		d.Type = other.Type
		changesMade = true
	}
	if other.DeviceTypeSlug != d.DeviceTypeSlug {
		d.DeviceTypeSlug = other.DeviceTypeSlug
		changesMade = true
	}
	if other.Vendor != d.Vendor {
		d.Vendor = other.Vendor
		changesMade = true
	}
	if other.Architecture != d.Architecture {
		d.Architecture = other.Architecture
		changesMade = true
	}
	if other.Model != d.Model {
		d.Model = other.Model
		changesMade = true
	}
	if other.Status != d.Status {
		d.Status = other.Status
		changesMade = true
	}
	// if other.LocationOrdinal != nil {
	// 	d.LocationOrdinal = other.LocationOrdinal
	// 	changesMade = true
	// }

	// Merge maps
	if d.Properties == nil {
		d.Properties = make(map[string]interface{})
	}
	if other.Properties != nil {
		for k, v := range other.Properties {
			d.Properties[k] = v
			// changesMade = true
		}
	}

	// Merge provider metadata
	if d.ProviderMetadata == nil {
		d.ProviderMetadata = make(map[string]interface{})
	}
	if other.ProviderMetadata != nil {
		for k, v := range other.ProviderMetadata {
			d.ProviderMetadata[k] = v
			// changesMade = true
		}
	}
	return changesMade
}

// Add this function to check and fix parent-child relationships
func (inv *Inventory) VerifyParentChildRelationships() {
	// For each device in the inventory
	for id, device := range inv.Devices {
		// Skip devices with no parent
		if device == nil || device.Parent == uuid.Nil {
			continue
		}

		// Get the parent device
		parentDevice, exists := inv.Devices[device.Parent]
		if !exists {
			// Parent doesn't exist, maybe log a warning
			log.Printf("Warning: device %s references non-existent parent %s",
				device.Name, device.Parent)
			continue
		}

		// Check if this device is in the parent's Children list
		found := false
		for _, childID := range parentDevice.Children {
			if childID == id {
				// log.Printf("Device %s is already a child of %s", device.Name, parentDevice.Name)
				found = true
				break
			}
		}

		// If not found, add it
		if !found {
			parentDevice.Children = append(parentDevice.Children, id)
			log.Printf("Fixed relationship: added %s as child of %s",
				device.Name, parentDevice.Name)
		}
	}
}
