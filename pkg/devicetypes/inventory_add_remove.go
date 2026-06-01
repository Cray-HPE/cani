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

// AddModule inserts a single module into the inventory.
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
	inv.Modules[mod.ID] = mod
	inv.VerifyParentChildRelationships()
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
	inv.VerifyParentChildRelationships()
	return nil
}

// RemoveLocation deletes a location and unlinks it from parent/children.
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
	// Unlink from parent
	if loc.Parent != uuid.Nil {
		if parent, ok := inv.Locations[loc.Parent]; ok {
			parent.Children = removeUUID(parent.Children, id)
		}
	}
	delete(inv.Locations, id)
	return nil
}

// RemoveRack deletes a rack and moves any contained devices to orphaned state.
func (inv *Inventory) RemoveRack(id uuid.UUID) error {
	rack, exists := inv.Racks[id]
	if !exists {
		return fmt.Errorf("rack %s not found", id)
	}
	// Unlink devices from rack
	for _, deviceID := range rack.Devices {
		if device, ok := inv.Devices[deviceID]; ok {
			device.Parent = uuid.Nil
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
	mod, exists := inv.Modules[id]
	if !exists {
		return fmt.Errorf("module %s not found", id)
	}
	// Remove cables referencing this module's parent device
	_ = mod // module has ParentDevice but cables ref devices, not modules
	delete(inv.Modules, id)
	return nil
}

// RemoveCable deletes a cable from the inventory.
func (inv *Inventory) RemoveCable(id uuid.UUID) error {
	if _, exists := inv.Cables[id]; !exists {
		return fmt.Errorf("cable %s not found", id)
	}
	delete(inv.Cables, id)
	return nil
}

// AddVLAN inserts a single VLAN into the inventory.
func (inv *Inventory) AddVLAN(vlan *CaniVLAN) error {
	if vlan == nil {
		return fmt.Errorf("vlan must not be nil")
	}
	if _, exists := inv.VLANs[vlan.ID]; exists {
		return fmt.Errorf("vlan %s already exists", vlan.ID)
	}
	inv.VLANs[vlan.ID] = vlan
	return nil
}

// AddPrefix inserts a single prefix into the inventory and auto-computes its parent.
func (inv *Inventory) AddPrefix(prefix *CaniPrefix) error {
	if prefix == nil {
		return fmt.Errorf("prefix must not be nil")
	}
	if _, exists := inv.Prefixes[prefix.ID]; exists {
		return fmt.Errorf("prefix %s already exists", prefix.ID)
	}
	if err := ParsePrefix(prefix); err != nil {
		return fmt.Errorf("invalid prefix: %w", err)
	}
	if prefix.Parent == uuid.Nil {
		prefix.Parent = FindParentPrefix(prefix, inv.Prefixes)
	}
	inv.Prefixes[prefix.ID] = prefix
	return nil
}

// AddIPAddress inserts a single IP address into the inventory and auto-computes its parent prefix.
func (inv *Inventory) AddIPAddress(addr *CaniIPAddress) error {
	if addr == nil {
		return fmt.Errorf("ip address must not be nil")
	}
	if _, exists := inv.IPAddresses[addr.ID]; exists {
		return fmt.Errorf("ip address %s already exists", addr.ID)
	}
	if err := ParseIPAddress(addr); err != nil {
		return fmt.Errorf("invalid ip address: %w", err)
	}
	if addr.Parent == uuid.Nil {
		addr.Parent = FindParentPrefixForIP(addr, inv.Prefixes)
	}
	inv.IPAddresses[addr.ID] = addr
	return nil
}
