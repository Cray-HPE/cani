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
package transform

import (
	"testing"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// TestMapRacks verifies MapRacks converts racks, resolves each rack's status
// name and location to a CANI UUID, passes custom fields through, and skips
// racks with a nil Id.
//
// Why it matters: racks are imported after locations and before devices; a rack's
// location must be rewritten from the Nautobot UUID to the CANI UUID via the
// location map, and an unknown location must resolve to uuid.Nil so the rack does
// not point at a location that was not imported.
// Inputs: a location lookup map plus racks that are nil-Id, carry a known
// location, carry an unknown location, and carry custom fields. Outputs: the CANI
// rack map and Nautobot->CANI UUID map with name, status, serial, asset tag,
// and location.
// Data choice: a populated location map proves the known location resolves to its
// CANI UUID, while a deliberately absent location UUID proves the miss leaves
// Location at uuid.Nil.
func TestMapRacks(t *testing.T) {
	locNBID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	locCaniID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	locationMap := map[uuid.UUID]uuid.UUID{locNBID: locCaniID}

	t.Run("empty input returns empty maps", func(t *testing.T) {
		racks, nbMap := MapRacks(nil, locationMap, nil, nil)
		if len(racks) != 0 {
			t.Errorf("expected 0 racks, got %d", len(racks))
		}
		if len(nbMap) != 0 {
			t.Errorf("expected 0 mappings, got %d", len(nbMap))
		}
	})

	t.Run("rack with nil ID is skipped", func(t *testing.T) {
		raw := []nautobotapi.Rack{
			{Name: "orphan", Id: nil, Status: makeStatusRefFromUUID(uuid.New()), Location: makeStatusRefFromUUID(uuid.New())},
		}
		racks, _ := MapRacks(raw, locationMap, nil, nil)
		if len(racks) != 0 {
			t.Errorf("expected 0 racks, got %d", len(racks))
		}
	})

	t.Run("single rack with location resolved", func(t *testing.T) {
		nbID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
		oaID := openapi_types.UUID(nbID)
		serial := "SN-RACK1"
		tag := "ASSET-001"
		comment := "floor 2"
		statusID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
		statusNameMap := map[uuid.UUID]string{statusID: "Active"}

		raw := []nautobotapi.Rack{
			{
				Id:       &oaID,
				Name:     "Rack-A",
				Serial:   &serial,
				AssetTag: &tag,
				Comments: &comment,
				Status:   makeStatusRefFromUUID(statusID),
				Location: makeStatusRefFromUUID(locNBID),
			},
		}

		racks, nbMap := MapRacks(raw, locationMap, statusNameMap, nil)
		if len(racks) != 1 {
			t.Fatalf("expected 1 rack, got %d", len(racks))
		}

		caniID := nbMap[nbID]
		rack := racks[caniID]
		if rack == nil {
			t.Fatal("rack not found by CANI ID")
		}
		if rack.Name != "Rack-A" {
			t.Errorf("Name = %q, want %q", rack.Name, "Rack-A")
		}
		if rack.Serial != "SN-RACK1" {
			t.Errorf("Serial = %q, want %q", rack.Serial, "SN-RACK1")
		}
		if rack.AssetTag != "ASSET-001" {
			t.Errorf("AssetTag = %q, want %q", rack.AssetTag, "ASSET-001")
		}
		if rack.Comments != "floor 2" {
			t.Errorf("Comments = %q, want %q", rack.Comments, "floor 2")
		}
		if rack.Status != "Active" {
			t.Errorf("Status = %q, want %q", rack.Status, "Active")
		}
		if rack.Location != locCaniID {
			t.Errorf("Location = %s, want %s", rack.Location, locCaniID)
		}
		if rack.ObjectMeta.ExternalIDs["nautobot"] != nbID {
			t.Errorf("ExternalIDs[nautobot] = %s, want %s", rack.ObjectMeta.ExternalIDs["nautobot"], nbID)
		}
	})

	t.Run("unknown location is not resolved", func(t *testing.T) {
		unknownLocID := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
		nbID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
		oaID := openapi_types.UUID(nbID)

		raw := []nautobotapi.Rack{
			{
				Id:       &oaID,
				Name:     "Floating-Rack",
				Status:   makeStatusRefFromUUID(uuid.New()),
				Location: makeStatusRefFromUUID(unknownLocID),
			},
		}

		racks, nbMap := MapRacks(raw, locationMap, nil, nil)
		caniID := nbMap[nbID]
		rack := racks[caniID]

		if rack.Location != uuid.Nil {
			t.Errorf("Location = %s, want Nil for unknown location", rack.Location)
		}
	})

	t.Run("custom fields are passed through", func(t *testing.T) {
		nbID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
		oaID := openapi_types.UUID(nbID)
		cf := map[string]interface{}{"zone": "east"}

		raw := []nautobotapi.Rack{
			{
				Id:           &oaID,
				Name:         "CF-Rack",
				CustomFields: &cf,
				Status:       makeStatusRefFromUUID(uuid.New()),
				Location:     makeStatusRefFromUUID(uuid.New()),
			},
		}

		racks, nbMap := MapRacks(raw, locationMap, nil, nil)
		caniID := nbMap[nbID]
		rack := racks[caniID]

		if rack.CustomFields == nil {
			t.Fatal("expected CustomFields to be set")
		}
		if rack.CustomFields["zone"] != "east" {
			t.Errorf("CustomFields[zone] = %v, want %q", rack.CustomFields["zone"], "east")
		}
	})
}

// TestMapRacks_OptionalDimensionFields verifies MapRacks copies the optional
// resolved role and physical-dimension fields (outer unit, width, type) when
// present.
//
// Why it matters: racks anchor the device hierarchy in the imported CANI
// inventory; their physical dimensions and role drive rack elevation rendering
// and placement, so each optional Nautobot field must survive the transform
// rather than silently defaulting to empty.
// Inputs: a single rack with Role.Id, OuterUnit, Width, and Type all set.
// Outputs: the mapped CaniRackType whose Role, OuterUnit, Width, and RackType
// fields are asserted.
// Data choice: distinct, schema-valid enum values ("in", 19, "4-post-cabinet")
// are used so each assertion pins one specific field and the integer Width
// proves the strconv conversion to "19".
func TestMapRacks_OptionalDimensionFields(t *testing.T) {
	nbID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	oaID := openapi_types.UUID(nbID)
	roleID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000002")
	roleURL := "http://api/rack-roles/network/"
	roleRef := makeTenantRefFromUUID(roleID)
	roleRef.Url = strPtr(roleURL)
	outerUnit := nautobotapi.RackOuterUnitValueIn
	width := nautobotapi.RackWidthValueN19
	rackType := nautobotapi.N4PostCabinet

	raw := []nautobotapi.Rack{
		{
			Id:        &oaID,
			Name:      "Rack-Dim",
			Status:    makeStatusRefFromUUID(uuid.New()),
			Location:  makeStatusRefFromUUID(uuid.New()),
			Role:      &roleRef,
			OuterUnit: &nautobotapi.RackOuterUnit{Value: &outerUnit},
			Width:     &nautobotapi.RackWidth{Value: &width},
			Type:      &nautobotapi.RackType{Value: &rackType},
		},
	}

	racks, nbMap := MapRacks(raw, map[uuid.UUID]uuid.UUID{}, nil, map[uuid.UUID]string{roleID: "Network"})
	rack := racks[nbMap[nbID]]
	if rack == nil {
		t.Fatal("rack not found by CANI ID")
	}
	if rack.Role != "Network" {
		t.Errorf("Role = %q, want %q", rack.Role, "Network")
	}
	if rack.OuterUnit != "in" {
		t.Errorf("OuterUnit = %q, want %q", rack.OuterUnit, "in")
	}
	if rack.Width != "19" {
		t.Errorf("Width = %q, want %q", rack.Width, "19")
	}
	if rack.RackType != "4-post-cabinet" {
		t.Errorf("RackType = %q, want %q", rack.RackType, "4-post-cabinet")
	}
}
