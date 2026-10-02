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
	"strings"
	"testing"

	"github.com/google/uuid"
)

// TestMergeIPAddressesKeysOnHostWithinNamespace verifies the same host with a
// different mask merges inside one namespace while the same host in another
// namespace stays separate, with a parented address taking its namespace from
// the parent prefix.
//
// Why it matters: this is the ticket's discriminating fixture — identical
// address text must coexist once per namespace, and a mask change must never
// create a second logical address.
// Inputs: existing 10.0.0.5/24 parented to a tenant-a prefix; incoming
// 10.0.0.5/25 parented to the same prefix and 10.0.0.5/24 parented to a
// tenant-b prefix. Outputs: two addresses total, the /25 remapped onto the
// existing ID, and the tenant-b copy inserted under its own ID.
// Data choice: the incoming addresses carry no Namespace field so only the
// parent-derived scope can separate them.
func TestMergeIPAddressesKeysOnHostWithinNamespace(t *testing.T) {
	inventory := NewInventory()
	prefixA, prefixB, existing := uuid.New(), uuid.New(), uuid.New()
	inventory.Prefixes[prefixA] = &CaniPrefix{ID: prefixA, Prefix: "10.0.0.0/24", Namespace: "tenant-a"}
	inventory.Prefixes[prefixB] = &CaniPrefix{ID: prefixB, Prefix: "10.0.0.0/24", Namespace: "tenant-b"}
	inventory.IPAddresses[existing] = &CaniIPAddress{ID: existing, Host: "10.0.0.5", Address: "10.0.0.5/24", Parent: prefixA}

	narrower, foreign := uuid.New(), uuid.New()
	remap := merged(t)(inventory.MergeIPAddresses(map[uuid.UUID]*CaniIPAddress{
		narrower: {ID: narrower, Host: "10.0.0.5", Address: "10.0.0.5/25", MaskLength: 25, Parent: prefixA},
		foreign:  {ID: foreign, Host: "10.0.0.5", Address: "10.0.0.5/24", Parent: prefixB},
	}))

	if len(inventory.IPAddresses) != 2 {
		t.Fatalf("address count = %d, want 2 (one per namespace)", len(inventory.IPAddresses))
	}
	if remap[narrower] != existing {
		t.Errorf("narrower remap = %v, want existing %v (mask must not split identity)", remap[narrower], existing)
	}
	if remap[foreign] != foreign {
		t.Errorf("foreign remap = %v, want identity %v (other namespace must stay separate)", remap[foreign], foreign)
	}
	if got := inventory.IPAddresses[existing].MaskLength; got != 25 {
		t.Errorf("merged mask = %d, want 25 (incoming attributes win)", got)
	}
}

// TestMergeIPAMPrefersSharedExternalIDOverNaturalKey verifies an incoming
// object that shares a provider external ID merges onto that object even when
// its natural key changed.
//
// Why it matters: a prefix re-masked or re-scoped in the source system is
// still the same object; identity must follow the source UUID, not the text.
// Inputs: an existing Global 10.0.0.0/24 with nautobot external ID X and an
// incoming tenant-a 10.0.0.0/25 carrying the same X. Outputs: one prefix,
// remapped onto the existing ID, now carrying the incoming scope and CIDR.
// Data choice: both the namespace and the CIDR differ so neither half of the
// natural key can explain the match.
func TestMergeIPAMPrefersSharedExternalIDOverNaturalKey(t *testing.T) {
	inventory := NewInventory()
	existing, source := uuid.New(), uuid.New()
	inventory.Prefixes[existing] = &CaniPrefix{
		ID: existing, Prefix: "10.0.0.0/24", ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": source}},
	}

	incoming := uuid.New()
	remap := merged(t)(inventory.MergePrefixes(map[uuid.UUID]*CaniPrefix{
		incoming: {
			ID: incoming, Prefix: "10.0.0.0/25", Namespace: "tenant-a",
			ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": source}},
		},
	}))

	if len(inventory.Prefixes) != 1 || remap[incoming] != existing {
		t.Fatalf("prefixes = %d, remap = %v; want 1 prefix remapped onto %v", len(inventory.Prefixes), remap[incoming], existing)
	}
	if got := inventory.Prefixes[existing]; got.Prefix != "10.0.0.0/25" || got.Namespace != "tenant-a" {
		t.Errorf("merged prefix = %s in %q, want 10.0.0.0/25 in tenant-a", got.Prefix, got.Namespace)
	}
}

// TestMergeIPAMRejectsConflictingAndAmbiguousIdentity verifies a natural-key
// match with a different source identity, and a key shared by several
// existing records, both return errors instead of picking a winner.
//
// Why it matters: silently overwriting a prefix that Nautobot knows under a
// different UUID, or merging onto an arbitrary legacy duplicate, would corrupt
// the mapping that every later export relies on.
// Inputs: an existing Global 10.0.0.0/24 with external ID X merged with the
// same key carrying external ID Y; and two existing Global VRFs named "red"
// merged with a third "red". Outputs: an error mentioning source identity, an
// error mentioning ambiguity, and unchanged collections.
// Data choice: VRFs supply the ambiguity case because same-named VRFs are the
// legacy shape the ticket calls out explicitly.
func TestMergeIPAMRejectsConflictingAndAmbiguousIdentity(t *testing.T) {
	inventory := NewInventory()
	existing := uuid.New()
	inventory.Prefixes[existing] = &CaniPrefix{
		ID: existing, Prefix: "10.0.0.0/24", ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": uuid.New()}},
	}
	redOne, redTwo := uuid.New(), uuid.New()
	inventory.VRFs[redOne] = &CaniVRF{ID: redOne, Name: "red"}
	inventory.VRFs[redTwo] = &CaniVRF{ID: redTwo, Name: "red"}

	_, err := inventory.MergePrefixes(map[uuid.UUID]*CaniPrefix{
		uuid.New(): {Prefix: "10.0.0.0/24", ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": uuid.New()}}},
	})
	if err == nil || !strings.Contains(err.Error(), "different source identity") {
		t.Errorf("MergePrefixes(conflict) error = %v, want source-identity conflict", err)
	}
	_, err = inventory.MergeVRFs(map[uuid.UUID]*CaniVRF{uuid.New(): {Name: "red"}})
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("MergeVRFs(duplicates) error = %v, want ambiguity error", err)
	}
	if len(inventory.Prefixes) != 1 || len(inventory.VRFs) != 2 {
		t.Errorf("collections changed on error: prefixes = %d, vrfs = %d", len(inventory.Prefixes), len(inventory.VRFs))
	}
}

// TestMergeIPAMDeduplicatesWithinOneBatch verifies two incoming objects that
// share a natural key collapse onto one record during a single merge.
//
// Why it matters: a provider may emit the same prefix twice in one transform
// (for example from two location assignments); the index must see the first
// insert so the second does not slip past as a new record.
// Inputs: an empty inventory and a batch of two Global 10.0.0.0/24 prefixes.
// Outputs: one prefix and both incoming IDs remapped to the same UUID.
// Data choice: an empty inventory isolates the in-batch path from any
// pre-existing match.
func TestMergeIPAMDeduplicatesWithinOneBatch(t *testing.T) {
	inventory := NewInventory()
	first, second := uuid.New(), uuid.New()

	remap := merged(t)(inventory.MergePrefixes(map[uuid.UUID]*CaniPrefix{
		first:  {ID: first, Prefix: "10.0.0.0/24"},
		second: {ID: second, Prefix: "10.0.0.0/24"},
	}))

	if len(inventory.Prefixes) != 1 {
		t.Fatalf("prefix count = %d, want 1", len(inventory.Prefixes))
	}
	if remap[first] != remap[second] {
		t.Errorf("remaps differ (%v, %v); both incoming IDs must resolve to one record", remap[first], remap[second])
	}
}
