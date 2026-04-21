/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func TestTransform(t *testing.T) {
	t.Run("nil raw data uses legacy path", func(t *testing.T) {
		devID := uuid.New()
		existing := devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
				devID: {ID: devID, Name: "server-1"},
			},
		}

		result, err := Transform(existing, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if len(result.Devices) != 1 {
			t.Errorf("expected 1 device, got %d", len(result.Devices))
		}
		if result.Devices[devID] == nil {
			t.Error("expected device to be present in result")
		}
		if result.Devices[devID].Name != "server-1" {
			t.Errorf("Name = %q, want %q", result.Devices[devID].Name, "server-1")
		}
	})

	t.Run("empty raw data uses raw path with empty results", func(t *testing.T) {
		raw := &RawData{}
		result, err := Transform(devicetypes.Inventory{}, raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if len(result.Locations) != 0 {
			t.Errorf("expected 0 locations, got %d", len(result.Locations))
		}
		if len(result.Devices) != 0 {
			t.Errorf("expected 0 devices, got %d", len(result.Devices))
		}
	})

	t.Run("end-to-end transform with raw data", func(t *testing.T) {
		locNBID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
		rackNBID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		devNBID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
		dtNBID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
		statusID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
		roleID := uuid.MustParse("66666666-6666-6666-6666-666666666666")

		oaLocID := openapi_types.UUID(locNBID)
		oaRackID := openapi_types.UUID(rackNBID)
		oaDevID := openapi_types.UUID(devNBID)
		oaDtID := openapi_types.UUID(dtNBID)
		oaStatusID := openapi_types.UUID(statusID)
		oaRoleID := openapi_types.UUID(roleID)

		devName := "compute-001"
		rackRef := makeTenantRefFromUUID(rackNBID)

		raw := &RawData{
			Locations: []nautobotapi.Location{
				{Id: &oaLocID, Name: "Site-A", Status: makeStatusRefFromUUID(statusID), LocationType: makeStatusRefFromUUID(uuid.New())},
			},
			Racks: []nautobotapi.Rack{
				{Id: &oaRackID, Name: "Rack-1", Status: makeStatusRefFromUUID(statusID), Location: makeStatusRefFromUUID(locNBID)},
			},
			Devices: []nautobotapi.Device{
				{
					Id:         &oaDevID,
					Name:       &devName,
					Status:     makeStatusRefFromUUID(statusID),
					DeviceType: makeStatusRefFromUUID(dtNBID),
					Rack:       &rackRef,
					Role: func() nautobotapi.BulkWritableCableRequestStatus {
						r := makeStatusRefFromUUID(roleID)
						url := "http://api/roles/compute/"
						r.Url = &url
						return r
					}(),
				},
			},
			DeviceTypes: []nautobotapi.DeviceType{
				{Id: &oaDtID, Model: "DL380", Manufacturer: makeStatusRefFromUUID(uuid.New())},
			},
			Statuses: []nautobotapi.Status{
				{Id: &oaStatusID, Name: "Active"},
			},
			Roles: []nautobotapi.Role{
				{Id: &oaRoleID, Name: "Compute"},
			},
		}

		result, err := Transform(devicetypes.Inventory{}, raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Locations) != 1 {
			t.Errorf("expected 1 location, got %d", len(result.Locations))
		}
		if len(result.Racks) != 1 {
			t.Errorf("expected 1 rack, got %d", len(result.Racks))
		}
		if len(result.Devices) != 1 {
			t.Errorf("expected 1 device, got %d", len(result.Devices))
		}

		// Verify device has correct model and role.
		for _, dev := range result.Devices {
			if dev.Name != "compute-001" {
				t.Errorf("Device.Name = %q, want %q", dev.Name, "compute-001")
			}
			if dev.Model != "DL380" {
				t.Errorf("Device.Model = %q, want %q", dev.Model, "DL380")
			}
			if dev.ObjectMeta.Status != "Active" {
				t.Errorf("Device.Status = %q, want %q", dev.ObjectMeta.Status, "Active")
			}
			if dev.Role != "Compute" {
				t.Errorf("Device.Role = %q, want %q", dev.Role, "Compute")
			}
		}
	})
}

func TestTransformLegacy(t *testing.T) {
	t.Run("empty inventory returns empty result", func(t *testing.T) {
		result, err := transformLegacy(devicetypes.Inventory{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Devices) != 0 {
			t.Errorf("expected 0 devices, got %d", len(result.Devices))
		}
	})

	t.Run("devices are copied to result", func(t *testing.T) {
		id1 := uuid.New()
		id2 := uuid.New()
		existing := devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
				id1: {ID: id1, Name: "server-1"},
				id2: {ID: id2, Name: "server-2"},
			},
		}

		result, err := transformLegacy(existing)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Devices) != 2 {
			t.Errorf("expected 2 devices, got %d", len(result.Devices))
		}
	})
}
