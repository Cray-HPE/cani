package import_

import (
	"testing"
)

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

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
