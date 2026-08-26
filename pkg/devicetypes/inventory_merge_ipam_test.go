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
	"testing"

	"github.com/google/uuid"
)

// TestMergeIPAMDeduplicatesByNaturalKey verifies each IPAM merge collapses an
// incoming object onto the existing one that shares its natural key.
//
// Why it matters: a provider mints fresh UUIDs on every import, so re-importing
// the same source must not double every VLAN, prefix, address and VRF in the
// inventory. UUID equality cannot do that job; only the natural key can. The
// returned remap is what lets every reference to the discarded UUID be rewritten
// afterwards, so a merge that deduplicated but returned the wrong mapping would
// leave dangling references.
// Inputs: an inventory seeded with one of each IPAM type, then a merge of a
// second object of each type carrying a different UUID but the same natural key.
// Outputs: the collection size is unchanged and the remap points the incoming
// UUID at the pre-existing one.
// Data choice: the incoming objects differ in a non-key field (Description) so
// the test also shows the incoming value wins on collision, which is what makes
// re-import an update rather than a no-op.
func TestMergeIPAMDeduplicatesByNaturalKey(t *testing.T) {
	location := uuid.New()
	existing := struct{ vlan, prefix, address, vrf uuid.UUID }{
		uuid.New(), uuid.New(), uuid.New(), uuid.New(),
	}

	inventory := NewInventory()
	inventory.VLANs[existing.vlan] = &CaniVLAN{ID: existing.vlan, VID: 100, Location: location}
	inventory.Prefixes[existing.prefix] = &CaniPrefix{ID: existing.prefix, Prefix: "10.0.0.0/24", VRF: "blue"}
	inventory.IPAddresses[existing.address] = &CaniIPAddress{ID: existing.address, Address: "10.0.0.5/24"}
	inventory.VRFs[existing.vrf] = &CaniVRF{ID: existing.vrf, Name: "blue"}

	incoming := struct{ vlan, prefix, address, vrf uuid.UUID }{
		uuid.New(), uuid.New(), uuid.New(), uuid.New(),
	}

	vlanRemap := inventory.MergeVLANs(map[uuid.UUID]*CaniVLAN{
		incoming.vlan: {ID: incoming.vlan, VID: 100, Location: location, Description: "merged"},
	})
	prefixRemap := inventory.MergePrefixes(map[uuid.UUID]*CaniPrefix{
		incoming.prefix: {ID: incoming.prefix, Prefix: "10.0.0.0/24", VRF: "blue", Description: "merged"},
	})
	addressRemap := inventory.MergeIPAddresses(map[uuid.UUID]*CaniIPAddress{
		incoming.address: {ID: incoming.address, Address: "10.0.0.5/24", Description: "merged"},
	})
	vrfRemap := inventory.MergeVRFs(map[uuid.UUID]*CaniVRF{
		incoming.vrf: {ID: incoming.vrf, Name: "blue", Description: "merged"},
	})

	for _, check := range []struct {
		name        string
		size, want  int
		got, wantID uuid.UUID
	}{
		{"vlan", len(inventory.VLANs), 1, vlanRemap[incoming.vlan], existing.vlan},
		{"prefix", len(inventory.Prefixes), 1, prefixRemap[incoming.prefix], existing.prefix},
		{"address", len(inventory.IPAddresses), 1, addressRemap[incoming.address], existing.address},
		{"vrf", len(inventory.VRFs), 1, vrfRemap[incoming.vrf], existing.vrf},
	} {
		if check.size != check.want {
			t.Errorf("%s count = %d, want %d (natural-key merge duplicated instead)", check.name, check.size, check.want)
		}
		if check.got != check.wantID {
			t.Errorf("%s remap = %v, want %v", check.name, check.got, check.wantID)
		}
	}

	if got := inventory.VLANs[existing.vlan].Description; got != "merged" {
		t.Errorf("vlan description = %q, want %q (incoming value should win)", got, "merged")
	}
	if got := inventory.VRFs[existing.vrf].GetID(); got != existing.vrf {
		t.Errorf("vrf GetID = %v, want %v", got, existing.vrf)
	}
}

// TestMergeIPAMInsertsWhenNaturalKeyDiffers verifies a genuinely new IPAM object
// is added rather than folded onto an unrelated one.
//
// Why it matters: the dedup above is only safe if the natural key is precise. A
// VLAN is keyed on VID *and* location, so the same VID in two locations must
// stay two VLANs; a prefix is keyed on CIDR *and* VRF for the same reason.
// Collapsing those would silently destroy inventory.
// Inputs: an inventory holding VID 100 in one location and prefix 10.0.0.0/24 in
// VRF "blue", merged with the same VID in a different location and the same CIDR
// in a different VRF.
// Outputs: both collections grow to two, and each remap is an identity mapping.
// Data choice: only the second half of each composite key varies, so a merge
// that considered just VID or just CIDR would fail here and nowhere else.
func TestMergeIPAMInsertsWhenNaturalKeyDiffers(t *testing.T) {
	firstLocation, secondLocation := uuid.New(), uuid.New()
	existingVLAN, existingPrefix := uuid.New(), uuid.New()

	inventory := NewInventory()
	inventory.VLANs[existingVLAN] = &CaniVLAN{ID: existingVLAN, VID: 100, Location: firstLocation}
	inventory.Prefixes[existingPrefix] = &CaniPrefix{ID: existingPrefix, Prefix: "10.0.0.0/24", VRF: "blue"}

	incomingVLAN, incomingPrefix := uuid.New(), uuid.New()
	vlanRemap := inventory.MergeVLANs(map[uuid.UUID]*CaniVLAN{
		incomingVLAN: {ID: incomingVLAN, VID: 100, Location: secondLocation},
	})
	prefixRemap := inventory.MergePrefixes(map[uuid.UUID]*CaniPrefix{
		incomingPrefix: {ID: incomingPrefix, Prefix: "10.0.0.0/24", VRF: "green"},
	})

	if len(inventory.VLANs) != 2 {
		t.Errorf("VLAN count = %d, want 2 (same VID in a different location is a different VLAN)", len(inventory.VLANs))
	}
	if len(inventory.Prefixes) != 2 {
		t.Errorf("prefix count = %d, want 2 (same CIDR in a different VRF is a different prefix)", len(inventory.Prefixes))
	}
	if vlanRemap[incomingVLAN] != incomingVLAN {
		t.Errorf("vlan remap = %v, want identity %v", vlanRemap[incomingVLAN], incomingVLAN)
	}
	if prefixRemap[incomingPrefix] != incomingPrefix {
		t.Errorf("prefix remap = %v, want identity %v", prefixRemap[incomingPrefix], incomingPrefix)
	}
}

// TestMergeIPAMSkipsNilEntries verifies a nil value in the incoming map is
// ignored rather than inserted or dereferenced.
//
// Why it matters: these maps come from provider Transform output, and a
// provider that allocates a key before deciding it has nothing to put there
// would otherwise panic the merge or plant a nil that every later traversal has
// to defend against.
// Inputs: a one-entry map of each IPAM type whose only value is nil.
// Outputs: no panic, an empty collection and an empty remap.
// Data choice: nil is the only value that distinguishes "skipped" from
// "inserted", since any non-nil object would legitimately be added.
func TestMergeIPAMSkipsNilEntries(t *testing.T) {
	inventory := NewInventory()
	id := uuid.New()

	vlanRemap := inventory.MergeVLANs(map[uuid.UUID]*CaniVLAN{id: nil})
	prefixRemap := inventory.MergePrefixes(map[uuid.UUID]*CaniPrefix{id: nil})
	addressRemap := inventory.MergeIPAddresses(map[uuid.UUID]*CaniIPAddress{id: nil})
	vrfRemap := inventory.MergeVRFs(map[uuid.UUID]*CaniVRF{id: nil})

	for _, check := range []struct {
		name        string
		size, remap int
	}{
		{"vlan", len(inventory.VLANs), len(vlanRemap)},
		{"prefix", len(inventory.Prefixes), len(prefixRemap)},
		{"address", len(inventory.IPAddresses), len(addressRemap)},
		{"vrf", len(inventory.VRFs), len(vrfRemap)},
	} {
		if check.size != 0 {
			t.Errorf("%s count = %d, want 0 (nil entry must not be inserted)", check.name, check.size)
		}
		if check.remap != 0 {
			t.Errorf("%s remap size = %d, want 0", check.name, check.remap)
		}
	}
}
