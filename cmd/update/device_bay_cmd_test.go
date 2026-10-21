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
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
)

// TestUpdateDeviceInstallsAndMovesBays verifies --parent into a chassis picks
// a bay, --bay alone moves within the chassis, and --parent to a rack clears
// the bay.
//
// Why it matters: this is how scripts place existing switches into a chassis
// and later rearrange them; a stale bay after leaving the chassis would make
// export install the device somewhere it no longer is. Inputs: a rack-mounted
// device moved into a two-bay chassis, then to Bay 2, then back to the rack.
// Outputs: Bay 1, Bay 2, then no bay with the rack as parent.
func TestUpdateDeviceInstallsAndMovesBays(t *testing.T) {
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	chassisID := cmdtest.AddChassisWithBays(inventory, rackID, "chassis-15", 15, "Bay 1", "Bay 2")
	deviceID := cmdtest.AddDevice(inventory, rackID, "sw-a", 20)

	steps := []struct {
		flags      map[string][]string
		wantParent string
		wantBay    string
	}{
		{map[string][]string{"parent": {"chassis-15"}}, "chassis", "Bay 1"},
		{map[string][]string{"bay": {"Bay 2"}}, "chassis", "Bay 2"},
		{map[string][]string{"parent": {"rack-01"}}, "rack", ""},
	}
	for i, step := range steps {
		harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)
		if err := harness.Run(t, step.flags, "sw-a"); err != nil {
			t.Fatalf("step %d: update: %v", i, err)
		}
		inventory = harness.Inventory()
		device := inventory.Devices[deviceID]
		wantParent := map[string]any{"chassis": chassisID, "rack": rackID}[step.wantParent]
		if device.Parent != wantParent || device.ParentBay != step.wantBay {
			t.Errorf("step %d: parent %s bay %q, want %s %q", i, device.Parent, device.ParentBay, step.wantParent, step.wantBay)
		}
	}
}

// TestUpdateDeviceBayErrors verifies --bay without a bay-capable parent and an
// occupied bay are rejected without saving.
//
// Why it matters: a rejected move must leave the inventory as it was. Inputs:
// --bay on a rack-mounted device, and --bay naming a sibling's bay. Outputs:
// an error and zero saves for each.
func TestUpdateDeviceBayErrors(t *testing.T) {
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	chassisID := cmdtest.AddChassisWithBays(inventory, rackID, "chassis-15", 15, "Bay 1", "Bay 2")
	cmdtest.AddDevice(inventory, rackID, "racked", 20)
	for _, name := range []string{"sw-a", "sw-b"} {
		id := cmdtest.AddChildDevice(inventory, chassisID, name)
		inventory.Devices[id].ParentBay = map[string]string{"sw-a": "Bay 1", "sw-b": "Bay 2"}[name]
	}

	cases := []struct{ device, bay, want string }{
		{"racked", "Bay 1", "--bay requires the parent"},
		{"sw-a", "Bay 2", `occupied by "sw-b"`},
	}
	for _, tc := range cases {
		harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)
		err := harness.Run(t, map[string][]string{"bay": {tc.bay}}, tc.device)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s --bay %s: err = %v, want containing %q", tc.device, tc.bay, err, tc.want)
		}
		if harness.Store.Saves != 0 {
			t.Errorf("%s: Saves = %d, want 0", tc.device, harness.Store.Saves)
		}
	}
}
