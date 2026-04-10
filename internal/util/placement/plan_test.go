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

func makeRack(name string, uHeight int) *devicetypes.CaniRackType {
	return &devicetypes.CaniRackType{
		ID:            uuid.New(),
		Name:          name,
		UHeight:       uHeight,
		OccupiedSlots: make(map[int]map[string]uuid.UUID),
	}
}

func TestParseStrategyFill(t *testing.T) {
	s, ok := ParseStrategy("%{FILL}")
	if !ok || s != StrategyFill {
		t.Fatalf("expected FILL, got %q ok=%v", s, ok)
	}
}

func TestParseStrategySpread(t *testing.T) {
	s, ok := ParseStrategy("%{SPREAD}")
	if !ok || s != StrategySpread {
		t.Fatalf("expected SPREAD, got %q ok=%v", s, ok)
	}
}

func TestParseStrategyCaseInsensitive(t *testing.T) {
	s, ok := ParseStrategy("%{fill}")
	if !ok || s != StrategyFill {
		t.Fatalf("expected FILL, got %q ok=%v", s, ok)
	}
}

func TestParseStrategyLiteral(t *testing.T) {
	_, ok := ParseStrategy("rack-01")
	if ok {
		t.Fatal("literal rack name should not parse as strategy")
	}
}

func TestFillAllInOneRack(t *testing.T) {
	racks := []*devicetypes.CaniRackType{
		makeRack("r1", 48),
		makeRack("r2", 48),
	}
	entries, err := Plan(racks, 5, "front", false, 3, StrategyFill, "", "blade")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	for i, e := range entries {
		if e.RackName != "r1" {
			t.Errorf("entry %d: expected rack r1, got %s", i, e.RackName)
		}
	}
}

func TestFillOverflow(t *testing.T) {
	// 14U racks: default top=4U, middle=U1–10 (10U).
	// 3 devices × 5U = 15U. Two fit in r1 (10U), overflow 1 to r2.
	racks := []*devicetypes.CaniRackType{
		makeRack("r1", 14),
		makeRack("r2", 14),
	}
	entries, err := Plan(racks, 5, "front", false, 3, StrategyFill, ZoneMiddle, "blade")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	r1Count, r2Count := 0, 0
	for _, e := range entries {
		if e.RackName == "r1" {
			r1Count++
		} else {
			r2Count++
		}
	}
	if r1Count != 2 || r2Count != 1 {
		t.Fatalf("expected 2 in r1 and 1 in r2, got %d and %d", r1Count, r2Count)
	}
}

func TestFillCapacityExceeded(t *testing.T) {
	// 14U rack: middle = U1–10 (10U). 3 × 5U = 15U won't fit.
	racks := []*devicetypes.CaniRackType{makeRack("r1", 14)}
	_, err := Plan(racks, 5, "front", false, 3, StrategyFill, ZoneMiddle, "blade")
	if err == nil {
		t.Fatal("expected capacity error")
	}
}

func TestSpreadRoundRobin(t *testing.T) {
	racks := []*devicetypes.CaniRackType{
		makeRack("r1", 48),
		makeRack("r2", 48),
	}
	entries, err := Plan(racks, 5, "front", false, 6, StrategySpread, "", "blade")
	if err != nil {
		t.Fatal(err)
	}
	r1Count, r2Count := 0, 0
	for _, e := range entries {
		if e.RackName == "r1" {
			r1Count++
		} else {
			r2Count++
		}
	}
	if r1Count != 3 || r2Count != 3 {
		t.Fatalf("expected 3 per rack, got r1=%d r2=%d", r1Count, r2Count)
	}
}

func TestSpreadAlternation(t *testing.T) {
	racks := []*devicetypes.CaniRackType{
		makeRack("r1", 48),
		makeRack("r2", 48),
	}
	entries, err := Plan(racks, 5, "front", false, 4, StrategySpread, "", "blade")
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"r1", "r2", "r1", "r2"}
	for i, e := range entries {
		if e.RackName != expected[i] {
			t.Errorf("entry %d: expected %s, got %s", i, expected[i], e.RackName)
		}
	}
}

func TestSpreadRackFull(t *testing.T) {
	// 9U racks: middle = U1–5 (5U). Each fits one 5U device.
	// 3rd device round-robins back to r1 which is full → error.
	racks := []*devicetypes.CaniRackType{
		makeRack("r1", 9),
		makeRack("r2", 9),
	}
	_, err := Plan(racks, 5, "front", false, 3, StrategySpread, ZoneMiddle, "blade")
	if err == nil {
		t.Fatal("expected error when rack is full")
	}
}

func TestRacksSortedByName(t *testing.T) {
	racks := []*devicetypes.CaniRackType{
		makeRack("z1", 48),
		makeRack("a1", 48),
		makeRack("m1", 48),
	}
	entries, err := Plan(racks, 5, "front", false, 3, StrategyFill, "", "blade")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.RackName != "a1" {
			t.Errorf("expected a1 (alphabetically first), got %s", e.RackName)
		}
	}
}

func TestSingleRack(t *testing.T) {
	racks := []*devicetypes.CaniRackType{makeRack("solo", 48)}
	entries, err := Plan(racks, 5, "front", false, 2, StrategyFill, "", "blade")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2, got %d", len(entries))
	}
}

func TestNoRacks(t *testing.T) {
	_, err := Plan(nil, 5, "front", false, 1, StrategyFill, "", "blade")
	if err == nil {
		t.Fatal("expected error for no racks")
	}
}

func TestDefaultFace(t *testing.T) {
	racks := []*devicetypes.CaniRackType{makeRack("r1", 48)}
	entries, err := Plan(racks, 5, "", false, 1, StrategyFill, "", "blade")
	if err != nil {
		t.Fatal(err)
	}
	if entries[0].Face != devicetypes.FaceFront {
		t.Errorf("expected front face default, got %s", entries[0].Face)
	}
}

func TestTopDownPlacement(t *testing.T) {
	racks := []*devicetypes.CaniRackType{makeRack("r1", 48)}
	entries, err := Plan(racks, 5, "front", false, 2, StrategyFill, "", "blade")
	if err != nil {
		t.Fatal(err)
	}
	// Middle zone default: U1–44 (top 4U reserved).
	// First device: 44 - 5 + 1 = 40
	if entries[0].StartU != 40 {
		t.Errorf("expected startU=40, got %d", entries[0].StartU)
	}
	if entries[1].StartU != 35 {
		t.Errorf("expected startU=35, got %d", entries[1].StartU)
	}
}

func TestDoesNotMutateOriginal(t *testing.T) {
	rack := makeRack("r1", 48)
	racks := []*devicetypes.CaniRackType{rack}
	_, err := Plan(racks, 5, "front", false, 3, StrategyFill, "", "blade")
	if err != nil {
		t.Fatal(err)
	}
	if len(rack.OccupiedSlots) != 0 {
		t.Fatal("Plan mutated original rack")
	}
}

// ─── Zone-aware placement tests ───

func TestFillMiddleZoneOnly(t *testing.T) {
	// 48U rack, default top=4U. Middle zone = U1–44.
	// 9 devices × 5U = 45U. Middle has 44U → only 8 fit in one rack.
	racks := []*devicetypes.CaniRackType{
		makeRack("r1", 48),
		makeRack("r2", 48),
	}
	entries, err := Plan(racks, 5, "front", false, 9, StrategyFill, "", "blade")
	if err != nil {
		t.Fatal(err)
	}
	// All entries must be in the middle zone.
	for i, e := range entries {
		if e.Zone != "middle" {
			t.Errorf("entry %d: expected zone=middle, got %s", i, e.Zone)
		}
	}
	// 8 devices fit in r1's middle (8×5=40 ≤ 44), 1 overflows to r2.
	r1Count := 0
	for _, e := range entries {
		if e.RackName == "r1" {
			r1Count++
		}
	}
	if r1Count != 8 {
		t.Errorf("expected 8 in r1, got %d", r1Count)
	}
}

func TestFillTopZoneSwitches(t *testing.T) {
	// Switches auto-detect to top zone. Default top = U45–48 (4U).
	// Two 1U switches should fit at U48 and U47.
	racks := []*devicetypes.CaniRackType{makeRack("r1", 48)}
	entries, err := Plan(racks, 1, "front", false, 2, StrategyFill, "", "mgmt-switch")
	if err != nil {
		t.Fatal(err)
	}
	if entries[0].StartU != 48 {
		t.Errorf("expected first switch at U48, got U%d", entries[0].StartU)
	}
	if entries[1].StartU != 47 {
		t.Errorf("expected second switch at U47, got U%d", entries[1].StartU)
	}
	for _, e := range entries {
		if e.Zone != "top" {
			t.Errorf("expected zone=top, got %s", e.Zone)
		}
	}
}

func TestFillTopZoneOverflow(t *testing.T) {
	// 4U top zone can't hold 3 × 2U switches in one rack.
	racks := []*devicetypes.CaniRackType{
		makeRack("r1", 48),
		makeRack("r2", 48),
	}
	entries, err := Plan(racks, 2, "front", false, 3, StrategyFill, "", "switch")
	if err != nil {
		t.Fatal(err)
	}
	// r1 gets 2 (4U), r2 gets 1.
	r1Count := 0
	for _, e := range entries {
		if e.RackName == "r1" {
			r1Count++
		}
	}
	if r1Count != 2 {
		t.Errorf("expected 2 in r1 top zone, got %d", r1Count)
	}
}

func TestZoneOverrideForceBottom(t *testing.T) {
	// Force a "blade" device into the bottom zone via explicit zone.
	rack := makeRack("r1", 48)
	rack.BottomZoneHeight = 6
	racks := []*devicetypes.CaniRackType{rack}
	entries, err := Plan(racks, 2, "front", false, 1, StrategyFill, ZoneBottom, "blade")
	if err != nil {
		t.Fatal(err)
	}
	if entries[0].Zone != "bottom" {
		t.Errorf("expected zone=bottom, got %s", entries[0].Zone)
	}
	// Bottom zone U1–6, top-down → startU = 5
	if entries[0].StartU != 5 {
		t.Errorf("expected startU=5, got %d", entries[0].StartU)
	}
}

func TestZoneEntryHasZoneField(t *testing.T) {
	racks := []*devicetypes.CaniRackType{makeRack("r1", 48)}
	entries, err := Plan(racks, 1, "front", false, 1, StrategyFill, ZoneTop, "blade")
	if err != nil {
		t.Fatal(err)
	}
	if entries[0].Zone != "top" {
		t.Errorf("PlacementEntry.Zone = %q, want \"top\"", entries[0].Zone)
	}
}
