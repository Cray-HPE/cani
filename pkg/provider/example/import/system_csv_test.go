package import_

import (
	"os"
	"testing"
)

// TestIsSystemCSV verifies IsSystemCSV detects headers that identify the system
// CSV format.
//
// Why it matters: ImportCSV routes to the system parser based only on the header,
// so Section aliases and spacing must be recognized while BOM headers stay false.
// Inputs: headers with Section variants, RecordType alias, BOM-only columns, and
// an empty header. Outputs: boolean format-detection results.
// Data choice: the table covers the positive aliases, whitespace trimming, and
// the negative BOM/empty cases that drive routing.
func TestIsSystemCSV(t *testing.T) {
	tests := []struct {
		name   string
		header []string
		want   bool
	}{
		{"has Section column", []string{"Section", "PartNumber", "Name"}, true},
		{"lowercase section", []string{"section", "partnumber", "name"}, true},
		{"no Section column", []string{"PartNumber", "Description", "Quantity"}, false},
		{"empty header", []string{}, false},
		{"section with spaces", []string{" Section ", "PartNumber"}, true},
		{"RecordType alias", []string{"RecordType", "PartNumber"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSystemCSV(tt.header)
			if got != tt.want {
				t.Errorf("IsSystemCSV(%v) = %v, want %v", tt.header, got, tt.want)
			}
		})
	}
}

// TestParseSystemCSV verifies ParseSystemCSV parses the system fixture into all
// expected section buckets and surfaces file-open errors.
//
// Why it matters: system CSV import relies on bucket counts before transform can
// create roles, racks, devices, modules, interfaces, and connections.
// Inputs: the system.csv fixture and a nonexistent path. Outputs: populated
// SystemCSV data or an open error.
// Data choice: the fixture has a known mixed-section shape, while the bad path
// isolates the file-open branch.
func TestParseSystemCSV(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		expectErr bool
		errorMsg  string
		validate  func(t *testing.T, data *SystemCSV)
	}{
		{
			name: "valid system.csv fixture",
			path: "../../../../testdata/fixtures/example/system.csv",
			validate: func(t *testing.T, data *SystemCSV) {
				if len(data.Roles) != 3 {
					t.Errorf("expected 3 roles, got %d", len(data.Roles))
				}
				if len(data.Racks) != 2 {
					t.Errorf("expected 2 racks, got %d", len(data.Racks))
				}
				if len(data.Devices) != 6 {
					t.Errorf("expected 6 devices, got %d", len(data.Devices))
				}
				if len(data.Modules) != 3 {
					t.Errorf("expected 3 modules, got %d", len(data.Modules))
				}
				if len(data.Interfaces) != 2 {
					t.Errorf("expected 2 interfaces, got %d", len(data.Interfaces))
				}
				if len(data.Connections) != 3 {
					t.Errorf("expected 3 connections, got %d", len(data.Connections))
				}
			},
		},
		{
			name:      "bad path",
			path:      "nonexistent.csv",
			expectErr: true,
			errorMsg:  "failed to open system CSV file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ParseSystemCSV(tt.path)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errorMsg)
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, data)
			}
		})
	}
}

// TestParseSystemCSV_Defaults verifies parsed global defaults apply to sparse
// records without overriding explicit values.
//
// Why it matters: system CSV defaults let operators avoid repeating status and
// role values, so default layering must preserve explicit row data.
// Inputs: the system.csv fixture plus synthetic device records with empty and
// explicit Status fields. Outputs: merged SystemRecord values.
// Data choice: the fixture's Active default and an explicit Planned status prove
// both fill and preserve behavior.
func TestParseSystemCSV_Defaults(t *testing.T) {
	data, err := ParseSystemCSV("../../../../testdata/fixtures/example/system.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Global _defaults row sets Status=Active
	if data.Defaults.Status != "Active" {
		t.Errorf("global defaults Status = %q, want %q", data.Defaults.Status, "Active")
	}

	// Verify defaults are applied
	rec := data.ApplyDefaults(SystemRecord{Section: "device", Name: "test"})
	if rec.Status != "Active" {
		t.Errorf("applied defaults Status = %q, want %q", rec.Status, "Active")
	}

	// Explicit value takes precedence
	rec = data.ApplyDefaults(SystemRecord{Section: "device", Name: "test", Status: "Planned"})
	if rec.Status != "Planned" {
		t.Errorf("explicit Status should be preserved, got %q", rec.Status)
	}
}

// TestParseSystemCSV_RoleFields verifies role rows preserve names and content
// type strings from the system fixture.
//
// Why it matters: role records feed inventory metadata used by later device and
// rack role references.
// Inputs: the system.csv fixture. Outputs: the first parsed role record.
// Data choice: ComputeNode is the first fixture role and has a known dcim.device
// content type, making the assertion stable.
func TestParseSystemCSV_RoleFields(t *testing.T) {
	data, err := ParseSystemCSV("../../../../testdata/fixtures/example/system.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(data.Roles) < 1 {
		t.Fatal("expected at least 1 role")
	}

	role := data.Roles[0]
	if role.Name != "ComputeNode" {
		t.Errorf("first role Name = %q, want %q", role.Name, "ComputeNode")
	}
	if role.ContentTypes != "dcim.device" {
		t.Errorf("first role ContentTypes = %q, want %q", role.ContentTypes, "dcim.device")
	}
}

// TestParseSystemCSV_StatusSection verifies status rows are bucketed into the
// Statuses slice with their content types preserved.
//
// Why it matters: statuses are a first-class catalog the transform turns into
// metadata, so the parser must route `status` rows separately from roles and
// keep their content-type column intact.
// Inputs: a temp CSV with one status row carrying two content types. Outputs: a
// single Statuses record named "Planned".
// Data choice: a status row with two comma-separated content types proves both
// section routing and that the ContentTypes column is captured verbatim.
func TestParseSystemCSV_StatusSection(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/status.csv"
	content := "Section,Name,ContentTypes\n" +
		"status,Planned,\"dcim.device,dcim.rack\"\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	data, err := ParseSystemCSV(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data.Statuses) != 1 {
		t.Fatalf("Statuses = %d, want 1", len(data.Statuses))
	}
	if data.Statuses[0].Name != "Planned" {
		t.Errorf("status Name = %q, want %q", data.Statuses[0].Name, "Planned")
	}
	if data.Statuses[0].ContentTypes != "dcim.device,dcim.rack" {
		t.Errorf("status ContentTypes = %q, want %q", data.Statuses[0].ContentTypes, "dcim.device,dcim.rack")
	}
}

// TestParseSystemCSV_RoleColorDescription verifies a role row's Color and
// Description columns are parsed into the record.
//
// Why it matters: roles support a display color and description; the parser must
// capture both columns so the transform can thread them into the catalog.
// Inputs: a temp CSV with a role row setting Color and Description. Outputs: the
// parsed role record with both fields.
// Data choice: a single role with a color and a description isolates the new
// column parsing from other sections.
func TestParseSystemCSV_RoleColorDescription(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/role.csv"
	content := "Section,Name,Color,Description\n" +
		"role,ComputeNode,blue,GPU compute nodes\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	data, err := ParseSystemCSV(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data.Roles) != 1 {
		t.Fatalf("Roles = %d, want 1", len(data.Roles))
	}
	if data.Roles[0].Color != "blue" {
		t.Errorf("role Color = %q, want %q", data.Roles[0].Color, "blue")
	}
	if data.Roles[0].Description != "GPU compute nodes" {
		t.Errorf("role Description = %q, want %q", data.Roles[0].Description, "GPU compute nodes")
	}
}

// TestParseSystemCSV_DeviceFields verifies a device row preserves part number,
// rack placement, role, face, and serial values.
//
// Why it matters: these fields become the core CANI device placement and
// identity data during transform.
// Inputs: the system.csv fixture. Outputs: the GH-x3701u34 SystemRecord fields.
// Data choice: GH-x3701u34 is a fixture device with rack, position, face, role,
// and serial populated, exercising the full device field set.
func TestParseSystemCSV_DeviceFields(t *testing.T) {
	data, err := ParseSystemCSV("../../../../testdata/fixtures/example/system.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find the GH-x3701u34 device
	var found *SystemRecord
	for _, d := range data.Devices {
		if d.Name == "GH-x3701u34" {
			found = &d
			break
		}
	}
	if found == nil {
		t.Fatal("device GH-x3701u34 not found")
	}

	if found.PartNumber != "hpe-xd670" {
		t.Errorf("PartNumber = %q, want %q", found.PartNumber, "hpe-xd670")
	}
	if found.Rack != "x3701" {
		t.Errorf("Rack = %q, want %q", found.Rack, "x3701")
	}
	if found.Position != 34 {
		t.Errorf("Position = %d, want %d", found.Position, 34)
	}
	if found.Face != "front" {
		t.Errorf("Face = %q, want %q", found.Face, "front")
	}
	if found.Role != "ComputeNode" {
		t.Errorf("Role = %q, want %q", found.Role, "ComputeNode")
	}
	if found.Serial != "5UF435KF42" {
		t.Errorf("Serial = %q, want %q", found.Serial, "5UF435KF42")
	}
}

// TestParseSystemCSV_ConnectionFields verifies connection rows preserve both
// endpoints and cable properties from the system fixture.
//
// Why it matters: transform turns connection records into CANI cables, so endpoint
// device/port and cable metadata must survive parsing exactly.
// Inputs: the system.csv fixture. Outputs: the first parsed connection record.
// Data choice: the first fixture connection includes endpoints, part number, and
// color, covering the fields needed for cable creation.
func TestParseSystemCSV_ConnectionFields(t *testing.T) {
	data, err := ParseSystemCSV("../../../../testdata/fixtures/example/system.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(data.Connections) < 1 {
		t.Fatal("expected at least 1 connection")
	}

	conn := data.Connections[0]
	if conn.ADevice != "GH-x3701u34" {
		t.Errorf("ADevice = %q, want %q", conn.ADevice, "GH-x3701u34")
	}
	if conn.APort != "iLO" {
		t.Errorf("APort = %q, want %q", conn.APort, "iLO")
	}
	if conn.BDevice != "MAN-x3701u48" {
		t.Errorf("BDevice = %q, want %q", conn.BDevice, "MAN-x3701u48")
	}
	if conn.BPort != "1" {
		t.Errorf("BPort = %q, want %q", conn.BPort, "1")
	}
	if conn.PartNumber != "hpe-3m-cat6-stp" {
		t.Errorf("PartNumber = %q, want %q", conn.PartNumber, "hpe-3m-cat6-stp")
	}
	if conn.Color != "blue" {
		t.Errorf("Color = %q, want %q", conn.Color, "blue")
	}
}

// TestParseSystemCSV_InterfaceFields verifies interface rows preserve owner,
// interface name, and MAC address fields.
//
// Why it matters: interface rows annotate device/module interface specs during
// transform, so the owner/name lookup keys and MAC value must parse intact.
// Inputs: the system.csv fixture. Outputs: the first parsed interface record.
// Data choice: the fixture's iLO row has all interface-specific fields populated,
// making it a compact contract check.
func TestParseSystemCSV_InterfaceFields(t *testing.T) {
	data, err := ParseSystemCSV("../../../../testdata/fixtures/example/system.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(data.Interfaces) < 1 {
		t.Fatal("expected at least 1 interface")
	}

	iface := data.Interfaces[0]
	if iface.Device != "GH-x3701u34" {
		t.Errorf("Device = %q, want %q", iface.Device, "GH-x3701u34")
	}
	if iface.Name != "iLO" {
		t.Errorf("Name = %q, want %q", iface.Name, "iLO")
	}
	if iface.MacAddress != "aa:bb:cc:dd:ee:01" {
		t.Errorf("MacAddress = %q, want %q", iface.MacAddress, "aa:bb:cc:dd:ee:01")
	}
}

// TestNormalizeSystemHeader verifies system CSV header aliases normalize to the
// parser's canonical column keys.
//
// Why it matters: system CSV files can use human-friendly headers, and parser
// routing depends on aliases landing on the right SystemRecord fields.
// Inputs: aliases for section, part number, rack face, serial, endpoints,
// content types, module bay, position, and cable color. Outputs: normalized keys.
// Data choice: the table samples one alias from each major section field family.
func TestNormalizeSystemHeader(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Section", "section"},
		{"Part Number", "partnumber"},
		{"part_number", "partnumber"},
		{"Rack Face", "face"},
		{"Serial Number", "serial"},
		{"A Device", "adevice"},
		{"Content Types", "contenttypes"},
		{"Module Bay", "bay"},
		{"U Position", "position"},
		{"Cable Color", "color"},
		{"Description", "description"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeSystemHeader(tt.input)
			if got != tt.want {
				t.Errorf("normalizeSystemHeader(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestParseSystemRow_EmptySection verifies parseSystemRow rejects rows without a
// Section value.
//
// Why it matters: Section controls bucket routing, so an empty section cannot be
// safely imported.
// Inputs: a row with an empty Section cell. Outputs: a non-nil empty Section
// error.
// Data choice: a one-cell row isolates the Section guard without involving other
// optional fields.
func TestParseSystemRow_EmptySection(t *testing.T) {
	idx := systemColumnIndex{section: 0}
	_, err := parseSystemRow([]string{""}, idx, 2)
	if err == nil {
		t.Fatal("expected error for empty section")
	}
	if !contains(err.Error(), "empty Section") {
		t.Errorf("expected 'empty Section' error, got %q", err.Error())
	}
}

// TestParseSystemRow_InvalidQty verifies parseSystemRow rejects non-numeric Qty
// values.
//
// Why it matters: Qty controls object multiplication, so an unparsable value must
// be dropped instead of silently defaulting.
// Inputs: a row whose Qty cell is "abc". Outputs: a non-nil error.
// Data choice: the non-numeric value isolates strconv.Atoi failure for Qty.
func TestParseSystemRow_InvalidQty(t *testing.T) {
	idx := systemColumnIndex{section: 0, qty: 1, partNumber: -1, name: -1, rack: -1,
		position: -1, face: -1, role: -1, status: -1, serial: -1,
		device: -1, bay: -1, aDevice: -1, aPort: -1, bDevice: -1, bPort: -1,
		color: -1, length: -1, lengthUnit: -1, location: -1, contentTypes: -1}
	_, err := parseSystemRow([]string{"device", "abc"}, idx, 2)
	if err == nil {
		t.Fatal("expected error for invalid qty")
	}
}

// TestParseSystemRow_DefaultQty verifies parseSystemRow defaults Qty to one when
// the column is absent.
//
// Why it matters: system CSV rows commonly omit Qty, and transform expects a
// positive count for single-object rows.
// Inputs: a row with only a Section cell and no Qty column. Outputs: a
// SystemRecord with Qty set to 1.
// Data choice: omitting the Qty index entirely proves the default branch rather
// than an empty string branch.
func TestParseSystemRow_DefaultQty(t *testing.T) {
	idx := systemColumnIndex{section: 0, qty: -1, partNumber: -1, name: -1, rack: -1,
		position: -1, face: -1, role: -1, status: -1, serial: -1,
		device: -1, bay: -1, aDevice: -1, aPort: -1, bDevice: -1, bPort: -1,
		color: -1, length: -1, lengthUnit: -1, location: -1, contentTypes: -1}
	rec, err := parseSystemRow([]string{"device"}, idx, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Qty != 1 {
		t.Errorf("default Qty = %d, want 1", rec.Qty)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

// TestParseSystemRow_ZeroQty verifies a Qty below one is rejected.
//
// Why it matters: quantity multiplies a row into N objects, so a zero or negative
// count is meaningless and must fail the row rather than create nothing silently.
// Inputs: a row whose Qty cell is "0". Outputs: a non-nil "Qty must be >= 1"
// error.
// Data choice: zero is the boundary value just below the minimum, pinning the
// >= 1 guard.
func TestParseSystemRow_ZeroQty(t *testing.T) {
	idx := systemColumnIndex{section: 0, qty: 1, partNumber: -1, name: -1, rack: -1,
		position: -1, face: -1, role: -1, status: -1, serial: -1,
		device: -1, bay: -1, aDevice: -1, aPort: -1, bDevice: -1, bPort: -1,
		color: -1, length: -1, lengthUnit: -1, location: -1, contentTypes: -1, macAddress: -1}
	_, err := parseSystemRow([]string{"device", "0"}, idx, 2)
	if err == nil {
		t.Fatal("expected error for zero qty")
	}
	if !contains(err.Error(), "Qty must be >= 1") {
		t.Errorf("error = %q, want containing 'Qty must be >= 1'", err.Error())
	}
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestParseSystemHeader_MissingSection verifies the header parser rejects a header
// with no Section column.
//
// Why it matters: the Section column drives every row's routing, so a header that
// omits it cannot be parsed and must fail with a clear message instead of
// silently mis-bucketing rows.
// Inputs: a header slice without a Section column. Outputs: a non-nil error
// naming the missing column.
// Data choice: a two-column header with only PartNumber and Name is the smallest
// header that is well-formed yet lacks Section.
func TestParseSystemHeader_MissingSection(t *testing.T) {
	_, err := parseSystemHeader([]string{"PartNumber", "Name"})
	if err == nil {
		t.Fatal("expected error for missing Section column")
	}
	if !contains(err.Error(), "missing required column: Section") {
		t.Errorf("error = %q, want containing 'missing required column: Section'", err.Error())
	}
}

// TestParseSystemRow_InvalidPosition verifies a non-numeric Position value is
// rejected.
//
// Why it matters: Position is the device's U slot used for rack placement, so a
// non-integer must fail the row rather than default to an arbitrary slot.
// Inputs: a row whose Position cell is "notanumber". Outputs: a non-nil "invalid
// Position" error.
// Data choice: a clearly non-numeric string isolates the strconv failure from the
// Qty path, which has its own test.
func TestParseSystemRow_InvalidPosition(t *testing.T) {
	idx := systemColumnIndex{section: 0, position: 1, partNumber: -1, name: -1, qty: -1,
		rack: -1, face: -1, role: -1, status: -1, serial: -1, device: -1, bay: -1,
		aDevice: -1, aPort: -1, bDevice: -1, bPort: -1, color: -1, length: -1,
		lengthUnit: -1, location: -1, contentTypes: -1, macAddress: -1}
	_, err := parseSystemRow([]string{"device", "notanumber"}, idx, 2)
	if err == nil {
		t.Fatal("expected error for invalid position")
	}
	if !contains(err.Error(), "invalid Position") {
		t.Errorf("error = %q, want containing 'invalid Position'", err.Error())
	}
}

// TestMergeSystemDefaults verifies each defaultable field is filled from defaults
// only when empty and is otherwise preserved.
//
// Why it matters: defaults backfill sparse system CSV rows without clobbering
// values an operator set explicitly, so every defaultable field must follow the
// fill-if-empty rule.
// Inputs: an all-empty record and a fully-set record, each merged against the
// same defaults. Outputs: the merged records.
// Data choice: the empty record exercises every fill branch; the fully-set record
// exercises every preserve branch, together covering both directions of all six
// fields.
func TestMergeSystemDefaults(t *testing.T) {
	defaults := SystemRecord{
		Status: "Active", Role: "ComputeNode", Face: "front",
		Location: "DC1", Color: "blue", LengthUnit: "m",
	}
	tests := []struct {
		name string
		rec  SystemRecord
		want SystemRecord
	}{
		{
			name: "fills all empty fields from defaults",
			rec:  SystemRecord{},
			want: SystemRecord{Status: "Active", Role: "ComputeNode", Face: "front",
				Location: "DC1", Color: "blue", LengthUnit: "m"},
		},
		{
			name: "preserves all set fields",
			rec: SystemRecord{Status: "Planned", Role: "HSNSwitch", Face: "rear",
				Location: "DC2", Color: "green", LengthUnit: "ft"},
			want: SystemRecord{Status: "Planned", Role: "HSNSwitch", Face: "rear",
				Location: "DC2", Color: "green", LengthUnit: "ft"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeSystemDefaults(tt.rec, defaults); got != tt.want {
				t.Errorf("mergeSystemDefaults() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestApplyDefaults_SectionDefaults verifies section-specific defaults fill fields
// the global defaults left empty, while explicit values win over both.
//
// Why it matters: system CSV supports per-section default rows so, for example,
// every device can inherit a role; these must layer on top of global defaults
// without overriding values a row already carries.
// Inputs: a SystemCSV with a global Status default and a device-section Role/Face
// default, applied to a bare device row and to a device row with an explicit
// Face. Outputs: the merged records.
// Data choice: keeping Status only global and Role/Face only in the section
// proves the section layer is consulted; the explicit Face proves precedence.
func TestApplyDefaults_SectionDefaults(t *testing.T) {
	data := &SystemCSV{
		Defaults: SystemRecord{Status: "Active"},
		SectionDefaults: map[string]SystemRecord{
			"device": {Role: "ComputeNode", Face: "rear"},
		},
	}
	t.Run("section defaults fill fields global left empty", func(t *testing.T) {
		got := data.ApplyDefaults(SystemRecord{Section: "device", Name: "d1"})
		if got.Status != "Active" {
			t.Errorf("Status = %q, want Active (global)", got.Status)
		}
		if got.Role != "ComputeNode" {
			t.Errorf("Role = %q, want ComputeNode (section)", got.Role)
		}
		if got.Face != "rear" {
			t.Errorf("Face = %q, want rear (section)", got.Face)
		}
	})
	t.Run("explicit value beats all defaults", func(t *testing.T) {
		got := data.ApplyDefaults(SystemRecord{Section: "device", Face: "side"})
		if got.Face != "side" {
			t.Errorf("Face = %q, want side (explicit)", got.Face)
		}
	})
}

// TestParseSystemCSV_SectionsAndSkips verifies parsing buckets every known
// section, records global and section defaults, skips unknown sections, and skips
// rows that fail validation.
//
// Why it matters: a real system CSV interleaves defaults, all section types, and
// occasional bad rows, so the parser must route the good rows, capture defaults,
// and drop the rest without aborting the whole import.
// Inputs: a temp CSV with _defaults, device_defaults, one row per section, an
// unknown "mystery" section, and a device row with a non-numeric Qty. Outputs: a
// SystemCSV with one Location, one Device (bad row skipped), and a recorded
// device section default.
// Data choice: exactly one valid Device plus one invalid Device proves the
// skip path without ambiguity, and the lone location row covers the otherwise
// untested Locations bucket.
func TestParseSystemCSV_SectionsAndSkips(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/mixed.csv"
	content := "Section,PartNumber,Name,Qty,Position,Location\n" +
		"_defaults,,,,,DC1\n" +
		"device_defaults,,,,,\n" +
		"location,,DC1,,,\n" +
		"role,,ComputeNode,,,\n" +
		"rack,rack-pn,r1,1,,\n" +
		"device,dev-pn,d1,1,10,\n" +
		"module,mod-pn,,1,,\n" +
		"interface,,iLO,,,\n" +
		"connection,cable-pn,,,,\n" +
		"mystery,,x,,,\n" +
		"device,bad,d2,notanumber,,\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	data, err := ParseSystemCSV(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data.Locations) != 1 {
		t.Errorf("Locations = %d, want 1", len(data.Locations))
	}
	if len(data.Devices) != 1 {
		t.Errorf("Devices = %d, want 1 (bad-qty row skipped)", len(data.Devices))
	}
	if _, ok := data.SectionDefaults["device"]; !ok {
		t.Error("expected device section defaults to be recorded")
	}
	if data.Defaults.Location != "DC1" {
		t.Errorf("global Defaults.Location = %q, want DC1", data.Defaults.Location)
	}
}

// TestParseSystemCSV_TooFewRows verifies a header-only system CSV is rejected.
//
// Why it matters: a file with no data rows carries no inventory and almost
// certainly indicates a truncated or wrong file, so it must error rather than
// silently import nothing.
// Inputs: a temp CSV containing only a header line. Outputs: a non-nil error.
// Data choice: a header with no data rows is the exact boundary the row-count
// guard protects.
func TestParseSystemCSV_TooFewRows(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/headeronly.csv"
	if err := os.WriteFile(path, []byte("Section,Name\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := ParseSystemCSV(path)
	if err == nil {
		t.Fatal("expected error for header-only system CSV")
	}
	if !contains(err.Error(), "header row and at least one data row") {
		t.Errorf("error = %q, want containing 'header row and at least one data row'", err.Error())
	}
}

// TestParseSystemCSV_MissingSectionColumn verifies parsing fails when the header
// has no Section column.
//
// Why it matters: ParseSystemCSV is only reached after header detection, but it
// must still defend its own contract and surface the header parser's error rather
// than mis-parse the rows.
// Inputs: a temp CSV whose header omits Section. Outputs: a non-nil "missing
// required column: Section" error.
// Data choice: a header of unrelated columns is the simplest way to drive the
// header-error return path through ParseSystemCSV.
func TestParseSystemCSV_MissingSectionColumn(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/nosection.csv"
	if err := os.WriteFile(path, []byte("PartNumber,Name\nx,y\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := ParseSystemCSV(path)
	if err == nil {
		t.Fatal("expected error when Section column is missing")
	}
	if !contains(err.Error(), "missing required column: Section") {
		t.Errorf("error = %q, want containing 'missing required column: Section'", err.Error())
	}
}

// TestParseSystemCSV_ReadError verifies a system CSV the reader rejects surfaces a
// wrapped read error.
//
// Why it matters: a corrupt system file with broken quoting cannot be parsed, so
// the import must fail loudly with a wrapped error rather than return partial
// data.
// Inputs: a temp CSV whose data row contains a bare double-quote in an unquoted
// field. Outputs: a non-nil "failed to read system CSV" error.
// Data choice: a bare quote is the simplest input that makes the CSV reader's
// ReadAll fail even with relaxed field counts, isolating the read-error wrap.
func TestParseSystemCSV_ReadError(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/malformed.csv"
	if err := os.WriteFile(path, []byte("Section,Name\nrole,a\"b\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := ParseSystemCSV(path)
	if err == nil {
		t.Fatal("expected error for malformed system CSV quoting")
	}
	if !contains(err.Error(), "failed to read system CSV") {
		t.Errorf("error = %q, want containing 'failed to read system CSV'", err.Error())
	}
}
