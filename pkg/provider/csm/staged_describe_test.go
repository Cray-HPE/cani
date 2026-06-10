package csm

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func TestParseXnameParts(t *testing.T) {
	cases := []struct {
		xn              string
		cab, chas, slot int
	}{
		{"x9000c1s0", 9000, 1, 0},
		{"x8000", 8000, 0, 0},
		{"x8000c3", 8000, 3, 0},
		{"x8000c3s7", 8000, 3, 7},
		{"x9000c1s0b0n1", 9000, 1, 0},
		{"bogus", 0, 0, 0},
		{"", 0, 0, 0},
	}
	for _, tc := range cases {
		cab, chas, slot := parseXnameParts(tc.xn)
		if cab != tc.cab || chas != tc.chas || slot != tc.slot {
			t.Errorf("parseXnameParts(%q) = (%d,%d,%d), want (%d,%d,%d)",
				tc.xn, cab, chas, slot, tc.cab, tc.chas, tc.slot)
		}
	}
}

func TestDescribeStagedDevice(t *testing.T) {
	p := New()

	dev := &devicetypes.CaniDeviceType{}
	dev.ID = uuid.New()
	dev.SetProviderMeta(p.Slug(), "xname", "x9000c1s7")
	lines := p.DescribeStagedDevice(dev)
	want := []string{"Cabinet: 9000", "Chassis: 1", "Blade: 7"}
	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d: %v", len(lines), len(want), lines)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("line %d = %q, want %q", i, lines[i], want[i])
		}
	}

	if got := p.DescribeStagedDevice(&devicetypes.CaniDeviceType{}); got != nil {
		t.Errorf("expected nil for device without metadata, got %v", got)
	}

	dev2 := &devicetypes.CaniDeviceType{}
	dev2.SetProviderMeta(p.Slug(), "xname", "")
	if got := p.DescribeStagedDevice(dev2); got != nil {
		t.Errorf("expected nil for empty xname, got %v", got)
	}
}
