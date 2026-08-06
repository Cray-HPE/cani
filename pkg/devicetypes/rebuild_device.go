/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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

	"github.com/google/uuid"
)

// clearDeviceReverseLists resets all derived fields on racks and devices
// so they can be rebuilt from Parent fields.
func (inv *Inventory) clearDeviceReverseLists() {
	for _, rack := range inv.Racks {
		if rack != nil {
			rack.Devices = nil
			rack.OccupiedSlots = nil
		}
	}
	for _, device := range inv.Devices {
		if device != nil {
			device.Children = nil
			device.Rack = uuid.Nil
			device.Location = uuid.Nil
			device.ParentDevice = uuid.Nil
		}
	}
}

// linkDeviceToRack links a device to its parent rack and returns a fix message.
// Returns true if the parent was a rack.
func (inv *Inventory) linkDeviceToRack(id uuid.UUID, device *CaniDeviceType) (bool, string) {
	rack, ok := inv.Racks[device.Parent]
	if !ok {
		return false, ""
	}
	rack.addDevice(id)
	device.Rack = device.Parent
	device.Location = rack.Location

	// Rebuild OccupiedSlots from the device's stored position.
	if device.RackPosition > 0 {
		height := device.UHeight
		if height < 1 {
			height = 1
		}
		rack.PlaceDevice(id, device.RackPosition, height, device.Face, device.IsFullDepth)
	}

	return true, fmt.Sprintf("device %q added to rack %q", device.Name, rack.Name)
}

// linkDeviceToParentDevice links a device to its parent device and returns a fix message.
// Returns true if the parent was a device.
func (inv *Inventory) linkDeviceToParentDevice(id uuid.UUID, device *CaniDeviceType) (bool, string) {
	parent, ok := inv.Devices[device.Parent]
	if !ok {
		return false, ""
	}
	if !containsUUID(parent.Children, id) {
		parent.Children = append(parent.Children, id)
	}
	device.ParentDevice = device.Parent
	device.Rack = parent.Rack
	device.Location = parent.Location

	// Also register with the rack so rack views can enumerate all
	// mounted devices, not just direct rack children.
	if rack, ok := inv.Racks[parent.Rack]; ok {
		rack.addDevice(id)
	}

	return true, fmt.Sprintf("device %q added as child of device %q", device.Name, parent.Name)
}

// rebuildDeviceRelationships clears and rebuilds Rack.Devices,
// Device.Children, and explicit FK fields (Rack, Location, ParentDevice)
// from each device's Parent field.
//
// The rebuild uses a multi-phase approach so that parent devices are
// always resolved before their children, regardless of map iteration
// order, making the result deterministic and idempotent.
func (inv *Inventory) rebuildDeviceRelationships() *RelationshipResult {
	res := &RelationshipResult{}

	// Snapshot derived fields for change detection.
	type deviceLink struct {
		rack, location, parentDevice uuid.UUID
	}
	old := make(map[uuid.UUID]deviceLink)
	for id, d := range inv.Devices {
		if d != nil {
			old[id] = deviceLink{d.Rack, d.Location, d.ParentDevice}
		}
	}

	inv.clearDeviceReverseLists()

	// Phase 1: link devices whose parent is a rack.
	for id, device := range inv.Devices {
		if device == nil || device.Parent == uuid.Nil {
			continue
		}
		inv.linkDeviceToRack(id, device)
	}

	// Phase 2: iteratively link devices whose parent is another device.
	// Repeats until no progress is made, handling arbitrary nesting
	// depth regardless of map iteration order.
	for {
		progress := false
		for id, device := range inv.Devices {
			if device == nil || device.Parent == uuid.Nil {
				continue
			}
			if inv.deviceAlreadyLinked(device) {
				continue
			}
			if _, isRack := inv.Racks[device.Parent]; isRack {
				continue
			}
			parent, ok := inv.Devices[device.Parent]
			if !ok || parent == nil {
				continue
			}
			if !inv.deviceAlreadyLinked(parent) {
				continue
			}
			if linked, _ := inv.linkDeviceToParentDevice(id, device); linked {
				progress = true
			}
		}
		if !progress {
			break
		}
	}

	// Phase 2b: link devices whose parent is an orphan device (no rack
	// ancestor). This handles device-bay hierarchies added without a
	// rack placement — children still need ParentDevice set so they
	// are not reported as errors.
	for {
		progress := false
		for id, device := range inv.Devices {
			if device == nil || device.Parent == uuid.Nil {
				continue
			}
			if inv.deviceAlreadyLinked(device) {
				continue
			}
			parent, ok := inv.Devices[device.Parent]
			if !ok || parent == nil {
				continue
			}
			device.ParentDevice = device.Parent
			if !containsUUID(parent.Children, id) {
				parent.Children = append(parent.Children, id)
			}
			progress = true
		}
		if !progress {
			break
		}
	}

	// Phase 3: report only actual changes and problems.
	for id, device := range inv.Devices {
		if device == nil {
			continue
		}
		if device.Parent == uuid.Nil {
			res.Warnings = append(res.Warnings,
				fmt.Sprintf("device %q (%s): no parent assigned (orphan)",
					device.Name, id))
			res.Orphans = append(res.Orphans, OrphanItem{
				ID:               device.ID,
				Name:             device.Name,
				Kind:             "device",
				DeviceType:       string(device.Type),
				Model:            device.Model,
				Manufacturer:     device.Manufacturer,
				ProviderMetadata: device.ProviderMetadata,
			})
			continue
		}
		if !inv.deviceAlreadyLinked(device) {
			res.Errors = append(res.Errors,
				fmt.Errorf("device %q (%s): parent %s not found in racks or devices",
					device.Name, id, device.Parent))
			continue
		}
		prev := old[id]
		if device.Rack == prev.rack && device.Location == prev.location &&
			device.ParentDevice == prev.parentDevice {
			continue
		}
		if device.ParentDevice != uuid.Nil {
			parent := inv.Devices[device.ParentDevice]
			if parent != nil {
				res.Fixed = append(res.Fixed,
					fmt.Sprintf("device %q added as child of device %q",
						device.Name, parent.Name))
			}
		} else if device.Rack != uuid.Nil {
			rack := inv.Racks[device.Rack]
			if rack != nil {
				res.Fixed = append(res.Fixed,
					fmt.Sprintf("device %q added to rack %q",
						device.Name, rack.Name))
			}
		}
	}

	return res
}

// deviceAlreadyLinked returns true when a device has been resolved to
// a rack (directly or through an ancestor device).
func (inv *Inventory) deviceAlreadyLinked(device *CaniDeviceType) bool {
	return device.Rack != uuid.Nil || device.ParentDevice != uuid.Nil
}
