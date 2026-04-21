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
	"os"
	"strings"
	"testing"
)

func TestParseInterfacesCSV_Fixture(t *testing.T) {
	f, err := os.Open("../../../testdata/fixtures/cani/connections_interfaces.csv")
	if err != nil {
		t.Fatalf("opening fixture: %v", err)
	}
	defer f.Close()

	cm, err := ParseInterfacesCSV(f)
	if err != nil {
		t.Fatalf("ParseInterfacesCSV: %v", err)
	}

	if cm.Version != "v1" {
		t.Errorf("version = %q, want v1", cm.Version)
	}

	// Fixture has 6 cables (12 interface rows, 2 per cable)
	if got := len(cm.Connections); got != 6 {
		t.Fatalf("connections count = %d, want 6", got)
	}

	// Verify all expected device pairs appear (order within pair is by UUID sort)
	type pair struct{ aDevice, aPort, bDevice, bPort string }
	found := make(map[pair]bool)
	for _, c := range cm.Connections {
		found[pair{c.A.Device, c.A.Port, c.B.Device, c.B.Port}] = true
	}

	expected := []pair{
		{"GH-x3701u34", "iLO", "MAN-x3701u48", "1"},
		{"GH-x3701u34", "HSN 0", "HSNS-x3701u43", "1"},
		{"BBL-x3701u45", "1/1/1", "BBS-x3516u39", "1/1/25"},
		{"HSNS-x3701u43", "33", "HSNS-x3702u43", "33"},
		{"DL-x3507u25", "Gig-E 1", "MANB-x3516u27", "1/1/3"},
		{"GH-x3701u26", "MGMT 0", "BBL-x3701u45", "1/1/20"},
	}

	for _, e := range expected {
		// Try both orderings since pair ordering is by UUID
		if !found[e] && !found[pair{e.bDevice, e.bPort, e.aDevice, e.aPort}] {
			t.Errorf("missing connection: %s:%s <-> %s:%s", e.aDevice, e.aPort, e.bDevice, e.bPort)
		}
	}
}

func TestParseInterfacesCSV_Label(t *testing.T) {
	f, err := os.Open("../../../testdata/fixtures/cani/connections_interfaces.csv")
	if err != nil {
		t.Fatalf("opening fixture: %v", err)
	}
	defer f.Close()

	cm, err := ParseInterfacesCSV(f)
	if err != nil {
		t.Fatalf("ParseInterfacesCSV: %v", err)
	}

	// The DL-x3507u25 <-> MANB-x3516u27 cable has label "mgmt-data"
	var labelFound bool
	for _, c := range cm.Connections {
		isDLPair := (c.A.Device == "DL-x3507u25" || c.B.Device == "DL-x3507u25") &&
			(c.A.Device == "MANB-x3516u27" || c.B.Device == "MANB-x3516u27")
		if isDLPair && c.Cable != nil && c.Cable.Label == "mgmt-data" {
			labelFound = true
		}
	}
	if !labelFound {
		t.Error("expected cable label 'mgmt-data' on DL-x3507u25 <-> MANB-x3516u27 pair")
	}
}

func TestParseInterfacesCSV_SkipsUncabled(t *testing.T) {
	csv := `name,device__name,id,cable_peer,cable__pk,type,status__name,label
iLO,GH-x3701u34,aaaa0001-0000-0000-0000-000000000001,aaaa0002-0000-0000-0000-000000000001,cccc0001-0000-0000-0000-000000000001,1000base-t,Active,
1,MAN-x3701u48,aaaa0002-0000-0000-0000-000000000001,aaaa0001-0000-0000-0000-000000000001,cccc0001-0000-0000-0000-000000000001,1000base-t,Active,
2,MAN-x3701u48,aaaa0099-0000-0000-0000-000000000001,NULL,NULL,1000base-t,Active,
3,MAN-x3701u48,aaaa0098-0000-0000-0000-000000000001,,NULL,1000base-t,Active,
`
	cm, err := ParseInterfacesCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseInterfacesCSV: %v", err)
	}

	// Should only get 1 connection (the cabled pair), not the uncabled rows
	if got := len(cm.Connections); got != 1 {
		t.Fatalf("connections = %d, want 1", got)
	}
}

func TestParseInterfacesCSV_Dedup(t *testing.T) {
	// Same cable appears from both sides — should produce exactly 1 connection
	csv := `name,device__name,id,cable_peer
port1,switch-a,id-aaa,id-bbb
port2,switch-b,id-bbb,id-aaa
`
	cm, err := ParseInterfacesCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseInterfacesCSV: %v", err)
	}
	if got := len(cm.Connections); got != 1 {
		t.Errorf("connections = %d, want 1 (dedup failed)", got)
	}
}

func TestParseInterfacesCSV_MissingColumns(t *testing.T) {
	csv := `name,device__name,id
port1,switch-a,id-aaa
`
	_, err := ParseInterfacesCSV(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for missing cable_peer column")
	}
	if !strings.Contains(err.Error(), "cable_peer") {
		t.Errorf("error should mention cable_peer, got: %v", err)
	}
}

func TestParseInterfacesCSV_MinimalColumns(t *testing.T) {
	// Only the 4 required columns — no optional ones
	csv := `name,device__name,id,cable_peer
port1,switch-a,id-aaa,id-bbb
port2,switch-b,id-bbb,id-aaa
`
	cm, err := ParseInterfacesCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseInterfacesCSV: %v", err)
	}
	if got := len(cm.Connections); got != 1 {
		t.Errorf("connections = %d, want 1", got)
	}
	c := cm.Connections[0]
	// Verify device and port names are populated
	if c.A.Device == "" || c.A.Port == "" || c.B.Device == "" || c.B.Port == "" {
		t.Errorf("endpoints not fully populated: A=%+v, B=%+v", c.A, c.B)
	}
}

func TestParseInterfacesCSV_NoCabledRows(t *testing.T) {
	csv := `name,device__name,id,cable_peer
port1,switch-a,id-aaa,NULL
port2,switch-b,id-bbb,
`
	_, err := ParseInterfacesCSV(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error when no cables found")
	}
}

func TestParseInterfacesCSV_EmptyFile(t *testing.T) {
	_, err := ParseInterfacesCSV(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error on empty input")
	}
}

func TestParseInterfacesCSV_PartialExport(t *testing.T) {
	// Peer id-ccc is not in the export — should be silently skipped
	csv := `name,device__name,id,cable_peer
port1,switch-a,id-aaa,id-bbb
port2,switch-b,id-bbb,id-aaa
port3,switch-c,id-ccc-has-peer,id-ddd-missing
`
	cm, err := ParseInterfacesCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseInterfacesCSV: %v", err)
	}
	// Only the aaa<->bbb pair should be included
	if got := len(cm.Connections); got != 1 {
		t.Errorf("connections = %d, want 1 (partial export)", got)
	}
}

// ── ParseConnectionsCSV (human-friendly) ─────────────────────────

func TestParseConnectionsCSV_Fixture(t *testing.T) {
	f, err := os.Open("../../../testdata/fixtures/cani/connections.csv")
	if err != nil {
		t.Fatalf("opening fixture: %v", err)
	}
	defer f.Close()

	cm, err := ParseConnectionsCSV(f)
	if err != nil {
		t.Fatalf("ParseConnectionsCSV: %v", err)
	}

	if cm.Version != "v1" {
		t.Errorf("version = %q, want v1", cm.Version)
	}

	// Defaults row should populate CableDefaults
	if cm.CableDefaults == nil {
		t.Fatal("expected CableDefaults from _defaults row")
	}
	if cm.CableDefaults.Status != "Connected" {
		t.Errorf("CableDefaults.Status = %q, want Connected", cm.CableDefaults.Status)
	}

	if got := len(cm.Connections); got != 6 {
		t.Fatalf("connections = %d, want 6", got)
	}

	// First connection: GH-x3701u34:iLO <-> MAN-x3701u48:1 with cat6a, 3m
	c := cm.Connections[0]
	if c.A.Device != "GH-x3701u34" || c.A.Port != "iLO" {
		t.Errorf("connection 0 A = %s:%s, want GH-x3701u34:iLO", c.A.Device, c.A.Port)
	}
	if c.B.Device != "MAN-x3701u48" || c.B.Port != "1" {
		t.Errorf("connection 0 B = %s:%s, want MAN-x3701u48:1", c.B.Device, c.B.Port)
	}
	if c.Cable == nil || c.Cable.Type != "cat6a" {
		t.Errorf("connection 0 cable type = %v, want cat6a", c.Cable)
	}
	if c.Cable.Length == nil || *c.Cable.Length != 3.0 {
		t.Errorf("connection 0 cable length = %v, want 3.0", c.Cable.Length)
	}
	if c.Cable.LengthUnit != "m" {
		t.Errorf("connection 0 cable length_unit = %q, want m", c.Cable.LengthUnit)
	}

	// Connection 3 (ISL): 15m NDR MPO cable
	c3 := cm.Connections[3]
	if c3.Cable == nil || c3.Cable.Type != "hpe-ib-ndr-mpo-mpo-sm-15m" {
		t.Errorf("connection 3 cable type = %v, want hpe-ib-ndr-mpo-mpo-sm-15m", c3.Cable)
	}
	if c3.Cable.Length == nil || *c3.Cable.Length != 15.0 {
		t.Errorf("connection 3 cable length = %v, want 15.0", c3.Cable.Length)
	}
}

func TestParseConnectionsCSV_Label(t *testing.T) {
	f, err := os.Open("../../../testdata/fixtures/cani/connections.csv")
	if err != nil {
		t.Fatalf("opening fixture: %v", err)
	}
	defer f.Close()

	cm, err := ParseConnectionsCSV(f)
	if err != nil {
		t.Fatalf("ParseConnectionsCSV: %v", err)
	}

	// Row 5 (DL-x3507u25 <-> MANB-x3516u27) has label "mgmt-data" and type cat6a
	c := cm.Connections[4]
	if c.Cable == nil || c.Cable.Label != "mgmt-data" {
		t.Error("expected cable label 'mgmt-data' on connection 4")
	}
	if c.Cable.Type != "cat6a" {
		t.Errorf("connection 4 cable type = %q, want cat6a", c.Cable.Type)
	}
}

func TestParseConnectionsCSV_Defaults(t *testing.T) {
	csv := `a_device,a_port,b_device,b_port,type,color,status,length_unit
_defaults,,,,cat6a,blue,Connected,m
switch-a,1,switch-b,1,,,,
switch-a,2,switch-b,2,fiber,red,,
`
	cm, err := ParseConnectionsCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseConnectionsCSV: %v", err)
	}

	if cm.CableDefaults == nil {
		t.Fatal("expected CableDefaults")
	}
	if cm.CableDefaults.Type != "cat6a" {
		t.Errorf("default type = %q, want cat6a", cm.CableDefaults.Type)
	}
	if cm.CableDefaults.Color != "blue" {
		t.Errorf("default color = %q, want blue", cm.CableDefaults.Color)
	}
	if cm.CableDefaults.Status != "Connected" {
		t.Errorf("default status = %q, want Connected", cm.CableDefaults.Status)
	}
	if cm.CableDefaults.LengthUnit != "m" {
		t.Errorf("default length_unit = %q, want m", cm.CableDefaults.LengthUnit)
	}

	// 2 connections (defaults row not counted)
	if got := len(cm.Connections); got != 2 {
		t.Fatalf("connections = %d, want 2", got)
	}

	// Row 1 has no per-row cable props — Cable should be nil (defaults apply at resolve time)
	if cm.Connections[0].Cable != nil {
		t.Error("expected nil Cable on row with no per-row props")
	}

	// Row 2 overrides type and color
	c1 := cm.Connections[1]
	if c1.Cable == nil || c1.Cable.Type != "fiber" {
		t.Errorf("row 2 cable type = %v, want fiber", c1.Cable)
	}
	if c1.Cable.Color != "red" {
		t.Errorf("row 2 cable color = %q, want red", c1.Cable.Color)
	}
}

func TestParseConnectionsCSV_MinimalColumns(t *testing.T) {
	csv := `a_device,a_port,b_device,b_port
switch-a,1,switch-b,1
switch-a,2,switch-b,2
`
	cm, err := ParseConnectionsCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseConnectionsCSV: %v", err)
	}
	if got := len(cm.Connections); got != 2 {
		t.Errorf("connections = %d, want 2", got)
	}
	// No cable props when optional columns absent
	if cm.Connections[0].Cable != nil {
		t.Error("expected nil Cable when no optional columns")
	}
}

func TestParseConnectionsCSV_SkipsIncomplete(t *testing.T) {
	csv := `a_device,a_port,b_device,b_port
switch-a,1,switch-b,1
switch-a,,switch-b,2
,1,switch-b,3
`
	cm, err := ParseConnectionsCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseConnectionsCSV: %v", err)
	}
	if got := len(cm.Connections); got != 1 {
		t.Errorf("connections = %d, want 1 (skip incomplete)", got)
	}
}

func TestParseConnectionsCSV_MissingColumns(t *testing.T) {
	csv := `a_device,a_port,b_device
switch-a,1,switch-b
`
	_, err := ParseConnectionsCSV(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for missing b_port column")
	}
	if !strings.Contains(err.Error(), "b_port") {
		t.Errorf("error should mention b_port, got: %v", err)
	}
}

func TestParseConnectionsCSV_NoRows(t *testing.T) {
	csv := `a_device,a_port,b_device,b_port
`
	_, err := ParseConnectionsCSV(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error when no data rows")
	}
}

func TestParseConnectionsCSV_CableProps(t *testing.T) {
	csv := `a_device,a_port,b_device,b_port,type,label,color,length,length_unit,status
switch-a,1,switch-b,1,cat6,uplink,blue,3,m,Active
`
	cm, err := ParseConnectionsCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseConnectionsCSV: %v", err)
	}
	c := cm.Connections[0]
	if c.Cable == nil {
		t.Fatal("expected Cable to be populated")
	}
	if c.Cable.Type != "cat6" {
		t.Errorf("cable type = %q, want cat6", c.Cable.Type)
	}
	if c.Cable.Label != "uplink" {
		t.Errorf("cable label = %q, want uplink", c.Cable.Label)
	}
	if c.Cable.Color != "blue" {
		t.Errorf("cable color = %q, want blue", c.Cable.Color)
	}
	if c.Cable.Length == nil || *c.Cable.Length != 3.0 {
		t.Errorf("cable length = %v, want 3.0", c.Cable.Length)
	}
	if c.Cable.LengthUnit != "m" {
		t.Errorf("cable length_unit = %q, want m", c.Cable.LengthUnit)
	}
	if c.Cable.Status != "Active" {
		t.Errorf("cable status = %q, want Active", c.Cable.Status)
	}
}

// ── ParseCSV (auto-detection) ────────────────────────────────────

func TestParseCSV_DetectsHumanFormat(t *testing.T) {
	csv := `a_device,a_port,b_device,b_port
switch-a,1,switch-b,1
`
	cm, err := ParseCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseCSV: %v", err)
	}
	if got := len(cm.Connections); got != 1 {
		t.Errorf("connections = %d, want 1", got)
	}
}

func TestParseCSV_DetectsNautobotFormat(t *testing.T) {
	csv := `name,device__name,id,cable_peer
port1,switch-a,id-aaa,id-bbb
port2,switch-b,id-bbb,id-aaa
`
	cm, err := ParseCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseCSV: %v", err)
	}
	if got := len(cm.Connections); got != 1 {
		t.Errorf("connections = %d, want 1", got)
	}
}

func TestParseCSV_UnknownFormat(t *testing.T) {
	csv := `foo,bar,baz
1,2,3
`
	_, err := ParseCSV(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for unknown CSV format")
	}
	if !strings.Contains(err.Error(), "not recognized") {
		t.Errorf("error should mention format not recognized, got: %v", err)
	}
}
