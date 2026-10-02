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
// identity conflicts, or that is ambiguous among legacy duplicates, is an
// error rather than a silent overwrite.
func (inv *Inventory) MergePrefixes(incoming map[uuid.UUID]*CaniPrefix) (map[uuid.UUID]uuid.UUID, error) {
	if inv.Prefixes == nil {
		inv.Prefixes = make(map[uuid.UUID]*CaniPrefix)
	}
	index := indexByKey(inv.Prefixes, prefixNaturalKey)
	remap := make(map[uuid.UUID]uuid.UUID, len(incoming))
	for _, incomingID := range sortedIDs(incoming) {
		prefix := incoming[incomingID]
		if prefix == nil {
			continue
		}
		key := prefixNaturalKey(prefix)
		resolvedID, err := resolveScopedIdentity(incomingID, prefix.ExternalIDs, inv.Prefixes,
			index[key], func(p *CaniPrefix) map[string]uuid.UUID { return p.ExternalIDs })
		if err != nil {
			return nil, fmt.Errorf("prefix %s in namespace %q: %w", prefix.Prefix, prefix.EffectiveNamespace(), err)
		}
		prefix.ID = resolvedID
		inv.Prefixes[resolvedID] = prefix
		index[key] = appendUnique(index[key], resolvedID)
		remap[incomingID] = resolvedID
	}
	return remap, nil
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
	index := indexByKey(inv.IPAddresses, keyOf)
	remap := make(map[uuid.UUID]uuid.UUID, len(incoming))
	for _, incomingID := range sortedIDs(incoming) {
		address := incoming[incomingID]
		if address == nil {
			continue
		}
		key := keyOf(address)
		resolvedID, err := resolveScopedIdentity(incomingID, address.ExternalIDs, inv.IPAddresses,
			index[key], func(a *CaniIPAddress) map[string]uuid.UUID { return a.ExternalIDs })
		if err != nil {
			return nil, fmt.Errorf("IP address %s in namespace %q: %w", key.Host, key.Namespace, err)
		}
		address.ID = resolvedID
		inv.IPAddresses[resolvedID] = address
		index[key] = appendUnique(index[key], resolvedID)
		remap[incomingID] = resolvedID
	}
	return remap, nil
}

// MergeVRFs merges VRFs by UUID, then by shared external ID, then by
// namespace + name, then inserts. Same-named VRFs in one namespace make a
// name match ambiguous and are reported instead of resolved to the first.
func (inv *Inventory) MergeVRFs(incoming map[uuid.UUID]*CaniVRF) (map[uuid.UUID]uuid.UUID, error) {
	if inv.VRFs == nil {
		inv.VRFs = make(map[uuid.UUID]*CaniVRF)
	}
	index := indexByKey(inv.VRFs, vrfNaturalKey)
	remap := make(map[uuid.UUID]uuid.UUID, len(incoming))
	for _, incomingID := range sortedIDs(incoming) {
		vrf := incoming[incomingID]
		if vrf == nil || vrf.Name == "" {
			continue
		}
		key := vrfNaturalKey(vrf)
		resolvedID, err := resolveScopedIdentity(incomingID, vrf.ExternalIDs, inv.VRFs,
			index[key], func(v *CaniVRF) map[string]uuid.UUID { return v.ExternalIDs })
		if err != nil {
			return nil, fmt.Errorf("VRF %s in namespace %q: %w", vrf.Name, vrf.EffectiveNamespace(), err)
		}
		vrf.ID = resolvedID
		inv.VRFs[resolvedID] = vrf
		index[key] = appendUnique(index[key], resolvedID)
		remap[incomingID] = resolvedID
	}
	return remap, nil
}

func appendUnique(ids []uuid.UUID, id uuid.UUID) []uuid.UUID {
	if containsUUID(ids, id) {
		return ids
	}
	return append(ids, id)
}

// indexByKey groups existing object IDs by natural key in a stable order.
func indexByKey[T any, K comparable](items map[uuid.UUID]*T, keyOf func(*T) K) map[K][]uuid.UUID {
	index := make(map[K][]uuid.UUID, len(items))
	for _, id := range sortedIDs(items) {
		if item := items[id]; item != nil {
			key := keyOf(item)
			index[key] = append(index[key], id)
		}
	}
	return index
}

// resolveScopedIdentity picks the inventory UUID an incoming object should
// take: its own when already present, otherwise the existing object sharing
// a provider external ID, otherwise the single existing object with the same
// natural key, otherwise itself (insert). Conflicting external IDs on a key
// match and several key matches are errors.
func resolveScopedIdentity[T any](
	incomingID uuid.UUID,
	incomingExternalIDs map[string]uuid.UUID,
	existing map[uuid.UUID]*T,
	keyMatches []uuid.UUID,
	externalIDsOf func(*T) map[string]uuid.UUID,
) (uuid.UUID, error) {
	if _, ok := existing[incomingID]; ok {
		return incomingID, nil
	}
	for _, id := range sortedIDs(existing) {
		if sharedExternalID(externalIDsOf(existing[id]), incomingExternalIDs) {
			return id, nil
		}
	}
	switch len(keyMatches) {
	case 0:
		return incomingID, nil
	case 1:
		if conflictingExternalID(externalIDsOf(existing[keyMatches[0]]), incomingExternalIDs) {
			return uuid.Nil, fmt.Errorf("existing %s has a different source identity; "+
				"resolve the conflict explicitly", keyMatches[0])
		}
		return keyMatches[0], nil
	default:
		return uuid.Nil, fmt.Errorf("natural key is ambiguous among %d existing records (%v); "+
			"resolve the duplicates explicitly", len(keyMatches), keyMatches)
	}
}
