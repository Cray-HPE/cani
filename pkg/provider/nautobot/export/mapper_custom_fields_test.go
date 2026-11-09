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
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// explicitCustomFieldDevice is a mappable device with one explicit custom
// field of each falsy kind and one provider-metadata key.
func explicitCustomFieldDevice(name string) *devicetypes.CaniDeviceType {
	dev := newMappableDevice(name)
	dev.CustomFields = map[string]any{"decommissioned": false, "spare_slots": 0, "tier": "gold"}
	dev.ProviderMetadata = map[string]any{"csm": map[string]any{"xname": "x3000c0s1b0"}}
	return dev
}

// wantDeviceCustomFields is what every device mapper must send for
// explicitCustomFieldDevice.
var wantDeviceCustomFields = map[string]any{"decommissioned": false, "spare_slots": 0, "tier": "gold", "xname": "x3000c0s1b0"}

// assertCustomFields fails unless got carries exactly the want keys with the
// want values.
func assertCustomFields(t *testing.T, got *map[string]*interface{}, want map[string]any) {
	t.Helper()
	if got == nil {
		t.Fatalf("CustomFields = nil, want %v", want)
	}
	if len(*got) != len(want) {
		t.Errorf("CustomFields has %d key(s), want %d: %v", len(*got), len(want), derefCustomFields(got))
	}
	for key, value := range want {
		p, ok := (*got)[key]
		if !ok || p == nil || *p != value {
			t.Errorf("CustomFields[%q] = %v, want %v", key, p, value)
		}
	}
}

// TestMapToWritableDeviceRequest_SendsExplicitCustomFields verifies that
// the single-create mapper carries explicit custom fields, including false
// and zero, alongside flattened provider metadata.
//
// Why it matters: before the shared projection the device create sent only
// provider metadata, so an inventory imported from one Nautobot lost every
// explicit custom field when seeded into another.
// Inputs: a mappable device with explicit {decommissioned: false,
// spare_slots: 0, tier} and provider {xname}.
// Outputs: custom_fields with all four keys and original values.
// Data choice: the falsy values prove nothing filters on truthiness.
func TestMapToWritableDeviceRequest_SendsExplicitCustomFields(t *testing.T) {
	e, cleanup := newExporterWithServer(t, jsonHandler(new(int), 200, `{}`))
	defer cleanup()
	seedDeviceRefs(t, e)

	req, err := newCrudMapper(e).MapToWritableDeviceRequest(explicitCustomFieldDevice("compute-001"))
	if err != nil {
		t.Fatalf("MapToWritableDeviceRequest() error = %v", err)
	}
	assertCustomFields(t, req.CustomFields, wantDeviceCustomFields)
}

// TestMapToNautobotDevice_SendsExplicitCustomFields verifies that the
// bulk-create mapper sends the same custom-field set as the single create.
//
// Why it matters: the Story requires one projection for every write; a bulk
// path that diverged would make fidelity depend on batch size.
// Inputs: the same device as the single-create test.
// Outputs: an identical custom_fields map.
// Data choice: reusing the fixture makes any divergence between the two
// mappers a direct assertion failure.
func TestMapToNautobotDevice_SendsExplicitCustomFields(t *testing.T) {
	e, cleanup := newExporterWithServer(t, jsonHandler(new(int), 200, `{}`))
	defer cleanup()
	seedDeviceRefs(t, e)

	req, err := newCrudMapper(e).MapToNautobotDevice(explicitCustomFieldDevice("compute-002"))
	if err != nil {
		t.Fatalf("MapToNautobotDevice() error = %v", err)
	}
	assertCustomFields(t, req.CustomFields, wantDeviceCustomFields)
}

// TestMapToPatchRequest_SendsExplicitCustomFields verifies that the merge
// PATCH mapper carries explicit custom fields.
//
// Why it matters: --merge is how a changed explicit custom field reaches an
// object that already exists; without it the create and merge paths send
// different sets, which the Story forbids.
// Inputs: the same device as the create tests and an arbitrary existing ID.
// Outputs: an identical custom_fields map on the PATCH request.
// Data choice: the existing ID is irrelevant to the payload and is random.
func TestMapToPatchRequest_SendsExplicitCustomFields(t *testing.T) {
	e, cleanup := newExporterWithServer(t, jsonHandler(new(int), 200, `{}`))
	defer cleanup()
	seedDeviceRefs(t, e)

	req, err := newCrudMapper(e).MapToPatchRequest(explicitCustomFieldDevice("compute-003"), uuid.New())
	if err != nil {
		t.Fatalf("MapToPatchRequest() error = %v", err)
	}
	assertCustomFields(t, req.CustomFields, wantDeviceCustomFields)
}

// TestMapToWritableRackRequest_SendsExplicitCustomFieldsButNotGeometry
// verifies that the legacy rack mapper carries explicit custom fields while
// still keeping u_height and rack_position out of custom_fields.
//
// Why it matters: rack geometry has dedicated Nautobot fields; sending it as
// a custom field would create spurious definitions, but the exclusion must
// not take explicit custom fields with it.
// Inputs: a rack-category device with explicit {owner} and provider
// {u_height: 42, rack_position: 1, color}.
// Outputs: custom_fields {owner, color} and UHeight 42.
// Data choice: one excluded key of each name next to one key that must
// survive from each source.
func TestMapToWritableRackRequest_SendsExplicitCustomFieldsButNotGeometry(t *testing.T) {
	e, cleanup := newExporterWithServer(t, jsonHandler(new(int), 200, `{}`))
	defer cleanup()
	seedDeviceRefs(t, e)

	rack := &devicetypes.CaniDeviceType{
		Name: "x3000",
		Type: devicetypes.Rack,
		ObjectMeta: devicetypes.ObjectMeta{
			Status:           "Active",
			CustomFields:     map[string]any{"owner": "ops"},
			ProviderMetadata: map[string]any{"u_height": 42, "rack_position": 1, "color": "black"},
		},
	}

	req, err := newCrudMapper(e).MapToWritableRackRequest(rack)
	if err != nil {
		t.Fatalf("MapToWritableRackRequest() error = %v", err)
	}
	assertCustomFields(t, req.CustomFields, map[string]any{"owner": "ops", "color": "black"})
	if req.UHeight == nil || *req.UHeight != 42 {
		t.Errorf("UHeight = %v, want 42", req.UHeight)
	}
}

// TestCompareDeviceFields_ReportsExplicitCustomFieldDrift verifies that the
// --merge comparison sees a changed explicit custom field.
//
// Why it matters: compareDeviceFields drove the merge decision from provider
// metadata alone, so an explicit value that differed remotely produced no
// diff and no PATCH; the compare must use the same projection as the write.
// Inputs: a device with explicit {tier: gold} and a remote device whose
// custom_fields say tier silver.
// Outputs: exactly one diff, for custom_field:tier, gold over silver.
// Data choice: no provider metadata at all, so the diff can only come from
// the explicit map.
func TestCompareDeviceFields_ReportsExplicitCustomFieldDrift(t *testing.T) {
	e, cleanup := newExporterWithServer(t, jsonHandler(new(int), 200, `{}`))
	defer cleanup()
	seedDeviceRefs(t, e)
	mapper := newCrudMapper(e)

	dev := newMappableDevice("compute-004")
	dev.CustomFields = map[string]any{"tier": "gold"}
	remote := &nautobotapi.Device{Name: &dev.Name, CustomFields: nbCF(map[string]any{"tier": "silver"})}

	var found *FieldDiff
	for _, d := range compareDeviceFields(dev, remote, mapper) {
		if d.Field == "custom_field:tier" {
			d := d
			found = &d
		}
	}
	if found == nil {
		t.Fatal("no diff for custom_field:tier; explicit custom fields are not compared")
	}
	if found.LocalVal != "gold" || found.RemoteVal != "silver" {
		t.Errorf("diff = %q over %q, want gold over silver", found.LocalVal, found.RemoteVal)
	}
}
