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
	"github.com/google/uuid"
)

// childrenOf returns the devices whose parent is parentID, keyed by name.
func childrenOf(inv *devicetypes.Inventory, parentID uuid.UUID) map[string]*devicetypes.CaniDeviceType {
	out := make(map[string]*devicetypes.CaniDeviceType)
	for _, d := range inv.Devices {
		if d.Parent == parentID {
			out[d.Name] = d
		}
	}
	return out
}

// TestAddDeviceInstallsIntoBay verifies --parent accepts a chassis name and
// installs the child in the first free bay, or the bay named by --bay.
//
// Why it matters: half-width switches and blades live in a parent device's
// bays; the bay is what Nautobot export installs them into. Inputs: a 1U
// chassis with two bays; one add without --bay, one with --bay "bay 2" (case
// differs). Outputs: Bay 1 then Bay 2, parent = chassis, no own rack position.
func TestAddDeviceInstallsIntoBay(t *testing.T) {
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	chassisID := cmdtest.AddChassisWithBays(inventory, rackID, "chassis-15", 15, "Bay 1", "Bay 2")

	for _, tc := range []struct{ name, bay string }{{"sw-a", ""}, {"sw-b", "bay 2"}} {
		harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)
		flags := map[string][]string{"parent": {"chassis-15"}, "name": {tc.name}}
		if tc.bay != "" {
			flags["bay"] = []string{tc.bay}
		}
		if err := harness.Run(t, flags, cmdtest.DeviceSlug); err != nil {
			t.Fatalf("add %s: %v", tc.name, err)
		}
		inventory = harness.Inventory()
	}

	kids := childrenOf(inventory, chassisID)
	if kids["sw-a"] == nil || kids["sw-a"].ParentBay != "Bay 1" || kids["sw-b"] == nil || kids["sw-b"].ParentBay != "Bay 2" {
		t.Fatalf("children = %+v, want sw-a in Bay 1 and sw-b in Bay 2", kids)
	}
	if kids["sw-a"].RackPosition != 0 || kids["sw-a"].Face != "" {
		t.Errorf("child kept its own placement: U%d %q", kids["sw-a"].RackPosition, kids["sw-a"].Face)
	}
}

// TestAddDeviceQtyFillsBays verifies a multi-quantity add assigns each new
// child a distinct bay.
//
// Why it matters: all devices in one add are built before saving, so the bay
// check must see earlier devices from the same command. Inputs: --qty 2 into
// an empty two-bay chassis. Outputs: one child in each bay.
func TestAddDeviceQtyFillsBays(t *testing.T) {
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	chassisID := cmdtest.AddChassisWithBays(inventory, rackID, "chassis-15", 15, "Bay 1", "Bay 2")
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	err := harness.Run(t, map[string][]string{"parent": {"chassis-15"}, "qty": {"2"}, "prefix": {"sw-"}}, cmdtest.DeviceSlug)
	if err != nil {
		t.Fatalf("add --qty 2: %v", err)
	}
	bays := map[string]bool{}
	for _, child := range childrenOf(harness.Inventory(), chassisID) {
		bays[child.ParentBay] = true
	}
	if !bays["Bay 1"] || !bays["Bay 2"] {
		t.Errorf("bays used = %v, want Bay 1 and Bay 2", bays)
	}
}

// TestAddDeviceBayErrors verifies bay mistakes fail before anything is saved.
//
// Why it matters: a bad bay caught at add time is far easier to fix than a
// Nautobot rejection during export. Inputs: --bay with --rack, an unknown bay,
// and a third child for a full two-bay chassis. Outputs: an error naming the
// problem and zero saves for each.
func TestAddDeviceBayErrors(t *testing.T) {
	cases := []struct {
		name  string
		flags map[string][]string
		want  string
	}{
		{"bay with rack", map[string][]string{"rack": {"rack-01"}, "position": {"20"}, "bay": {"Bay 1"}}, "--bay requires --parent"},
		{"unknown bay", map[string][]string{"parent": {"chassis-15"}, "bay": {"Bay 9"}}, `has no bay "Bay 9"`},
		{"full chassis", map[string][]string{"parent": {"full-15"}}, "no free bay"},
		{"unknown parent", map[string][]string{"parent": {"nope"}}, "not found as rack or device"},
	}
	for _, tc := range cases {
		inventory, rackID := cmdtest.InventoryWithRack("rack-01")
		cmdtest.AddChassisWithBays(inventory, rackID, "chassis-15", 15, "Bay 1", "Bay 2")
		fullID := cmdtest.AddChassisWithBays(inventory, rackID, "full-15", 16, "Bay 1")
		occupant := uuid.New()
		inventory.Devices[occupant] = &devicetypes.CaniDeviceType{ID: occupant, Name: "x", Parent: fullID, ParentBay: "Bay 1"}
		harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

		err := harness.Run(t, tc.flags, cmdtest.DeviceSlug)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want containing %q", tc.name, err, tc.want)
		}
		if harness.Store.Saves != 0 {
			t.Errorf("%s: Saves = %d, want 0", tc.name, harness.Store.Saves)
		}
	}
}

// TestAddDeviceParentByNameWithoutBays verifies --parent by name still works
// for a rack and for a parent device that has no bays.
//
// Why it matters: --parent previously took only a UUID; names must resolve,
// and parents without bays must keep the old plain-parent behavior. Inputs: a
// rack name, and a server name. Outputs: the device parented to each, with no
// bay recorded.
func TestAddDeviceParentByNameWithoutBays(t *testing.T) {
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	serverID := cmdtest.AddDevice(inventory, rackID, "server-01", 10)

	for _, parent := range []struct {
		ref string
		id  uuid.UUID
	}{{"rack-01", rackID}, {"server-01", serverID}} {
		harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)
		name := "child-of-" + parent.ref
		if err := harness.Run(t, map[string][]string{"parent": {parent.ref}, "name": {name}}, cmdtest.DeviceSlug); err != nil {
			t.Fatalf("add --parent %s: %v", parent.ref, err)
		}
		inventory = harness.Inventory()
		child := childrenOf(inventory, parent.id)[name]
		if child == nil || child.ParentBay != "" {
			t.Errorf("--parent %s: child = %+v, want parented with no bay", parent.ref, child)
		}
	}
}
