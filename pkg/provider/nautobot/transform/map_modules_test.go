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

// TestBuildModuleBayMap verifies BuildModuleBayMap indexes module bays by UUID to
// their (parent-device UUID, bay name), skipping nil-Id bays and recording a nil
// device UUID when the bay has no parent device.
//
// Why it matters: MapModules resolves a module's parent device and bay name
// through this map, so a missing bay Id must be skipped and a parentless bay must
// still index by name rather than panic.
// Inputs: nil, a nil-Id bay, a bay with a parent device, and a bay with no parent
// device. Outputs: the UUID->moduleBayRef map.
// Data choice: distinct bay and device UUIDs prove the parent link is captured,
// while the parentless case proves the nil-device guard yields uuid.Nil.
func TestBuildModuleBayMap(t *testing.T) {
	t.Run("empty input returns empty map", func(t *testing.T) {
		got := BuildModuleBayMap(nil)
		if len(got) != 0 {
			t.Errorf("expected 0, got %d", len(got))
		}
	})

	t.Run("module bay with nil ID is skipped", func(t *testing.T) {
		bays := []nautobotapi.ModuleBay{
			{Id: nil, Name: "bay-0"},
		}
		got := BuildModuleBayMap(bays)
		if len(got) != 0 {
			t.Errorf("expected 0, got %d", len(got))
		}
	})

	t.Run("module bays are indexed correctly", func(t *testing.T) {
		bayID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
		devID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		oaBayID := openapi_types.UUID(bayID)
		devRef := makeTenantRefFromUUID(devID)

		bays := []nautobotapi.ModuleBay{
			{Id: &oaBayID, Name: "bay-1", ParentDevice: &devRef},
		}

		got := BuildModuleBayMap(bays)
		if len(got) != 1 {
			t.Fatalf("expected 1, got %d", len(got))
		}
		ref := got[bayID]
		if ref.deviceID != devID {
			t.Errorf("deviceID = %s, want %s", ref.deviceID, devID)
		}
		if ref.name != "bay-1" {
			t.Errorf("name = %q, want %q", ref.name, "bay-1")
		}
	})

	t.Run("module bay without parent device has nil device ID", func(t *testing.T) {
		bayID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
		oaBayID := openapi_types.UUID(bayID)

		bays := []nautobotapi.ModuleBay{
			{Id: &oaBayID, Name: "orphan-bay", ParentDevice: nil},
		}

		got := BuildModuleBayMap(bays)
		ref := got[bayID]
		if ref.deviceID != uuid.Nil {
			t.Errorf("deviceID = %s, want Nil", ref.deviceID)
		}
	})
}

// TestMapModules verifies MapModules converts modules, resolves the parent device
// and bay name through the module-bay map, passes custom fields through, and
// skips modules with a nil Id.
//
// Why it matters: modules are imported after devices and attach via their parent
// bay; the bay reference must resolve to a CANI device and bay name so the module
// lands on the right host, and a nil-Id module must be dropped.
// Inputs: module-bay and device lookup maps plus modules that are nil-Id, carry a
// resolvable parent bay, and carry custom fields. Outputs: the CANI module map
// with serial, asset tag, parent device, bay name, and custom fields.
// Data choice: a bay map whose ref points at a known device UUID proves the
// two-hop bay->device resolution lands on the mapped CANI device rather than a
// coincidental value.
func TestMapModules(t *testing.T) {
	devNBID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	devCaniID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	deviceMap := map[uuid.UUID]uuid.UUID{devNBID: devCaniID}

	bayID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	moduleBayMap := map[uuid.UUID]moduleBayRef{
		bayID: {deviceID: devNBID, name: "bay-1"},
	}

	t.Run("empty input returns empty map", func(t *testing.T) {
		got := MapModules(nil, moduleBayMap, deviceMap, nil, nil, nil)
		if len(got) != 0 {
			t.Errorf("expected 0, got %d", len(got))
		}
	})

	t.Run("module with nil ID is skipped", func(t *testing.T) {
		raw := []nautobotapi.Module{
			{Id: nil, Status: makeStatusRefFromUUID(uuid.New())},
		}
		got := MapModules(raw, moduleBayMap, deviceMap, nil, nil, nil)
		if len(got) != 0 {
			t.Errorf("expected 0, got %d", len(got))
		}
	})

	t.Run("module with parent bay resolves to device", func(t *testing.T) {
		modID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
		oaModID := openapi_types.UUID(modID)
		serial := "MOD-SN1"
		tag := "MOD-TAG"
		statusID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
		bayRef := makeTenantRefFromUUID(bayID)

		raw := []nautobotapi.Module{
			{
				Id:              &oaModID,
				Serial:          &serial,
				AssetTag:        &tag,
				Status:          makeStatusRefFromUUID(statusID),
				ParentModuleBay: &bayRef,
			},
		}

		got := MapModules(raw, moduleBayMap, deviceMap, nil, map[uuid.UUID]string{statusID: "Active"}, nil)
		if len(got) != 1 {
			t.Fatalf("expected 1, got %d", len(got))
		}

		// Find the module (key is a new CANI UUID).
		var mod *struct {
			found bool
		}
		for _, m := range got {
			if m.Serial != "MOD-SN1" {
				continue
			}
			if m.ParentDevice != devCaniID {
				t.Errorf("ParentDevice = %s, want %s", m.ParentDevice, devCaniID)
			}
			if m.ModuleBayName != "bay-1" {
				t.Errorf("ModuleBayName = %q, want %q", m.ModuleBayName, "bay-1")
			}
			if m.AssetTag != "MOD-TAG" {
				t.Errorf("AssetTag = %q, want %q", m.AssetTag, "MOD-TAG")
			}
			if m.Status != "Active" {
				t.Errorf("Status = %q, want %q", m.Status, "Active")
			}
			mod = &struct{ found bool }{true}
		}
		if mod == nil {
			t.Fatal("module not found in result")
		}
	})

	t.Run("custom fields are passed through", func(t *testing.T) {
		modID := uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
		oaModID := openapi_types.UUID(modID)
		cf := map[string]interface{}{"firmware": "v2.1"}

		raw := []nautobotapi.Module{
			{
				Id:           &oaModID,
				Status:       makeStatusRefFromUUID(uuid.New()),
				CustomFields: &cf,
			},
		}

		got := MapModules(raw, moduleBayMap, deviceMap, nil, nil, nil)
		for _, m := range got {
			if m.CustomFields == nil {
				t.Fatal("expected CustomFields to be set")
			}
			if m.CustomFields["firmware"] != "v2.1" {
				t.Errorf("CustomFields[firmware] = %v, want %q", m.CustomFields["firmware"], "v2.1")
			}
		}
	})
}

// TestMapModules_LocationAndRole verifies MapModules resolves a module's
// optional location tenant reference to a CANI location UUID and its role/status
// references to names.
//
// Why it matters: modules (e.g. line cards, NICs) carry their own location and
// role independent of a parent bay; the transform preserves these as CANI-native
// values so imported modules retain placement, lifecycle, and classification.
// Inputs: a module with Location, Role, and Status references set (no parent bay).
// Outputs: the mapped CaniModuleType whose Location equals the CANI location UUID
// and whose Role/Status equal resolved names.
// Data choice: three distinct UUIDs are used for location, role, and status so
// the test proves each reference is read from its own field rather than
// cross-wired.
func TestMapModules_LocationAndRole(t *testing.T) {
	modID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001")
	oaModID := openapi_types.UUID(modID)
	locID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000002")
	locCaniID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000004")
	roleID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000003")
	statusID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000005")
	locRef := makeTenantRefFromUUID(locID)
	roleRef := makeTenantRefFromUUID(roleID)

	raw := []nautobotapi.Module{
		{
			Id:       &oaModID,
			Status:   makeStatusRefFromUUID(statusID),
			Location: &locRef,
			Role:     &roleRef,
		},
	}

	got := MapModules(
		raw,
		map[uuid.UUID]moduleBayRef{},
		map[uuid.UUID]uuid.UUID{},
		map[uuid.UUID]uuid.UUID{locID: locCaniID},
		map[uuid.UUID]string{statusID: "Active"},
		map[uuid.UUID]string{roleID: "Line Card"},
	)
	if len(got) != 1 {
		t.Fatalf("expected 1 module, got %d", len(got))
	}
	for _, m := range got {
		if m.Location != locCaniID {
			t.Errorf("Location = %s, want %s", m.Location, locCaniID)
		}
		if m.Role != "Line Card" {
			t.Errorf("Role = %q, want %q", m.Role, "Line Card")
		}
		if m.Status != "Active" {
			t.Errorf("Status = %q, want %q", m.Status, "Active")
		}
	}
}
