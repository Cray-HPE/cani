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
	"slices"

	"github.com/google/uuid"
)

// validateIPAMScope checks that every IPAM relationship stays inside one
// namespace and reports legacy duplicates that only a namespace could have
// separated. Cross-namespace links are errors; duplicates are unresolved
// conflicts, so legacy inventories still load and every change lists them
// until they are resolved explicitly.
func (inv *Inventory) validateIPAMScope(result *RelationshipResult) {
	inv.validatePrefixScope(result)
	inv.validateIPAddressScope(result)
	inv.reportDuplicatePrefixes(result)
	inv.reportDuplicateIPAddresses(result)
	inv.reportDuplicateVRFs(result)
}

func (inv *Inventory) validatePrefixScope(result *RelationshipResult) {
	for prefixID, prefix := range inv.Prefixes {
		if prefix == nil {
			continue
		}
		owner := fmt.Sprintf("prefix %q (%s)", prefix.Prefix, prefixID)
		inv.validatePrefixParentScope(result, owner, prefix)
		inv.validatePrefixVRFScope(result, owner, prefix)
		if prefix.VRF != "" {
			result.Unresolved = append(result.Unresolved, fmt.Sprintf(
				"%s: legacy VRF name %q does not resolve to exactly one VRF in namespace %q; set vrfs explicitly",
				owner, prefix.VRF, prefix.EffectiveNamespace()))
		}
	}
}

func (inv *Inventory) validatePrefixParentScope(result *RelationshipResult, owner string, prefix *CaniPrefix) {
	parent := inv.Prefixes[prefix.Parent]
	if parent == nil || parent.EffectiveNamespace() == prefix.EffectiveNamespace() {
		return
	}
	result.Errors = append(result.Errors, fmt.Errorf("%s: parent prefix %q is in namespace %q, not %q",
		owner, parent.Prefix, parent.EffectiveNamespace(), prefix.EffectiveNamespace()))
}

func (inv *Inventory) validatePrefixVRFScope(result *RelationshipResult, owner string, prefix *CaniPrefix) {
	for _, vrfID := range prefix.VRFs {
		if err := inv.prefixVRFError(prefix, vrfID); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("%s: %w", owner, err))
		}
	}
}

// checkPrefixVRFs rejects a VRF membership that is missing or outside the
// prefix's namespace.
func (inv *Inventory) checkPrefixVRFs(prefix *CaniPrefix) error {
	for _, vrfID := range prefix.VRFs {
		if err := inv.prefixVRFError(prefix, vrfID); err != nil {
			return fmt.Errorf("prefix %s: %w", prefix.Prefix, err)
		}
	}
	return nil
}

// AddPrefixVRF makes prefix a member of the VRF, which must exist in the
// prefix's namespace, and reports whether anything changed. An explicit
// membership supersedes a legacy VRF name, so it clears one the migration
// could not fold.
func (inv *Inventory) AddPrefixVRF(prefix *CaniPrefix, vrfID uuid.UUID) (bool, error) {
	if err := inv.prefixVRFError(prefix, vrfID); err != nil {
		return false, fmt.Errorf("prefix %s: %w", prefix.Prefix, err)
	}
	cleared := prefix.ClearLegacyVRF()
	if slices.Contains(prefix.VRFs, vrfID) {
		return cleared, nil
	}
	prefix.VRFs = append(prefix.VRFs, vrfID)
	return true, nil
}

// RemoveVRF drops the prefix's membership of the VRF and reports whether it
// was a member. The VRF need not exist, so a membership left behind by a VRF
// that is gone can be dropped too.
func (p *CaniPrefix) RemoveVRF(vrfID uuid.UUID) bool {
	if !slices.Contains(p.VRFs, vrfID) {
		return false
	}
	p.VRFs = removeUUID(p.VRFs, vrfID)
	return true
}

// ClearLegacyVRF drops the legacy VRF name the migration could not fold into
// memberships and reports whether there was one.
func (p *CaniPrefix) ClearLegacyVRF() bool {
	if p.VRF == "" {
		return false
	}
	p.VRF = ""
	return true
}

// prefixVRFError reports why vrfID cannot be a membership of prefix: the VRF
// is missing or belongs to another namespace.
func (inv *Inventory) prefixVRFError(prefix *CaniPrefix, vrfID uuid.UUID) error {
	vrf := inv.VRFs[vrfID]
	switch {
	case vrf == nil:
		return fmt.Errorf("VRF %s not found", vrfID)
	case vrf.EffectiveNamespace() != prefix.EffectiveNamespace():
		return fmt.Errorf("VRF %q is in namespace %q, not %q",
			vrf.Name, vrf.EffectiveNamespace(), prefix.EffectiveNamespace())
	}
	return nil
}

func (inv *Inventory) validateIPAddressScope(result *RelationshipResult) {
	for addressID, address := range inv.IPAddresses {
		if address == nil {
			continue
		}
		if parent := inv.Prefixes[address.Parent]; parent != nil {
			owner := fmt.Sprintf("IP address %q (%s)", address.Address, addressID)
			validateIPAddressParentScope(result, owner, address, parent)
		}
	}
}

func validateIPAddressParentScope(result *RelationshipResult, owner string, address *CaniIPAddress, parent *CaniPrefix) {
	if address.Namespace != "" && NamespaceOrDefault(address.Namespace) != parent.EffectiveNamespace() {
		result.Errors = append(result.Errors, fmt.Errorf("%s: parent prefix %q is in namespace %q, not %q",
			owner, parent.Prefix, parent.EffectiveNamespace(), address.Namespace))
	}
	if !prefixContainsHost(parent.Prefix, address.Host) {
		result.Errors = append(result.Errors, fmt.Errorf("%s: parent prefix %q does not contain the address",
			owner, parent.Prefix))
	}
}

// prefixContainsHost reports whether the CIDR contains the host. Unparseable
// input is treated as contained so the CIDR/address parsers report the fault.
func prefixContainsHost(cidr, host string) bool {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return true
	}
	return network.Contains(ip)
}

func (inv *Inventory) reportDuplicatePrefixes(result *RelationshipResult) {
	seen := make(map[prefixKey]uuid.UUID, len(inv.Prefixes))
	for _, id := range sortedIDs(inv.Prefixes) {
		prefix := inv.Prefixes[id]
		if prefix == nil {
			continue
		}
		key := prefixNaturalKey(prefix)
		if first, dup := seen[key]; dup {
			result.Unresolved = append(result.Unresolved, fmt.Sprintf(
				"prefix %s exists twice in namespace %q (%s, %s); resolve the duplicate explicitly",
				key.CIDR, key.Namespace, first, id))
			continue
		}
		seen[key] = id
	}
}

func (inv *Inventory) reportDuplicateIPAddresses(result *RelationshipResult) {
	seen := make(map[ipAddressKey]uuid.UUID, len(inv.IPAddresses))
	for _, id := range sortedIDs(inv.IPAddresses) {
		address := inv.IPAddresses[id]
		if address == nil {
			continue
		}
		key := ipAddressNaturalKey(address, inv.IPAddressNamespace(address))
		if first, dup := seen[key]; dup {
			result.Unresolved = append(result.Unresolved, fmt.Sprintf(
				"IP address %s exists twice in namespace %q (%s, %s); resolve the duplicate explicitly",
				key.Host, key.Namespace, first, id))
			continue
		}
		seen[key] = id
	}
}

func (inv *Inventory) reportDuplicateVRFs(result *RelationshipResult) {
	seen := make(map[vrfKey]uuid.UUID, len(inv.VRFs))
	for _, id := range sortedIDs(inv.VRFs) {
		vrf := inv.VRFs[id]
		if vrf == nil {
			continue
		}
		key := vrfNaturalKey(vrf)
		if first, dup := seen[key]; dup {
			result.Unresolved = append(result.Unresolved, fmt.Sprintf(
				"VRF %q exists twice in namespace %q (%s, %s); qualified name lookups will be ambiguous",
				key.Name, key.Namespace, first, id))
			continue
		}
		seen[key] = id
	}
}
