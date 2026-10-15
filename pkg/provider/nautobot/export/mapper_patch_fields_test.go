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

// ---------- MapToPatchRequest additional branches ----------

// TestMapToPatchRequestSerial verifies MapToPatchRequest copies a device's
// Serial into the patch request as a non-nil pointer.
//
// Why it matters: PATCH updates an existing Nautobot device; serial numbers
// must survive the mapping so inventory stays accurate on re-sync.
// Inputs: a fully cached mapper and a CaniDeviceType with Serial "SN123456".
// Outputs: a patch request whose *Serial equals "SN123456".
// Data choice: device-type/location/status/role are all pre-cached so mapping
// succeeds and the test isolates the Serial field.
func TestMapToPatchRequestSerial(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()
	cache.deviceTypesMu.Lock()
	cache.deviceTypes["patch-serial"] = &CachedItem{ID: dtID, Name: "Patch Serial"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-PS"] = &CachedItem{ID: locID, Name: "DC-PS"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-PS",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		Name:   "serial-device",
		Slug:   "patch-serial",
		Serial: "SN123456",
	}

	req, err := mapper.MapToPatchRequest(dev, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Serial == nil || *req.Serial != "SN123456" {
		t.Errorf("Serial = %v, want SN123456", req.Serial)
	}
}

// TestMapToPatchRequestAssetTag verifies MapToPatchRequest copies a device's
// AssetTag into the patch request as a non-nil pointer.
//
// Why it matters: PATCH updates an existing Nautobot device; asset tags must
// survive the mapping so inventory stays accurate on re-sync.
// Inputs: a fully cached mapper and a device with AssetTag "ASSET-001".
// Outputs: a patch request whose *AssetTag equals "ASSET-001".
// Data choice: all FK caches are pre-seeded so mapping succeeds and the test
// isolates the AssetTag field.
func TestMapToPatchRequestAssetTag(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()
	cache.deviceTypesMu.Lock()
	cache.deviceTypes["patch-at"] = &CachedItem{ID: dtID, Name: "Patch AT"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-AT"] = &CachedItem{ID: locID, Name: "DC-AT"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-AT",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		Name:     "asset-device",
		Slug:     "patch-at",
		AssetTag: "ASSET-001",
	}

	req, err := mapper.MapToPatchRequest(dev, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.AssetTag == nil || *req.AssetTag != "ASSET-001" {
		t.Errorf("AssetTag = %v, want ASSET-001", req.AssetTag)
	}
}

// TestMapToPatchRequestComments verifies MapToPatchRequest copies a device's
// Comments into the patch request as a non-nil pointer.
//
// Why it matters: free-form comments carry operator context that must be
// preserved when patching an existing Nautobot device.
// Inputs: a mapper with empty opts and a device carrying only Comments.
// Outputs: a patch request whose *Comments matches the input.
// Data choice: omitting a slug/defaults keeps the test on the comment-copy
// branch with no FK resolution needed.
func TestMapToPatchRequestComments(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	dev := &devicetypes.CaniDeviceType{
		Name:     "comments-device",
		Comments: "This is a test comment",
	}

	req, err := mapper.MapToPatchRequest(dev, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Comments == nil || *req.Comments != "This is a test comment" {
		t.Errorf("Comments = %v, want 'This is a test comment'", req.Comments)
	}
}

// TestMapToPatchRequestNoSlugSkipsDeviceType verifies that an empty Slug makes
// MapToPatchRequest leave DeviceType nil instead of resolving one.
//
// Why it matters: a PATCH should only touch the device type when the caller
// supplies one; otherwise the existing Nautobot device type must be left
// untouched.
// Inputs: a device with Name set but Slug "". Outputs: a patch request whose
// DeviceType pointer is nil.
// Data choice: an empty slug directly exercises the skip branch without needing
// a populated device-type cache.
func TestMapToPatchRequestNoSlugSkipsDeviceType(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	dev := &devicetypes.CaniDeviceType{
		Name: "no-slug-patch",
		Slug: "", // no slug means device type resolution is skipped
	}

	req, err := mapper.MapToPatchRequest(dev, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.DeviceType != nil {
		t.Error("expected nil DeviceType when slug is empty")
	}
}

// ---------- MapToPatchRequest with serial, asset tag, comments, custom fields ----------

// TestMapToPatchRequestSerialAssetComments verifies MapToPatchRequest carries
// Serial, AssetTag, Comments, and custom fields into the patch body.
//
// Why it matters: an in-place update of an existing Nautobot device must
// propagate edits to these tracking fields, not just structural ones.
// Inputs: a resolvable device with serial/asset/comments and a custom_key in
// ProviderMetadata. Outputs: a patch request mirroring all four values.
// Data choice: distinct "PATCH" sentinel strings make a dropped field obvious.
func TestMapToPatchRequestSerialAssetComments(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["test-dt"] = &CachedItem{ID: dtID, Name: "test-dt"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-01"] = &CachedItem{ID: locID, Name: "DC-01"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{DefaultLocation: "DC-01", DefaultStatus: "Active", DefaultRole: "Server"})
	dev := &devicetypes.CaniDeviceType{
		Name:     "patch-dev",
		Slug:     "test-dt",
		Serial:   "SN-PATCH-1",
		AssetTag: "TAG-PATCH-1",
		Comments: "updated via patch",
		ObjectMeta: devicetypes.ObjectMeta{
			ProviderMetadata: map[string]any{
				"custom_key": "custom_val",
			},
		},
	}

	req, err := mapper.MapToPatchRequest(dev, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Serial == nil || *req.Serial != "SN-PATCH-1" {
		t.Errorf("Serial = %v, want 'SN-PATCH-1'", req.Serial)
	}
	if req.AssetTag == nil || *req.AssetTag != "TAG-PATCH-1" {
		t.Errorf("AssetTag = %v, want 'TAG-PATCH-1'", req.AssetTag)
	}
	if req.Comments == nil || *req.Comments != "updated via patch" {
		t.Errorf("Comments = %v, want 'updated via patch'", req.Comments)
	}
	if req.CustomFields == nil {
		t.Fatal("CustomFields should not be nil")
	}
	cf := derefCustomFields(req.CustomFields)
	if cf["custom_key"] != "custom_val" {
		t.Errorf("CustomFields[custom_key] = %v, want 'custom_val'", cf["custom_key"])
	}
}

// TestMapToPatchRequestNilDevice verifies MapToPatchRequest returns an error when
// handed a nil device.
//
// Why it matters: a nil device must be rejected defensively so the exporter
// never dereferences it while building an update.
// Inputs: nil device plus a random target UUID. Outputs: a non-nil error.
// Data choice: nil is the degenerate input that guards the mapper's entry check.
func TestMapToPatchRequestNilDevice(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	_, err := mapper.MapToPatchRequest(nil, uuid.New())
	if err == nil {
		t.Fatal("expected error for nil device")
	}
}
