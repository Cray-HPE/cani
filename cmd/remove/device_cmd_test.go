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

package remove

import (
	"testing"

	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
)

// TestRemoveDeviceDeletesDeviceByName verifies that "remove device <name>"
// resolves the device by name, deletes it, and persists exactly once.
//
// Why it matters: remove is the destructive half of provider-agnostic CRUD. It
// must operate on the portable model alone, with no provider registered.
// Inputs: an inventory holding one rack and one device named "cn-01".
// Outputs: a nil error, one Save call, and an inventory with no devices.
// Data choice: addressing the device by name rather than UUID exercises the
// resolver, which is the path users actually take.
func TestRemoveDeviceDeletesDeviceByName(t *testing.T) {
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	cmdtest.AddDevice(inventory, rackID, "cn-01", 10)
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	if err := harness.Run(t, nil, "cn-01"); err != nil {
		t.Fatalf("remove device: unexpected error: %v", err)
	}

	if harness.Store.Saves != 1 {
		t.Errorf("Saves = %d, want 1", harness.Store.Saves)
	}
	if got := len(harness.Inventory().Devices); got != 0 {
		t.Errorf("device count = %d, want 0", got)
	}
}

// TestRemoveDeviceRejectsUnknownName verifies an unresolvable device name
// fails without persisting.
//
// Why it matters: a typo must not silently succeed or write a mutated
// inventory back to the datastore.
// Inputs: an inventory holding one device, and a name that does not match it.
// Outputs: a non-nil error and zero Save calls.
// Data choice: a name close to nothing in the inventory isolates resolution
// failure from any cascade behavior.
func TestRemoveDeviceRejectsUnknownName(t *testing.T) {
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	cmdtest.AddDevice(inventory, rackID, "cn-01", 10)
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	err := harness.Run(t, nil, "does-not-exist")
	if err == nil {
		t.Fatal("expected an error for an unknown device, got nil")
	}
	if harness.Store.Saves != 0 {
		t.Errorf("Saves = %d, want 0 (nothing should persist on failure)", harness.Store.Saves)
	}
}
