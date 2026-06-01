package devicetypes

import "testing"

func TestInferInterfaceRole(t *testing.T) {
	tests := []struct {
		name     string
		ifName   string
		ifType   InterfacesElemType
		mgmtOnly bool
		want     string
	}{
		{"mgmt_only flag", "eth0", InterfacesElemTypeA1000BaseT, true, InterfaceRoleManagement},
		{"iLO name", "iLO", InterfacesElemTypeA1000BaseT, false, InterfaceRoleManagement},
		{"BMC name", "BMC", InterfacesElemTypeA1000BaseT, false, InterfaceRoleManagement},
		{"mgmt0 name", "mgmt0", InterfacesElemTypeA1000BaseT, false, InterfaceRoleManagement},
		{"oob name", "oob0", InterfacesElemTypeA1000BaseT, false, InterfaceRoleManagement},
		{"HSN name", "HSN 0", InterfacesElemTypeA100GbaseXQsfp28, false, InterfaceRoleHSN},
		{"ib name", "ib0", InterfacesElemTypeA200GbaseXQsfp56, false, InterfaceRoleHSN},
		{"osfp prefix", "osfp1", InterfacesElemTypeA400GbaseXOsfp, false, InterfaceRoleHSN},
		{"400G type", "port0", InterfacesElemTypeA400GbaseXQsfpdd, false, InterfaceRoleHSN},
		{"200G type", "port0", InterfacesElemTypeA200GbaseXQsfp56, false, InterfaceRoleHSN},
		{"plain eth no role", "eth0", InterfacesElemTypeA1000BaseT, false, ""},
		{"10G SFP no role", "sfp1", InterfacesElemTypeA10GbaseXSfpp, false, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := InferInterfaceRole(tc.ifName, tc.ifType, tc.mgmtOnly)
			if got != tc.want {
				t.Errorf("InferInterfaceRole(%q, %q, %v) = %q, want %q",
					tc.ifName, tc.ifType, tc.mgmtOnly, got, tc.want)
			}
		})
	}
}

func TestResolveInterfaceRole(t *testing.T) {
	// Explicit role takes priority over inference
	got := ResolveInterfaceRole("storage", "iLO", InterfacesElemTypeA1000BaseT, true)
	if got != "storage" {
		t.Errorf("ResolveInterfaceRole with explicit role = %q, want %q", got, "storage")
	}

	// Empty explicit falls through to inference
	got = ResolveInterfaceRole("", "iLO", InterfacesElemTypeA1000BaseT, true)
	if got != InterfaceRoleManagement {
		t.Errorf("ResolveInterfaceRole with empty explicit = %q, want %q", got, InterfaceRoleManagement)
	}
}
