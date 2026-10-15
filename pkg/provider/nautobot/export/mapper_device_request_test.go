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

// ---------- MapToWritableDeviceRequest (full path with populated cache) ----------

// TestMapToWritableDeviceRequestFull verifies that MapToWritableDeviceRequest
// builds a complete request (name, serial, asset_tag, comments) when all
// reference caches are pre-seeded.
//
// Why it matters: this is the single-device create/update path; like the bulk
// path it must resolve every FK and copy scalar fields so updates to Nautobot
// faithfully reflect cani data.
// Inputs: a populated device and a cache with the DL380 slug, location,
// "Planned" status, and "Compute" role. Outputs: a non-nil request, fields set.
// Data choice: a different model/status/role than the bulk test guards against
// hard-coded values leaking between the two mapping methods.
func TestMapToWritableDeviceRequestFull(t *testing.T) {
	cache := NewLookupCache(nil)

	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["hpe-dl380-gen10"] = &CachedItem{ID: dtID, Name: "HPE DL380 Gen10"}
	cache.deviceTypesMu.Unlock()

	cache.locationsMu.Lock()
	cache.locations["DC-02"] = &CachedItem{ID: locID, Name: "DC-02"}
	cache.locationsMu.Unlock()

	cache.statusesMu.Lock()
	cache.statuses["Planned"] = &CachedItem{ID: statusID, Name: "Planned"}
	cache.statusesMu.Unlock()

	cache.rolesMu.Lock()
	cache.roles["Compute"] = &CachedItem{ID: roleID, Name: "Compute"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultStatus:   "Planned",
		DefaultRole:     "Compute",
		DefaultLocation: "DC-02",
	})

	dev := &devicetypes.CaniDeviceType{
		ID:       uuid.New(),
		Name:     "compute-01",
		Slug:     "hpe-dl380-gen10",
		Serial:   "SN456",
		AssetTag: "ASSET-002",
		Comments: "Compute node",
	}

	req, err := mapper.MapToWritableDeviceRequest(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req == nil {
		t.Fatal("expected non-nil request")
	}
	if *req.Name != "compute-01" {
		t.Errorf("name = %q, want 'compute-01'", *req.Name)
	}
	if req.Serial == nil || *req.Serial != "SN456" {
		t.Error("serial not set correctly")
	}
	if req.AssetTag == nil || *req.AssetTag != "ASSET-002" {
		t.Error("asset_tag not set correctly")
	}
	if req.Comments == nil || *req.Comments != "Compute node" {
		t.Error("comments not set correctly")
	}
}

// ---------- MapToWritableDeviceRequest with rack/position ----------

// TestMapToWritableDeviceRequestWithRack verifies that MapToWritableDeviceRequest
// succeeds and leaves Rack nil when the device has a rack position and face but
// no resolvable parent rack (rack lookup needs an API client, which is nil).
//
// Why it matters: devices may carry position/face hints before their rack is
// known; the mapper must still produce a valid request and simply omit the rack
// FK rather than fail or panic.
// Inputs: a device with RackPosition=10 and Face="rear" but no Parent.
// Outputs: a non-nil request with Rack==nil.
// Data choice: the nil API client deliberately suppresses rack resolution.
// NOTE: the name says "WithRack" yet it asserts the no-rack-resolution branch.
func TestMapToWritableDeviceRequestWithRack(t *testing.T) {
	cache := NewLookupCache(nil)

	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["hpe-dl360"] = &CachedItem{ID: dtID, Name: "HPE DL360"}
	cache.deviceTypesMu.Unlock()

	cache.locationsMu.Lock()
	cache.locations["Room-A"] = &CachedItem{ID: locID, Name: "Room-A"}
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
		DefaultLocation: "Room-A",
	})

	// Device without rack FK (rack resolution needs API which is nil).
	// This tests that the path through optional fields + no rack works.
	dev := &devicetypes.CaniDeviceType{
		ID:           uuid.New(),
		Name:         "racked-server",
		Slug:         "hpe-dl360",
		RackPosition: 10,
		Face:         "rear",
	}

	req, err := mapper.MapToWritableDeviceRequest(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req == nil {
		t.Fatal("expected non-nil request")
	}
	if *req.Name != "racked-server" {
		t.Errorf("name = %q, want 'racked-server'", *req.Name)
	}
	// Without rack resolution, Rack should be nil
	if req.Rack != nil {
		t.Error("expected nil Rack (no API client)")
	}
}

// ---------- MapToWritableDeviceRequest with custom fields ----------

// TestMapToWritableDeviceRequestCustomFields verifies that
// MapToWritableDeviceRequest populates the request's CustomFields from the
// device's flattened ProviderMetadata (e.g. "custom_key").
//
// Why it matters: the single-create export path (not just bulk) must also carry
// cani provider metadata into Nautobot custom fields for parity and traceability.
// Inputs: a classified CaniDeviceType with ProviderMetadata {"custom_key":
// "custom_val"} and a fully seeded mapper. Outputs: a WritableDeviceRequest whose
// *CustomFields contains custom_key.
// Data choice: a flat top-level metadata key (no "nautobot" nesting) confirms the
// mapper passes through simple custom fields on the single-create request.
func TestMapToWritableDeviceRequestCustomFields(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["cf-slug"] = &CachedItem{ID: dtID, Name: "CF-Slug"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-WR"] = &CachedItem{ID: locID, Name: "DC-WR"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-WR",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		ID:   uuid.New(),
		Name: "wr-cf-device",
		Slug: "cf-slug",
		ObjectMeta: devicetypes.ObjectMeta{
			ProviderMetadata: map[string]any{
				"custom_key": "custom_val",
			},
		},
	}

	req, err := mapper.MapToWritableDeviceRequest(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.CustomFields == nil {
		t.Fatal("expected non-nil CustomFields")
	}
	cf := derefCustomFields(req.CustomFields)
	if cf["custom_key"] != "custom_val" {
		t.Errorf("custom_key = %v", cf["custom_key"])
	}
}

// ---------- MapToWritableDeviceRequest with serial/asset/comments ----------

// TestMapToWritableDeviceRequestSerial verifies MapToWritableDeviceRequest
// copies Serial, AssetTag, and Comments into the create request.
//
// Why it matters: creating a new Nautobot device must carry over all
// identifying metadata so the exported record matches cani inventory.
// Inputs: a fully cached mapper and a device with Serial/AssetTag/Comments set.
// Outputs: a request whose *Serial, *AssetTag, and *Comments match.
// Data choice: all FK caches are pre-seeded so the create path succeeds and the
// test focuses on the optional string fields.
func TestMapToWritableDeviceRequestSerial(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["wr-serial"] = &CachedItem{ID: dtID, Name: "WR-Serial"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-WS2"] = &CachedItem{ID: locID, Name: "DC-WS2"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-WS2",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		Name:     "wr-serial-dev",
		Slug:     "wr-serial",
		Serial:   "SER-XYZ",
		AssetTag: "ASSET-99",
		Comments: "Test device comments",
	}

	req, err := mapper.MapToWritableDeviceRequest(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Serial == nil || *req.Serial != "SER-XYZ" {
		t.Errorf("Serial = %v, want SER-XYZ", req.Serial)
	}
	if req.AssetTag == nil || *req.AssetTag != "ASSET-99" {
		t.Errorf("AssetTag = %v, want ASSET-99", req.AssetTag)
	}
	if req.Comments == nil || *req.Comments != "Test device comments" {
		t.Errorf("Comments = %v, want 'Test device comments'", req.Comments)
	}
}

// ---------- MapToWritableDeviceRequest with rack from inventory ----------

// TestMapToWritableDeviceRequestNoParentNoRack verifies
// MapToWritableDeviceRequest leaves Rack and Position unset when the device has
// no parent (Parent == uuid.Nil).
//
// Why it matters: only rack-mounted devices should carry rack/position in
// Nautobot; an unparented device must export without a phantom rack assignment.
// Inputs: a device with Parent=uuid.Nil but RackPosition/Face set, plus a
// seeded cache and inventory. Outputs: a request with nil Rack and Position.
// Data choice: RackPosition/Face are populated to prove they are ignored when
// no parent resolves to a rack, avoiding the nil-client GetRackByName path.
func TestMapToWritableDeviceRequestNoParentNoRack(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["server-dt"] = &CachedItem{ID: dtID, Name: "Server DT"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-Rack"] = &CachedItem{ID: locID, Name: "DC-Rack"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-Rack",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	// Set up inventory with a rack
	rackID := uuid.New()
	devID := uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackID: {ID: rackID, Name: "Rack-Test-01"},
		},
	}
	mapper.SetInventory(inv)

	// Mock GetRackByName by pre-populating the devices cache won't work for racks
	// The rack lookup uses GetRackByName which hits the API — can't test that path
	// without a nil client panic. But we CAN test the path where GetRackID returns
	// uuid.Nil (no Parent set), which skips the rack block entirely.

	dev := &devicetypes.CaniDeviceType{
		Name:         "server-in-rack",
		Slug:         "server-dt",
		RackPosition: 10,
		Face:         "rear",
		Parent:       uuid.Nil, // no parent → no rack resolution
	}
	inv.Devices[devID] = dev

	req, err := mapper.MapToWritableDeviceRequest(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No rack should be set since Parent is Nil
	if req.Rack != nil {
		t.Error("expected nil Rack when device has no parent")
	}
	// Position should not be set either
	if req.Position != nil {
		t.Error("expected nil Position when no rack assigned")
	}
}

// ---------- MapToWritableDeviceRequest with rack in Racks collection ----------

// TestMapToWritableDeviceRequestWithRacksCollection verifies
// MapToWritableDeviceRequest resolves a rack from the inventory's Racks
// collection through the real Nautobot rack lookup path.
//
// Why it matters: new device creates must include rack UUID, position, and face
// so Nautobot stores the device in the same physical slot as the cani inventory.
// Inputs: a classified device whose Rack field points at inventory rack
// "rack-1" and a fake Nautobot server returning that rack UUID. Outputs: a
// writable request with Rack, Position=22, and Face=rear set.
// Data choice: Rack (not Parent) is the field the create mapper resolves via
// GetRackID, so this fixture exercises the create-specific rack branch.
func TestMapToWritableDeviceRequestWithRacksCollection(t *testing.T) {
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
		Name:         "rack-srv",
		Slug:         "hpe-dl380",
		Rack:         rackID,
		RackPosition: 22,
		Face:         "rear",
		ObjectMeta:   devicetypes.ObjectMeta{Status: "Active", Role: "Compute"},
	}

	req, err := mapper.MapToWritableDeviceRequest(dev)
	if err != nil {
		t.Fatalf("MapToWritableDeviceRequest() error = %v", err)
	}
	if req.Rack == nil || req.Rack.Id == nil {
		t.Fatal("expected rack reference to be set")
	}
	if got := extractRackID(t, req.Rack.Id); got != nautobotRackID {
		t.Errorf("rack id = %s, want %s", got, nautobotRackID)
	}
	if req.Position == nil || *req.Position != 22 {
		t.Errorf("position = %v, want 22", req.Position)
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
