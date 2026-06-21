package import_

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestParseCSV verifies ParseCSV reads BOM CSV files, returns parsed records,
// and surfaces file/header errors.
//
// Why it matters: BOM CSV import depends on ParseCSV to preserve row identity and
// reject unusable input before records reach the provider.
// Inputs: the cables fixture, a missing path, a fixture missing a required
// column, and a header-only fixture. Outputs: parsed CsvRecord slices or wrapped
// errors.
// Data choice: the cables fixture has five known records, and each error fixture
// isolates one parser failure path.
func TestParseCSV(t *testing.T) {

	goodCableResponse := []CsvRecord{
		{PartNumber: "P9K58A", Description: "HPE 48U 800mmx1200mm G2 Enterprise Shock Rack", Quantity: 2, ConfigGroup: "0100"},
		{PartNumber: "R9G23A", Description: "HPE Aruba Networking 8360-48Y6C v2 Power-to-Port Airflow 5 Fans 2 PSU Attached Bundle", Quantity: 2, ConfigGroup: "0200"},
		{PartNumber: "P67287-B21", Description: "XD670", Quantity: 6, ConfigGroup: "0300"},
		{PartNumber: "R0Z28A", Description: "HPE 400G QSFP-DD DAC 3m", Quantity: 1, ConfigGroup: "0900"},
		{PartNumber: "C7536A", Description: "HPE Cat6 RJ45 M/M 2m", Quantity: 6, ConfigGroup: "0900"},
	}
	tests := []struct {
		name      string
		path      string
		expected  []CsvRecord
		expectErr bool
		errorMsg  string
	}{
		{
			name:      "Test good data in cables.csv",
			path:      "../../../../testdata/fixtures/example/cables.csv",
			expectErr: false,
			expected:  goodCableResponse,
			errorMsg:  "",
		},
		{
			name:      "Bad path",
			path:      "bad_path.csv",
			expectErr: true,
			expected:  nil,
			errorMsg:  "failed to open CSV file",
		},
		{
			name:      "Missing required column",
			path:      "../../../../testdata/fixtures/example/missing_column.csv",
			expectErr: true,
			expected:  nil,
			errorMsg:  "missing required column: Description (or Name, ProductDescription)",
		},
		{
			name:      "Empty file",
			path:      "../../../../testdata/fixtures/example/empty.csv",
			expectErr: true,
			expected:  nil,
			errorMsg:  "CSV must have a header row and at least one data row",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCSV(tt.path)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q but got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Error: %e", err)
			}
			if len(got) != len(tt.expected) {
				t.Fatalf("ParseCSV() returned %d records, want %d", len(got), len(tt.expected))
			}

			for i, row := range got {
				if row != tt.expected[i] {

					t.Errorf("parseCSV() = \n%v, wanted \n%v", row, tt.expected[i])
				}

			}

		})
	}
}

// TestParseHeader verifies parseHeader recognizes required and optional BOM CSV
// columns, including aliases, and errors when required columns are missing.
//
// Why it matters: header parsing defines how every later row cell is interpreted,
// so aliases and missing-column errors must be deterministic.
// Inputs: standard headers, alias headers, optional cable headers, and headers
// missing PartNumber, Description, or Quantity. Outputs: column indexes or
// errors.
// Data choice: the table uses the smallest headers that prove each required
// index and each missing-column guard.
func TestParseHeader(t *testing.T) {
	tests := []struct {
		name      string
		header    []string
		expectErr bool
		errorMsg  string
		wantPN    int
		wantDesc  int
		wantQty   int
		wantCfg   int
	}{
		{
			name:     "standard headers",
			header:   []string{"PartNumber", "Description", "Quantity"},
			wantPN:   0,
			wantDesc: 1,
			wantQty:  2,
			wantCfg:  -1,
		},
		{
			name:     "alias headers",
			header:   []string{"ProductNumber", "Name", "Qty", "ConfigGroup"},
			wantPN:   0,
			wantDesc: 1,
			wantQty:  2,
			wantCfg:  3,
		},
		{
			name:     "all optional cable columns",
			header:   []string{"PartNumber", "Description", "Quantity", "SourceDevice", "SourcePort", "DestDevice", "DestPort", "CableType", "CableLength"},
			wantPN:   0,
			wantDesc: 1,
			wantQty:  2,
			wantCfg:  -1,
		},
		{
			name:      "missing PartNumber",
			header:    []string{"Description", "Quantity"},
			expectErr: true,
			errorMsg:  "missing required column: PartNumber",
		},
		{
			name:      "missing Description",
			header:    []string{"PartNumber", "Quantity"},
			expectErr: true,
			errorMsg:  "missing required column: Description",
		},
		{
			name:      "missing Quantity",
			header:    []string{"PartNumber", "Description"},
			expectErr: true,
			errorMsg:  "missing required column: Quantity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx, err := parseHeader(tt.header)
			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("error = %q, want containing %q", err.Error(), tt.errorMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if idx.partNumber != tt.wantPN {
				t.Errorf("partNumber index = %d, want %d", idx.partNumber, tt.wantPN)
			}
			if idx.description != tt.wantDesc {
				t.Errorf("description index = %d, want %d", idx.description, tt.wantDesc)
			}
			if idx.quantity != tt.wantQty {
				t.Errorf("quantity index = %d, want %d", idx.quantity, tt.wantQty)
			}
			if idx.configGroup != tt.wantCfg {
				t.Errorf("configGroup index = %d, want %d", idx.configGroup, tt.wantCfg)
			}
		})
	}
}

// TestNormalizeHeader verifies BOM CSV header aliases normalize to canonical
// parser keys.
//
// Why it matters: exported BOMs use several naming conventions, and import must
// route them to the same CsvRecord fields.
// Inputs: aliases for product, description, quantity, config, cable endpoint,
// type, and length fields plus unknown and separator-heavy headers. Outputs:
// normalized header strings.
// Data choice: representative aliases and whitespace/separator variants prove
// both synonym mapping and cleanup behavior.
func TestNormalizeHeader(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// PartNumber aliases
		{"PartNumber", "partnumber"},
		{"ProductNumber", "partnumber"},
		{"HPEProductNumber", "partnumber"},
		{"Part", "partnumber"},
		{"Product", "partnumber"},
		{"part_number", "partnumber"},
		{"Part-Number", "partnumber"},
		// Description aliases
		{"Description", "description"},
		{"Name", "description"},
		{"ProductDescription", "description"},
		{"ProductName", "description"},
		{"Desc", "description"},
		// Quantity aliases
		{"Quantity", "quantity"},
		{"Qty", "quantity"},
		{"Count", "quantity"},
		{"OrderQuantity", "quantity"},
		// ConfigGroup aliases
		{"ConfigGroup", "configgroup"},
		{"ConfigItemNumber", "configgroup"},
		{"Config", "configgroup"},
		{"Group", "configgroup"},
		// Cable field aliases
		{"SourceDevice", "sourcedevice"},
		{"SrcDevice", "sourcedevice"},
		{"DeviceA", "sourcedevice"},
		{"SourcePort", "sourceport"},
		{"SrcPort", "sourceport"},
		{"PortA", "sourceport"},
		{"DestDevice", "destdevice"},
		{"DestinationDevice", "destdevice"},
		{"DeviceB", "destdevice"},
		{"DestPort", "destport"},
		{"DestinationPort", "destport"},
		{"PortB", "destport"},
		{"CableType", "cabletype"},
		{"Type", "cabletype"},
		{"CableLength", "cablelength"},
		{"Length", "cablelength"},
		// Unknown returns lowercase
		{"SomethingElse", "somethingelse"},
		// Whitespace and separators stripped
		{" Part Number ", "partnumber"},
		{"part-number", "partnumber"},
		{"part_number", "partnumber"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeHeader(tt.input)
			if got != tt.want {
				t.Errorf("normalizeHeader(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestParseRow verifies parseRow converts a data row into CsvRecord fields and
// rejects malformed row values.
//
// Why it matters: row parsing is the last validation step before raw records are
// stored for transform, so invalid rows must be skipped and valid optional cable
// fields must survive.
// Inputs: valid rows with and without config groups, malformed rows, and a row
// with explicit cable endpoint fields. Outputs: CsvRecord values or errors.
// Data choice: each malformed row trips one guard, while the cable row proves all
// optional endpoint columns are copied.
func TestParseRow(t *testing.T) {
	baseIdx := columnIndex{
		partNumber:   0,
		description:  1,
		quantity:     2,
		configGroup:  3,
		sourceDevice: -1,
		sourcePort:   -1,
		destDevice:   -1,
		destPort:     -1,
		cableType:    -1,
		cableLength:  -1,
	}

	tests := []struct {
		name      string
		row       []string
		idx       columnIndex
		expectErr bool
		errorMsg  string
		expected  CsvRecord
	}{
		{
			name: "valid row with config group",
			row:  []string{"P9K58A", "HPE Rack", "2", "0100"},
			idx:  baseIdx,
			expected: CsvRecord{
				PartNumber:  "P9K58A",
				Description: "HPE Rack",
				Quantity:    2,
				ConfigGroup: "0100",
			},
		},
		{
			name: "valid row without config group",
			row:  []string{"P9K58A", "HPE Rack", "1"},
			idx: columnIndex{
				partNumber:   0,
				description:  1,
				quantity:     2,
				configGroup:  -1,
				sourceDevice: -1,
				sourcePort:   -1,
				destDevice:   -1,
				destPort:     -1,
				cableType:    -1,
				cableLength:  -1,
			},
			expected: CsvRecord{
				PartNumber:  "P9K58A",
				Description: "HPE Rack",
				Quantity:    1,
			},
		},
		{
			name:      "insufficient columns",
			row:       []string{"P9K58A"},
			idx:       baseIdx,
			expectErr: true,
			errorMsg:  "insufficient columns",
		},
		{
			name:      "empty PartNumber",
			row:       []string{"", "HPE Rack", "2", "0100"},
			idx:       baseIdx,
			expectErr: true,
			errorMsg:  "empty PartNumber",
		},
		{
			name:      "empty Description",
			row:       []string{"P9K58A", "", "2", "0100"},
			idx:       baseIdx,
			expectErr: true,
			errorMsg:  "empty Description",
		},
		{
			name:      "invalid Quantity",
			row:       []string{"P9K58A", "HPE Rack", "abc", "0100"},
			idx:       baseIdx,
			expectErr: true,
			errorMsg:  "invalid Quantity",
		},
		{
			name:      "zero Quantity",
			row:       []string{"P9K58A", "HPE Rack", "0", "0100"},
			idx:       baseIdx,
			expectErr: true,
			errorMsg:  "Quantity must be >= 1",
		},
		{
			name: "with cable fields",
			row:  []string{"C7536A", "Cat6 Cable", "1", "", "switch-01", "eth0", "server-01", "eth1", "cat6", "2m"},
			idx: columnIndex{
				partNumber:   0,
				description:  1,
				quantity:     2,
				configGroup:  3,
				sourceDevice: 4,
				sourcePort:   5,
				destDevice:   6,
				destPort:     7,
				cableType:    8,
				cableLength:  9,
			},
			expected: CsvRecord{
				PartNumber:   "C7536A",
				Description:  "Cat6 Cable",
				Quantity:     1,
				SourceDevice: "switch-01",
				SourcePort:   "eth0",
				DestDevice:   "server-01",
				DestPort:     "eth1",
				CableType:    "cat6",
				CableLength:  "2m",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRow(tt.row, tt.idx, 2)
			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("error = %q, want containing %q", err.Error(), tt.errorMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("parseRow() = %+v, want %+v", got, tt.expected)
			}
		})
	}
}

// TestGetColumnValue verifies getColumnValue trims an in-range cell and returns
// empty for absent optional columns.
//
// Why it matters: optional BOM columns should not panic or leak whitespace when
// a row is shorter than the header map.
// Inputs: rows with valid, padded, out-of-range, and negative indexes. Outputs:
// the extracted string.
// Data choice: the cases cover the helper's two guard branches and its trim
// behavior with minimal row data.
func TestGetColumnValue(t *testing.T) {
	tests := []struct {
		name     string
		row      []string
		idx      int
		expected string
	}{
		{
			name:     "Get index 0",
			row:      []string{"R0Z28A", "HPE 400G QSFP-DD DAC 3m"},
			idx:      0,
			expected: "R0Z28A",
		},
		{
			name:     "trim spaces",
			row:      []string{"R0Z28A", " HPE 400G QSFP-DD DAC 3m "},
			idx:      1,
			expected: "HPE 400G QSFP-DD DAC 3m",
		},
		{
			name:     "index out of range",
			row:      []string{"R0Z28A", "HPE 400G QSFP-DD DAC 3m"},
			idx:      3,
			expected: "",
		},
		{
			name:     "negative index",
			row:      []string{"R0Z28A"},
			idx:      -1,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getColumnValue(tt.row, tt.idx)
			if got != tt.expected {
				t.Errorf("getColumnValue() = %q, want %q", got, tt.expected)
			}
		})
	}

}

// TestIsCableRecord verifies IsCableRecord classifies only records with both
// source and destination devices as explicit cable records.
//
// Why it matters: cable records are routed away from device/rack transform logic,
// so partial endpoint data must not be misclassified.
// Inputs: records with both endpoints, one endpoint, or no endpoints. Outputs:
// boolean classification results.
// Data choice: the table isolates each endpoint presence combination.
func TestIsCableRecord(t *testing.T) {
	tests := []struct {
		name     string
		record   CsvRecord
		expected bool
	}{
		{
			name: "SourceDevice and DestDevice set",
			record: CsvRecord{
				SourceDevice: "switch01",
				DestDevice:   "server01",
			},
			expected: true,
		},
		{
			name: "Only SourceDevice set",
			record: CsvRecord{
				SourceDevice: "switch01",
			},
			expected: false,
		},
		{
			name: "Only DestDevice set",
			record: CsvRecord{
				DestDevice: "server01",
			},
			expected: false,
		},
		{
			name:     "Empty CsvRecord",
			record:   CsvRecord{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCableRecord(tt.record)
			if got != tt.expected {
				t.Errorf("IsCableRecord() = %t, want %t", got, tt.expected)
			}
		})
	}
}

// TestIsConfigParent verifies isConfigParent recognizes top-level config groups
// ending in 00.
//
// Why it matters: config groups drive rack/device hierarchy, so parent groups
// must be detected predictably.
// Inputs: a parent group, a child group, and a too-short group. Outputs: boolean
// parent classification results.
// Data choice: 0200 and 0315 model the normal parent/child pattern; "1" covers
// the length guard.
func TestIsConfigParent(t *testing.T) {
	tests := []struct {
		configGroup string
		expected    bool
	}{
		{
			configGroup: "0200",
			expected:    true,
		},
		{
			configGroup: "0315",
			expected:    false,
		},
		{
			configGroup: "1",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s->%t", tt.configGroup, tt.expected), func(t *testing.T) {
			got := isConfigParent(tt.configGroup)
			if got != tt.expected {
				t.Errorf("isConfigParent() = %t, want %t", got, tt.expected)
			}
		})
	}

}

// TestGetConfigParentGroup verifies getConfigParentGroup maps child config
// groups to their parent group names.
//
// Why it matters: transform uses this relationship to parent devices under the
// correct rack group.
// Inputs: a parent group, a child group, and a too-short group. Outputs: the
// derived parent group string.
// Data choice: 0200 proves an existing parent remains stable, 0315 proves suffix
// replacement, and "1" proves the guard path.
func TestGetConfigParentGroup(t *testing.T) {
	tests := []struct {
		configGroup string
		expected    string
	}{
		{
			configGroup: "0200",
			expected:    "0200",
		},
		{
			configGroup: "0315",
			expected:    "0300",
		},
		{
			configGroup: "1",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s->%s", tt.configGroup, tt.expected), func(t *testing.T) {
			got := getConfigParentGroup(tt.configGroup)
			if got != tt.expected {
				t.Errorf("getConfigParentGroup() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestGetConfigGroupPrefix verifies getConfigGroupPrefix extracts the two-digit
// config group prefix when present.
//
// Why it matters: prefix comparisons decide which groups are racks, devices, and
// related cable groups.
// Inputs: three-character, four-character, and too-short group strings. Outputs:
// the two-character prefix or empty string.
// Data choice: 020 and 0315 prove the helper only needs the first two
// characters, while "1" proves the length guard.
func TestGetConfigGroupPrefix(t *testing.T) {
	tests := []struct {
		configGroup string
		expected    string
	}{
		{
			configGroup: "020",
			expected:    "02",
		},
		{
			configGroup: "0315",
			expected:    "03",
		},
		{
			configGroup: "1",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s->%s", tt.configGroup, tt.expected), func(t *testing.T) {
			got := getConfigGroupPrefix(tt.configGroup)
			if got != tt.expected {
				t.Errorf("getConfigGroupPrefix() = %q, want %q", got, tt.expected)
			}
		})
	}

}

// TestParseCSV_SkipsMalformedRows verifies ParseCSV drops rows that fail row
// validation while keeping the valid ones.
//
// Why it matters: a BOM CSV exported by hand often has stray bad rows, so the
// parser must skip them with a warning and still return the usable records rather
// than failing the whole import.
// Inputs: a temp CSV with one valid row, one empty-PartNumber row, and one
// zero-Quantity row. Outputs: a single CsvRecord for the valid row, nil error.
// Data choice: the two bad rows trip the empty-PartNumber and Quantity<1 guards
// respectively, proving both skip branches; the lone good row makes the surviving
// record unambiguous.
func TestParseCSV_SkipsMalformedRows(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/mixed.csv"
	content := "PartNumber,Description,Quantity\n" +
		"P1,Good Widget,2\n" +
		",Missing PN,1\n" +
		"P3,Another,0\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := ParseCSV(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 valid record, got %d", len(got))
	}
	if got[0].PartNumber != "P1" || got[0].Quantity != 2 {
		t.Errorf("record = %+v, want PartNumber P1 Quantity 2", got[0])
	}
}

// TestParseCSV_ReadError verifies a CSV the encoding/csv reader rejects surfaces
// a wrapped read error.
//
// Why it matters: a corrupt export with broken quoting cannot be parsed at all,
// so the import must fail loudly with a wrapped error rather than return partial
// data.
// Inputs: a temp CSV whose data row contains a bare double-quote in an unquoted
// field. Outputs: a non-nil "failed to read CSV" error.
// Data choice: a bare quote is the simplest input that makes the standard CSV
// reader's ReadAll fail, isolating the read-error wrap.
func TestParseCSV_ReadError(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/malformed.csv"
	if err := os.WriteFile(path, []byte("PartNumber,Description,Quantity\na\"b,c,1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := ParseCSV(path)
	if err == nil {
		t.Fatal("expected error for malformed CSV quoting")
	}
	if !strings.Contains(err.Error(), "failed to read CSV") {
		t.Errorf("error = %q, want containing 'failed to read CSV'", err.Error())
	}
}
