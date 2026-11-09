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
	"testing"

	"github.com/google/uuid"
)

// TestMergeDevicesCarriesCustomFields verifies a re-imported device's custom
// fields, tags and tenant reach the existing record whether it is matched by
// UUID or by name, both when the value is new and when it changed.
//
// Why it matters: re-importing from Nautobot onto an existing inventory is the
// main round trip; before the fix MergeProperties skipped ObjectMeta and every
// custom field value set in Nautobot was dropped (PR #281 review).
// Inputs: an existing device with no custom field (or x=old), merged with an
// incoming record carrying x=y, a tag and a tenant.
// Expected: the existing record holds x=y, the tag and tenant.
func TestMergeDevicesCarriesCustomFields(t *testing.T) {
	for _, tc := range []struct {
		name     string
		byName   bool
		existing map[string]any
	}{
		{"uuid, new value", false, nil},
		{"uuid, changed value", false, map[string]any{"x": "old"}},
		{"name, new value", true, nil},
		{"name, changed value", true, map[string]any{"x": "old"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			inv := NewInventory()
			id := uuid.New()
			existing := &CaniDeviceType{ID: id, Name: "node-1"}
			existing.CustomFields = tc.existing
			inv.Devices[id] = existing
			inID := id
			if tc.byName {
				inID = uuid.New()
			}
			in := &CaniDeviceType{ID: inID, Name: "node-1"}
			in.CustomFields = map[string]any{"x": "y"}
			in.Tags = []string{"gpu"}
			in.Tenant = "t1"

			inv.MergeDevices(map[uuid.UUID]*CaniDeviceType{inID: in})

			got := inv.Devices[id]
			if got.CustomFields["x"] != "y" || got.Tenant != "t1" || len(got.Tags) != 1 {
				t.Errorf("meta not merged: %+v", got.ObjectMeta)
			}
		})
	}
}

// TestMergePropertiesReportsCustomFieldChange verifies a custom-field-only
// difference is reported as a change, and an identical one is not.
//
// Why it matters: the import decides whether to save on this flag, so a
// custom-field-only edit made in Nautobot must not be treated as a no-op.
// Inputs: a device with x=old merged with x=new, then with x=new again.
// Expected: true, then false.
func TestMergePropertiesReportsCustomFieldChange(t *testing.T) {
	c := &CaniDeviceType{Name: "n"}
	c.CustomFields = map[string]any{"x": "old"}
	other := &CaniDeviceType{Name: "n"}
	other.CustomFields = map[string]any{"x": "new"}

	first := c.MergeProperties(other)
	second := c.MergeProperties(other)

	if !first || second {
		t.Errorf("first=%v second=%v, want true then false", first, second)
	}
}
