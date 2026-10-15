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
	"github.com/google/uuid"
)

// ---------- MapToWritableRackRequest ----------

// TestMapToWritableRackRequestNil verifies that MapToWritableRackRequest returns
// an error when given a nil device.
//
// Why it matters: rack creation must reject nil input rather than panic, so a
// missing rack record fails cleanly during export.
// Inputs: a nil device. Outputs: a non-nil error.
// Data choice: nil is the boundary case guarded before any field access.
func TestMapToWritableRackRequestNil(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	_, err := mapper.MapToWritableRackRequest(nil)
	if err == nil {
		t.Error("expected error for nil device")
	}
}

// TestMapToWritableRackRequestBasic verifies that MapToWritableRackRequest sets
// the rack name and defaults UHeight to 48 when no u_height metadata is present.
//
// Why it matters: racks export with a sane default height so Nautobot can place
// devices by rack unit even when the cani record omits the height.
// Inputs: a rack device named "Rack-A01" with location/status cached and no
// u_height metadata. Outputs: a request with Name "Rack-A01" and UHeight 48.
// Data choice: omitting u_height drives the 48U default branch.
func TestMapToWritableRackRequestBasic(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	statusID := uuid.New()

	cache.locationsMu.Lock()
	cache.locations["Floor-1"] = &CachedItem{ID: locID, Name: "Floor-1"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "Floor-1",
		DefaultStatus:   "Active",
	})

	dev := &devicetypes.CaniDeviceType{
		ID:   uuid.New(),
		Name: "Rack-A01",
	}

	req, err := mapper.MapToWritableRackRequest(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req == nil {
		t.Fatal("expected non-nil request")
	}
	if req.Name != "Rack-A01" {
		t.Errorf("name = %q", req.Name)
	}
	if req.UHeight == nil || *req.UHeight != 48 {
		t.Errorf("expected default u_height 48, got %v", req.UHeight)
	}
}

// TestMapToWritableRackRequestCustomHeight verifies that MapToWritableRackRequest
// honors ProviderMetadata["u_height"] instead of the 48U default.
//
// Why it matters: non-standard racks (e.g. 42U) must export with their real
// height so device positions remain valid in Nautobot.
// Inputs: a rack device with ProviderMetadata{"u_height":42} and cached
// location/status. Outputs: a request with UHeight 42.
// Data choice: 42 (not the 48 default) proves the metadata value overrides the
// default.
func TestMapToWritableRackRequestCustomHeight(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	statusID := uuid.New()

	cache.locationsMu.Lock()
	cache.locations["Room-C"] = &CachedItem{ID: locID, Name: "Room-C"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "Room-C",
		DefaultStatus:   "Active",
	})

	dev := &devicetypes.CaniDeviceType{
		ID:   uuid.New(),
		Name: "Rack-B02",
		ObjectMeta: devicetypes.ObjectMeta{
			ProviderMetadata: map[string]any{"u_height": 42},
		},
	}

	req, err := mapper.MapToWritableRackRequest(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.UHeight == nil || *req.UHeight != 42 {
		t.Errorf("expected u_height 42, got %v", req.UHeight)
	}
}

// TestMapToWritableRackRequestNoLocation verifies that MapToWritableRackRequest
// returns an error when no location can be resolved (no metadata, no default, no
// parent rack).
//
// Why it matters: Nautobot racks require a location; failing early stops an
// invalid rack create instead of letting the API reject it later.
// Inputs: a rack device with only Name/ID and a mapper with no DefaultLocation.
// Outputs: a non-nil error.
// Data choice: an empty MapperOpts removes every location source, forcing the
// "location is required" error.
func TestMapToWritableRackRequestNoLocation(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	dev := &devicetypes.CaniDeviceType{
		ID:   uuid.New(),
		Name: "Rack-NoLoc",
	}

	_, err := mapper.MapToWritableRackRequest(dev)
	if err == nil {
		t.Error("expected error when no location available")
	}
}

// ---------- MapToWritableRackRequest error path ----------

// TestMapToWritableRackRequestNoStatus verifies that MapToWritableRackRequest
// returns a non-nil error when the rack's status cannot be resolved.
//
// Why it matters: racks exported to Nautobot require a status; mapping must fail
// rather than create a rack with a missing status reference.
// Inputs: a CaniDeviceType (Name "Rack-Fail") with its location cached but a
// mapper with empty DefaultStatus. Outputs: a non-nil error.
// Data choice: caching the location while leaving status unsatisfiable isolates
// status resolution as the failure point for the rack request.
func TestMapToWritableRackRequestNoStatus(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	cache.locationsMu.Lock()
	cache.locations["Floor"] = &CachedItem{ID: locID, Name: "Floor"}
	cache.locationsMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "Floor",
		DefaultStatus:   "", // no status
	})
	dev := &devicetypes.CaniDeviceType{Name: "Rack-Fail"}
	_, err := mapper.MapToWritableRackRequest(dev)
	if err == nil {
		t.Fatal("expected error when status can't be resolved for rack")
	}
}

// ---------- MapToWritableRackRequest custom fields filtering ----------

// TestMapToWritableRackRequestCustomFieldsFiltered verifies
// MapToWritableRackRequest drops u_height and rack_position from CustomFields
// while preserving genuine custom keys.
//
// Why it matters: u_height/rack_position map to first-class Nautobot rack
// fields, so leaking them into CustomFields would duplicate data and risk a
// schema-invalid rack create.
// Inputs: a CaniDeviceType whose ProviderMetadata mixes the reserved keys with
// custom_key/site_code. Outputs: a WritableRackRequest with filtered
// CustomFields.
// Data choice: the metadata deliberately blends reserved and arbitrary keys to
// prove only the reserved ones are stripped.
func TestMapToWritableRackRequestCustomFieldsFiltered(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	statusID := uuid.New()

	cache.locationsMu.Lock()
	cache.locations["Room-CF"] = &CachedItem{ID: locID, Name: "Room-CF"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "Room-CF",
		DefaultStatus:   "Active",
	})

	dev := &devicetypes.CaniDeviceType{
		Name: "Rack-CF",
		ObjectMeta: devicetypes.ObjectMeta{
			ProviderMetadata: map[string]any{
				"u_height":      42,
				"rack_position": 5,
				"custom_key":    "custom_val",
				"site_code":     "S01",
			},
		},
	}

	req, err := mapper.MapToWritableRackRequest(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// u_height and rack_position should be filtered out of CustomFields
	if req.CustomFields == nil {
		t.Fatal("expected non-nil CustomFields")
	}
	cf := derefCustomFields(req.CustomFields)
	if _, ok := cf["u_height"]; ok {
		t.Error("u_height should be filtered from CustomFields")
	}
	if _, ok := cf["rack_position"]; ok {
		t.Error("rack_position should be filtered from CustomFields")
	}
	if cf["custom_key"] != "custom_val" {
		t.Errorf("custom_key = %v, want 'custom_val'", cf["custom_key"])
	}
	if cf["site_code"] != "S01" {
		t.Errorf("site_code = %v, want 'S01'", cf["site_code"])
	}
}

// TestMapToWritableRackRequestDefaultHeight verifies MapToWritableRackRequest
// defaults UHeight to 48 when the device carries no u_height metadata.
//
// Why it matters: a rack created in Nautobot needs a sane height, so cani must
// supply a default rather than emit a zero-height rack.
// Inputs: a CaniDeviceType with no ProviderMetadata. Outputs: a
// WritableRackRequest whose UHeight is 48.
// Data choice: omitting all metadata isolates the default-height branch with no
// other fields interfering.
func TestMapToWritableRackRequestDefaultHeight(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	statusID := uuid.New()

	cache.locationsMu.Lock()
	cache.locations["Room-DH"] = &CachedItem{ID: locID, Name: "Room-DH"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "Room-DH",
		DefaultStatus:   "Active",
	})

	// No ProviderMetadata → default to 48U
	dev := &devicetypes.CaniDeviceType{Name: "Rack-DH"}

	req, err := mapper.MapToWritableRackRequest(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.UHeight == nil || *req.UHeight != 48 {
		t.Errorf("expected default u_height 48, got %v", req.UHeight)
	}
}
