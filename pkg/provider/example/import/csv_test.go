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

func TestParseHeader(t *testing.T) {}

func TestNormalizeHeader(t *testing.T) {}

func TestParseRow(t *testing.T) {}

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
