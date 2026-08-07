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
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
)

// -----------------------------------------------------------------------------
// compareCustomFields — pure logic
// -----------------------------------------------------------------------------

func TestCompareCustomFields_DetectsChangedValue(t *testing.T) {
	local := map[string]any{"tier": "gold"}
	remote := map[string]interface{}{"tier": "silver"}
	diffs := compareCustomFields(local, &remote)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Field != "custom_field:tier" {
		t.Errorf("Field = %q, want custom_field:tier", diffs[0].Field)
	}
	if diffs[0].LocalVal != "gold" || diffs[0].RemoteVal != "silver" {
		t.Errorf("values = %q/%q, want gold/silver", diffs[0].LocalVal, diffs[0].RemoteVal)
	}
}

func TestCompareCustomFields_DetectsNewKey(t *testing.T) {
	local := map[string]any{"env": "prod"}
	remote := map[string]interface{}{}
	diffs := compareCustomFields(local, &remote)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].RemoteVal != "(none)" {
		t.Errorf("RemoteVal = %q, want (none)", diffs[0].RemoteVal)
	}
}

func TestCompareCustomFields_NoDiffWhenEqual(t *testing.T) {
	local := map[string]any{"tier": "gold", "env": "prod"}
	remote := map[string]interface{}{"tier": "gold", "env": "prod"}
	diffs := compareCustomFields(local, &remote)
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs, got %d: %v", len(diffs), diffs)
	}
}

func TestCompareCustomFields_NilRemote(t *testing.T) {
	local := map[string]any{"tier": "gold"}
	diffs := compareCustomFields(local, nil)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff for nil remote, got %d", len(diffs))
	}
}

func TestCompareCustomFields_EmptyLocal(t *testing.T) {
	remote := map[string]interface{}{"tier": "gold"}
	diffs := compareCustomFields(nil, &remote)
	if diffs != nil {
		t.Errorf("expected nil for empty local, got %v", diffs)
	}
}

// -----------------------------------------------------------------------------
// sortedMapKeys — pure logic
// -----------------------------------------------------------------------------

func TestSortedMapKeys_DeterministicOrder(t *testing.T) {
	m := map[string]any{"z": 1, "a": 2, "m": 3}
	keys := sortedMapKeys(m)
	if len(keys) != 3 || keys[0] != "a" || keys[1] != "m" || keys[2] != "z" {
		t.Errorf("sortedMapKeys = %v, want [a m z]", keys)
	}
}

func TestSortedMapKeys_EmptyMap(t *testing.T) {
	keys := sortedMapKeys(map[string]any{})
	if len(keys) != 0 {
		t.Errorf("expected empty slice, got %v", keys)
	}
}

// -----------------------------------------------------------------------------
// compareDeviceFields — custom field integration
// -----------------------------------------------------------------------------

func TestCompareDeviceFields_IncludesCustomFieldDiffs(t *testing.T) {
	device := &devicetypes.CaniDeviceType{
		ObjectMeta: devicetypes.ObjectMeta{
			ProviderMetadata: map[string]any{
				"nautobot": map[string]any{"rack_elevation": "top"},
			},
		},
	}

	remoteCF := map[string]interface{}{"rack_elevation": "bottom"}
	remote := &nautobotapi.Device{
		DeviceType:   nautobotapi.BulkWritableCableRequestStatus{},
		Location:     nautobotapi.BulkWritableCableRequestStatus{},
		Status:       nautobotapi.BulkWritableCableRequestStatus{},
		CustomFields: &remoteCF,
	}

	// Use a real mapper with empty cache so resolvers return nil (not panic)
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, 200, emptyListJSON))
	defer cleanup()
	mapper := NewDeviceMapper(e.Cache, &MapperOpts{})

	diffs := compareDeviceFields(device, remote, mapper)
	found := false
	for _, d := range diffs {
		if d.Field == "custom_field:rack_elevation" {
			found = true
			if d.LocalVal != "top" || d.RemoteVal != "bottom" {
				t.Errorf("custom field diff values = %q/%q", d.LocalVal, d.RemoteVal)
			}
		}
	}
	if !found {
		t.Error("compareDeviceFields did not include custom_field:rack_elevation diff")
	}
}

func TestCompareDeviceFields_NoDiffWhenCustomFieldsMatch(t *testing.T) {
	device := &devicetypes.CaniDeviceType{
		ObjectMeta: devicetypes.ObjectMeta{
			ProviderMetadata: map[string]any{
				"nautobot": map[string]any{"tier": "1"},
			},
		},
	}

	remoteCF := map[string]interface{}{"tier": "1"}
	remote := &nautobotapi.Device{
		DeviceType:   nautobotapi.BulkWritableCableRequestStatus{},
		Location:     nautobotapi.BulkWritableCableRequestStatus{},
		Status:       nautobotapi.BulkWritableCableRequestStatus{},
		CustomFields: &remoteCF,
	}

	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, 200, emptyListJSON))
	defer cleanup()
	mapper := NewDeviceMapper(e.Cache, &MapperOpts{})

	diffs := compareDeviceFields(device, remote, mapper)
	for _, d := range diffs {
		if d.Field == "custom_field:tier" {
			t.Errorf("unexpected diff for tier: %+v", d)
		}
	}
}
