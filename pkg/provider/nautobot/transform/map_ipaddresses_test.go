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

// TestMapIPAddresses verifies MapIPAddresses converts IP addresses, derives the
// bare host from the CIDR address, resolves status and role (as IPRole), maps the
// address type, passes custom fields through, and skips addresses with a nil Id.
//
// Why it matters: an imported address must retain both its full CIDR and the bare
// host used for interface assignment; if the host were not split from the mask or
// the type/role were dropped, downstream interface and DNS wiring would bind to
// malformed or unclassified addresses.
// Inputs: IP addresses that are nil-Id, fully populated ("10.0.0.5/24" with type,
// role, status, DNS name, custom fields), and one address with no mask. Outputs:
// the CANI IP map with address, host, mask length, IP version, type, IP role,
// status, DNS name, and custom fields.
// Data choice: "10.0.0.5/24" proves the host splits to "10.0.0.5"; a maskless
// "10.0.0.9" proves hostFromAddress returns the input unchanged.
func TestMapIPAddresses(t *testing.T) {
	t.Run("empty input returns empty map", func(t *testing.T) {
		if got := MapIPAddresses(nil, nil, nil); len(got) != 0 {
			t.Fatalf("expected 0 addresses, got %d", len(got))
		}
	})

	t.Run("address with nil ID is skipped", func(t *testing.T) {
		raw := []nautobotapi.IPAddress{{Address: "10.0.0.1/24", Id: nil, Status: makeStatusRefFromUUID(uuid.New())}}
		if got := MapIPAddresses(raw, nil, nil); len(got) != 0 {
			t.Errorf("expected 0 addresses, got %d", len(got))
		}
	})

	t.Run("full address is mapped", func(t *testing.T) {
		nbID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
		oaID := openapi_types.UUID(nbID)
		statusID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
		roleID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
		statusNameMap := map[uuid.UUID]string{statusID: "Active"}
		roleNameMap := map[uuid.UUID]string{roleID: "loopback"}
		role := makeObjectRefFromUUID(roleID)
		ipType := nautobotapi.IPAddressTypeChoices("host")

		raw := []nautobotapi.IPAddress{
			{
				Id:           &oaID,
				Address:      "10.0.0.5/24",
				MaskLength:   intPtr(24),
				IpVersion:    intPtr(4),
				DnsName:      strPtr("host1.example.com"),
				Description:  strPtr("loopback addr"),
				Status:       makeStatusRefFromUUID(statusID),
				Role:         &role,
				Type:         &ipType,
				CustomFields: &map[string]interface{}{"rack": "r1"},
			},
		}

		got := onlyValue(t, MapIPAddresses(raw, statusNameMap, roleNameMap))
		got.ID = uuid.Nil // normalize randomly generated ID
		want := &devicetypes.CaniIPAddress{
			Address:     "10.0.0.5/24",
			Host:        "10.0.0.5",
			MaskLength:  24,
			IPVersion:   4,
			Type:        devicetypes.IPAddressTypeHost,
			IPRole:      devicetypes.IPAddressRoleLoopback,
			DNSName:     "host1.example.com",
			Description: "loopback addr",
			ObjectMeta: devicetypes.ObjectMeta{
				Status:       "Active",
				CustomFields: map[string]any{"rack": "r1"},
				ExternalIDs:  map[string]uuid.UUID{"nautobot": nbID},
			},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("ip =\n%+v\nwant\n%+v", got, want)
		}
	})

	t.Run("address without mask keeps host unchanged", func(t *testing.T) {
		nbID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
		oaID := openapi_types.UUID(nbID)
		raw := []nautobotapi.IPAddress{
			{Id: &oaID, Address: "10.0.0.9", Status: makeStatusRefFromUUID(uuid.New())},
		}
		got := onlyValue(t, MapIPAddresses(raw, nil, nil))
		if got.Host != "10.0.0.9" {
			t.Errorf("Host = %q, want 10.0.0.9", got.Host)
		}
	})
}
