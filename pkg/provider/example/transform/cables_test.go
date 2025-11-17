package transform

import (
	"testing"
)

func TestInferHardwareType(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        string
	}{
		{"48U rack", "HPE 48U 800mmx1200mm G2 Enterprise Shock Rack", "rack"},
		{"cabinet", "HPE Cabinet G2", "rack"},
		{"aruba switch", "HPE Aruba Networking 8360-48Y6C v2", "switch"},
		{"generic switch", "HPE 48-Port Ethernet Switch", "switch"},
		{"proliant server", "HPE ProLiant DL380 Gen11", "node"},
		{"blade server", "HPE BladeSystem c7000 Blade", "node"},
		{"cat6 cable", "HPE Cat6 RJ45 M/M 2m", "cable"},
		{"cat5e cable", "HPE CAT5e RJ45 2.3m Cable", "cable"},
		{"DAC cable", "HPE 400G QSFP-DD DAC 3m", "cable"},
		{"direct attach cable", "HPE 100Gb Direct Attach Copper Cable", "cable"},
		{"AOC cable", "HPE Aruba 100G QSFP28 15m AOC", "cable"},
		{"active optical cable", "HPE Active Optical Cable 30m", "cable"},
		{"OM4 fiber", "HPE Premier Flex LC LC OM4 15m", "cable"},
		{"power jumper", "HPE C19 C20 250V 16A 2m Jumper", "cable"},
		{"RJ45 cable", "HPE RJ45 to RJ45 Cat5e Black", "cable"},
		{"QSFP cable", "HPE QSFP28 to QSFP28 Cable", "cable"},
		{"generic cable", "HPE Data Cable 3m", "cable"},
		{"unknown device", "XD670", ""},
		{"memory module", "HPE 64GB DDR5 Memory Kit", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferHardwareType(tt.description)
			if got != tt.want {
				t.Errorf("inferHardwareType(%q) = %q, want %q", tt.description, got, tt.want)
			}
		})
	}
}

func TestInferCableTypeSlug(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        string
	}{
		{"cat5 cable", "HPE CAT5 RJ45 1.2m Cable", cableTypeCat5e},
		{"cat5e cable", "HPE CAT5e RJ45 2.3m Cable", cableTypeCat5e},
		{"cat6 cable", "HPE Cat6 RJ45 M/M 2m", cableTypeCat6},
		{"cat6a cable", "HPE Cat6a Shielded Cable 3m", cableTypeCat6a},
		{"DAC cable", "HPE 400G QSFP-DD DAC 3m", cableTypeDacPassive},
		{"direct attach cable", "HPE 100Gb Direct Attach Copper Cable", cableTypeDacPassive},
		{"AOC cable", "HPE Aruba 100G QSFP28 15m AOC", cableTypeAoc},
		{"active optical cable", "HPE Active Optical Cable 30m", cableTypeAoc},
		{"OM3 fiber", "HPE LC LC OM3 2F 30m", cableTypeMmfOm4},
		{"OM4 fiber", "HPE Premier Flex LC LC OM4 15m", cableTypeMmfOm4},
		{"MMF fiber", "HPE InfiniBand NDR MPO MPO MM 10m", cableTypeMmfOm4},
		{"SMF fiber", "HPE InfiniBand NDR MPO MPO SM 10m", cableTypeSmf},
		{"single mode fiber", "HPE Single Mode Fiber Cable 20m", cableTypeSmf},
		{"power jumper", "HPE C19 C20 250V 16A 2m Jumper", cableTypePower},
		{"power cord", "HPE Power Cord 2m", cableTypePower},
		{"generic cable", "HPE Data Cable 3m", cableTypeOther},
		{"QSFP without type", "HPE QSFP28 Cable", cableTypeOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferCableTypeSlug(tt.description)
			if got != tt.want {
				t.Errorf("inferCableTypeSlug(%q) = %q, want %q", tt.description, got, tt.want)
			}
		})
	}
}

func TestParseLengthFromDescription(t *testing.T) {
	tests := []struct {
		name        string
		description string
		wantLength  float64
		wantUnit    string
	}{
		{"meters", "HPE Cat6 RJ45 M/M 2m", 2, "m"},
		{"feet", "HPE Cable 10ft", 10, "ft"},
		{"decimal meters", "HPE 1.5m DAC Cable", 1.5, "m"},
		{"centimeters", "HPE 50cm Patch Cable", 50, "cm"},
		{"no length", "HPE Generic Cable", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLength, gotUnit := parseLengthFromDescription(tt.description)
			if gotLength != tt.wantLength {
				t.Errorf("parseLengthFromDescription(%q) length = %v, want %v", tt.description, gotLength, tt.wantLength)
			}
			if gotUnit != tt.wantUnit {
				t.Errorf("parseLengthFromDescription(%q) unit = %q, want %q", tt.description, gotUnit, tt.wantUnit)
			}
		})
	}
}

func TestGenerateCableLabel(t *testing.T) {
	tests := []struct {
		name        string
		description string
		index       int
		total       int
		want        string
	}{
		{"single cable", "HPE Cat6 Cable", 0, 1, "HPE Cat6 Cable"},
		{"first of multiple", "HPE Cat6 Cable", 0, 3, "HPE Cat6 Cable-001"},
		{"second of multiple", "HPE Cat6 Cable", 1, 3, "HPE Cat6 Cable-002"},
		{"tenth cable", "HPE DAC Cable", 9, 20, "HPE DAC Cable-010"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateCableLabel(tt.description, tt.index, tt.total)
			if got != tt.want {
				t.Errorf("generateCableLabel(%q, %d, %d) = %q, want %q",
					tt.description, tt.index, tt.total, got, tt.want)
			}
		})
	}
}
