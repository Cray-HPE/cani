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

import "github.com/google/uuid"

// MergeVLANs merges VLANs by UUID, then by VID + Location, then inserts.
func (inv *Inventory) MergeVLANs(incoming map[uuid.UUID]*CaniVLAN) map[uuid.UUID]uuid.UUID {
	if inv.VLANs == nil {
		inv.VLANs = make(map[uuid.UUID]*CaniVLAN)
	}
	remap := make(map[uuid.UUID]uuid.UUID, len(incoming))
	for incomingID, vlan := range incoming {
		if vlan == nil {
			continue
		}
		resolvedID := incomingID
		if _, ok := inv.VLANs[incomingID]; !ok {
			for existingID, existing := range inv.VLANs {
				if existing != nil && existing.VID == vlan.VID && existing.Location == vlan.Location {
					resolvedID = existingID
					break
				}
			}
		}
		vlan.ID = resolvedID
		inv.VLANs[resolvedID] = vlan
		remap[incomingID] = resolvedID
	}
	return remap
}

// MergePrefixes merges prefixes by UUID, then by CIDR + VRF, then inserts.
func (inv *Inventory) MergePrefixes(incoming map[uuid.UUID]*CaniPrefix) map[uuid.UUID]uuid.UUID {
	if inv.Prefixes == nil {
		inv.Prefixes = make(map[uuid.UUID]*CaniPrefix)
	}
	remap := make(map[uuid.UUID]uuid.UUID, len(incoming))
	for incomingID, prefix := range incoming {
		if prefix == nil {
			continue
		}
		resolvedID := incomingID
		if _, ok := inv.Prefixes[incomingID]; !ok {
			for existingID, existing := range inv.Prefixes {
				if existing != nil && existing.Prefix == prefix.Prefix && existing.VRF == prefix.VRF {
					resolvedID = existingID
					break
				}
			}
		}
		prefix.ID = resolvedID
		inv.Prefixes[resolvedID] = prefix
		remap[incomingID] = resolvedID
	}
	return remap
}

// MergeIPAddresses merges IP addresses by UUID, then by address, then inserts.
func (inv *Inventory) MergeIPAddresses(incoming map[uuid.UUID]*CaniIPAddress) map[uuid.UUID]uuid.UUID {
	if inv.IPAddresses == nil {
		inv.IPAddresses = make(map[uuid.UUID]*CaniIPAddress)
	}
	remap := make(map[uuid.UUID]uuid.UUID, len(incoming))
	for incomingID, address := range incoming {
		if address == nil {
			continue
		}
		resolvedID := incomingID
		if _, ok := inv.IPAddresses[incomingID]; !ok {
			for existingID, existing := range inv.IPAddresses {
				if existing != nil && existing.Address == address.Address {
					resolvedID = existingID
					break
				}
			}
		}
		address.ID = resolvedID
		inv.IPAddresses[resolvedID] = address
		remap[incomingID] = resolvedID
	}
	return remap
}

// MergeVRFs merges VRFs by UUID, then by name, then inserts.
func (inv *Inventory) MergeVRFs(incoming map[uuid.UUID]*CaniVRF) map[uuid.UUID]uuid.UUID {
	if inv.VRFs == nil {
		inv.VRFs = make(map[uuid.UUID]*CaniVRF)
	}
	remap := make(map[uuid.UUID]uuid.UUID, len(incoming))
	for incomingID, vrf := range incoming {
		if vrf == nil || vrf.Name == "" {
			continue
		}
		resolvedID := incomingID
		if _, ok := inv.VRFs[incomingID]; !ok {
			for existingID, existing := range inv.VRFs {
				if existing != nil && existing.Name == vrf.Name {
					resolvedID = existingID
					break
				}
			}
		}
		vrf.ID = resolvedID
		inv.VRFs[resolvedID] = vrf
		remap[incomingID] = resolvedID
	}
	return remap
}
