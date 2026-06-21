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
	"errors"
	"strings"
	"testing"
	"testing/iotest"
)

// TestParseInterfacesCSV_SkipsUncabled verifies the Nautobot interfaces parser
// emits a connection only for cabled interface pairs, skipping rows whose
// cable_peer is empty or "NULL".
//
// Why it matters: a Nautobot export lists every interface, but only cabled ones
// describe a physical link, so uncabled rows must not become phantom
// connections.
// Inputs: a CSV with one cabled pair and two uncabled rows (NULL and empty
// cable_peer). Outputs: a ConnectionMap with exactly one connection.
// Data choice: including both a literal "NULL" and an empty cable_peer covers the
// two distinct uncabled sentinels the parser must treat as "no cable".
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

// TestParseInterfacesCSV_Dedup verifies a cable appearing once per side is
// collapsed into a single connection.
//
// Why it matters: every cable shows up twice in a Nautobot interface export, so
// the parser must deduplicate or it would double every link.
// Inputs: two rows describing the same cable from each end (id/cable_peer
// swapped). Outputs: a ConnectionMap with exactly one connection.
// Data choice: mirrored id/cable_peer values are the minimal case that forces the
// lexicographic-smaller-id dedup rule to fire.
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

// TestParseInterfacesCSV_MissingColumns verifies the parser rejects a CSV that
// omits the required cable_peer column.
//
// Why it matters: without cable_peer the parser cannot pair interfaces, so it
// must fail fast with a message naming the missing column.
// Inputs: a CSV header missing cable_peer. Outputs: an error mentioning
// "cable_peer".
// Data choice: dropping exactly the cable_peer column isolates the
// required-column check and lets the assertion match on its name.
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

// TestParseInterfacesCSV_MinimalColumns verifies a CSV with only the four
// required columns parses and fully populates both endpoints.
//
// Why it matters: optional columns (type, status, label) may be absent from an
// export, so the parser must still produce usable device/port endpoints.
// Inputs: a CSV with only name, device__name, id, cable_peer for one pair.
// Outputs: one connection whose A and B device and port fields are all non-empty.
// Data choice: providing strictly the required columns proves the parser does not
// depend on any optional field to build an endpoint.
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

// TestParseInterfacesCSV_NoCabledRows verifies the parser errors when no row has
// a usable cable_peer.
//
// Why it matters: a CSV with zero cables yields no connections, and the parser
// must signal that rather than return an empty, silently-useless map.
// Inputs: a CSV whose only rows have NULL and empty cable_peer values. Outputs:
// an error.
// Data choice: both uncabled sentinels appear so the "zero connections" error is
// reached only after every row is correctly skipped.
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

// TestParseInterfacesCSV_EmptyFile verifies the parser errors on empty input.
//
// Why it matters: an empty file has no header row, so the parser must fail
// rather than treat it as a valid zero-cable export.
// Inputs: an empty reader. Outputs: an error.
// Data choice: the empty string is the smallest input that exercises the
// header-read failure path.
func TestParseInterfacesCSV_EmptyFile(t *testing.T) {
	_, err := ParseInterfacesCSV(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error on empty input")
	}
}

// TestParseInterfacesCSV_PartialExport verifies a cabled interface whose peer is
// absent from the export is silently skipped.
//
// Why it matters: a filtered or partial export can reference a peer that was not
// included, and the parser must drop that half-link instead of inventing a
// one-sided connection.
// Inputs: a CSV with one complete pair plus a third row whose cable_peer is not
// present. Outputs: a ConnectionMap with exactly one connection.
// Data choice: a dangling peer id alongside a valid pair proves only the
// resolvable pair survives.
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

// TestParseConnectionsCSV_Defaults verifies a _defaults sentinel row populates
// CableDefaults and that per-row props override while bare rows stay nil.
//
// Why it matters: the human CSV format mirrors YAML cable_defaults, so the
// sentinel row must become defaults (applied later at resolve time) and explicit
// per-row props must win.
// Inputs: a CSV with a _defaults row plus a bare connection and a second
// connection overriding type and color. Outputs: CableDefaults set from the
// sentinel, a nil Cable on the bare row, and the override values on the second.
// Data choice: distinct default and override values (cat6a/blue vs fiber/red)
// make it unambiguous which row supplied each field.
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

// TestParseConnectionsCSV_MinimalColumns verifies rows with only the four
// required columns parse with no cable properties attached.
//
// Why it matters: the simplest human CSV lists just endpoints, and the parser
// must accept it without fabricating an empty Cable block.
// Inputs: a CSV with only a_device, a_port, b_device, b_port for two rows.
// Outputs: two connections, the first with a nil Cable.
// Data choice: omitting every optional column drives the no-cable-props branch of
// the row builder.
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

// TestParseConnectionsCSV_SkipsIncomplete verifies rows missing a required
// endpoint field are skipped rather than parsed into broken connections.
//
// Why it matters: a half-filled row cannot describe a cable, so it must be
// dropped to keep the resulting map valid.
// Inputs: a CSV with one complete row, one missing a_port, and one missing
// a_device. Outputs: a ConnectionMap with exactly one connection.
// Data choice: omitting a different required field in each bad row proves the
// skip applies to any missing endpoint component, not just one column.
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

// TestParseConnectionsCSV_MissingColumns verifies the parser rejects a CSV that
// omits a required endpoint column.
//
// Why it matters: without b_port the parser cannot form the B endpoint, so it
// must fail with a message naming the missing column.
// Inputs: a CSV header missing b_port. Outputs: an error mentioning "b_port".
// Data choice: dropping exactly b_port isolates the required-column check and
// lets the assertion match on its name.
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

// TestParseConnectionsCSV_NoRows verifies a header-only CSV errors because it
// contains no connections.
//
// Why it matters: a CSV with valid columns but no data rows yields nothing to
// import, and the parser must say so rather than return an empty map.
// Inputs: a CSV with only the header line. Outputs: an error.
// Data choice: the header-only input is the minimal case that passes column
// validation yet produces zero complete rows.
func TestParseConnectionsCSV_NoRows(t *testing.T) {
	csv := `a_device,a_port,b_device,b_port
`
	_, err := ParseConnectionsCSV(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error when no data rows")
	}
}

// TestParseConnectionsCSV_CableProps verifies every optional cable column is
// parsed onto the connection's Cable, including numeric length.
//
// Why it matters: the human CSV must round-trip full cable metadata, so each
// optional column has to map to its CableProps field with correct typing.
// Inputs: one row setting type, label, color, length, length_unit, and status.
// Outputs: a Cable with all six fields populated and Length parsed to 3.0.
// Data choice: a distinct value per column (and an integer length) makes a
// mis-mapped or unparsed field immediately visible.
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

// TestParseCSV_DetectsHumanFormat verifies ParseCSV routes an a_device header to
// the human-friendly parser.
//
// Why it matters: ParseCSV auto-detects format so callers need not specify it,
// and an a_device header must select the human parser.
// Inputs: a CSV whose header contains a_device. Outputs: a ConnectionMap with one
// connection.
// Data choice: the a_device column is the documented discriminator for the human
// format, so its presence alone should drive detection.
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

// TestParseCSV_DetectsNautobotFormat verifies ParseCSV routes a cable_peer header
// to the Nautobot interfaces parser.
//
// Why it matters: auto-detection must select the Nautobot parser when it sees the
// cable_peer column so a raw export imports without manual format selection.
// Inputs: a CSV whose header contains cable_peer for one mirrored pair. Outputs:
// a ConnectionMap with one connection.
// Data choice: the cable_peer column is the documented discriminator for the
// Nautobot format, so its presence alone should drive detection.
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

// TestParseCSV_UnknownFormat verifies ParseCSV errors when the header matches
// neither known format.
//
// Why it matters: an unrecognized header means the importer cannot know how to
// interpret the rows, so it must fail with a clear "not recognized" message
// rather than guess.
// Inputs: a CSV with arbitrary columns (foo, bar, baz). Outputs: an error
// mentioning "not recognized".
// Data choice: column names that contain neither a_device nor cable_peer
// guarantee both detection branches miss, reaching the unrecognized-format error.
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

// ========== additional branch-coverage tests ==========

// TestParseCSV_ReadError verifies ParseCSV surfaces an error when the underlying
// reader fails before any header can be read.
//
// Why it matters: ParseCSV reads the whole stream up front, so an I/O failure
// must propagate instead of being mistaken for empty input.
// Inputs: a reader that always returns an error. Outputs: a non-nil error.
// Data choice: iotest.ErrReader is the stdlib way to force io.ReadAll to fail,
// driving the read-error branch deterministically.
func TestParseCSV_ReadError(t *testing.T) {
	_, err := ParseCSV(iotest.ErrReader(errors.New("boom")))
	if err == nil {
		t.Fatal("expected error when the reader fails")
	}
}

// TestParseCSV_EmptyInput verifies ParseCSV errors on empty input that has no
// header row to detect a format from.
//
// Why it matters: format detection reads the header first, so empty input must
// fail at the header-read step rather than silently choosing a parser.
// Inputs: an empty reader. Outputs: a non-nil error.
// Data choice: the empty string yields an immediate EOF on the header read, the
// exact branch under test.
func TestParseCSV_EmptyInput(t *testing.T) {
	_, err := ParseCSV(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error on empty input (no header)")
	}
}

// TestParseConnectionsCSV_EmptyInput verifies the human parser errors when the
// header row cannot be read.
//
// Why it matters: column validation depends on the header, so empty input must
// fail at the header read instead of proceeding with no columns.
// Inputs: an empty reader. Outputs: a non-nil error.
// Data choice: the empty string forces the header-read EOF branch that the
// populated-CSV tests never reach.
func TestParseConnectionsCSV_EmptyInput(t *testing.T) {
	_, err := ParseConnectionsCSV(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error on empty input (no header)")
	}
}

// TestParseConnectionsCSV_MalformedRow verifies the human parser surfaces a
// row-level CSV parse error.
//
// Why it matters: a corrupt data row must abort the import with an error rather
// than be silently dropped or mis-parsed.
// Inputs: a valid header followed by a row containing a bare double quote.
// Outputs: a non-nil error.
// Data choice: a bare quote in a non-quoted field is the canonical malformed-CSV
// trigger for encoding/csv, reaching the per-row read-error branch.
func TestParseConnectionsCSV_MalformedRow(t *testing.T) {
	csv := "a_device,a_port,b_device,b_port\nx\"y,1,switch-b,2\n"
	_, err := ParseConnectionsCSV(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for malformed CSV row")
	}
}

// TestParseConnectionsCSV_EmptyDefaults verifies an all-empty _defaults row
// produces no CableDefaults while normal rows still parse.
//
// Why it matters: a sentinel row that sets nothing should not attach an empty
// defaults block, which would otherwise mask the "no defaults" state downstream.
// Inputs: a CSV with a _defaults row whose optional fields are all blank plus one
// real connection. Outputs: a nil CableDefaults and one connection.
// Data choice: blanking every default column drives buildCableDefaults' return-nil
// branch that the populated-defaults test cannot reach.
func TestParseConnectionsCSV_EmptyDefaults(t *testing.T) {
	csv := `a_device,a_port,b_device,b_port,type,color,status,length_unit
_defaults,,,,,,,
switch-a,1,switch-b,1
`
	cm, err := ParseConnectionsCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseConnectionsCSV: %v", err)
	}
	if cm.CableDefaults != nil {
		t.Errorf("CableDefaults = %+v, want nil for an all-empty _defaults row", cm.CableDefaults)
	}
	if len(cm.Connections) != 1 {
		t.Errorf("connections = %d, want 1", len(cm.Connections))
	}
}

// TestParseInterfacesCSV_MalformedRow verifies the Nautobot parser surfaces a
// row-level CSV parse error.
//
// Why it matters: a corrupt interface row must abort the import with an error so
// a malformed export is never partially and silently accepted.
// Inputs: a valid header followed by a row containing a bare double quote.
// Outputs: a non-nil error.
// Data choice: a bare quote in a non-quoted field is the canonical malformed-CSV
// trigger, reaching the readInterfaceRows error path and its propagation.
func TestParseInterfacesCSV_MalformedRow(t *testing.T) {
	csv := "name,device__name,id,cable_peer\nx\"y,sw,id-a,id-b\n"
	_, err := ParseInterfacesCSV(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for malformed CSV row")
	}
}

// TestParseInterfacesCSV_SkipsEmptyID verifies a cabled row with an empty id is
// skipped while a valid pair still produces a connection.
//
// Why it matters: an interface with a cable_peer but no id cannot be paired, so
// it must be dropped rather than corrupt the interface index.
// Inputs: a CSV with one complete mirrored pair plus a third row that has a
// cable_peer but a blank id. Outputs: a ConnectionMap with exactly one
// connection.
// Data choice: pairing the empty-id row with a resolvable pair drives the
// empty-id skip branch without collapsing the whole parse to zero connections.
func TestParseInterfacesCSV_SkipsEmptyID(t *testing.T) {
	csv := `name,device__name,id,cable_peer
portX,switch-a,id-aaa,id-bbb
portY,switch-b,id-bbb,id-aaa
portZ,switch-c,,id-ccc
`
	cm, err := ParseInterfacesCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseInterfacesCSV: %v", err)
	}
	if len(cm.Connections) != 1 {
		t.Errorf("connections = %d, want 1 (empty-id row skipped)", len(cm.Connections))
	}
}
