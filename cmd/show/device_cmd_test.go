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

package show

import (
	"testing"

	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
)

// TestShowDeviceIsReadOnly verifies that "show device" loads the inventory and
// never writes it back.
//
// Why it matters: show is the only CRUD verb with no mutation semantics. A
// stray Save here would rewrite the datastore on a read, which is how derived
// reverse indices leak onto disk.
// Inputs: an inventory with one rack and one device, rendered as JSON.
// Outputs: a nil error, at least one Load call, and zero Save calls.
// Data choice: the json format avoids the table and tree renderers so the
// assertion isolates datastore behavior from presentation.
func TestShowDeviceIsReadOnly(t *testing.T) {
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	cmdtest.AddDevice(inventory, rackID, "cn-01", 10)
	harness := cmdtest.New(t, NewCommand(), newDeviceShowCommand(), inventory)

	if err := harness.Run(t, map[string][]string{"format": {"json"}}); err != nil {
		t.Fatalf("show device: unexpected error: %v", err)
	}

	if harness.Store.Loads == 0 {
		t.Error("Loads = 0, want at least 1")
	}
	if harness.Store.Saves != 0 {
		t.Errorf("Saves = %d, want 0 (show must never write)", harness.Store.Saves)
	}
}

// TestShowDeviceRejectsUnknownName verifies that requesting a single device
// that does not exist is an error rather than empty output.
//
// Why it matters: silently printing nothing for a mistyped name hides the
// mistake; users need to know the lookup failed.
// Inputs: an inventory holding one device, and a name that does not match it.
// Outputs: a non-nil error and zero Save calls.
// Data choice: the single-device path is separate from the list path, so it
// needs its own coverage.
func TestShowDeviceRejectsUnknownName(t *testing.T) {
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	cmdtest.AddDevice(inventory, rackID, "cn-01", 10)
	harness := cmdtest.New(t, NewCommand(), newDeviceShowCommand(), inventory)

	err := harness.Run(t, map[string][]string{"format": {"json"}}, "does-not-exist")
	if err == nil {
		t.Fatal("expected an error for an unknown device, got nil")
	}
	if harness.Store.Saves != 0 {
		t.Errorf("Saves = %d, want 0 (show must never write)", harness.Store.Saves)
	}
}

// TestShowDeviceRendersEveryFormatReadOnly verifies each output format renders
// without error and without writing.
//
// Why it matters: the table and tree renderers walk the inventory and are the
// most likely place for presentation code to mutate the model it is drawing.
// Inputs: an inventory with a rack and a device, rendered as table, tree and
// json in turn.
// Outputs: a nil error and zero Save calls for every format.
// Data choice: all three formats are exercised because they take independent
// code paths, and only json avoids the visual package entirely.
func TestShowDeviceRendersEveryFormatReadOnly(t *testing.T) {
	for _, format := range []string{"table", "tree", "json"} {
		t.Run(format, func(t *testing.T) {
			inventory, rackID := cmdtest.InventoryWithRack("rack-01")
			cmdtest.AddDevice(inventory, rackID, "cn-01", 10)
			harness := cmdtest.New(t, NewCommand(), newDeviceShowCommand(), inventory)

			if err := harness.Run(t, map[string][]string{"format": {format}}); err != nil {
				t.Fatalf("show device --format %s: unexpected error: %v", format, err)
			}
			if harness.Store.Saves != 0 {
				t.Errorf("Saves = %d, want 0 (show must never write)", harness.Store.Saves)
			}
			if got := len(harness.Inventory().Devices); got != 1 {
				t.Errorf("device count = %d, want 1 (rendering must not mutate)", got)
			}
		})
	}
}

// TestShowIsProviderAgnostic verifies show works with no provider registered.
//
// Why it matters: show resolves its format list partly from providers, so it is
// the verb most at risk of requiring one. It must still render the base formats
// with an empty registry.
// Inputs: the registry as it stands during this package's tests, plus a listing.
// Outputs: an empty provider registry and a successful render.
// Data choice: pairing the registry assertion with a real render proves the verb
// genuinely runs without providers rather than merely importing none.
func TestShowIsProviderAgnostic(t *testing.T) {
	cmdtest.RequireNoProviders(t)

	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	cmdtest.AddDevice(inventory, rackID, "cn-01", 10)
	harness := cmdtest.New(t, NewCommand(), newDeviceShowCommand(), inventory)

	if err := harness.Run(t, map[string][]string{"format": {"json"}}); err != nil {
		t.Fatalf("show device without providers: %v", err)
	}
	if harness.Store.Saves != 0 {
		t.Errorf("Saves = %d, want 0", harness.Store.Saves)
	}
}
