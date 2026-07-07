/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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
package export

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// TestInterfaceNeedsEnrichment verifies that only interfaces carrying LAG, mode,
// or VLAN settings are selected for the enrichment pass.
//
// Why it matters: the enrichment pass issues an extra PATCH per interface;
// selecting a plain interface would waste an API call, while missing a
// configured one would drop LAG/VLAN state from the source of truth.
// Inputs: a plain spec plus one spec per enrichable field. Outputs: false for
// the plain spec, true for each configured spec.
// Data choice: one case per field ensures every predicate term is covered.
func TestInterfaceNeedsEnrichment(t *testing.T) {
	cases := []struct {
		name string
		spec interfaceSpec
		want bool
	}{
		{"plain", interfaceSpec{Name: "eth0"}, false},
		{"lag", interfaceSpec{Name: "1/1/1", Lag: "lag256"}, true},
		{"mode", interfaceSpec{Name: "1/1/1", Mode: "tagged"}, true},
		{"untagged", interfaceSpec{Name: "1/1/1", UntaggedVLAN: 2000}, true},
		{"tagged", interfaceSpec{Name: "1/1/1", TaggedVLANs: []int{1}}, true},
	}
	for _, c := range cases {
		if got := interfaceNeedsEnrichment(c.spec); got != c.want {
			t.Errorf("%s: interfaceNeedsEnrichment = %v, want %v", c.name, got, c.want)
		}
	}
}

// TestBuildVIDMap verifies VLAN IDs are mapped to the Nautobot UUIDs produced by
// the VLAN phase.
//
// Why it matters: interface VLAN assignment references VLANs by Nautobot UUID,
// but inventory expresses membership by VLAN ID; a wrong mapping would attach an
// interface to the wrong (or no) VLAN.
// Inputs: one inventory VLAN (VID 2000) and a cani-to-Nautobot ID map. Outputs:
// a map whose VID 2000 entry is the Nautobot UUID.
// Data choice: VID 2000 mirrors the legacy CSM VLAN used on real trunks.
func TestBuildVIDMap(t *testing.T) {
	caniVLANID := uuid.New()
	nautobotVLANID := uuid.New()
	inv := &devicetypes.Inventory{
		VLANs: map[uuid.UUID]*devicetypes.CaniVLAN{
			caniVLANID: {ID: caniVLANID, VID: 2000, Name: "Legacy"},
		},
	}
	created := map[uuid.UUID]uuid.UUID{caniVLANID: nautobotVLANID}

	m := buildVIDMap(inv, created)
	if m[2000] != nautobotVLANID {
		t.Errorf("buildVIDMap[2000] = %s, want %s", m[2000], nautobotVLANID)
	}
}

// TestModeRef verifies switchport-mode reference construction.
//
// Why it matters: an empty mode must leave the field unset (nil) so the PATCH
// does not clear an interface's mode, while a valid mode must produce a non-nil
// reference for the wire payload.
// Inputs: the empty string and "tagged". Outputs: nil for empty, non-nil for
// "tagged".
// Data choice: covers the unset and set branches that drive the enrichment
// "changed" flag.
func TestModeRef(t *testing.T) {
	if ref := modeRef(""); ref != nil {
		t.Errorf("modeRef(\"\") = %v, want nil", ref)
	}
	if ref := modeRef("tagged"); ref == nil {
		t.Error("modeRef(\"tagged\") = nil, want non-nil")
	}
}

// TestBuildInterfaceEnrichment_SetsLagModeAndVLANs verifies the enrichment PATCH
// carries the LAG parent, switchport mode, and untagged/tagged VLAN references.
//
// Why it matters: this is the payload that records aggregation and trunk
// membership in Nautobot; missing any of these would leave the fabric interface
// misconfigured in the source of truth.
// Inputs: a member interface spec referencing a cached LAG and VLAN 2000, with a
// VID-to-VLAN map resolving 2000. Outputs: changed=true and a payload containing
// the LAG UUID, the VLAN UUID, and the mode "tagged".
// Data choice: a tagged member joining lag256 with VLAN 2000 mirrors a real VSX
// trunk port, exercising every enrichment field at once.
func TestBuildInterfaceEnrichment_SetsLagModeAndVLANs(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{}`))
	defer cleanup()

	deviceID := uuid.New()
	lagID := uuid.New()
	e.Cache.CacheInterface(deviceID, "lag256", &CachedItem{ID: lagID, Name: "lag256"})

	vlanID := uuid.New()
	vidToVLAN := map[int]uuid.UUID{2000: vlanID}

	spec := interfaceSpec{
		Name:         "1/1/1",
		Lag:          "lag256",
		Mode:         "tagged",
		UntaggedVLAN: 2000,
		TaggedVLANs:  []int{2000},
	}

	req, changed := e.buildInterfaceEnrichment(deviceID, spec, vidToVLAN)
	if !changed {
		t.Fatal("buildInterfaceEnrichment: changed = false, want true")
	}
	blob, _ := json.Marshal(req)
	for _, want := range []string{lagID.String(), vlanID.String(), "tagged"} {
		if !strings.Contains(string(blob), want) {
			t.Errorf("enrichment payload missing %q:\n%s", want, blob)
		}
	}
}
