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
	"strings"
	"testing"
)

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

	wantHeader := "a_device,a_port,b_device,b_port,type,label,color,length,length_unit,status"
	if lines[0] != wantHeader {
		t.Errorf("header = %q, want %q", lines[0], wantHeader)
	}

	wantRow := "sw-leaf01,1/1/1,node01,HSN 0,cat6a,mgmt,,3,m,Connected"
	if lines[1] != wantRow {
		t.Errorf("row = %q, want %q", lines[1], wantRow)
	}
}

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
