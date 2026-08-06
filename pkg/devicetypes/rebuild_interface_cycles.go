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

// rebuildInterfaceRelationships clears and rebuilds Inventory.Interfaces
// from each device's and module's embedded Interfaces slices,
// reporting only newly indexed interfaces.
func (inv *Inventory) rebuildInterfaceRelationships() *RelationshipResult {
	res := &RelationshipResult{}

	// Snapshot existing interface IDs for change detection.
	oldIfaces := make(map[uuid.UUID]bool, len(inv.Interfaces))
	for id := range inv.Interfaces {
		oldIfaces[id] = true
	}

	inv.Interfaces = make(map[uuid.UUID]*CaniInterface)

	for deviceID, device := range inv.Devices {
		if device == nil {
			continue
		}
		inv.indexInterfaceSpecs(device.Interfaces, deviceID, "device", device.Name, oldIfaces, res)
	}

	for _, mod := range inv.Modules {
		if mod == nil {
			continue
		}
		inv.indexInterfaceSpecs(mod.Interfaces, mod.ParentDevice, "module", mod.Name, oldIfaces, res)
	}

	return res
}

func (inv *Inventory) indexInterfaceSpecs(interfaces []InterfaceSpec, deviceID uuid.UUID,
	ownerKind, ownerName string, oldIfaces map[uuid.UUID]bool, result *RelationshipResult) {
	for i := range interfaces {
		iface := &interfaces[i]
		if iface.ID == uuid.Nil {
			iface.ID = uuid.New()
		}
		inv.Interfaces[iface.ID] = interfaceInstanceFromSpec(iface, deviceID)
		if !oldIfaces[iface.ID] {
			result.Fixed = append(result.Fixed,
				fmt.Sprintf("interface %q (%s) indexed from %s %q",
					iface.Name, iface.ID, ownerKind, ownerName))
		}
	}
}

func interfaceInstanceFromSpec(iface *InterfaceSpec, deviceID uuid.UUID) *CaniInterface {
	mgmtOnly := iface.MgmtOnly != nil && *iface.MgmtOnly
	role := ResolveInterfaceRole(iface.Role, iface.Name, iface.Type, mgmtOnly)
	return &CaniInterface{
		ID:            iface.ID,
		Name:          iface.Name,
		InterfaceType: iface.Type,
		DeviceID:      deviceID,
		ObjectMeta: ObjectMeta{
			Status: string(StatusActive),
			Role:   role,
			Tags:   append([]string(nil), iface.Tags...),
		},
		MgmtOnly:       mgmtOnly,
		Label:          iface.Label,
		MacAddress:     iface.MacAddress,
		Lag:            iface.Lag,
		Mode:           iface.Mode,
		UntaggedVLAN:   iface.UntaggedVLAN,
		TaggedVLANs:    append([]int(nil), iface.TaggedVLANs...),
		VRF:            iface.VRF,
		ConnectedCable: iface.ConnectedCable,
	}
}

// detectCircularLocationRefs walks location parent chains to find cycles.
func (inv *Inventory) detectCircularLocationRefs() *RelationshipResult {
	res := &RelationshipResult{}
	for id, loc := range inv.Locations {
		if loc == nil {
			continue
		}
		if inv.hasLocationCycle(id) {
			res.Errors = append(res.Errors,
				fmt.Errorf("location %q (%s): circular parent reference detected",
					loc.Name, id))
		}
	}
	return res
}

// hasLocationCycle returns true if following Parent pointers from id
// leads back to id.
func (inv *Inventory) hasLocationCycle(id uuid.UUID) bool {
	visited := map[uuid.UUID]bool{}
	cur := id
	for {
		loc, ok := inv.Locations[cur]
		if !ok || loc == nil || loc.Parent == uuid.Nil {
			return false
		}
		if visited[cur] {
			return true
		}
		visited[cur] = true
		cur = loc.Parent
	}
}

// detectCircularDeviceRefs walks device parent chains to find cycles.
func (inv *Inventory) detectCircularDeviceRefs() *RelationshipResult {
	res := &RelationshipResult{}
	for id, device := range inv.Devices {
		if device == nil {
			continue
		}
		if inv.hasDeviceCycle(id) {
			res.Errors = append(res.Errors,
				fmt.Errorf("device %q (%s): circular parent reference detected",
					device.Name, id))
		}
	}
	return res
}

// hasDeviceCycle returns true if following device Parent pointers from
// id leads back to id (only follows device→device links).
func (inv *Inventory) hasDeviceCycle(id uuid.UUID) bool {
	visited := map[uuid.UUID]bool{}
	cur := id
	for {
		device, ok := inv.Devices[cur]
		if !ok || device == nil || device.Parent == uuid.Nil {
			return false
		}
		// Stop if parent is a rack (not a device cycle).
		if _, isRack := inv.Racks[device.Parent]; isRack {
			return false
		}
		if visited[cur] {
			return true
		}
		visited[cur] = true
		cur = device.Parent
	}
}
