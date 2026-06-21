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
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// TestGroupInterfacesByDevice verifies GroupInterfacesByDevice buckets interfaces
// under their parent device's Nautobot UUID and skips interfaces with no device.
//
// Why it matters: MapDevices attaches interfaces to each device from this
// grouping, so interfaces must land under the correct device UUID and orphaned
// interfaces must be dropped rather than misattached.
// Inputs: nil, an interface with a nil device, and three interfaces split across
// two devices. Outputs: the device-UUID->interface-slice map.
// Data choice: a 2-and-1 split across two distinct devices proves grouping keys
// on device identity rather than coincidentally passing with a single bucket.
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
			t.Fatalf("device1 interfaces = %d, want 2", len(got[devID]))
		}
		if got[devID][0].Name != "eth0" || got[devID][1].Name != "eth1" {
			t.Errorf("device1 interface names = %q/%q, want eth0/eth1", got[devID][0].Name, got[devID][1].Name)
		}
		if len(got[dev2ID]) != 1 {
			t.Fatalf("device2 interfaces = %d, want 1", len(got[dev2ID]))
		}
		if got[dev2ID][0].Name != "mgmt0" {
			t.Errorf("device2 interface name = %q, want mgmt0", got[dev2ID][0].Name)
		}
	})
}

// TestBuildDeviceTypeMap verifies BuildDeviceTypeMap indexes device types by
// their Nautobot UUID and skips entries with a nil Id.
//
// Why it matters: MapDevices resolves each device's model, U-height, depth, and
// slug from its device type via this map, so a device type must be reachable by
// UUID and a nil-Id entry must be excluded rather than indexed under uuid.Nil.
// Inputs: nil, a device type with a UUID, and a nil-Id device type. Outputs: the
// UUID->DeviceType map, asserted by length and model.
// Data choice: a recognizable model ("DL380") confirms the stored value is the
// indexed device type, and the nil-Id case isolates the skip guard.
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

// TestMapDevices verifies MapDevices maps devices end to end - resolving device
// type, rack parent, location, status/role names, comments, and interfaces -
// while skipping nil-Id devices and defaulting face to "front" when a position
// is set without a face.
//
// Why it matters: devices are the core imported entity; their CANI records must
// carry the right model and dimensions, attach to the correct rack and location,
// resolve status/role to names, preserve comments, and obey Nautobot's rule that
// a racked device needs a face, so position-without-face must default rather
// than import as invalid.
// Inputs: rack/location/device-type/status/role lookup maps plus devices that
// are nil-Id, fully populated with one interface, and positioned without a face.
// Outputs: the CANI device map and Nautobot->CANI UUID map, asserted across every
// resolved field.
// Data choice: a device type present in the maps but whose slug is absent from
// the embedded library proves slug resolution returns "" without breaking the
// rest, and the separate position-only device isolates the face default.
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
		comment := "installed from Nautobot"
		pos := 10
		faceVal := nautobotapi.DeviceFaceValue("front")
		rackRef := makeTenantRefFromUUID(rackNBID)
		roleRef := makeStatusRefFromUUID(roleID)

		raw := []nautobotapi.Device{
			{
				Id:         &oaID,
				Name:       &name,
				Serial:     &serial,
				AssetTag:   &tag,
				Status:     makeStatusRefFromUUID(statusID),
				DeviceType: makeStatusRefFromUUID(dtID),
				Comments:   &comment,
				Location:   makeStatusRefFromUUID(locNBID),
				Rack:       &rackRef,
				Position:   &pos,
				Face:       &nautobotapi.DeviceFace{Value: &faceVal},
				Role:       roleRef,
			},
		}

		// Add interfaces for this device.
		devRef := makeTenantRefFromUUID(devNBID)
		ifaceNBID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
		ifaceOAID := openapi_types.UUID(ifaceNBID)
		mgmtOnly := true
		ifType := nautobotapi.InterfaceTypeValue("1000base-t")
		ifaces := map[uuid.UUID][]nautobotapi.Interface{
			devNBID: {
				{Id: &ifaceOAID, Name: "eth0", Device: &devRef, MgmtOnly: &mgmtOnly, Type: nautobotapi.InterfaceType{Value: &ifType}},
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
		if dev.Comments != "installed from Nautobot" {
			t.Errorf("Comments = %q, want %q", dev.Comments, "installed from Nautobot")
		}
		if dev.Description != "" {
			t.Errorf("Description = %q, want empty because Nautobot comments map to CANI Comments", dev.Description)
		}
		if dev.Model != "DL380" {
			t.Errorf("Model = %q, want %q", dev.Model, "DL380")
		}
		// "hpe-dl380" is not in the library, so resolveDeviceSlug returns empty.
		if dev.Slug != "" {
			t.Errorf("Slug = %q, want %q (not in library)", dev.Slug, "")
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
		if dev.Location != locCaniID {
			t.Errorf("Location = %s, want %s (location CANI ID)", dev.Location, locCaniID)
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
		if dev.Interfaces[0].ID != ifaceNBID {
			t.Errorf("Interface ID = %s, want %s", dev.Interfaces[0].ID, ifaceNBID)
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

// TestMapDevices_CustomFields verifies MapDevices copies a device's custom
// fields through to the mapped CaniDeviceType.
//
// Why it matters: devices are the core of the imported CANI inventory; custom
// fields hold operator-defined metadata (e.g. asset owner, rack elevation hints)
// that must not be dropped during the Nautobot-to-CANI transform.
// Inputs: a single device with a non-nil CustomFields map and otherwise minimal
// references. Outputs: the mapped CaniDeviceType whose CustomFields are asserted.
// Data choice: empty rack/location/device-type maps isolate the CustomFields
// branch, proving it runs independently of foreign-key resolution.
func TestMapDevices_CustomFields(t *testing.T) {
	devNBID := uuid.MustParse("dddddddd-0000-0000-0000-000000000001")
	oaID := openapi_types.UUID(devNBID)
	cf := map[string]interface{}{"owner": "hpc-ops"}

	raw := []nautobotapi.Device{
		{
			Id:           &oaID,
			Status:       makeStatusRefFromUUID(uuid.New()),
			DeviceType:   makeStatusRefFromUUID(uuid.New()),
			Role:         makeStatusRefFromUUID(uuid.New()),
			CustomFields: &cf,
		},
	}

	devs, nbMap := MapDevices(
		raw,
		map[uuid.UUID]uuid.UUID{},
		map[uuid.UUID]uuid.UUID{},
		map[uuid.UUID]nautobotapi.DeviceType{},
		nil,
		map[uuid.UUID]string{},
		map[uuid.UUID]string{},
	)
	if len(devs) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devs))
	}
	dev := devs[nbMap[devNBID]]
	if dev.CustomFields == nil {
		t.Fatal("expected CustomFields to be set")
	}
	if dev.CustomFields["owner"] != "hpc-ops" {
		t.Errorf("CustomFields[owner] = %v, want %q", dev.CustomFields["owner"], "hpc-ops")
	}
}

// TestResolveDeviceSlug verifies resolveDeviceSlug matches a Nautobot device
// type to a cani library slug by natural slug first, then case-insensitive
// model name, and returns empty when neither matches.
//
// Why it matters: the slug links an imported device to the cani device-type
// library, which supplies the hardware definition (U-height, ports, FRUs); a
// wrong or missing slug breaks classification, so both resolution strategies and
// the no-match fallback must behave correctly.
// Inputs: synthetic DeviceTypes built from a real library entry. Outputs: the
// resolved slug string.
// Data choice: a known library entry (slug "hpe-dl380-gen-11") is fetched at
// runtime so the fixture stays valid; the model case is upper-cased to prove the
// comparison is case-insensitive rather than an exact-string coincidence.
func TestResolveDeviceSlug(t *testing.T) {
	const knownSlug = "hpe-dl380-gen-11"
	libDT, ok := devicetypes.GetBySlug(knownSlug)
	if !ok {
		t.Fatalf("fixture slug %q missing from device-type library; pick another", knownSlug)
	}

	t.Run("natural slug present in library is returned", func(t *testing.T) {
		dt := nautobotapi.DeviceType{NaturalSlug: strPtr(knownSlug), Model: "unused-model"}
		if got := resolveDeviceSlug(dt); got != knownSlug {
			t.Errorf("resolveDeviceSlug() = %q, want %q", got, knownSlug)
		}
	})

	t.Run("falls back to case-insensitive model match", func(t *testing.T) {
		dt := nautobotapi.DeviceType{Model: strings.ToUpper(libDT.Model)}
		got := resolveDeviceSlug(dt)
		if got == "" {
			t.Fatalf("resolveDeviceSlug() = empty, want a model match for %q", libDT.Model)
		}
		matched, found := devicetypes.GetBySlug(got)
		if !found {
			t.Fatalf("returned slug %q not present in library", got)
		}
		if !strings.EqualFold(matched.Model, libDT.Model) {
			t.Errorf("matched model = %q, want case-insensitive match of %q", matched.Model, libDT.Model)
		}
	})

	t.Run("no slug or model match returns empty", func(t *testing.T) {
		dt := nautobotapi.DeviceType{NaturalSlug: strPtr("zzz-nonexistent-slug"), Model: "zzz-nonexistent-model"}
		if got := resolveDeviceSlug(dt); got != "" {
			t.Errorf("resolveDeviceSlug() = %q, want empty", got)
		}
	})
}
