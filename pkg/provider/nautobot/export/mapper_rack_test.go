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
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// rackLookupServer returns a handler that answers the rack-name lookup
// (GET /dcim/racks/) with a single rack carrying nautobotRackID and replies to
// every other request with an empty result list. seedDeviceRefs pre-caches the
// device type, location, role and status so the mapper resolves those offline,
// leaving the rack lookup as the only HTTP call this handler must satisfy.
func rackLookupServer(nautobotRackID uuid.UUID) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if strings.Contains(r.URL.Path, "dcim/racks") {
			_, _ = fmt.Fprintf(w, `{"count":1,"results":[%s]}`,
				createdItemJSON(nautobotRackID, "rack-1"))
			return
		}
		_, _ = w.Write([]byte(emptyListJSON))
	}
}

// extractRackID unwraps the rack reference union on a write request into a UUID.
func extractRackID(t *testing.T, id *nautobotapi.BulkWritableCableRequestStatusId) uuid.UUID {
	t.Helper()
	got, err := id.AsBulkWritableCableRequestStatusId0()
	if err != nil {
		t.Fatalf("decode rack id union: %v", err)
	}
	return uuid.UUID(got)
}

// TestMapToWritableDeviceRequest_FullRackPlacement verifies the single-create
// mapper populates the rack reference, U-position, face, and the optional
// serial/asset-tag/comments/custom-field attributes when a device is mounted in
// a rack present in the inventory's Racks collection.
//
// Why it matters: a device exported to Nautobot is anchored to its rack slot and
// carries operational metadata; if the mapper omitted the rack/position/face or
// dropped the optional fields, the remote device would be created floating and
// stripped of its provenance, breaking placement and audit.
// Inputs: a fully populated CaniDeviceType (rack FK, RackPosition=12, Face=rear,
// serial, asset tag, comments, provider metadata) whose references are seeded in
// the cache, with the parent rack resolved by a fake server. Outputs: a
// WritableDeviceRequest whose Rack decodes to the Nautobot rack UUID, Position
// is 12, Face is rear, and Serial/AssetTag/Comments/CustomFields are set.
// Data choice: Face "rear" is the non-default mounting face and a non-zero
// position plus every optional field exercises each conditional copy in the
// rack branch that the nil/minimal-device tests skip.
func TestMapToWritableDeviceRequest_FullRackPlacement(t *testing.T) {
	caniRackID := uuid.New()
	nautobotRackID := uuid.New()
	e, cleanup := newExporterWithServer(t, rackLookupServer(nautobotRackID))
	defer cleanup()
	seedDeviceRefs(t, e)

	device := &devicetypes.CaniDeviceType{
		Name:         "compute-001",
		Slug:         "hpe-dl380",
		Serial:       "SGH123",
		AssetTag:     "ASSET-9",
		Comments:     "rack unit 12",
		Rack:         caniRackID,
		RackPosition: 12,
		Face:         "rear",
		ObjectMeta: devicetypes.ObjectMeta{
			Status:           "Active",
			Role:             "Compute",
			ProviderMetadata: map[string]any{"nautobot": map[string]any{"nid": "42"}},
		},
	}

	inv := devicetypes.NewInventory()
	inv.Racks[caniRackID] = &devicetypes.CaniRackType{ID: caniRackID, Name: "rack-1"}
	mapper := newCrudMapper(e)
	mapper.SetInventory(inv)

	req, err := mapper.MapToWritableDeviceRequest(device)
	if err != nil {
		t.Fatalf("MapToWritableDeviceRequest() error = %v", err)
	}

	if req.Rack == nil || req.Rack.Id == nil {
		t.Fatal("expected rack reference to be set")
	}
	if got := extractRackID(t, req.Rack.Id); got != nautobotRackID {
		t.Errorf("rack id = %s, want %s", got, nautobotRackID)
	}
	if req.Position == nil || *req.Position != 12 {
		t.Errorf("position = %v, want 12", req.Position)
	}
	if req.Face == nil {
		t.Fatal("expected face to be set")
	}
	if face, _ := req.Face.AsFaceEnum(); face != nautobotapi.FaceEnumRear {
		t.Errorf("face = %v, want rear", face)
	}
	if req.Serial == nil || *req.Serial != "SGH123" {
		t.Errorf("serial = %v, want SGH123", req.Serial)
	}
	if req.AssetTag == nil || *req.AssetTag != "ASSET-9" {
		t.Errorf("asset tag = %v, want ASSET-9", req.AssetTag)
	}
	if req.Comments == nil || *req.Comments != "rack unit 12" {
		t.Errorf("comments = %v, want 'rack unit 12'", req.Comments)
	}
	if req.CustomFields == nil || (*req.CustomFields)["nid"] != "42" {
		t.Errorf("custom fields = %v, want nid=42", req.CustomFields)
	}
}

// TestMapToWritableDeviceRequest_LegacyRackDevice verifies the single-create
// mapper resolves a rack that is modelled as a rack-type entry in the Devices
// collection (the legacy representation) rather than the Racks collection.
//
// Why it matters: older inventories store racks as devices with Type=rack, and
// the exporter must still place child devices into those racks; without the
// fallback branch, legacy-sourced devices would lose their rack placement.
// Inputs: a device whose Rack FK points at a Type=rack entry in inv.Devices,
// with RackPosition=3 and the rack resolvable via the fake server. Outputs: a
// WritableDeviceRequest whose Rack decodes to the Nautobot rack UUID and
// Position is 3. Data choice: putting the rack only in inv.Devices (not
// inv.Racks) forces the Racks-collection lookup to miss and exercises the
// else-if legacy fallback exclusively.
func TestMapToWritableDeviceRequest_LegacyRackDevice(t *testing.T) {
	legacyRackID := uuid.New()
	nautobotRackID := uuid.New()
	e, cleanup := newExporterWithServer(t, rackLookupServer(nautobotRackID))
	defer cleanup()
	seedDeviceRefs(t, e)

	device := &devicetypes.CaniDeviceType{
		Name:         "compute-002",
		Slug:         "hpe-dl380",
		Rack:         legacyRackID,
		RackPosition: 3,
		ObjectMeta:   devicetypes.ObjectMeta{Status: "Active", Role: "Compute"},
	}

	inv := devicetypes.NewInventory()
	inv.Devices[legacyRackID] = &devicetypes.CaniDeviceType{
		ID: legacyRackID, Name: "rack-1", Type: devicetypes.Rack,
	}
	mapper := newCrudMapper(e)
	mapper.SetInventory(inv)

	req, err := mapper.MapToWritableDeviceRequest(device)
	if err != nil {
		t.Fatalf("MapToWritableDeviceRequest() error = %v", err)
	}
	if req.Rack == nil || req.Rack.Id == nil {
		t.Fatal("expected rack reference from legacy device-as-rack fallback")
	}
	if got := extractRackID(t, req.Rack.Id); got != nautobotRackID {
		t.Errorf("rack id = %s, want %s", got, nautobotRackID)
	}
	if req.Position == nil || *req.Position != 3 {
		t.Errorf("position = %v, want 3", req.Position)
	}
}

// TestMapToPatchRequest_FullRackPlacement verifies the update (PATCH) mapper
// populates the rack reference, position, face, and optional attributes when the
// device's parent is a rack in the inventory's Racks collection.
//
// Why it matters: merge exports update existing Nautobot devices in place, so a
// re-placed or re-tagged device must carry its new rack slot and metadata into
// the PATCH; dropping them would silently revert remote state on every sync.
// Inputs: a device whose Parent points at a rack in inv.Racks, with
// RackPosition=7, Face=rear, serial, asset tag, comments and provider metadata,
// references seeded in the cache, rack resolved via the fake server, and an
// existing device UUID. Outputs: a PatchedWritableDeviceRequest whose Rack
// decodes to the Nautobot rack UUID, Position is 7, Face is rear, and the
// optional fields are set. Data choice: driving the patch path via device.Parent
// (not the Rack FK) matches how MapToPatchRequest resolves the rack and covers
// its dedicated rack branch independent of the create path.
func TestMapToPatchRequest_FullRackPlacement(t *testing.T) {
	caniRackID := uuid.New()
	nautobotRackID := uuid.New()
	e, cleanup := newExporterWithServer(t, rackLookupServer(nautobotRackID))
	defer cleanup()
	seedDeviceRefs(t, e)

	device := &devicetypes.CaniDeviceType{
		Name:         "compute-003",
		Slug:         "hpe-dl380",
		Serial:       "SGH777",
		AssetTag:     "ASSET-7",
		Comments:     "updated placement",
		Parent:       caniRackID,
		RackPosition: 7,
		Face:         "rear",
		ObjectMeta: devicetypes.ObjectMeta{
			Status:           "Active",
			Role:             "Compute",
			ProviderMetadata: map[string]any{"nautobot": map[string]any{"alias": "c3"}},
		},
	}

	inv := devicetypes.NewInventory()
	inv.Racks[caniRackID] = &devicetypes.CaniRackType{ID: caniRackID, Name: "rack-1"}
	mapper := newCrudMapper(e)
	mapper.SetInventory(inv)

	req, err := mapper.MapToPatchRequest(device, uuid.New())
	if err != nil {
		t.Fatalf("MapToPatchRequest() error = %v", err)
	}

	if req.Rack == nil || req.Rack.Id == nil {
		t.Fatal("expected rack reference to be set on patch")
	}
	if got := extractRackID(t, req.Rack.Id); got != nautobotRackID {
		t.Errorf("rack id = %s, want %s", got, nautobotRackID)
	}
	if req.Position == nil || *req.Position != 7 {
		t.Errorf("position = %v, want 7", req.Position)
	}
	if req.Face == nil {
		t.Fatal("expected face to be set on patch")
	}
	if face, _ := req.Face.AsFaceEnum(); face != nautobotapi.FaceEnumRear {
		t.Errorf("face = %v, want rear", face)
	}
	if req.Serial == nil || *req.Serial != "SGH777" {
		t.Errorf("serial = %v, want SGH777", req.Serial)
	}
	if req.AssetTag == nil || *req.AssetTag != "ASSET-7" {
		t.Errorf("asset tag = %v, want ASSET-7", req.AssetTag)
	}
	if req.Comments == nil || *req.Comments != "updated placement" {
		t.Errorf("comments = %v, want 'updated placement'", req.Comments)
	}
	if req.CustomFields == nil || (*req.CustomFields)["alias"] != "c3" {
		t.Errorf("custom fields = %v, want alias=c3", req.CustomFields)
	}
}

// TestMapToPatchRequest_LegacyRackDevice verifies the update (PATCH) mapper
// resolves a rack modelled as a rack-type entry in the Devices collection (the
// legacy representation) when reached through the device's parent reference.
//
// Why it matters: merge syncs of legacy-sourced inventories must keep updating
// the rack placement of child devices whose rack is stored as a Type=rack
// device; without the patch-path fallback, every merge would strip those
// devices' rack and position.
// Inputs: a device whose Parent points at a Type=rack entry in inv.Devices (and
// deliberately absent from inv.Racks), with RackPosition=5, references seeded in
// the cache, and the rack resolvable via the fake server. Outputs: a
// PatchedWritableDeviceRequest whose Rack decodes to the Nautobot rack UUID and
// Position is 5. Data choice: placing the rack only in inv.Devices forces the
// primary Racks lookup to miss and exercises the patch-path legacy else-if
// fallback exclusively.
func TestMapToPatchRequest_LegacyRackDevice(t *testing.T) {
	legacyRackID := uuid.New()
	nautobotRackID := uuid.New()
	e, cleanup := newExporterWithServer(t, rackLookupServer(nautobotRackID))
	defer cleanup()
	seedDeviceRefs(t, e)

	device := &devicetypes.CaniDeviceType{
		Name:         "compute-004",
		Slug:         "hpe-dl380",
		Parent:       legacyRackID,
		RackPosition: 5,
		ObjectMeta:   devicetypes.ObjectMeta{Status: "Active", Role: "Compute"},
	}

	inv := devicetypes.NewInventory()
	inv.Devices[legacyRackID] = &devicetypes.CaniDeviceType{
		ID: legacyRackID, Name: "rack-1", Type: devicetypes.Rack,
	}
	mapper := newCrudMapper(e)
	mapper.SetInventory(inv)

	req, err := mapper.MapToPatchRequest(device, uuid.New())
	if err != nil {
		t.Fatalf("MapToPatchRequest() error = %v", err)
	}
	if req.Rack == nil || req.Rack.Id == nil {
		t.Fatal("expected rack reference from legacy device-as-rack fallback")
	}
	if got := extractRackID(t, req.Rack.Id); got != nautobotRackID {
		t.Errorf("rack id = %s, want %s", got, nautobotRackID)
	}
	if req.Position == nil || *req.Position != 5 {
		t.Errorf("position = %v, want 5", req.Position)
	}
}
