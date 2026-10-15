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

// TestAddDeviceRejectsMetadataMissingItsValue verifies the command line a
// reviewer actually typed — a --metadata with no value in front of --status —
// fails at parse time and writes nothing.
//
// Why it matters: the parser used to hand the next token to --metadata
// whatever it looked like, so "--status" became a metadata value; the old code
// then dropped every pair silently, never saw --status, and saved a device
// with the default status and no role. The user only found out after an
// export.
// Inputs: the reported argv against an inventory with one rack.
// Outputs: the parser's missing-argument error naming --metadata, zero Save
// calls and no device.
// Data choice: flag order and tokens match the report, including the valid
// role pair that used to be lost along with the bad one.
func TestAddDeviceRejectsMetadataMissingItsValue(t *testing.T) {
	// Arrange.
	inventory, _ := cmdtest.InventoryWithRack("rack-01")
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	// Act.
	err := harness.Execute(t, "add", "device", cmdtest.DeviceSlug,
		"--rack", "rack-01", "--position", "10",
		"--metadata", "role=SpineSwitch", "--metadata", "--status", "Active")

	// Assert.
	if err == nil {
		t.Fatal("expected a parse error, got nil")
	}
	if !strings.Contains(err.Error(), "flag needs an argument: --metadata") {
		t.Errorf("error = %q, want the missing-argument error for --metadata", err)
	}
	if harness.Store.Saves != 0 {
		t.Errorf("Saves = %d, want 0", harness.Store.Saves)
	}
	if n := len(harness.Inventory().Devices); n != 0 {
		t.Errorf("devices = %d, want 0", n)
	}
}

// TestAddDeviceRejectsMalformedMetadata verifies a --metadata value without a
// key=value shape fails at parse time, naming the flag and the value.
//
// Why it matters: this is the case the parser's flag-token guard cannot see;
// the key=value flag type itself has to refuse it before the command runs.
// Inputs: "--metadata bogus" on an otherwise valid add device.
// Outputs: an error naming the value and --metadata, and zero Save calls.
// Data choice: "bogus" has no "=" at all, the simplest malformed pair.
func TestAddDeviceRejectsMalformedMetadata(t *testing.T) {
	// Arrange.
	inventory, _ := cmdtest.InventoryWithRack("rack-01")
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	// Act.
	err := harness.Execute(t, "add", "device", cmdtest.DeviceSlug,
		"--rack", "rack-01", "--position", "10", "--metadata", "bogus")

	// Assert.
	if err == nil {
		t.Fatal("expected a parse error, got nil")
	}
	if !strings.Contains(err.Error(), `invalid argument "bogus" for --metadata: expected key=value`) {
		t.Errorf("error = %q, want the value and flag named", err)
	}
	if harness.Store.Saves != 0 {
		t.Errorf("Saves = %d, want 0", harness.Store.Saves)
	}
}

// TestAddDeviceAppliesMetadataAndStatus verifies the corrected form of the
// same command parses --metadata and --status and persists both.
//
// Why it matters: this is the regression guard for the report's intent, and
// it proves the stricter flag does not reject well-formed input.
// Inputs: the reviewer's argv without the stray --metadata.
// Outputs: one Save, exactly one device with status Active, and the role pair
// available to providers. No provider is registered here; the dispatch of the
// pairs to MetadataApplier is covered by the cmdtest capability tests.
// Data choice: "Active" is a built-in status so no status definition is
// needed in the inventory.
func TestAddDeviceAppliesMetadataAndStatus(t *testing.T) {
	// Arrange.
	inventory, _ := cmdtest.InventoryWithRack("rack-01")
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	// Act.
	err := harness.Execute(t, "add", "device", cmdtest.DeviceSlug,
		"--rack", "rack-01", "--position", "10", "--name", "cn-01",
		"--metadata", "role=SpineSwitch", "--status", "Active")

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
	}
	if got := collectProviderMetadata(harness.Target); got["role"] != "SpineSwitch" {
		t.Errorf("metadata = %v, want role=SpineSwitch", got)
	}
}
