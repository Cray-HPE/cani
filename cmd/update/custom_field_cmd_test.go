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

package update

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// customFieldDevice returns a harness over an inventory holding "cn-01",
// which carries a defined text field, a required date and two keys with no
// definition, and the ID of that device.
func customFieldDevice(t *testing.T) (*cmdtest.Harness, uuid.UUID) {
	t.Helper()
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	deviceID := cmdtest.AddDevice(inventory, rackID, "cn-01", 10)
	cmdtest.DefineCustomField(inventory, "site_code", "text", devicetypes.ContentTypeDevice)
	cmdtest.DefineCustomField(inventory, "spare_slots", "integer", devicetypes.ContentTypeDevice)
	inventory.Metadata.CustomFields = append(inventory.Metadata.CustomFields, devicetypes.CustomFieldDefinition{
		Key: "install_date", Label: "install_date", Type: "date", Required: true,
		ContentTypes: []string{devicetypes.ContentTypeDevice},
	})
	inventory.Devices[deviceID].CustomFields = map[string]any{
		"site_code": "dc1", "install_date": "2026-11-17", "legacy_note": "typed long ago", "owner": "ops",
	}
	return cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory), deviceID
}

// TestUpdateDeviceRejectsCustomFieldAtParseTime verifies that update's
// --custom-field fails before the command runs, with the same errors add
// gives, for both a malformed pair and a flag written without its value.
//
// Why it matters: add and update expose the same flag; a pair the parser let
// through would be reported after the device had been loaded and resolved,
// and "--status" taken as a value would silently lose the flag after it.
// Inputs: an inventory holding "cn-01"; "--custom-field bogus", then
// "--custom-field --status Active".
// Outputs: an error naming --custom-field for each, zero Save calls.
// Data choice: the second argv is the one first reported against add,
// replayed against update to prove the parser guard is not add-specific.
func TestUpdateDeviceRejectsCustomFieldAtParseTime(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"malformed pair", []string{"--custom-field", "bogus"}, `invalid argument "bogus" for --custom-field: expected key=value`},
		{"swallowed flag", []string{"--custom-field", "--status", "Active"}, "flag needs an argument: --custom-field"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			harness, _ := customFieldDevice(t)

			// Act.
			err := harness.Execute(t, append([]string{"update", "device", "cn-01"}, tc.args...)...)

			// Assert.
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want %q", err, tc.want)
			}
			if harness.Store.Saves != 0 {
				t.Errorf("Saves = %d, want 0", harness.Store.Saves)
			}
		})
	}
}

// TestUpdateDeviceSetsAndClearsCustomFields verifies one update can set a
// field, null a defined one and delete a key that has no definition, and
// leaves every other key alone.
//
// Why it matters: this is the only way to change a custom field from the
// command line. A defined field must become an explicit null so the next
// export clears it remotely, and a key with no definition, such as a pair
// carried over from --metadata, must be removable or it stays for good.
// Inputs: `update device cn-01` setting an integer and clearing a defined
// text field and an undefined key.
// Outputs: one Save; the integer stored as an int, the text field present
// as nil, the undefined key gone, the required date and the other undefined
// key untouched.
// Data choice: the device holds two undefined keys so that deleting one is
// distinguishable from dropping every key without a definition.
func TestUpdateDeviceSetsAndClearsCustomFields(t *testing.T) {
	// Arrange.
	harness, deviceID := customFieldDevice(t)

	// Act.
	err := harness.Execute(t, "update", "device", "cn-01", "--custom-field", "spare_slots=2",
		"--clear-custom-field", "site_code", "--clear-custom-field", "legacy_note")

	// Assert.
	if err != nil {
		t.Fatalf("update device: unexpected error: %v", err)
	}
	if harness.Store.Saves != 1 {
		t.Errorf("Saves = %d, want 1", harness.Store.Saves)
	}
	want := map[string]any{"site_code": nil, "spare_slots": 2, "install_date": "2026-11-17", "owner": "ops"}
	if got := harness.Inventory().Devices[deviceID].CustomFields; !reflect.DeepEqual(got, want) {
		t.Errorf("custom fields = %#v\nwant %#v", got, want)
	}
}

// TestUpdateDeviceRefusesCustomFieldEditsAndSavesNothing verifies an edit the
// definitions do not allow fails the whole update, valid flags included.
//
// Why it matters: an update that saved its valid half would leave the
// operator to work out which flags took effect.
// Inputs: a valid --description with, in turn, a key that has no
// definition, a value of the wrong type, a clear of a required field, a
// clear of a key the device neither holds nor has a definition for, and one
// key both set and cleared.
// Outputs: an error holding each wanted fragment, and zero Save calls.
// Data choice: each refusal is paired with a flag that would otherwise
// succeed, so a partial save would be visible as a Save call.
func TestUpdateDeviceRefusesCustomFieldEditsAndSavesNothing(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want []string
	}{
		{"no definition", []string{"--custom-field", "site_cod=dc2"}, []string{"--custom-field", `"site_cod"`, "add metadata custom-field"}},
		{"wrong type", []string{"--custom-field", "spare_slots=4U"}, []string{"--custom-field", `"spare_slots"`, "not an integer"}},
		{"required field", []string{"--clear-custom-field", "install_date"}, []string{"--clear-custom-field", `"install_date"`, "required"}},
		{"unknown key", []string{"--clear-custom-field", "no_such_key"}, []string{"--clear-custom-field", `"no_such_key"`}},
		{"set and cleared", []string{"--custom-field", "site_code=dc2", "--clear-custom-field", "site_code"},
			[]string{"--custom-field", "--clear-custom-field", `"site_code"`}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			harness, _ := customFieldDevice(t)
			argv := append([]string{"update", "device", "cn-01", "--description", "audited"}, tc.args...)

			// Act.
			err := harness.Execute(t, argv...)

			// Assert.
			if err == nil {
				t.Fatal("expected a refusal, got nil")
			}
			for _, want := range tc.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error = %q, want it to contain %q", err, want)
				}
			}
			if harness.Store.Saves != 0 {
				t.Errorf("Saves = %d, want 0", harness.Store.Saves)
			}
		})
	}
}
