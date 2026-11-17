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

package add

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// addNounContentTypes names, for each addNounCases entry that stores custom
// fields, the content type its items are checked against.
var addNounContentTypes = map[string]string{
	"device":           devicetypes.ContentTypeDevice,
	"rack":             devicetypes.ContentTypeRack,
	"module":           devicetypes.ContentTypeModule,
	"location":         devicetypes.ContentTypeLocation,
	"cable":            devicetypes.ContentTypeCable,
	"vlan":             devicetypes.ContentTypeVLAN,
	"prefix":           devicetypes.ContentTypePrefix,
	"ip":               devicetypes.ContentTypeIPAddress,
	"vrf":              devicetypes.ContentTypeVRF,
	"bare device slug": devicetypes.ContentTypeDevice,
	"bare rack slug":   devicetypes.ContentTypeRack,
	"bare module slug": devicetypes.ContentTypeModule,
	"bare cable slug":  devicetypes.ContentTypeCable,
}

// TestAddChecksCustomFieldAgainstTheNounsOwnKind verifies each noun accepts a
// custom field defined for the kind of item it creates and for no other.
//
// Why it matters: a definition lists the kinds of object it applies to, and
// Nautobot drops a value sent for any other kind; a noun that checked against
// the wrong kind would refuse a valid field or store one that never arrives.
// Inputs: every noun that stores custom fields, run with --custom-field
// against an inventory that defines the field for that noun's kind only,
// then for interfaces only.
// Outputs: the field stored in the first run, and in the second an error
// naming the flag and no save.
// Data choice: one content type per run, so a noun using another noun's
// kind fails the first run; interfaces are the kind no listed noun creates.
func TestAddChecksCustomFieldAgainstTheNounsOwnKind(t *testing.T) {
	flag := []string{"--custom-field", inheritedFlagField + "=checked"}
	for _, tc := range addNounCases(t) {
		contentType, stores := addNounContentTypes[tc.name]
		if !stores {
			continue
		}
		// Arrange.
		argv := append(append([]string{"add"}, tc.argv...), flag...)
		baseline, err := runAddNounDefining(t, tc.argv, contentType)
		if err != nil {
			t.Errorf("add %s: the flagless run failed, so the case proves nothing: %v", tc.name, err)
			continue
		}
		// Act. A harness installs its store when it is made, so each is run
		// before the next is made.
		own := addNounHarness(t, contentType)
		ownErr := own.Execute(t, argv...)
		other := addNounHarness(t, devicetypes.ContentTypeInterface)
		otherErr := other.Execute(t, argv...)

		// Assert.
		if ownErr != nil || cmdtest.Fingerprint(own.Inventory()) == baseline {
			t.Errorf("add %s: a field defined for %s was not stored (error: %v)", tc.name, contentType, ownErr)
		}
		if otherErr == nil || !strings.Contains(otherErr.Error(), "--custom-field") {
			t.Errorf("add %s: a field defined for interfaces only: error = %v, want a refusal naming the flag", tc.name, otherErr)
		}
		if other.Store.Saves != 0 {
			t.Errorf("add %s: a refused field was followed by %d save(s)", tc.name, other.Store.Saves)
		}
	}
}

// TestAddDeviceStampsEachDeviceWithItsOwnTypedValues verifies one command
// line gives every device it adds the custom fields in the type their
// definitions declare, and that no two devices share a value.
//
// Why it matters: --qty builds each device from the same parsed flags; a
// number stored as text is rejected on export, and a shared list would let an
// edit to one device change the others.
// Inputs: `add device --qty 2` with an integer and a multi-select field.
// Outputs: two devices, each holding an int and its own list of two choices.
// Data choice: the selection holds a comma, which only a flag type that
// splits on the first "=" alone delivers whole.
func TestAddDeviceStampsEachDeviceWithItsOwnTypedValues(t *testing.T) {
	// Arrange.
	inventory, _ := cmdtest.InventoryWithRack("rack-01")
	cmdtest.DefineCustomField(inventory, "spare_slots", "integer", devicetypes.ContentTypeDevice)
	inventory.Metadata.CustomFields = append(inventory.Metadata.CustomFields, devicetypes.CustomFieldDefinition{
		Key: "fabrics", Label: "fabrics", Type: "multi-select", Choices: []string{"hsn", "mgmt"},
		ContentTypes: []string{devicetypes.ContentTypeDevice},
	})
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	// Act.
	err := harness.Execute(t, "add", "device", cmdtest.DeviceSlug, "--qty", "2", "--prefix", "cn",
		"--custom-field", "spare_slots=4", "--custom-field", "fabrics=hsn,mgmt")

	// Assert.
	if err != nil {
		t.Fatalf("add device: unexpected error: %v", err)
	}
	var lists [][]string
	for _, device := range harness.Inventory().Devices {
		want := map[string]any{"spare_slots": 4, "fabrics": []string{"hsn", "mgmt"}}
		if !reflect.DeepEqual(device.CustomFields, want) {
			t.Errorf("%s custom fields = %#v\nwant %#v", device.Name, device.CustomFields, want)
			continue
		}
		lists = append(lists, device.CustomFields["fabrics"].([]string))
	}
	if len(lists) != 2 {
		t.Fatalf("devices holding the fields = %d, want 2", len(lists))
	}
	if &lists[0][0] == &lists[1][0] {
		t.Error("the two devices share one list of choices")
	}
}

// TestAddRefusesCustomFieldWithNoDefinition verifies a key the inventory does
// not define is refused before anything is added, with an error that says
// how to define it.
//
// Why it matters: an undefined key used to be stored and was then dropped by
// Nautobot, which answers 2xx for it; the operator has to learn at the
// command line, not after an export.
// Inputs: `add device` with --custom-field for a key one letter short of a
// defined one.
// Outputs: an error naming the flag, the key and the command that defines a
// field; zero Save calls and no device.
// Data choice: a near-miss key is the mistake a definition check exists for.
func TestAddRefusesCustomFieldWithNoDefinition(t *testing.T) {
	// Arrange.
	inventory, _ := cmdtest.InventoryWithRack("rack-01")
	cmdtest.DefineCustomField(inventory, "site_code", "text", devicetypes.ContentTypeDevice)
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	// Act.
	err := harness.Execute(t, "add", "device", cmdtest.DeviceSlug,
		"--rack", "rack-01", "--position", "10", "--custom-field", "site_cod=dc1")

	// Assert.
	if err == nil {
		t.Fatal("expected a refusal, got nil")
	}
	for _, want := range []string{"--custom-field", `"site_cod"`, "add metadata custom-field"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want it to contain %q", err, want)
		}
	}
	if harness.Store.Saves != 0 || len(harness.Inventory().Devices) != 0 {
		t.Errorf("Saves = %d, devices = %d, want neither", harness.Store.Saves, len(harness.Inventory().Devices))
	}
}
