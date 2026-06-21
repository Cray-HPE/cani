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
package export

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
)

// TestComparePosition verifies comparePosition emits a FieldDiff only when a
// meaningful local rack position differs from the remote one.
//
// Why it matters: under --merge this decides whether an existing Nautobot device
// gets its U-position rewritten; treating a zero local position as "unset"
// prevents cani from clobbering a position it never tracked.
// Inputs: a CaniDeviceType (RackPosition) and a nautobotapi.Device (Position
// pointer). Outputs: a []FieldDiff, empty when no change is warranted.
// Data choice: cases cover differ, match, local-zero (skip), and remote-nil
// (treated as 0) to pin every branch of the guard.
func TestComparePosition(t *testing.T) {
	pos42 := 42
	pos10 := 10

	tests := []struct {
		name     string
		device   *devicetypes.CaniDeviceType
		remote   *nautobotapi.Device
		wantDiff bool
	}{
		{
			name:     "positions differ returns diff",
			device:   &devicetypes.CaniDeviceType{RackPosition: 42},
			remote:   &nautobotapi.Device{Position: &pos10},
			wantDiff: true,
		},
		{
			name:     "positions match returns no diff",
			device:   &devicetypes.CaniDeviceType{RackPosition: 42},
			remote:   &nautobotapi.Device{Position: &pos42},
			wantDiff: false,
		},
		{
			name:     "local position zero skips comparison",
			device:   &devicetypes.CaniDeviceType{RackPosition: 0},
			remote:   &nautobotapi.Device{Position: &pos10},
			wantDiff: false,
		},
		{
			name:     "remote position nil returns diff",
			device:   &devicetypes.CaniDeviceType{RackPosition: 5},
			remote:   &nautobotapi.Device{Position: nil},
			wantDiff: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diffs := comparePosition(tt.device, tt.remote)
			if tt.wantDiff && len(diffs) == 0 {
				t.Error("expected diff but got none")
			}
			if !tt.wantDiff && len(diffs) > 0 {
				t.Errorf("expected no diff but got %v", diffs)
			}
		})
	}
}

// TestCompareFace verifies compareFace emits a FieldDiff only when a non-empty
// local rack face differs from the remote face value.
//
// Why it matters: rack face (front/rear) drives device placement in Nautobot;
// skipping an empty local face avoids overwriting remote data cani has no
// opinion about, while real differences must be flagged for a merge update.
// Inputs: a CaniDeviceType (Face string) and a nautobotapi.Device whose Face
// holds an optional value pointer. Outputs: a []FieldDiff.
// Data choice: front-vs-rear mismatch, matching rear, and empty-local cover the
// diff, no-diff, and skip branches respectively.
func TestCompareFace(t *testing.T) {
	frontVal := nautobotapi.DeviceFaceValue("front")
	rearVal := nautobotapi.DeviceFaceValue("rear")

	tests := []struct {
		name     string
		device   *devicetypes.CaniDeviceType
		remote   *nautobotapi.Device
		wantDiff bool
	}{
		{
			name:   "faces differ returns diff",
			device: &devicetypes.CaniDeviceType{Face: "rear"},
			remote: &nautobotapi.Device{
				Face: &nautobotapi.DeviceFace{Value: &frontVal},
			},
			wantDiff: true,
		},
		{
			name:   "faces match returns no diff",
			device: &devicetypes.CaniDeviceType{Face: "rear"},
			remote: &nautobotapi.Device{
				Face: &nautobotapi.DeviceFace{Value: &rearVal},
			},
			wantDiff: false,
		},
		{
			name:     "empty local face skips comparison",
			device:   &devicetypes.CaniDeviceType{Face: ""},
			remote:   &nautobotapi.Device{},
			wantDiff: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diffs := compareFace(tt.device, tt.remote)
			if tt.wantDiff && len(diffs) == 0 {
				t.Error("expected diff but got none")
			}
			if !tt.wantDiff && len(diffs) > 0 {
				t.Errorf("expected no diff but got %v", diffs)
			}
		})
	}
}

// TestPtrStr verifies ptrStr dereferences a *string, yielding "" for nil.
//
// Why it matters: Nautobot API models expose most fields as pointers, and diff
// rendering reads them constantly; a nil-safe accessor keeps the exporter from
// panicking on absent optional fields such as Name or Display.
// Inputs: a *string (set, then nil). Outputs: the dereferenced value or "".
// Data choice: present and nil are the helper's only two branches, so the table
// exhaustively covers it.
func TestPtrStr(t *testing.T) {
	val := "hello"

	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{
			name:     "non-nil returns value",
			input:    &val,
			expected: "hello",
		},
		{
			name:     "nil returns empty string",
			input:    nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ptrStr(tt.input)
			if got != tt.expected {
				t.Errorf("ptrStr() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestOrNone verifies orNone returns its input unchanged, or "(none)" when empty.
//
// Why it matters: merge diffs are printed for operators, and showing "(none)"
// instead of a blank makes a missing remote value (e.g. an unset rack or face)
// legible in the change report.
// Inputs: a string (non-empty, then empty). Outputs: the same string or the
// "(none)" sentinel.
// Data choice: non-empty and empty are the helper's only two branches, so the
// table is complete.
func TestOrNone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "non-empty returns input",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "empty returns (none)",
			input:    "",
			expected: "(none)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := orNone(tt.input)
			if got != tt.expected {
				t.Errorf("orNone() = %q, want %q", got, tt.expected)
			}
		})
	}
}
