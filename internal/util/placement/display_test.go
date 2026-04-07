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
	"bytes"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestPrintPlanWithHeight(t *testing.T) {
	entries := []PlacementEntry{
		{RackID: uuid.New(), RackName: "x3701", StartU: 44, Face: "front", DeviceIndex: 0},
		{RackID: uuid.New(), RackName: "x3701", StartU: 39, Face: "front", DeviceIndex: 1},
		{RackID: uuid.New(), RackName: "x3702", StartU: 44, Face: "front", DeviceIndex: 2},
	}
	names := []string{"gh-x3701u44", "gh-x3701u39", "gh-x3702u44"}

	var buf bytes.Buffer
	PrintPlanWithHeight(&buf, entries, names, 5)
	output := buf.String()

	// Verify header and all rows present.
	if !strings.Contains(output, "#") || !strings.Contains(output, "Rack") {
		t.Fatal("missing header")
	}
	for _, name := range names {
		if !strings.Contains(output, name) {
			t.Errorf("missing name %q in output", name)
		}
	}
	// EndU = StartU + height - 1 = 44 + 5 - 1 = 48
	if !strings.Contains(output, "48") {
		t.Error("expected EndU=48 for StartU=44 height=5")
	}
}

func TestPrintPlanEmpty(t *testing.T) {
	var buf bytes.Buffer
	PrintPlanWithHeight(&buf, nil, nil, 5)
	// Should just have header lines, no panic.
	if !strings.Contains(buf.String(), "#") {
		t.Fatal("expected header even for empty plan")
	}
}
