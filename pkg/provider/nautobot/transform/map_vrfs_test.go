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

	openapi_types "github.com/Cray-HPE/cani/internal/openapi/types"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

func TestMapVRFs(t *testing.T) {
	statusID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	statusNameMap := map[uuid.UUID]string{statusID: "Active"}

	t.Run("empty input returns empty map", func(t *testing.T) {
		result := MapVRFs(nil, nil, nil, statusNameMap)
		if len(result) != 0 {
			t.Fatalf("expected empty, got %d", len(result))
		}
	})

	t.Run("vrf with nil ID is skipped", func(t *testing.T) {
		raw := []nautobotapi.VRF{{Name: "orphan", Id: nil}}
		result := MapVRFs(raw, nil, nil, statusNameMap)
		if len(result) != 0 {
			t.Errorf("expected 0, got %d", len(result))
		}
	})

	t.Run("basic VRF fields are mapped", func(t *testing.T) {
		nbID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
		oaID := openapi_types.UUID(nbID)
		rd := "65000:1"
		desc := "test vrf"
		statusRef := makeObjectRefFromUUID(statusID)

		raw := []nautobotapi.VRF{{
			Id:          &oaID,
			Name:        "PROD",
			Rd:          &rd,
			Description: &desc,
			Status:      &statusRef,
		}}

		result := MapVRFs(raw, nil, nil, statusNameMap)
		if len(result) != 1 {
			t.Fatalf("expected 1 VRF, got %d", len(result))
		}
		for _, v := range result {
			if v.Name != "PROD" {
				t.Errorf("Name = %q, want PROD", v.Name)
			}
			if v.RD != "65000:1" {
				t.Errorf("RD = %q, want 65000:1", v.RD)
			}
			if v.Description != "test vrf" {
				t.Errorf("Description = %q, want 'test vrf'", v.Description)
			}
			if v.ObjectMeta.Status != "Active" {
				t.Errorf("Status = %q, want Active", v.ObjectMeta.Status)
			}
			if v.ObjectMeta.ExternalIDs["nautobot"] != nbID {
				t.Errorf("ExternalIDs[nautobot] = %v, want %v", v.ObjectMeta.ExternalIDs["nautobot"], nbID)
			}
		}
	})

	t.Run("device assignments are resolved", func(t *testing.T) {
		vrfNbID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		oaID := openapi_types.UUID(vrfNbID)
		deviceNbID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
		deviceCaniID := uuid.MustParse("44444444-4444-4444-4444-444444444444")

		raw := []nautobotapi.VRF{{Id: &oaID, Name: "VRF1"}}
		deviceMap := map[uuid.UUID]uuid.UUID{deviceNbID: deviceCaniID}
		deviceRef := makeObjectRefFromUUID(deviceNbID)
		assignments := []nautobotapi.VRFDeviceAssignment{{
			Vrf:    makeStatusRefFromUUID(vrfNbID),
			Device: &deviceRef,
		}}

		result := MapVRFs(raw, assignments, deviceMap, statusNameMap)
		if len(result) != 1 {
			t.Fatalf("expected 1 VRF, got %d", len(result))
		}
		for _, v := range result {
			if len(v.Devices) != 1 {
				t.Fatalf("expected 1 device, got %d", len(v.Devices))
			}
			if v.Devices[0] != deviceCaniID {
				t.Errorf("Devices[0] = %v, want %v", v.Devices[0], deviceCaniID)
			}
		}
	})

	t.Run("assignment with nil device is skipped", func(t *testing.T) {
		vrfNbID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
		oaID := openapi_types.UUID(vrfNbID)

		raw := []nautobotapi.VRF{{Id: &oaID, Name: "VRF2"}}
		assignments := []nautobotapi.VRFDeviceAssignment{{
			Vrf:    makeStatusRefFromUUID(vrfNbID),
			Device: nil,
		}}

		result := MapVRFs(raw, assignments, nil, statusNameMap)
		for _, v := range result {
			if len(v.Devices) != 0 {
				t.Errorf("expected 0 devices, got %d", len(v.Devices))
			}
		}
	})

	t.Run("assignment for unknown VRF is skipped", func(t *testing.T) {
		vrfNbID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
		oaID := openapi_types.UUID(vrfNbID)
		unknownVRF := uuid.MustParse("77777777-7777-7777-7777-777777777777")
		deviceNbID := uuid.MustParse("88888888-8888-8888-8888-888888888888")

		raw := []nautobotapi.VRF{{Id: &oaID, Name: "VRF3"}}
		deviceRef := makeObjectRefFromUUID(deviceNbID)
		assignments := []nautobotapi.VRFDeviceAssignment{{
			Vrf:    makeStatusRefFromUUID(unknownVRF),
			Device: &deviceRef,
		}}
		deviceMap := map[uuid.UUID]uuid.UUID{deviceNbID: uuid.New()}

		result := MapVRFs(raw, assignments, deviceMap, statusNameMap)
		for _, v := range result {
			if len(v.Devices) != 0 {
				t.Errorf("expected 0 devices, got %d", len(v.Devices))
			}
		}
	})

	t.Run("assignment for unmapped device is skipped", func(t *testing.T) {
		vrfNbID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
		oaID := openapi_types.UUID(vrfNbID)
		deviceNbID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")

		raw := []nautobotapi.VRF{{Id: &oaID, Name: "VRF4"}}
		deviceRef := makeObjectRefFromUUID(deviceNbID)
		assignments := []nautobotapi.VRFDeviceAssignment{{
			Vrf:    makeStatusRefFromUUID(vrfNbID),
			Device: &deviceRef,
		}}
		// deviceMap does not contain deviceNbID
		deviceMap := map[uuid.UUID]uuid.UUID{}

		result := MapVRFs(raw, assignments, deviceMap, statusNameMap)
		for _, v := range result {
			if len(v.Devices) != 0 {
				t.Errorf("expected 0 devices, got %d", len(v.Devices))
			}
		}
	})

	t.Run("duplicate assignments are deduplicated", func(t *testing.T) {
		vrfNbID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
		oaID := openapi_types.UUID(vrfNbID)
		deviceNbID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
		deviceCaniID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")

		raw := []nautobotapi.VRF{{Id: &oaID, Name: "VRF-DUP"}}
		deviceMap := map[uuid.UUID]uuid.UUID{deviceNbID: deviceCaniID}
		deviceRef := makeObjectRefFromUUID(deviceNbID)
		assignments := []nautobotapi.VRFDeviceAssignment{
			{Vrf: makeStatusRefFromUUID(vrfNbID), Device: &deviceRef},
			{Vrf: makeStatusRefFromUUID(vrfNbID), Device: &deviceRef},
		}

		result := MapVRFs(raw, assignments, deviceMap, statusNameMap)
		for _, v := range result {
			if len(v.Devices) != 1 {
				t.Errorf("expected 1 device (deduplicated), got %d", len(v.Devices))
			}
		}
	})
}
