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
	"github.com/google/uuid"
)

// ---------- MapToPatchRequest ----------

// TestMapToPatchRequestNil verifies that MapToPatchRequest returns an error when
// given a nil device.
//
// Why it matters: PATCH builds an update body for an existing Nautobot device;
// guarding nil input prevents a panic mid-export and surfaces the programming
// error instead.
// Inputs: a nil device and an arbitrary existing UUID. Outputs: a non-nil error.
// Data choice: nil is the boundary input that must be rejected before any field
// access.
func TestMapToPatchRequestNil(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	_, err := mapper.MapToPatchRequest(nil, uuid.New())
	if err == nil {
		t.Error("expected error for nil device")
	}
}

// TestMapToPatchRequestBasic verifies that MapToPatchRequest populates name,
// serial, asset_tag, comments and resolves device_type, location, status, and
// role when all caches are seeded.
//
// Why it matters: updating an existing device must carry both scalar edits and
// resolved FKs so an in-place Nautobot PATCH fully reconciles the record.
// Inputs: a populated device plus caches for slug, location, status, role.
// Outputs: a request with all scalars and all four references non-nil.
// Data choice: every reference pre-seeded drives the all-fields-set path with no
// skipped FK.
func TestMapToPatchRequestBasic(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["test-slug"] = &CachedItem{ID: dtID, Name: "TestSlug"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["Site-X"] = &CachedItem{ID: locID, Name: "Site-X"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "Site-X",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		ID:       uuid.New(),
		Name:     "patched-device",
		Slug:     "test-slug",
		Serial:   "SN-PATCH",
		AssetTag: "AT-PATCH",
		Comments: "patched",
	}

	req, err := mapper.MapToPatchRequest(dev, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req == nil {
		t.Fatal("expected non-nil request")
	}
	if *req.Name != "patched-device" {
		t.Errorf("name = %q", *req.Name)
	}
	if req.Serial == nil || *req.Serial != "SN-PATCH" {
		t.Error("serial not set")
	}
	if req.AssetTag == nil || *req.AssetTag != "AT-PATCH" {
		t.Error("asset_tag not set")
	}
	if req.Comments == nil || *req.Comments != "patched" {
		t.Error("comments not set")
	}
	if req.DeviceType == nil {
		t.Error("device_type not resolved")
	}
	if req.Location == nil {
		t.Error("location not resolved")
	}
	if req.Status == nil {
		t.Error("status not resolved")
	}
	if req.Role == nil {
		t.Error("role not resolved")
	}
}

// TestMapToPatchRequestNoSlug verifies that MapToPatchRequest leaves DeviceType
// nil when the device has no slug, while still building the rest of the request.
//
// Why it matters: a PATCH should not overwrite a device's type with an empty
// value; omitting the device_type FK when the slug is unknown avoids corrupting
// the Nautobot record.
// Inputs: a device with only a Name and caches for location/status/role.
// Outputs: a request with DeviceType==nil.
// Data choice: an empty slug specifically targets the skip-device-type branch.
func TestMapToPatchRequestNoSlug(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.locationsMu.Lock()
	cache.locations["Site-Y"] = &CachedItem{ID: locID, Name: "Site-Y"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "Site-Y",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		ID:   uuid.New(),
		Name: "no-slug-device",
	}

	req, err := mapper.MapToPatchRequest(dev, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// DeviceType should be nil (slug is empty, resolveDeviceType would error)
	if req.DeviceType != nil {
		t.Error("expected nil DeviceType for empty slug")
	}
}

// TestMapToPatchRequestWithCustomFields verifies that MapToPatchRequest copies
// flattened ProviderMetadata into the request's CustomFields, e.g. a nested
// nautobot.xname becomes the "xname" custom field.
//
// Why it matters: cani stores Nautobot-specific attributes (like xname) in
// provider metadata; flattening them into custom fields is how that data reaches
// Nautobot on update.
// Inputs: a device whose ProviderMetadata nests {"nautobot":{"xname":...}}.
// Outputs: CustomFields["xname"]=="x1000c0s0b0".
// Data choice: a nested map proves FlattenProviderMetadata unwraps the
// "nautobot" namespace into a top-level custom field.
func TestMapToPatchRequestWithCustomFields(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.locationsMu.Lock()
	cache.locations["DC"] = &CachedItem{ID: locID, Name: "DC"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Generic"] = &CachedItem{ID: roleID, Name: "Generic"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC",
		DefaultStatus:   "Active",
		DefaultRole:     "Generic",
	})

	dev := &devicetypes.CaniDeviceType{
		ID:   uuid.New(),
		Name: "cf-device",
		ObjectMeta: devicetypes.ObjectMeta{
			ProviderMetadata: map[string]any{
				"nautobot": map[string]any{
					"xname": "x1000c0s0b0",
				},
			},
		},
	}

	req, err := mapper.MapToPatchRequest(dev, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.CustomFields == nil {
		t.Fatal("expected non-nil CustomFields")
	}
	cf := derefCustomFields(req.CustomFields)
	if cf["xname"] != "x1000c0s0b0" {
		t.Errorf("custom field xname = %v", cf["xname"])
	}
}

// ---------- MapToPatchRequest with rack position via inventory ----------

// TestMapToPatchRequestWithRackPositionNoParent verifies MapToPatchRequest skips
// the rack block (Rack stays nil) when the device has no parent.
//
// Why it matters: updating an existing Nautobot device must not attach a rack
// when the local device is unparented, even if a RackPosition is present.
// Inputs: a device with Parent=uuid.Nil, RackPosition=15, Face="front", and a
// seeded cache/inventory. Outputs: a patch request with nil Rack.
// Data choice: RackPosition/Face are set to confirm they are ignored without a
// parent, sidestepping the nil-client rack API call.
func TestMapToPatchRequestWithRackPositionNoParent(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["patch-rp"] = &CachedItem{ID: dtID, Name: "Patch RP"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-RP"] = &CachedItem{ID: locID, Name: "DC-RP"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-RP",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})
	mapper.SetInventory(&devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
		Racks:   map[uuid.UUID]*devicetypes.CaniRackType{},
	})

	dev := &devicetypes.CaniDeviceType{
		Name:         "patch-rp-dev",
		Slug:         "patch-rp",
		RackPosition: 15,
		Face:         "front",
		Parent:       uuid.Nil, // no parent → rack/position block is skipped
	}

	req, err := mapper.MapToPatchRequest(dev, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Without a parent, rack block should be skipped
	if req.Rack != nil {
		t.Error("expected nil Rack when device has no parent")
	}
}

// ---------- MapToPatchRequest rack via Racks collection ----------

// TestMapToPatchRequestRackViaRacksCollection verifies MapToPatchRequest
// resolves a parent rack from the inventory's Racks collection through the real
// Nautobot rack lookup path.
//
// Why it matters: merge exports PATCH existing Nautobot devices; if rack
// placement is not resolved into a rack UUID, moved devices lose their slot.
// Inputs: a device whose Parent is an inventory rack named "rack-1" and a fake
// Nautobot server returning that rack UUID. Outputs: a patch request with Rack,
// Position=10, and Face=rear set.
// Data choice: Parent (not Rack) is the field the patch mapper uses, so this
// fixture exercises the update-specific rack branch directly.
func TestMapToPatchRequestRackViaRacksCollection(t *testing.T) {
	nautobotRackID := uuid.New()
	rackID := uuid.New()
	e, cleanup := newExporterWithServer(t, rackLookupServer(nautobotRackID))
	defer cleanup()
	seedDeviceRefs(t, e)

	mapper := newCrudMapper(e)
	inv := devicetypes.NewInventory()
	inv.Racks[rackID] = &devicetypes.CaniRackType{ID: rackID, Name: "rack-1"}
	mapper.SetInventory(inv)

	dev := &devicetypes.CaniDeviceType{
		Name:         "rack-dev",
		Parent:       rackID,
		RackPosition: 10,
		Face:         "rear",
		ObjectMeta:   devicetypes.ObjectMeta{Status: "Active", Role: "Compute"},
	}

	req, err := mapper.MapToPatchRequest(dev, uuid.New())
	if err != nil {
		t.Fatalf("MapToPatchRequest() error = %v", err)
	}
	if req.Rack == nil || req.Rack.Id == nil {
		t.Fatal("expected rack reference to be set")
	}
	if got := extractRackID(t, req.Rack.Id); got != nautobotRackID {
		t.Errorf("rack id = %s, want %s", got, nautobotRackID)
	}
	if req.Position == nil || *req.Position != 10 {
		t.Errorf("position = %v, want 10", req.Position)
	}
	if req.Face == nil {
		t.Fatal("expected face to be set")
	}
	face, err := req.Face.AsFaceEnum()
	if err != nil {
		t.Fatalf("decode face: %v", err)
	}
	if face != nautobotapi.FaceEnumRear {
		t.Errorf("face = %v, want %v", face, nautobotapi.FaceEnumRear)
	}
}
