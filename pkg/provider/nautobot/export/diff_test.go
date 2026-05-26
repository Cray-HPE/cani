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
