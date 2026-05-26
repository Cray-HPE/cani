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

func TestMapRacks(t *testing.T) {
	locNBID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	locCaniID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	locationMap := map[uuid.UUID]uuid.UUID{locNBID: locCaniID}

	t.Run("empty input returns empty maps", func(t *testing.T) {
		racks, nbMap := MapRacks(nil, locationMap)
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
		racks, _ := MapRacks(raw, locationMap)
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

		raw := []nautobotapi.Rack{
			{
				Id:       &oaID,
				Name:     "Rack-A",
				Serial:   &serial,
				AssetTag: &tag,
				Comments: &comment,
				Status:   makeStatusRefFromUUID(uuid.New()),
				Location: makeStatusRefFromUUID(locNBID),
			},
		}

		racks, nbMap := MapRacks(raw, locationMap)
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

		racks, nbMap := MapRacks(raw, locationMap)
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

		racks, nbMap := MapRacks(raw, locationMap)
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
