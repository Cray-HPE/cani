package export

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func TestCapitalizeType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "passing test capitalizes first letter",
			input:    "node",
			expected: "Node",
		},
		{
			name:     "failing test empty string returns empty",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := capitalizeType(tt.input)
			if got != tt.expected {
				t.Errorf("capitalizeType() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestPrintSummary(t *testing.T) {
	deviceID := uuid.New()

	tests := []struct {
		name     string
		inv      devicetypes.Inventory
		stats    reconcileStats
		contains string
	}{
		{
			name: "passing test with staged items",
			inv: devicetypes.Inventory{
				Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
					deviceID: {
						ID:   deviceID,
						Type: devicetypes.TypeNode,
						ObjectMeta: devicetypes.ObjectMeta{
							Status: "staged",
						},
					},
				},
			},
			stats:    reconcileStats{},
			contains: "1 new hardware item(s) are in the inventory",
		},
		{
			name: "failing test with empty inventory",
			inv: devicetypes.Inventory{
				Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
			},
			stats:    reconcileStats{},
			contains: "0 new hardware item(s) are in the inventory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printSummary(&buf, tt.inv, tt.stats)
			got := buf.String()

			if !strings.Contains(got, tt.contains) {
				t.Errorf("printSummary() output %q does not contain %q", got, tt.contains)
			}
			if !strings.Contains(got, "Summary:") {
				t.Errorf("printSummary() output missing header")
			}
		})
	}
}
