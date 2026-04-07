package export

import (
	"encoding/json"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/csm/import"
	"github.com/google/uuid"
)

func TestExtractXname(t *testing.T) {
	tests := []struct {
		name     string
		dev      *devicetypes.CaniDeviceType
		expected string
	}{
		{
			name: "passing test with valid xname",
			dev: &devicetypes.CaniDeviceType{
				ProviderMetadata: map[string]any{
					"csm": map[string]any{
						"xname": "x9000c1s0b0n0",
					},
				},
			},
			expected: "x9000c1s0b0n0",
		},
		{
			name:     "failing test with nil device",
			dev:      nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractXname(tt.dev)
			if got != tt.expected {
				t.Errorf("extractXname() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestCaniStatus(t *testing.T) {
	tests := []struct {
		name     string
		dev      *devicetypes.CaniDeviceType
		expected string
	}{
		{
			name: "passing test staged node returns provisioned",
			dev: &devicetypes.CaniDeviceType{
				Type:   devicetypes.TypeNode,
				Status: "staged",
			},
			expected: "provisioned",
		},
		{
			name: "failing test non-staged node returns empty",
			dev: &devicetypes.CaniDeviceType{
				Type:   devicetypes.TypeNode,
				Status: "active",
			},
			expected: "empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := caniStatus(tt.dev)
			if got != tt.expected {
				t.Errorf("caniStatus() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestInjectCaniMetadata(t *testing.T) {
	tests := []struct {
		name     string
		existing map[string]any
		caniID   string
		status   string
		wantKeys []string
	}{
		{
			name:     "passing test injects all four keys",
			existing: map[string]any{"Role": "Compute"},
			caniID:   "test-id",
			status:   "provisioned",
			wantKeys: []string{"@cani.id", "@cani.lastModified", "@cani.slsSchemaVersion", "@cani.status", "Role"},
		},
		{
			name:     "failing test with nil existing map still injects keys",
			existing: nil,
			caniID:   "test-id",
			status:   "empty",
			wantKeys: []string{"@cani.id", "@cani.lastModified", "@cani.slsSchemaVersion", "@cani.status"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := injectCaniMetadata(tt.existing, tt.caniID, tt.status)

			for _, key := range tt.wantKeys {
				if _, ok := got[key]; !ok {
					t.Errorf("injectCaniMetadata() missing key %q", key)
				}
			}
			if got["@cani.id"] != tt.caniID {
				t.Errorf("@cani.id = %q, want %q", got["@cani.id"], tt.caniID)
			}
			if got["@cani.status"] != tt.status {
				t.Errorf("@cani.status = %q, want %q", got["@cani.status"], tt.status)
			}
		})
	}
}

func TestDeriveParentXname(t *testing.T) {
	tests := []struct {
		name     string
		xname    string
		expected string
	}{
		{
			name:     "passing test multi-component xname",
			xname:    "x9000c1s0b0n2",
			expected: "x9000c1s0b0",
		},
		{
			name:     "failing test single-component xname returns empty prefix",
			xname:    "x9000",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveParentXname(tt.xname)
			if got != tt.expected {
				t.Errorf("deriveParentXname() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGetIntMeta(t *testing.T) {
	tests := []struct {
		name     string
		meta     map[string]any
		key      string
		expected int
		ok       bool
	}{
		{
			name:     "passing test with int value",
			meta:     map[string]any{"nid": 42},
			key:      "nid",
			expected: 42,
			ok:       true,
		},
		{
			name:     "failing test with missing key",
			meta:     map[string]any{},
			key:      "nid",
			expected: 0,
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := getIntMeta(tt.meta, tt.key)
			if got != tt.expected {
				t.Errorf("getIntMeta() value = %d, want %d", got, tt.expected)
			}
			if ok != tt.ok {
				t.Errorf("getIntMeta() ok = %v, want %v", ok, tt.ok)
			}
		})
	}
}

func TestGetStringSliceMeta(t *testing.T) {
	tests := []struct {
		name     string
		meta     map[string]any
		key      string
		expected []string
	}{
		{
			name:     "passing test with string slice",
			meta:     map[string]any{"aliases": []string{"nid001", "nid002"}},
			key:      "aliases",
			expected: []string{"nid001", "nid002"},
		},
		{
			name:     "failing test with missing key",
			meta:     map[string]any{},
			key:      "aliases",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStringSliceMeta(tt.meta, tt.key)

			if len(got) != len(tt.expected) {
				t.Errorf("getStringSliceMeta() returned %d items, want %d", len(got), len(tt.expected))
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("getStringSliceMeta()[%d] = %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestMarshalHardware(t *testing.T) {
	tests := []struct {
		name        string
		hw          import_.SlsHardware
		wantField   string
		expectedErr bool
	}{
		{
			name: "passing test produces valid JSON",
			hw: import_.SlsHardware{
				Xname:      "x9000c1",
				Parent:     "s0",
				Type:       "comptype_cabinet",
				TypeString: "Cabinet",
				Class:      "Mountain",
			},
			wantField:   "x9000c1",
			expectedErr: false,
		},
		{
			name: "failing test still produces JSON for minimal entry",
			hw: import_.SlsHardware{
				Xname: "",
			},
			wantField:   "",
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := marshalHardware(tt.hw)

			if tt.expectedErr && err == nil {
				t.Error("expected error but got nil")
				return
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			var result map[string]any
			if err := json.Unmarshal(data, &result); err != nil {
				t.Errorf("marshalHardware() produced invalid JSON: %v", err)
				return
			}
			if xname, _ := result["Xname"].(string); xname != tt.wantField {
				t.Errorf("Xname = %q, want %q", xname, tt.wantField)
			}
		})
	}
}

func TestBuildNewNodeEntry(t *testing.T) {
	tests := []struct {
		name      string
		dev       *devicetypes.CaniDeviceType
		xname     string
		wantType  string
		wantClass string
	}{
		{
			name: "passing test builds node entry with class",
			dev: &devicetypes.CaniDeviceType{
				Type:   devicetypes.TypeNode,
				Status: "staged",
				Role:   "Compute",
				ProviderMetadata: map[string]any{
					"csm": map[string]any{
						"xname":   "x9000c1s0b0n0",
						"class":   "Mountain",
						"nid":     42,
						"role":    "Compute",
						"aliases": []string{"nid001"},
					},
				},
			},
			xname:     "x9000c1s0b0n0",
			wantType:  "comptype_node",
			wantClass: "Mountain",
		},
		{
			name: "failing test builds entry with empty class",
			dev: &devicetypes.CaniDeviceType{
				Type:   devicetypes.TypeNode,
				Status: "staged",
				ProviderMetadata: map[string]any{
					"csm": map[string]any{
						"xname": "x9000c1s0b0n0",
					},
				},
			},
			xname:     "x9000c1s0b0n0",
			wantType:  "comptype_node",
			wantClass: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hw := buildNewNodeEntry(tt.dev, tt.xname)

			if hw.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", hw.Type, tt.wantType)
			}
			if hw.Class != tt.wantClass {
				t.Errorf("Class = %q, want %q", hw.Class, tt.wantClass)
			}
			if hw.Xname != tt.xname {
				t.Errorf("Xname = %q, want %q", hw.Xname, tt.xname)
			}
		})
	}
}

func TestBuildNewCabinetEntry(t *testing.T) {
	tests := []struct {
		name      string
		dev       *devicetypes.CaniDeviceType
		xname     string
		wantType  string
		wantClass string
	}{
		{
			name: "passing test builds cabinet entry",
			dev: &devicetypes.CaniDeviceType{
				Type:   devicetypes.TypeCabinet,
				Status: "staged",
				ProviderMetadata: map[string]any{
					"csm": map[string]any{
						"xname": "x9000",
						"class": "Mountain",
					},
				},
			},
			xname:     "x9000",
			wantType:  "comptype_cabinet",
			wantClass: "Mountain",
		},
		{
			name: "failing test builds cabinet with empty class",
			dev: &devicetypes.CaniDeviceType{
				Type:   devicetypes.TypeCabinet,
				Status: "staged",
				ProviderMetadata: map[string]any{
					"csm": map[string]any{
						"xname": "x9000",
					},
				},
			},
			xname:     "x9000",
			wantType:  "comptype_cabinet",
			wantClass: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hw := buildNewCabinetEntry(tt.dev, tt.xname)

			if hw.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", hw.Type, tt.wantType)
			}
			if hw.Class != tt.wantClass {
				t.Errorf("Class = %q, want %q", hw.Class, tt.wantClass)
			}
			if hw.Parent != "s0" {
				t.Errorf("Parent = %q, want %q", hw.Parent, "s0")
			}
		})
	}
}

func TestFindParentDevice(t *testing.T) {
	deviceID := uuid.New()

	tests := []struct {
		name     string
		xname    string
		inv      devicetypes.Inventory
		wantNil  bool
	}{
		{
			name:  "passing test finds parent by xname prefix",
			xname: "x9000c1b0",
			inv: devicetypes.Inventory{
				Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
					deviceID: {
						ID: deviceID,
						ProviderMetadata: map[string]any{
							"csm": map[string]any{
								"xname": "x9000c1",
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			name:  "failing test no matching parent",
			xname: "x9000c1b0",
			inv: devicetypes.Inventory{
				Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findParentDevice(tt.xname, tt.inv)
			if tt.wantNil && got != nil {
				t.Error("expected nil but got a device")
			}
			if !tt.wantNil && got == nil {
				t.Error("expected a device but got nil")
			}
		})
	}
}
