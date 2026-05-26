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

func TestMapModules(t *testing.T) {
	devNBID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	devCaniID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	deviceMap := map[uuid.UUID]uuid.UUID{devNBID: devCaniID}

	bayID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	moduleBayMap := map[uuid.UUID]moduleBayRef{
		bayID: {deviceID: devNBID, name: "bay-1"},
	}

	t.Run("empty input returns empty map", func(t *testing.T) {
		got := MapModules(nil, moduleBayMap, deviceMap)
		if len(got) != 0 {
			t.Errorf("expected 0, got %d", len(got))
		}
	})

	t.Run("module with nil ID is skipped", func(t *testing.T) {
		raw := []nautobotapi.Module{
			{Id: nil, Status: makeStatusRefFromUUID(uuid.New())},
		}
		got := MapModules(raw, moduleBayMap, deviceMap)
		if len(got) != 0 {
			t.Errorf("expected 0, got %d", len(got))
		}
	})

	t.Run("module with parent bay resolves to device", func(t *testing.T) {
		modID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
		oaModID := openapi_types.UUID(modID)
		serial := "MOD-SN1"
		tag := "MOD-TAG"
		bayRef := makeTenantRefFromUUID(bayID)

		raw := []nautobotapi.Module{
			{
				Id:              &oaModID,
				Serial:          &serial,
				AssetTag:        &tag,
				Status:          makeStatusRefFromUUID(uuid.New()),
				ParentModuleBay: &bayRef,
			},
		}

		got := MapModules(raw, moduleBayMap, deviceMap)
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

		got := MapModules(raw, moduleBayMap, deviceMap)
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
