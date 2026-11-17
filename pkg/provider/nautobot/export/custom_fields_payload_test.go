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
	"github.com/google/uuid"
)

// TestCustomFieldProjection_SendsCustomFieldsAndNoProviderMetadata verifies
// that, with no provider-metadata key listed, an object's custom fields are
// all an export sends, whatever its provider metadata holds.
//
// Why it matters: provider metadata used to be sent as custom fields over
// the object's own, so one value had two homes and the copy in provider
// metadata silently won; a value another provider imported could overwrite
// a field an operator set, and nothing said which was sent.
// Inputs: custom fields {owner, tier} beside a CSM bucket that holds owner
// and xname, a nautobot bucket, and a key outside any bucket.
// Outputs: exactly the two custom fields, and nothing reported as replaced.
// Data choice: "owner" is in both, so a projection that still merged
// provider metadata would change it.
func TestCustomFieldProjection_SendsCustomFieldsAndNoProviderMetadata(t *testing.T) {
	// Arrange.
	meta := &devicetypes.ObjectMeta{
		CustomFields: map[string]any{"owner": "ops", "tier": "gold"},
		ProviderMetadata: map[string]any{
			"csm":      map[string]any{"owner": "csm-owner", "xname": "x3000c0s1b0"},
			"nautobot": map[string]any{"location": "dc-9"},
			"cost":     "cc-4411",
		},
	}

	// Act.
	set := customFieldProjection(nil).of(meta)

	// Assert.
	if want := map[string]any{"owner": "ops", "tier": "gold"}; !reflect.DeepEqual(set.values, want) {
		t.Errorf("values = %v, want %v", set.values, want)
	}
	if len(set.overridden) != 0 {
		t.Errorf("overridden = %v, want none", set.overridden)
	}
}

// TestCustomFieldProjection_SendsListedProviderKeys verifies that a
// provider-metadata key the operator listed is sent as a custom field, under
// its own name or the one given, and that no unlisted key comes with it.
//
// Why it matters: listing a key is the only way another provider's metadata
// reaches Nautobot, so it has to send exactly the listed key of exactly the
// listed provider; and because the listed key belongs to its provider, its
// value has to win over the copy an import read back, or the first import
// would freeze the field at its old value.
// Inputs: an object with custom fields, CSM and Redfish buckets and a key
// outside any bucket, and one list per case.
// Outputs: the values sent, and the listed keys whose value replaced a
// different custom field.
// Data choice: the Redfish bucket holds the same key as the CSM one, a
// provider with no bucket is asked for a key the other two hold, and the
// agreeing custom field is the float64 a JSON import yields for CSM's int.
func TestCustomFieldProjection_SendsListedProviderKeys(t *testing.T) {
	tests := []struct {
		name           string
		list           []string
		custom         map[string]any
		want           map[string]any
		wantOverridden []string
	}{
		{
			name: "listed key under its own name",
			list: []string{"csm.xname"},
			want: map[string]any{"tier": "gold", "xname": "x3000c0s1b0"},
		},
		{
			name: "listed key under another name",
			list: []string{"csm.nid=node_id"},
			want: map[string]any{"tier": "gold", "node_id": 1001},
		},
		{
			name: "only the listed provider's copy of a key",
			list: []string{"redfish.owner"},
			want: map[string]any{"tier": "gold", "owner": "bmc-team"},
		},
		{
			name: "key the object's provider metadata does not hold",
			list: []string{"hpcm.hpcm_uuid", "csm.subRole"},
			want: map[string]any{"tier": "gold"},
		},
		{
			name: "key held only by another provider or outside any bucket",
			list: []string{"hpcm.owner", "csm.cost"},
			want: map[string]any{"tier": "gold"},
		},
		{
			name:           "replaces a different custom field and says so",
			list:           []string{"csm.owner"},
			custom:         map[string]any{"owner": "edited-in-nautobot"},
			want:           map[string]any{"tier": "gold", "owner": "csm-team"},
			wantOverridden: []string{"owner"},
		},
		{
			name:   "custom field an import read back agrees",
			list:   []string{"csm.nid=node_id"},
			custom: map[string]any{"node_id": float64(1001)},
			want:   map[string]any{"tier": "gold", "node_id": 1001},
		},
		{
			name:   "custom field an import read back as unset is not a different value",
			list:   []string{"csm.nid=node_id"},
			custom: map[string]any{"node_id": nil},
			want:   map[string]any{"tier": "gold", "node_id": 1001},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			fields, err := ParseProviderMetadataFields(tc.list)
			if err != nil {
				t.Fatalf("ParseProviderMetadataFields(%v): %v", tc.list, err)
			}
			meta := &devicetypes.ObjectMeta{
				CustomFields: map[string]any{"tier": "gold"},
				ProviderMetadata: map[string]any{
					"csm":     map[string]any{"owner": "csm-team", "xname": "x3000c0s1b0", "nid": 1001},
					"redfish": map[string]any{"owner": "bmc-team"},
					"cost":    "cc-4411",
				},
			}
			for key, value := range tc.custom {
				meta.CustomFields[key] = value
			}

			// Act.
			set := customFieldProjection(fields).of(meta)

			// Assert.
			if !reflect.DeepEqual(set.values, tc.want) {
				t.Errorf("values = %#v\nwant %#v", set.values, tc.want)
			}
			var overridden []string
			for _, field := range set.overridden {
				overridden = append(overridden, field.CustomField)
			}
			if !reflect.DeepEqual(overridden, tc.wantOverridden) {
				t.Errorf("overridden = %v, want %v", overridden, tc.wantOverridden)
			}
		})
	}
}

// TestCustomFieldProjection_LeavesTheObjectsOwnFieldsAlone verifies that
// projecting an object with a listed provider key does not write that key
// into the object's custom fields.
//
// Why it matters: the projection runs on every create, merge and
// comparison; if it wrote into the inventory, an export would turn provider
// metadata into custom fields that the next save persists.
// Inputs: an object with one custom field and a listed CSM key.
// Outputs: the object's custom fields unchanged after the projection.
// Data choice: the object already has a custom-field map, the case in which
// a projection that skipped the copy would write into it.
func TestCustomFieldProjection_LeavesTheObjectsOwnFieldsAlone(t *testing.T) {
	// Arrange.
	meta := &devicetypes.ObjectMeta{
		CustomFields:     map[string]any{"tier": "gold"},
		ProviderMetadata: map[string]any{"csm": map[string]any{"xname": "x3000c0s1b0"}},
	}
	projection := customFieldProjection{{Provider: "csm", Key: "xname", CustomField: "xname"}}

	// Act.
	projection.payload(kindDevice, "cn-01", meta)

	// Assert.
	if want := map[string]any{"tier": "gold"}; !reflect.DeepEqual(meta.CustomFields, want) {
		t.Errorf("custom fields = %v, want them untouched: %v", meta.CustomFields, want)
	}
}

// TestCustomFieldProjection_KeepsFalseZeroEmptyAndNullValues verifies that
// false, 0, "" and an explicit null survive the projection and reach the
// request payload.
//
// Why it matters: Nautobot fills an omitted key from the definition's
// default, so dropping a falsy value would silently change it to the default
// on a fresh destination; and an explicit null is how a field is cleared, so
// dropping it would leave the old value in place.
// Inputs: custom fields {active: false, slots: 0, note: "", retired: nil}.
// Outputs: a payload carrying all four keys with their original values.
// Data choice: one value per falsy Go kind that a len() or truthiness check
// would discard, and the nil --clear-custom-field stores.
func TestCustomFieldProjection_KeepsFalseZeroEmptyAndNullValues(t *testing.T) {
	// Arrange.
	meta := &devicetypes.ObjectMeta{CustomFields: map[string]any{"active": false, "slots": 0, "note": "", "retired": nil}}

	// Act.
	payload := customFieldProjection(nil).of(meta).payload()

	// Assert.
	if payload == nil {
		t.Fatal("payload = nil, want four keys")
	}
	for key, want := range map[string]any{"active": false, "slots": 0, "note": "", "retired": nil} {
		got, ok := (*payload)[key]
		if !ok || got == nil || *got != want {
			t.Errorf("payload[%q] = %v, want a pointer to %v", key, got, want)
		}
	}
}

// TestCustomFieldProjection_SendsTheValueImportReadBack verifies that after
// an import refreshes a custom field, the projection sends the imported
// value even though another provider's metadata holds the same key.
//
// Why it matters: this is the round trip that reverted an edit made in
// Nautobot. A stale copy of the key in provider metadata outlived the import
// and won on the next --merge export.
// Inputs: a device whose custom field and CSM bucket both hold vsx_role,
// merged with an imported record carrying the value edited in Nautobot.
// Outputs: the imported value, with nothing reported as replaced.
// Data choice: the stale copy is in another provider's bucket, the case the
// import cannot retire because that provider did not report the field.
func TestCustomFieldProjection_SendsTheValueImportReadBack(t *testing.T) {
	// Arrange.
	meta := &devicetypes.ObjectMeta{
		CustomFields:     map[string]any{"vsx_role": "secondary"},
		ProviderMetadata: map[string]any{"csm": map[string]any{"vsx_role": "secondary"}},
	}
	meta.Merge(devicetypes.ObjectMeta{
		CustomFields: map[string]any{"vsx_role": "primary"},
		ExternalIDs:  map[string]uuid.UUID{"nautobot": uuid.New()},
	})

	// Act.
	set := customFieldProjection(nil).of(meta)

	// Assert.
	if set.values["vsx_role"] != "primary" || len(set.overridden) != 0 {
		t.Errorf("values = %v, overridden = %v, want primary and none", set.values, set.overridden)
	}
}

// TestCustomFieldProjection_NothingToSendYieldsNilPayload verifies that nil
// metadata, metadata without custom fields, and metadata holding only
// unlisted provider metadata all produce a nil payload.
//
// Why it matters: a nil CustomFields pointer keeps custom_fields out of the
// request body altogether; an empty object would be read by Nautobot as a
// write.
// Inputs: a nil *ObjectMeta, an ObjectMeta with only a status, and one with
// only a CSM bucket.
// Outputs: nil payloads from all three.
// Data choice: the last case is the common one for a CSM-imported device.
func TestCustomFieldProjection_NothingToSendYieldsNilPayload(t *testing.T) {
	// Arrange.
	projection := customFieldProjection(nil)
	providerOnly := &devicetypes.ObjectMeta{ProviderMetadata: map[string]any{"csm": map[string]any{"xname": "x3000"}}}

	// Act and assert.
	if got := projection.of(nil).payload(); got != nil {
		t.Errorf("payload(nil meta) = %v, want nil", got)
	}
	if got := projection.of(&devicetypes.ObjectMeta{Status: "Active"}).payload(); got != nil {
		t.Errorf("payload(no custom fields) = %v, want nil", got)
	}
	if got := projection.payload(kindDevice, "cn-01", providerOnly); got != nil {
		t.Errorf("payload(provider metadata only) = %v, want nil", got)
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
