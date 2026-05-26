package import_

import (
	"encoding/csv"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// newTestDevice creates a CaniDeviceType with an initialised ProviderMetadata map.
func newTestDevice() *devicetypes.CaniDeviceType {
	return &devicetypes.CaniDeviceType{
		ObjectMeta: devicetypes.ObjectMeta{
			ProviderMetadata: make(map[string]any),
		},
	}
}

// newTestInventory creates an Inventory pre-populated with the supplied devices.
func newTestInventory(devices map[uuid.UUID]*devicetypes.CaniDeviceType) *devicetypes.Inventory {
	return &devicetypes.Inventory{
		Devices: devices,
	}
}

func TestImportCSVFromReader(t *testing.T) {
	devID := uuid.New()
	inv := newTestInventory(map[uuid.UUID]*devicetypes.CaniDeviceType{
		devID: newTestDevice(),
	})

	tests := []struct {
		name         string
		csv          string
		wantModified int
		wantTotal    int
		expectErr    bool
	}{
		{
			name:         "updates device from valid CSV",
			csv:          "ID,Role,SubRole,Nid\n" + devID.String() + ",Compute,UAN,42\n",
			wantModified: 1,
			wantTotal:    1,
			expectErr:    false,
		},
		{
			name:      "empty CSV returns error",
			csv:       "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := csv.NewReader(strings.NewReader(tt.csv))
			mod, total, err := importCSVFromReader(reader, inv)

			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectErr {
				if mod != tt.wantModified {
					t.Errorf("modified = %d, want %d", mod, tt.wantModified)
				}
				if total != tt.wantTotal {
					t.Errorf("total = %d, want %d", total, tt.wantTotal)
				}
			}
		})
	}
}

func TestHasIDColumn(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		want    bool
	}{
		{
			name:    "contains ID header",
			headers: []string{"Name", "ID", "Role"},
			want:    true,
		},
		{
			name:    "missing ID header",
			headers: []string{"Name", "Role"},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasIDColumn(tt.headers)
			if got != tt.want {
				t.Errorf("hasIDColumn() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestRowToMap(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		row     []string
		wantKey string
		wantVal string
	}{
		{
			name:    "maps header to normalised key",
			headers: []string{"id", "role"},
			row:     []string{"abc-123", "Compute"},
			wantKey: "Role",
			wantVal: "Compute",
		},
		{
			name:    "row shorter than headers",
			headers: []string{"id", "role"},
			row:     []string{"abc-123"},
			wantKey: "Role",
			wantVal: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := rowToMap(tt.headers, tt.row)
			got := m[tt.wantKey]
			if got != tt.wantVal {
				t.Errorf("rowToMap()[%q] = %q, want %q", tt.wantKey, got, tt.wantVal)
			}
		})
	}
}

func TestNormalizeHeader(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "known header normalises",
			input: " role ",
			want:  "Role",
		},
		{
			name:  "unknown header passes through",
			input: "CustomField",
			want:  "CustomField",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeHeader(tt.input)
			if got != tt.want {
				t.Errorf("normalizeHeader(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSetDeviceFields(t *testing.T) {
	tests := []struct {
		name    string
		values  map[string]string
		want    bool
		checkFn func(t *testing.T, dev *devicetypes.CaniDeviceType)
	}{
		{
			name:   "sets Role and Nid",
			values: map[string]string{"Role": "Compute", "Nid": "42"},
			want:   true,
			checkFn: func(t *testing.T, dev *devicetypes.CaniDeviceType) {
				sub, _ := dev.GetProviderSubMap("csm")
				if sub["role"] != "Compute" {
					t.Errorf("role = %v, want Compute", sub["role"])
				}
				if sub["nid"] != 42 {
					t.Errorf("nid = %v, want 42", sub["nid"])
				}
			},
		},
		{
			name:   "no matching fields returns false",
			values: map[string]string{"Unknown": "value"},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dev := newTestDevice()
			got := setDeviceFields(dev, tt.values)
			if got != tt.want {
				t.Errorf("setDeviceFields() = %t, want %t", got, tt.want)
			}
			if tt.checkFn != nil {
				tt.checkFn(t, dev)
			}
		})
	}
}

func TestSetCSMMetaString(t *testing.T) {
	tests := []struct {
		name    string
		initial string
		value   string
		want    bool
	}{
		{
			name:    "sets new value",
			initial: "",
			value:   "Compute",
			want:    true,
		},
		{
			name:    "same value returns false",
			initial: "Compute",
			value:   "Compute",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dev := newTestDevice()
			if tt.initial != "" {
				dev.SetProviderMeta("csm", "role", tt.initial)
			}
			got := setCSMMetaString(dev, "role", tt.value)
			if got != tt.want {
				t.Errorf("setCSMMetaString() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestSetCSMMetaNid(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "sets valid integer",
			value: "42",
			want:  true,
		},
		{
			name:  "invalid integer returns false",
			value: "not-a-number",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dev := newTestDevice()
			got := setCSMMetaNid(dev, tt.value)
			if got != tt.want {
				t.Errorf("setCSMMetaNid() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestSetCSMMetaAlias(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "sets new alias",
			value: "nid000042",
			want:  true,
		},
		{
			name:  "empty value on nil aliases returns false",
			value: "",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dev := newTestDevice()
			got := setCSMMetaAlias(dev, tt.value)
			if got != tt.want {
				t.Errorf("setCSMMetaAlias() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestSetCSMMetaVlan(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "sets valid vlan",
			value: "100",
			want:  true,
		},
		{
			name:  "invalid vlan returns false",
			value: "abc",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dev := newTestDevice()
			got := setCSMMetaVlan(dev, tt.value)
			if got != tt.want {
				t.Errorf("setCSMMetaVlan() = %t, want %t", got, tt.want)
			}
		})
	}
}
