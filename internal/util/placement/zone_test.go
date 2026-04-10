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
package placement

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// ─── ComputeZoneBounds ───

func TestComputeZoneBounds48U(t *testing.T) {
	b := ComputeZoneBounds(48, 4, 0)
	if b.Top.StartU != 45 || b.Top.EndU != 48 {
		t.Errorf("top: want U45–48, got U%d–%d", b.Top.StartU, b.Top.EndU)
	}
	if b.Middle.StartU != 1 || b.Middle.EndU != 44 {
		t.Errorf("middle: want U1–44, got U%d–%d", b.Middle.StartU, b.Middle.EndU)
	}
	if b.Bottom.Height() != 0 {
		t.Errorf("bottom: want empty, got U%d–%d", b.Bottom.StartU, b.Bottom.EndU)
	}
}

func TestComputeZoneBoundsWithBottom(t *testing.T) {
	b := ComputeZoneBounds(48, 4, 6)
	if b.Top.StartU != 45 || b.Top.EndU != 48 {
		t.Errorf("top: want U45–48, got U%d–%d", b.Top.StartU, b.Top.EndU)
	}
	if b.Bottom.StartU != 1 || b.Bottom.EndU != 6 {
		t.Errorf("bottom: want U1–6, got U%d–%d", b.Bottom.StartU, b.Bottom.EndU)
	}
	if b.Middle.StartU != 7 || b.Middle.EndU != 44 {
		t.Errorf("middle: want U7–44, got U%d–%d", b.Middle.StartU, b.Middle.EndU)
	}
}

func TestComputeZoneBoundsNoZones(t *testing.T) {
	b := ComputeZoneBounds(48, 0, 0)
	if b.Top.Height() != 0 {
		t.Errorf("top should be empty when topHeight=0, got U%d–%d", b.Top.StartU, b.Top.EndU)
	}
	if b.Bottom.Height() != 0 {
		t.Errorf("bottom should be empty when bottomHeight=0, got U%d–%d", b.Bottom.StartU, b.Bottom.EndU)
	}
	if b.Middle.StartU != 1 || b.Middle.EndU != 48 {
		t.Errorf("middle: want U1–48, got U%d–%d", b.Middle.StartU, b.Middle.EndU)
	}
}

// ─── ZoneForHardwareType ───

func TestZoneForSwitch(t *testing.T) {
	for _, hw := range []string{"switch", "mgmt-switch", "hsn-switch"} {
		if z := ZoneForHardwareType(hw); z != ZoneTop {
			t.Errorf("ZoneForHardwareType(%q) = %q, want top", hw, z)
		}
	}
}

func TestZoneForPDU(t *testing.T) {
	for _, hw := range []string{"cabinet-pdu", "cdu"} {
		if z := ZoneForHardwareType(hw); z != ZoneBottom {
			t.Errorf("ZoneForHardwareType(%q) = %q, want bottom", hw, z)
		}
	}
}

func TestZoneForCompute(t *testing.T) {
	for _, hw := range []string{"blade", "node", "chassis", "nodecard", ""} {
		if z := ZoneForHardwareType(hw); z != ZoneMiddle {
			t.Errorf("ZoneForHardwareType(%q) = %q, want middle", hw, z)
		}
	}
}

// ─── ParseZone ───

func TestParseZoneValid(t *testing.T) {
	cases := map[string]Zone{"top": ZoneTop, "MIDDLE": ZoneMiddle, " Bottom ": ZoneBottom}
	for input, want := range cases {
		got, err := ParseZone(input)
		if err != nil || got != want {
			t.Errorf("ParseZone(%q) = (%q, %v), want (%q, nil)", input, got, err, want)
		}
	}
}

func TestParseZoneInvalid(t *testing.T) {
	_, err := ParseZone("side")
	if err == nil {
		t.Fatal("expected error for unknown zone")
	}
}

// ─── FindSlotInZone ───

func TestFindSlotInZoneTopDown(t *testing.T) {
	rack := makeRack("r1", 48)
	zr := URange{StartU: 1, EndU: 44} // middle zone
	slot := FindSlotInZone(rack, 5, "front", false, zr)
	// Top-down within zone: 44 - 5 + 1 = 40
	if slot != 40 {
		t.Errorf("expected startU=40, got %d", slot)
	}
}

func TestFindSlotInZoneRespectsBounds(t *testing.T) {
	rack := makeRack("r1", 48)
	zr := URange{StartU: 45, EndU: 48} // top zone, 4U
	slot := FindSlotInZone(rack, 5, "front", false, zr)
	// 5U device cannot fit in 4U zone
	if slot != 0 {
		t.Errorf("expected 0 (doesn't fit), got %d", slot)
	}
}

func TestFindSlotInZoneOccupied(t *testing.T) {
	rack := makeRack("r1", 48)
	// Occupy the top slot of the middle zone (U40–44)
	rack.PlaceDevice(uuid.New(), 40, 5, "front", false)

	zr := URange{StartU: 1, EndU: 44}
	slot := FindSlotInZone(rack, 5, "front", false, zr)
	// Next available: 35
	if slot != 35 {
		t.Errorf("expected startU=35, got %d", slot)
	}
}

func TestFindSlotInZoneFull(t *testing.T) {
	rack := makeRack("r1", 10)
	rack.PlaceDevice(uuid.New(), 1, 10, "front", false)

	zr := URange{StartU: 1, EndU: 10}
	slot := FindSlotInZone(rack, 1, "front", false, zr)
	if slot != 0 {
		t.Errorf("expected 0 (full), got %d", slot)
	}
}

// ─── ResolveZoneBounds ───

func TestResolveZoneBoundsDefaults(t *testing.T) {
	rack := &devicetypes.CaniRackType{UHeight: 48}
	b := ResolveZoneBounds(rack)
	if b.Top.StartU != 45 || b.Top.EndU != 48 {
		t.Errorf("top: want U45–48, got U%d–%d", b.Top.StartU, b.Top.EndU)
	}
	if b.Middle.StartU != 1 || b.Middle.EndU != 44 {
		t.Errorf("middle: want U1–44, got U%d–%d", b.Middle.StartU, b.Middle.EndU)
	}
}

func TestResolveZoneBoundsCustom(t *testing.T) {
	rack := &devicetypes.CaniRackType{
		UHeight:          48,
		TopZoneHeight:    6,
		BottomZoneHeight: 3,
	}
	b := ResolveZoneBounds(rack)
	if b.Top.StartU != 43 || b.Top.EndU != 48 {
		t.Errorf("top: want U43–48, got U%d–%d", b.Top.StartU, b.Top.EndU)
	}
	if b.Bottom.StartU != 1 || b.Bottom.EndU != 3 {
		t.Errorf("bottom: want U1–3, got U%d–%d", b.Bottom.StartU, b.Bottom.EndU)
	}
	if b.Middle.StartU != 4 || b.Middle.EndU != 42 {
		t.Errorf("middle: want U4–42, got U%d–%d", b.Middle.StartU, b.Middle.EndU)
	}
}

// ─── RangeForZone ───

func TestRangeForZone(t *testing.T) {
	b := ComputeZoneBounds(48, 4, 6)
	if r := b.RangeForZone(ZoneTop); r.StartU != 45 {
		t.Errorf("top: got startU=%d", r.StartU)
	}
	if r := b.RangeForZone(ZoneBottom); r.StartU != 1 {
		t.Errorf("bottom: got startU=%d", r.StartU)
	}
	if r := b.RangeForZone(ZoneMiddle); r.StartU != 7 {
		t.Errorf("middle: got startU=%d", r.StartU)
	}
}
