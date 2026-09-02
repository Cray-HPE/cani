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
