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
package devicetypes

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// positionedChassis returns the fixture chassis with bay positions, so a port
// template resolves to a name that records which bay the module sits in.
func positionedChassis() (*Inventory, *CaniDeviceType, uuid.UUID) {
	inv, chassis := chassisWithBays()
	chassis.ModuleBays = []ModuleBaySpec{{Name: "slot1", Position: "1"}, {Name: "slot2", Position: "2"}}
	return inv, chassis, siteOf(inv, chassis)
}

// templatedModule returns a module whose single port is named after its bay.
func templatedModule(name string) *CaniModuleType {
	return &CaniModuleType{ID: uuid.New(), Name: name, Model: "nic", Interfaces: []InterfaceSpec{{Name: "p" + ModuleToken}}}
}

// placementOf reports a module's placement and first port name on one line.
func placementOf(mod *CaniModuleType) string {
	parent := "-"
	if mod.ParentDevice != uuid.Nil {
		parent = "device"
	}
	location := "-"
	if mod.Location != uuid.Nil {
		location = "location"
	}
	return parent + " " + mod.ModuleBayName + " " + location + " " + mod.Interfaces[0].Name
}

// TestPlaceModuleMovesPlacementAsOneUnit verifies PlaceModule installs a spare
// into a declared bay, moves an installed module between bays, and stores an
// installed module at a location, each time replacing the whole placement and
// resolving the port names the new bay gives the module.
//
// Why it matters: a module that keeps its old location when installed, or its
// old bay when stored, fails placement validation and is refused by export;
// the port names must follow the bay as they do for `add module`.
// Inputs: a templated module stored at the site, placed into "SLOT2", then
// into "slot1", then back at the site.
// Outputs: installed in "slot2" (declared spelling) with port "p2" and no
// location; then "slot1" with "p1"; then a spare with no bay and port "p1".
// Data choice: the upper-case bay proves the parent's spelling is stored, and
// the positioned bays make the port name a witness of the move.
func TestPlaceModuleMovesPlacementAsOneUnit(t *testing.T) {
	inv, chassis, site := positionedChassis()
	mod := templatedModule("nic")
	mod.Location = site
	if err := inv.AddModule(mod); err != nil {
		t.Fatalf("seed spare: %v", err)
	}
	steps := []struct {
		target ModulePlacement
		want   string
	}{
		{ModulePlacement{ParentDevice: chassis.ID, Bay: "SLOT2"}, "device slot2 - p2"},
		{ModulePlacement{ParentDevice: chassis.ID, Bay: "slot1"}, "device slot1 - p1"},
		{ModulePlacement{Location: site}, "-  location p1"},
	}
	for _, step := range steps {
		if err := inv.PlaceModule(mod.ID, step.target); err != nil {
			t.Fatalf("PlaceModule(%+v): %v", step.target, err)
		}
		if got := placementOf(mod); got != step.want {
			t.Errorf("after PlaceModule(%+v) = %q, want %q", step.target, got, step.want)
		}
	}
	if res := inv.validatePlacement(); len(res.Errors) != 0 || len(res.Warnings) != 0 {
		t.Errorf("placement diagnostics after moves = %v / %v, want none", res.Errors, res.Warnings)
	}
}

// TestPlaceModuleRefusesAndRestores verifies every placement PlaceModule
// rejects leaves the module exactly where it was, including the port names
// an attempted bay would have given it.
//
// Why it matters: a half-applied move would leave the datastore both
// installed and stored, or with ports named for a bay the module is not in.
// Inputs: a spare at the site, placed into: an undeclared bay, an occupied
// bay, both a device and a location, neither, an unknown device, no bay on a
// parent that declares bays, an unknown location, and a location whose types
// exclude modules; plus an unknown module ID.
// Outputs: an error naming each rule, and the spare unchanged after each.
// Data choice: the occupied bay is the one case refused only after the
// placement is applied, so it proves the snapshot restore rather than the
// up-front checks.
func TestPlaceModuleRefusesAndRestores(t *testing.T) {
	inv, chassis, site := positionedChassis()
	racksOnly := uuid.New()
	inv.Locations[racksOnly] = &CaniLocationType{ID: racksOnly, Name: "racks-only", ContentTypes: []string{"rack"}}
	occupant := templatedModule("occupant")
	occupant.ParentDevice, occupant.ModuleBayName = chassis.ID, "slot1"
	spare := templatedModule("spare")
	spare.Location = site
	for _, mod := range []*CaniModuleType{occupant, spare} {
		if err := inv.AddModule(mod); err != nil {
			t.Fatalf("seed %s: %v", mod.Name, err)
		}
	}
	before := placementOf(spare)

	cases := []struct {
		name   string
		target ModulePlacement
		want   string
	}{
		{"undeclared bay", ModulePlacement{ParentDevice: chassis.ID, Bay: "slot9"}, `has no module bay "slot9" (bays: slot1, slot2)`},
		{"occupied bay", ModulePlacement{ParentDevice: chassis.ID, Bay: "slot1"}, `module bay "slot1" on "chassis" is already`},
		{"device and location", ModulePlacement{ParentDevice: chassis.ID, Bay: "slot2", Location: site}, "not both"},
		{"neither", ModulePlacement{}, "name a parent device or a location"},
		{"unknown device", ModulePlacement{ParentDevice: uuid.New(), Bay: "slot2"}, "not found"},
		{"no bay on a parent with bays", ModulePlacement{ParentDevice: chassis.ID}, "declares module bays; name the bay"},
		{"unknown location", ModulePlacement{Location: uuid.New()}, "not found"},
		{"location excludes modules", ModulePlacement{Location: racksOnly}, `content type "module" is not permitted`},
	}
	for _, tc := range cases {
		err := inv.PlaceModule(spare.ID, tc.target)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error = %v, want %q", tc.name, err, tc.want)
		}
		if got := placementOf(spare); got != before {
			t.Errorf("%s: spare after refusal = %q, want unchanged %q", tc.name, got, before)
		}
	}
	if err := inv.PlaceModule(uuid.New(), ModulePlacement{Location: site}); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("unknown module error = %v, want not found", err)
	}
	if res := inv.validatePlacement(); len(res.Errors) != 0 {
		t.Errorf("placement errors after refusals = %v, want none", res.Errors)
	}
}

// TestSwapModulesExchangesPlacements verifies SwapModules installs the spare
// in the outgoing module's bay and stores the outgoing module at the parent's
// location by default or at a named location, with port names following.
//
// Why it matters: this is the model behind `swap module`, the one-step
// replacement of a failed part by a spare the reviewer asked for.
// Inputs: a templated module in slot1 and a templated spare at the site;
// swapped with no location, then swapped back into a second location.
// Outputs: the spare in slot1 with port "p1" and no location, the outgoing
// module at the site with no bay and port "p1"; after the second swap the
// roles reverse and the stored module sits at the named location.
// Data choice: the second swap reuses the pair in reverse, proving a stored
// module is a valid spare for the next swap.
func TestSwapModulesExchangesPlacements(t *testing.T) {
	inv, chassis, site := positionedChassis()
	stock := uuid.New()
	inv.Locations[stock] = &CaniLocationType{ID: stock, Name: "stock"}
	installed, spare := templatedModule("bad"), templatedModule("spare")
	installed.ParentDevice, installed.ModuleBayName = chassis.ID, "slot1"
	spare.Location = site
	for _, mod := range []*CaniModuleType{installed, spare} {
		if err := inv.AddModule(mod); err != nil {
			t.Fatalf("seed %s: %v", mod.Name, err)
		}
	}

	if err := inv.SwapModules(installed.ID, spare.ID, uuid.Nil); err != nil {
		t.Fatalf("SwapModules: %v", err)
	}
	if got := placementOf(spare); got != "device slot1 - p1" {
		t.Errorf("spare after swap = %q, want installed in slot1", got)
	}
	if got := placementOf(installed); got != "-  location p1" || installed.Location != site {
		t.Errorf("outgoing after swap = %q at %s, want stored at the parent's site %s", got, installed.Location, site)
	}

	if err := inv.SwapModules(spare.ID, installed.ID, stock); err != nil {
		t.Fatalf("reverse SwapModules: %v", err)
	}
	if installed.ParentDevice != chassis.ID || installed.ModuleBayName != "slot1" || spare.Location != stock || !spare.IsSpare() {
		t.Errorf("reverse swap = installed {%v %q} spare {%v}; want installed back in slot1 and spare at stock",
			installed.ParentDevice, installed.ModuleBayName, spare.Location)
	}
	if res := inv.validatePlacement(); len(res.Errors) != 0 || len(res.Warnings) != 0 {
		t.Errorf("placement diagnostics after swaps = %v / %v, want none", res.Errors, res.Warnings)
	}
}

// TestSwapModulesRefusesInvalidPairs verifies the pair and target checks, and
// that a swap the rebuilt inventory refuses restores both modules.
//
// Why it matters: a swap must never install a spare into a bay named after the
// module it replaces, mix hardware types silently, or leave one module moved
// when the other is refused.
// Inputs: against the fixture, swaps of: the same module twice, a spare as the
// installed side, an installed module as the spare side, a GPU spare for a NIC,
// an outgoing module with no named bay, a parent with no location and no
// --location, a location that excludes modules; and a swap into a bay a
// third module also claims.
// Outputs: an error naming each rule; for the contested bay, both modules
// unchanged afterwards.
// Data choice: the contested bay is the one refusal that happens after the
// placements are applied, so it exercises the two-module restore.
func TestSwapModulesRefusesInvalidPairs(t *testing.T) {
	inv, chassis, site := positionedChassis()
	racksOnly := uuid.New()
	inv.Locations[racksOnly] = &CaniLocationType{ID: racksOnly, Name: "racks-only", ContentTypes: []string{"rack"}}
	installed, spare, gpu := templatedModule("nic-a"), templatedModule("nic-spare"), templatedModule("gpu-spare")
	installed.ParentDevice, installed.ModuleBayName = chassis.ID, "slot1"
	spare.Location, gpu.Location, gpu.Model = site, site, "gpu"
	for _, mod := range []*CaniModuleType{installed, spare, gpu} {
		if err := inv.AddModule(mod); err != nil {
			t.Fatalf("seed %s: %v", mod.Name, err)
		}
	}
	legacy := &CaniModuleType{ID: uuid.New(), Name: "legacy", ParentDevice: chassis.ID}
	inv.Modules[legacy.ID] = legacy
	homeless := &CaniDeviceType{ID: uuid.New(), Name: "homeless", ModuleBays: []ModuleBaySpec{{Name: "s1"}}}
	inv.Devices[homeless.ID] = homeless
	stranded := &CaniModuleType{ID: uuid.New(), Name: "stranded", ParentDevice: homeless.ID, ModuleBayName: "s1"}
	inv.Modules[stranded.ID] = stranded

	cases := []struct {
		name                 string
		installedID, spareID uuid.UUID
		storeAt              uuid.UUID
		want                 string
	}{
		{"same module", installed.ID, installed.ID, uuid.Nil, "cannot be swapped with itself"},
		{"spare as installed side", spare.ID, gpu.ID, uuid.Nil, "is not installed in a device"},
		{"installed as spare side", installed.ID, legacy.ID, uuid.Nil, "is not a spare stored at a location"},
		{"different hardware types", installed.ID, gpu.ID, uuid.Nil, "swap like for like"},
		{"outgoing module without a bay", legacy.ID, spare.ID, uuid.Nil, "has no named bay"},
		{"parent without a location", stranded.ID, spare.ID, uuid.Nil, "has no location to store"},
		{"location excludes modules", installed.ID, spare.ID, racksOnly, `content type "module" is not permitted`},
		{"unknown module", uuid.New(), spare.ID, uuid.Nil, "not found"},
	}
	for _, tc := range cases {
		err := inv.SwapModules(tc.installedID, tc.spareID, tc.storeAt)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error = %v, want %q", tc.name, err, tc.want)
		}
	}

	rival := &CaniModuleType{ID: uuid.New(), Name: "rival", ParentDevice: chassis.ID, ModuleBayName: "slot1"}
	inv.Modules[rival.ID] = rival
	err := inv.SwapModules(installed.ID, spare.ID, uuid.Nil)
	if err == nil || !strings.Contains(err.Error(), "already occupied") {
		t.Errorf("contested bay error = %v, want already occupied", err)
	}
	if installed.ParentDevice != chassis.ID || installed.ModuleBayName != "slot1" || installed.Location != uuid.Nil {
		t.Errorf("outgoing module after refused swap = {%v %q %v}, want unchanged in slot1", installed.ParentDevice, installed.ModuleBayName, installed.Location)
	}
	if got := placementOf(spare); got != "-  location p"+ModuleToken {
		t.Errorf("spare after refused swap = %q, want unchanged spare with its template port", got)
	}
}
