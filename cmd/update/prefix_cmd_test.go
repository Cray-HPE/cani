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
	"reflect"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// legacyPrefixInventory returns a tenant-a 10.0.0.0/24 whose legacy VRF name
// is legacy, and tenant-a VRFs with the given names.
func legacyPrefixInventory(legacy string, vrfNames ...string) (*devicetypes.Inventory, uuid.UUID, []uuid.UUID) {
	inventory := devicetypes.NewInventory()
	prefix := uuid.New()
	inventory.Prefixes[prefix] = &devicetypes.CaniPrefix{ID: prefix, Prefix: "10.0.0.0/24", Namespace: "tenant-a", VRF: legacy}
	vrfs := make([]uuid.UUID, 0, len(vrfNames))
	for _, name := range vrfNames {
		id := uuid.New()
		inventory.VRFs[id] = &devicetypes.CaniVRF{ID: id, Name: name, Namespace: "tenant-a"}
		vrfs = append(vrfs, id)
	}
	return inventory, prefix, vrfs
}

// TestUpdatePrefixAddsAVRFAndClearsTheLegacyName verifies --vrf adds the
// membership and clears the legacy VRF name it replaces.
//
// Why it matters: the legacy name is reported on every change and fails each
// export until memberships are set, and there was no command to set them.
// Inputs: tenant-a 10.0.0.0/24 with legacy name red and two tenant-a VRFs
// named red; "tenant-a/10.0.0.0/24 --vrf <first red's UUID>".
// Outputs: one save, the first red as the only membership, no legacy name.
// Data choice: two VRFs named red are why the legacy name never migrated, so
// only the UUID can choose.
func TestUpdatePrefixAddsAVRFAndClearsTheLegacyName(t *testing.T) {
	// Arrange.
	inventory, prefix, vrfs := legacyPrefixInventory("red", "red", "red")
	harness := cmdtest.New(t, NewCommand(), newPrefixCommand(), inventory)

	// Act.
	err := harness.Run(t, map[string][]string{flagVRF: {vrfs[0].String()}}, "tenant-a/10.0.0.0/24")

	// Assert.
	if err != nil || harness.Store.Saves != 1 {
		t.Fatalf("update prefix: error = %v, saves = %d; want nil, 1", err, harness.Store.Saves)
	}
	got := harness.Inventory().Prefixes[prefix]
	if !reflect.DeepEqual(got.VRFs, []uuid.UUID{vrfs[0]}) || got.VRF != "" {
		t.Errorf("VRFs = %v, legacy name = %q; want [%v] and none", got.VRFs, got.VRF, vrfs[0])
	}
}

// TestUpdatePrefixRemoveVRFClearsALegacyName verifies --remove-vrf removes a
// membership by name and clears a legacy name that matches no VRF.
//
// Why it matters: a stale legacy name has no VRF to resolve to, so naming it
// is the only way to clear it.
// Inputs: tenant-a 10.0.0.0/24 in VRF blue with legacy name stale;
// "--remove-vrf stale --remove-vrf blue".
// Outputs: one save, no memberships and no legacy name.
// Data choice: stale exists nowhere, so resolving it as a VRF would fail.
func TestUpdatePrefixRemoveVRFClearsALegacyName(t *testing.T) {
	// Arrange.
	inventory, prefix, vrfs := legacyPrefixInventory("stale", "blue")
	inventory.Prefixes[prefix].VRFs = []uuid.UUID{vrfs[0]}
	harness := cmdtest.New(t, NewCommand(), newPrefixCommand(), inventory)

	// Act.
	err := harness.Run(t, map[string][]string{flagRemoveVRF: {"stale", "blue"}, flagNamespace: {"tenant-a"}}, "10.0.0.0/24")

	// Assert.
	if err != nil || harness.Store.Saves != 1 {
		t.Fatalf("update prefix: error = %v, saves = %d; want nil, 1", err, harness.Store.Saves)
	}
	if got := harness.Inventory().Prefixes[prefix]; len(got.VRFs) != 0 || got.VRF != "" {
		t.Errorf("VRFs = %v, legacy name = %q; want none", got.VRFs, got.VRF)
	}
}

// TestUpdatePrefixRemoveVRFDropsAMembershipWhoseVRFIsGone verifies
// --remove-vrf <uuid> drops a membership even when no VRF has that UUID any
// more, so the prefix can be saved again.
//
// Why it matters: a dangling membership fails relationship validation, which
// refuses every save of the inventory; if the only remedy first resolved the
// UUID to a VRF, the operator could never clear it from the CLI.
// Inputs: tenant-a 10.0.0.0/24 whose only membership names a UUID with no
// VRF; "--remove-vrf <that uuid>".
// Outputs: one save and no memberships.
// Data choice: no VRF exists at all, so a lookup of the UUID must fail and
// only the model's remove-by-ID path can succeed.
func TestUpdatePrefixRemoveVRFDropsAMembershipWhoseVRFIsGone(t *testing.T) {
	// Arrange.
	inventory, prefix, _ := legacyPrefixInventory("")
	gone := uuid.New()
	inventory.Prefixes[prefix].VRFs = []uuid.UUID{gone}
	harness := cmdtest.New(t, NewCommand(), newPrefixCommand(), inventory)

	// Act.
	err := harness.Run(t, map[string][]string{flagRemoveVRF: {gone.String()}}, "tenant-a/10.0.0.0/24")

	// Assert.
	if err != nil || harness.Store.Saves != 1 {
		t.Fatalf("update prefix: error = %v, saves = %d; want nil, 1", err, harness.Store.Saves)
	}
	if got := harness.Inventory().Prefixes[prefix].VRFs; len(got) != 0 {
		t.Errorf("VRFs = %v, want none", got)
	}
}

// TestUpdatePrefixRejectsAVRFFromAnotherNamespace verifies a VRF outside the
// prefix's namespace is refused and nothing is saved.
//
// Why it matters: Nautobot rejects a VRF assignment across namespaces, so the
// membership could never be exported.
// Inputs: tenant-a 10.0.0.0/24 and a Global VRF blue given by UUID.
// Outputs: an error naming both namespaces and no save.
// Data choice: a UUID bypasses the namespace-scoped name lookup, leaving the
// model's check as the only guard.
func TestUpdatePrefixRejectsAVRFFromAnotherNamespace(t *testing.T) {
	// Arrange.
	inventory, _, _ := legacyPrefixInventory("")
	global := uuid.New()
	inventory.VRFs[global] = &devicetypes.CaniVRF{ID: global, Name: "blue"}
	harness := cmdtest.New(t, NewCommand(), newPrefixCommand(), inventory)

	// Act.
	err := harness.Run(t, map[string][]string{flagVRF: {global.String()}}, "tenant-a/10.0.0.0/24")

	// Assert.
	if err == nil || !strings.Contains(err.Error(), `is in namespace "Global", not "tenant-a"`) || harness.Store.Saves != 0 {
		t.Errorf("error = %v, saves = %d; want a namespace error and no save", err, harness.Store.Saves)
	}
}

// TestUpdatePrefixSetsTheDescription verifies --description alone is a change
// that is saved.
//
// Why it matters: the change check must count every flag, or a description
// update fails as "no changes specified".
// Inputs: tenant-a 10.0.0.0/24 and --description Management.
// Outputs: one save and the new description.
// Data choice: no VRF flags, so the description is the only change.
func TestUpdatePrefixSetsTheDescription(t *testing.T) {
	// Arrange.
	inventory, prefix, _ := legacyPrefixInventory("")
	harness := cmdtest.New(t, NewCommand(), newPrefixCommand(), inventory)

	// Act.
	err := harness.Run(t, map[string][]string{flagDescription: {"Management"}}, "tenant-a/10.0.0.0/24")

	// Assert.
	if err != nil || harness.Store.Saves != 1 || harness.Inventory().Prefixes[prefix].Description != "Management" {
		t.Errorf("error = %v, saves = %d, description = %q; want nil, 1, Management",
			err, harness.Store.Saves, harness.Inventory().Prefixes[prefix].Description)
	}
}

// TestUpdatePrefixWithoutChangesFails verifies a call that changes nothing
// fails rather than rewriting the datastore.
//
// Why it matters: the other update commands behave this way, and a silent
// success hides a mistyped flag.
// Inputs: tenant-a 10.0.0.0/24 and no flags.
// Outputs: "no changes specified" and no save.
// Data choice: the prefix resolves, so only the empty change set can fail.
func TestUpdatePrefixWithoutChangesFails(t *testing.T) {
	// Arrange.
	inventory, _, _ := legacyPrefixInventory("")
	harness := cmdtest.New(t, NewCommand(), newPrefixCommand(), inventory)

	// Act.
	err := harness.Run(t, nil, "tenant-a/10.0.0.0/24")

	// Assert.
	if err == nil || !strings.Contains(err.Error(), "no changes specified") || harness.Store.Saves != 0 {
		t.Errorf("error = %v, saves = %d; want no changes specified and no save", err, harness.Store.Saves)
	}
}
