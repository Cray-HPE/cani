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

// TestMapFrus verifies MapFrus converts inventory items to CANI FRUs, resolving
// the device reference to a CANI UUID, preserving parent and manufacturer
// references, and skipping items with a nil Id.
//
// Why it matters: FRUs are imported after devices and hang off them; the device
// reference must be re-pointed to the CANI device while unknown devices degrade
// to uuid.Nil, so a FRU is never attached to a device that was not imported.
// Inputs: a device lookup map plus inventory items that are nil-Id, fully
// populated, reference an unknown device, carry a parent, and carry a
// manufacturer. Outputs: the CANI FRU map with name, label, serial, part, and
// resolved references.
// Data choice: a known device UUID proves resolution, an unknown UUID proves the
// miss leaves Device at uuid.Nil, and separate parent/manufacturer cases prove
// those references are retained as their Nautobot UUIDs.
func TestMapFrus(t *testing.T) {
	devNBID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	devCaniID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	deviceMap := map[uuid.UUID]uuid.UUID{devNBID: devCaniID}

	t.Run("empty input returns empty map", func(t *testing.T) {
		got := MapFrus(nil, deviceMap)
		if len(got) != 0 {
			t.Errorf("expected 0, got %d", len(got))
		}
	})

	t.Run("item with nil ID is skipped", func(t *testing.T) {
		raw := []nautobotapi.InventoryItem{
			{Id: nil, Name: "orphan", Device: makeStatusRefFromUUID(uuid.New())},
		}
		got := MapFrus(raw, deviceMap)
		if len(got) != 0 {
			t.Errorf("expected 0, got %d", len(got))
		}
	})

	t.Run("full inventory item mapping", func(t *testing.T) {
		itemID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
		oaItemID := openapi_types.UUID(itemID)
		label := "GPU-0"
		serial := "GPU-SN1"
		desc := "NVIDIA A100"
		partID := "PN-A100"
		tag := "GPU-ASSET"
		discovered := true

		raw := []nautobotapi.InventoryItem{
			{
				Id:          &oaItemID,
				Name:        "gpu-0",
				Label:       &label,
				Serial:      &serial,
				Description: &desc,
				PartId:      &partID,
				AssetTag:    &tag,
				Discovered:  &discovered,
				Device:      makeStatusRefFromUUID(devNBID),
			},
		}

		got := MapFrus(raw, deviceMap)
		if len(got) != 1 {
			t.Fatalf("expected 1, got %d", len(got))
		}

		for _, fru := range got {
			if fru.Name != "gpu-0" {
				t.Errorf("Name = %q, want %q", fru.Name, "gpu-0")
			}
			if fru.Label != "GPU-0" {
				t.Errorf("Label = %q, want %q", fru.Label, "GPU-0")
			}
			if fru.Serial != "GPU-SN1" {
				t.Errorf("Serial = %q, want %q", fru.Serial, "GPU-SN1")
			}
			if fru.Description != "NVIDIA A100" {
				t.Errorf("Description = %q, want %q", fru.Description, "NVIDIA A100")
			}
			if fru.PartNumber != "PN-A100" {
				t.Errorf("PartNumber = %q, want %q", fru.PartNumber, "PN-A100")
			}
			if fru.AssetTag != "GPU-ASSET" {
				t.Errorf("AssetTag = %q, want %q", fru.AssetTag, "GPU-ASSET")
			}
			if !fru.Discovered {
				t.Error("Discovered = false, want true")
			}
			if fru.Device != devCaniID {
				t.Errorf("Device = %s, want %s", fru.Device, devCaniID)
			}
		}
	})

	t.Run("unknown device reference is not resolved", func(t *testing.T) {
		unknownDevID := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
		itemID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
		oaItemID := openapi_types.UUID(itemID)

		raw := []nautobotapi.InventoryItem{
			{
				Id:     &oaItemID,
				Name:   "orphan-fru",
				Device: makeStatusRefFromUUID(unknownDevID),
			},
		}

		got := MapFrus(raw, deviceMap)
		for _, fru := range got {
			if fru.Device != uuid.Nil {
				t.Errorf("Device = %s, want Nil for unknown device", fru.Device)
			}
		}
	})

	t.Run("parent inventory item stores nautobot UUID", func(t *testing.T) {
		parentNBID := uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
		itemID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
		oaItemID := openapi_types.UUID(itemID)
		parentRef := makeTenantRefFromUUID(parentNBID)

		raw := []nautobotapi.InventoryItem{
			{
				Id:     &oaItemID,
				Name:   "child-fru",
				Parent: &parentRef,
				Device: makeStatusRefFromUUID(devNBID),
			},
		}

		got := MapFrus(raw, deviceMap)
		for _, fru := range got {
			if fru.Parent != parentNBID {
				t.Errorf("Parent = %s, want %s", fru.Parent, parentNBID)
			}
		}
	})

	t.Run("manufacturer is extracted", func(t *testing.T) {
		mfgID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
		itemID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
		oaItemID := openapi_types.UUID(itemID)
		mfgRef := makeTenantRefFromUUID(mfgID)

		raw := []nautobotapi.InventoryItem{
			{
				Id:           &oaItemID,
				Name:         "mfg-fru",
				Manufacturer: &mfgRef,
				Device:       makeStatusRefFromUUID(devNBID),
			},
		}

		got := MapFrus(raw, deviceMap)
		for _, fru := range got {
			if fru.Manufacturer != mfgID.String() {
				t.Errorf("Manufacturer = %q, want %q", fru.Manufacturer, mfgID.String())
			}
		}
	})
}

// TestMapFrus_CustomFields verifies MapFrus copies a FRU's custom fields through
// to the mapped CaniFruType.
//
// Why it matters: inventory items (FRUs) frequently carry site-specific custom
// fields (revision, firmware, warranty); the transform must preserve them so the
// imported CANI inventory does not lose operator metadata.
// Inputs: a single inventory item with a non-nil CustomFields map.
// Outputs: the mapped CaniFruType whose CustomFields entry is asserted.
// Data choice: a single keyed value ("rev":"A1") is the minimal map that proves
// the pointer is dereferenced and the contents are carried, not just allocated.
func TestMapFrus_CustomFields(t *testing.T) {
	itemID := uuid.MustParse("cccccccc-0000-0000-0000-000000000001")
	oaItemID := openapi_types.UUID(itemID)
	cf := map[string]interface{}{"rev": "A1"}

	raw := []nautobotapi.InventoryItem{
		{
			Id:           &oaItemID,
			Name:         "cf-fru",
			CustomFields: &cf,
			Device:       makeStatusRefFromUUID(uuid.New()),
		},
	}

	got := MapFrus(raw, map[uuid.UUID]uuid.UUID{})
	if len(got) != 1 {
		t.Fatalf("expected 1 FRU, got %d", len(got))
	}
	for _, fru := range got {
		if fru.CustomFields == nil {
			t.Fatal("expected CustomFields to be set")
		}
		if fru.CustomFields["rev"] != "A1" {
			t.Errorf("CustomFields[rev] = %v, want %q", fru.CustomFields["rev"], "A1")
		}
	}
}
