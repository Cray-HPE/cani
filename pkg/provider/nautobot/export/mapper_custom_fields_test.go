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

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/Cray-HPE/cani/pkg/provider/nautobot/transform"
	"github.com/google/uuid"
)

// customFieldDevice is a mappable device with one custom field of each falsy
// kind and a CSM bucket that holds an xname.
func customFieldDevice(name string) *devicetypes.CaniDeviceType {
	dev := newMappableDevice(name)
	dev.CustomFields = map[string]any{"decommissioned": false, "spare_slots": 0, "tier": "gold"}
	dev.ProviderMetadata = map[string]any{"csm": map[string]any{"xname": "x3000c0s1b0"}}
	return dev
}

// deviceRequests names the three device mappers; each returns the
// custom_fields of the request it builds.
var deviceRequests = map[string]func(*DeviceMapper, *devicetypes.CaniDeviceType) (*map[string]*interface{}, error){
	"single create": func(m *DeviceMapper, dev *devicetypes.CaniDeviceType) (*map[string]*interface{}, error) {
		req, err := m.MapToWritableDeviceRequest(dev)
		if err != nil {
			return nil, err
		}
		return req.CustomFields, nil
	},
	"bulk create": func(m *DeviceMapper, dev *devicetypes.CaniDeviceType) (*map[string]*interface{}, error) {
		req, err := m.MapToNautobotDevice(dev)
		if err != nil {
			return nil, err
		}
		return req.CustomFields, nil
	},
	"merge patch": func(m *DeviceMapper, dev *devicetypes.CaniDeviceType) (*map[string]*interface{}, error) {
		req, err := m.MapToPatchRequest(dev, uuid.New())
		if err != nil {
			return nil, err
		}
		return req.CustomFields, nil
	},
}

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

// TestDeviceMappers_SendCustomFieldsAndNoProviderMetadata verifies that
// every device mapper sends the device's custom fields, including false and
// zero, and none of its provider metadata.
//
// Why it matters: the three mappers are every way a device is written, so
// one that projected differently would make what Nautobot receives depend
// on batch size or on --merge; and provider metadata sent as custom fields
// is what let another provider's value overwrite a field set in Nautobot.
// Inputs: a mappable device with custom fields {decommissioned: false,
// spare_slots: 0, tier} and a CSM bucket {xname}, through each mapper.
// Outputs: custom_fields with the three custom fields and no xname.
// Data choice: the falsy values prove nothing filters on truthiness, and
// one fixture for all three makes any divergence a direct failure.
func TestDeviceMappers_SendCustomFieldsAndNoProviderMetadata(t *testing.T) {
	want := map[string]any{"decommissioned": false, "spare_slots": 0, "tier": "gold"}
	for name, request := range deviceRequests {
		t.Run(name, func(t *testing.T) {
			// Arrange.
			e, cleanup := newExporterWithServer(t, jsonHandler(new(int), 200, `{}`))
			defer cleanup()
			seedDeviceRefs(t, e)

			// Act.
			fields, err := request(newCrudMapper(e), customFieldDevice("compute-001"))

			// Assert.
			if err != nil {
				t.Fatalf("mapper error = %v", err)
			}
			assertCustomFields(t, fields, want)
		})
	}
}

// TestDeviceMappers_SendAListedProviderKey verifies that every device
// mapper sends a provider-metadata key the export options list, beside the
// device's own custom fields.
//
// Why it matters: the list is the only way a provider's metadata reaches
// Nautobot; a mapper that ignored it would drop the field from creates,
// bulk creates or merges without a word.
// Inputs: the same device, through each mapper built with csm.xname listed.
// Outputs: custom_fields with the three custom fields and the xname.
// Data choice: the fixture of the test above, so the listed key is the
// only difference between the two outcomes.
func TestDeviceMappers_SendAListedProviderKey(t *testing.T) {
	want := map[string]any{"decommissioned": false, "spare_slots": 0, "tier": "gold", "xname": "x3000c0s1b0"}
	for name, request := range deviceRequests {
		t.Run(name, func(t *testing.T) {
			// Arrange.
			e, cleanup := newExporterWithServer(t, jsonHandler(new(int), 200, `{}`))
			defer cleanup()
			seedDeviceRefs(t, e)
			mapper := NewDeviceMapper(e.Cache, &MapperOpts{
				DefaultLocation: "DC1", DefaultStatus: "Active", DefaultRole: "Compute",
				ProviderMetadataFields: []ProviderMetadataField{{Provider: "csm", Key: "xname", CustomField: "xname"}},
			})

			// Act.
			fields, err := request(mapper, customFieldDevice("compute-002"))

			// Assert.
			if err != nil {
				t.Fatalf("mapper error = %v", err)
			}
			assertCustomFields(t, fields, want)
		})
	}
}

// TestMapToWritableRackRequest_SendsCustomFieldsAndTheDevicesHeight verifies
// that a rack kept as a device is created with the device's own custom
// fields and height, and with nothing from its provider metadata.
//
// Why it matters: the rack mapper read a height from provider metadata that
// no stored inventory could supply, so every such rack was created 48U
// tall, and it sent the rest of the provider metadata as custom fields.
// Inputs: a rack-category device with a custom field, a height of 42 and
// provider metadata holding u_height, rack_position and a CSM bucket.
// Outputs: custom_fields {owner} and UHeight 42.
// Data choice: the provider-metadata height differs from the field, so
// reading it again would show.
func TestMapToWritableRackRequest_SendsCustomFieldsAndTheDevicesHeight(t *testing.T) {
	// Arrange.
	e, cleanup := newExporterWithServer(t, jsonHandler(new(int), 200, `{}`))
	defer cleanup()
	seedDeviceRefs(t, e)
	rack := &devicetypes.CaniDeviceType{
		Name:    "x3000",
		Type:    devicetypes.Rack,
		UHeight: 42,
		ObjectMeta: devicetypes.ObjectMeta{
			Status:       "Active",
			CustomFields: map[string]any{"owner": "ops"},
			ProviderMetadata: map[string]any{
				"u_height": 47, "rack_position": 1, "csm": map[string]any{"xname": "x3000"},
			},
		},
	}

	// Act.
	req, err := newCrudMapper(e).MapToWritableRackRequest(rack)

	// Assert.
	if err != nil {
		t.Fatalf("MapToWritableRackRequest() error = %v", err)
	}
	assertCustomFields(t, req.CustomFields, map[string]any{"owner": "ops"})
	if req.UHeight == nil {
		t.Fatal("UHeight is unset, want 42")
	}
	if *req.UHeight != 42 {
		t.Errorf("UHeight = %d, want 42", *req.UHeight)
	}
}

// TestCompareDeviceFields_ReportsCustomFieldDrift verifies that the --merge
// comparison sees a changed custom field.
//
// Why it matters: compareDeviceFields drives the merge decision, so a value
// that differs remotely must produce a diff and a PATCH; the compare has to
// use the same projection as the write.
// Inputs: a device with custom field {tier: gold} and a remote device whose
// custom_fields say tier silver.
// Outputs: exactly one diff, for custom_field:tier, gold over silver.
// Data choice: no provider metadata at all, so the diff can only come from
// the device's custom fields.
func TestCompareDeviceFields_ReportsCustomFieldDrift(t *testing.T) {
	// Arrange.
	e, cleanup := newExporterWithServer(t, jsonHandler(new(int), 200, `{}`))
	defer cleanup()
	seedDeviceRefs(t, e)
	mapper := newCrudMapper(e)
	dev := newMappableDevice("compute-004")
	dev.CustomFields = map[string]any{"tier": "gold"}
	remote := &nautobotapi.Device{Name: &dev.Name, CustomFields: nbCF(map[string]any{"tier": "silver"})}

	// Act.
	var found *FieldDiff
	for _, d := range compareDeviceFields(dev, remote, mapper) {
		if d.Field == "custom_field:tier" {
			d := d
			found = &d
		}
	}

	// Assert.
	if found == nil {
		t.Fatal("no diff for custom_field:tier; custom fields are not compared")
	}
	if found.LocalVal != "gold" || found.RemoteVal != "silver" {
		t.Errorf("diff = %q over %q, want gold over silver", found.LocalVal, found.RemoteVal)
	}
}

// providerRoleDevice is a device with no role field whose provider metadata
// holds a role and a location in two buckets and outside any.
func providerRoleDevice(name string) *devicetypes.CaniDeviceType {
	dev := newMappableDevice(name)
	dev.Role = ""
	named := map[string]any{"role": "Storage", "location": "DC9"}
	dev.ProviderMetadata = map[string]any{"csm": named, "nautobot": named, "role": "Storage", "location": "DC9"}
	return dev
}

// TestDeviceMappers_ProviderMetadataSetsNeitherRoleNorLocation verifies
// that every device mapper takes a device's role and location from the
// device and the export defaults, never from provider metadata, and sends
// neither as a custom field.
//
// Why it matters: a role or location in provider metadata used to override
// the inventory's own, so two places could set one Nautobot field and the
// one a command could not show won. The role field and the device's
// placement are now the only sources.
// Inputs: a device with no role field and no placement whose provider
// metadata names role Storage and location DC9, through each mapper with
// defaults Compute and DC1.
// Outputs: role Compute and location DC1 on each request, and no custom
// fields.
// Data choice: the names in provider metadata are ones the destination
// does not have, so a mapper that read them would fail or create them.
func TestDeviceMappers_ProviderMetadataSetsNeitherRoleNorLocation(t *testing.T) {
	// Arrange.
	e, cleanup := newExporterWithServer(t, jsonHandler(new(int), 200, `{}`))
	defer cleanup()
	seedDeviceRefs(t, e)
	mapper := newCrudMapper(e)
	wantRole, wantLocation := e.Cache.roles["Compute"].ID, e.Cache.locations["DC1"].ID

	// Act.
	single, singleErr := mapper.MapToWritableDeviceRequest(providerRoleDevice("compute-005"))
	bulk, bulkErr := mapper.MapToNautobotDevice(providerRoleDevice("compute-006"))
	patch, patchErr := mapper.MapToPatchRequest(providerRoleDevice("compute-007"), uuid.New())

	// Assert.
	if singleErr != nil || bulkErr != nil || patchErr != nil {
		t.Fatalf("mapper errors = %v, %v, %v", singleErr, bulkErr, patchErr)
	}
	if role, location := transform.RefUUID(single.Role.Id), transform.RefUUID(single.Location.Id); role != wantRole || location != wantLocation {
		t.Errorf("single create: role = %s, location = %s, want the defaults %s and %s", role, location, wantRole, wantLocation)
	}
	if role, location := transform.RefUUID(bulk.Role.Id), transform.RefUUID(bulk.Location.Id); role != wantRole || location != wantLocation {
		t.Errorf("bulk create: role = %s, location = %s, want the defaults %s and %s", role, location, wantRole, wantLocation)
	}
	if patch.Role == nil || patch.Location == nil {
		t.Fatalf("merge patch: role = %v, location = %v, want both set from the defaults", patch.Role, patch.Location)
	}
	if role, location := transform.RefUUID(patch.Role.Id), transform.RefUUID(patch.Location.Id); role != wantRole || location != wantLocation {
		t.Errorf("merge patch: role = %s, location = %s, want the defaults %s and %s", role, location, wantRole, wantLocation)
	}
	if single.CustomFields != nil || bulk.CustomFields != nil || patch.CustomFields != nil {
		t.Errorf("custom fields = %v, %v, %v, want none", single.CustomFields, bulk.CustomFields, patch.CustomFields)
	}
}

// TestCompareDeviceFields_ProviderMetadataIsNotACustomFieldDiff verifies
// that the --merge comparison reports no custom-field diff for provider
// metadata the remote does not carry.
//
// Why it matters: Nautobot never returns a custom-field key it has no
// definition for, so comparing provider metadata as custom fields produced
// a diff and a PATCH on every --merge run and the export never settled.
// Inputs: a device whose provider metadata holds a role, a location and an
// xname, and a remote with no custom fields.
// Outputs: no diff whose field starts with custom_field:.
// Data choice: three keys the old projection treated three different ways.
func TestCompareDeviceFields_ProviderMetadataIsNotACustomFieldDiff(t *testing.T) {
	// Arrange.
	e, cleanup := newExporterWithServer(t, jsonHandler(new(int), 200, `{}`))
	defer cleanup()
	seedDeviceRefs(t, e)
	dev := providerRoleDevice("compute-008")
	dev.ProviderMetadata["csm"] = map[string]any{"xname": "x3000c0s1b0"}
	remote := &nautobotapi.Device{Name: &dev.Name}

	// Act.
	diffs := compareDeviceFields(dev, remote, newCrudMapper(e))

	// Assert.
	for _, d := range diffs {
		if strings.HasPrefix(d.Field, "custom_field:") {
			t.Errorf("unexpected diff %s: %q over %q", d.Field, d.LocalVal, d.RemoteVal)
		}
	}
}

// TestCompareDeviceFields_ComparesListedProviderKeys verifies that the
// --merge comparison sees a listed provider-metadata key the way the write
// sends it: a diff when the remote value differs, none when it agrees.
//
// Why it matters: the comparison decides whether a device is patched at
// all; if it left listed keys out, a device whose only change was a
// provider's value would never be updated, and if it compared unlisted
// ones the export would never settle.
// Inputs: a device whose CSM bucket holds xname and nid, a mapper listing
// csm.xname, and a remote whose xname differs in one case and agrees in
// the other.
// Outputs: one custom-field diff, for xname, in the first case and none in
// the second.
// Data choice: nid is in the same bucket and not listed, so a comparison
// of the whole bucket would add a second diff.
func TestCompareDeviceFields_ComparesListedProviderKeys(t *testing.T) {
	// Arrange.
	e, cleanup := newExporterWithServer(t, jsonHandler(new(int), 200, `{}`))
	defer cleanup()
	seedDeviceRefs(t, e)
	mapper := NewDeviceMapper(e.Cache, &MapperOpts{
		DefaultLocation: "DC1", DefaultStatus: "Active", DefaultRole: "Compute",
		ProviderMetadataFields: []ProviderMetadataField{{Provider: "csm", Key: "xname", CustomField: "xname"}},
	})
	dev := newMappableDevice("compute-009")
	dev.ProviderMetadata = map[string]any{"csm": map[string]any{"xname": "x3000c0s1b0", "nid": 1001}}
	customFieldDiffs := func(remoteXname string) []string {
		remote := &nautobotapi.Device{Name: &dev.Name, CustomFields: nbCF(map[string]any{"xname": remoteXname})}
		var fields []string
		for _, d := range compareDeviceFields(dev, remote, mapper) {
			if strings.HasPrefix(d.Field, "custom_field:") {
				fields = append(fields, d.Field)
			}
		}
		return fields
	}

	// Act.
	drifted := customFieldDiffs("x3000c0s9b0")
	settled := customFieldDiffs("x3000c0s1b0")

	// Assert.
	if len(drifted) != 1 || drifted[0] != "custom_field:xname" {
		t.Errorf("remote xname differs: custom-field diffs = %v, want custom_field:xname alone", drifted)
	}
	if len(settled) != 0 {
		t.Errorf("remote xname agrees: custom-field diffs = %v, want none", settled)
	}
}

// TestMapToWritableRackRequest_ProviderMetadataDoesNotPlaceTheRack verifies
// that a rack kept as a device is placed by the export default when it has
// no placement of its own, whatever provider metadata names.
//
// Why it matters: the rack mapper shares resolveLocation with the device
// mappers, so it took a location from provider metadata the same way.
// Inputs: a rack-category device with no location whose provider metadata
// names DC9, mapped with default location DC1 and then with none.
// Outputs: location DC1 in the first request; an error asking for a
// location in the second; no custom fields.
// Data choice: DC9 is not in the destination, and with no default the
// request can only be built if provider metadata was read.
func TestMapToWritableRackRequest_ProviderMetadataDoesNotPlaceTheRack(t *testing.T) {
	// Arrange.
	e, cleanup := newExporterWithServer(t, jsonHandler(new(int), 200, `{}`))
	defer cleanup()
	seedDeviceRefs(t, e)
	rack := func() *devicetypes.CaniDeviceType {
		return &devicetypes.CaniDeviceType{
			Name: "x3001",
			Type: devicetypes.Rack,
			ObjectMeta: devicetypes.ObjectMeta{
				Status:           "Active",
				ProviderMetadata: map[string]any{"location": "DC9", "nautobot": map[string]any{"location": "DC9"}},
			},
		}
	}

	// Act.
	placed, err := newCrudMapper(e).MapToWritableRackRequest(rack())
	_, unplacedErr := NewDeviceMapper(e.Cache, &MapperOpts{DefaultStatus: "Active"}).MapToWritableRackRequest(rack())

	// Assert.
	if err != nil {
		t.Fatalf("MapToWritableRackRequest() error = %v", err)
	}
	if got, want := transform.RefUUID(placed.Location.Id), e.Cache.locations["DC1"].ID; got != want {
		t.Errorf("location = %s, want the default %s", got, want)
	}
	if placed.CustomFields != nil {
		t.Errorf("custom fields = %v, want none", derefCustomFields(placed.CustomFields))
	}
	if unplacedErr == nil || !strings.Contains(unplacedErr.Error(), "location is required") {
		t.Errorf("no default location: error = %v, want one asking for a location", unplacedErr)
	}
}
