package transform

import (
	"regexp"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/ochami/import"
)

// Cable type slug constants.
const (
	cableTypeDacPassive = "dac-passive"
	cableTypeAoc        = "aoc"
	cableTypeMmfOm4     = "mmf-om4"
	cableTypeSmf        = "smf"
	cableTypePower      = "power"
	cableTypeOther      = "other"
)

// Cable type inference patterns (case-insensitive).
var (
	dacPattern   = regexp.MustCompile(`(?i)dac|direct.?attach`)
	aocPattern   = regexp.MustCompile(`(?i)aoc|active.?optical`)
	mmfPattern   = regexp.MustCompile(`(?i)om[34]|mmf|\bMM\b`)
	smfPattern   = regexp.MustCompile(`(?i)smf|single.?mode|\bSM\b`)
	powerPattern = regexp.MustCompile(`(?i)power.?cord|jumper`)
)

// createCable builds a CaniCableType from a JSONDeviceRecord.
func createCable(rec import_.JSONDeviceRecord) *devicetypes.CaniCableType {
	slug := resolveCableTypeSlug(rec.PartNumber, rec.DeviceType)
	cable := devicetypes.NewCable(slug, rec.SerialNumber)
	cable.Manufacturer = rec.Manufacturer
	cable.PartNumber = rec.PartNumber
	return cable
}

// resolveCableTypeSlug resolves the cable type slug using the cascade:
// 1. Lookup by part number in the cable type library
// 2. Infer from description patterns
func resolveCableTypeSlug(partNumber, description string) string {
	if partNumber != "" {
		if ct, ok := devicetypes.GetCableTypeByPartNumber(partNumber); ok {
			return ct.Slug
		}
	}
	return inferCableTypeSlug(description)
}

// inferCableTypeSlug derives a cable type slug from description patterns.
func inferCableTypeSlug(description string) string {
	switch {
	case dacPattern.MatchString(description):
		return cableTypeDacPassive
	case aocPattern.MatchString(description):
		return cableTypeAoc
	case mmfPattern.MatchString(description):
		return cableTypeMmfOm4
	case smfPattern.MatchString(description):
		return cableTypeSmf
	case powerPattern.MatchString(description):
		return cableTypePower
	default:
		return cableTypeOther
	}
}
