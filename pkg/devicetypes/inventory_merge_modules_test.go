package devicetypes

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// TestMergeModules verifies MergeModules resolves incoming modules by UUID
// match, then name match, then insert, while skipping nil and invalid
// modules; an unnamed module without identity is a module like any other and
// is inserted.
//
// Why it matters: module re-import must converge without duplicates and must
// reject modules whose slug is absent from the library, since those cannot be
// rendered downstream; a module needs no name, as in Nautobot.
// Inputs: an inventory with a "nic-0" module, then a merge containing a
// UUID-matched update, a name-matched "nic-0", a new "nic-1", a nil, an unnamed
// module, and one with an unregistered slug. Outputs: the UUID and name matches
// folded in, "nic-1" and the unnamed module inserted, and the two invalid
// entries skipped.
// Data choice: the unregistered-slug entry drives the Validate()-failure skip
// branch that a simpler test would miss.
func TestMergeModules(t *testing.T) {
	inv := NewInventory()
	devID := uuid.New()
	idA := uuid.New()
	inv.Modules[idA] = &CaniModuleType{ID: idA, Name: "nic-0", ParentDevice: devID, ObjectMeta: ObjectMeta{Status: "old"}}

	inv.MergeModules(map[uuid.UUID]*CaniModuleType{
		idA:        {ID: idA, Name: "nic-0", ObjectMeta: ObjectMeta{Status: "updated"}}, // UUID match
		uuid.New(): {Name: "nic-0", Serial: "SN-name"},                                  // name match
		uuid.New(): {Name: "nic-1"},                                                     // insert
		uuid.New(): nil,                                                                 // skipped (nil)
		uuid.New(): {Name: ""},                                                          // inserted (unnamed)
		uuid.New(): {Name: "bad", Slug: "not-in-library"},                               // skipped (invalid)
	})

	if inv.Modules[idA].Status != "updated" {
		t.Errorf("UUID-matched module Status = %q, want %q", inv.Modules[idA].Status, "updated")
	}
	if inv.Modules[idA].Serial != "SN-name" {
		t.Errorf("name-matched merge did not apply, Serial = %q, want %q", inv.Modules[idA].Serial, "SN-name")
	}
	var hasNic1 bool
	for _, m := range inv.Modules {
		if m != nil && m.Name == "nic-1" {
			hasNic1 = true
		}
	}
	if !hasNic1 {
		t.Error("expected nic-1 to be inserted")
	}
	if len(inv.Modules) != 3 {
		t.Errorf("final module count = %d, want 3 (nic-0 + nic-1 + the unnamed module)", len(inv.Modules))
	}
}

// TestMergeModulesMatchesByIdentity verifies MergeModules folds an incoming
// module into the existing one that shares its identity, Nautobot's rule:
// the same serial on the same hardware type, or the same asset tag; and that
// neither a shared serial on another type nor two blank names is a match.
//
// Why it matters: spares are authored without names, so a re-import or a
// second authoring of the same card must converge on identity, and two
// unnamed cards must never be folded into one because their names agree.
// Inputs: an unnamed spare of type "nic" with serial SN-1 and asset tag AT-1;
// merges of: same type and serial, same asset tag without a serial, same
// serial on another type, and two unnamed modules with no identity.
// Outputs: the first two fold in (status updated), the last three are inserted.
// Data choice: the asset-tag case carries no serial, proving either identity
// value alone is enough; the other-type case proves a serial is scoped to
// its hardware type, as Nautobot's uniqueness rule is.
func TestMergeModulesMatchesByIdentity(t *testing.T) {
	inv := NewInventory()
	spare := uuid.New()
	inv.Modules[spare] = &CaniModuleType{ID: spare, Model: "nic", Serial: "SN-1", AssetTag: "AT-1", ObjectMeta: ObjectMeta{Status: "old"}}

	inv.MergeModules(map[uuid.UUID]*CaniModuleType{
		uuid.New(): {Model: "nic", Serial: "sn-1", ObjectMeta: ObjectMeta{Status: "by-serial"}},
	})
	if got := inv.Modules[spare].Status; got != "by-serial" || len(inv.Modules) != 1 {
		t.Errorf("same type and serial: status %q, %d module(s); want folded in", got, len(inv.Modules))
	}
	inv.MergeModules(map[uuid.UUID]*CaniModuleType{
		uuid.New(): {Model: "nic", AssetTag: "at-1", ObjectMeta: ObjectMeta{Status: "by-asset-tag"}},
	})
	if got := inv.Modules[spare].Status; got != "by-asset-tag" || len(inv.Modules) != 1 {
		t.Errorf("same asset tag: status %q, %d module(s); want folded in", got, len(inv.Modules))
	}
	inv.MergeModules(map[uuid.UUID]*CaniModuleType{
		uuid.New(): {Model: "hba", Serial: "SN-1"},
		uuid.New(): {},
	})
	if len(inv.Modules) != 3 {
		t.Errorf("serial on another type and an identity-less module: %d module(s), want 3 (both inserted)", len(inv.Modules))
	}
	inv.MergeModules(map[uuid.UUID]*CaniModuleType{uuid.New(): {}})
	if len(inv.Modules) != 4 {
		t.Errorf("second identity-less module: %d module(s), want 4; blank names must not match", len(inv.Modules))
	}
}

// TestMergeModulesMatchesSerialWhenOneTypeIsUnknown verifies a module that
// records no hardware type folds into the existing module of its serial,
// while a module of another recorded type with the same serial does not.
//
// Why it matters: Nautobot's module type reference carries no model, so an
// imported module has no type key; requiring equal keys made the import of a
// card authored in cani a duplicate rather than the same card, and
// contradicted the swap rule, which compares types only when both record one.
// Inputs: an authored spare of model "nic" with serial SN-1; merges of a
// typeless module with serial sn-1 and of an "hba" with serial SN-1.
// Outputs: the first folds in (status updated, one module), the second is
// inserted (two modules).
// Data choice: the lower-case serial proves the comparison stays
// case-insensitive on the path that no longer compares types.
func TestMergeModulesMatchesSerialWhenOneTypeIsUnknown(t *testing.T) {
	inv := NewInventory()
	spare := uuid.New()
	inv.Modules[spare] = &CaniModuleType{ID: spare, Model: "nic", Serial: "SN-1", ObjectMeta: ObjectMeta{Status: "old"}}

	if _, err := inv.MergeModules(map[uuid.UUID]*CaniModuleType{
		uuid.New(): {Serial: "sn-1", ObjectMeta: ObjectMeta{Status: "imported"}},
	}); err != nil {
		t.Fatalf("typeless merge: %v", err)
	}
	if got := inv.Modules[spare].Status; got != "imported" || len(inv.Modules) != 1 {
		t.Errorf("typeless module with the serial: status %q, %d module(s); want folded in", got, len(inv.Modules))
	}
	if _, err := inv.MergeModules(map[uuid.UUID]*CaniModuleType{uuid.New(): {Model: "hba", Serial: "SN-1"}}); err != nil {
		t.Fatalf("other-type merge: %v", err)
	}
	if len(inv.Modules) != 2 {
		t.Errorf("other type with the serial: %d module(s), want 2 (inserted)", len(inv.Modules))
	}
}

// TestMergeModulesRefusesAmbiguousIdentity verifies an incoming module whose
// serial names one existing module and whose asset tag names another is
// refused with an error naming all three, and that nothing is merged or
// inserted, through MergeModules and through MergeTransformResult.
//
// Why it matters: folding the module into either record would copy both
// identifiers onto it and leave two records claiming one identity, with the
// winner chosen by map iteration order; the merge must refuse what it cannot
// decide, as the IPAM merges do.
// Inputs: existing A (serial SN-1) and B (asset tag AT-9), both "nic"; an
// incoming "nic" with SN-1 and AT-9, merged directly and as a transform
// result. Outputs: an error naming A, B and the incoming label each time;
// A and B unchanged; still two modules.
// Data choice: each existing module carries only one identifier so that only
// the incoming module ties them together.
func TestMergeModulesRefusesAmbiguousIdentity(t *testing.T) {
	inv := NewInventory()
	a, b := uuid.New(), uuid.New()
	inv.Modules[a] = &CaniModuleType{ID: a, Model: "nic", Serial: "SN-1"}
	inv.Modules[b] = &CaniModuleType{ID: b, Model: "nic", AssetTag: "AT-9"}
	incoming := &CaniModuleType{ID: uuid.New(), Model: "nic", Serial: "SN-1", AssetTag: "AT-9", ObjectMeta: ObjectMeta{Status: "new"}}

	_, err := inv.MergeModules(map[uuid.UUID]*CaniModuleType{incoming.ID: incoming})
	for _, want := range []string{"2 existing modules", a.String(), b.String(), incoming.Label()} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("MergeModules error = %v, want it to name %q", err, want)
		}
	}
	if _, err := mergeModules(inv, incoming); err == nil || !strings.Contains(err.Error(), "merge modules:") {
		t.Errorf("MergeTransformResult error = %v, want the module merge refusal", err)
	}
	if len(inv.Modules) != 2 || inv.Modules[a].AssetTag != "" || inv.Modules[b].Serial != "" ||
		inv.Modules[a].Status != "" || inv.Modules[b].Status != "" {
		t.Errorf("modules after refusal = %d, A = {%q %q %q}, B = {%q %q %q}; want both untouched and nothing inserted",
			len(inv.Modules), inv.Modules[a].Serial, inv.Modules[a].AssetTag, inv.Modules[a].Status,
			inv.Modules[b].Serial, inv.Modules[b].AssetTag, inv.Modules[b].Status)
	}
}
