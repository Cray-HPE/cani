package transform

import "testing"

func TestParseXname(t *testing.T) {
	tests := []struct {
		xname   string
		wantTyp string
		cabinet int
		chassis int
		slot    int
		bmc     int
		node    int
		port    int
	}{
		{"x3000", XnameTypeCabinet, 3000, 0, 0, 0, 0, 0},
		{"x3000c0", XnameTypeChassis, 3000, 0, 0, 0, 0, 0},
		{"x3000c0s9", XnameTypeComputeModule, 3000, 0, 9, 0, 0, 0},
		{"x3000c0s9b0", XnameTypeNodeBMC, 3000, 0, 9, 0, 0, 0},
		{"x3000c0s9b0n0", XnameTypeNode, 3000, 0, 9, 0, 0, 0},
		{"x3000c0r7", XnameTypeRouterModule, 3000, 0, 7, 0, 0, 0},
		{"x3000c0r7b0", XnameTypeRouterBMC, 3000, 0, 7, 0, 0, 0},
		{"x3000c0w14", XnameTypeMgmtSwitch, 3000, 0, 14, 0, 0, 0},
		{"x3000c0w14j36", XnameTypeMgmtSwitchConnector, 3000, 0, 14, 0, 0, 36},
		{"x3000c0h1s1", XnameTypeMgmtHLSwitch, 3000, 0, 1, 1, 0, 0},
		{"x3000c0b0", XnameTypeChassisBMC, 3000, 0, 0, 0, 0, 0},
		{"x3000c0s9e0", XnameTypeNodeEnclosure, 3000, 0, 9, 0, 0, 0},
		{"x3000c0r7e0", XnameTypeHSNBoard, 3000, 0, 7, 0, 0, 0},
		{"x9000c1s3b2n1", XnameTypeNode, 9000, 1, 3, 2, 1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.xname, func(t *testing.T) {
			info := ParseXname(tt.xname)
			if info.Type != tt.wantTyp {
				t.Errorf("Type = %q, want %q", info.Type, tt.wantTyp)
			}
			if info.Cabinet != tt.cabinet {
				t.Errorf("Cabinet = %d, want %d", info.Cabinet, tt.cabinet)
			}
			if info.Chassis != tt.chassis {
				t.Errorf("Chassis = %d, want %d", info.Chassis, tt.chassis)
			}
			if info.Slot != tt.slot {
				t.Errorf("Slot = %d, want %d", info.Slot, tt.slot)
			}
			if info.Port != tt.port {
				t.Errorf("Port = %d, want %d", info.Port, tt.port)
			}
		})
	}
}

func TestParseXname_Unknown(t *testing.T) {
	info := ParseXname("unknown")
	if info.Type != "" {
		t.Errorf("Type = %q, want empty", info.Type)
	}
	if info.Raw != "unknown" {
		t.Errorf("Raw = %q, want unknown", info.Raw)
	}
}

func TestXnameParent(t *testing.T) {
	tests := []struct {
		xname  string
		parent string
	}{
		{"x3000", "s0"},
		{"x3000c0", "x3000"},
		{"x3000c0s9", "x3000c0"},
		{"x3000c0s9b0", "x3000c0s9"},
		{"x3000c0s9b0n0", "x3000c0s9b0"},
		{"x3000c0r7", "x3000c0"},
		{"x3000c0w14", "x3000c0"},
		{"x3000c0h1s1", "x3000c0h1"},
		{"x3000c0w14j36", "x3000c0w14"},
		{"x3000c0r7b0", "x3000c0r7"},
		{"x3000c0r7e0", "x3000c0r7"},
		{"x3000c0s9e0", "x3000c0s9"},
		{"x3000c0b0", "x3000c0"},
	}
	for _, tt := range tests {
		t.Run(tt.xname, func(t *testing.T) {
			info := ParseXname(tt.xname)
			got := info.Parent()
			if got != tt.parent {
				t.Errorf("Parent() = %q, want %q", got, tt.parent)
			}
		})
	}
}

func TestGetParentXname(t *testing.T) {
	tests := []struct {
		xname  string
		parent string
	}{
		{"x3000", "s0"},
		{"x3000c0", "x3000"},
		{"x3000c0s9b0n0", "x3000c0s9b0"},
		{"x3000c0h1s1", "x3000c0h1"},
		{"", ""},
	}
	for _, tt := range tests {
		name := tt.xname
		if name == "" {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			got := GetParentXname(tt.xname)
			if got != tt.parent {
				t.Errorf("GetParentXname(%q) = %q, want %q", tt.xname, got, tt.parent)
			}
		})
	}
}
