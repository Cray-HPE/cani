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
package connections

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// TestInventoryToConnectionMap_SortsByCableUUID verifies the exporter emits one
// entry per cable in ascending cable-UUID order with device names resolved.
//
// Why it matters: the connection map is the round-trippable export of inventory
// cabling, so its output must be deterministic (UUID-sorted) and human-readable
// (names, not UUIDs) regardless of Go's randomized map iteration.
// Inputs: an inventory with two named devices and two cables whose IDs are fixed
// so ...0001 sorts before ...0002. Outputs: a ConnectionMap with version v1 and
// two entries ordered by cable UUID, each endpoint showing the device name.
// Data choice: hand-assigned ascending UUIDs make the sort observable, and the
// lower-UUID cable is deliberately added with the later port so a stable sort is
// required (not insertion or map order) for the assertion to hold.
func TestInventoryToConnectionMap_SortsByCableUUID(t *testing.T) {
	inv := devicetypes.NewInventory()
	devA := uuid.New()
	devB := uuid.New()
	inv.Devices[devA] = &devicetypes.CaniDeviceType{ID: devA, Name: "alpha"}
	inv.Devices[devB] = &devicetypes.CaniDeviceType{ID: devB, Name: "beta"}

	cableLo := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	cableHi := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	inv.Cables[cableHi] = &devicetypes.CaniCableType{
		ID:                 cableHi,
		TerminationADevice: devB,
		TerminationAPort:   "p3",
		TerminationBDevice: devA,
		TerminationBPort:   "p4",
	}
	inv.Cables[cableLo] = &devicetypes.CaniCableType{
		ID:                 cableLo,
		TerminationADevice: devA,
		TerminationAPort:   "p1",
		TerminationBDevice: devB,
		TerminationBPort:   "p2",
	}

	cm := InventoryToConnectionMap(inv)
	if cm.Version != "v1" {
		t.Errorf("version = %q, want v1", cm.Version)
	}
	if len(cm.Connections) != 2 {
		t.Fatalf("connections = %d, want 2", len(cm.Connections))
	}
	if cm.Connections[0].A.Port != "p1" {
		t.Errorf("entry[0] A.Port = %q, want p1 (lower-UUID cable first)", cm.Connections[0].A.Port)
	}
	if cm.Connections[0].A.Device != "alpha" {
		t.Errorf("entry[0] A.Device = %q, want alpha", cm.Connections[0].A.Device)
	}
	if cm.Connections[1].A.Port != "p3" {
		t.Errorf("entry[1] A.Port = %q, want p3 (higher-UUID cable second)", cm.Connections[1].A.Port)
	}
}

// TestInventoryToConnectionMap_SkipsNilCable verifies a nil entry in the cable
// map is skipped rather than emitted or panicked on.
//
// Why it matters: a sparse or partially populated cable map must not crash the
// exporter or produce a phantom connection.
// Inputs: an inventory whose cable map holds one real cable and one nil value.
// Outputs: a ConnectionMap with exactly one entry.
// Data choice: pairing a valid cable with an explicit nil isolates the nil-skip
// branch while proving the surrounding cable is still exported.
func TestInventoryToConnectionMap_SkipsNilCable(t *testing.T) {
	inv := devicetypes.NewInventory()
	dev := uuid.New()
	inv.Devices[dev] = &devicetypes.CaniDeviceType{ID: dev, Name: "sw"}

	good := uuid.New()
	inv.Cables[good] = &devicetypes.CaniCableType{
		ID:                 good,
		TerminationADevice: dev,
		TerminationAPort:   "1",
		TerminationBDevice: dev,
		TerminationBPort:   "2",
	}
	inv.Cables[uuid.New()] = nil

	cm := InventoryToConnectionMap(inv)
	if len(cm.Connections) != 1 {
		t.Fatalf("connections = %d, want 1 (nil cable skipped)", len(cm.Connections))
	}
}

// TestInventoryToConnectionMap_EmitsPropsConditionally verifies cable properties
// are attached only when at least one of slug, label, color, or length is set.
//
// Why it matters: the export omits an empty cable block so round-tripped YAML/CSV
// stays minimal, while a cable carrying metadata must preserve it.
// Inputs: per case, an inventory with a single cable that either sets a slug or
// leaves all property fields empty. Outputs: an entry whose Cable is populated
// (Type from slug) or nil respectively.
// Data choice: the two cases sit on opposite sides of the property guard — one
// field set versus all empty — so both the emit and omit branches are exercised.
func TestInventoryToConnectionMap_EmitsPropsConditionally(t *testing.T) {
	cases := []struct {
		name     string
		mutate   func(c *devicetypes.CaniCableType)
		wantCabl bool
	}{
		{"slug set emits props", func(c *devicetypes.CaniCableType) { c.Slug = "cat6a" }, true},
		{"no props omits cable", func(c *devicetypes.CaniCableType) {}, false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			inv := devicetypes.NewInventory()
			dev := uuid.New()
			inv.Devices[dev] = &devicetypes.CaniDeviceType{ID: dev, Name: "sw"}
			id := uuid.New()
			cable := &devicetypes.CaniCableType{
				ID:                 id,
				TerminationADevice: dev,
				TerminationAPort:   "1",
				TerminationBDevice: dev,
				TerminationBPort:   "2",
			}
			tt.mutate(cable)
			inv.Cables[id] = cable

			cm := InventoryToConnectionMap(inv)
			if len(cm.Connections) != 1 {
				t.Fatalf("connections = %d, want 1", len(cm.Connections))
			}
			gotCable := cm.Connections[0].Cable
			if tt.wantCabl && gotCable == nil {
				t.Fatal("expected cable props, got nil")
			}
			if !tt.wantCabl && gotCable != nil {
				t.Fatalf("expected nil cable props, got %+v", gotCable)
			}
			if tt.wantCabl && gotCable.Type != "cat6a" {
				t.Errorf("cable type = %q, want cat6a", gotCable.Type)
			}
		})
	}
}

// TestResolveDeviceName verifies resolveDeviceName maps a device UUID to its name
// and falls back to the UUID string for missing or unnamed devices.
//
// Why it matters: export uses readable device names, but it must degrade
// gracefully to a UUID string when a name is unavailable so no endpoint is ever
// blank-by-accident.
// Inputs: an inventory with one named and one unnamed device, queried with the
// nil UUID, the named ID, an absent ID, and the unnamed ID. Outputs: "", the
// name, the absent UUID string, and the unnamed UUID string respectively.
// Data choice: the four cases map one-to-one onto resolveDeviceName's branches —
// nil guard, name hit, not-in-inventory fallback, and present-but-empty-name
// fallback.
func TestResolveDeviceName(t *testing.T) {
	inv := devicetypes.NewInventory()
	named := uuid.New()
	inv.Devices[named] = &devicetypes.CaniDeviceType{ID: named, Name: "switch-a"}
	noName := uuid.New()
	inv.Devices[noName] = &devicetypes.CaniDeviceType{ID: noName, Name: ""}
	missing := uuid.New()

	cases := []struct {
		name string
		id   uuid.UUID
		want string
	}{
		{"nil uuid", uuid.Nil, ""},
		{"named device", named, "switch-a"},
		{"missing device falls back to uuid", missing, missing.String()},
		{"empty name falls back to uuid", noName, noName.String()},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveDeviceName(tt.id, inv); got != tt.want {
				t.Errorf("resolveDeviceName(%v) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}
