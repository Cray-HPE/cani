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
	"slices"
	"testing"

	"github.com/google/uuid"
)

// TestObjectMetaMergeCarriesSetFields verifies every set ObjectMeta field is
// carried into the receiver and the merge reports the change.
//
// Why it matters: devices, modules, locations and racks all merge through this
// one function, so a field it skips is lost on every re-import.
// Inputs: an empty receiver and an incoming meta with every field set.
// Expected: every field arrives and Merge returns true.
func TestObjectMetaMergeCarriesSetFields(t *testing.T) {
	ext := uuid.New()
	m := ObjectMeta{}
	in := ObjectMeta{
		Status: "active", Role: "compute", Tenant: "t1", Tags: []string{"a"},
		CustomFields:     map[string]any{"x": "y"},
		ExternalIDs:      map[string]uuid.UUID{"nautobot": ext},
		ProviderMetadata: map[string]any{"csm": map[string]any{"xname": "x1"}},
	}

	changed := m.Merge(in)

	if !changed {
		t.Fatal("expected a change to be reported")
	}
	if m.Status != "active" || m.Role != "compute" || m.Tenant != "t1" || !slices.Equal(m.Tags, []string{"a"}) {
		t.Errorf("scalar fields not merged: %+v", m)
	}
	if m.CustomFields["x"] != "y" || m.ExternalIDs["nautobot"] != ext {
		t.Errorf("maps not merged: %+v", m)
	}
	if sub, _ := m.ProviderMetadata["csm"].(map[string]any); sub["xname"] != "x1" {
		t.Errorf("provider metadata not merged: %+v", m.ProviderMetadata)
	}
}

// TestObjectMetaMergeKeepsAndOverwritesKeys verifies custom fields merge key
// by key: an incoming key overwrites, an absent key is kept.
//
// Why it matters: a value changed in the source system must replace the local
// one without wiping custom fields the source did not send.
// Inputs: a receiver with keep=1 and x=old; incoming x=new.
// Expected: keep=1, x=new, change reported.
func TestObjectMetaMergeKeepsAndOverwritesKeys(t *testing.T) {
	m := ObjectMeta{CustomFields: map[string]any{"keep": 1, "x": "old"}}

	changed := m.Merge(ObjectMeta{CustomFields: map[string]any{"x": "new"}})

	if !changed || m.CustomFields["x"] != "new" || m.CustomFields["keep"] != 1 {
		t.Errorf("changed=%v custom fields=%v", changed, m.CustomFields)
	}
}

// TestObjectMetaMergeRetiresSupersededProviderKeys verifies a custom field
// reported by a source system removes the same key from that system's
// provider-metadata bucket and nothing else.
//
// Why it matters: a value first set with --metadata lives in the provider
// bucket and wins on export, so an edit made in the source system and read
// back by import was reverted by the next export.
// Inputs: a receiver whose nautobot bucket holds vsx_role=secondary and
// role, and whose csm bucket holds vsx_role too; incoming identified by a
// nautobot external ID with custom field vsx_role=primary.
// Expected: custom field primary; the nautobot bucket keeps only role; the
// csm bucket is untouched; a change is reported. A second receiver whose
// bucket held only the superseded key loses the bucket.
func TestObjectMetaMergeRetiresSupersededProviderKeys(t *testing.T) {
	in := ObjectMeta{
		CustomFields: map[string]any{"vsx_role": "primary"},
		ExternalIDs:  map[string]uuid.UUID{"nautobot": uuid.New()},
	}
	m := ObjectMeta{ProviderMetadata: map[string]any{
		"nautobot": map[string]any{"vsx_role": "secondary", "role": "LeafSwitch"},
		"csm":      map[string]any{"vsx_role": "kept"},
	}}

	if !m.Merge(in) {
		t.Fatal("expected a change to be reported")
	}
	if m.CustomFields["vsx_role"] != "primary" {
		t.Errorf("custom field = %v, want primary", m.CustomFields["vsx_role"])
	}
	nb, _ := m.ProviderMetadata["nautobot"].(map[string]any)
	if _, stale := nb["vsx_role"]; stale || nb["role"] != "LeafSwitch" {
		t.Errorf("nautobot bucket = %v, want only role", nb)
	}
	if csm, _ := m.ProviderMetadata["csm"].(map[string]any); csm["vsx_role"] != "kept" {
		t.Errorf("csm bucket = %v, want it untouched", csm)
	}

	only := ObjectMeta{ProviderMetadata: map[string]any{"nautobot": map[string]any{"vsx_role": "secondary"}}}
	only.Merge(in)
	if _, left := only.ProviderMetadata["nautobot"]; left {
		t.Errorf("emptied bucket was kept: %v", only.ProviderMetadata)
	}
}

// TestObjectMetaMergeNoOp verifies an empty or identical incoming meta leaves
// the receiver untouched and reports no change.
//
// Why it matters: a spurious change marks every re-import as dirty, and an
// empty value must never clear what the inventory already holds.
// Inputs: a populated receiver merged with an empty meta, then with itself.
// Expected: no change reported either time; values intact.
func TestObjectMetaMergeNoOp(t *testing.T) {
	m := ObjectMeta{Status: "active", Tags: []string{"a"}, CustomFields: map[string]any{"x": "y"}}
	same := ObjectMeta{Status: "active", Tags: []string{"a"}, CustomFields: map[string]any{"x": "y"}}

	emptyChanged := m.Merge(ObjectMeta{})
	sameChanged := m.Merge(same)

	if emptyChanged || sameChanged {
		t.Errorf("expected no change, got empty=%v same=%v", emptyChanged, sameChanged)
	}
	if m.Status != "active" || m.CustomFields["x"] != "y" || len(m.Tags) != 1 {
		t.Errorf("receiver altered: %+v", m)
	}
}
