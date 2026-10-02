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

// newModule returns a module with a fresh ID and one port, so a rollback can
// be observed through the interface index as well as the module map.
func newModule(name string) *CaniModuleType {
	return &CaniModuleType{ID: uuid.New(), Name: name, Interfaces: []InterfaceSpec{{Name: "p1"}}}
}

// TestAddModuleRejectsInvalidPlacementAndRollsBack verifies AddModule refuses
// a module the placement rules reject and leaves no trace of it, while a
// module in a declared free bay is accepted.
//
// Why it matters: AddModule is the authoring path behind `add module`; the
// placement rules only protect the datastore if a rejected module is not
// persisted and does not leave its ports in the interface index.
// Inputs: against the fixture chassis, a module in undeclared "slot9", one in
// "slot1", a second authored into "slot1", a spare at a location that does not
// exist, and a spare at a location whose content types exclude modules.
// Outputs: four errors naming the rule that fired; only the slot1 module is
// present afterwards and only its port is indexed.
// Data choice: every rejected module carries a port so the rollback of the
// derived index is proven, not only the removal from the module map.
func TestAddModuleRejectsInvalidPlacementAndRollsBack(t *testing.T) {
	inv, chassis := chassisWithBays()
	noModules := uuid.New()
	inv.Locations[noModules] = &CaniLocationType{ID: noModules, Name: "racks-only", ContentTypes: []string{"rack"}}
	accepted := newModule("nic-ok")
	accepted.ParentDevice, accepted.ModuleBayName = chassis.ID, "slot1"
	if err := inv.AddModule(accepted); err != nil {
		t.Fatalf("declared free bay rejected: %v", err)
	}

	cases := map[string]*CaniModuleType{
		`module bay "slot9" is not declared`:         {ParentDevice: chassis.ID, ModuleBayName: "slot9"},
		`module bay "slot1" on "chassis" is already`: {ParentDevice: chassis.ID, ModuleBayName: "slot1"},
		"which does not exist":                       {Location: uuid.New()},
		`content type "module" is not permitted`:     {Location: noModules},
	}
	for want, mod := range cases {
		mod.ID, mod.Name, mod.Interfaces = uuid.New(), "rejected", []InterfaceSpec{{Name: "p1"}}
		err := inv.AddModule(mod)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("AddModule(%+v) error = %v, want %q", mod, err, want)
		}
		if inv.Modules[mod.ID] != nil {
			t.Errorf("rejected module %q was kept in the inventory", want)
		}
	}
	if len(inv.Modules) != 1 || len(inv.Interfaces) != 1 {
		t.Errorf("modules = %d, interfaces = %d after rollbacks; want 1 and 1", len(inv.Modules), len(inv.Interfaces))
	}
}

// TestAddModuleReportsOwnErrorsFirst verifies a rejected module's error names
// only that module's problems when it has any, and that an inventory broken
// elsewhere still refuses a module whose own placement is valid.
//
// Why it matters: an operator fixing one `add module` should see that
// module's fault, not every pre-existing fault; an inventory with errors must
// still not accept further writes, matching AddDevices and AddCable.
// Inputs: a chassis with a child in undeclared device bay "bay9", then a
// module in undeclared "slot9" and a module in declared "slot1".
// Outputs: the first error mentions slot9 and not bay9; the second mentions
// bay9 and the module is not added.
// Data choice: a device-bay fault is unrelated to modules, so its presence or
// absence in each message isolates the scoping rule.
func TestAddModuleReportsOwnErrorsFirst(t *testing.T) {
	inv, chassis := chassisWithBays()
	bladeID := uuid.New()
	inv.Devices[bladeID] = &CaniDeviceType{ID: bladeID, Name: "blade", Parent: chassis.ID, ParentBay: "bay9"}

	bad := newModule("bad")
	bad.ParentDevice, bad.ModuleBayName = chassis.ID, "slot9"
	err := inv.AddModule(bad)
	if err == nil || !strings.Contains(err.Error(), "slot9") || strings.Contains(err.Error(), "bay9") {
		t.Errorf("own-fault error = %v, want slot9 only", err)
	}

	good := newModule("good")
	good.ParentDevice, good.ModuleBayName = chassis.ID, "slot1"
	err = inv.AddModule(good)
	if err == nil || !strings.Contains(err.Error(), "bay9") || inv.Modules[good.ID] != nil {
		t.Errorf("broken-inventory error = %v (kept=%v), want bay9 and no insertion", err, inv.Modules[good.ID] != nil)
	}
}

// TestModuleBayNamesCompareCaseInsensitively verifies a module bay authored in
// a different case than the parent declares is accepted, that two spellings
// of one bay are one claim, and that a parent saved without bay specs is
// judged by its library type's bays.
//
// Why it matters: ResolveModuleNames and InstallInBay already admit any case,
// so the placement rules must not report a bay those paths accepted; and a
// parent imported without bay specs must not have its modules' bays checked
// against an empty list.
// Inputs: modules in "SLOT1" and "slot1" on the fixture chassis; a module in
// the library bay, in upper case, on a parent carrying only a library slug.
// Outputs: one "already occupied" error and no "not declared" error for the
// chassis pair; no error for the library-backed parent.
// Data choice: the library-backed case uses a real embedded type so the
// fallback is exercised against data that ships with the binary.
func TestModuleBayNamesCompareCaseInsensitively(t *testing.T) {
	inv, chassis := chassisWithBays()
	for _, bay := range []string{"SLOT1", "slot1"} {
		id := uuid.New()
		inv.Modules[id] = &CaniModuleType{ID: id, Name: "nic-" + bay, ParentDevice: chassis.ID, ModuleBayName: bay}
	}
	res := inv.validatePlacement()
	msg := res.Err()
	if len(res.Errors) != 1 || !strings.Contains(msg.Error(), "is already occupied") {
		t.Errorf("errors = %v, want exactly one collision and no undeclared report", res.Errors)
	}

	slug, bay := libraryTypeWithModuleBay(t)
	lib := NewInventory()
	parentID, modID := uuid.New(), uuid.New()
	lib.Devices[parentID] = &CaniDeviceType{ID: parentID, Name: "lib-parent", Slug: slug}
	lib.Modules[modID] = &CaniModuleType{ID: modID, Name: "nic", ParentDevice: parentID, ModuleBayName: strings.ToUpper(bay)}
	if res := lib.validatePlacement(); res.HasErrors() {
		t.Errorf("library-backed parent errors = %v, want none for bay %q", res.Errors, bay)
	}
}

// TestResolveModuleBayNameCanonicalises verifies ResolveModuleBayName returns
// the declared spelling for any case, passes through an empty name or a
// parent without bays, and rejects an undeclared bay naming the declared ones.
//
// Why it matters: the authoring commands store what this returns, so the
// datastore and the Nautobot ModuleBay it creates carry one spelling per bay.
// Inputs: "SLOT2", "", and "slot9" against the fixture chassis; "anything"
// against a device with no bays and against an unknown parent.
// Outputs: "slot2"; "" ; an error listing slot1 and slot2; "anything" twice.
// Data choice: the unknown-parent case is what `add module` passes when the
// device argument did not resolve, which must not become a bay error.
func TestResolveModuleBayNameCanonicalises(t *testing.T) {
	inv, chassis := chassisWithBays()
	bare := uuid.New()
	inv.Devices[bare] = &CaniDeviceType{ID: bare, Name: "bare"}

	if got, err := inv.ResolveModuleBayName(chassis.ID, "SLOT2"); err != nil || got != "slot2" {
		t.Errorf("SLOT2 = %q, %v; want slot2", got, err)
	}
	if got, err := inv.ResolveModuleBayName(chassis.ID, ""); err != nil || got != "" {
		t.Errorf("empty = %q, %v; want empty", got, err)
	}
	if _, err := inv.ResolveModuleBayName(chassis.ID, "slot9"); err == nil || !strings.Contains(err.Error(), "slot1, slot2") {
		t.Errorf("slot9 error = %v, want the declared bays listed", err)
	}
	for _, parent := range []uuid.UUID{bare, uuid.New()} {
		if got, err := inv.ResolveModuleBayName(parent, "anything"); err != nil || got != "anything" {
			t.Errorf("no-bays parent %s = %q, %v; want pass-through", parent, got, err)
		}
	}
}

// TestValidateAcceptsSparesAndChecksTheirLocation verifies Inventory.Validate
// accepts a module with no parent device and rejects a module whose location
// does not exist.
//
// Why it matters: Validate is the referential-integrity contract; a spare is
// a valid shape since bays were modelled, and a location is a forward FK that
// Nautobot protects, so a missing one is a broken reference.
// Inputs: a spare at the fixture site, then a module whose Location is a
// random UUID. Outputs: nil, then an error naming the module and "location".
// Data choice: the second module is also parentless so the only broken
// reference is the location.
func TestValidateAcceptsSparesAndChecksTheirLocation(t *testing.T) {
	inv, chassis := chassisWithBays()
	spareID := uuid.New()
	inv.Modules[spareID] = &CaniModuleType{ID: spareID, Name: "spare", Location: siteOf(inv, chassis)}
	if err := inv.Validate(); err != nil {
		t.Fatalf("Validate() with a spare = %v, want nil", err)
	}

	lostID := uuid.New()
	inv.Modules[lostID] = &CaniModuleType{ID: lostID, Name: "lost", Location: uuid.New()}
	err := inv.Validate()
	if err == nil || !strings.Contains(err.Error(), `module "lost"`) || !strings.Contains(err.Error(), "location") {
		t.Errorf("Validate() with a lost spare = %v, want a module location error", err)
	}
}

// TestRemoveLocationProtectsSpares verifies a location holding a spare module
// cannot be removed, and that a spare whose location is gone is a placement
// error rather than a silent dangling reference.
//
// Why it matters: Nautobot's Module.location is on_delete=PROTECT; mirroring
// it keeps the inventory exportable, and reporting the dangling case covers
// inventories edited by hand or imported in that state.
// Inputs: an empty location with one spare; then the spare re-pointed at a
// UUID that is not a location. Outputs: an error counting one spare module
// and the location still present; then one placement error.
// Data choice: an otherwise empty location isolates the spare as the only
// reason the removal is refused.
func TestRemoveLocationProtectsSpares(t *testing.T) {
	inv := NewInventory()
	storeID, spareID := uuid.New(), uuid.New()
	inv.Locations[storeID] = &CaniLocationType{ID: storeID, Name: "store"}
	inv.Modules[spareID] = &CaniModuleType{ID: spareID, Name: "spare", Location: storeID}

	err := inv.RemoveLocation(storeID)
	if err == nil || !strings.Contains(err.Error(), "1 spare module(s)") || inv.Locations[storeID] == nil {
		t.Errorf("RemoveLocation = %v (kept=%v), want a spare-module refusal", err, inv.Locations[storeID] != nil)
	}

	inv.Modules[spareID].Location = uuid.New()
	res := inv.validatePlacement()
	if len(res.Errors) != 1 || !strings.Contains(res.Err().Error(), "which does not exist") {
		t.Errorf("dangling spare errors = %v, want one missing-location error", res.Errors)
	}
}

// libraryTypeWithModuleBay returns a registered device type that declares
// module bays, and the name of its first module bay.
func libraryTypeWithModuleBay(t *testing.T) (string, string) {
	t.Helper()
	for slug, dt := range allDeviceTypes {
		if len(dt.ModuleBays) > 0 {
			return slug, dt.ModuleBays[0].Name
		}
	}
	t.Skip("no library device type with module bays")
	return "", ""
}

// mergeModules merges incoming modules into inv through MergeTransformResult
// and returns the summary and error.
func mergeModules(inv *Inventory, mods ...*CaniModuleType) (*TransformMergeSummary, error) {
	incoming := &TransformResult{Modules: map[uuid.UUID]*CaniModuleType{}}
	for _, mod := range mods {
		incoming.Modules[mod.ID] = mod
	}
	return inv.MergeTransformResult(incoming)
}

// TestMergeReconcilesProviderModuleBays verifies a provider transform result
// whose modules name bays in their own spelling merges without aborting: a
// declared bay in another case is stored as declared, an undeclared bay is
// cleared and reported, and a parent that declares no bays keeps the
// provider's string.
//
// Why it matters: HPCM synthesises bay names from tray and node numbers and
// the example provider passes a sheet's bay through verbatim; a model rule
// that only a library-aware provider can satisfy must be reconciled at the
// seam every provider passes through, not allowed to refuse a whole import.
// Inputs: against the fixture chassis, modules in "SLOT1" and "tray-1/node-2";
// against a bare device, a module in "anything".
// Outputs: no error; bays "slot1", "" and "anything"; one Unresolved line
// naming the undeclared bay, the declared bays and `update module --bay`.
// Data choice: the three cases are the three outcomes ResolveModuleBayName
// has, exercised through the merge rather than directly.
func TestMergeReconcilesProviderModuleBays(t *testing.T) {
	inv, chassis := chassisWithBays()
	bare := uuid.New()
	inv.Devices[bare] = &CaniDeviceType{ID: bare, Name: "bare"}
	cased := &CaniModuleType{ID: uuid.New(), Name: "nic-cased", ParentDevice: chassis.ID, ModuleBayName: "SLOT1"}
	hpcm := &CaniModuleType{ID: uuid.New(), Name: "node-2", ParentDevice: chassis.ID, ModuleBayName: "tray-1/node-2"}
	loose := &CaniModuleType{ID: uuid.New(), Name: "loose", ParentDevice: bare, ModuleBayName: "anything"}

	summary, err := mergeModules(inv, cased, hpcm, loose)

	if err != nil {
		t.Fatalf("merge aborted: %v", err)
	}
	got := []string{inv.Modules[cased.ID].ModuleBayName, inv.Modules[hpcm.ID].ModuleBayName, inv.Modules[loose.ID].ModuleBayName}
	if got[0] != "slot1" || got[1] != "" || got[2] != "anything" {
		t.Errorf("bays after merge = %q, want [slot1 \"\" anything]", got)
	}
	if len(summary.Unresolved) != 1 {
		t.Fatalf("unresolved = %v, want exactly one report", summary.Unresolved)
	}
	for _, want := range []string{`module "node-2"`, `"tray-1/node-2"`, "slot1, slot2", "update module --bay"} {
		if !strings.Contains(summary.Unresolved[0], want) {
			t.Errorf("unresolved report missing %q: %s", want, summary.Unresolved[0])
		}
	}
}

// TestMergeStillRefusesContestedBaysAfterReconciliation verifies that
// canonicalising spellings does not weaken the double-claim rule: two
// incoming modules authored into one bay in different cases are refused.
//
// Why it matters: reconciliation exists to stop naming differences from
// aborting imports; it must not let two modules occupy one bay by spelling
// it differently.
// Inputs: incoming modules in "slot1" and "SLOT1" on the fixture chassis.
// Outputs: an error naming the occupied bay, and the inventory unchanged.
// Data choice: the pair differs only in case so reconciliation is the only
// thing that can make them collide.
func TestMergeStillRefusesContestedBaysAfterReconciliation(t *testing.T) {
	inv, chassis := chassisWithBays()
	a := &CaniModuleType{ID: uuid.New(), Name: "nic-a", ParentDevice: chassis.ID, ModuleBayName: "slot1"}
	b := &CaniModuleType{ID: uuid.New(), Name: "nic-b", ParentDevice: chassis.ID, ModuleBayName: "SLOT1"}

	_, err := mergeModules(inv, a, b)

	if err == nil || !strings.Contains(err.Error(), `module bay "slot1" on "chassis" is already occupied`) {
		t.Errorf("merge error = %v, want a contested-bay refusal", err)
	}
	if len(inv.Modules) != 0 {
		t.Errorf("modules after refused merge = %d, want 0", len(inv.Modules))
	}
}

// TestMergeTakesProviderModulePlacement verifies a re-import places an
// existing module where the provider now reports it, taking the parent, bay
// and location as one unit, while a placement the provider does not report is
// kept.
//
// Why it matters: a module moved, installed or stored as a spare in the
// provider must not stay where the datastore last saw it, and a provider that
// does not know the library's bay names must not clear an authored bay.
// Inputs: an existing module matched by its Nautobot ID, merged with an
// incoming record that moves it to "SLOT2", names no bay on the same parent,
// stores it at the site, installs it from the site, moves it to a device that
// declares no bays, or carries no placement. Outputs: the parent, bay and
// location after the merge, the authored name kept, and a reported asset tag
// taken.
// Data choice: "SLOT2" differs from the declared "slot2" only in case, which
// proves the moved bay still passes through bay reconciliation.
func TestMergeTakesProviderModulePlacement(t *testing.T) {
	type placement struct {
		parent   uuid.UUID
		bay      string
		location uuid.UUID
	}
	inv, chassis := chassisWithBays()
	site := siteOf(inv, chassis)
	bare := uuid.New()
	inv.Devices[bare] = &CaniDeviceType{ID: bare, Name: "bare"}
	installed := placement{chassis.ID, "slot1", uuid.Nil}
	stored := placement{uuid.Nil, "", site}
	cases := []struct {
		name     string
		existing placement
		incoming placement
		want     placement
	}{
		{"moved to another declared bay", installed, placement{chassis.ID, "SLOT2", uuid.Nil}, placement{chassis.ID, "slot2", uuid.Nil}},
		{"bay the provider does not name", installed, placement{chassis.ID, "", uuid.Nil}, installed},
		{"stored as a spare", installed, stored, stored},
		{"installed from stock", stored, placement{chassis.ID, "slot2", uuid.Nil}, placement{chassis.ID, "slot2", uuid.Nil}},
		{"moved to a device without declared bays", installed, placement{bare, "", uuid.Nil}, placement{bare, "", uuid.Nil}},
		{"no placement reported", installed, placement{}, installed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nautobotID := uuid.New()
			existing := &CaniModuleType{ID: uuid.New(), Name: "nic-a", AssetTag: "AT-1",
				ParentDevice: tc.existing.parent, ModuleBayName: tc.existing.bay, Location: tc.existing.location,
				ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": nautobotID}}}
			inv.Modules = map[uuid.UUID]*CaniModuleType{existing.ID: existing}
			incoming := &CaniModuleType{ID: uuid.New(), Name: "NIC installed in chassis", AssetTag: "AT-2",
				ParentDevice: tc.incoming.parent, ModuleBayName: tc.incoming.bay, Location: tc.incoming.location,
				ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": nautobotID}}}

			if _, err := mergeModules(inv, incoming); err != nil {
				t.Fatalf("merge: %v", err)
			}

			got := inv.Modules[existing.ID]
			if got == nil || len(inv.Modules) != 1 {
				t.Fatalf("modules after merge = %d, want the existing module only", len(inv.Modules))
			}
			if have := (placement{got.ParentDevice, got.ModuleBayName, got.Location}); have != tc.want {
				t.Errorf("placement = %+v, want %+v", have, tc.want)
			}
			if got.Name != "nic-a" || got.AssetTag != "AT-2" {
				t.Errorf("name, asset tag = %q, %q; want the authored name and the reported tag", got.Name, got.AssetTag)
			}
		})
	}
}
