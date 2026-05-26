package transform

import (
	"testing"

	"github.com/google/uuid"
)

func TestParseGeoloc(t *testing.T) {
	tests := []struct {
		xname       string
		wantValid   bool
		wantCabinet int
		wantChassis int
	}{
		{"x9000c1s7b0n0", true, 9000, 1},
		{"x9000c1s4b0n0", true, 9000, 1},
		{"x9000c3", true, 9000, 3},
		{"x3000c0s9b0n0", true, 3000, 0},
		{"x3000c0", true, 3000, 0},
		{"x3000", false, 0, 0},
		{"", false, 0, 0},
		{"antero001", false, 0, 0},
		{"service20900014000", false, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.xname, func(t *testing.T) {
			info := ParseGeoloc(tt.xname)
			if info.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", info.Valid, tt.wantValid)
			}
			if info.Cabinet != tt.wantCabinet {
				t.Errorf("Cabinet = %d, want %d", info.Cabinet, tt.wantCabinet)
			}
			if info.Chassis != tt.wantChassis {
				t.Errorf("Chassis = %d, want %d", info.Chassis, tt.wantChassis)
			}
			if info.Raw != tt.xname {
				t.Errorf("Raw = %q, want %q", info.Raw, tt.xname)
			}
		})
	}
}

func TestParentChassisXname(t *testing.T) {
	tests := []struct {
		xname string
		want  string
	}{
		{"x9000c1s7b0n0", "x9000c1"},
		{"x9000c1s4b0n0", "x9000c1"},
		{"x9000c3s1b0n2", "x9000c3"},
		{"x3000c0s9b0n0", "x3000c0"},
		{"x9000c1", "x9000c1"},
		{"x3000", ""},
		{"", ""},
		{"antero001", ""},
	}
	for _, tt := range tests {
		t.Run(tt.xname, func(t *testing.T) {
			got := ParentChassisXname(tt.xname)
			if got != tt.want {
				t.Errorf("ParentChassisXname(%q) = %q, want %q", tt.xname, got, tt.want)
			}
		})
	}
}

func TestNodeGeolocXname(t *testing.T) {
	tests := []struct {
		name      string
		inventory map[string]string
		aliases   map[string]string
		want      string
	}{
		{
			name:      "from_inventory",
			inventory: map[string]string{"geoloc": "x9000c1s7b0n0"},
			want:      "x9000c1s7b0n0",
		},
		{
			name:    "from_alias",
			aliases: map[string]string{"cm-geo-name": "x9000c1s4b0n0"},
			want:    "x9000c1s4b0n0",
		},
		{
			name:      "inventory_preferred",
			inventory: map[string]string{"geoloc": "x9000c1s7b0n0"},
			aliases:   map[string]string{"cm-geo-name": "x9000c1s4b0n0"},
			want:      "x9000c1s7b0n0",
		},
		{
			name: "nil_both",
			want: "",
		},
		{
			name:      "empty_geoloc",
			inventory: map[string]string{"geoloc": ""},
			aliases:   map[string]string{"cm-geo-name": "x9000c3s1b0n0"},
			want:      "x9000c3s1b0n0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nodeGeolocXname(tt.inventory, tt.aliases)
			if got != tt.want {
				t.Errorf("nodeGeolocXname() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveGeolocParent(t *testing.T) {
	chassisID := uuid.New()
	byLoc := map[string]uuid.UUID{"9000-1": chassisID}
	byXname := map[string]uuid.UUID{"x9000c1": chassisID}

	tests := []struct {
		name    string
		geoloc  string
		byLoc   map[string]uuid.UUID
		byXname map[string]uuid.UUID
		wantID  uuid.UUID
	}{
		{
			name:    "found_by_loc",
			geoloc:  "x9000c1s7b0n0",
			byLoc:   byLoc,
			byXname: byXname,
			wantID:  chassisID,
		},
		{
			name:    "found_by_xname_only",
			geoloc:  "x9000c1s7b0n0",
			byLoc:   map[string]uuid.UUID{},
			byXname: byXname,
			wantID:  chassisID,
		},
		{
			name:    "empty_geoloc",
			geoloc:  "",
			byLoc:   byLoc,
			byXname: byXname,
			wantID:  uuid.Nil,
		},
		{
			name:    "invalid_xname",
			geoloc:  "not-an-xname",
			byLoc:   byLoc,
			byXname: byXname,
			wantID:  uuid.Nil,
		},
		{
			name:    "chassis_not_found",
			geoloc:  "x5000c2s1b0n0",
			byLoc:   byLoc,
			byXname: byXname,
			wantID:  uuid.Nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveGeolocParent(tt.geoloc, tt.byLoc, tt.byXname)
			if got != tt.wantID {
				t.Errorf("resolveGeolocParent(%q) = %s, want %s",
					tt.geoloc, got, tt.wantID)
			}
		})
	}
}
