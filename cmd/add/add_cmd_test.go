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

// TestAddDeviceQuantityFansOut verifies --qty adds that many distinct devices
// in a single invocation and still persists only once.
//
// Why it matters: bulk add is the common path for filling a rack. Reusing one
// UUID or saving per device would either lose hardware or multiply datastore
// writes.
// Inputs: an inventory with one empty rack, --qty 3 with a sequential naming
// prefix and no explicit position.
// Outputs: a nil error, one Save call, and three uniquely named devices.
// Data choice: omitting --position avoids rack slot contention so the test
// isolates fan-out from placement.
func TestAddDeviceQuantityFansOut(t *testing.T) {
	inventory, _ := cmdtest.InventoryWithRack("rack-01")
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	err := harness.Run(t, map[string]string{
		"rack":   "rack-01",
		"qty":    "3",
		"prefix": "cn-",
		"start":  "1",
	}, cmdtest.DeviceSlug)
	if err != nil {
		t.Fatalf("add device --qty 3: unexpected error: %v", err)
	}

	if harness.Store.Saves != 1 {
		t.Errorf("Saves = %d, want 1 (bulk add should persist once)", harness.Store.Saves)
	}

	names := map[string]bool{}
	ids := map[string]bool{}
	for id, device := range harness.Inventory().Devices {
		names[device.Name] = true
		ids[id.String()] = true
	}
	if len(ids) < 3 {
		t.Errorf("distinct device IDs = %d, want at least 3", len(ids))
	}
	for _, want := range []string{"cn-1", "cn-2", "cn-3"} {
		if !names[want] {
			t.Errorf("device %q missing; got names %v", want, names)
		}
	}
}

// TestAddDeviceDryRunDoesNotPersist verifies --dry-run with a placement
// strategy plans without writing.
//
// Why it matters: dry-run is the safety valve users rely on before a bulk add.
// If it saves, it silently mutates the inventory it claimed only to preview.
// Inputs: an inventory with one empty rack and --rack %{FILL} --dry-run.
// Outputs: a nil error, zero Save calls, and an unchanged device count.
// Data choice: %{FILL} is used because dry-run is only meaningful on the
// strategy path, where cani computes placements rather than taking them.
func TestAddDeviceDryRunDoesNotPersist(t *testing.T) {
	inventory, _ := cmdtest.InventoryWithRack("rack-01")
	before := len(inventory.Devices)
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	err := harness.Run(t, map[string]string{
		"rack":    "%{FILL}",
		"qty":     "2",
		"dry-run": "true",
	}, cmdtest.DeviceSlug)
	if err != nil {
		t.Fatalf("add device --dry-run: unexpected error: %v", err)
	}

	if harness.Store.Saves != 0 {
		t.Errorf("Saves = %d, want 0 (dry-run must not persist)", harness.Store.Saves)
	}
	if got := len(harness.Inventory().Devices); got != before {
		t.Errorf("device count = %d, want %d (unchanged)", got, before)
	}
}

// TestAddDeviceRejectsOccupiedPosition verifies a second device cannot be
// placed into a slot that is already taken.
//
// Why it matters: rack slots are physical; double-booking a U would render the
// rack view and any downstream export wrong.
// Inputs: an inventory whose rack already holds a device at U10, and an add
// targeting that same position and face.
// Outputs: a non-nil error and zero Save calls.
// Data choice: the same face is used deliberately, since front and rear are
// independently addressable and would not collide.
func TestAddDeviceRejectsOccupiedPosition(t *testing.T) {
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	cmdtest.AddDevice(inventory, rackID, "cn-01", 10)
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	err := harness.Run(t, map[string]string{
		"rack":     "rack-01",
		"position": "10",
		"face":     "front",
		"name":     "cn-02",
	}, cmdtest.DeviceSlug)
	if err == nil {
		t.Fatal("expected an error placing a device into an occupied slot, got nil")
	}
	if harness.Store.Saves != 0 {
		t.Errorf("Saves = %d, want 0 (nothing should persist on failure)", harness.Store.Saves)
	}
}

// TestAddRackAddsRackFromSlug verifies "add rack <slug>" creates a rack from
// the embedded library rather than from hand-supplied dimensions.
//
// Why it matters: resolving the slug is what keeps the same hardware identical
// across providers; hand-coding a rack would let two providers disagree.
// Inputs: an empty inventory and a real rack slug with --name.
// Outputs: a nil error, one Save call, and a rack carrying the requested name
// and a non-zero U height taken from the library.
// Data choice: hpe-42u-800mmx1200mm-g2-enterprise-shock-rack is used because it
// declares u_height in the library, so a populated UHeight proves the template
// was applied rather than a bare struct being created.
func TestAddRackAddsRackFromSlug(t *testing.T) {
	harness := cmdtest.New(t, NewCommand(), newRackAddCommand(), nil)

	err := harness.Run(t, map[string]string{"name": "rack-99"}, "hpe-42u-800mmx1200mm-g2-enterprise-shock-rack")
	if err != nil {
		t.Fatalf("add rack: unexpected error: %v", err)
	}

	if harness.Store.Saves != 1 {
		t.Errorf("Saves = %d, want 1", harness.Store.Saves)
	}

	var found bool
	for _, rack := range harness.Inventory().Racks {
		if rack.Name != "rack-99" {
			continue
		}
		found = true
		if rack.UHeight != 42 {
			t.Errorf("UHeight = %d, want 42 (should come from the device-type library)", rack.UHeight)
		}
	}
	if !found {
		t.Error("rack \"rack-99\" not found in inventory")
	}
}

// TestAddIsProviderAgnostic verifies add works with no provider registered.
//
// Why it matters: this is the executable form of the claim that add/remove/
// update/show behave identically for every provider. A provider import creeping
// into cmd/add would fail here, at the point it is introduced.
// Inputs: the registry as it stands during this package's tests, plus a normal
// device add.
// Outputs: an empty provider registry and a device created from the library.
// Data choice: pairing the registry assertion with a real add proves the verb
// genuinely runs without providers rather than merely importing none.
func TestAddIsProviderAgnostic(t *testing.T) {
	cmdtest.RequireNoProviders(t)

	inventory, _ := cmdtest.InventoryWithRack("rack-01")
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	err := harness.Run(t, map[string]string{
		"rack": "rack-01",
		"name": "cn-01",
	}, cmdtest.DeviceSlug)
	if err != nil {
		t.Fatalf("add device without providers: %v", err)
	}
	if len(harness.Inventory().Devices) == 0 {
		t.Error("no device was added")
	}
}
