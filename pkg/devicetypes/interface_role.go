package devicetypes

import (
	"fmt"
	"strings"
)

// Well-known interface role names.
const (
	InterfaceRoleManagement = "management"
	InterfaceRoleHSN        = "hsn"
	InterfaceRoleStorage    = "storage"
	InterfaceRoleAccess     = "access"
	InterfaceRoleUplink     = "uplink"
	InterfaceRolePeer       = "peer"

	// Interface-specific roles (for Nautobot dcim.interface content type)
	InterfaceRoleManagementIface = "ManagementInterface"
	InterfaceRoleHSNIface        = "HSNInterface"
	InterfaceRoleDataIface       = "DataInterface"
	InterfaceRoleUplinkIface     = "UplinkInterface"
	InterfaceRoleStorageIface    = "StorageInterface"
)

// knownInterfaceRoles is the set of well-known role names.
var knownInterfaceRoles = map[string]bool{
	InterfaceRoleManagement:      true,
	InterfaceRoleHSN:             true,
	InterfaceRoleStorage:         true,
	InterfaceRoleAccess:          true,
	InterfaceRoleUplink:          true,
	InterfaceRolePeer:            true,
	InterfaceRoleManagementIface: true,
	InterfaceRoleHSNIface:        true,
	InterfaceRoleDataIface:       true,
	InterfaceRoleUplinkIface:     true,
	InterfaceRoleStorageIface:    true,
}

// ValidateInterfaceRole returns a warning message if the role is not a
// well-known role name. Returns empty string if the role is recognized.
// Unknown roles are allowed (Nautobot supports custom roles) but emit a warning.
func ValidateInterfaceRole(role string) string {
	if role == "" {
		return ""
	}
	if knownInterfaceRoles[role] {
		return ""
	}
	// Also check lowercase for the legacy role names
	if knownInterfaceRoles[strings.ToLower(role)] {
		return ""
	}
	return fmt.Sprintf("role %q is not a well-known role; known roles: management, hsn, storage, access, uplink, peer, ManagementInterface, HSNInterface, DataInterface, UplinkInterface, StorageInterface", role)
}

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
