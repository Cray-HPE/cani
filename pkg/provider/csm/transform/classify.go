package transform

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/csm/import"
)

// CsmClassification holds the result of classifying an SLS hardware entry.
type CsmClassification struct {
	Category devicetypes.Category
	CaniType devicetypes.Type
	Xname    XnameInfo
	Hardware import_.SlsHardware
	Smd      *import_.SmdComponent // nil if no SMD match
	Skip     bool                  // true if this entry should be skipped
	Warning  string                // non-empty when type is unusual for the class
}

// classifyHardware maps an SLS hardware entry to a CANI classification.
// It validates that the hardware TypeString is expected for the cabinet
// class (River, Mountain, Hill). Mismatches set Warning but still allow
// the entry to be classified normally.
func classifyHardware(hw import_.SlsHardware, smdMap map[string]import_.SmdComponent) CsmClassification {
	xname := ParseXname(hw.Xname)
	cl := CsmClassification{
		Xname:    xname,
		Hardware: hw,
	}

	// Attach SMD data if available.
	if smd, ok := smdMap[hw.Xname]; ok {
		cl.Smd = &smd
	}

	// Resolve the effective class: prefer the explicit Class field,
	// then fall back to the conventional cabinet-number range.
	effectiveClass := hw.Class
	if effectiveClass == "" {
		effectiveClass = classForCabinetNumber(xname.Cabinet)
	}

	// Map SLS TypeString to CANI type and category.
	classifyByTypeString(&cl, hw.TypeString)

	// Validate that the TypeString is expected for the cabinet class.
	if !cl.Skip && effectiveClass != "" {
		if !validTypeForClass(hw.TypeString, effectiveClass) {
			cl.Warning = hw.TypeString + " is unexpected in " + effectiveClass + " cabinet " + hw.Xname
		}
	}

	return cl
}

// classifyByTypeString sets Category, CaniType, and Skip on cl
// based on the SLS TypeString value.
func classifyByTypeString(cl *CsmClassification, typeString string) {
	switch typeString {
	case XnameTypeCabinet:
		cl.Category = devicetypes.CategoryDevice
		cl.CaniType = devicetypes.TypeCabinet

	case XnameTypeChassis:
		cl.Category = devicetypes.CategoryDevice
		cl.CaniType = devicetypes.TypeChassis

	case XnameTypeChassisBMC:
		cl.Category = devicetypes.CategoryModule
		cl.CaniType = devicetypes.TypeModule
		cl.Skip = true // chassis BMC metadata tracked on chassis device

	case XnameTypeComputeModule:
		cl.Category = devicetypes.CategoryDevice
		cl.CaniType = devicetypes.TypeBlade

	case XnameTypeNodeBMC:
		cl.Category = devicetypes.CategoryDevice
		cl.CaniType = devicetypes.TypeNodeCard

	case XnameTypeNode:
		cl.Category = devicetypes.CategoryDevice
		cl.CaniType = devicetypes.TypeNode

	case XnameTypeRouterModule:
		cl.Category = devicetypes.CategoryDevice
		cl.CaniType = devicetypes.TypeHsnSwitch

	case XnameTypeRouterBMC:
		cl.Category = devicetypes.CategoryModule
		cl.CaniType = devicetypes.TypeModule
		cl.Skip = true // router BMC metadata tracked on router device

	case XnameTypeMgmtSwitch:
		cl.Category = devicetypes.CategoryDevice
		cl.CaniType = devicetypes.TypeMgmtSwitch

	case XnameTypeMgmtHLSwitch:
		cl.Category = devicetypes.CategoryDevice
		cl.CaniType = devicetypes.TypeSwitch

	case XnameTypeMgmtSwitchConnector:
		cl.Skip = true // connectors are port-level, not devices

	case XnameTypeNodeEnclosure:
		cl.Skip = true // enclosures are physical containers, not CANI devices

	case XnameTypeHSNBoard:
		cl.Skip = true // HSN boards tracked via RouterModule

	case XnameTypeCabinetPDUController:
		cl.Category = devicetypes.CategoryDevice
		cl.CaniType = devicetypes.TypeCabinetPDU

	case XnameTypeMgmtCDUSwitch:
		cl.Category = devicetypes.CategoryDevice
		cl.CaniType = devicetypes.TypeCDU

	default:
		cl.Skip = true
	}
}
