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
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// TestAddDevicePlacesDeviceInRack verifies that "add device <slug>" resolves
// the rack by name, places the device at the requested position and face, and
// persists exactly once.
//
// Why it matters: this is the reference case for provider-agnostic CRUD. The
// command runs with no provider registered, proving add operates purely on the
// portable model through the datastore interface.
// Inputs: an inventory with one empty 42U rack, and --rack/--position/--face
// flags with a real device slug from the embedded library.
// Outputs: a nil error, one Save call, and a device carrying the rack as its
// parent FK at the requested position.
// Data choice: a named rack (not a UUID) exercises the resolver, and U10 is
// clear of the rack's boundaries so placement cannot fail for position reasons.
func TestAddDevicePlacesDeviceInRack(t *testing.T) {
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	err := harness.Run(t, map[string][]string{
		"rack":     {"rack-01"},
		"position": {"10"},
		"face":     {"front"},
		"name":     {"cn-01"},
	}, cmdtest.DeviceSlug)
	if err != nil {
		t.Fatalf("add device: unexpected error: %v", err)
	}

	if harness.Store.Saves != 1 {
		t.Errorf("Saves = %d, want 1", harness.Store.Saves)
	}

	devices := harness.Inventory().Devices
	if len(devices) == 0 {
		t.Fatal("no devices were added to the inventory")
	}

	var found bool
	for _, device := range devices {
		if device.Name != "cn-01" {
			continue
		}
		found = true
		if device.Parent != rackID {
			t.Errorf("Parent = %s, want %s", device.Parent, rackID)
		}
		if device.RackPosition != 10 {
			t.Errorf("RackPosition = %d, want 10", device.RackPosition)
		}
	}
	if !found {
		t.Errorf("device %q not found among %d devices", "cn-01", len(devices))
	}
}

// TestAddDeviceRejectsUnknownSlug verifies an unresolvable slug fails before
// anything is written.
//
// Why it matters: the slug is the contract with the embedded device-type
// library; a typo must not produce a half-populated device or a silent no-op.
// Inputs: an inventory with one rack and a slug that is not in the library.
// Outputs: a non-nil error and zero Save calls.
// Data choice: a clearly fictitious slug avoids colliding with any real part
// number as the library grows.
func TestAddDeviceRejectsUnknownSlug(t *testing.T) {
	inventory, _ := cmdtest.InventoryWithRack("rack-01")
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	err := harness.Run(t, map[string][]string{"rack": {"rack-01"}}, "not-a-real-slug")
	if err == nil {
		t.Fatal("expected an error for an unknown slug, got nil")
	}
	if harness.Store.Saves != 0 {
		t.Errorf("Saves = %d, want 0 (nothing should persist on failure)", harness.Store.Saves)
	}
}

// TestAddDeviceRejectsCustomFieldMissingItsValue verifies a key=value flag
// with no value in front of another flag fails at parse time and writes
// nothing.
//
// Why it matters: the parser used to hand the next token to a key=value flag
// whatever it looked like, so "--status" became a value; the pairs were then
// dropped silently, --status was never seen, and a device was saved with the
// default status. The user only found out after an export.
// Inputs: a valid pair, then --custom-field directly followed by --status,
// against an inventory with one rack.
// Outputs: the parser's missing-argument error naming --custom-field, zero
// Save calls and no device.
// Data choice: the valid pair comes first, because it used to be lost along
// with the bad one.
func TestAddDeviceRejectsCustomFieldMissingItsValue(t *testing.T) {
	// Arrange.
	inventory, _ := cmdtest.InventoryWithRack("rack-01")
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	// Act.
	err := harness.Execute(t, "add", "device", cmdtest.DeviceSlug,
		"--rack", "rack-01", "--position", "10",
		"--custom-field", "site_code=dc1", "--custom-field", "--status", "Active")

	// Assert.
	if err == nil {
		t.Fatal("expected a parse error, got nil")
	}
	if !strings.Contains(err.Error(), "flag needs an argument: --custom-field") {
		t.Errorf("error = %q, want the missing-argument error for --custom-field", err)
	}
	if harness.Store.Saves != 0 {
		t.Errorf("Saves = %d, want 0", harness.Store.Saves)
	}
	if n := len(harness.Inventory().Devices); n != 0 {
		t.Errorf("devices = %d, want 0", n)
	}
}

// TestAddDeviceRejectsMalformedCustomField verifies a --custom-field value
// without a key=value shape fails at parse time, naming the flag and the
// value.
//
// Why it matters: this is the case the parser's flag-token guard cannot see;
// the key=value flag type itself has to refuse it before the command runs.
// Inputs: "--custom-field bogus" on an otherwise valid add device.
// Outputs: an error naming the value and --custom-field, and zero Save calls.
// Data choice: "bogus" has no "=" at all, the simplest malformed pair.
func TestAddDeviceRejectsMalformedCustomField(t *testing.T) {
	// Arrange.
	inventory, _ := cmdtest.InventoryWithRack("rack-01")
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	// Act.
	err := harness.Execute(t, "add", "device", cmdtest.DeviceSlug,
		"--rack", "rack-01", "--position", "10", "--custom-field", "bogus")

	// Assert.
	if err == nil {
		t.Fatal("expected a parse error, got nil")
	}
	if !strings.Contains(err.Error(), `invalid argument "bogus" for --custom-field: expected key=value`) {
		t.Errorf("error = %q, want the value and flag named", err)
	}
	if harness.Store.Saves != 0 {
		t.Errorf("Saves = %d, want 0", harness.Store.Saves)
	}
}

// TestAddDeviceAppliesCustomFieldAndStatus verifies the well-formed form of
// the same command stores the custom field and the status that follows it.
//
// Why it matters: it proves the stricter flag does not reject well-formed
// input, and that a flag after --custom-field is still seen.
// Inputs: one --custom-field pair for a field defined for devices, then
// --status.
// Outputs: one Save, exactly one device with status Active and the field.
// Data choice: "Active" is a built-in status so no status definition is
// needed in the inventory; the value holds an "=" to prove the pair is split
// at the first one only.
func TestAddDeviceAppliesCustomFieldAndStatus(t *testing.T) {
	// Arrange.
	inventory, _ := cmdtest.InventoryWithRack("rack-01")
	cmdtest.DefineCustomField(inventory, "site_code", "text", devicetypes.ContentTypeDevice)
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	// Act.
	err := harness.Execute(t, "add", "device", cmdtest.DeviceSlug,
		"--rack", "rack-01", "--position", "10", "--name", "cn-01",
		"--custom-field", "site_code=dc1=east", "--status", "Active")

	// Assert.
	if err != nil {
		t.Fatalf("add device: unexpected error: %v", err)
	}
	if harness.Store.Saves != 1 {
		t.Errorf("Saves = %d, want 1", harness.Store.Saves)
	}
	devices := harness.Inventory().Devices
	if len(devices) != 1 {
		t.Fatalf("devices = %d, want 1", len(devices))
	}
	for _, device := range devices {
		if device.Status != "Active" {
			t.Errorf("Status = %q, want %q", device.Status, "Active")
		}
		if got := device.CustomFields["site_code"]; got != "dc1=east" {
			t.Errorf("custom field site_code = %v, want dc1=east", got)
		}
	}
}

// TestAddDeviceRejectsExtraPositionalArgs verifies a second positional token
// is an error rather than silently ignored.
//
// Why it matters: addDevice reads args[0] only, so a status or name typed
// without its flag used to vanish; the device was added as if it were absent.
// Inputs: a valid slug followed by a stray "Active" token, with a valid rack.
// Outputs: an error quoting the stray token, zero Save calls.
// Data choice: "Active" is what the reported command left behind once the
// parser had consumed "--status" as a metadata value.
func TestAddDeviceRejectsExtraPositionalArgs(t *testing.T) {
	// Arrange.
	inventory, _ := cmdtest.InventoryWithRack("rack-01")
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	// Act.
	err := harness.Execute(t, "add", "device", cmdtest.DeviceSlug, "Active",
		"--rack", "rack-01", "--position", "10")

	// Assert.
	if err == nil {
		t.Fatal("expected an error for a stray positional argument, got nil")
	}
	if !strings.Contains(err.Error(), `unexpected argument(s) ["Active"]; add device takes a single slug or part number`) {
		t.Errorf("error = %q, want the stray token quoted", err)
	}
	if harness.Store.Saves != 0 {
		t.Errorf("Saves = %d, want 0", harness.Store.Saves)
	}
}
