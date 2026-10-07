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

// AddVRF inserts a single VRF into the inventory. Its name must be unique in
// its namespace; names are compared exactly, because VRF identity is
// case-sensitive.
func (inv *Inventory) AddVRF(vrf *CaniVRF) error {
	if vrf == nil {
		return fmt.Errorf("vrf must not be nil")
	}
	if _, exists := inv.VRFs[vrf.ID]; exists {
		return fmt.Errorf("vrf %s already exists", vrf.ID)
	}
	key := vrfNaturalKey(vrf)
	if holder, held := holderOfKey(inv.VRFs, key, vrfNaturalKey); held {
		return fmt.Errorf("VRF %q already exists in namespace %q as %s", key.Name, key.Namespace, holder)
	}
	for _, devID := range vrf.Devices {
		if _, ok := inv.Devices[devID]; !ok {
			return fmt.Errorf("device %s not found", devID)
		}
	}
	inv.VRFs[vrf.ID] = vrf
	return nil
}

// AddPrefix inserts a single prefix into the inventory and auto-computes its
// parent. Its CIDR must be unique in its namespace, and each VRF membership
// must name an existing VRF in that namespace.
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
	key := prefixNaturalKey(prefix)
	if holder, held := holderOfKey(inv.Prefixes, key, prefixNaturalKey); held {
		return fmt.Errorf("prefix %s already exists in namespace %q as %s", key.CIDR, key.Namespace, holder)
	}
	if err := inv.checkPrefixVRFs(prefix); err != nil {
		return err
	}
	if prefix.Parent == uuid.Nil {
		prefix.Parent = FindParentPrefix(prefix, inv.Prefixes)
	}
	inv.Prefixes[prefix.ID] = prefix
	return nil
}

// AddIPAddress inserts a single IP address into the inventory and
// auto-computes its parent prefix. Its host must be unique in its namespace;
// the mask is not part of an address's identity.
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
	keyOf := func(a *CaniIPAddress) ipAddressKey { return ipAddressNaturalKey(a, inv.IPAddressNamespace(a)) }
	key := keyOf(addr)
	if holder, held := holderOfKey(inv.IPAddresses, key, keyOf); held {
		return fmt.Errorf("IP address %s already exists in namespace %q as %s", key.Host, key.Namespace, holder)
	}
	if addr.Parent == uuid.Nil {
		addr.Parent = FindParentPrefixForIP(addr, inv.Prefixes)
	}
	inv.IPAddresses[addr.ID] = addr
	return nil
}

// holderOfKey returns the lowest-UUID record whose natural key is key.
func holderOfKey[T any, K comparable](items map[uuid.UUID]*T, key K, keyOf func(*T) K) (uuid.UUID, bool) {
	if matches := recordsWithKey(items, key, keyOf); len(matches) > 0 {
		return matches[0], true
	}
	return uuid.Nil, false
}

// recordsWithKey returns the UUIDs of the records whose natural key is key,
// in UUID order.
func recordsWithKey[T any, K comparable](items map[uuid.UUID]*T, key K, keyOf func(*T) K) []uuid.UUID {
	var matches []uuid.UUID
	for _, id := range sortedIDs(items) {
		if item := items[id]; item != nil && keyOf(item) == key {
			matches = append(matches, id)
		}
	}
	return matches
}
