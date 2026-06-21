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
	"bytes"
	"errors"
	"strings"
	"testing"
)

// TestWriteCSV_Basic verifies WriteCSV emits the fixed header and a single data
// row whose columns exactly match a fully-specified cable.
//
// Why it matters: the human CSV is a round-trippable export, so the exact column
// order and values must be stable for downstream parsing and diffing.
// Inputs: a ConnectionMap with one connection carrying type, label, length, unit,
// and status. Outputs: a two-line CSV whose header and row equal the expected
// literal strings.
// Data choice: an integer length (3) and populated optional fields make the
// expected row a precise literal, so any column drift or formatting change fails
// the test.
func TestWriteCSV_Basic(t *testing.T) {
	length := 3.0
	cm := ConnectionMap{
		Version: "v1",
		Connections: []ConnectionEntry{
			{
				A: Endpoint{Device: "sw-leaf01", Port: "1/1/1"},
				B: Endpoint{Device: "node01", Port: "HSN 0"},
				Cable: &CableProps{
					Type:       "cat6a",
					Label:      "mgmt",
					Length:     &length,
					LengthUnit: "m",
					Status:     "Connected",
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := WriteCSV(&buf, cm); err != nil {
		t.Fatalf("WriteCSV: %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (header + 1 row), got %d:\n%s", len(lines), got)
	}

	wantHeader := "a_device,a_port,a_mac,b_device,b_port,b_mac,type,label,color,length,length_unit,status"
	if lines[0] != wantHeader {
		t.Errorf("header = %q, want %q", lines[0], wantHeader)
	}

	wantRow := "sw-leaf01,1/1/1,,node01,HSN 0,,cat6a,mgmt,,3,m,Connected"
	if lines[1] != wantRow {
		t.Errorf("row = %q, want %q", lines[1], wantRow)
	}
}

// TestWriteCSV_WithDefaults verifies a CableDefaults block is emitted as a
// _defaults sentinel row between the header and the data rows.
//
// Why it matters: cable defaults must survive export so a re-import applies the
// same fallback properties, and the sentinel row is how the CSV format encodes
// them.
// Inputs: a ConnectionMap with CableDefaults and one bare connection. Outputs: a
// three-line CSV whose second line begins with "_defaults,".
// Data choice: setting only status and length_unit on the defaults keeps the
// focus on the sentinel-row placement rather than per-column values.
func TestWriteCSV_WithDefaults(t *testing.T) {
	cm := ConnectionMap{
		Version: "v1",
		CableDefaults: &CableDefaults{
			Status:     "Connected",
			LengthUnit: "m",
		},
		Connections: []ConnectionEntry{
			{
				A: Endpoint{Device: "sw01", Port: "1"},
				B: Endpoint{Device: "sw02", Port: "1"},
			},
		},
	}

	var buf bytes.Buffer
	if err := WriteCSV(&buf, cm); err != nil {
		t.Fatalf("WriteCSV: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + defaults + 1 row), got %d:\n%s", len(lines), buf.String())
	}

	if !strings.HasPrefix(lines[1], "_defaults,") {
		t.Errorf("defaults row = %q, expected prefix '_defaults,'", lines[1])
	}
}

// TestWriteCSV_RoundTrip verifies output from WriteCSV parses back through
// ParseConnectionsCSV with endpoints, cable, and defaults preserved.
//
// Why it matters: write and parse are inverse halves of the human CSV workflow,
// so a value written must read back identically or edits would silently drift.
// Inputs: a ConnectionMap with defaults and one connection, written then
// re-parsed. Outputs: a parsed map whose endpoints, cable type, and default
// status match the original.
// Data choice: combining a defaults block with a per-connection cable exercises
// both the sentinel-row and data-row round trips in one pass.
func TestWriteCSV_RoundTrip(t *testing.T) {
	length := 15.0
	cm := ConnectionMap{
		Version: "v1",
		CableDefaults: &CableDefaults{
			Status: "Connected",
		},
		Connections: []ConnectionEntry{
			{
				A:     Endpoint{Device: "sw01", Port: "1/1/1"},
				B:     Endpoint{Device: "sw02", Port: "1/1/25"},
				Cable: &CableProps{Type: "cat6a", Length: &length, LengthUnit: "m"},
			},
		},
	}

	var buf bytes.Buffer
	if err := WriteCSV(&buf, cm); err != nil {
		t.Fatalf("WriteCSV: %v", err)
	}

	parsed, err := ParseConnectionsCSV(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("ParseConnectionsCSV round-trip: %v", err)
	}

	if len(parsed.Connections) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(parsed.Connections))
	}

	entry := parsed.Connections[0]
	if entry.A.Device != "sw01" || entry.A.Port != "1/1/1" {
		t.Errorf("A endpoint = %+v, want {sw01, 1/1/1}", entry.A)
	}
	if entry.B.Device != "sw02" || entry.B.Port != "1/1/25" {
		t.Errorf("B endpoint = %+v, want {sw02, 1/1/25}", entry.B)
	}
	if entry.Cable == nil || entry.Cable.Type != "cat6a" {
		t.Errorf("cable type = %v, want cat6a", entry.Cable)
	}
	if parsed.CableDefaults == nil || parsed.CableDefaults.Status != "Connected" {
		t.Errorf("cable defaults = %+v, want Status=Connected", parsed.CableDefaults)
	}
}

// ========== additional branch-coverage tests ==========

// failingWriter is an io.Writer that always fails, used to drive WriteCSV's
// write-error branches.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

// TestFormatLength verifies formatLength renders whole numbers without a decimal
// and fractional values with one.
//
// Why it matters: cable length is written into CSV, so a 3-meter cable must read
// as "3" (not "3.0") while a 2.5-meter cable keeps its fraction for fidelity.
// Inputs: an integral and a fractional float64. Outputs: "3" and "2.5"
// respectively.
// Data choice: 3.0 drives the integer-notation branch and 2.5 drives the %g
// fractional branch, the two formatting paths the function distinguishes.
func TestFormatLength(t *testing.T) {
	cases := []struct {
		name string
		in   float64
		want string
	}{
		{"integer", 3.0, "3"},
		{"fractional", 2.5, "2.5"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatLength(tt.in); got != tt.want {
				t.Errorf("formatLength(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestWriteCSV_WriteErrors verifies WriteCSV returns a wrapped error when the
// underlying writer fails while emitting the defaults row or a data row.
//
// Why it matters: a failed write (full disk, broken pipe) must surface as an
// error identifying which row failed rather than producing a truncated file
// silently.
// Inputs: a failing writer plus a ConnectionMap whose relevant row contains a
// field larger than the csv writer's 4 KiB buffer. Outputs: a non-nil error
// containing the row-specific message.
// Data choice: a 5000-character field forces the buffered csv writer to flush
// mid-row, the only way to make writer.Write itself return the underlying error
// and reach the per-row error branches.
func TestWriteCSV_WriteErrors(t *testing.T) {
	bigField := strings.Repeat("x", 5000)
	cases := []struct {
		name    string
		cm      ConnectionMap
		wantSub string
	}{
		{
			name:    "defaults row",
			cm:      ConnectionMap{CableDefaults: &CableDefaults{Status: bigField}},
			wantSub: "writing CSV defaults row",
		},
		{
			name: "data row",
			cm: ConnectionMap{
				Connections: []ConnectionEntry{
					{A: Endpoint{Device: "a", Port: bigField}, B: Endpoint{Device: "b", Port: "2"}},
				},
			},
			wantSub: "writing CSV row",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := WriteCSV(failingWriter{}, tt.cm)
			if err == nil {
				t.Fatal("expected error from failing writer")
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.wantSub)
			}
		})
	}
}
