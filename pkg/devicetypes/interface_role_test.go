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

// TestInferInterfaceRole100GAmbiguous verifies InferInterfaceRole leaves a 100G
// interface unclassified when its name gives no high-speed hint.
//
// Why it matters: 100G links can be either HSN or uplink, so the inferer must
// only claim HSN when the name corroborates it; otherwise the role stays empty
// for the operator to decide.
// Inputs: name "port0" with type 100GBASE-X QSFP28 and mgmtOnly false. Outputs:
// an empty role string.
// Data choice: a neutral name (no "hsn"/"ib"/"fabric" and no qsfp prefix) is the
// only way to reach the false branch of the 100G name check, which the existing
// table's "HSN name" case does not cover.
func TestInferInterfaceRole100GAmbiguous(t *testing.T) {
	if got := InferInterfaceRole("port0", InterfacesElemTypeA100GbaseXQsfp28, false); got != "" {
		t.Errorf("InferInterfaceRole(100G, neutral name) = %q, want empty", got)
	}
}

// TestValidateInterfaceRole verifies ValidateInterfaceRole returns no warning
// for empty, canonical, and legacy-lowercase roles, and a warning for an
// unknown role.
//
// Why it matters: unknown roles are still allowed (Nautobot supports custom
// roles) but must produce an operator-facing warning, so the function gates on
// recognition without rejecting.
// Inputs: "" (skipped), "management" (canonical), "HSN" (an uppercased form of
// the known "hsn"), and "bogus" (unknown). Outputs: empty string for the first
// three; a non-empty warning for "bogus".
// Data choice: "HSN" specifically exercises the strings.ToLower lookup branch
// (its exact case is absent from the map but its lowercase is present), distinct
// from the direct-map hit of "management".
func TestValidateInterfaceRole(t *testing.T) {
	cases := []struct {
		name     string
		role     string
		wantWarn bool
	}{
		{"empty", "", false},
		{"canonical known", InterfaceRoleManagement, false},
		{"uppercased legacy", "HSN", false},
		{"unknown", "bogus", true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			warn := ValidateInterfaceRole(tt.role)
			if tt.wantWarn && warn == "" {
				t.Errorf("ValidateInterfaceRole(%q) = empty, want a warning", tt.role)
			}
			if !tt.wantWarn && warn != "" {
				t.Errorf("ValidateInterfaceRole(%q) = %q, want empty", tt.role, warn)
			}
		})
	}
}
