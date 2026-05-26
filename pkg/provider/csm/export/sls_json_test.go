package export

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/csm/import"
	"github.com/google/uuid"
)

func TestValidateSLSHardware(t *testing.T) {
	deviceID := uuid.New()

	tests := []struct {
		name        string
		expected    map[string]import_.SlsHardware
		inv         devicetypes.Inventory
		expectedErr bool
	}{
		{
			name: "passing test all staged devices present with valid class",
			expected: map[string]import_.SlsHardware{
				"x9000c1s0b0n0": {
					Xname: "x9000c1s0b0n0",
					Class: "Mountain",
				},
			},
			inv: devicetypes.Inventory{
				Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
					deviceID: {
						ID: deviceID,
						ObjectMeta: devicetypes.ObjectMeta{
							Status: "staged",
							ProviderMetadata: map[string]any{
								"csm": map[string]any{
									"xname": "x9000c1s0b0n0",
								},
							},
						},
					},
				},
			},
			expectedErr: false,
		},
		{
			name:     "failing test staged device missing from expected map",
			expected: map[string]import_.SlsHardware{},
			inv: devicetypes.Inventory{
				Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
					deviceID: {
						ID: deviceID,
						ObjectMeta: devicetypes.ObjectMeta{
							Status: "staged",
							ProviderMetadata: map[string]any{
								"csm": map[string]any{
									"xname": "x9000c1s0b0n0",
								},
							},
						},
					},
				},
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSLSHardware(tt.expected, tt.inv)

			if tt.expectedErr && err == nil {
				t.Error("expected error but got nil")
				return
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestWriteSLSJSON(t *testing.T) {
	tests := []struct {
		name        string
		expected    map[string]import_.SlsHardware
		contains    string
		expectedErr bool
	}{
		{
			name: "passing test writes sorted JSON",
			expected: map[string]import_.SlsHardware{
				"x9000c1": {
					Xname:      "x9000c1",
					Parent:     "x9000",
					Type:       "comptype_chassis",
					TypeString: "Chassis",
					Class:      "Mountain",
				},
				"x9000": {
					Xname:      "x9000",
					Parent:     "s0",
					Type:       "comptype_cabinet",
					TypeString: "Cabinet",
					Class:      "Mountain",
				},
			},
			contains:    `"x9000"`,
			expectedErr: false,
		},
		{
			name:        "failing test empty map produces empty JSON object",
			expected:    map[string]import_.SlsHardware{},
			contains:    "{}",
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := writeSLSJSON(&buf, tt.expected)

			if tt.expectedErr && err == nil {
				t.Error("expected error but got nil")
				return
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !strings.Contains(buf.String(), tt.contains) {
				t.Errorf("output %q does not contain %q", buf.String(), tt.contains)
			}
		})
	}
}
