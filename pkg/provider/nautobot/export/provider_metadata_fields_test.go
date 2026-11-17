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
	"strings"
	"testing"
)

// TestParseProviderMetadataFields_ReadsBothEntryForms verifies an entry
// names a provider and a key, and optionally the custom field to send it as.
//
// Why it matters: the list is typed by hand into a config file; a misread
// entry would send the wrong provider's value, or send it under a key
// Nautobot has no definition for and drops.
// Inputs: "csm.xname", a renamed "csm.nid=node_id", a key that itself
// holds a dot, an entry with spaces around it and around its "=", and an
// entry listed twice.
// Outputs: the provider, key and custom-field key of each, once.
// Data choice: the dotted key proves only the first dot ends the provider,
// since a provider slug has no dot and a key may; the spaced rename is how
// the entry reads most naturally in a YAML list.
func TestParseProviderMetadataFields_ReadsBothEntryForms(t *testing.T) {
	// Act.
	fields, err := ParseProviderMetadataFields([]string{
		"csm.xname", "csm.nid=node_id", "redfish.bmc.fqdn=bmc_fqdn", " hpcm.hpcm_uuid ",
		"csm.sub_role = subrole", "csm.xname",
	})

	// Assert.
	if err != nil {
		t.Fatalf("ParseProviderMetadataFields: %v", err)
	}
	want := []ProviderMetadataField{
		{Provider: "csm", Key: "xname", CustomField: "xname"},
		{Provider: "csm", Key: "nid", CustomField: "node_id"},
		{Provider: "redfish", Key: "bmc.fqdn", CustomField: "bmc_fqdn"},
		{Provider: "hpcm", Key: "hpcm_uuid", CustomField: "hpcm_uuid"},
		{Provider: "csm", Key: "sub_role", CustomField: "subrole"},
	}
	if !reflect.DeepEqual(fields, want) {
		t.Errorf("fields = %+v\nwant %+v", fields, want)
	}
}

// TestParseProviderMetadataFields_RefusesWhatItCannotSend verifies an entry
// with no provider, no key or no custom-field key is an error that quotes
// it, and that two keys cannot be sent as one custom field.
//
// Why it matters: an entry that was skipped would leave a field the
// operator asked for out of every export without a word, and two sources
// for one custom field would make the value depend on their order.
// Inputs: one malformed list per case.
// Outputs: an error that says what it is about and quotes the entry.
// Data choice: each malformed entry is a plausible typo: a bare key, a
// missing key, a dangling "=", a key that is only a space, and a rename
// that collides with another entry's default name.
func TestParseProviderMetadataFields_RefusesWhatItCannotSend(t *testing.T) {
	tests := []struct {
		name string
		list []string
		want string
	}{
		{"key with no provider", []string{"xname"}, `"xname"`},
		{"provider with no key", []string{"csm."}, `"csm."`},
		{"no provider before the dot", []string{".xname"}, `".xname"`},
		{"rename to nothing", []string{"csm.nid="}, `"csm.nid="`},
		{"key that is only a space", []string{"csm. =node_id"}, `"csm. =node_id"`},
		{"empty entry", []string{""}, `""`},
		{"two keys sent as one custom field", []string{"csm.xname", "hpcm.geoloc=xname"},
			`csm.xname and hpcm.geoloc are both sent as custom field "xname"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act.
			fields, err := ParseProviderMetadataFields(tc.list)

			// Assert.
			if err == nil {
				t.Fatalf("ParseProviderMetadataFields(%v) = %+v, want an error", tc.list, fields)
			}
			if !strings.Contains(err.Error(), "provider metadata field") || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %q, want it to say provider metadata field and %s", err, tc.want)
			}
		})
	}
}

// TestParseProviderMetadataFields_NoEntriesSendNothing verifies an empty
// list, the default, yields an empty projection.
//
// Why it matters: the default has to mean that no provider metadata is
// exported, not an error.
// Inputs: a nil list and an empty one. Outputs: no fields and no error.
// Data choice: both, because the option arrives as either depending on
// whether the config file has the key.
func TestParseProviderMetadataFields_NoEntriesSendNothing(t *testing.T) {
	for _, list := range [][]string{nil, {}} {
		// Act.
		fields, err := ParseProviderMetadataFields(list)

		// Assert.
		if err != nil || len(fields) != 0 {
			t.Errorf("ParseProviderMetadataFields(%v) = %+v, %v; want none and no error", list, fields, err)
		}
	}
}
