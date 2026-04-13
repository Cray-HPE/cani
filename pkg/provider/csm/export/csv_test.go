package export

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func TestExportCSV(t *testing.T) {
	deviceID := uuid.New()

	tests := []struct {
		name        string
		inv         devicetypes.Inventory
		headers     []string
		types       []string
		contains    string
		expectedErr bool
	}{
		{
			name: "passing test with valid headers and types",
			inv: devicetypes.Inventory{
				Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
					deviceID: {
						ID:   deviceID,
						Name: "nid001",
						Type: devicetypes.TypeNode,
						ObjectMeta: devicetypes.ObjectMeta{
							ProviderMetadata: map[string]any{
								"csm": map[string]any{
									"role": "Compute",
								},
							},
						},
					},
				},
			},
			headers:     []string{"Name", "Type", "Role"},
			types:       []string{"node"},
			contains:    "nid001,Node,Compute",
			expectedErr: false,
		},
		{
			name: "failing test with invalid header",
			inv: devicetypes.Inventory{
				Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
			},
			headers:     []string{"BadHeader"},
			types:       []string{},
			contains:    "",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := ExportCSV(&buf, tt.inv, tt.headers, tt.types)

			if tt.expectedErr && err == nil {
				t.Error("expected error but got nil")
				return
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if tt.contains != "" && !strings.Contains(buf.String(), tt.contains) {
				t.Errorf("output %q does not contain %q", buf.String(), tt.contains)
			}
		})
	}
}

func TestNormalizeHeaders(t *testing.T) {
	tests := []struct {
		name        string
		headers     []string
		expected    []string
		expectedErr bool
	}{
		{
			name:        "passing test with valid headers",
			headers:     []string{"name", "type", "status"},
			expected:    []string{"Name", "Type", "Status"},
			expectedErr: false,
		},
		{
			name:        "failing test with invalid header",
			headers:     []string{"name", "badheader"},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeHeaders(tt.headers)

			if tt.expectedErr && err == nil {
				t.Error("expected error but got nil")
				return
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !tt.expectedErr {
				for i, h := range got {
					if h != tt.expected[i] {
						t.Errorf("header[%d] = %q, want %q", i, h, tt.expected[i])
					}
				}
			}
		})
	}
}

func TestBuildTypeSet(t *testing.T) {
	tests := []struct {
		name     string
		types    []string
		expected map[string]struct{}
	}{
		{
			name:  "passing test with known aliases",
			types: []string{"nodeblade", "cabinet"},
			expected: map[string]struct{}{
				string(devicetypes.TypeNodeCard): {},
				string(devicetypes.TypeCabinet):  {},
			},
		},
		{
			name:     "failing test with empty input",
			types:    []string{},
			expected: map[string]struct{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTypeSet(tt.types)

			if len(got) != len(tt.expected) {
				t.Errorf("buildTypeSet() returned %d entries, want %d", len(got), len(tt.expected))
				return
			}
			for k := range tt.expected {
				if _, ok := got[k]; !ok {
					t.Errorf("buildTypeSet() missing key %q", k)
				}
			}
		})
	}
}

func TestMatchesType(t *testing.T) {
	tests := []struct {
		name     string
		dev      *devicetypes.CaniDeviceType
		typeSet  map[string]struct{}
		expected bool
	}{
		{
			name:     "passing test with matching type",
			dev:      &devicetypes.CaniDeviceType{Type: devicetypes.TypeNode},
			typeSet:  map[string]struct{}{string(devicetypes.TypeNode): {}},
			expected: true,
		},
		{
			name:     "failing test with non-matching type",
			dev:      &devicetypes.CaniDeviceType{Type: devicetypes.TypeCabinet},
			typeSet:  map[string]struct{}{string(devicetypes.TypeNode): {}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesType(tt.dev, tt.typeSet)
			if got != tt.expected {
				t.Errorf("matchesType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCanonicalTypeName(t *testing.T) {
	tests := []struct {
		name     string
		typeName devicetypes.Type
		expected string
	}{
		{
			name:     "passing test with known type",
			typeName: devicetypes.TypeNode,
			expected: "Node",
		},
		{
			name:     "failing test with empty type",
			typeName: devicetypes.Type(""),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := canonicalTypeName(tt.typeName)
			if got != tt.expected {
				t.Errorf("canonicalTypeName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGetField(t *testing.T) {
	deviceID := uuid.New()

	tests := []struct {
		name     string
		dev      *devicetypes.CaniDeviceType
		header   string
		expected string
	}{
		{
			name: "passing test with Name header",
			dev: &devicetypes.CaniDeviceType{
				ID:   deviceID,
				Name: "nid001",
			},
			header:   "Name",
			expected: "nid001",
		},
		{
			name:     "failing test with unknown header",
			dev:      &devicetypes.CaniDeviceType{},
			header:   "Unknown",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getField(tt.dev, tt.header)
			if got != tt.expected {
				t.Errorf("getField() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGetCSMMetaString(t *testing.T) {
	tests := []struct {
		name     string
		dev      *devicetypes.CaniDeviceType
		key      string
		expected string
	}{
		{
			name: "passing test with valid key",
			dev: &devicetypes.CaniDeviceType{
				ProviderMetadata: map[string]any{
					"csm": map[string]any{
						"role": "Compute",
					},
				},
			},
			key:      "role",
			expected: "Compute",
		},
		{
			name:     "failing test with no provider metadata",
			dev:      &devicetypes.CaniDeviceType{},
			key:      "role",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCSMMetaString(tt.dev, tt.key)
			if got != tt.expected {
				t.Errorf("getCSMMetaString() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGetCSMMetaFirstAlias(t *testing.T) {
	tests := []struct {
		name     string
		dev      *devicetypes.CaniDeviceType
		expected string
	}{
		{
			name: "passing test with string slice aliases",
			dev: &devicetypes.CaniDeviceType{
				ObjectMeta: devicetypes.ObjectMeta{
					ProviderMetadata: map[string]any{
						"csm": map[string]any{
							"aliases": []string{"nid001", "nid002"},
						},
					},
				},
			},
			expected: "nid001",
		},
		{
			name:     "failing test with no aliases",
			dev:      &devicetypes.CaniDeviceType{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCSMMetaFirstAlias(tt.dev)
			if got != tt.expected {
				t.Errorf("getCSMMetaFirstAlias() = %q, want %q", got, tt.expected)
			}
		})
	}
}
