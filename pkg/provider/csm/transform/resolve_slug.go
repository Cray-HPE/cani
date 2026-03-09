package transform

import (
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/csm/import"
)

// defaultSlugs maps (class, CaniType) to a reasonable default slug
// when SLS does not provide enough info to resolve a specific model.
var defaultSlugs = map[string]map[devicetypes.Type]string{
	ClassRiver: {
		devicetypes.TypeCabinet:    "hpe-eia-cabinet",
		devicetypes.TypeChassis:    "hpe-eia-chassis",
		devicetypes.TypeMgmtSwitch: "hpe-aruba-6300m-48g",
		devicetypes.TypeSwitch:     "hpe-aruba-8325-32c",
		devicetypes.TypeBlade:      "hpe-dl380-gen-11",
		devicetypes.TypeNodeCard:   "hpe-standard-node-bmc",
		devicetypes.TypeNode:       "cray-xd225v",
	},
	ClassMountain: {
		devicetypes.TypeCabinet:  "hpe-ex2000",
		devicetypes.TypeChassis:  "hpe-crayex-chassis",
		devicetypes.TypeBlade:    "hpe-crayex-ex235a-compute-blade",
		devicetypes.TypeNodeCard: "hpe-crayex-ex235a-compute-blade-bard-peak-node-card",
		devicetypes.TypeNode:     "hpe-crayex-ex235a-compute-node",
	},
	// Hill cabinets use the same EX2000 enclosure as Mountain.
	// Other Hill component types fall through to the Hill-specific
	// fallback in defaultSlugForClass (River then Mountain).
	ClassHill: {
		devicetypes.TypeCabinet: "hpe-ex2000",
	},
}

// resolveSlug attempts to determine a device-type slug for a classified
// SLS hardware entry. It tries, in order:
//
//  1. Brand+Model from SLS ExtraProperties (switches have these fields).
//  2. Class-based default for the CANI type.
//
// Returns "" when no slug can be determined.
func resolveSlug(cl CsmClassification) string {
	// 1. Try to match from SLS ExtraProperties.
	if slug := slugFromExtraProperties(cl); slug != "" {
		return slug
	}

	// 2. Fall back to class-based default.
	return defaultSlugForClass(cl)
}

// slugFromExtraProperties attempts to resolve a slug by looking at
// Brand/Model fields in SLS ExtraProperties.
func slugFromExtraProperties(cl CsmClassification) string {
	if cl.Hardware.ExtraProperties == nil {
		return ""
	}
	var brand, model string

	switch cl.Hardware.TypeString {
	case XnameTypeMgmtSwitch:
		ep, err := import_.DecodeExtraProperties[import_.SlsMgmtSwitchExtraProperties](
			cl.Hardware.ExtraProperties,
		)
		if err == nil {
			brand = ep.Brand
			model = ep.Model
		}
	case XnameTypeMgmtHLSwitch:
		ep, err := import_.DecodeExtraProperties[import_.SlsMgmtHLSwitchExtraProperties](
			cl.Hardware.ExtraProperties,
		)
		if err == nil {
			brand = ep.Brand
			model = ep.Model
		}
	default:
		return ""
	}

	return lookupByBrandModel(brand, model)
}

// lookupByBrandModel builds a query from brand and model and attempts a
// library lookup. Returns the slug if the match score is acceptable.
// When model is empty the brand alone is too ambiguous for fuzzy matching
// so we return "" and let the caller fall through to class-based defaults.
func lookupByBrandModel(brand, model string) string {
	if strings.TrimSpace(model) == "" {
		return ""
	}
	query := strings.TrimSpace(brand + " " + model)
	if query == "" {
		return ""
	}
	dt, score := devicetypes.LookupScored(query)
	if score >= 30 && dt.Slug != "" {
		return dt.Slug
	}
	return ""
}

// defaultSlugForClass returns a sensible default slug for the given
// CANI type and cabinet class. Hill tries River then Mountain defaults.
func defaultSlugForClass(cl CsmClassification) string {
	class := cl.Hardware.Class
	if class == "" {
		class = classForCabinetNumber(cl.Xname.Cabinet)
	}

	if slugs, ok := defaultSlugs[class]; ok {
		if slug, ok := slugs[cl.CaniType]; ok {
			return slug
		}
	}

	// Hill: try River defaults, then Mountain.
	if class == ClassHill {
		if slug, ok := defaultSlugs[ClassRiver][cl.CaniType]; ok {
			return slug
		}
		if slug, ok := defaultSlugs[ClassMountain][cl.CaniType]; ok {
			return slug
		}
	}

	return ""
}
