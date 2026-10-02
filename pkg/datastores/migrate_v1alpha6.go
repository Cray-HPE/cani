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
package datastores

import "github.com/Cray-HPE/cani/pkg/devicetypes"

// migrateV1Alpha6 upgrades a v1alpha6 inventory to v1alpha7 in place.
// v1alpha7 adds namespace scope to prefixes and IP addresses and replaces the
// legacy single VRF name on a prefix with canonical VRF memberships.
//
// Legacy prefixes and VRFs genuinely omitted a namespace, so the migration
// stamps them Global rather than leaving scope implicit. A legacy VRF name is
// folded into VRFs only when exactly one VRF of that name exists in the
// prefix's namespace; otherwise the name is retained so the intent is reported
// instead of guessed. Duplicate CIDRs that only the VRF name separated are
// kept as distinct records for explicit resolution.
func migrateV1Alpha6(inv *devicetypes.Inventory) {
	for _, vrf := range inv.VRFs {
		if vrf != nil && vrf.Namespace == "" {
			vrf.Namespace = devicetypes.DefaultNamespace
		}
	}
	for _, prefix := range inv.Prefixes {
		if prefix == nil {
			continue
		}
		if prefix.Namespace == "" {
			prefix.Namespace = devicetypes.DefaultNamespace
		}
		resolveLegacyPrefixVRF(inv, prefix)
	}
	inv.SchemaVersion = devicetypes.SchemaVersionV1Alpha7
}

// resolveLegacyPrefixVRF folds a uniquely resolvable legacy VRF name into the
// prefix's VRF memberships and clears the name; anything else is left as is.
func resolveLegacyPrefixVRF(inv *devicetypes.Inventory, prefix *devicetypes.CaniPrefix) {
	if prefix.VRF == "" {
		return
	}
	matches := inv.FindVRFsByName(prefix.EffectiveNamespace(), prefix.VRF)
	if len(matches) != 1 {
		return
	}
	for _, existing := range prefix.VRFs {
		if existing == matches[0].ID {
			prefix.VRF = ""
			return
		}
	}
	prefix.VRFs = append(prefix.VRFs, matches[0].ID)
	prefix.VRF = ""
}
