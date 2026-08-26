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
	"testing"

	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
)

// TestUpdateDeviceRenamesDevice verifies that "update device <name> --name"
// renames the resolved device, leaves its placement untouched, and persists
// exactly once.
//
// Why it matters: update must mutate only the fields the user named. Silently
// clearing placement or parentage would corrupt the rack view for every
// provider, since all of them share this code path.
// Inputs: an inventory with one rack and a device "cn-01" at U10, and a
// --name flag renaming it.
// Outputs: a nil error, one Save call, the new name, and an unchanged rack
// parent and position.
// Data choice: asserting on the untouched fields alongside the changed one is
// what makes this a regression test rather than a smoke test.
func TestUpdateDeviceRenamesDevice(t *testing.T) {
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	deviceID := cmdtest.AddDevice(inventory, rackID, "cn-01", 10)
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	if err := harness.Run(t, map[string]string{"name": "cn-99"}, "cn-01"); err != nil {
		t.Fatalf("update device: unexpected error: %v", err)
	}

	if harness.Store.Saves != 1 {
		t.Errorf("Saves = %d, want 1", harness.Store.Saves)
	}

	device := harness.Inventory().Devices[deviceID]
	if device == nil {
		t.Fatalf("device %s missing after update", deviceID)
	}
	if device.Name != "cn-99" {
		t.Errorf("Name = %q, want %q", device.Name, "cn-99")
	}
	if device.Parent != rackID {
		t.Errorf("Parent = %s, want %s (should be preserved)", device.Parent, rackID)
	}
	if device.RackPosition != 10 {
		t.Errorf("RackPosition = %d, want 10 (should be preserved)", device.RackPosition)
	}
}

// TestUpdateDeviceRejectsUnknownName verifies an unresolvable device name
// fails without persisting.
//
// Why it matters: update loads, mutates, then saves; a failed resolution must
// abort before the save so a no-op does not rewrite the datastore.
// Inputs: an inventory holding one device, and a name that does not match it.
// Outputs: a non-nil error and zero Save calls.
// Data choice: mirrors the remove-command failure case so both destructive
// verbs are held to the same standard.
func TestUpdateDeviceRejectsUnknownName(t *testing.T) {
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	cmdtest.AddDevice(inventory, rackID, "cn-01", 10)
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	err := harness.Run(t, map[string]string{"name": "cn-99"}, "does-not-exist")
	if err == nil {
		t.Fatal("expected an error for an unknown device, got nil")
	}
	if harness.Store.Saves != 0 {
		t.Errorf("Saves = %d, want 0 (nothing should persist on failure)", harness.Store.Saves)
	}
}
