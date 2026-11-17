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

// Library slugs the role tests add; each is a real entry so the add runs
// the same lookup a user's would.
const (
	roleTestRackSlug   = "hpe-42u-800mmx1200mm-g2-enterprise-shock-rack"
	roleTestModuleSlug = "hpe-2p-25gbe-pcie-nic"
	roleTestCableSlug  = "hpe-aruba-100g-qsfp28-1m-dac"
)

// TestAddStampsRoleOnDeviceRackAndModule verifies --role sets the role field
// of a device, a rack and a module, and that a later add without the flag
// leaves the role empty.
//
// Why it matters: `add` had no way to set a role, so operators wrote it into
// provider metadata, where one exporter read it back for devices only and
// nothing else in cani could see it.
// Inputs: `add device --role LeafSwitch`, `add rack --role StorageRack`, a
// spare `add module --role Spare`, then `add rack` with no role.
// Outputs: each item carries its role in the role field; the second rack
// carries none.
// Data choice: three nouns with three different roles show the flag is read
// per command, and the roleless rack shows an unset flag stamps nothing.
func TestAddStampsRoleOnDeviceRackAndModule(t *testing.T) {
	// Arrange.
	inventory, _ := cmdtest.InventoryWithRack("rack-01")
	var siteName string
	for _, location := range inventory.Locations {
		siteName = location.Name
	}

	// Act.
	devices := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)
	if err := devices.Run(t, map[string][]string{
		"rack": {"rack-01"}, "position": {"10"}, "name": {"leaf-01"}, flagRole: {"LeafSwitch"},
	}, cmdtest.DeviceSlug); err != nil {
		t.Fatalf("add device --role: unexpected error: %v", err)
	}
	racks := cmdtest.New(t, NewCommand(), newRackAddCommand(), devices.Inventory())
	if err := racks.Run(t, map[string][]string{"name": {"rack-02"}, flagRole: {"StorageRack"}}, roleTestRackSlug); err != nil {
		t.Fatalf("add rack --role: unexpected error: %v", err)
	}
	modules := cmdtest.New(t, NewCommand(), newModuleCommand(), racks.Inventory())
	if err := modules.Run(t, map[string][]string{
		"spare": {"true"}, "location": {siteName}, "name": {"nic-spare"}, flagRole: {"Spare"},
	}, roleTestModuleSlug); err != nil {
		t.Fatalf("add module --role: unexpected error: %v", err)
	}
	plain := cmdtest.New(t, NewCommand(), newRackAddCommand(), modules.Inventory())
	if err := plain.Run(t, map[string][]string{"name": {"rack-03"}}, roleTestRackSlug); err != nil {
		t.Fatalf("add rack without --role: unexpected error: %v", err)
	}

	// Assert.
	final := plain.Inventory()
	if got := deviceNamed(t, final, "leaf-01").Role; got != "LeafSwitch" {
		t.Errorf("device role = %q, want LeafSwitch", got)
	}
	rackRoles := map[string]string{}
	for _, rack := range final.Racks {
		rackRoles[rack.Name] = rack.Role
	}
	if rackRoles["rack-02"] != "StorageRack" || rackRoles["rack-03"] != "" {
		t.Errorf("rack roles = %v, want rack-02 StorageRack and rack-03 empty", rackRoles)
	}
	if len(final.Modules) != 1 {
		t.Fatalf("modules = %d, want 1", len(final.Modules))
	}
	for _, module := range final.Modules {
		if module.Role != "Spare" {
			t.Errorf("module role = %q, want Spare", module.Role)
		}
	}
}

// TestAddBySlugStampsRoleAndRefusesItForACable verifies the bare
// `add <slug>` form stamps --role on what it resolves to, and refuses the
// flag for a cable instead of dropping it.
//
// Why it matters: the bare form picks the noun from the slug, so a flag that
// one noun cannot store must fail loudly; a role typed for a cable would
// otherwise vanish without a word.
// Inputs: `add <device slug> --name leaf-02 --role LeafSwitch`, then
// `add <cable slug> --role LeafSwitch`.
// Outputs: a device with the role; an error naming --role for the cable, no
// cable and no second save.
// Data choice: the same role on both commands makes the noun the only
// difference between the accepted and the refused add.
func TestAddBySlugStampsRoleAndRefusesItForACable(t *testing.T) {
	// Arrange.
	inventory, _ := cmdtest.InventoryWithRack("rack-01")
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	// Act.
	deviceErr := harness.Execute(t, "add", cmdtest.DeviceSlug, "--name", "leaf-02", "--role", "LeafSwitch")
	savesAfterDevice := harness.Store.Saves
	cableErr := harness.Execute(t, "add", roleTestCableSlug, "--role", "LeafSwitch")

	// Assert.
	if deviceErr != nil {
		t.Fatalf("add <device slug> --role: unexpected error: %v", deviceErr)
	}
	if got := deviceNamed(t, harness.Inventory(), "leaf-02").Role; got != "LeafSwitch" {
		t.Errorf("device role = %q, want LeafSwitch", got)
	}
	if cableErr == nil || !strings.Contains(cableErr.Error(), "--role does not apply to a cable") {
		t.Errorf("add <cable slug> --role error = %v, want a refusal naming --role", cableErr)
	}
	if harness.Store.Saves != savesAfterDevice || len(harness.Inventory().Cables) != 0 {
		t.Errorf("after the refused cable: saves = %d (was %d), cables = %d; want nothing written",
			harness.Store.Saves, savesAfterDevice, len(harness.Inventory().Cables))
	}
}
