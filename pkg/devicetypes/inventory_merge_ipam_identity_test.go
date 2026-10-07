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
	"errors"
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

// TestMergeIPAMRejectsConflictingIdentity verifies a natural-key match with a
// different source identity returns ErrSourceIdentityConflict naming the
// record, both IDs and the remedy, and leaves the collection unchanged.
//
// Why it matters: Nautobot gives a deleted and recreated prefix a new UUID.
// Silently overwriting the record would corrupt the mapping every later export
// relies on, while a bare error would leave the operator with nothing to act
// on; the typed error is what the import command turns into a remove command.
// Inputs: an existing Global 10.0.0.0/24 with nautobot ID X merged with the
// same key carrying nautobot ID Y. Outputs: an error that unwraps to
// ErrSourceIdentityConflict and carries kind prefix, the record, X and Y; one
// prefix still holding X.
// Data choice: the incoming object has its own UUID, as every import mints
// one, so only the natural key can match it.
func TestMergeIPAMRejectsConflictingIdentity(t *testing.T) {
	// Arrange.
	inventory := NewInventory()
	existing, before, after := uuid.New(), uuid.New(), uuid.New()
	inventory.Prefixes[existing] = &CaniPrefix{
		ID: existing, Prefix: "10.0.0.0/24", ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": before}},
	}

	// Act.
	_, err := inventory.MergePrefixes(map[uuid.UUID]*CaniPrefix{
		uuid.New(): {Prefix: "10.0.0.0/24", ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": after}}},
	})

	// Assert.
	if !errors.Is(err, ErrSourceIdentityConflict) {
		t.Fatalf("MergePrefixes(conflict) error = %v, want ErrSourceIdentityConflict", err)
	}
	var conflict *SourceIdentityConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("MergePrefixes(conflict) error = %T, want *SourceIdentityConflictError", err)
	}
	want := &SourceIdentityConflictError{Kind: IPAMKindPrefix, Record: existing, Source: "nautobot", Existing: before, Incoming: after}
	if *conflict != *want {
		t.Errorf("conflict = %+v, want %+v", *conflict, *want)
	}
	if !strings.HasPrefix(err.Error(), `prefix 10.0.0.0/24 in namespace "Global": record `+existing.String()) {
		t.Errorf("error = %q, want it to name the prefix and the record", err)
	}
	if got := inventory.Prefixes[existing].ExternalIDs["nautobot"]; len(inventory.Prefixes) != 1 || got != before {
		t.Errorf("collection changed on error: %d prefixes, record nautobot ID %v; want 1 still holding %v", len(inventory.Prefixes), got, before)
	}
}

// TestMergeIPAMRejectsAmbiguousIdentity verifies a natural key shared by
// several existing records returns an error instead of picking a winner.
//
// Why it matters: merging onto an arbitrary legacy duplicate would corrupt
// the mapping that every later export relies on.
// Inputs: two existing Global VRFs named "red" merged with a third "red".
// Outputs: an error mentioning ambiguity and an unchanged collection.
// Data choice: same-named VRFs are the legacy shape the ticket calls out.
func TestMergeIPAMRejectsAmbiguousIdentity(t *testing.T) {
	// Arrange.
	inventory := NewInventory()
	redOne, redTwo := uuid.New(), uuid.New()
	inventory.VRFs[redOne] = &CaniVRF{ID: redOne, Name: "red"}
	inventory.VRFs[redTwo] = &CaniVRF{ID: redTwo, Name: "red"}

	// Act.
	_, err := inventory.MergeVRFs(map[uuid.UUID]*CaniVRF{uuid.New(): {Name: "red"}})

	// Assert.
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("MergeVRFs(duplicates) error = %v, want ambiguity error", err)
	}
	if errors.Is(err, ErrSourceIdentityConflict) {
		t.Errorf("MergeVRFs(duplicates) error = %v, must not read as a source identity conflict", err)
	}
	if len(inventory.VRFs) != 2 {
		t.Errorf("VRFs changed on error: %d records, want 2", len(inventory.VRFs))
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
