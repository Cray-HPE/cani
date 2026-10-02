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
package transform

import (
	"fmt"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// External-ID accessors for nautobotIDIndex.
func prefixExternalIDs(p *devicetypes.CaniPrefix) map[string]uuid.UUID     { return p.ExternalIDs }
func vrfExternalIDs(v *devicetypes.CaniVRF) map[string]uuid.UUID           { return v.ExternalIDs }
func addressExternalIDs(a *devicetypes.CaniIPAddress) map[string]uuid.UUID { return a.ExternalIDs }

// BuildNamespaceNameMap creates a lookup from namespace UUID to name.
func BuildNamespaceNameMap(namespaces []nautobotapi.Namespace) map[uuid.UUID]string {
	m := make(map[uuid.UUID]string, len(namespaces))
	for _, ns := range namespaces {
		if id := directUUID(ns.Id); id != uuid.Nil {
			m[id] = ns.Name
		}
	}
	return m
}

// BuildVRFPrefixMembershipMap groups VRF-prefix assignments by Nautobot
// prefix UUID so each prefix can list the VRFs it belongs to.
func BuildVRFPrefixMembershipMap(assignments []nautobotapi.VRFPrefixAssignment) map[uuid.UUID][]uuid.UUID {
	m := make(map[uuid.UUID][]uuid.UUID, len(assignments))
	for _, a := range assignments {
		prefixID, vrfID := refID(a.Prefix), refID(a.Vrf)
		if prefixID != uuid.Nil && vrfID != uuid.Nil {
			m[prefixID] = append(m[prefixID], vrfID)
		}
	}
	return m
}

// nautobotIDIndex maps each object's Nautobot external ID back to its CANI ID.
func nautobotIDIndex[T any](items map[uuid.UUID]*T, externalIDs func(*T) map[string]uuid.UUID) map[uuid.UUID]uuid.UUID {
	index := make(map[uuid.UUID]uuid.UUID, len(items))
	for caniID, item := range items {
		if item == nil {
			continue
		}
		if nbID := externalIDs(item)[externalIDKeyNautobot]; nbID != uuid.Nil {
			index[nbID] = caniID
		}
	}
	return index
}

// resolveNamespaceName returns the namespace name for a reference. A nil
// reference means the source omitted scope and keeps the Global default; a
// reference that is present but unknown is reported as unresolved.
func resolveNamespaceName(ref any, namespaceNameMap map[uuid.UUID]string) (string, bool) {
	id := refID(ref)
	if id == uuid.Nil {
		return "", true
	}
	name, ok := namespaceNameMap[id]
	return name, ok
}

// ScopeVRFs stamps each mapped VRF with its Nautobot namespace name. A VRF
// whose namespace cannot be resolved is left out of the result rather than
// filed under Global; the returned lines report each one.
func ScopeVRFs(vrfs map[uuid.UUID]*devicetypes.CaniVRF, raw []nautobotapi.VRF, namespaceNameMap map[uuid.UUID]string) []string {
	index := nautobotIDIndex(vrfs, vrfExternalIDs)
	var warnings []string
	for _, rawVRF := range raw {
		caniID, ok := index[directUUID(rawVRF.Id)]
		if !ok {
			continue
		}
		name, resolved := resolveNamespaceName(rawVRF.Namespace, namespaceNameMap)
		if !resolved {
			warnings = append(warnings, fmt.Sprintf("VRF %s: namespace %s not found", rawVRF.Name, refID(rawVRF.Namespace)))
			delete(vrfs, caniID)
			continue
		}
		vrfs[caniID].Namespace = name
	}
	return warnings
}

// ScopePrefixes stamps each mapped prefix with its namespace, then links the
// retained prefixes: the Nautobot parent is kept as mapped import scope and
// VRF memberships come from the assignment records. A prefix whose namespace
// cannot be resolved is left out and reported; a child whose parent was left
// out keeps no parent, so the model derives one inside its namespace.
func ScopePrefixes(
	prefixes map[uuid.UUID]*devicetypes.CaniPrefix,
	raw []nautobotapi.Prefix,
	namespaceNameMap map[uuid.UUID]string,
	vrfs map[uuid.UUID]*devicetypes.CaniVRF,
	memberships map[uuid.UUID][]uuid.UUID,
) []string {
	warnings := scopePrefixNamespaces(prefixes, raw, namespaceNameMap)
	linkPrefixes(prefixes, raw, vrfs, memberships)
	return warnings
}

// scopePrefixNamespaces resolves each prefix's namespace reference, removing
// the prefixes whose reference is unknown.
func scopePrefixNamespaces(prefixes map[uuid.UUID]*devicetypes.CaniPrefix, raw []nautobotapi.Prefix, namespaceNameMap map[uuid.UUID]string) []string {
	index := nautobotIDIndex(prefixes, prefixExternalIDs)
	var warnings []string
	for _, rawPrefix := range raw {
		caniID, ok := index[directUUID(rawPrefix.Id)]
		if !ok {
			continue
		}
		name, resolved := resolveNamespaceName(rawPrefix.Namespace, namespaceNameMap)
		if !resolved {
			warnings = append(warnings, fmt.Sprintf("prefix %s: namespace %s not found", rawPrefix.Prefix, refID(rawPrefix.Namespace)))
			delete(prefixes, caniID)
			continue
		}
		prefixes[caniID].Namespace = name
	}
	return warnings
}

// linkPrefixes sets parents and VRF memberships from indexes built after the
// namespace pass, so a reference to a left-out prefix or VRF resolves to none.
func linkPrefixes(
	prefixes map[uuid.UUID]*devicetypes.CaniPrefix,
	raw []nautobotapi.Prefix,
	vrfs map[uuid.UUID]*devicetypes.CaniVRF,
	memberships map[uuid.UUID][]uuid.UUID,
) {
	prefixIndex := nautobotIDIndex(prefixes, prefixExternalIDs)
	vrfIndex := nautobotIDIndex(vrfs, vrfExternalIDs)
	for _, rawPrefix := range raw {
		nbID := directUUID(rawPrefix.Id)
		prefix, ok := prefixes[prefixIndex[nbID]]
		if !ok {
			continue
		}
		prefix.Parent = prefixIndex[refID(rawPrefix.Parent)]
		prefix.VRFs = mapVRFMemberships(memberships[nbID], vrfIndex)
	}
}

// mapVRFMemberships translates Nautobot VRF UUIDs to the CANI VRF IDs that
// were imported, skipping any VRF that was not.
func mapVRFMemberships(nautobotVRFs []uuid.UUID, vrfIndex map[uuid.UUID]uuid.UUID) []uuid.UUID {
	var caniVRFs []uuid.UUID
	for _, nbID := range nautobotVRFs {
		if caniID, ok := vrfIndex[nbID]; ok {
			caniVRFs = append(caniVRFs, caniID)
		}
	}
	return caniVRFs
}

// ScopeIPAddresses preserves each address's Nautobot parent prefix and takes
// its namespace from that parent. An address whose parent is not among the
// retained prefixes has no knowable scope and is left out rather than
// re-parented into Global; the returned lines report each one.
func ScopeIPAddresses(
	addresses map[uuid.UUID]*devicetypes.CaniIPAddress,
	raw []nautobotapi.IPAddress,
	prefixes map[uuid.UUID]*devicetypes.CaniPrefix,
) []string {
	addressIndex := nautobotIDIndex(addresses, addressExternalIDs)
	prefixIndex := nautobotIDIndex(prefixes, prefixExternalIDs)
	var warnings []string
	for _, rawIP := range raw {
		caniID, ok := addressIndex[directUUID(rawIP.Id)]
		if !ok {
			continue
		}
		parentNBID := refID(rawIP.Parent)
		if parentNBID == uuid.Nil {
			continue
		}
		parentCaniID, ok := prefixIndex[parentNBID]
		if !ok {
			warnings = append(warnings, fmt.Sprintf("IP address %s: parent prefix %s not imported", rawIP.Address, parentNBID))
			delete(addresses, caniID)
			continue
		}
		addresses[caniID].Parent = parentCaniID
		addresses[caniID].Namespace = prefixes[parentCaniID].EffectiveNamespace()
	}
	return warnings
}
