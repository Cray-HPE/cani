/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package devicetypes

import "testing"

// TestGetRoleExplicit verifies GetRole returns the explicit Role field when it
// is set, ignoring provider metadata.
//
// Why it matters: an operator-assigned role must always take precedence over
// any role inferred from imported provider metadata.
// Inputs: an ObjectMeta with Role "leader" and a conflicting provider-metadata
// role "follower". Outputs: "leader".
// Data choice: deliberately conflicting values prove precedence rather than a
// coincidental match.
func TestGetRoleExplicit(t *testing.T) {
	m := ObjectMeta{
		Role: "leader",
		ProviderMetadata: map[string]any{
			"csm": map[string]any{"role": "follower"},
		},
	}
	if got := m.GetRole(); got != "leader" {
		t.Errorf("GetRole() = %q, want %q", got, "leader")
	}
}

// TestGetRoleFromProviderMetadata verifies GetRole falls back to the first
// "role" value found in provider metadata when Role is empty.
//
// Why it matters: devices imported from a provider often carry their role only
// in provider metadata, so the accessor must surface it when no explicit role
// exists.
// Inputs: an ObjectMeta with empty Role and a provider bucket containing
// role "spine". Outputs: "spine".
// Data choice: a single provider bucket with a role key isolates the fallback
// branch without map-ordering ambiguity.
func TestGetRoleFromProviderMetadata(t *testing.T) {
	m := ObjectMeta{
		ProviderMetadata: map[string]any{
			"nautobot": map[string]any{"role": "spine"},
		},
	}
	if got := m.GetRole(); got != "spine" {
		t.Errorf("GetRole() = %q, want %q", got, "spine")
	}
}

// TestGetRoleEmpty verifies GetRole returns an empty string when neither an
// explicit role nor a provider-metadata role is present.
//
// Why it matters: callers treat an empty role as "unassigned", so the accessor
// must not invent a value or panic on non-map metadata.
// Inputs: an ObjectMeta with empty Role and a provider bucket that is a string
// (not a map) and a map bucket without a "role" key. Outputs: "".
// Data choice: the non-map value exercises the type-assertion guard and the
// keyless map exercises the missing-key path, covering both skip branches.
func TestGetRoleEmpty(t *testing.T) {
	m := ObjectMeta{
		ProviderMetadata: map[string]any{
			"scalar": "not-a-map",
			"empty":  map[string]any{"other": "value"},
		},
	}
	if got := m.GetRole(); got != "" {
		t.Errorf("GetRole() = %q, want empty string", got)
	}
}
