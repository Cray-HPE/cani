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
	"strings"
	"testing"

	openapi_types "github.com/Cray-HPE/cani/internal/openapi/types"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// -----------------------------------------------------------------------------
// buildLocationPatch — pure logic
// -----------------------------------------------------------------------------

func TestBuildLocationPatch_DetectsDescriptionDrift(t *testing.T) {
	loc := &devicetypes.CaniLocationType{
		ID:          uuid.New(),
		Name:        "DC1",
		Description: "new desc",
	}
	remoteDesc := "old desc"
	remote := &nautobotapi.Location{Name: "DC1", Description: &remoteDesc}

	req, drifted := buildLocationPatch(loc, remote)
	if !drifted {
		t.Fatal("expected drift for changed description")
	}
	if req.Description == nil || *req.Description != "new desc" {
		t.Errorf("Description = %v, want 'new desc'", req.Description)
	}
}

func TestBuildLocationPatch_DetectsFacilityDrift(t *testing.T) {
	loc := &devicetypes.CaniLocationType{
		ID:       uuid.New(),
		Name:     "DC1",
		Facility: "fac-2",
	}
	remoteFac := "fac-1"
	remote := &nautobotapi.Location{Name: "DC1", Facility: &remoteFac}

	req, drifted := buildLocationPatch(loc, remote)
	if !drifted {
		t.Fatal("expected drift for changed facility")
	}
	if req.Facility == nil || *req.Facility != "fac-2" {
		t.Errorf("Facility = %v, want 'fac-2'", req.Facility)
	}
}

func TestBuildLocationPatch_DetectsContactDrift(t *testing.T) {
	loc := &devicetypes.CaniLocationType{
		ID:           uuid.New(),
		Name:         "DC1",
		ContactName:  "Alice",
		ContactPhone: "555-0100",
		ContactEmail: "alice@example.com",
	}
	oldName := "Bob"
	oldPhone := "555-0000"
	oldEmail := openapi_types.Email("bob@example.com")
	remote := &nautobotapi.Location{
		Name:         "DC1",
		ContactName:  &oldName,
		ContactPhone: &oldPhone,
		ContactEmail: &oldEmail,
	}

	req, drifted := buildLocationPatch(loc, remote)
	if !drifted {
		t.Fatal("expected drift for changed contact fields")
	}
	if req.ContactName == nil || *req.ContactName != "Alice" {
		t.Errorf("ContactName = %v, want 'Alice'", req.ContactName)
	}
	if req.ContactPhone == nil || *req.ContactPhone != "555-0100" {
		t.Errorf("ContactPhone = %v, want '555-0100'", req.ContactPhone)
	}
	if req.ContactEmail == nil || *req.ContactEmail != "alice@example.com" {
		t.Errorf("ContactEmail = %v, want 'alice@example.com'", req.ContactEmail)
	}
}

func TestBuildLocationPatch_DetectsCustomFieldDrift(t *testing.T) {
	loc := &devicetypes.CaniLocationType{
		ID:   uuid.New(),
		Name: "DC1",
		ObjectMeta: devicetypes.ObjectMeta{
			CustomFields: map[string]any{"tier": "gold"},
		},
	}
	remoteCF := map[string]interface{}{"tier": "silver"}
	remote := &nautobotapi.Location{Name: "DC1", CustomFields: nbCF(remoteCF)}

	req, drifted := buildLocationPatch(loc, remote)
	if !drifted {
		t.Fatal("expected drift for changed custom fields")
	}
	if req.CustomFields == nil {
		t.Fatal("CustomFields patch should not be nil")
	}
	if derefCustomFields(req.CustomFields)["tier"] != "gold" {
		t.Errorf("CustomFields[tier] = %v, want 'gold'", derefCustomFields(req.CustomFields)["tier"])
	}
}

func TestBuildLocationPatch_NoDriftWhenIdentical(t *testing.T) {
	loc := &devicetypes.CaniLocationType{
		ID:          uuid.New(),
		Name:        "DC1",
		Description: "same",
		Facility:    "fac-1",
	}
	desc := "same"
	fac := "fac-1"
	remote := &nautobotapi.Location{Name: "DC1", Description: &desc, Facility: &fac}

	_, drifted := buildLocationPatch(loc, remote)
	if drifted {
		t.Error("expected no drift when all fields match")
	}
}

func TestBuildLocationPatch_EmptyLocalClearsRemote(t *testing.T) {
	loc := &devicetypes.CaniLocationType{
		ID:   uuid.New(),
		Name: "DC1",
		// All optional fields empty — should converge remote to empty
	}
	desc := "remote has desc"
	remote := &nautobotapi.Location{Name: "DC1", Description: &desc}

	req, drifted := buildLocationPatch(loc, remote)
	if !drifted {
		t.Error("should detect drift when local is empty but remote is not")
	}
	if req.Description == nil || *req.Description != "" {
		t.Errorf("Description = %v, want empty string to clear remote", req.Description)
	}
}

// -----------------------------------------------------------------------------
// buildVLANPatch — pure logic
// -----------------------------------------------------------------------------

func TestBuildVLANPatch_DetectsNameDrift(t *testing.T) {
	vlan := &devicetypes.CaniVLAN{ID: uuid.New(), VID: 100, Name: "new-name"}
	remote := &nautobotapi.VLAN{Name: "old-name", Vid: 100}

	req, drifted := buildVLANPatch(vlan, remote)
	if !drifted {
		t.Fatal("expected drift for changed name")
	}
	if req.Name == nil || *req.Name != "new-name" {
		t.Errorf("Name = %v, want 'new-name'", req.Name)
	}
}

func TestBuildVLANPatch_DetectsDescriptionDrift(t *testing.T) {
	vlan := &devicetypes.CaniVLAN{ID: uuid.New(), VID: 200, Name: "mgmt", Description: "updated"}
	oldDesc := "original"
	remote := &nautobotapi.VLAN{Name: "mgmt", Vid: 200, Description: &oldDesc}

	req, drifted := buildVLANPatch(vlan, remote)
	if !drifted {
		t.Fatal("expected drift for changed description")
	}
	if req.Description == nil || *req.Description != "updated" {
		t.Errorf("Description = %v, want 'updated'", req.Description)
	}
}

func TestBuildVLANPatch_DetectsCustomFieldDrift(t *testing.T) {
	vlan := &devicetypes.CaniVLAN{
		ID:   uuid.New(),
		VID:  300,
		Name: "storage",
		ObjectMeta: devicetypes.ObjectMeta{
			CustomFields: map[string]any{"priority": "high"},
		},
	}
	remoteCF := map[string]interface{}{"priority": "low"}
	remote := &nautobotapi.VLAN{Name: "storage", Vid: 300, CustomFields: nbCF(remoteCF)}

	req, drifted := buildVLANPatch(vlan, remote)
	if !drifted {
		t.Fatal("expected drift for changed custom fields")
	}
	if req.CustomFields == nil {
		t.Fatal("CustomFields patch should not be nil")
	}
}

func TestBuildVLANPatch_NoDriftWhenIdentical(t *testing.T) {
	vlan := &devicetypes.CaniVLAN{ID: uuid.New(), VID: 100, Name: "mgmt", Description: "same"}
	desc := "same"
	remote := &nautobotapi.VLAN{Name: "mgmt", Vid: 100, Description: &desc}

	_, drifted := buildVLANPatch(vlan, remote)
	if drifted {
		t.Error("expected no drift when all fields match")
	}
}

// TestMergePatchesSendTheProjectionAndNameACollision verifies that the
// location and VLAN merge paths send the same custom-field projection a
// create sends: on a key set both explicitly and by provider metadata the
// provider value is patched and the collision is logged by object and key,
// never by value.
//
// Why it matters: the merge paths used to take the projection's values
// without its collision report, so a key an operator set two ways was
// resolved silently on --merge and named only on create.
// Inputs: a location and a VLAN whose explicit custom field "owner" and
// provider metadata "owner" disagree, each against a remote that holds a
// third value.
// Outputs: the patch carries the provider value; the log names the object
// and the key and contains neither value.
// Data choice: three distinct values make the precedence visible in the
// patch and prove the log carries none of them.
func TestMergePatchesSendTheProjectionAndNameACollision(t *testing.T) {
	meta := func() devicetypes.ObjectMeta {
		return devicetypes.ObjectMeta{
			CustomFields:     map[string]any{"owner": "explicit-team"},
			ProviderMetadata: map[string]any{"nautobot": map[string]any{"owner": "provider-team"}},
		}
	}
	cases := map[string]struct {
		patch    func() *map[string]*interface{}
		wantLine string
	}{
		"location": {
			patch: func() *map[string]*interface{} {
				loc := &devicetypes.CaniLocationType{ID: uuid.New(), Name: "DC1", ObjectMeta: meta()}
				req, _ := buildLocationPatch(loc, &nautobotapi.Location{Name: "DC1", CustomFields: nbCF(map[string]interface{}{"owner": "remote-team"})})
				return req.CustomFields
			},
			wantLine: `location DC1: custom field "owner" is set both explicitly and by provider metadata`,
		},
		"VLAN": {
			patch: func() *map[string]*interface{} {
				vlan := &devicetypes.CaniVLAN{ID: uuid.New(), VID: 300, Name: "storage", ObjectMeta: meta()}
				req, _ := buildVLANPatch(vlan, &nautobotapi.VLAN{Name: "storage", Vid: 300, CustomFields: nbCF(map[string]interface{}{"owner": "remote-team"})})
				return req.CustomFields
			},
			wantLine: `VLAN storage: custom field "owner" is set both explicitly and by provider metadata`,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			logged := captureClog(t)

			sent := derefCustomFields(tc.patch())

			if sent["owner"] != "provider-team" {
				t.Errorf("patched owner = %v, want the provider value", sent["owner"])
			}
			if !strings.Contains(logged.String(), tc.wantLine) {
				t.Errorf("log = %q, want it to contain %q", logged.String(), tc.wantLine)
			}
			for _, value := range []string{"explicit-team", "provider-team", "remote-team"} {
				if strings.Contains(logged.String(), value) {
					t.Errorf("log names the value %q; a finding must name only the object and the key", value)
				}
			}
		})
	}
}

// -----------------------------------------------------------------------------
// mergedCustomFields — pure logic
// -----------------------------------------------------------------------------

func TestMergedCustomFields_CombinesBothMaps(t *testing.T) {
	explicit := map[string]any{"a": "1"}
	flat := map[string]any{"b": "2"}
	merged := mergedCustomFields(explicit, flat)
	if merged["a"] != "1" || merged["b"] != "2" {
		t.Errorf("merged = %v, want {a:1, b:2}", merged)
	}
}

func TestMergedCustomFields_FlatOverridesExplicit(t *testing.T) {
	explicit := map[string]any{"key": "explicit"}
	flat := map[string]any{"key": "flat"}
	merged := mergedCustomFields(explicit, flat)
	if merged["key"] != "flat" {
		t.Errorf("merged[key] = %v, want 'flat'", merged["key"])
	}
}

func TestMergedCustomFields_EmptyInputs(t *testing.T) {
	merged := mergedCustomFields(nil, nil)
	if len(merged) != 0 {
		t.Errorf("expected empty map, got %v", merged)
	}
}

// -----------------------------------------------------------------------------
// customFieldsDrifted — pure logic
// -----------------------------------------------------------------------------

func TestCustomFieldsDrifted_TrueWhenDifferent(t *testing.T) {
	local := map[string]interface{}{"tier": "gold"}
	remote := map[string]interface{}{"tier": "silver"}
	if !customFieldsDrifted(local, nbCF(remote)) {
		t.Error("expected true when values differ")
	}
}

func TestCustomFieldsDrifted_TrueWhenNewKey(t *testing.T) {
	local := map[string]interface{}{"env": "prod"}
	remote := map[string]interface{}{}
	if !customFieldsDrifted(local, nbCF(remote)) {
		t.Error("expected true when local has new key")
	}
}

func TestCustomFieldsDrifted_FalseWhenEqual(t *testing.T) {
	local := map[string]interface{}{"tier": "gold"}
	remote := map[string]interface{}{"tier": "gold"}
	if customFieldsDrifted(local, nbCF(remote)) {
		t.Error("expected false when values match")
	}
}

func TestCustomFieldsDrifted_TrueWhenRemoteNil(t *testing.T) {
	local := map[string]interface{}{"tier": "gold"}
	if !customFieldsDrifted(local, nil) {
		t.Error("expected true when remote is nil")
	}
}
