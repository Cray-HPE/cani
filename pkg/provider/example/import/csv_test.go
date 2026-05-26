package import_

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseCSV(t *testing.T) {

	good_cable_response := []CsvRecord{
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
			expected:  good_cable_response,
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

			for i, row := range got {
				if row != tt.expected[i] {

					t.Errorf("parseCSV() = \n%v, wanted \n%v", row, tt.expected[i])
				}

			}

		})
	}
}

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
