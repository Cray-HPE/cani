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

// TestMapPrefixes verifies MapPrefixes converts prefixes, resolves status and
// role names, rewrites the first location to a CANI UUID, resolves the VLAN
// association through the VLAN ID map, maps the prefix type, passes custom
// fields through, and skips prefixes with a nil Id.
//
// Why it matters: prefixes are transformed after VLANs so a prefix can link to
// its VLAN via the Nautobot->CANI VLAN map; a prefix whose VLAN or location is
// not rewritten would point at objects that were never imported, and a prefix
// type that is dropped would lose the container/network/pool classification the
// address plan depends on.
// Inputs: location and VLAN lookup maps plus prefixes that are nil-Id, fully
// populated (known location, VLAN, status, role, type, custom fields), and one
// carrying an unknown VLAN. Outputs: the CANI prefix map with CIDR, length, IP
// version, type, location, VLAN, status, role, and custom fields.
// Data choice: "10.0.0.0/24" type "network" with a populated VLAN map proves the
// VLAN resolves; a deliberately absent VLAN UUID proves the miss leaves VLAN at
// uuid.Nil. Parent is intentionally not asserted because CANI recomputes it.
func TestMapPrefixes(t *testing.T) {
	locNBID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	locCaniID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	locationMap := map[uuid.UUID]uuid.UUID{locNBID: locCaniID}
	vlanNBID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	vlanCaniID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	vlanMap := map[uuid.UUID]uuid.UUID{vlanNBID: vlanCaniID}

	t.Run("empty input returns empty map", func(t *testing.T) {
		if got := MapPrefixes(nil, locationMap, vlanMap, nil, nil); len(got) != 0 {
			t.Fatalf("expected 0 prefixes, got %d", len(got))
		}
	})

	t.Run("prefix with nil ID is skipped", func(t *testing.T) {
		raw := []nautobotapi.Prefix{{Prefix: "10.0.0.0/24", Id: nil, Status: makeStatusRefFromUUID(uuid.New())}}
		if got := MapPrefixes(raw, locationMap, vlanMap, nil, nil); len(got) != 0 {
			t.Errorf("expected 0 prefixes, got %d", len(got))
		}
	})

	t.Run("full prefix is mapped", func(t *testing.T) {
		nbID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
		oaID := openapi_types.UUID(nbID)
		statusID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
		roleID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
		statusNameMap := map[uuid.UUID]string{statusID: "Active"}
		roleNameMap := map[uuid.UUID]string{roleID: "mgmt"}
		role := makeObjectRefFromUUID(roleID)
		vlanRef := makeObjectRefFromUUID(vlanNBID)
		prefixType := nautobotapi.PrefixTypeValue("network")

		raw := []nautobotapi.Prefix{
			{
				Id:           &oaID,
				Prefix:       "10.0.0.0/24",
				PrefixLength: intPtr(24),
				IpVersion:    intPtr(4),
				Description:  strPtr("mgmt net"),
				Status:       makeStatusRefFromUUID(statusID),
				Role:         &role,
				Locations:    &[]nautobotapi.BulkWritableCableRequestStatus{makeStatusRefFromUUID(locNBID)},
				Vlan:         &vlanRef,
				Type:         &nautobotapi.PrefixType{Value: &prefixType},
				CustomFields: &map[string]interface{}{"env": "prod"},
			},
		}

		got := onlyValue(t, MapPrefixes(raw, locationMap, vlanMap, statusNameMap, roleNameMap))
		got.ID = uuid.Nil // normalize randomly generated ID
		want := &devicetypes.CaniPrefix{
			Prefix:      "10.0.0.0/24",
			PrefixLen:   24,
			IPVersion:   4,
			Type:        devicetypes.PrefixTypeNetwork,
			Description: "mgmt net",
			Location:    locCaniID,
			VLAN:        vlanCaniID,
			ObjectMeta: devicetypes.ObjectMeta{
				Status:       "Active",
				Role:         "mgmt",
				CustomFields: map[string]any{"env": "prod"},
				ExternalIDs:  map[string]uuid.UUID{"nautobot": nbID},
			},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("prefix =\n%+v\nwant\n%+v", got, want)
		}
	})

	t.Run("unknown vlan resolves to nil", func(t *testing.T) {
		nbID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
		oaID := openapi_types.UUID(nbID)
		unknownVLAN := makeObjectRefFromUUID(uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"))
		raw := []nautobotapi.Prefix{
			{
				Id:     &oaID,
				Prefix: "10.1.0.0/24",
				Status: makeStatusRefFromUUID(uuid.New()),
				Vlan:   &unknownVLAN,
			},
		}
		got := onlyValue(t, MapPrefixes(raw, locationMap, vlanMap, nil, nil))
		if got.VLAN != uuid.Nil {
			t.Errorf("VLAN = %s, want Nil", got.VLAN)
		}
	})
}
