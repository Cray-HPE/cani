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
	"reflect"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func testInventoryWithDevices(names ...string) *devicetypes.Inventory {
	inv := devicetypes.NewInventory()
	for _, name := range names {
		id := uuid.New()
		inv.Devices[id] = &devicetypes.CaniDeviceType{
			ID:   id,
			Name: name,
			Interfaces: []devicetypes.InterfaceSpec{
				{ID: uuid.New(), Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT},
				{ID: uuid.New(), Name: "eth1", Type: devicetypes.InterfacesElemTypeA1000BaseT},
			},
		}
	}
	return inv
}

// TestResolveConnectionMap_SingleConnection verifies a single literal connection
// resolves to one ResolvedConnection with its ports preserved.
//
// Why it matters: resolution turns declarative name-based intent into concrete
// device-id cables, so the simplest one-to-one case must map straight through.
// Inputs: an inventory with node-01 and switch-01 and a one-entry ConnectionMap
// linking their eth0/eth1 ports. Outputs: no errors and one resolved connection
// with APort eth0 and BPort eth1.
// Data choice: two distinct device names and ports make a transposition or
// drop immediately visible.
func TestResolveConnectionMap_SingleConnection(t *testing.T) {
	inv := testInventoryWithDevices("node-01", "switch-01")

	cm := &ConnectionMap{
		Version: "v1",
		Connections: []ConnectionEntry{
			{
				A: Endpoint{Device: "node-01", Port: "eth0"},
				B: Endpoint{Device: "switch-01", Port: "eth1"},
			},
		},
	}

	resolved, errs := ResolveConnectionMap(cm, inv)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(resolved) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(resolved))
	}
	if resolved[0].APort != "eth0" || resolved[0].BPort != "eth1" {
		t.Errorf("unexpected ports: %s -> %s", resolved[0].APort, resolved[0].BPort)
	}
}

// TestResolveConnectionMap_PatternExpansion verifies brace patterns on both
// endpoints expand and zip into aligned per-link connections.
//
// Why it matters: bulk cabling is expressed with brace ranges, so the resolver
// must expand and pair them index-by-index to produce the right concrete links.
// Inputs: node-{01..03} on eth0 against switch-01 on eth{1..3}. Outputs: three
// connections, each with APort eth0 and BPort eth1/eth2/eth3 in order.
// Data choice: a 3-wide device range zipped against a 3-wide port range proves
// positional alignment rather than a coincidental single match.
func TestResolveConnectionMap_PatternExpansion(t *testing.T) {
	inv := testInventoryWithDevices("node-01", "node-02", "node-03", "switch-01")

	cm := &ConnectionMap{
		Version: "v1",
		Connections: []ConnectionEntry{
			{
				A: Endpoint{Device: "node-{01..03}", Port: "eth0"},
				B: Endpoint{Device: "switch-01", Port: "eth{1..3}"},
			},
		},
	}

	resolved, errs := ResolveConnectionMap(cm, inv)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(resolved) != 3 {
		t.Fatalf("expected 3 connections, got %d", len(resolved))
	}

	expectedPorts := []string{"eth1", "eth2", "eth3"}
	for i, conn := range resolved {
		if conn.APort != "eth0" {
			t.Errorf("connection[%d]: expected A port eth0, got %s", i, conn.APort)
		}
		if conn.BPort != expectedPorts[i] {
			t.Errorf("connection[%d]: expected B port %s, got %s", i, expectedPorts[i], conn.BPort)
		}
	}
}

// TestResolveConnectionMap_DeviceNotFound verifies a connection naming an unknown
// B device produces an error and no resolved connection.
//
// Why it matters: cabling to a device absent from inventory is a user error that
// must be reported, not silently skipped or mapped to a nil id.
// Inputs: an inventory with only node-01 and an entry whose B device is
// "nonexistent". Outputs: a non-empty error slice and zero resolved connections.
// Data choice: leaving the B endpoint unresolved while A exists isolates the
// missing-device path on the B lookup.
func TestResolveConnectionMap_DeviceNotFound(t *testing.T) {
	inv := testInventoryWithDevices("node-01")

	cm := &ConnectionMap{
		Version: "v1",
		Connections: []ConnectionEntry{
			{
				A: Endpoint{Device: "node-01", Port: "eth0"},
				B: Endpoint{Device: "nonexistent", Port: "eth0"},
			},
		},
	}

	resolved, errs := ResolveConnectionMap(cm, inv)
	if len(errs) == 0 {
		t.Fatal("expected errors for missing device")
	}
	if len(resolved) != 0 {
		t.Errorf("expected 0 resolved connections, got %d", len(resolved))
	}
}

// TestResolveConnectionMap_MacCarried verifies per-endpoint MAC addresses survive
// resolution onto the ResolvedConnection.
//
// Why it matters: a MAC pins a cable end to a specific physical interface, so it
// must flow through resolution to be applied at cable-creation time.
// Inputs: a single connection with distinct A and B MAC strings. Outputs: one
// resolved connection whose AMac and BMac equal the inputs.
// Data choice: two different MACs ensure each is carried to its own side without
// being swapped or shared.
func TestResolveConnectionMap_MacCarried(t *testing.T) {
	inv := testInventoryWithDevices("node-01", "switch-01")

	cm := &ConnectionMap{
		Version: "v1",
		Connections: []ConnectionEntry{
			{
				A: Endpoint{Device: "node-01", Port: "eth0", Mac: "aa:bb:cc:dd:ee:01"},
				B: Endpoint{Device: "switch-01", Port: "eth1", Mac: "aa:bb:cc:dd:ee:02"},
			},
		},
	}

	resolved, errs := ResolveConnectionMap(cm, inv)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(resolved) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(resolved))
	}
	if resolved[0].AMac != "aa:bb:cc:dd:ee:01" {
		t.Errorf("AMac = %q, want %q", resolved[0].AMac, "aa:bb:cc:dd:ee:01")
	}
	if resolved[0].BMac != "aa:bb:cc:dd:ee:02" {
		t.Errorf("BMac = %q, want %q", resolved[0].BMac, "aa:bb:cc:dd:ee:02")
	}
}

// TestResolveConnectionMap_MacBraceExpandError verifies a MAC on a brace-expanded
// A endpoint is rejected.
//
// Why it matters: a MAC identifies one physical interface, so it cannot be
// applied across an endpoint that expands to many cables; the resolver must
// reject that ambiguity.
// Inputs: an A endpoint with a brace range and a MAC, against a matching B range.
// Outputs: a non-empty error slice and zero resolved connections.
// Data choice: putting the MAC on the expanding A side drives the A-side
// mac-with-expansion guard specifically.
func TestResolveConnectionMap_MacBraceExpandError(t *testing.T) {
	inv := testInventoryWithDevices("node-01", "node-02", "switch-01")

	cm := &ConnectionMap{
		Version: "v1",
		Connections: []ConnectionEntry{
			{
				A: Endpoint{Device: "node-{01..02}", Port: "eth0", Mac: "aa:bb:cc:dd:ee:01"},
				B: Endpoint{Device: "switch-01", Port: "eth{0..1}"},
			},
		},
	}

	resolved, errs := ResolveConnectionMap(cm, inv)
	if len(errs) == 0 {
		t.Fatal("expected error when mac is applied to a brace-expanded endpoint")
	}
	if len(resolved) != 0 {
		t.Errorf("expected 0 resolved connections, got %d", len(resolved))
	}
}

// TestResolveConnectionMap_CableDefaults verifies map-level cable defaults are
// applied to a connection that sets no per-entry cable.
//
// Why it matters: defaults let one block describe shared cable properties, so a
// bare connection must inherit them at resolve time.
// Inputs: a ConnectionMap with CableDefaults (type cat6a, color blue) and one
// connection with no Cable. Outputs: a resolved cable whose type is cat6a and
// color is blue.
// Data choice: distinct type and color defaults confirm multiple default fields
// are inherited, not just one.
func TestResolveConnectionMap_CableDefaults(t *testing.T) {
	inv := testInventoryWithDevices("node-01", "switch-01")

	cm := &ConnectionMap{
		Version:       "v1",
		CableDefaults: &CableDefaults{Type: "cat6a", Status: "connected", Color: "blue"},
		Connections: []ConnectionEntry{
			{
				A: Endpoint{Device: "node-01", Port: "eth0"},
				B: Endpoint{Device: "switch-01", Port: "eth0"},
			},
		},
	}

	resolved, errs := ResolveConnectionMap(cm, inv)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if resolved[0].Cable.Type != "cat6a" {
		t.Errorf("expected cable type cat6a, got %s", resolved[0].Cable.Type)
	}
	if resolved[0].Cable.Color != "blue" {
		t.Errorf("expected cable color blue, got %s", resolved[0].Cable.Color)
	}
}

// TestResolveConnectionMap_PerEntryOverride verifies a per-entry cable overrides
// matching defaults while inheriting the rest.
//
// Why it matters: a connection must be able to specialize shared defaults (e.g. a
// different cable type) without restating every property.
// Inputs: defaults (type cat6a, color blue) plus an entry overriding type to
// dac-passive and adding length. Outputs: a resolved cable with type
// dac-passive, inherited color blue, and length 3.0.
// Data choice: overriding type but not color, plus adding a brand-new length,
// exercises override, inheritance, and addition in a single case.
func TestResolveConnectionMap_PerEntryOverride(t *testing.T) {
	inv := testInventoryWithDevices("node-01", "switch-01")

	length := 3.0
	cm := &ConnectionMap{
		Version:       "v1",
		CableDefaults: &CableDefaults{Type: "cat6a", Color: "blue"},
		Connections: []ConnectionEntry{
			{
				A:     Endpoint{Device: "node-01", Port: "eth0"},
				B:     Endpoint{Device: "switch-01", Port: "eth0"},
				Cable: &CableProps{Type: "dac-passive", Length: &length},
			},
		},
	}

	resolved, errs := ResolveConnectionMap(cm, inv)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	c := resolved[0].Cable
	if c.Type != "dac-passive" {
		t.Errorf("expected type override dac-passive, got %s", c.Type)
	}
	if c.Color != "blue" {
		t.Errorf("expected inherited color blue, got %s", c.Color)
	}
	if c.Length == nil || *c.Length != 3.0 {
		t.Errorf("expected length 3.0, got %v", c.Length)
	}
}

// TestResolveConnectionMap_PatternLengthMismatch verifies mismatched brace ranges
// on the two endpoints produce an error.
//
// Why it matters: zip semantics require expandable endpoints to agree in count,
// so a 2-vs-3 mismatch is an unresolvable user error that must be reported.
// Inputs: A device range of width 2 against a B port range of width 3. Outputs: a
// non-empty error slice.
// Data choice: widths 2 and 3 are the smallest pair that cannot be broadcast to a
// common count, forcing the mismatch error.
func TestResolveConnectionMap_PatternLengthMismatch(t *testing.T) {
	inv := testInventoryWithDevices("node-01", "node-02", "switch-01")

	cm := &ConnectionMap{
		Version: "v1",
		Connections: []ConnectionEntry{
			{
				A: Endpoint{Device: "node-{01..02}", Port: "eth0"},
				B: Endpoint{Device: "switch-01", Port: "eth{1..3}"},
			},
		},
	}

	_, errs := ResolveConnectionMap(cm, inv)
	if len(errs) == 0 {
		t.Fatal("expected error for pattern length mismatch")
	}
}

// TestZipCount verifies ZipCount returns the aligned count for compatible lengths
// and errors on a conflict.
//
// Why it matters: ZipCount is the core of zip semantics, deciding how many cables
// an entry yields, so it must broadcast 0/1 lengths and reject disagreeing
// multi-element lengths.
// Inputs: a table of length tuples (all ones, one expanded, two matching,
// mismatch, all zero). Outputs: the expected aligned count and error flag per
// case.
// Data choice: the cases span every rule — broadcast, agreement, conflict, and
// the all-zero edge that still yields 1.
func TestZipCount(t *testing.T) {
	tests := []struct {
		name    string
		lengths []int
		want    int
		wantErr bool
	}{
		{"all ones", []int{1, 1, 1, 1}, 1, false},
		{"one expanded", []int{4, 1, 1, 1}, 4, false},
		{"two matching", []int{4, 1, 4, 1}, 4, false},
		{"mismatch", []int{3, 1, 4, 1}, 0, true},
		{"all zero", []int{0, 0, 0, 0}, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ZipCount(tt.lengths...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ZipCount() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ZipCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestBroadcast verifies Broadcast replicates a single element, passes through a
// full-length slice, and zero-fills from nil.
//
// Why it matters: Broadcast aligns endpoint lists to the zip count, so a
// single value must fan out, an already-aligned list must be untouched, and an
// empty input must produce safe empty strings.
// Inputs: a one-element slice to length 3, a three-element slice to length 3, and
// nil to length 2. Outputs: a replicated slice, the same slice, and a slice of
// empty strings respectively.
// Data choice: these three shapes cover the replicate, passthrough, and nil
// branches that the resolver relies on.
func TestBroadcast(t *testing.T) {
	result := Broadcast([]string{"x"}, 3)
	if len(result) != 3 || result[0] != "x" || result[2] != "x" {
		t.Errorf("Broadcast single: got %v", result)
	}

	result = Broadcast([]string{"a", "b", "c"}, 3)
	if len(result) != 3 || result[0] != "a" {
		t.Errorf("Broadcast passthrough: got %v", result)
	}

	result = Broadcast(nil, 2)
	if len(result) != 2 || result[0] != "" {
		t.Errorf("Broadcast nil: got %v", result)
	}
}

// ========== additional branch-coverage tests ==========

// TestResolveConnectionMap_MacBraceExpandErrorBSide verifies a MAC on a
// brace-expanded B endpoint is rejected, mirroring the A-side guard.
//
// Why it matters: the ambiguity of one MAC across many expanded cables applies to
// both endpoints, so the B side must be guarded just like the A side.
// Inputs: an A port range and a B device range (count > 1) with a MAC on the B
// endpoint. Outputs: a non-empty error slice.
// Data choice: putting the MAC on the expanding B side, with A's MAC empty,
// reaches the B-side guard that the existing A-side test cannot.
func TestResolveConnectionMap_MacBraceExpandErrorBSide(t *testing.T) {
	inv := testInventoryWithDevices("node-01", "switch-01", "switch-02")

	cm := &ConnectionMap{
		Version: "v1",
		Connections: []ConnectionEntry{
			{
				A: Endpoint{Device: "node-01", Port: "eth{0..1}"},
				B: Endpoint{Device: "switch-{01..02}", Port: "eth0", Mac: "aa:bb:cc:dd:ee:02"},
			},
		},
	}

	_, errs := ResolveConnectionMap(cm, inv)
	if len(errs) == 0 {
		t.Fatal("expected error when B mac is applied to a brace-expanded endpoint")
	}
}

// TestResolveConnectionMap_ADeviceNotFound verifies an unknown A device produces
// an error and no resolved connection.
//
// Why it matters: the A-side device lookup must fail loudly just like the B side,
// so a typo'd source device is caught rather than mapped to a nil id.
// Inputs: an inventory with only switch-01 and an entry whose A device is
// "nonexistent". Outputs: a non-empty error slice and zero resolved connections.
// Data choice: leaving the A endpoint unresolved while B exists isolates the
// missing-device path on the A lookup, the complement of the existing B-side
// test.
func TestResolveConnectionMap_ADeviceNotFound(t *testing.T) {
	inv := testInventoryWithDevices("switch-01")

	cm := &ConnectionMap{
		Version: "v1",
		Connections: []ConnectionEntry{
			{
				A: Endpoint{Device: "nonexistent", Port: "eth0"},
				B: Endpoint{Device: "switch-01", Port: "eth1"},
			},
		},
	}

	resolved, errs := ResolveConnectionMap(cm, inv)
	if len(errs) == 0 {
		t.Fatal("expected error for missing A device")
	}
	if len(resolved) != 0 {
		t.Errorf("expected 0 resolved connections, got %d", len(resolved))
	}
}

// TestExpandPattern verifies expandPattern returns a single-element slice for
// empty or plain input and a full expansion for a brace range.
//
// Why it matters: expandPattern underlies every endpoint expansion, so an empty
// string must yield one empty element (not zero), a plain string must pass
// through, and a brace range must expand.
// Inputs: an empty string, a plain "eth0", and a brace range "eth{0..2}".
// Outputs: [""], ["eth0"], and ["eth0","eth1","eth2"] respectively.
// Data choice: the empty case drives the early-return branch that the
// resolver-level tests never reach, while the plain and brace cases anchor the
// passthrough and expansion behavior.
func TestExpandPattern(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", []string{""}},
		{"no braces", "eth0", []string{"eth0"}},
		{"brace range", "eth{0..2}", []string{"eth0", "eth1", "eth2"}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := expandPattern(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expandPattern(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// TestMergeProps verifies mergeProps lets a per-entry cable override every
// default field and inherits all defaults when the entry is nil.
//
// Why it matters: mergeProps is the precedence engine for cable properties, so
// each entry field must win over its default and a missing entry must fall back
// to defaults across the board.
// Inputs: defaults setting type/status/color/length_unit, then an entry setting
// all six fields, then a nil entry. Outputs: the entry's values in the first
// case and the defaults in the second.
// Data choice: making every entry field differ from its default exposes any field
// that fails to override, while the nil-entry case confirms full inheritance.
func TestMergeProps(t *testing.T) {
	length := 5.0
	defaults := &CableDefaults{Type: "cat6a", Status: "Planned", Color: "blue", LengthUnit: "m"}

	t.Run("entry overrides every default", func(t *testing.T) {
		entry := &CableProps{
			Type:       "dac-passive",
			Label:      "uplink",
			Color:      "red",
			Length:     &length,
			LengthUnit: "ft",
			Status:     "Active",
		}
		got := mergeProps(entry, defaults)
		if got.Type != "dac-passive" || got.Label != "uplink" || got.Color != "red" ||
			got.LengthUnit != "ft" || got.Status != "Active" || got.Length == nil || *got.Length != 5.0 {
			t.Errorf("mergeProps override = %+v", got)
		}
	})

	t.Run("nil entry inherits defaults", func(t *testing.T) {
		got := mergeProps(nil, defaults)
		if got.Type != "cat6a" || got.Status != "Planned" || got.Color != "blue" || got.LengthUnit != "m" {
			t.Errorf("mergeProps inherit = %+v", got)
		}
	})
}
