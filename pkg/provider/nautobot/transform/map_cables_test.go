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

// TestBuildInterfaceMap verifies BuildInterfaceMap indexes interfaces by UUID to
// their (device UUID, name), skipping interfaces with a nil Id and recording a
// nil device UUID when the interface has no device.
//
// Why it matters: cable terminations reference interfaces by UUID; MapCables uses
// this map to resolve each end to a device and port, so a missing interface Id
// must be skipped and a deviceless interface must still index by name.
// Inputs: nil, a nil-Id interface, an interface with a device ref, and an
// interface with no device. Outputs: the UUID->ifaceRef map.
// Data choice: distinct interface and device UUIDs prove the device link is
// captured, while the deviceless case proves the nil-device guard yields
// uuid.Nil rather than panicking.
func TestBuildInterfaceMap(t *testing.T) {
	t.Run("empty input returns empty map", func(t *testing.T) {
		got := BuildInterfaceMap(nil)
		if len(got) != 0 {
			t.Errorf("expected 0, got %d", len(got))
		}
	})

	t.Run("interface with nil ID is skipped", func(t *testing.T) {
		ifaces := []nautobotapi.Interface{
			{Id: nil, Name: "eth0"},
		}
		got := BuildInterfaceMap(ifaces)
		if len(got) != 0 {
			t.Errorf("expected 0, got %d", len(got))
		}
	})

	t.Run("interfaces are indexed with device ref", func(t *testing.T) {
		ifaceID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
		devID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		oaIfaceID := openapi_types.UUID(ifaceID)
		devRef := makeTenantRefFromUUID(devID)

		ifaces := []nautobotapi.Interface{
			{Id: &oaIfaceID, Name: "eth0", Device: &devRef},
		}

		got := BuildInterfaceMap(ifaces)
		if len(got) != 1 {
			t.Fatalf("expected 1, got %d", len(got))
		}
		ref := got[ifaceID]
		if ref.deviceID != devID {
			t.Errorf("deviceID = %s, want %s", ref.deviceID, devID)
		}
		if ref.name != "eth0" {
			t.Errorf("name = %q, want %q", ref.name, "eth0")
		}
	})

	t.Run("interface without device has nil device ID", func(t *testing.T) {
		ifaceID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
		oaIfaceID := openapi_types.UUID(ifaceID)

		ifaces := []nautobotapi.Interface{
			{Id: &oaIfaceID, Name: "orphan-eth", Device: nil},
		}

		got := BuildInterfaceMap(ifaces)
		ref := got[ifaceID]
		if ref.deviceID != uuid.Nil {
			t.Errorf("deviceID = %s, want Nil", ref.deviceID)
		}
	})
}

// TestMapCables verifies MapCables converts cables and resolves status plus both
// terminations to CANI interface IDs and device/port pairs, skipping nil-Id
// cables and leaving terminations empty when an interface reference is unknown.
//
// Why it matters: cables are imported after devices and interfaces; each
// termination must preserve the imported interface UUID and resolve to the CANI
// device plus port name, so unresolved references must degrade to empty rather
// than wire a cable to the wrong device.
// Inputs: device and interface lookup maps plus cables that are nil-Id, fully
// resolvable, and reference unknown interfaces. Outputs: the CANI cable map with
// label, color, type, length/unit, and both termination device/port fields.
// Data choice: two distinct devices and named interfaces (eth0/eth1) prove each
// termination resolves to its own end, and unknown UUIDs prove the miss path
// leaves both termination devices at uuid.Nil.
func TestMapCables(t *testing.T) {
	devANBID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	devACaniID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	devBNBID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	devBCaniID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	deviceMap := map[uuid.UUID]uuid.UUID{
		devANBID: devACaniID,
		devBNBID: devBCaniID,
	}

	ifaceAID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	ifaceBID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	ifaceMap := map[uuid.UUID]ifaceRef{
		ifaceAID: {deviceID: devANBID, name: "eth0"},
		ifaceBID: {deviceID: devBNBID, name: "eth1"},
	}

	t.Run("empty input returns empty map", func(t *testing.T) {
		got := MapCables(nil, deviceMap, ifaceMap, nil)
		if len(got) != 0 {
			t.Errorf("expected 0, got %d", len(got))
		}
	})

	t.Run("cable with nil ID is skipped", func(t *testing.T) {
		raw := []nautobotapi.Cable{
			{Id: nil, Status: makeStatusRefFromUUID(uuid.New()),
				TerminationAId: openapi_types.UUID(ifaceAID), TerminationBId: openapi_types.UUID(ifaceBID)},
		}
		got := MapCables(raw, deviceMap, ifaceMap, nil)
		if len(got) != 0 {
			t.Errorf("expected 0, got %d", len(got))
		}
	})

	t.Run("cable with terminations resolved", func(t *testing.T) {
		cableID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
		oaCableID := openapi_types.UUID(cableID)
		label := "uplink-1"
		color := "blue"
		length := float64(3)
		statusID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
		cableType := nautobotapi.CableTypeValue("cat6")
		lengthUnit := nautobotapi.CableLengthUnitValue("m")
		termAType := "dcim.interface"
		termBType := "dcim.interface"

		raw := []nautobotapi.Cable{
			{
				Id:               &oaCableID,
				Label:            &label,
				Color:            &color,
				Status:           makeStatusRefFromUUID(statusID),
				TerminationAId:   openapi_types.UUID(ifaceAID),
				TerminationBId:   openapi_types.UUID(ifaceBID),
				TerminationAType: termAType,
				TerminationBType: termBType,
				Type:             &nautobotapi.CableType{Value: &cableType},
				Length:           intPtr(3),
				LengthUnit:       &nautobotapi.CableLengthUnit{Value: &lengthUnit},
			},
		}

		got := MapCables(raw, deviceMap, ifaceMap, map[uuid.UUID]string{statusID: "Connected"})
		if len(got) != 1 {
			t.Fatalf("expected 1, got %d", len(got))
		}

		for _, cable := range got {
			if cable.Label != "uplink-1" {
				t.Errorf("Label = %q, want %q", cable.Label, "uplink-1")
			}
			if cable.Color != "blue" {
				t.Errorf("Color = %q, want %q", cable.Color, "blue")
			}
			if cable.CableType != "cat6" {
				t.Errorf("CableType = %q, want %q", cable.CableType, "cat6")
			}
			if cable.Length == nil || *cable.Length != length {
				t.Errorf("Length = %v, want %v", cable.Length, length)
			}
			if cable.LengthUnit != "m" {
				t.Errorf("LengthUnit = %q, want %q", cable.LengthUnit, "m")
			}
			if cable.Status != "Connected" {
				t.Errorf("Status = %q, want %q", cable.Status, "Connected")
			}
			if cable.TerminationA != ifaceAID {
				t.Errorf("TerminationA = %s, want %s", cable.TerminationA, ifaceAID)
			}
			if cable.TerminationADevice != devACaniID {
				t.Errorf("TerminationADevice = %s, want %s", cable.TerminationADevice, devACaniID)
			}
			if cable.TerminationAPort != "eth0" {
				t.Errorf("TerminationAPort = %q, want %q", cable.TerminationAPort, "eth0")
			}
			if cable.TerminationB != ifaceBID {
				t.Errorf("TerminationB = %s, want %s", cable.TerminationB, ifaceBID)
			}
			if cable.TerminationBDevice != devBCaniID {
				t.Errorf("TerminationBDevice = %s, want %s", cable.TerminationBDevice, devBCaniID)
			}
			if cable.TerminationBPort != "eth1" {
				t.Errorf("TerminationBPort = %q, want %q", cable.TerminationBPort, "eth1")
			}
			if cable.TerminationAType != "dcim.interface" {
				t.Errorf("TerminationAType = %q, want %q", cable.TerminationAType, "dcim.interface")
			}
			if cable.TerminationBType != "dcim.interface" {
				t.Errorf("TerminationBType = %q, want %q", cable.TerminationBType, "dcim.interface")
			}
		}
	})

	t.Run("unknown interface refs leave terminations empty", func(t *testing.T) {
		cableID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
		oaCableID := openapi_types.UUID(cableID)
		unknownA := uuid.MustParse("99999999-9999-9999-9999-999999999999")
		unknownB := uuid.MustParse("88888888-8888-8888-8888-888888888888")

		raw := []nautobotapi.Cable{
			{
				Id:             &oaCableID,
				Status:         makeStatusRefFromUUID(uuid.New()),
				TerminationAId: openapi_types.UUID(unknownA),
				TerminationBId: openapi_types.UUID(unknownB),
			},
		}

		got := MapCables(raw, deviceMap, ifaceMap, nil)
		for _, cable := range got {
			if cable.TerminationA != uuid.Nil {
				t.Errorf("TerminationA = %s, want Nil", cable.TerminationA)
			}
			if cable.TerminationADevice != uuid.Nil {
				t.Errorf("TerminationADevice = %s, want Nil", cable.TerminationADevice)
			}
			if cable.TerminationB != uuid.Nil {
				t.Errorf("TerminationB = %s, want Nil", cable.TerminationB)
			}
			if cable.TerminationBDevice != uuid.Nil {
				t.Errorf("TerminationBDevice = %s, want Nil", cable.TerminationBDevice)
			}
		}
	})

	t.Run("custom fields are passed through", func(t *testing.T) {
		cableID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
		oaCableID := openapi_types.UUID(cableID)
		cf := map[string]interface{}{"vendor": "Panduit"}

		raw := []nautobotapi.Cable{
			{
				Id:             &oaCableID,
				Status:         makeStatusRefFromUUID(uuid.New()),
				CustomFields:   &cf,
				TerminationAId: openapi_types.UUID(ifaceAID),
				TerminationBId: openapi_types.UUID(ifaceBID),
			},
		}

		got := MapCables(raw, deviceMap, ifaceMap, nil)
		for _, cable := range got {
			if cable.CustomFields == nil {
				t.Fatal("expected CustomFields to be set")
			}
			if cable.CustomFields["vendor"] != "Panduit" {
				t.Errorf("CustomFields[vendor] = %v, want %q", cable.CustomFields["vendor"], "Panduit")
			}
		}
	})
}
