package transform

import (
	"testing"

	import_ "github.com/Cray-HPE/cani/pkg/provider/ochami/import"
)

func TestInferCableTypeSlug_Pass(t *testing.T) {
	tests := []struct {
		description string
		want        string
	}{
		{"DAC cable 3m", cableTypeDacPassive},
		{"direct-attach copper", cableTypeDacPassive},
		{"AOC 10m active optical", cableTypeAoc},
		{"OM4 multimode fiber", cableTypeMmfOm4},
		{"MMF patch cord", cableTypeMmfOm4},
		{"single-mode fiber SMF", cableTypeSmf},
		{"power cord 2m", cableTypePower},
		{"jumper cable", cableTypePower},
		{"serial console adapter", cableTypeOther},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			got := inferCableTypeSlug(tt.description)
			if got != tt.want {
				t.Errorf("inferCableTypeSlug(%q) = %q, want %q", tt.description, got, tt.want)
			}
		})
	}
}

func TestInferCableTypeSlug_NoMatch(t *testing.T) {
	// Descriptions that match no known pattern should fall back to "other".
	tests := []struct {
		description string
	}{
		{"serial console adapter"},
		{""},
		{"unknown cable xyz"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			got := inferCableTypeSlug(tt.description)
			if got != cableTypeOther {
				t.Errorf("inferCableTypeSlug(%q) = %q, want %q", tt.description, got, cableTypeOther)
			}
		})
	}
}

func TestResolveCableTypeSlug_Pass(t *testing.T) {
	// With an empty part number the function falls back to description inference.
	got := resolveCableTypeSlug("", "active optical cable")
	if got != cableTypeAoc {
		t.Errorf("resolveCableTypeSlug(%q, %q) = %q, want %q", "", "active optical cable", got, cableTypeAoc)
	}
}

func TestResolveCableTypeSlug_UnknownPartNumber(t *testing.T) {
	// An unregistered part number should fall back to description-based inference.
	got := resolveCableTypeSlug("BOGUS-PN-999", "active optical cable")
	if got != cableTypeAoc {
		t.Errorf("resolveCableTypeSlug(%q, %q) = %q, want %q", "BOGUS-PN-999", "active optical cable", got, cableTypeAoc)
	}

	// An unregistered part number with an unrecognised description should return "other".
	got = resolveCableTypeSlug("BOGUS-PN-000", "mystery cable")
	if got != cableTypeOther {
		t.Errorf("resolveCableTypeSlug(%q, %q) = %q, want %q", "BOGUS-PN-000", "mystery cable", got, cableTypeOther)
	}
}

func TestCreateCable_Pass(t *testing.T) {
	rec := import_.JSONDeviceRecord{
		DeviceType:   "cable",
		SerialNumber: "CBL-001",
		Manufacturer: "Mellanox",
		PartNumber:   "UNKNOWN-PN-999",
	}

	cable := createCable(rec)

	if cable.Label != "CBL-001" {
		t.Errorf("Label = %q, want %q", cable.Label, "CBL-001")
	}
	if cable.Manufacturer != "Mellanox" {
		t.Errorf("Manufacturer = %q, want %q", cable.Manufacturer, "Mellanox")
	}
	if cable.PartNumber != "UNKNOWN-PN-999" {
		t.Errorf("PartNumber = %q, want %q", cable.PartNumber, "UNKNOWN-PN-999")
	}
	if cable.Status != "connected" {
		t.Errorf("Status = %q, want %q", cable.Status, "connected")
	}
}

func TestCreateCable_EmptyFields(t *testing.T) {
	// A record with no manufacturer, no part number, and no serial number
	// should still produce a valid cable with sensible defaults.
	rec := import_.JSONDeviceRecord{
		DeviceType: "cable",
	}

	cable := createCable(rec)

	if cable.Slug != cableTypeOther {
		t.Errorf("Slug = %q, want %q", cable.Slug, cableTypeOther)
	}
	if cable.Label != "" {
		t.Errorf("Label = %q, want empty string", cable.Label)
	}
	if cable.Manufacturer != "" {
		t.Errorf("Manufacturer = %q, want empty string", cable.Manufacturer)
	}
	if cable.PartNumber != "" {
		t.Errorf("PartNumber = %q, want empty string", cable.PartNumber)
	}
	if cable.Status != "connected" {
		t.Errorf("Status = %q, want %q", cable.Status, "connected")
	}
}
