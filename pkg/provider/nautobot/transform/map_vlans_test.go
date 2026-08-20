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
	"reflect"
	"testing"

	openapi_types "github.com/Cray-HPE/cani/internal/openapi/types"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// TestMapVLANs verifies MapVLANs converts VLANs, resolves each VLAN's status and
// role name, rewrites its first location to a CANI UUID, passes custom fields
// through, records the Nautobot->CANI ID mapping, and skips VLANs with a nil Id.
//
// Why it matters: VLANs are transformed before prefixes so that a prefix can
// resolve its VLAN association through the returned ID map; a VLAN whose location
// is not rewritten (or whose nil Id is not skipped) would corrupt the imported
// L2 topology or point at a location that was never imported.
// Inputs: a location lookup map plus VLANs that are nil-Id, fully populated
// (known location, status, role, custom fields), and one carrying an unknown
// location. Outputs: the CANI VLAN map and Nautobot->CANI UUID map with VID,
// name, description, status, role, location, and custom fields.
// Data choice: VID 100 ("prod") with a populated location map proves the known
// location resolves; a deliberately absent location UUID proves the miss leaves
// Location at uuid.Nil.
func TestMapVLANs(t *testing.T) {
	locNBID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	locCaniID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	locationMap := map[uuid.UUID]uuid.UUID{locNBID: locCaniID}

	t.Run("empty input returns empty maps", func(t *testing.T) {
		vlans, nbMap := MapVLANs(nil, locationMap, nil, nil)
		if len(vlans) != 0 || len(nbMap) != 0 {
			t.Fatalf("expected empty maps, got %d vlans, %d mappings", len(vlans), len(nbMap))
		}
	})

	t.Run("vlan with nil ID is skipped", func(t *testing.T) {
		raw := []nautobotapi.VLAN{{Name: "orphan", Vid: 10, Id: nil, Status: makeStatusRefFromUUID(uuid.New())}}
		vlans, _ := MapVLANs(raw, locationMap, nil, nil)
		if len(vlans) != 0 {
			t.Errorf("expected 0 vlans, got %d", len(vlans))
		}
	})

	t.Run("full vlan is mapped", func(t *testing.T) {
		nbID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
		oaID := openapi_types.UUID(nbID)
		statusID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
		roleID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
		statusNameMap := map[uuid.UUID]string{statusID: "Active"}
		roleNameMap := map[uuid.UUID]string{roleID: "server"}
		role := makeObjectRefFromUUID(roleID)

		raw := []nautobotapi.VLAN{
			{
				Id:           &oaID,
				Vid:          100,
				Name:         "prod",
				Description:  strPtr("production vlan"),
				Status:       makeStatusRefFromUUID(statusID),
				Role:         &role,
				Locations:    &[]nautobotapi.BulkWritableCableRequestStatus{makeStatusRefFromUUID(locNBID)},
				CustomFields: &map[string]interface{}{"zone": "a"},
			},
		}

		vlans, nbMap := MapVLANs(raw, locationMap, statusNameMap, roleNameMap)
		got := vlans[nbMap[nbID]]
		if got == nil {
			t.Fatal("vlan not found by CANI ID")
		}
		got.ID = uuid.Nil // normalize randomly generated ID
		want := &devicetypes.CaniVLAN{
			VID:         100,
			Name:        "prod",
			Description: "production vlan",
			Location:    locCaniID,
			ObjectMeta: devicetypes.ObjectMeta{
				Status:       "Active",
				Role:         "server",
				CustomFields: map[string]any{"zone": "a"},
				ExternalIDs:  map[string]uuid.UUID{"nautobot": nbID},
			},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("vlan =\n%+v\nwant\n%+v", got, want)
		}
	})

	t.Run("unknown location resolves to nil", func(t *testing.T) {
		nbID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
		oaID := openapi_types.UUID(nbID)
		unknownLoc := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
		raw := []nautobotapi.VLAN{
			{
				Id:        &oaID,
				Vid:       200,
				Name:      "floating",
				Status:    makeStatusRefFromUUID(uuid.New()),
				Locations: &[]nautobotapi.BulkWritableCableRequestStatus{makeStatusRefFromUUID(unknownLoc)},
			},
		}
		vlans, _ := MapVLANs(raw, locationMap, nil, nil)
		if got := onlyValue(t, vlans); got.Location != uuid.Nil {
			t.Errorf("Location = %s, want Nil", got.Location)
		}
	})
}
