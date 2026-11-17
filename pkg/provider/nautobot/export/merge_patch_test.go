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

	req, drifted := buildLocationPatch(loc, remote, nil)
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

	req, drifted := buildLocationPatch(loc, remote, nil)
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

	req, drifted := buildLocationPatch(loc, remote, nil)
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

	req, drifted := buildLocationPatch(loc, remote, nil)
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

	_, drifted := buildLocationPatch(loc, remote, nil)
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

	req, drifted := buildLocationPatch(loc, remote, nil)
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

	req, drifted := buildVLANPatch(vlan, remote, nil)
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

	req, drifted := buildVLANPatch(vlan, remote, nil)
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

	req, drifted := buildVLANPatch(vlan, remote, nil)
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

	_, drifted := buildVLANPatch(vlan, remote, nil)
	if drifted {
		t.Error("expected no drift when all fields match")
	}
}

// TestMergePatchesSendTheProjection verifies that the location and VLAN
// merge paths send the same custom-field projection a create sends: the
// object's own custom field, unless the operator listed a provider key for
// it, in which case the provider's value is patched and the replaced custom
// field is noted by object, key and provider, never by value.
//
// Why it matters: a merge that projected differently from a create would
// patch one value and compare another, so every --merge would report a
// change; and a note that printed a value could leak a credential.
// Inputs: a location and a VLAN whose custom field "owner" and CSM
// metadata "owner" disagree, each against a remote that holds a third
// value, merged with no listed key and then with csm.owner listed.
// Outputs: the patch carries the custom field, then the provider value; the
// note appears only in the second run and contains none of the values.
// Data choice: three distinct values make the precedence visible in the
// patch and prove the log carries none of them.
func TestMergePatchesSendTheProjection(t *testing.T) {
	meta := func() devicetypes.ObjectMeta {
		return devicetypes.ObjectMeta{
			CustomFields:     map[string]any{"owner": "own-team"},
			ProviderMetadata: map[string]any{"csm": map[string]any{"owner": "provider-team"}},
		}
	}
	remoteFields := nbCF(map[string]interface{}{"owner": "remote-team"})
	listed := customFieldProjection{{Provider: "csm", Key: "owner", CustomField: "owner"}}
	patches := map[string]struct {
		patch    func(fields customFieldProjection) *map[string]*interface{}
		wantNote string
	}{
		"location": {
			patch: func(fields customFieldProjection) *map[string]*interface{} {
				loc := &devicetypes.CaniLocationType{ID: uuid.New(), Name: "DC1", ObjectMeta: meta()}
				req, _ := buildLocationPatch(loc, &nautobotapi.Location{Name: "DC1", CustomFields: remoteFields}, fields)
				return req.CustomFields
			},
			wantNote: `location DC1: custom field "owner" comes from csm provider metadata`,
		},
		"VLAN": {
			patch: func(fields customFieldProjection) *map[string]*interface{} {
				vlan := &devicetypes.CaniVLAN{ID: uuid.New(), VID: 300, Name: "storage", ObjectMeta: meta()}
				req, _ := buildVLANPatch(vlan, &nautobotapi.VLAN{Name: "storage", Vid: 300, CustomFields: remoteFields}, fields)
				return req.CustomFields
			},
			wantNote: `VLAN storage: custom field "owner" comes from csm provider metadata`,
		},
	}
	for name, tc := range patches {
		t.Run(name, func(t *testing.T) {
			// Arrange.
			logged := captureClog(t)

			// Act.
			own := derefCustomFields(tc.patch(nil))
			unlistedLog := logged.String()
			fromProvider := derefCustomFields(tc.patch(listed))

			// Assert.
			if own["owner"] != "own-team" || strings.Contains(unlistedLog, "provider metadata") {
				t.Errorf("no key listed: patched owner = %v, log = %q; want the custom field and no note", own["owner"], unlistedLog)
			}
			if fromProvider["owner"] != "provider-team" {
				t.Errorf("csm.owner listed: patched owner = %v, want the provider value", fromProvider["owner"])
			}
			if !strings.Contains(logged.String(), tc.wantNote) {
				t.Errorf("log = %q, want it to contain %q", logged.String(), tc.wantNote)
			}
			for _, value := range []string{"own-team", "provider-team", "remote-team"} {
				if strings.Contains(logged.String(), value) {
					t.Errorf("log names the value %q; a note must name only the object, the key and the provider", value)
				}
			}
		})
	}
}

// TestMergeComparisonsNoteAReplacedCustomFieldWithoutAWrite verifies that
// each merge comparison names a custom field a listed provider key replaces
// even when the remote already holds the provider's value and nothing is
// patched, and that a device update that follows does not name it twice.
//
// Why it matters: with Nautobot already holding the provider's value there
// is nothing to send, so a custom field an operator had just set was passed
// over without a word. The note used to be made only while a write was
// built; an export against a live Nautobot showed the gap.
// Inputs: a location, a VLAN and a device whose custom field "owner" says
// own-team and whose CSM metadata says provider-team, each compared with a
// remote that holds provider-team, with csm.owner listed.
// Outputs: no custom-field change to send, and one note per object, still
// one after the device's patch request is built.
// Data choice: the remote equals the provider value, the one case in which
// the comparison finds nothing to write.
func TestMergeComparisonsNoteAReplacedCustomFieldWithoutAWrite(t *testing.T) {
	// Arrange.
	logged := captureClog(t)
	e, cleanup := newExporterWithServer(t, jsonHandler(new(int), 200, `{}`))
	defer cleanup()
	seedDeviceRefs(t, e)
	listed := customFieldProjection{{Provider: "csm", Key: "owner", CustomField: "owner"}}
	meta := devicetypes.ObjectMeta{
		CustomFields:     map[string]any{"owner": "own-team"},
		ProviderMetadata: map[string]any{"csm": map[string]any{"owner": "provider-team"}},
	}
	remoteFields := nbCF(map[string]interface{}{"owner": "provider-team"})
	loc := &devicetypes.CaniLocationType{ID: uuid.New(), Name: "DC1", ObjectMeta: meta}
	vlan := &devicetypes.CaniVLAN{ID: uuid.New(), VID: 300, Name: "storage", ObjectMeta: meta}
	dev := newMappableDevice("compute-010")
	dev.CustomFields, dev.ProviderMetadata = meta.CustomFields, meta.ProviderMetadata
	mapper := NewDeviceMapper(e.Cache, &MapperOpts{
		DefaultLocation: "DC1", DefaultStatus: "Active", DefaultRole: "Compute", ProviderMetadataFields: listed,
	})

	// Act.
	locPatch, _ := buildLocationPatch(loc, &nautobotapi.Location{Name: "DC1", CustomFields: remoteFields}, listed)
	vlanPatch, _ := buildVLANPatch(vlan, &nautobotapi.VLAN{Name: "storage", Vid: 300, CustomFields: remoteFields}, listed)
	diffs := compareDeviceFields(dev, &nautobotapi.Device{Name: &dev.Name, CustomFields: remoteFields}, mapper)
	_, patchErr := mapper.MapToPatchRequest(dev, uuid.New())

	// Assert.
	if patchErr != nil {
		t.Fatalf("MapToPatchRequest: %v", patchErr)
	}
	if locPatch.CustomFields != nil || vlanPatch.CustomFields != nil {
		t.Errorf("patched custom fields = %v, %v; want none, the remote already holds the provider value",
			locPatch.CustomFields, vlanPatch.CustomFields)
	}
	for _, d := range diffs {
		if strings.HasPrefix(d.Field, "custom_field:") {
			t.Errorf("unexpected device diff %s: %q over %q", d.Field, d.LocalVal, d.RemoteVal)
		}
	}
	for _, object := range []string{"location DC1", "VLAN storage", "device compute-010"} {
		note := object + `: custom field "owner" comes from csm provider metadata`
		if got := strings.Count(logged.String(), note); got != 1 {
			t.Errorf("%q noted %d times, want once; log = %q", object, got, logged.String())
		}
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
