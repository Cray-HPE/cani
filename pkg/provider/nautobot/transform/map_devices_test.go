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

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func TestGroupInterfacesByDevice(t *testing.T) {
	devID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	devRef := makeTenantRefFromUUID(devID)

	t.Run("empty input returns empty map", func(t *testing.T) {
		got := GroupInterfacesByDevice(nil)
		if len(got) != 0 {
			t.Errorf("expected 0, got %d", len(got))
		}
	})

	t.Run("interface with nil device is skipped", func(t *testing.T) {
		ifaces := []nautobotapi.Interface{
			{Name: "eth0", Device: nil},
		}
		got := GroupInterfacesByDevice(ifaces)
		if len(got) != 0 {
			t.Errorf("expected 0, got %d", len(got))
		}
	})

	t.Run("interfaces grouped by device", func(t *testing.T) {
		dev2ID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		dev2Ref := makeTenantRefFromUUID(dev2ID)
		ifaces := []nautobotapi.Interface{
			{Name: "eth0", Device: &devRef},
			{Name: "eth1", Device: &devRef},
			{Name: "mgmt0", Device: &dev2Ref},
		}

		got := GroupInterfacesByDevice(ifaces)
		if len(got[devID]) != 2 {
			t.Errorf("device1 interfaces = %d, want 2", len(got[devID]))
		}
		if len(got[dev2ID]) != 1 {
			t.Errorf("device2 interfaces = %d, want 1", len(got[dev2ID]))
		}
	})
}

func TestBuildDeviceTypeMap(t *testing.T) {
	t.Run("empty input returns empty map", func(t *testing.T) {
		got := BuildDeviceTypeMap(nil)
		if len(got) != 0 {
			t.Errorf("expected 0, got %d", len(got))
		}
	})

	t.Run("device types indexed by UUID", func(t *testing.T) {
		id := uuid.MustParse("33333333-3333-3333-3333-333333333333")
		oaID := openapi_types.UUID(id)
		dts := []nautobotapi.DeviceType{
			{Id: &oaID, Model: "DL380"},
		}

		got := BuildDeviceTypeMap(dts)
		if len(got) != 1 {
			t.Fatalf("expected 1, got %d", len(got))
		}
		if got[id].Model != "DL380" {
			t.Errorf("Model = %q, want %q", got[id].Model, "DL380")
		}
	})

	t.Run("nil ID device type is skipped", func(t *testing.T) {
		dts := []nautobotapi.DeviceType{
			{Id: nil, Model: "orphan"},
		}
		got := BuildDeviceTypeMap(dts)
		if len(got) != 0 {
			t.Errorf("expected 0, got %d", len(got))
		}
	})
}

func TestMapDevices(t *testing.T) {
	rackNBID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	rackCaniID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	rackMap := map[uuid.UUID]uuid.UUID{rackNBID: rackCaniID}

	locNBID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	locCaniID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	locationMap := map[uuid.UUID]uuid.UUID{locNBID: locCaniID}

	statusID := uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	roleID := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
	statusNameMap := map[uuid.UUID]string{statusID: "Active"}
	roleNameMap := map[uuid.UUID]string{roleID: "Compute"}

	dtID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	slug := "hpe-dl380"
	uHeight := 2
	fullDepth := true
	deviceTypeMap := map[uuid.UUID]nautobotapi.DeviceType{
		dtID: {Model: "DL380", NaturalSlug: &slug, UHeight: &uHeight, IsFullDepth: &fullDepth, Manufacturer: makeStatusRefFromUUID(uuid.New())},
	}

	t.Run("empty input returns empty maps", func(t *testing.T) {
		devs, nbMap := MapDevices(nil, rackMap, locationMap, deviceTypeMap, nil, statusNameMap, roleNameMap)
		if len(devs) != 0 || len(nbMap) != 0 {
			t.Errorf("expected empty, got %d devs, %d map entries", len(devs), len(nbMap))
		}
	})

	t.Run("device with nil ID is skipped", func(t *testing.T) {
		raw := []nautobotapi.Device{
			{Id: nil, Status: makeStatusRefFromUUID(statusID), DeviceType: makeStatusRefFromUUID(dtID), Role: makeStatusRefFromUUID(roleID)},
		}
		devs, _ := MapDevices(raw, rackMap, locationMap, deviceTypeMap, nil, statusNameMap, roleNameMap)
		if len(devs) != 0 {
			t.Errorf("expected 0, got %d", len(devs))
		}
	})

	t.Run("full device mapping", func(t *testing.T) {
		devNBID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
		oaID := openapi_types.UUID(devNBID)
		name := "server-1"
		serial := "SN001"
		tag := "TAG001"
		pos := 10
		faceVal := nautobotapi.DeviceFaceValue("front")
		rackRef := makeTenantRefFromUUID(rackNBID)
		roleRef := makeStatusRefFromUUID(roleID)
		roleURL := "http://api/roles/compute/"
		roleRef.Url = &roleURL

		raw := []nautobotapi.Device{
			{
				Id:         &oaID,
				Name:       &name,
				Serial:     &serial,
				AssetTag:   &tag,
				Status:     makeStatusRefFromUUID(statusID),
				DeviceType: makeStatusRefFromUUID(dtID),
				Rack:       &rackRef,
				Position:   &pos,
				Face:       &nautobotapi.DeviceFace{Value: &faceVal},
				Role:       roleRef,
			},
		}

		// Add interfaces for this device.
		devRef := makeTenantRefFromUUID(devNBID)
		mgmtOnly := true
		ifType := nautobotapi.InterfaceTypeValue("1000base-t")
		ifaces := map[uuid.UUID][]nautobotapi.Interface{
			devNBID: {
				{Name: "eth0", Device: &devRef, MgmtOnly: &mgmtOnly, Type: nautobotapi.InterfaceType{Value: &ifType}},
			},
		}

		devs, nbMap := MapDevices(raw, rackMap, locationMap, deviceTypeMap, ifaces, statusNameMap, roleNameMap)
		if len(devs) != 1 {
			t.Fatalf("expected 1 device, got %d", len(devs))
		}

		caniID := nbMap[devNBID]
		dev := devs[caniID]

		if dev.Name != "server-1" {
			t.Errorf("Name = %q, want %q", dev.Name, "server-1")
		}
		if dev.Serial != "SN001" {
			t.Errorf("Serial = %q, want %q", dev.Serial, "SN001")
		}
		if dev.AssetTag != "TAG001" {
			t.Errorf("AssetTag = %q, want %q", dev.AssetTag, "TAG001")
		}
		if dev.Model != "DL380" {
			t.Errorf("Model = %q, want %q", dev.Model, "DL380")
		}
		if dev.Slug != "hpe-dl380" {
			t.Errorf("Slug = %q, want %q", dev.Slug, "hpe-dl380")
		}
		if dev.UHeight != 2 {
			t.Errorf("UHeight = %d, want 2", dev.UHeight)
		}
		if !dev.IsFullDepth {
			t.Error("IsFullDepth = false, want true")
		}
		if dev.Parent != rackCaniID {
			t.Errorf("Parent = %s, want %s (rack CANI ID)", dev.Parent, rackCaniID)
		}
		if dev.RackPosition != 10 {
			t.Errorf("RackPosition = %d, want 10", dev.RackPosition)
		}
		if dev.Face != "front" {
			t.Errorf("Face = %q, want %q", dev.Face, "front")
		}
		if dev.ObjectMeta.Status != "Active" {
			t.Errorf("Status = %q, want %q", dev.ObjectMeta.Status, "Active")
		}
		if dev.Role != "Compute" {
			t.Errorf("Role = %q, want %q", dev.Role, "Compute")
		}
		if len(dev.Interfaces) != 1 {
			t.Fatalf("expected 1 interface, got %d", len(dev.Interfaces))
		}
		if dev.Interfaces[0].Name != "eth0" {
			t.Errorf("Interface Name = %q, want %q", dev.Interfaces[0].Name, "eth0")
		}
		if dev.Interfaces[0].MgmtOnly == nil || !*dev.Interfaces[0].MgmtOnly {
			t.Error("Interface MgmtOnly should be true")
		}
	})

	t.Run("device with position but no face defaults to front", func(t *testing.T) {
		devNBID := uuid.MustParse("88888888-8888-8888-8888-888888888888")
		oaID := openapi_types.UUID(devNBID)
		pos := 5
		rackRef := makeTenantRefFromUUID(rackNBID)

		raw := []nautobotapi.Device{
			{
				Id:         &oaID,
				Status:     makeStatusRefFromUUID(statusID),
				DeviceType: makeStatusRefFromUUID(dtID),
				Rack:       &rackRef,
				Position:   &pos,
				Role:       makeStatusRefFromUUID(roleID),
			},
		}

		devs, nbMap := MapDevices(raw, rackMap, locationMap, deviceTypeMap, nil, statusNameMap, roleNameMap)
		caniID := nbMap[devNBID]
		dev := devs[caniID]

		if dev.Face != "front" {
			t.Errorf("Face = %q, want %q (default)", dev.Face, "front")
		}
	})
}
