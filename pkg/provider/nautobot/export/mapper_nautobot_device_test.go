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

// ---------- MapToNautobotDevice (full path with populated cache) ----------

// TestMapToNautobotDeviceFull verifies that MapToNautobotDevice produces a
// fully-populated BulkWritableDeviceRequest (name, serial, asset_tag, comments)
// when device-type, location, status, and role caches are all pre-seeded.
//
// Why it matters: bulk device creation is the primary export path; every
// required FK must resolve from cache and every scalar field must round-trip so
// Nautobot records match the cani inventory.
// Inputs: a fully-populated CaniDeviceType plus a cache holding the slug,
// location, status, and role. Outputs: a non-nil request with all fields set.
// Data choice: realistic HPE DL360 values exercise the all-fields-present path
// with no auto-creation or error branches.
func TestMapToNautobotDeviceFull(t *testing.T) {
	cache := NewLookupCache(nil)

	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	// Pre-populate caches
	cache.deviceTypesMu.Lock()
	cache.deviceTypes["hpe-dl360-gen10"] = &CachedItem{ID: dtID, Name: "HPE DL360 Gen10"}
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

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
		DefaultLocation: "DC-01",
	})

	dev := &devicetypes.CaniDeviceType{
		ID:       uuid.New(),
		Name:     "server-01",
		Slug:     "hpe-dl360-gen10",
		Serial:   "SN123",
		AssetTag: "ASSET-001",
		Comments: "Test server",
	}

	req, err := mapper.MapToNautobotDevice(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req == nil {
		t.Fatal("expected non-nil request")
	}
	if *req.Name != "server-01" {
		t.Errorf("name = %q, want 'server-01'", *req.Name)
	}
	if req.Serial == nil || *req.Serial != "SN123" {
		t.Error("serial not set correctly")
	}
	if req.AssetTag == nil || *req.AssetTag != "ASSET-001" {
		t.Error("asset_tag not set correctly")
	}
	if req.Comments == nil || *req.Comments != "Test server" {
		t.Error("comments not set correctly")
	}
}

// ---------- MapToNautobotDevice with ProviderMetadata (custom fields) ----------

// TestMapToNautobotDeviceWithCustomFields verifies that MapToNautobotDevice
// flattens a device's nested ProviderMetadata ("nautobot" map) into the request's
// CustomFields, surfacing keys like "xname" and "import_date" at top level.
//
// Why it matters: cani provider metadata (e.g. HPC xnames, import dates) must be
// carried into Nautobot custom fields so exported devices stay traceable to source.
// Inputs: a fully classified CaniDeviceType whose ProviderMetadata nests a
// "nautobot" map. Outputs: a BulkWritableDeviceRequest whose *CustomFields holds
// the flattened keys.
// Data choice: nesting under "nautobot" exercises FlattenProviderMetadata's
// lifting of nested keys; all reference caches are seeded so mapping reaches the
// custom-field step without resolution errors.
func TestMapToNautobotDeviceWithCustomFields(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["test-dt"] = &CachedItem{ID: dtID, Name: "TestDT"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-CF"] = &CachedItem{ID: locID, Name: "DC-CF"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-CF",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		ID:   uuid.New(),
		Name: "custom-fields-device",
		Slug: "test-dt",
		ObjectMeta: devicetypes.ObjectMeta{
			ProviderMetadata: map[string]any{
				"nautobot": map[string]any{
					"xname":       "x3000c0s1b0",
					"import_date": "2026-01-01",
				},
			},
		},
	}

	req, err := mapper.MapToNautobotDevice(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.CustomFields == nil {
		t.Fatal("expected non-nil CustomFields")
	}
	cf := derefCustomFields(req.CustomFields)
	if cf["xname"] != "x3000c0s1b0" {
		t.Errorf("xname = %v", cf["xname"])
	}
	if cf["import_date"] != "2026-01-01" {
		t.Errorf("import_date = %v", cf["import_date"])
	}
}

// ---------- MapToNautobotDevice with comments ----------

// TestMapToNautobotDeviceComments verifies MapToNautobotDevice copies the cani
// Comments field into the Nautobot device request.
//
// Why it matters: operator notes on a device must survive the export so they
// remain visible in Nautobot.
// Inputs: a fully resolvable device (seeded type/location/status/role) with
// Comments="Important device". Outputs: a request whose Comments points to that
// string.
// Data choice: all FK caches are pre-seeded so the test isolates comment mapping
// from FK-resolution failures.
func TestMapToNautobotDeviceComments(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["cmt-dt"] = &CachedItem{ID: dtID, Name: "CMT-DT"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-C"] = &CachedItem{ID: locID, Name: "DC-C"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-C",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		Name:     "comment-dev",
		Slug:     "cmt-dt",
		Comments: "Important device",
	}

	req, err := mapper.MapToNautobotDevice(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Comments == nil || *req.Comments != "Important device" {
		t.Errorf("Comments = %v, want 'Important device'", req.Comments)
	}
}

// ---------- MapToNautobotDevice with serial and asset_tag ----------

// TestMapToNautobotDeviceSerialAndAssetTag verifies MapToNautobotDevice copies
// Serial and AssetTag into the Nautobot device request.
//
// Why it matters: serial and asset tag are key asset-tracking fields that must
// transfer accurately from cani into Nautobot.
// Inputs: a resolvable device with Serial="SN-99887766" and AssetTag="ASSET-XYZ".
// Outputs: a request whose Serial and AssetTag pointers hold those values.
// Data choice: distinctive sentinel strings make an accidental swap or drop
// obvious.
func TestMapToNautobotDeviceSerialAndAssetTag(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["sa-dt"] = &CachedItem{ID: dtID, Name: "SA-DT"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-SA"] = &CachedItem{ID: locID, Name: "DC-SA"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-SA",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		Name:     "serial-asset-dev",
		Slug:     "sa-dt",
		Serial:   "SN-99887766",
		AssetTag: "ASSET-XYZ",
	}

	req, err := mapper.MapToNautobotDevice(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Serial == nil || *req.Serial != "SN-99887766" {
		t.Errorf("Serial = %v, want 'SN-99887766'", req.Serial)
	}
	if req.AssetTag == nil || *req.AssetTag != "ASSET-XYZ" {
		t.Errorf("AssetTag = %v, want 'ASSET-XYZ'", req.AssetTag)
	}
}
