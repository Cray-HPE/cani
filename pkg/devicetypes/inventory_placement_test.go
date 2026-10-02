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

// chassisWithBays returns an inventory holding a racked chassis that declares
// device bays "bay1"/"bay2" and module bays "slot1"/"slot2"; the rack sits at
// a site so the chassis derives that location.
func chassisWithBays() (*Inventory, *CaniDeviceType) {
	inv := NewInventory()
	siteID, rackID, chassisID := uuid.New(), uuid.New(), uuid.New()
	inv.Locations[siteID] = &CaniLocationType{ID: siteID, Name: "site"}
	inv.Racks[rackID] = &CaniRackType{ID: rackID, Name: "rack", Location: siteID}
	chassis := &CaniDeviceType{
		ID: chassisID, Name: "chassis", Parent: rackID,
		DeviceBays: []DeviceBaySpec{{Name: "bay1"}, {Name: "bay2"}},
		ModuleBays: []ModuleBaySpec{{Name: "slot1"}, {Name: "slot2"}},
	}
	inv.Devices[chassisID] = chassis
	return inv, chassis
}

// siteOf returns the location the fixture chassis derives from its rack.
func siteOf(inv *Inventory, chassis *CaniDeviceType) uuid.UUID {
	return inv.Racks[chassis.Parent].Location
}

func joinWarnings(res *RelationshipResult) string { return strings.Join(res.Warnings, "\n") }

// TestPlacementAcceptsDeclaredBaysAndSpares verifies children in declared
// bays, modules in declared slots, and a spare module at a location pass
// without errors or warnings.
//
// Why it matters: the happy path must stay silent so the diagnostics that
// follow are meaningful; it also proves the spare shape (location, no parent)
// is a valid placement.
// Inputs: two blades in bay1/bay2, two modules in slot1/slot2 (one with its
// location authored equal to the chassis's), and a spare module at the site.
// Outputs: an empty RelationshipResult.
// Data choice: the authored-equal location exercises the legacy inherited
// location rule without triggering the conflict error.
func TestPlacementAcceptsDeclaredBaysAndSpares(t *testing.T) {
	inv, chassis := chassisWithBays()
	for i, bay := range []string{"bay1", "bay2"} {
		id := uuid.New()
		inv.Devices[id] = &CaniDeviceType{ID: id, Name: "blade" + string(rune('1'+i)), Parent: chassis.ID, ParentBay: bay}
	}
	one, two, spare := uuid.New(), uuid.New(), uuid.New()
	inv.Modules[one] = &CaniModuleType{ID: one, Name: "nic1", ParentDevice: chassis.ID, ModuleBayName: "slot1"}
	inv.Modules[two] = &CaniModuleType{ID: two, Name: "nic2", ParentDevice: chassis.ID, ModuleBayName: "slot2", Location: siteOf(inv, chassis)}
	inv.Modules[spare] = &CaniModuleType{ID: spare, Name: "spare-nic", Location: siteOf(inv, chassis)}

	res := inv.RebuildDerivedState()

	if len(res.Errors) != 0 || len(res.Warnings) != 0 {
		t.Fatalf("errors = %v, warnings = %v; want none", res.Errors, res.Warnings)
	}
	if !inv.Modules[spare].IsSpare() || inv.Modules[one].IsSpare() {
		t.Error("IsSpare must be true only for the location-only module")
	}
}

// TestPlacementRejectsUndeclaredDuplicateAndConflicting verifies an undeclared
// bay, a second claim on one bay, and a module whose authored location
// disagrees with its parent's are all errors.
//
// Why it matters: these are the two ways a seed would describe hardware
// Nautobot cannot install; each must block rather than be repaired. An
// installed module's authored location is stale data the parent's location
// overrides, so it must be reported without refusing every write to an
// inventory whose device moved.
// Inputs: a blade in "bay9", two blades both in "bay1", a module in "slot9",
// two modules both in "slot1", and a module in slot2 authored at a different
// location. Outputs: four errors naming the bays — one per undeclared bay,
// one for the second claimant of each contested bay — and one warning naming
// both locations.
// Data choice: every other attribute is valid so only the placement rules can
// explain each report.
func TestPlacementRejectsUndeclaredDuplicateAndConflicting(t *testing.T) {
	inv, chassis := chassisWithBays()
	elsewhere := uuid.New()
	inv.Locations[elsewhere] = &CaniLocationType{ID: elsewhere, Name: "elsewhere"}
	for name, bay := range map[string]string{"blade-undeclared": "bay9", "blade-a": "bay1", "blade-b": "bay1"} {
		id := uuid.New()
		inv.Devices[id] = &CaniDeviceType{ID: id, Name: name, Parent: chassis.ID, ParentBay: bay}
	}
	for name, mod := range map[string]*CaniModuleType{
		"nic-undeclared": {ParentDevice: chassis.ID, ModuleBayName: "slot9"},
		"nic-a":          {ParentDevice: chassis.ID, ModuleBayName: "slot1"},
		"nic-b":          {ParentDevice: chassis.ID, ModuleBayName: "slot1"},
		"nic-moved":      {ParentDevice: chassis.ID, ModuleBayName: "slot2", Location: elsewhere},
	} {
		mod.ID, mod.Name = uuid.New(), name
		inv.Modules[mod.ID] = mod
	}

	res := inv.RebuildDerivedState()

	if len(res.Errors) != 4 {
		t.Fatalf("errors = %d (%v), want 4", len(res.Errors), res.Errors)
	}
	msg := res.Err().Error()
	for _, want := range []string{
		`device bay "bay9" is not declared on parent "chassis"`,
		`device bay "bay1" on "chassis" is already occupied`,
		`module bay "slot9" is not declared on parent "chassis"`,
		`module bay "slot1" on "chassis" is already occupied`,
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("errors missing %q in:\n%s", want, msg)
		}
	}
	if want := `but authored at location ` + elsewhere.String(); !strings.Contains(joinWarnings(res), want) {
		t.Errorf("warnings missing %q in:\n%s", want, joinWarnings(res))
	}
}

// TestPlacementReportsMissingLegacyBayIdentity verifies a child device and a
// module installed without a named bay load with warnings, not errors, and
// that no bay is invented for them.
//
// Why it matters: inventories written before v1alpha8 carry no bay names;
// they must keep loading, and export must be told rather than guess.
// Inputs: a blade parented to the chassis with no ParentBay, and a module
// in the chassis with no ModuleBayName. Outputs: two warnings, zero errors,
// the device bay still empty and the module's effective bay equal to its name.
// Data choice: the chassis declares bays, which is exactly when guessing would
// be tempting.
func TestPlacementReportsMissingLegacyBayIdentity(t *testing.T) {
	inv, chassis := chassisWithBays()
	bladeID, modID := uuid.New(), uuid.New()
	inv.Devices[bladeID] = &CaniDeviceType{ID: bladeID, Name: "legacy-blade", Parent: chassis.ID}
	inv.Modules[modID] = &CaniModuleType{ID: modID, Name: "legacy-nic", ParentDevice: chassis.ID}

	res := inv.RebuildDerivedState()

	if len(res.Errors) != 0 {
		t.Fatalf("errors = %v, want none for legacy placement", res.Errors)
	}
	if len(res.Warnings) != 2 || !strings.Contains(joinWarnings(res), "without a named device bay") ||
		!strings.Contains(joinWarnings(res), "without a named module bay") {
		t.Errorf("warnings = %v, want one device-bay and one module-bay report", res.Warnings)
	}
	if inv.Devices[bladeID].ParentBay != "" || inv.Modules[modID].EffectiveModuleBayName() != "legacy-nic" {
		t.Error("placement must not be invented for legacy objects")
	}
}

// TestInterfaceIndexRecordsModuleOwner verifies a module's interface is
// indexed with ModuleID set to the module and DeviceID set to the parent,
// while a device interface leaves ModuleID empty, so OwnerID distinguishes
// same-named ports on two modules.
//
// Why it matters: typed ownership is what lets export create the port on the
// module and lets two modules with identical port names stay distinct.
// Inputs: a chassis with port "eth0" and two modules each with port "p1".
// Outputs: the device port has no ModuleID and OwnerID == chassis; each
// module port has ModuleID == its module and OwnerID differs between them.
// Data choice: identical port names on both modules make the owner the only
// distinguishing attribute.
func TestInterfaceIndexRecordsModuleOwner(t *testing.T) {
	inv, chassis := chassisWithBays()
	chassis.Interfaces = []InterfaceSpec{{Name: "eth0", Type: InterfacesElemTypeA1000BaseT}}
	modA, modB := uuid.New(), uuid.New()
	inv.Modules[modA] = &CaniModuleType{ID: modA, Name: "nic-a", ParentDevice: chassis.ID, ModuleBayName: "slot1",
		Interfaces: []InterfaceSpec{{Name: "p1", Type: InterfacesElemTypeA25GbaseXSfp28}}}
	inv.Modules[modB] = &CaniModuleType{ID: modB, Name: "nic-b", ParentDevice: chassis.ID, ModuleBayName: "slot2",
		Interfaces: []InterfaceSpec{{Name: "p1", Type: InterfacesElemTypeA25GbaseXSfp28}}}

	if err := inv.RebuildDerivedState().Err(); err != nil {
		t.Fatalf("RebuildDerivedState: %v", err)
	}

	owners := map[uuid.UUID]int{}
	for _, iface := range inv.Interfaces {
		switch iface.Name {
		case "eth0":
			if iface.ModuleID != uuid.Nil || iface.OwnerID() != chassis.ID || iface.DeviceID != chassis.ID {
				t.Errorf("device port = {device %v, module %v}, want chassis-owned", iface.DeviceID, iface.ModuleID)
			}
		case "p1":
			if iface.DeviceID != chassis.ID || iface.ModuleID == uuid.Nil || iface.OwnerID() != iface.ModuleID {
				t.Errorf("module port = {device %v, module %v}, want parent chassis and a module owner", iface.DeviceID, iface.ModuleID)
			}
			owners[iface.ModuleID]++
		}
	}
	if len(owners) != 2 || owners[modA] != 1 || owners[modB] != 1 {
		t.Errorf("module port owners = %v, want one p1 on each module", owners)
	}
}

// TestValidateModuleRelationshipsNilAndNoParent verifies validateModuleRelationships
// skips nil module entries and stays silent on a module that has no parent
// device, leaving the unplaced-module warning to validatePlacement so a spare
// stored at a location is not reported twice or at all.
//
// Why it matters: import data frequently contains modules not yet attached to a
// device; these must surface as a single warning rather than hard errors, and
// nil map slots must never panic the validator.
// Inputs: an inventory with a nil module entry and a module whose ParentDevice
// is uuid.Nil. Outputs: no errors and no warnings from the relationship pass,
// and exactly one warning from the placement pass. Data choice: pairing the nil
// entry with an unparented module covers both the continue guard and the
// hand-off to placement that the happy-path and missing-device tests skip.
func TestValidateModuleRelationshipsNilAndNoParent(t *testing.T) {
	inv := NewInventory()
	inv.Modules[uuid.New()] = nil
	inv.Modules[uuid.New()] = &CaniModuleType{Name: "floating"}

	res := inv.validateModuleRelationships()
	if res.HasErrors() || len(res.Warnings) != 0 {
		t.Fatalf("relationship pass = errors %v, warnings %v; want neither", res.Errors, res.Warnings)
	}
	placement := inv.validatePlacement()
	if placement.HasErrors() || len(placement.Warnings) != 1 {
		t.Fatalf("placement pass = errors %v, warnings %v; want exactly one unplaced warning", placement.Errors, placement.Warnings)
	}
}

// TestPlacementWarnsWhenLegacyBayNamesCollide verifies two modules that fall
// back to the module name as their bay, or one such module and one authored
// into the same bay, produce a warning and no error.
//
// Why it matters: before v1alpha9 two identical modules on one device were
// authored without bays and share the model name; treating that fallback as a
// hard claim would refuse every save of an upgraded inventory.
// Inputs: two bay-less modules both named "nic" on the chassis, plus a module
// authored into slot1 and a bay-less module named "slot1". Outputs: zero
// errors; warnings naming the "already occupied" bay twice plus the
// unnamed-bay report for each of the three bay-less modules.
// Data choice: the chassis declares bays so the fallback is reported as well,
// proving the collision does not escalate it.
func TestPlacementWarnsWhenLegacyBayNamesCollide(t *testing.T) {
	inv, chassis := chassisWithBays()
	for _, mod := range []*CaniModuleType{
		{Name: "nic", ParentDevice: chassis.ID},
		{Name: "nic", ParentDevice: chassis.ID},
		{Name: "authored", ParentDevice: chassis.ID, ModuleBayName: "slot1"},
		{Name: "slot1", ParentDevice: chassis.ID},
	} {
		mod.ID = uuid.New()
		inv.Modules[mod.ID] = mod
	}

	res := inv.RebuildDerivedState()

	if len(res.Errors) != 0 {
		t.Fatalf("errors = %v, want none for legacy bay collisions", res.Errors)
	}
	warnings := joinWarnings(res)
	if len(res.Warnings) != 5 || strings.Count(warnings, "is already occupied by") != 2 ||
		strings.Count(warnings, "without a named module bay") != 3 {
		t.Errorf("warnings = %v, want 2 collision and 3 unnamed-bay reports", res.Warnings)
	}
}

// TestPlacementReportsModuleWithNoBayAndNoName verifies an unnamed module
// installed without a bay on a parent that declares no bays is reported with
// the remedy and claims no bay, so two such modules do not collide.
//
// Why it matters: a module needs no name, so nothing stands in for a missing
// bay; export would otherwise send an empty bay name to Nautobot, and a claim
// on the empty name would report two such modules as occupying one bay.
// Inputs: a parent with no module bays holding two modules with neither a
// name nor a bay, and one named module without a bay. Outputs: no errors,
// exactly two warnings, both naming the remedy, and no occupancy report.
// Data choice: the named module proves the legacy name fallback still
// applies on a parent without bays and is not reported.
func TestPlacementReportsModuleWithNoBayAndNoName(t *testing.T) {
	inv, chassis := chassisWithBays()
	bare := uuid.New()
	inv.Devices[bare] = &CaniDeviceType{ID: bare, Name: "bare", Parent: chassis.Parent}
	for _, mod := range []*CaniModuleType{
		{ParentDevice: bare}, {ParentDevice: bare}, {Name: "legacy", ParentDevice: bare},
	} {
		mod.ID = uuid.New()
		inv.Modules[mod.ID] = mod
	}

	res := inv.RebuildDerivedState()

	if len(res.Errors) != 0 {
		t.Fatalf("errors = %v, want none", res.Errors)
	}
	warnings := joinWarnings(res)
	if len(res.Warnings) != 2 || strings.Count(warnings, "without a module bay or a name") != 2 ||
		strings.Contains(warnings, "already occupied") {
		t.Errorf("warnings = %v, want two remedy reports and no occupancy report", res.Warnings)
	}
}

// TestPlacementReportsBayNameOnSpare verifies a spare stored at a location
// that still names a module bay is accepted with a warning, not refused.
//
// Why it matters: a bay belongs to a parent device, so a bay on a spare is
// stale or mistaken data; the operator must be told, but a datastore holding
// such a spare must stay writable and the spare must stay a spare.
// Inputs: a spare at the site with ModuleBayName "slot1". Outputs: no
// errors, one warning naming the bay, IsSpare still true.
// Data choice: "slot1" is a bay the fixture chassis declares, proving the
// report does not depend on the bay being unknown.
func TestPlacementReportsBayNameOnSpare(t *testing.T) {
	inv, chassis := chassisWithBays()
	spare := &CaniModuleType{ID: uuid.New(), Location: siteOf(inv, chassis), ModuleBayName: "slot1"}
	inv.Modules[spare.ID] = spare

	res := inv.RebuildDerivedState()

	if len(res.Errors) != 0 || !spare.IsSpare() {
		t.Fatalf("errors = %v, spare = %v; want no errors and a spare", res.Errors, spare.IsSpare())
	}
	if len(res.Warnings) != 1 || !strings.Contains(joinWarnings(res), `names module bay "slot1"`) {
		t.Errorf("warnings = %v, want one naming the bay", res.Warnings)
	}
}
