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
package swap

import (
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// swapFixture returns an inventory with a racked node "cn-01" that declares
// module bay "PCIe1", NIC "nic-a" installed in it, NIC "nic-spare" stored at
// the site, and a second location "stock"; it returns the site and stock IDs.
func swapFixture(t *testing.T) (*devicetypes.Inventory, uuid.UUID, uuid.UUID) {
	t.Helper()
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	site := inventory.Racks[rackID].Location
	stock := uuid.New()
	inventory.Locations[stock] = &devicetypes.CaniLocationType{ID: stock, Name: "stock"}
	nodeID := uuid.New()
	inventory.Devices[nodeID] = &devicetypes.CaniDeviceType{
		ID: nodeID, Name: "cn-01", Parent: rackID, RackPosition: 10, Face: devicetypes.FaceFront, UHeight: 1,
		Type: devicetypes.TypeNode, ModuleBays: []devicetypes.ModuleBaySpec{{Name: "PCIe1", Position: "1"}},
	}
	installed := &devicetypes.CaniModuleType{
		ID: uuid.New(), Name: "nic-a", Model: "NIC", ParentDevice: nodeID, ModuleBayName: "PCIe1",
		Interfaces: []devicetypes.InterfaceSpec{{Name: "Port " + devicetypes.ModuleToken}},
	}
	spare := &devicetypes.CaniModuleType{
		ID: uuid.New(), Name: "nic-spare", Model: "NIC", Location: site,
		Interfaces: []devicetypes.InterfaceSpec{{Name: "Port " + devicetypes.ModuleToken}},
	}
	for _, mod := range []*devicetypes.CaniModuleType{installed, spare} {
		if err := inventory.AddModule(mod); err != nil {
			t.Fatalf("seed %s: %v", mod.Name, err)
		}
	}
	return inventory, site, stock
}

// cabledSwapFixture returns the swap fixture with a switch "sw-01" in the
// rack and two cables on nic-a's port "Port 1": "by-module" names nic-a as
// its owner, and "by-device" names cn-01 with the bay-derived port name, as
// `add cable --a-device cn-01` and connection imports record it.
func cabledSwapFixture(t *testing.T) (*devicetypes.Inventory, uuid.UUID, uuid.UUID) {
	t.Helper()
	inventory, site, stock := swapFixture(t)
	nic := moduleNamed(t, inventory, "nic-a")
	swID := uuid.New()
	inventory.Devices[swID] = &devicetypes.CaniDeviceType{
		ID: swID, Name: "sw-01", Parent: inventory.Devices[nic.ParentDevice].Parent, RackPosition: 20, Face: devicetypes.FaceFront, UHeight: 1,
		Type:       devicetypes.TypeSwitch,
		Interfaces: []devicetypes.InterfaceSpec{{ID: uuid.New(), Name: "swp1"}, {ID: uuid.New(), Name: "swp2"}},
	}
	cables := []*devicetypes.CaniCableType{
		{ID: uuid.New(), Label: "by-module", TerminationADevice: nic.ID, TerminationAPort: "Port 1", TerminationBDevice: swID, TerminationBPort: "swp1"},
		{ID: uuid.New(), Label: "by-device", TerminationADevice: nic.ParentDevice, TerminationAPort: "Port 1", TerminationBDevice: swID, TerminationBPort: "swp2"},
	}
	for _, cable := range cables {
		if err := inventory.AddCable(cable); err != nil {
			t.Fatalf("seed cable %s: %v", cable.Label, err)
		}
	}
	return inventory, site, stock
}

// moduleNamed returns the named module from the inventory.
func moduleNamed(t *testing.T, inventory *devicetypes.Inventory, name string) *devicetypes.CaniModuleType {
	t.Helper()
	for _, mod := range inventory.Modules {
		if mod.Name == name {
			return mod
		}
	}
	t.Fatalf("module %q not found", name)
	return nil
}

// moduleByName returns the named module from the store's inventory.
func moduleByName(t *testing.T, harness *cmdtest.Harness, name string) *devicetypes.CaniModuleType {
	t.Helper()
	return moduleNamed(t, harness.Inventory(), name)
}

// TestSwapModuleDryRunThenSwap verifies --dry-run reports the resulting
// placements without saving, and that the swap itself installs the spare in
// the bay, stores the replaced module at the parent's location, and saves
// once.
//
// Why it matters: this is the operator's one-step replacement of a failed
// module; a dry run that saved, or a swap that saved twice or stored the part
// elsewhere, would misreport what the next export moves.
// Inputs: the fixture, `swap module nic-a nic-spare --dry-run` then
// `swap module nic-a nic-spare` (the dry run first, because flag values
// persist on the shared command). Outputs: two "Would" lines and no save;
// then nic-spare in cn-01 bay PCIe1 with port "Port 1", nic-a a spare at the
// site, and one save.
// Data choice: the templated port name witnesses that the bay's position was
// applied to the incoming module through the command path.
func TestSwapModuleDryRunThenSwap(t *testing.T) {
	cmdtest.RequireNoProviders(t)
	inventory, site, _ := swapFixture(t)
	harness := cmdtest.New(t, NewCommand(), newModuleCommand(), inventory)

	if err := harness.Run(t, map[string][]string{flagDryRun: {"true"}}, "nic-a", "nic-spare"); err != nil {
		t.Fatalf("swap module --dry-run: unexpected error: %v", err)
	}
	out := harness.Out.String()
	if !strings.Contains(out, "Would install nic-spare in cn-01 bay PCIe1") || !strings.Contains(out, "Would store nic-a at location") {
		t.Errorf("dry-run output = %q, want both placement lines", out)
	}
	if harness.Store.Saves != 0 || !moduleByName(t, harness, "nic-spare").IsSpare() {
		t.Fatalf("dry run saved %d time(s) or changed the store; want no save and the spare still stored", harness.Store.Saves)
	}

	if err := harness.Run(t, map[string][]string{flagDryRun: {"false"}}, "nic-a", "nic-spare"); err != nil {
		t.Fatalf("swap module: unexpected error: %v", err)
	}
	spare, replaced := moduleByName(t, harness, "nic-spare"), moduleByName(t, harness, "nic-a")
	node := moduleByName(t, harness, "nic-spare").ParentDevice
	if harness.Inventory().Devices[node] == nil || spare.ModuleBayName != "PCIe1" || spare.Location != uuid.Nil || spare.Interfaces[0].Name != "Port 1" {
		t.Errorf("spare after swap = {parent %v, bay %q, location %v, port %q}; want installed in PCIe1 with port \"Port 1\"",
			spare.ParentDevice, spare.ModuleBayName, spare.Location, spare.Interfaces[0].Name)
	}
	if !replaced.IsSpare() || replaced.Location != site {
		t.Errorf("replaced module after swap = {parent %v, location %v}; want a spare at the site %v", replaced.ParentDevice, replaced.Location, site)
	}
	if harness.Store.Saves != 1 {
		t.Errorf("Saves = %d, want 1", harness.Store.Saves)
	}
}

// TestSwapModuleStoresAtNamedLocationAndRejectsUnknown verifies --location
// stores the replaced module at the named location, and that an unknown
// location or module is rejected before anything is saved.
//
// Why it matters: a failed part is often sent to a repair location rather
// than left at the site, and a typo must not store it somewhere else or
// half-apply the swap.
// Inputs: `swap module nic-a nic-spare --location nowhere`, then
// `--location stock`; then `swap module nic-a ghost` against a fresh fixture.
// Outputs: a "not found" error with no save; nic-a stored at stock with one
// save; a "no item found" error with no save.
// Data choice: the unknown-location case runs first on the same harness so a
// rejected flag value is proven not to leak into the datastore.
func TestSwapModuleStoresAtNamedLocationAndRejectsUnknown(t *testing.T) {
	inventory, _, stock := swapFixture(t)
	harness := cmdtest.New(t, NewCommand(), newModuleCommand(), inventory)

	err := harness.Run(t, map[string][]string{flagLocation: {"nowhere"}}, "nic-a", "nic-spare")
	if err == nil || !strings.Contains(err.Error(), `location "nowhere" not found`) {
		t.Errorf("unknown location error = %v, want not found", err)
	}
	if harness.Store.Saves != 0 {
		t.Fatalf("Saves after rejected location = %d, want 0", harness.Store.Saves)
	}

	if err := harness.Run(t, map[string][]string{flagLocation: {"stock"}}, "nic-a", "nic-spare"); err != nil {
		t.Fatalf("swap module --location stock: unexpected error: %v", err)
	}
	if replaced := moduleByName(t, harness, "nic-a"); replaced.Location != stock || !replaced.IsSpare() {
		t.Errorf("replaced module = {parent %v, location %v}; want a spare at stock %v", replaced.ParentDevice, replaced.Location, stock)
	}
	if harness.Store.Saves != 1 {
		t.Errorf("Saves = %d, want 1", harness.Store.Saves)
	}

	fresh, _, _ := swapFixture(t)
	harness = cmdtest.New(t, NewCommand(), newModuleCommand(), fresh)
	err = harness.Run(t, nil, "nic-a", "ghost")
	if err == nil || !strings.Contains(err.Error(), "no item found matching") || harness.Store.Saves != 0 {
		t.Errorf("unknown module error = %v (saves %d), want no item found and no save", err, harness.Store.Saves)
	}
}

// TestSwapModuleMovesCablesToTheSpare verifies the command re-terminates the
// cables on the replaced module's ports at the spare's same-named ports, for
// a cable that names the module and one that names the parent device, and
// that --dry-run reports the move without saving.
//
// Why it matters: a cable addressed through the parent device, the form
// `add cable --a-device <device>` and connection imports record, left the
// swap refused with a relationship error about the cable and rolled back;
// the store's JSON round trip also proves the re-terminated ends persist.
// Inputs: the cabled fixture; `swap module nic-a nic-spare --dry-run`, then
// `swap module nic-a nic-spare`.
// Outputs: "Would move 2 cable(s) to the ports of nic-spare", no save and
// both cables still on nic-a; then both cables resolve to nic-spare's
// "Port 1" interface, "by-module" owned by nic-spare and "by-device" still
// naming cn-01, nic-a with no cables, and one save.
// Data choice: two cables of different owner form on one port prove the
// command path moves both the same way.
func TestSwapModuleMovesCablesToTheSpare(t *testing.T) {
	inventory, _, _ := cabledSwapFixture(t)
	node := moduleNamed(t, inventory, "nic-a").ParentDevice
	harness := cmdtest.New(t, NewCommand(), newModuleCommand(), inventory)

	if err := harness.Run(t, map[string][]string{flagDryRun: {"true"}}, "nic-a", "nic-spare"); err != nil {
		t.Fatalf("swap module --dry-run: unexpected error: %v", err)
	}
	if out := harness.Out.String(); !strings.Contains(out, "Would move 2 cable(s) to the ports of nic-spare") {
		t.Errorf("dry-run output = %q, want the cable move reported", out)
	}
	if n := len(harness.Inventory().CablesOnModule(moduleByName(t, harness, "nic-a").ID)); harness.Store.Saves != 0 || n != 2 {
		t.Fatalf("dry run saved %d time(s) and left %d cable(s) on nic-a; want no save and both cables", harness.Store.Saves, n)
	}

	if err := harness.Run(t, map[string][]string{flagDryRun: {"false"}}, "nic-a", "nic-spare"); err != nil {
		t.Fatalf("swap module: unexpected error: %v", err)
	}
	saved := harness.Inventory()
	spare, replaced := moduleByName(t, harness, "nic-spare"), moduleByName(t, harness, "nic-a")
	want := spare.Interfaces[0].ID
	byModule, byDevice := saved.FindCableByLabel("by-module"), saved.FindCableByLabel("by-device")
	if byModule == nil || byModule.TerminationADevice != spare.ID || byModule.TerminationAPort != "Port 1" || byModule.TerminationA != want {
		t.Errorf("by-module end = %+v, want owned by nic-spare on \"Port 1\" %v", byModule, want)
	}
	if byDevice == nil || byDevice.TerminationADevice != node || byDevice.TerminationAPort != "Port 1" || byDevice.TerminationA != want {
		t.Errorf("by-device end = %+v, want cn-01, \"Port 1\" and nic-spare's interface %v", byDevice, want)
	}
	if n := len(saved.CablesOnModule(replaced.ID)); n != 0 {
		t.Errorf("cables on the replaced module = %d, want 0", n)
	}
	if harness.Store.Saves != 1 {
		t.Errorf("Saves = %d, want 1", harness.Store.Saves)
	}
}
