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

// requireAdded fails the test when arranging a record fails.
func requireAdded(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("arrange: %v", err)
	}
}

// TestAddIPAMRejectsANaturalKeyDuplicate verifies adding a prefix, address or
// VRF whose natural key another record holds fails, naming that record.
//
// Why it matters: Nautobot keeps one object per key in a namespace, so the
// second record could never be exported, and every later import would find
// the key ambiguous.
// Inputs: Global 10.0.0.0/24, 10.0.0.5/24 and VRF red; then the prefix
// 10.0.0.9/24, the address 10.0.0.5/25, and red with Global spelled out.
// Outputs: an error naming the existing record for each; nothing is added.
// Data choice: each duplicate is spelled differently from the original, so
// only the canonical natural key can catch it.
func TestAddIPAMRejectsANaturalKeyDuplicate(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	prefix, address, vrf := uuid.New(), uuid.New(), uuid.New()
	requireAdded(t, inv.AddPrefix(&CaniPrefix{ID: prefix, Prefix: "10.0.0.0/24"}))
	requireAdded(t, inv.AddIPAddress(&CaniIPAddress{ID: address, Address: "10.0.0.5/24"}))
	requireAdded(t, inv.AddVRF(&CaniVRF{ID: vrf, Name: "red"}))
	duplicates := []struct {
		name   string
		add    func() error
		holder uuid.UUID
	}{
		{"prefix", func() error { return inv.AddPrefix(&CaniPrefix{ID: uuid.New(), Prefix: "10.0.0.9/24"}) }, prefix},
		{"address", func() error { return inv.AddIPAddress(&CaniIPAddress{ID: uuid.New(), Address: "10.0.0.5/25"}) }, address},
		{"vrf", func() error { return inv.AddVRF(&CaniVRF{ID: uuid.New(), Name: "red", Namespace: "Global"}) }, vrf},
	}
	for _, tc := range duplicates {
		t.Run(tc.name, func(t *testing.T) {
			// Act.
			err := tc.add()

			// Assert.
			want := `already exists in namespace "Global" as ` + tc.holder.String()
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Errorf("error = %v, want it to contain %q", err, want)
			}
		})
	}
	if len(inv.Prefixes) != 1 || len(inv.IPAddresses) != 1 || len(inv.VRFs) != 1 {
		t.Errorf("records = %d prefixes, %d addresses, %d VRFs; want 1 each",
			len(inv.Prefixes), len(inv.IPAddresses), len(inv.VRFs))
	}
}

// TestAddIPAMAcceptsTheSameKeyElsewhere verifies uniqueness is scoped by
// namespace and compares VRF names exactly.
//
// Why it matters: namespaces exist to hold the same network twice, and
// Nautobot treats "red" and "Red" as different VRFs.
// Inputs: Global 10.0.0.0/24, 10.0.0.5/24 and VRF red; then the same prefix,
// address and VRF in tenant-a, and VRF Red in Global.
// Outputs: every add succeeds.
// Data choice: each record repeats a Global key except for its namespace or
// the case of its name.
func TestAddIPAMAcceptsTheSameKeyElsewhere(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	requireAdded(t, inv.AddPrefix(&CaniPrefix{ID: uuid.New(), Prefix: "10.0.0.0/24"}))
	requireAdded(t, inv.AddIPAddress(&CaniIPAddress{ID: uuid.New(), Address: "10.0.0.5/24"}))
	requireAdded(t, inv.AddVRF(&CaniVRF{ID: uuid.New(), Name: "red"}))

	// Act.
	errs := []error{
		inv.AddPrefix(&CaniPrefix{ID: uuid.New(), Prefix: "10.0.0.0/24", Namespace: "tenant-a"}),
		inv.AddIPAddress(&CaniIPAddress{ID: uuid.New(), Address: "10.0.0.5/24", Namespace: "tenant-a"}),
		inv.AddVRF(&CaniVRF{ID: uuid.New(), Name: "red", Namespace: "tenant-a"}),
		inv.AddVRF(&CaniVRF{ID: uuid.New(), Name: "Red"}),
	}

	// Assert.
	for i, err := range errs {
		if err != nil {
			t.Errorf("add %d returned error: %v", i, err)
		}
	}
	if len(inv.Prefixes) != 2 || len(inv.IPAddresses) != 2 || len(inv.VRFs) != 3 {
		t.Errorf("records = %d prefixes, %d addresses, %d VRFs; want 2, 2, 3",
			len(inv.Prefixes), len(inv.IPAddresses), len(inv.VRFs))
	}
}

// TestAddPrefixAdoptsWhatItHoldsMostClosely verifies a new prefix becomes the
// parent of each prefix and address in its namespace that it holds more
// closely than their current parent, and leaves the rest alone.
//
// Why it matters: adding the containing prefix is the fix for an address
// without a parent, and Nautobot re-parents the same way on create.
// Inputs: Global 10.0.0.0/8 holding 10.1.2.0/24 (which holds 10.1.2.0/28 and
// 10.1.2.9), 10.2.0.0/24, 10.1.0.5 and 10.2.0.1; orphans 10.1.3.7 in Global
// and in tenant-a; a tenant-a root 10.1.5.0/24. Then 10.1.0.0/16 is added in
// Global.
// Outputs: the /24, 10.1.0.5 and the Global orphan move to the /16; the /28
// and 10.1.2.9 keep the closer /24; 10.2.0.0/24 and 10.2.0.1, outside the
// /16, keep the /8; the tenant-a records keep no parent.
// Data choice: each record covers one rule: closer, already closer, outside,
// orphaned, and other namespace.
func TestAddPrefixAdoptsWhatItHoldsMostClosely(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	slash8, slash24, slash28, outside, tenant := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.Prefixes[slash8] = &CaniPrefix{ID: slash8, Prefix: "10.0.0.0/8"}
	inv.Prefixes[slash24] = &CaniPrefix{ID: slash24, Prefix: "10.1.2.0/24", Parent: slash8}
	inv.Prefixes[slash28] = &CaniPrefix{ID: slash28, Prefix: "10.1.2.0/28", Parent: slash24}
	inv.Prefixes[outside] = &CaniPrefix{ID: outside, Prefix: "10.2.0.0/24", Parent: slash8}
	inv.Prefixes[tenant] = &CaniPrefix{ID: tenant, Prefix: "10.1.5.0/24", Namespace: "tenant-a"}
	addresses := map[string]*CaniIPAddress{
		"under the /8":    {Host: "10.1.0.5", Parent: slash8},
		"under the /24":   {Host: "10.1.2.9", Parent: slash24},
		"outside the /16": {Host: "10.2.0.1", Parent: slash8},
		"Global orphan":   {Host: "10.1.3.7"},
		"tenant-a orphan": {Host: "10.1.3.7", Namespace: "tenant-a"},
	}
	for _, address := range addresses {
		address.ID = uuid.New()
		inv.IPAddresses[address.ID] = address
	}
	slash16 := &CaniPrefix{ID: uuid.New(), Prefix: "10.1.0.0/16"}

	// Act.
	err := inv.AddPrefix(slash16)

	// Assert.
	if err != nil {
		t.Fatalf("AddPrefix returned error: %v", err)
	}
	want := map[string]uuid.UUID{
		"the /16": slash8, "the /24": slash16.ID, "the /28": slash24, "the /24 outside": slash8,
		"under the /8": slash16.ID, "under the /24": slash24, "outside the /16": slash8,
		"Global orphan": slash16.ID, "tenant-a orphan": uuid.Nil, "the tenant-a /24": uuid.Nil,
	}
	got := map[string]uuid.UUID{
		"the /16": slash16.Parent, "the /24": inv.Prefixes[slash24].Parent, "the /28": inv.Prefixes[slash28].Parent,
		"the /24 outside": inv.Prefixes[outside].Parent, "the tenant-a /24": inv.Prefixes[tenant].Parent,
	}
	for name, address := range addresses {
		got[name] = address.Parent
	}
	for name, parent := range want {
		if got[name] != parent {
			t.Errorf("parent of %s = %v, want %v", name, got[name], parent)
		}
	}
}

// TestMergeTransformResultAdoptsLikeAddPrefix verifies a prefix a merge
// inserts re-parents the records it holds, as AddPrefix does, so the tree does
// not depend on whether a prefix was added or imported.
//
// Why it matters: an orphan address is fixed by supplying its containing
// prefix; an import that brings the prefix must fix it too, and a matched
// prefix must not disturb anything.
// Inputs: Global 10.0.0.0/8 holding 10.1.0.5; orphans 10.1.3.7 in Global and
// in tenant-a; a transform result carrying the same /8 and a new Global
// 10.1.0.0/16.
// Outputs: the /16 sits under the /8; 10.1.0.5 and the Global orphan move to
// it; the tenant-a orphan keeps no parent; the inventory holds two prefixes.
// Data choice: the records mirror the AddPrefix test so both paths are held
// to one rule, and the re-sent /8 is a match rather than an insert.
func TestMergeTransformResultAdoptsLikeAddPrefix(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	slash8, held, globalOrphan, tenantOrphan := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	requireAdded(t, inv.AddPrefix(&CaniPrefix{ID: slash8, Prefix: "10.0.0.0/8"}))
	inv.IPAddresses[held] = &CaniIPAddress{ID: held, Host: "10.1.0.5", Address: "10.1.0.5/8", Parent: slash8}
	inv.IPAddresses[globalOrphan] = &CaniIPAddress{ID: globalOrphan, Host: "10.1.3.7", Address: "10.1.3.7/32"}
	inv.IPAddresses[tenantOrphan] = &CaniIPAddress{ID: tenantOrphan, Host: "10.1.3.7", Address: "10.1.3.7/32", Namespace: "tenant-a"}
	resent := &CaniPrefix{ID: slash8, Prefix: "10.0.0.0/8"}
	slash16 := &CaniPrefix{ID: uuid.New(), Prefix: "10.1.0.0/16"}
	requireAdded(t, ParsePrefix(resent))
	requireAdded(t, ParsePrefix(slash16))
	result := &TransformResult{Prefixes: map[uuid.UUID]*CaniPrefix{resent.ID: resent, slash16.ID: slash16}}

	// Act.
	_, err := inv.MergeTransformResult(result)

	// Assert.
	if err != nil {
		t.Fatalf("MergeTransformResult returned error: %v", err)
	}
	if len(inv.Prefixes) != 2 {
		t.Errorf("prefixes = %d, want 2 (the re-sent /8 matched)", len(inv.Prefixes))
	}
	want := map[string]uuid.UUID{
		"the /16": slash8, "held address": slash16.ID, "Global orphan": slash16.ID, "tenant-a orphan": uuid.Nil,
	}
	got := map[string]uuid.UUID{
		"the /16": inv.Prefixes[slash16.ID].Parent, "held address": inv.IPAddresses[held].Parent,
		"Global orphan": inv.IPAddresses[globalOrphan].Parent, "tenant-a orphan": inv.IPAddresses[tenantOrphan].Parent,
	}
	for name, parent := range want {
		if got[name] != parent {
			t.Errorf("parent of %s = %v, want %v", name, got[name], parent)
		}
	}
}

// TestMergeTransformResultKeepsTheParentsTheBatchCarries verifies a merge
// adopts only the records it leaves alone: a record the batch re-sends keeps
// the parent the provider set, even when an inserted prefix holds it more
// closely.
//
// Why it matters: a parent set by the provider is mapped import scope and must
// survive the merge; only pre-existing records outside the batch follow the
// local re-parenting rule.
// Inputs: Global 10.0.0.0/8 holding 10.1.2.0/24 (re-sent by the batch with
// parent /8) and 10.1.3.0/24 (not in the batch); the batch also inserts
// 10.1.0.0/16.
// Outputs: the re-sent /24 stays under the /8; the untouched /24 moves to the
// /16.
// Data choice: the two /24s differ only in whether the batch carries them, so
// that is the only thing that can explain their different parents.
func TestMergeTransformResultKeepsTheParentsTheBatchCarries(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	slash8, carried, untouched := uuid.New(), uuid.New(), uuid.New()
	requireAdded(t, inv.AddPrefix(&CaniPrefix{ID: slash8, Prefix: "10.0.0.0/8"}))
	requireAdded(t, inv.AddPrefix(&CaniPrefix{ID: carried, Prefix: "10.1.2.0/24"}))
	requireAdded(t, inv.AddPrefix(&CaniPrefix{ID: untouched, Prefix: "10.1.3.0/24"}))
	resent := &CaniPrefix{ID: carried, Prefix: "10.1.2.0/24", Parent: slash8}
	slash16 := &CaniPrefix{ID: uuid.New(), Prefix: "10.1.0.0/16"}
	requireAdded(t, ParsePrefix(resent))
	requireAdded(t, ParsePrefix(slash16))
	result := &TransformResult{Prefixes: map[uuid.UUID]*CaniPrefix{resent.ID: resent, slash16.ID: slash16}}

	// Act.
	_, err := inv.MergeTransformResult(result)

	// Assert.
	if err != nil {
		t.Fatalf("MergeTransformResult returned error: %v", err)
	}
	if got := inv.Prefixes[carried].Parent; got != slash8 {
		t.Errorf("re-sent /24 parent = %v, want the provider's /8 %v", got, slash8)
	}
	if got := inv.Prefixes[untouched].Parent; got != slash16.ID {
		t.Errorf("untouched /24 parent = %v, want the inserted /16 %v", got, slash16.ID)
	}
}
