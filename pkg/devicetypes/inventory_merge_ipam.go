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

// --- IPAM merge (VLANs, prefixes, IP addresses, VRFs) ---

// MergeVLANs merges VLANs by UUID, then by VID + Location, then inserts. The
// existing UUID is preserved on a natural-key match so re-imports update in
// place instead of creating duplicates.
func (inv *Inventory) MergeVLANs(incoming map[uuid.UUID]*CaniVLAN) {
	if inv.VLANs == nil {
		inv.VLANs = make(map[uuid.UUID]*CaniVLAN)
	}
	for id, vlan := range incoming {
		if vlan == nil {
			continue
		}
		if _, ok := inv.VLANs[id]; ok {
			inv.VLANs[id] = vlan
			continue
		}
		matched := false
		for existingID, existing := range inv.VLANs {
			if existing != nil && existing.VID == vlan.VID && existing.Location == vlan.Location {
				vlan.ID = existingID
				inv.VLANs[existingID] = vlan
				matched = true
				break
			}
		}
		if !matched {
			inv.VLANs[id] = vlan
		}
	}
}

// MergePrefixes merges prefixes by UUID, then by CIDR + VRF, then inserts.
func (inv *Inventory) MergePrefixes(incoming map[uuid.UUID]*CaniPrefix) {
	if inv.Prefixes == nil {
		inv.Prefixes = make(map[uuid.UUID]*CaniPrefix)
	}
	for id, prefix := range incoming {
		if prefix == nil {
			continue
		}
		if _, ok := inv.Prefixes[id]; ok {
			inv.Prefixes[id] = prefix
			continue
		}
		matched := false
		for existingID, existing := range inv.Prefixes {
			if existing != nil && existing.Prefix == prefix.Prefix && existing.VRF == prefix.VRF {
				prefix.ID = existingID
				inv.Prefixes[existingID] = prefix
				matched = true
				break
			}
		}
		if !matched {
			inv.Prefixes[id] = prefix
		}
	}
}

// MergeIPAddresses merges IP addresses by UUID, then by address, then inserts.
func (inv *Inventory) MergeIPAddresses(incoming map[uuid.UUID]*CaniIPAddress) {
	if inv.IPAddresses == nil {
		inv.IPAddresses = make(map[uuid.UUID]*CaniIPAddress)
	}
	for id, ip := range incoming {
		if ip == nil {
			continue
		}
		if _, ok := inv.IPAddresses[id]; ok {
			inv.IPAddresses[id] = ip
			continue
		}
		matched := false
		for existingID, existing := range inv.IPAddresses {
			if existing != nil && existing.Address == ip.Address {
				ip.ID = existingID
				inv.IPAddresses[existingID] = ip
				matched = true
				break
			}
		}
		if !matched {
			inv.IPAddresses[id] = ip
		}
	}
}

// MergeVRFs merges VRFs by UUID, then by name, then inserts.
func (inv *Inventory) MergeVRFs(incoming map[uuid.UUID]*CaniVRF) {
	if inv.VRFs == nil {
		inv.VRFs = make(map[uuid.UUID]*CaniVRF)
	}
	for id, vrf := range incoming {
		if vrf == nil || vrf.Name == "" {
			continue
		}
		if _, ok := inv.VRFs[id]; ok {
			inv.VRFs[id] = vrf
			continue
		}
		matched := false
		for existingID, existing := range inv.VRFs {
			if existing != nil && existing.Name == vrf.Name {
				vrf.ID = existingID
				inv.VRFs[existingID] = vrf
				matched = true
				break
			}
		}
		if !matched {
			inv.VRFs[id] = vrf
		}
	}
}

// --- Relationship verification ---

// RebuildDerivedState reconstructs all bidirectional parent-child references
// across the full inventory hierarchy from the authoritative forward FKs:
//
//	Location → Child Locations  (Location.Parent ↔ Location.Children)
