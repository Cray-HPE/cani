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

// AddLocation inserts a single location into the inventory.
func (inv *Inventory) AddLocation(loc *CaniLocationType) error {
	if loc == nil {
		return fmt.Errorf("location must not be nil")
	}
	if _, exists := inv.Locations[loc.ID]; exists {
		return fmt.Errorf("location %s already exists", loc.ID)
	}
	inv.Locations[loc.ID] = loc
	inv.VerifyParentChildRelationships()
	return nil
}

// AddRack inserts a single rack into the inventory.
func (inv *Inventory) AddRack(rack *CaniRackType) error {
	if rack == nil {
		return fmt.Errorf("rack must not be nil")
	}
	if _, exists := inv.Racks[rack.ID]; exists {
		return fmt.Errorf("rack %s already exists", rack.ID)
	}
	if rack.Location != uuid.Nil {
		if loc, ok := inv.Locations[rack.Location]; ok {
			if err := loc.ValidateContentType("rack"); err != nil {
				return err
			}
		}
	}
	inv.Racks[rack.ID] = rack
	inv.VerifyParentChildRelationships()
	return nil
}

// AddModule inserts a single module into the inventory. A module whose
// placement the inventory rejects (an undeclared or occupied bay, a location
// that disagrees with its parent's, or a spare at a location that does not
// exist or admit modules) is not added, and the error is returned.
func (inv *Inventory) AddModule(mod *CaniModuleType) error {
	if mod == nil {
		return fmt.Errorf("module must not be nil")
	}
	if err := mod.Validate(); err != nil {
		return err
	}
	if _, exists := inv.Modules[mod.ID]; exists {
		return fmt.Errorf("module %s already exists", mod.ID)
	}
	if loc := inv.Locations[mod.Location]; mod.IsSpare() && loc != nil {
		if err := loc.ValidateContentType("module"); err != nil {
			return err
		}
	}
	inv.ResolveModuleNames(mod)
	inv.Modules[mod.ID] = mod
	result := inv.VerifyParentChildRelationships()
	if err := result.errorsAbout(mod.ID); err != nil {
		delete(inv.Modules, mod.ID)
		inv.RebuildDerivedState()
		return err
	}
	return nil
}

// AddCable inserts a single cable into the inventory.
func (inv *Inventory) AddCable(cable *CaniCableType) error {
	if cable == nil {
		return fmt.Errorf("cable must not be nil")
	}
	if _, exists := inv.Cables[cable.ID]; exists {
		return fmt.Errorf("cable %s already exists", cable.ID)
	}
	inv.Cables[cable.ID] = cable
	result := inv.VerifyParentChildRelationships()
	if err := result.Err(); err != nil {
		delete(inv.Cables, cable.ID)
		inv.RebuildDerivedState()
		return err
	}
	return nil
}

// RemoveLocation deletes a location and unlinks it from parent/children.
// Racks, child locations, devices placed directly in the location and spare
// modules stored there block the removal, as they would in Nautobot.
func (inv *Inventory) RemoveLocation(id uuid.UUID) error {
	loc, exists := inv.Locations[id]
	if !exists {
		return fmt.Errorf("location %s not found", id)
	}
	if len(loc.Racks) > 0 {
		return fmt.Errorf("location %s still has %d rack(s); remove them first", id, len(loc.Racks))
	}
	if len(loc.Children) > 0 {
		return fmt.Errorf("location %s still has %d child location(s); remove them first", id, len(loc.Children))
	}
	if n := inv.countDevicesPlacedIn(id); n > 0 {
		return fmt.Errorf("location %s still has %d device(s); remove them first", id, n)
	}
	if n := inv.countSparesStoredIn(id); n > 0 {
		return fmt.Errorf("location %s still has %d spare module(s); remove or install them first", id, n)
	}
	// Unlink from parent
	if loc.Parent != uuid.Nil {
		if parent, ok := inv.Locations[loc.Parent]; ok {
			parent.Children = removeUUID(parent.Children, id)
		}
	}
	delete(inv.Locations, id)
	return nil
}

// countDevicesPlacedIn counts the devices placed directly in a location,
// with no rack or parent device between them; a racked device is reached
// through its rack, which blocks the removal on its own.
func (inv *Inventory) countDevicesPlacedIn(id uuid.UUID) int {
	n := 0
	for _, device := range inv.Devices {
		if device != nil && device.Parent == uuid.Nil && device.Location == id {
			n++
		}
	}
	return n
}

// countSparesStoredIn counts the modules stored at a location rather than
// installed in a device; Nautobot protects such a location from deletion.
func (inv *Inventory) countSparesStoredIn(id uuid.UUID) int {
	n := 0
	for _, mod := range inv.Modules {
		if mod.IsSpare() && mod.Location == id {
			n++
		}
	}
	return n
}

// RemoveRack deletes a rack. The devices it held keep the rack's location and
// lose the rack, as a device unracked in Nautobot does; they are orphans only
// when the rack had no location to leave them in.
func (inv *Inventory) RemoveRack(id uuid.UUID) error {
	rack, exists := inv.Racks[id]
	if !exists {
		return fmt.Errorf("rack %s not found", id)
	}
	// Unlink devices from rack
	for _, deviceID := range rack.Devices {
		if device, ok := inv.Devices[deviceID]; ok {
			device.Parent = uuid.Nil
			device.Location = rack.Location
		}
	}
	// Unlink from location
	if rack.Location != uuid.Nil {
		if loc, ok := inv.Locations[rack.Location]; ok {
			loc.Racks = removeUUID(loc.Racks, id)
		}
	}
	delete(inv.Racks, id)
	return nil
}

// RemoveModule deletes a module from the inventory.
func (inv *Inventory) RemoveModule(id uuid.UUID) error {
	if inv.Modules[id] == nil {
		return fmt.Errorf("module %s not found", id)
	}
	inv.removeModule(id)
	return inv.RebuildDerivedState().Err()
}

// RemoveCable deletes a cable from the inventory.
func (inv *Inventory) RemoveCable(id uuid.UUID) error {
	if _, exists := inv.Cables[id]; !exists {
		return fmt.Errorf("cable %s not found", id)
	}
	delete(inv.Cables, id)
	inv.unlinkCableReferences(id)
	return nil
}

// AddInterface inserts a standalone interface into the inventory.
// The interface's DeviceID must reference an existing device or module.
func (inv *Inventory) AddInterface(iface *CaniInterface) error {
	if iface == nil {
		return fmt.Errorf("interface must not be nil")
	}
	if _, exists := inv.Interfaces[iface.ID]; exists {
		return fmt.Errorf("interface %s already exists", iface.ID)
	}
	if iface.DeviceID == uuid.Nil {
		return fmt.Errorf("interface device ID must not be nil")
	}
	dev, devOK := inv.Devices[iface.DeviceID]
	mod, modOK := inv.Modules[iface.DeviceID]
	if !devOK && !modOK {
		return fmt.Errorf("device or module %s not found", iface.DeviceID)
	}
	for _, existing := range inv.Interfaces {
		if existing.DeviceID == iface.DeviceID && existing.Name == iface.Name {
			return fmt.Errorf("interface %q already exists on device %s", iface.Name, iface.DeviceID)
		}
	}

	mgmtOnly := iface.MgmtOnly
	spec := InterfaceSpec{
		ID:               iface.ID,
		Name:             iface.Name,
		Type:             iface.InterfaceType,
		Label:            iface.Label,
		Role:             iface.Role,
		MacAddress:       iface.MacAddress,
		MgmtOnly:         &mgmtOnly,
		Tags:             iface.Tags,
		Lag:              iface.Lag,
		Mode:             iface.Mode,
		UntaggedVLAN:     iface.UntaggedVLAN,
		TaggedVLANs:      iface.TaggedVLANs,
		VRF:              iface.VRF,
		Description:      iface.Description,
		Status:           iface.Status,
		Tenant:           iface.Tenant,
		CustomFields:     iface.CustomFields,
		ExternalIDs:      iface.ExternalIDs,
		ContentType:      iface.ContentType,
		ConnectedCable:   iface.ConnectedCable,
		ProviderMetadata: iface.ProviderMetadata,
	}
	if devOK {
		dev.Interfaces = append(dev.Interfaces, spec)
	} else {
		mod.Interfaces = append(mod.Interfaces, spec)
	}

	inv.Interfaces[iface.ID] = iface
	return nil
}
