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
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
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
		{"vrf", interfaceSpec{Name: "1/1/25", VRF: "LEGACY"}, true},
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
	var empty nautobotapi.PatchedWritableInterfaceRequest
	setPatchedInterfaceMode(&empty, "")
	if empty.Mode != nil {
		t.Errorf("setPatchedInterfaceMode(\"\") Mode = %v, want nil", empty.Mode)
	}
	var tagged nautobotapi.PatchedWritableInterfaceRequest
	setPatchedInterfaceMode(&tagged, "tagged")
	if tagged.Mode == nil {
		t.Error("setPatchedInterfaceMode(\"tagged\") Mode = nil, want non-nil")
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

	req, changed, unresolved, err := e.buildInterfaceEnrichment(deviceID, spec, vidToVLAN)
	if err != nil {
		t.Fatalf("buildInterfaceEnrichment() error = %v", err)
	}
	if !changed {
		t.Fatal("buildInterfaceEnrichment: changed = false, want true")
	}
	if unresolved != 0 {
		t.Errorf("buildInterfaceEnrichment: unresolved = %d, want 0", unresolved)
	}
	blob, _ := json.Marshal(req)
	for _, want := range []string{lagID.String(), vlanID.String(), "tagged"} {
		if !strings.Contains(string(blob), want) {
			t.Errorf("enrichment payload missing %q:\n%s", want, blob)
		}
	}
}

// TestBuildInterfaceEnrichment_CountsUnresolvedRefs verifies that references
// which cannot be resolved (LAG, untagged/tagged VLANs, VRF) are reported via
// the unresolved count instead of being silently dropped.
//
// Why it matters: FORGE-216 warns on and skips dangling references at export;
// the count is what drives the aggregate warning and the LoadResult tally, so an
// operator can tell that fabric state was left unapplied.
// Inputs: a spec referencing a missing LAG, an unmapped untagged VLAN, two
// unmapped tagged VLANs, and an uncached VRF, against an empty cache/VID map.
// Outputs: changed=false (nothing resolved) and unresolved=5 (1 LAG + 1 untagged
// + 2 tagged + 1 VRF).
// Data choice: one unresolved item per reference kind proves each path counts.
func TestBuildInterfaceEnrichment_CountsUnresolvedRefs(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{}`))
	defer cleanup()

	deviceID := uuid.New()
	spec := interfaceSpec{
		Name:         "1/1/2",
		Lag:          "missing-lag",
		UntaggedVLAN: 999,
		TaggedVLANs:  []int{10, 20},
		VRF:          "GONE",
	}

	_, changed, unresolved, err := e.buildInterfaceEnrichment(deviceID, spec, map[int]uuid.UUID{})
	if err != nil {
		t.Fatalf("buildInterfaceEnrichment() error = %v", err)
	}
	if changed {
		t.Error("buildInterfaceEnrichment: changed = true, want false (nothing resolvable)")
	}
	if unresolved != 5 {
		t.Errorf("buildInterfaceEnrichment: unresolved = %d, want 5", unresolved)
	}
}

// TestEnrichOneInterface_AccumulatesUnresolvedRefs verifies enrichOneInterface
// records unresolved references into LoadResult.IfacesUnresolvedRefs.
//
// Why it matters: the export summary reports this tally; if enrichOneInterface
// dropped the count, operators would get no signal that LAG/VLAN state was
// skipped.
// Inputs: a cached target interface whose spec references a missing LAG and one
// unmapped tagged VLAN, an empty VID map, and a fresh LoadResult. Outputs:
// IfacesUnresolvedRefs = 2 (nothing resolved, so no PATCH is sent).
// Data choice: the target interface is cached so the flow reaches the build step
// rather than erroring on a missing interface.
func TestEnrichOneInterface_AccumulatesUnresolvedRefs(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{}`))
	defer cleanup()

	deviceID := uuid.New()
	e.Cache.CacheInterface(deviceID, "1/1/3", &CachedItem{ID: uuid.New(), Name: "1/1/3"})

	spec := interfaceSpec{Name: "1/1/3", Lag: "missing-lag", TaggedVLANs: []int{30}}
	result := &LoadResult{}

	e.enrichOneInterface(context.Background(), deviceID, spec, map[int]uuid.UUID{}, result)

	if result.IfacesUnresolvedRefs != 2 {
		t.Errorf("IfacesUnresolvedRefs = %d, want 2", result.IfacesUnresolvedRefs)
	}
}

// TestBuildInterfaceEnrichment_PreservesRoleAndDescription verifies the
// enrichment PATCH carries the interface role and description so that adding
// LAG/VRF settings does not clear fields set during initial creation.
func TestBuildInterfaceEnrichment_PreservesRoleAndDescription(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{}`))
	defer cleanup()

	deviceID := uuid.New()
	roleID := uuid.New()
	e.Cache.roles["UplinkInterface"] = &CachedItem{ID: roleID, Name: "UplinkInterface"}

	lagID := uuid.New()
	e.Cache.CacheInterface(deviceID, "lag256", &CachedItem{ID: lagID, Name: "lag256"})

	spec := interfaceSpec{
		Name:        "1/1/49",
		Lag:         "lag256",
		Role:        "UplinkInterface",
		Description: "ISL member",
	}

	req, changed, _, err := e.buildInterfaceEnrichment(deviceID, spec, map[int]uuid.UUID{})
	if err != nil {
		t.Fatalf("buildInterfaceEnrichment() error = %v", err)
	}
	if !changed {
		t.Fatal("buildInterfaceEnrichment: changed = false, want true")
	}

	blob, _ := json.Marshal(req)
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(blob, &payload); err != nil {
		t.Fatalf("unmarshal enrichment payload: %v", err)
	}
	if role := string(payload["role"]); role == "" || role == "null" || !strings.Contains(role, roleID.String()) {
		t.Errorf("enrichment role = %s, want reference to %s", role, roleID)
	}
	if device := string(payload["device"]); device == "" || device == "null" || !strings.Contains(device, deviceID.String()) {
		t.Errorf("enrichment device = %s, want reference to %s", device, deviceID)
	}
	if description := string(payload["description"]); description != `"ISL member"` {
		t.Errorf("enrichment description = %s, want %q", description, "ISL member")
	}
}

// TestBuildInterfaceEnrichment_RejectsUnresolvedRole verifies that a role which
// cannot be resolved aborts enrichment before a nil role can clear Nautobot.
func TestBuildInterfaceEnrichment_RejectsUnresolvedRole(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{}`))
	defer cleanup()

	deviceID := uuid.New()
	lagID := uuid.New()
	e.Cache.CacheInterface(deviceID, "lag256", &CachedItem{ID: lagID, Name: "lag256"})

	// Lag resolves (so changed=true), but the role is not in the cache.
	spec := interfaceSpec{Name: "1/1/49", Lag: "lag256", Role: "GhostRole"}

	_, changed, unresolved, err := e.buildInterfaceEnrichment(deviceID, spec, map[int]uuid.UUID{})
	if changed {
		t.Fatal("buildInterfaceEnrichment: changed = true, want false for unresolved role")
	}
	if unresolved != 1 {
		t.Errorf("unresolved = %d, want 1 (dangling role)", unresolved)
	}
	if err == nil {
		t.Fatal("buildInterfaceEnrichment() error = nil, want unresolved role error")
	}
}

// TestBuildInterfaceEnrichment_ClearsEmptyRoleOnWire verifies that an empty
// local role serializes as an intentional null while the required device FK is
// retained; an empty description remains omitted.
func TestBuildInterfaceEnrichment_ClearsEmptyRoleOnWire(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{}`))
	defer cleanup()

	deviceID := uuid.New()
	lagID := uuid.New()
	e.Cache.CacheInterface(deviceID, "lag256", &CachedItem{ID: lagID, Name: "lag256"})

	spec := interfaceSpec{
		Name: "1/1/49",
		Lag:  "lag256",
	}

	req, changed, _, err := e.buildInterfaceEnrichment(deviceID, spec, map[int]uuid.UUID{})
	if err != nil {
		t.Fatalf("buildInterfaceEnrichment() error = %v", err)
	}
	if !changed {
		t.Fatal("buildInterfaceEnrichment: changed = false, want true")
	}

	if req.Role != nil {
		t.Error("expected req.Role to be nil when spec.Role is empty")
	}
	if req.Description != nil {
		t.Error("expected req.Description to be nil when spec.Description is empty")
	}
	blob, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal enrichment payload: %v", err)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(blob, &payload); err != nil {
		t.Fatalf("unmarshal enrichment payload: %v", err)
	}
	if role := string(payload["role"]); role != "null" {
		t.Errorf("enrichment role = %s, want null clear", role)
	}
	if device := string(payload["device"]); device == "" || device == "null" || !strings.Contains(device, deviceID.String()) {
		t.Errorf("enrichment device = %s, want reference to %s", device, deviceID)
	}
	if _, exists := payload["description"]; exists {
		t.Errorf("enrichment description should be omitted, got %s", payload["description"])
	}
}
