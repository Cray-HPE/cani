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
package export

import (
	"reflect"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// explicitAndProviderMeta builds metadata whose explicit custom fields and
// provider metadata overlap on "owner" and otherwise differ.
func explicitAndProviderMeta() *devicetypes.ObjectMeta {
	return &devicetypes.ObjectMeta{
		CustomFields:     map[string]any{"owner": "explicit", "tier": "gold"},
		ProviderMetadata: map[string]any{"csm": map[string]any{"owner": "provider", "xname": "x3000c0s1b0"}},
	}
}

// TestCustomFieldsOf_ProviderMetadataWinsAndNamesTheCollision verifies that
// a key supplied by both sources takes the provider value and is listed as
// a collision, while keys from either source alone pass through.
//
// Why it matters: the mapping guide documents provider-metadata-over-explicit
// precedence; the Story requires collisions to be reported, not resolved
// differently, so an operator can see which explicit value did not arrive.
// Inputs: explicit {owner, tier} and provider {owner, xname}.
// Outputs: values {owner: provider, tier, xname} and collisions ["owner"].
// Data choice: one overlapping key and one unique key per source separate
// the precedence rule from plain pass-through.
func TestCustomFieldsOf_ProviderMetadataWinsAndNamesTheCollision(t *testing.T) {
	set := customFieldsOf(explicitAndProviderMeta())

	want := map[string]any{"owner": "provider", "tier": "gold", "xname": "x3000c0s1b0"}
	if !reflect.DeepEqual(set.values, want) {
		t.Errorf("values = %v, want %v", set.values, want)
	}
	if !reflect.DeepEqual(set.collisions, []string{"owner"}) {
		t.Errorf("collisions = %v, want [owner]", set.collisions)
	}
}

// TestCustomFieldsOf_KeepsFalseZeroAndEmptyValues verifies that explicit
// false, 0 and "" survive the projection and reach the request payload.
//
// Why it matters: Nautobot fills an omitted key from the definition's
// default, so dropping a falsy value would silently change it to the default
// on a fresh destination; the Story names false and zero explicitly.
// Inputs: explicit {active: false, slots: 0, note: ""}, no provider metadata.
// Outputs: a payload carrying all three keys with their original values.
// Data choice: one value per falsy Go kind that a len() or truthiness check
// would discard.
func TestCustomFieldsOf_KeepsFalseZeroAndEmptyValues(t *testing.T) {
	meta := &devicetypes.ObjectMeta{CustomFields: map[string]any{"active": false, "slots": 0, "note": ""}}

	payload := customFieldsOf(meta).payload()
	if payload == nil {
		t.Fatal("payload = nil, want three keys")
	}
	for key, want := range map[string]any{"active": false, "slots": 0, "note": ""} {
		got, ok := (*payload)[key]
		if !ok || got == nil || *got != want {
			t.Errorf("payload[%q] = %v, want %v", key, got, want)
		}
	}
}

// TestCustomFieldsOf_ExcludeDropsTheKeyAndItsCollision verifies that an
// excluded key is absent from the values and is not reported as a collision
// even when both sources supply it.
//
// Why it matters: the legacy rack mapper maps u_height to a dedicated field;
// it must neither be sent as a custom field nor produce a misleading warning.
// Inputs: explicit and provider both set u_height; provider also sets color.
// Outputs: values {color} only and no collisions.
// Data choice: the excluded key collides so the test proves exclusion is
// applied before collision reporting.
func TestCustomFieldsOf_ExcludeDropsTheKeyAndItsCollision(t *testing.T) {
	meta := &devicetypes.ObjectMeta{
		CustomFields:     map[string]any{"u_height": 42},
		ProviderMetadata: map[string]any{"u_height": 48, "color": "black"},
	}

	set := customFieldsOf(meta, "u_height")

	if !reflect.DeepEqual(set.values, map[string]any{"color": "black"}) {
		t.Errorf("values = %v, want {color: black}", set.values)
	}
	if len(set.collisions) != 0 {
		t.Errorf("collisions = %v, want none", set.collisions)
	}
}

// TestCustomFieldsOf_NothingToSendYieldsNilPayload verifies that nil
// metadata and metadata without custom fields both produce a nil payload.
//
// Why it matters: a nil CustomFields pointer keeps custom_fields out of the
// request body, which is how every writer behaved before the shared helper.
// Inputs: a nil *ObjectMeta and an ObjectMeta with only a status.
// Outputs: nil payloads from both.
// Data choice: the second case carries unrelated metadata so the test shows
// only custom-field sources are consulted.
func TestCustomFieldsOf_NothingToSendYieldsNilPayload(t *testing.T) {
	if got := customFieldsOf(nil).payload(); got != nil {
		t.Errorf("payload(nil meta) = %v, want nil", got)
	}
	if got := customFieldsOf(&devicetypes.ObjectMeta{Status: "Active"}).payload(); got != nil {
		t.Errorf("payload(no custom fields) = %v, want nil", got)
	}
}

// TestRecordDroppedCustomFields_CountsKeysMissingFromResponse verifies that
// each sent key absent from the response's custom_fields is counted once.
//
// Why it matters: Nautobot answers 2xx for a custom-field key with no
// definition for the object type and never exposes the value, so the
// response body is the only evidence it did not arrive; the lossy total
// must include it.
// Inputs: a payload of {tier, owner, note} and a body whose custom_fields
// holds tier and note only.
// Outputs: CustomFieldsDropped == 1.
// Data choice: two keys that arrive bracket the one that does not, so the
// count cannot be satisfied by "all or nothing" logic.
func TestRecordDroppedCustomFields_CountsKeysMissingFromResponse(t *testing.T) {
	sent := nbCF(map[string]any{"tier": "gold", "owner": "ops", "note": ""})
	body := []byte(`{"id":"a","custom_fields":{"tier":"gold","note":""}}`)
	result := &LoadResult{}

	recordDroppedCustomFields("device", "compute-001", sent, body, result)

	if result.CustomFieldsDropped != 1 {
		t.Errorf("CustomFieldsDropped = %d, want 1", result.CustomFieldsDropped)
	}
	if lossyExportCount(result) != 1 {
		t.Errorf("lossyExportCount = %d, want 1", lossyExportCount(result))
	}
}

// TestRecordDroppedCustomFields_IgnoresBodiesThatSayNothing verifies that a
// response without a custom_fields object, an empty body, or nothing sent
// never counts as a loss.
//
// Why it matters: a body that lacks custom_fields entirely (a minimal test
// double, a proxy) is not evidence of a drop; counting it would make the
// summary lie in the other direction.
// Inputs: a one-key payload with a body lacking custom_fields; the same
// payload with an empty body; a nil payload with a complete body.
// Outputs: CustomFieldsDropped stays 0 in every case.
// Data choice: the three ways the check can be inconclusive, each in
// isolation.
func TestRecordDroppedCustomFields_IgnoresBodiesThatSayNothing(t *testing.T) {
	sent := nbCF(map[string]any{"tier": "gold"})
	result := &LoadResult{}

	recordDroppedCustomFields("device", "d1", sent, []byte(`{"id":"a"}`), result)
	recordDroppedCustomFields("device", "d1", sent, nil, result)
	recordDroppedCustomFields("device", "d1", nil, []byte(`{"custom_fields":{}}`), result)

	if result.CustomFieldsDropped != 0 {
		t.Errorf("CustomFieldsDropped = %d, want 0", result.CustomFieldsDropped)
	}
}

// TestRecordDroppedCustomFields_NullValueStillArrived verifies that a key
// returned with a null value counts as arrived.
//
// Why it matters: Nautobot returns every defined key for the model, using
// null for an unset value; null means "defined, empty", not "dropped".
// Inputs: a payload {note: ""} and a body with "note": null.
// Outputs: CustomFieldsDropped == 0.
// Data choice: the empty string is the value most likely to come back as
// null after Nautobot's own normalisation.
func TestRecordDroppedCustomFields_NullValueStillArrived(t *testing.T) {
	result := &LoadResult{}

	recordDroppedCustomFields("device", "d1", nbCF(map[string]any{"note": ""}), []byte(`{"custom_fields":{"note":null}}`), result)

	if result.CustomFieldsDropped != 0 {
		t.Errorf("CustomFieldsDropped = %d, want 0", result.CustomFieldsDropped)
	}
}
