/*
 *
 *  MIT License
 *
 *  (C) Copyright 2024-2026 Hewlett Packard Enterprise Development LP
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
package visual

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

func captureStdout(t *testing.T, run func()) string {
	t.Helper()

	oldStdout := os.Stdout
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe returned error: %v", err)
	}

	os.Stdout = writePipe
	run()
	if err := writePipe.Close(); err != nil {
		os.Stdout = oldStdout
		t.Fatalf("closing stdout pipe returned error: %v", err)
	}
	os.Stdout = oldStdout

	var output bytes.Buffer
	if _, err := io.Copy(&output, readPipe); err != nil {
		t.Fatalf("reading captured stdout returned error: %v", err)
	}
	if err := readPipe.Close(); err != nil {
		t.Fatalf("closing stdout reader returned error: %v", err)
	}

	return output.String()
}

func tableRow(values []string, widths []int) string {
	parts := make([]string, 0, len(values))
	for i, value := range values {
		parts = append(parts, rightPadVisible(value, widths[i]))
	}
	return strings.Join(parts, "  ")
}

// TestPrintDeviceTableWithRoles verifies device table rendering emits exact
// fixed-width columns, resolved rack names, U positions, and role data.
//
// Why it matters: table output is a user-facing format used for quick inventory
// inspection, so column drift or unresolved foreign keys makes the output harder
// to scan.
// Inputs: two devices from the sample rack inventory rendered with role display
// enabled. Outputs: the complete table text written to stdout.
// Data choice: a node and a switch give distinct type, model, role, and U-position
// values while sharing one resolved rack name.
func TestPrintDeviceTableWithRoles(t *testing.T) {
	inventory := sampleRackInventory()
	devices := []*devicetypes.CaniDeviceType{
		inventory.Devices[sampleNodeAID],
		inventory.Devices[sampleSwitchID],
	}

	output := captureStdout(t, func() {
		PrintDeviceTable(devices, inventory, TreeFilter{Roles: true})
	})

	widths := []int{30, 15, 30, 10, 20, 6, 14}
	want := strings.Join([]string{
		tableRow([]string{"NAME", "TYPE", "MODEL", "STATUS", "RACK", "U-POS", "ROLE"}, widths),
		tableRow([]string{strings.Repeat("-", 30), strings.Repeat("-", 15), strings.Repeat("-", 30), strings.Repeat("-", 10), strings.Repeat("-", 20), strings.Repeat("-", 6), strings.Repeat("-", 14)}, widths),
		tableRow([]string{"Node-A", "node", "HPE DL360", "Active", "Rack-006U", "1", "compute"}, widths),
		tableRow([]string{"Leaf-1", "switch", "Aruba 8325", "Active", "Rack-006U", "6", "leaf"}, widths),
		"",
		"Total: 2 device(s)",
		"",
	}, "\n")
	assertExactOutput(t, output, want)
}

// TestPrintCableTableResolvesTerminations verifies cable table rendering resolves
// endpoint device names and ports into fixed-width termination columns.
//
// Why it matters: cable tables are the plain-text view of physical connectivity,
// and users need the visible endpoints to remain stable and readable.
// Inputs: the sample inventory cable rendered through PrintCableTable. Outputs:
// the complete cable table text written to stdout.
// Data choice: one node-to-switch cable proves both termination columns combine
// inventory device names with the stored port names.
func TestPrintCableTableResolvesTerminations(t *testing.T) {
	inventory := sampleRackInventory()
	cables := []*devicetypes.CaniCableType{inventory.Cables[sampleCableID]}

	output := captureStdout(t, func() {
		PrintCableTable(cables, inventory)
	})

	widths := []int{25, 15, 10, 25, 25}
	want := strings.Join([]string{
		tableRow([]string{"LABEL", "TYPE", "STATUS", "A TERMINATION", "B TERMINATION"}, widths),
		tableRow([]string{strings.Repeat("-", 25), strings.Repeat("-", 15), strings.Repeat("-", 10), strings.Repeat("-", 25), strings.Repeat("-", 25)}, widths),
		tableRow([]string{"node-a-to-leaf", "cat6", "Connected", "Node-A:eth0", "Leaf-1:1/1/1"}, widths),
		"",
		"Total: 1 cable(s)",
		"",
	}, "\n")
	assertExactOutput(t, output, want)
}
