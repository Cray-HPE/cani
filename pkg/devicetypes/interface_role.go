package devicetypes

import "strings"

// Well-known interface role names.
const (
	InterfaceRoleManagement = "management"
	InterfaceRoleHSN        = "hsn"
	InterfaceRoleStorage    = "storage"
	InterfaceRoleAccess     = "access"
)

// InferInterfaceRole returns a role name based on the interface's name,
// type, and mgmt_only flag. Returns empty string if no role can be inferred.
// Priority: mgmt_only flag > name patterns > type-based heuristics.
func InferInterfaceRole(name string, ifaceType InterfacesElemType, mgmtOnly bool) string {
	if mgmtOnly {
		return InterfaceRoleManagement
	}

	lower := strings.ToLower(name)

	// Management patterns
	if matchesAny(lower, "ilo", "bmc", "mgmt", "oob", "ipmi") {
		return InterfaceRoleManagement
	}

	// High-speed network patterns (by name)
	if matchesAny(lower, "hsn", "ib") || strings.HasPrefix(lower, "osfp") || strings.HasPrefix(lower, "qsfp") {
		return InterfaceRoleHSN
	}

	// High-speed network patterns (by type)
	switch ifaceType {
	case InterfacesElemTypeA200GbaseXQsfp56,
		InterfacesElemTypeA400GbaseXQsfpdd,
		InterfacesElemTypeA400GbaseXOsfp:
		return InterfaceRoleHSN
	case InterfacesElemTypeA100GbaseXQsfp28:
		// 100G could be HSN or uplink — infer HSN only if name suggests it
		if matchesAny(lower, "hsn", "ib", "fabric") {
			return InterfaceRoleHSN
		}
	}

	return ""
}

// ResolveInterfaceRole returns the explicit role if set, otherwise infers one.
func ResolveInterfaceRole(explicit string, name string, ifaceType InterfacesElemType, mgmtOnly bool) string {
	if explicit != "" {
		return explicit
	}
	return InferInterfaceRole(name, ifaceType, mgmtOnly)
}

// matchesAny returns true if s contains any of the given substrings.
func matchesAny(s string, patterns ...string) bool {
	for _, p := range patterns {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}
