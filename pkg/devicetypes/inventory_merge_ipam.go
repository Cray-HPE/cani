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
	"maps"
	"slices"

	"github.com/google/uuid"
)

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

// MergePrefixes merges prefixes by UUID, then by shared external ID, then by
// namespace + canonical CIDR, then inserts. A natural-key match whose source
// identity conflicts is an ErrSourceIdentityConflict naming the record to
// remove, and a key ambiguous among legacy duplicates is an error; neither is
// a silent overwrite. A prefix the merge inserts adopts, as AddPrefix does,
// the pre-existing records it holds more closely than their current parent;
// the records the batch itself carries keep the parent the provider set or
// the merge derived.
func (inv *Inventory) MergePrefixes(incoming map[uuid.UUID]*CaniPrefix) (map[uuid.UUID]uuid.UUID, error) {
	if inv.Prefixes == nil {
		inv.Prefixes = make(map[uuid.UUID]*CaniPrefix)
	}
	remap, err := resolveIdentities(IPAMKindPrefix, inv.Prefixes, incoming, prefixNaturalKey, prefixExternalIDs,
		func(p *CaniPrefix) string {
			return fmt.Sprintf("%s in namespace %q", p.Prefix, p.EffectiveNamespace())
		})
	if err != nil {
		return nil, err
	}
	inserted := insertedRecords(inv.Prefixes, incoming, remap)
	untouched := recordsOutside(inv.Prefixes, remap)
	addresses := slices.Collect(maps.Values(inv.IPAddresses)) // addresses merge after prefixes, so none is the batch's yet
	applyIdentities(inv.Prefixes, incoming, remap, func(p *CaniPrefix, id uuid.UUID) { p.ID = id })
	for _, id := range inserted {
		inv.adoptHeld(inv.Prefixes[id], untouched, addresses)
	}
	return remap, nil
}

// insertedRecords returns, once each and in incoming-UUID order, the resolved
// UUIDs that are not records yet, so a merge can tell an insert from an update.
func insertedRecords[T any](existing, incoming map[uuid.UUID]*T, remap map[uuid.UUID]uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]bool)
	var inserted []uuid.UUID
	for _, id := range sortedIDs(incoming) {
		target, ok := remap[id]
		if !ok || seen[target] || existing[target] != nil {
			continue
		}
		seen[target] = true
		inserted = append(inserted, target)
	}
	return inserted
}

// recordsOutside returns the records no incoming object resolves to: the
// ones a merge leaves alone.
func recordsOutside[T any](existing map[uuid.UUID]*T, remap map[uuid.UUID]uuid.UUID) []*T {
	targets := make(map[uuid.UUID]bool, len(remap))
	for _, target := range remap {
		targets[target] = true
	}
	var outside []*T
	for id, record := range existing {
		if record != nil && !targets[id] {
			outside = append(outside, record)
		}
	}
	return outside
}

// MergeIPAddresses merges addresses by UUID, then by shared external ID, then
// by namespace + canonical host (the mask is not part of identity), then
// inserts. The namespace of a parented address comes from its parent prefix,
// so callers must remap and merge prefixes first.
func (inv *Inventory) MergeIPAddresses(incoming map[uuid.UUID]*CaniIPAddress) (map[uuid.UUID]uuid.UUID, error) {
	if inv.IPAddresses == nil {
		inv.IPAddresses = make(map[uuid.UUID]*CaniIPAddress)
	}
	keyOf := func(addr *CaniIPAddress) ipAddressKey { return ipAddressNaturalKey(addr, inv.IPAddressNamespace(addr)) }
	remap, err := resolveIdentities(IPAMKindIPAddress, inv.IPAddresses, incoming, keyOf, ipAddressExternalIDs,
		func(addr *CaniIPAddress) string {
			key := keyOf(addr)
			return fmt.Sprintf("%s in namespace %q", key.Host, key.Namespace)
		})
	if err != nil {
		return nil, err
	}
	applyIdentities(inv.IPAddresses, incoming, remap, func(a *CaniIPAddress, id uuid.UUID) { a.ID = id })
	return remap, nil
}

// MergeVRFs merges VRFs by UUID, then by shared external ID, then by
// namespace + name, then inserts; a VRF without a name is skipped. Same-named
// VRFs in one namespace make a name match ambiguous and are reported instead
// of resolved to the first.
func (inv *Inventory) MergeVRFs(incoming map[uuid.UUID]*CaniVRF) (map[uuid.UUID]uuid.UUID, error) {
	if inv.VRFs == nil {
		inv.VRFs = make(map[uuid.UUID]*CaniVRF)
	}
	named := make(map[uuid.UUID]*CaniVRF, len(incoming))
	for id, vrf := range incoming {
		if vrf != nil && vrf.Name != "" {
			named[id] = vrf
		}
	}
	remap, err := resolveIdentities(IPAMKindVRF, inv.VRFs, named, vrfNaturalKey, vrfExternalIDs,
		func(v *CaniVRF) string { return fmt.Sprintf("%s in namespace %q", v.Name, v.EffectiveNamespace()) })
	if err != nil {
		return nil, err
	}
	applyIdentities(inv.VRFs, named, remap, func(v *CaniVRF, id uuid.UUID) { v.ID = id })
	return remap, nil
}

func prefixExternalIDs(p *CaniPrefix) map[string]uuid.UUID       { return p.ExternalIDs }
func ipAddressExternalIDs(a *CaniIPAddress) map[string]uuid.UUID { return a.ExternalIDs }
func vrfExternalIDs(v *CaniVRF) map[string]uuid.UUID             { return v.ExternalIDs }
