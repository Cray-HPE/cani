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
	"net"
	"strings"

	"github.com/google/uuid"
)

// ResolvePrefixReference resolves a prefix by UUID or by CIDR. A UUID names
// one record whatever its namespace, so the namespace argument is ignored. A
// CIDR may be qualified as "namespace/CIDR" (see splitQualified); an
// unqualified one is searched in the given namespace. Equivalent spellings of
// a CIDR match, and a CIDR held by several legacy records is reported with
// their UUIDs rather than resolved to one.
func (inv *Inventory) ResolvePrefixReference(namespace, ref string) (*CaniPrefix, error) {
	if id, err := uuid.Parse(ref); err == nil {
		if prefix := inv.Prefixes[id]; prefix != nil {
			return prefix, nil
		}
		return nil, fmt.Errorf("prefix with UUID %q not found", ref)
	}
	namespace, cidr := splitQualified(namespace, ref, isCIDR)
	key := prefixKey{Namespace: NamespaceOrDefault(namespace), CIDR: canonicalCIDR(cidr)}
	return singleMatch(inv.Prefixes, key, prefixNaturalKey, "prefix "+key.CIDR, key.Namespace)
}

// ResolveIPAddressReference resolves an address by UUID or by host, with or
// without a mask. A UUID names one record whatever its namespace. A host may
// be qualified as "namespace/host"; an unqualified one is searched in the
// given namespace.
func (inv *Inventory) ResolveIPAddressReference(namespace, ref string) (*CaniIPAddress, error) {
	if id, err := uuid.Parse(ref); err == nil {
		if address := inv.IPAddresses[id]; address != nil {
			return address, nil
		}
		return nil, fmt.Errorf("IP address with UUID %q not found", ref)
	}
	namespace, host := splitQualified(namespace, ref, isAddress)
	key := ipAddressKey{Namespace: NamespaceOrDefault(namespace), Host: canonicalHost(host)}
	keyOf := func(a *CaniIPAddress) ipAddressKey { return ipAddressNaturalKey(a, inv.IPAddressNamespace(a)) }
	return singleMatch(inv.IPAddresses, key, keyOf, "IP address "+key.Host, key.Namespace)
}

// splitQualified splits a "namespace/value" reference: the value is the
// shortest suffix after a slash that isValue accepts, so a CIDR keeps its mask
// and a namespace may itself contain slashes. A reference that is a value as a
// whole, or has no slash, is searched in the given namespace. Every resolver
// uses this one grammar.
func splitQualified(namespace, ref string, isValue func(string) bool) (string, string) {
	for idx := strings.LastIndexByte(ref, '/'); idx >= 0; idx = strings.LastIndexByte(ref[:idx], '/') {
		if isValue(ref[idx+1:]) {
			return ref[:idx], ref[idx+1:]
		}
	}
	return namespace, ref
}

func isCIDR(ref string) bool {
	_, _, err := net.ParseCIDR(ref)
	return err == nil
}

func isAddress(ref string) bool {
	return isCIDR(ref) || net.ParseIP(ref) != nil
}

// isVRFName accepts any non-empty name, so a qualified VRF reference splits at
// its last slash.
func isVRFName(ref string) bool {
	return ref != ""
}

// singleMatch returns the one record with key, or an error naming what was
// sought and, when several records hold the key, their UUIDs.
func singleMatch[T any, K comparable](
	items map[uuid.UUID]*T, key K, keyOf func(*T) K, what, namespace string,
) (*T, error) {
	matches := recordsWithKey(items, key, keyOf)
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("%s not found in namespace %q", what, namespace)
	case 1:
		return items[matches[0]], nil
	default:
		return nil, fmt.Errorf("%s is ambiguous in namespace %q; use one of the UUIDs %v", what, namespace, matches)
	}
}
