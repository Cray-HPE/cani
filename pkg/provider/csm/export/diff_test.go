package export

import (
	"testing"

	import_ "github.com/Cray-HPE/cani/pkg/provider/csm/import"
)

func TestDiffHardware(t *testing.T) {
	tests := []struct {
		name        string
		expected    map[string]import_.SlsHardware
		current     map[string]import_.SlsHardware
		wantAdded   int
		wantChanged int
	}{
		{
			name: "passing test adds new entry with CANI metadata",
			expected: map[string]import_.SlsHardware{
				"x9000c1s0b0n0": {
					Xname: "x9000c1s0b0n0",
					Class: "Mountain",
					ExtraProperties: map[string]any{
						"@cani.slsSchemaVersion": "v1alpha1",
						"@cani.status":           "provisioned",
					},
				},
			},
			current:     map[string]import_.SlsHardware{},
			wantAdded:   1,
			wantChanged: 0,
		},
		{
			name: "failing test skips entry without CANI metadata",
			expected: map[string]import_.SlsHardware{
				"x9000c1": {
					Xname:           "x9000c1",
					Class:           "Mountain",
					ExtraProperties: map[string]any{},
				},
			},
			current:     map[string]import_.SlsHardware{},
			wantAdded:   0,
			wantChanged: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := diffHardware(tt.expected, tt.current)

			if len(changes.Added) != tt.wantAdded {
				t.Errorf("Added = %d, want %d", len(changes.Added), tt.wantAdded)
			}
			if len(changes.Changed) != tt.wantChanged {
				t.Errorf("Changed = %d, want %d", len(changes.Changed), tt.wantChanged)
			}
		})
	}
}

func TestHasCaniMetadata(t *testing.T) {
	tests := []struct {
		name     string
		hw       import_.SlsHardware
		expected bool
	}{
		{
			name: "passing test with CANI metadata",
			hw: import_.SlsHardware{
				ExtraProperties: map[string]any{
					"@cani.slsSchemaVersion": "v1alpha1",
				},
			},
			expected: true,
		},
		{
			name: "failing test without CANI metadata",
			hw: import_.SlsHardware{
				ExtraProperties: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasCaniMetadata(tt.hw)
			if got != tt.expected {
				t.Errorf("hasCaniMetadata() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHardwareNeedsUpdate(t *testing.T) {
	tests := []struct {
		name     string
		current  import_.SlsHardware
		expected import_.SlsHardware
		want     bool
	}{
		{
			name: "passing test current lacks CANI metadata",
			current: import_.SlsHardware{
				ExtraProperties: map[string]any{},
			},
			expected: import_.SlsHardware{
				ExtraProperties: map[string]any{
					"@cani.slsSchemaVersion": "v1alpha1",
					"@cani.status":           "provisioned",
				},
			},
			want: true,
		},
		{
			name: "failing test matching status needs no update",
			current: import_.SlsHardware{
				ExtraProperties: map[string]any{
					"@cani.slsSchemaVersion": "v1alpha1",
					"@cani.status":           "provisioned",
				},
			},
			expected: import_.SlsHardware{
				ExtraProperties: map[string]any{
					"@cani.slsSchemaVersion": "v1alpha1",
					"@cani.status":           "provisioned",
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hardwareNeedsUpdate(tt.current, tt.expected)
			if got != tt.want {
				t.Errorf("hardwareNeedsUpdate() = %v, want %v", got, tt.want)
			}
		})
	}
}
