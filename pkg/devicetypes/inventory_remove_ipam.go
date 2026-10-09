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
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ErrPrefixHoldsAddresses marks a refused removal of a root prefix that still
// holds addresses; RemovePrefixWithAddresses removes them with the prefix.
var ErrPrefixHoldsAddresses = errors.New("remove them first")

// RemovePrefix deletes a prefix and moves its child prefixes and addresses up
// to the prefix's own parent, as Nautobot does on delete. Like Nautobot, it
// refuses to delete a root prefix that still holds addresses, because every
// address needs a containing prefix; a parent record that is missing counts
// as none.
func (inv *Inventory) RemovePrefix(id uuid.UUID) error {
	prefix := inv.Prefixes[id]
	if prefix == nil {
		return fmt.Errorf("prefix %s not found", id)
	}
	parent := inv.parentPrefixID(prefix)
	addresses := inv.prefixAddresses(id)
	if parent == uuid.Nil && len(addresses) > 0 {
		return fmt.Errorf("prefix %s (%s) holds %d IP address(es) that would have no parent prefix: %w",
			prefix.Prefix, id, len(addresses), ErrPrefixHoldsAddresses)
	}
	for _, child := range inv.Prefixes {
		if child != nil && child.Parent == id {
			child.Parent = parent
		}
	}
	for _, address := range addresses {
		address.Parent = parent
	}
	delete(inv.Prefixes, id)
	return nil
}

// RemovePrefixWithAddresses deletes a prefix like RemovePrefix, but when the
// prefix has no parent to move its addresses to, it removes them first rather
// than refusing. It returns how many addresses it removed.
func (inv *Inventory) RemovePrefixWithAddresses(id uuid.UUID) (int, error) {
	prefix := inv.Prefixes[id]
	if prefix == nil {
		return 0, fmt.Errorf("prefix %s not found", id)
	}
	var removed int
	if inv.parentPrefixID(prefix) == uuid.Nil {
		for _, address := range inv.prefixAddresses(id) {
			if inv.removeIPAddress(address.ID) == nil {
				removed++
			}
		}
	}
	if removed > 0 {
		inv.rebuildInterfaceIPAddresses()
	}
	return removed, inv.RemovePrefix(id)
}

// parentPrefixID returns the prefix's parent, or uuid.Nil when that record is
// missing, so a dangling parent counts as none.
func (inv *Inventory) parentPrefixID(prefix *CaniPrefix) uuid.UUID {
	if inv.Prefixes[prefix.Parent] == nil {
		return uuid.Nil
	}
	return prefix.Parent
}

func (inv *Inventory) prefixAddresses(id uuid.UUID) []*CaniIPAddress {
	var held []*CaniIPAddress
	for _, address := range inv.IPAddresses {
		if address != nil && address.Parent == id {
			held = append(held, address)
		}
	}
	return held
}

// RemoveIPAddress deletes an address and clears the device primary IPs and
// NAT inside references that pointed at it, as Nautobot does on delete. Its
// interface assignments go with it; the derived interface index is rebuilt.
func (inv *Inventory) RemoveIPAddress(id uuid.UUID) error {
	if err := inv.removeIPAddress(id); err != nil {
		return err
	}
	inv.rebuildInterfaceIPAddresses()
	return nil
}

// removeIPAddress deletes the address and the forward references to it; the
// caller rebuilds the derived interface index once it has finished removing.
func (inv *Inventory) removeIPAddress(id uuid.UUID) error {
	if inv.IPAddresses[id] == nil {
		return fmt.Errorf("ip address %s not found", id)
	}
	delete(inv.IPAddresses, id)
	inv.clearPrimaryIPs(id)
	for _, address := range inv.IPAddresses {
		if address != nil {
			address.NATInside = clearReference(address.NATInside, id)
		}
	}
	return nil
}

func (inv *Inventory) clearPrimaryIPs(id uuid.UUID) {
	for _, device := range inv.Devices {
		if device != nil {
			device.PrimaryIPv4 = clearReference(device.PrimaryIPv4, id)
			device.PrimaryIPv6 = clearReference(device.PrimaryIPv6, id)
		}
	}
}

// RemoveVRF deletes a VRF and its prefix memberships. Its device assignments
// are held on the VRF and go with it. Interface VRF settings are names that
// export resolves in the target system, so they are left as they are.
func (inv *Inventory) RemoveVRF(id uuid.UUID) error {
	if inv.VRFs[id] == nil {
		return fmt.Errorf("vrf %s not found", id)
	}
	delete(inv.VRFs, id)
	for _, prefix := range inv.Prefixes {
		if prefix != nil {
			prefix.VRFs = removeUUID(prefix.VRFs, id)
		}
	}
	return nil
}

func clearReference(ref, removed uuid.UUID) uuid.UUID {
	if ref == removed {
		return uuid.Nil
	}
	return ref
}
