package transform

// HardwareClass constants for HPE/Cray cabinet architectures.
const (
	ClassRiver    = "River"
	ClassMountain = "Mountain"
	ClassHill     = "Hill"
)

// riverTypes lists SLS TypeStrings valid in River cabinets.
// River cabinets are standard 19" racks with management switches,
// spine switches, switch connectors, compute modules (blades),
// node BMCs, nodes, and PDU controllers.
var riverTypes = map[string]bool{
	XnameTypeCabinet:              true,
	XnameTypeMgmtSwitch:           true,
	XnameTypeMgmtHLSwitch:         true,
	XnameTypeMgmtSwitchConnector:  true,
	XnameTypeComputeModule:        true,
	XnameTypeNodeBMC:              true,
	XnameTypeNode:                 true,
	XnameTypeNodeEnclosure:        true,
	XnameTypeCabinetPDUController: true,
}

// mountainTypes lists SLS TypeStrings valid in Mountain cabinets.
// Mountain cabinets are liquid-cooled chassis-based systems with
// chassis BMCs, router modules (HSN switches), and CDU switches.
// No management switches or switch connectors.
var mountainTypes = map[string]bool{
	XnameTypeCabinet:       true,
	XnameTypeChassis:       true,
	XnameTypeChassisBMC:    true,
	XnameTypeComputeModule: true,
	XnameTypeNodeBMC:       true,
	XnameTypeNode:          true,
	XnameTypeRouterModule:  true,
	XnameTypeRouterBMC:     true,
	XnameTypeHSNBoard:      true,
	XnameTypeMgmtCDUSwitch: true,
}

// validTypeForClass reports whether typeString is expected for the
// given cabinet class. Hill allows the union of River and Mountain.
// An empty or unrecognised class always returns true (permissive).
func validTypeForClass(typeString, class string) bool {
	switch class {
	case ClassRiver:
		return riverTypes[typeString]
	case ClassMountain:
		return mountainTypes[typeString]
	case ClassHill:
		return riverTypes[typeString] || mountainTypes[typeString]
	default:
		return true
	}
}

// classForCabinetNumber returns the conventional class based on
// HPE cabinet numbering ranges. Returns "" when the range is unknown.
//
//	x1000–x2999 → Mountain
//	x3000–x3999 → River
//	x9000–x9999 → Hill
func classForCabinetNumber(cabinet int) string {
	switch {
	case cabinet >= 1000 && cabinet <= 2999:
		return ClassMountain
	case cabinet >= 3000 && cabinet <= 3999:
		return ClassRiver
	case cabinet >= 9000 && cabinet <= 9999:
		return ClassHill
	default:
		return ""
	}
}
