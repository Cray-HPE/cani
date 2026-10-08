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

// DefaultNamespace is the IPAM namespace assumed when an object omits one.
// It matches Nautobot's built-in namespace so legacy single-namespace
// inventories keep their meaning.
const DefaultNamespace = "Global"

// NamespaceOrDefault returns name without surrounding whitespace, or
// DefaultNamespace when nothing is left, so " Global" and "Global" are one
// namespace wherever scope is compared.
func NamespaceOrDefault(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return DefaultNamespace
	}
	return name
}

// EffectiveNamespace returns the namespace that scopes the prefix.
func (p *CaniPrefix) EffectiveNamespace() string {
	if p == nil {
		return DefaultNamespace
	}
	return NamespaceOrDefault(p.Namespace)
}

// EffectiveNamespace returns the namespace that scopes the VRF.
func (v *CaniVRF) EffectiveNamespace() string {
	if v == nil {
		return DefaultNamespace
	}
	return NamespaceOrDefault(v.Namespace)
}

// IPAddressNamespace returns the namespace an address belongs to. Prefix
// scope is authoritative, so a parented address inherits its parent's
// namespace; a draft without a parent falls back to its own intent.
func (inv *Inventory) IPAddressNamespace(addr *CaniIPAddress) string {
	if addr == nil {
		return DefaultNamespace
	}
	if inv != nil && addr.Parent != uuid.Nil {
		if parent := inv.Prefixes[addr.Parent]; parent != nil {
			return parent.EffectiveNamespace()
		}
	}
	return NamespaceOrDefault(addr.Namespace)
}

// prefixKey is the natural key of a prefix: its namespace plus canonical CIDR.
type prefixKey struct {
	Namespace string
	CIDR      string
}

// ipAddressKey is the natural key of an address: its namespace plus canonical
// host. The mask length is deliberately excluded.
type ipAddressKey struct {
	Namespace string
	Host      string
}

// vrfKey is the natural key of a VRF: its namespace plus name.
type vrfKey struct {
	Namespace string
	Name      string
}

// canonicalCIDR normalizes a CIDR to its network form ("10.0.0.5/24" becomes
// "10.0.0.0/24") so equivalent spellings share one key. Unparseable input is
// returned trimmed so it still participates in exact-match comparisons.
func canonicalCIDR(cidr string) string {
	_, network, err := net.ParseCIDR(strings.TrimSpace(cidr))
	if err != nil {
		return strings.TrimSpace(cidr)
	}
	return network.String()
}

// canonicalHost normalizes an address or CIDR to its bare host form.
func canonicalHost(address string) string {
	host := strings.TrimSpace(address)
	if idx := strings.IndexByte(host, '/'); idx >= 0 {
		host = host[:idx]
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.String()
	}
	return host
}

func prefixNaturalKey(p *CaniPrefix) prefixKey {
	return prefixKey{Namespace: p.EffectiveNamespace(), CIDR: canonicalCIDR(p.Prefix)}
}

func ipAddressNaturalKey(addr *CaniIPAddress, namespace string) ipAddressKey {
	host := addr.Host
	if host == "" {
		host = addr.Address
	}
	return ipAddressKey{Namespace: NamespaceOrDefault(namespace), Host: canonicalHost(host)}
}

func vrfNaturalKey(v *CaniVRF) vrfKey {
	return vrfKey{Namespace: v.EffectiveNamespace(), Name: v.Name}
}

// FindVRFsByName returns the VRFs in the namespace named exactly name, or,
// when there are none, those whose name matches it case-insensitively, in a
// stable order. VRF identity is case-sensitive, as in Nautobot, so "red" and
// "Red" are different VRFs; the fallback only spares typing the exact case.
// Callers decide how to treat zero or several matches; the lookup never
// picks a winner for them.
func (inv *Inventory) FindVRFsByName(namespace, name string) []*CaniVRF {
	if inv == nil {
		return nil
	}
	want := NamespaceOrDefault(namespace)
	var exact, folded []*CaniVRF
	for _, id := range sortedIDs(inv.VRFs) {
		vrf := inv.VRFs[id]
		if vrf == nil || vrf.EffectiveNamespace() != want {
			continue
		}
		if vrf.Name == name {
			exact = append(exact, vrf)
		} else if strings.EqualFold(vrf.Name, name) {
			folded = append(folded, vrf)
		}
	}
	if len(exact) > 0 {
		return exact
	}
	return folded
}

// ResolveVRFReference resolves a VRF by UUID or by name. A name may be
// qualified as "namespace/name" (see splitQualified); an unqualified name is
// searched in the given namespace. Several same-named VRFs are reported as
// ambiguous rather than resolved to the first result.
func (inv *Inventory) ResolveVRFReference(namespace, ref string) (*CaniVRF, error) {
	if id, err := uuid.Parse(ref); err == nil {
		if vrf, ok := inv.VRFs[id]; ok && vrf != nil {
			return vrf, nil
		}
		return nil, fmt.Errorf("VRF with UUID %q not found", ref)
	}
	namespace, name := splitQualified(namespace, ref, isVRFName)
	matches := inv.FindVRFsByName(namespace, name)
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("VRF %q not found in namespace %q", name, NamespaceOrDefault(namespace))
	case 1:
		return matches[0], nil
	default:
		return nil, fmt.Errorf("VRF %q is ambiguous in namespace %q (%d matches); use the UUID",
			name, NamespaceOrDefault(namespace), len(matches))
	}
}
