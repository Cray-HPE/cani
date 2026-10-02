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

// ipamInventory seeds a Global and a tenant-a 10.0.0.0/16 plus a VRF named
// "blue" in each namespace, the shape every namespace flag test needs.
func ipamInventory() (*devicetypes.Inventory, map[string]uuid.UUID) {
	inventory := devicetypes.NewInventory()
	ids := map[string]uuid.UUID{}
	for _, namespace := range []string{"Global", "tenant-a"} {
		prefixID, vrfID := uuid.New(), uuid.New()
		ids[namespace+"/16"], ids[namespace+"/vrf"] = prefixID, vrfID
		inventory.Prefixes[prefixID] = &devicetypes.CaniPrefix{ID: prefixID, Prefix: "10.0.0.0/16", PrefixLen: 16, Namespace: namespace}
		inventory.VRFs[vrfID] = &devicetypes.CaniVRF{ID: vrfID, Name: "blue", Namespace: namespace}
	}
	return inventory, ids
}

// onlyNewPrefix returns the single prefix that was not seeded by ipamInventory.
func onlyNewPrefix(t *testing.T, inventory *devicetypes.Inventory, ids map[string]uuid.UUID) *devicetypes.CaniPrefix {
	t.Helper()
	for id, prefix := range inventory.Prefixes {
		if id != ids["Global/16"] && id != ids["tenant-a/16"] {
			return prefix
		}
	}
	t.Fatal("no new prefix was added")
	return nil
}

// TestAddPrefixScopesNamespaceParentAndVRF verifies --namespace scopes the
// new prefix, its parent is chosen inside that namespace, and --vrf resolves
// to a canonical membership in the same namespace.
//
// Why it matters: the command layer is the only place an operator can author
// scope; defaulting to Global or resolving the VRF name across namespaces
// would silently file the prefix in the wrong network.
// Inputs: a /16 and a VRF "blue" in both Global and tenant-a; add
// 10.0.1.0/24 --namespace tenant-a --vrf blue. Outputs: one save, a prefix in
// tenant-a parented to the tenant-a /16 with the tenant-a VRF as its only
// membership and no legacy VRF string.
// Data choice: identical prefix and VRF text in both namespaces means only
// the namespace flag can explain which objects were chosen.
func TestAddPrefixScopesNamespaceParentAndVRF(t *testing.T) {
	inventory, ids := ipamInventory()
	harness := cmdtest.New(t, NewCommand(), newPrefixCommand(), inventory)

	err := harness.Run(t, map[string][]string{
		"namespace": {"tenant-a"},
		"vrf":       {"blue"},
	}, "10.0.1.0/24")
	if err != nil {
		t.Fatalf("add prefix --namespace tenant-a --vrf blue: unexpected error: %v", err)
	}

	if harness.Store.Saves != 1 {
		t.Errorf("Saves = %d, want 1", harness.Store.Saves)
	}
	prefix := onlyNewPrefix(t, harness.Inventory(), ids)
	if prefix.Namespace != "tenant-a" || prefix.Parent != ids["tenant-a/16"] {
		t.Errorf("prefix = {ns %q, parent %v}, want tenant-a parented to %v", prefix.Namespace, prefix.Parent, ids["tenant-a/16"])
	}
	if len(prefix.VRFs) != 1 || prefix.VRFs[0] != ids["tenant-a/vrf"] || prefix.VRF != "" {
		t.Errorf("prefix VRFs = %v, legacy = %q; want [%v] and no legacy name", prefix.VRFs, prefix.VRF, ids["tenant-a/vrf"])
	}
}

// TestAddPrefixDefaultsToGlobal verifies an omitted --namespace means Global,
// so the parent is chosen among the Global prefixes.
//
// Why it matters: legacy invocations must keep their meaning when a second
// namespace holds the same CIDRs.
// Inputs: a /16 in Global and in tenant-a; add 10.0.2.0/24 with no flags.
// Outputs: one save and a prefix in Global parented to the Global /16.
// Data choice: the tenant-a /16 also contains the prefix, so only the default
// namespace can explain the parent chosen.
func TestAddPrefixDefaultsToGlobal(t *testing.T) {
	inventory, ids := ipamInventory()
	harness := cmdtest.New(t, NewCommand(), newPrefixCommand(), inventory)

	if err := harness.Run(t, nil, "10.0.2.0/24"); err != nil {
		t.Fatalf("add prefix without flags: unexpected error: %v", err)
	}

	if harness.Store.Saves != 1 {
		t.Errorf("Saves = %d, want 1", harness.Store.Saves)
	}
	prefix := onlyNewPrefix(t, harness.Inventory(), ids)
	if prefix.Namespace != devicetypes.DefaultNamespace || prefix.Parent != ids["Global/16"] {
		t.Errorf("prefix = {ns %q, parent %v}, want Global parented to %v", prefix.Namespace, prefix.Parent, ids["Global/16"])
	}
}

// TestAddIPScopesParentByNamespace verifies --namespace selects the parent
// prefix inside that namespace and records the intent on the address.
//
// Why it matters: the same host text exists in both networks; the address
// must land under the tenant's prefix, not the Global one.
// Inputs: a /16 in Global and in tenant-a; add ip 10.0.0.5/24 --namespace
// tenant-a. Outputs: an address parented to the tenant-a /16 with namespace
// tenant-a.
// Data choice: both /16s contain the host, so only the namespace can decide.
func TestAddIPScopesParentByNamespace(t *testing.T) {
	inventory, ids := ipamInventory()
	harness := cmdtest.New(t, NewCommand(), newIPCommand(), inventory)

	if err := harness.Run(t, map[string][]string{"namespace": {"tenant-a"}}, "10.0.0.5/24"); err != nil {
		t.Fatalf("add ip --namespace tenant-a: unexpected error: %v", err)
	}

	if len(harness.Inventory().IPAddresses) != 1 {
		t.Fatalf("addresses = %d, want 1", len(harness.Inventory().IPAddresses))
	}
	for _, addr := range harness.Inventory().IPAddresses {
		if addr.Namespace != "tenant-a" || addr.Parent != ids["tenant-a/16"] {
			t.Errorf("address = {ns %q, parent %v}, want tenant-a parented to %v", addr.Namespace, addr.Parent, ids["tenant-a/16"])
		}
	}
}

// TestAddVRFStoresGlobalByDefault verifies an omitted --namespace stores Global
// on the new VRF, as add prefix and add ip do.
//
// Why it matters: a blank stored namespace made add vrf the only IPAM writer
// whose saved scope differed from its effective one.
// Inputs: add vrf red with no flags, then add vrf red --namespace tenant-a.
// Outputs: two VRFs named red, one stored as Global and one as tenant-a.
// Data choice: reusing the name shows the flag, not the name, sets the scope.
func TestAddVRFStoresGlobalByDefault(t *testing.T) {
	harness := cmdtest.New(t, NewCommand(), newVRFCommand(), nil)

	if err := harness.Run(t, nil, "red"); err != nil {
		t.Fatalf("add vrf red: unexpected error: %v", err)
	}
	if err := harness.Run(t, map[string][]string{flagNamespace: {"tenant-a"}}, "red"); err != nil {
		t.Fatalf("add vrf red --namespace tenant-a: unexpected error: %v", err)
	}

	stored := map[string]int{}
	for _, vrf := range harness.Inventory().VRFs {
		stored[vrf.Namespace]++
	}
	if len(stored) != 2 || stored[devicetypes.DefaultNamespace] != 1 || stored["tenant-a"] != 1 {
		t.Errorf("stored VRF namespaces = %v, want one Global and one tenant-a", stored)
	}
}

// vrfInventory seeds one Global VRF named blue and returns its ID.
func vrfInventory() (*devicetypes.Inventory, uuid.UUID) {
	inventory := devicetypes.NewInventory()
	vrfID := uuid.New()
	inventory.VRFs[vrfID] = &devicetypes.CaniVRF{ID: vrfID, Name: "blue"}
	return inventory, vrfID
}

// onlyPrefix returns the single prefix in the inventory.
func onlyPrefix(t *testing.T, inventory *devicetypes.Inventory) *devicetypes.CaniPrefix {
	t.Helper()
	if len(inventory.Prefixes) != 1 {
		t.Fatalf("prefixes = %d, want 1", len(inventory.Prefixes))
	}
	for _, prefix := range inventory.Prefixes {
		return prefix
	}
	return nil
}

// TestAddPrefixResolvesVRFToMembership verifies --vrf stores a canonical VRF
// membership and no legacy VRF name.
//
// Why it matters: the model identifies a prefix's VRFs by UUID within its
// namespace and reports a legacy name as an unresolved conflict, so the
// command must never author one.
// Inputs: a Global VRF blue; add 10.0.0.0/24 --vrf blue. Outputs: one save,
// a prefix whose VRFs holds blue's UUID and whose VRF string is empty.
// Data choice: a bare name exercises the namespace-scoped name lookup rather
// than the UUID shortcut.
func TestAddPrefixResolvesVRFToMembership(t *testing.T) {
	// Arrange.
	inventory, vrfID := vrfInventory()
	harness := cmdtest.New(t, NewCommand(), newPrefixCommand(), inventory)

	// Act.
	err := harness.Run(t, map[string][]string{"vrf": {"blue"}}, "10.0.0.0/24")

	// Assert.
	if err != nil {
		t.Fatalf("add prefix --vrf blue: unexpected error: %v", err)
	}
	if harness.Store.Saves != 1 {
		t.Errorf("Saves = %d, want 1", harness.Store.Saves)
	}
	prefix := onlyPrefix(t, harness.Inventory())
	if len(prefix.VRFs) != 1 || prefix.VRFs[0] != vrfID || prefix.VRF != "" {
		t.Errorf("prefix VRFs = %v, legacy = %q; want [%v] and no legacy name", prefix.VRFs, prefix.VRF, vrfID)
	}
}

// TestAddPrefixRejectsUnknownVRF verifies an unresolvable --vrf reference
// fails the command before anything is saved.
//
// Why it matters: an unknown name must not be stored as intent to be guessed
// at later; the operator adds the VRF first.
// Inputs: a Global VRF blue; add 10.0.0.0/24 --vrf red. Outputs: an error
// naming the reference and zero saves.
// Data choice: red differs from blue in more than case, so no fallback match
// can rescue it.
func TestAddPrefixRejectsUnknownVRF(t *testing.T) {
	// Arrange.
	inventory, _ := vrfInventory()
	harness := cmdtest.New(t, NewCommand(), newPrefixCommand(), inventory)

	// Act.
	err := harness.Run(t, map[string][]string{"vrf": {"red"}}, "10.0.0.0/24")

	// Assert.
	if err == nil || !strings.Contains(err.Error(), `resolving --vrf "red"`) {
		t.Fatalf("add prefix --vrf red error = %v, want a resolution error naming the reference", err)
	}
	if harness.Store.Saves != 0 {
		t.Errorf("Saves = %d, want 0 (nothing should persist on failure)", harness.Store.Saves)
	}
}

// TestAddPrefixRejectsForeignVRF verifies a --vrf reference qualified into
// another namespace is refused by the model's membership check before
// anything is saved.
//
// Why it matters: the command only resolves the reference; the namespace
// boundary is the model's rule, so the refusal must reach the operator from
// AddPrefix and leave no prefix behind.
// Inputs: a tenant-a VRF blue; add 10.0.0.0/24 --vrf tenant-a/blue, so the
// prefix defaults to Global. Outputs: an error naming both namespaces and
// zero saves.
// Data choice: a qualified name is the only way to reach a VRF outside the
// prefix's namespace.
func TestAddPrefixRejectsForeignVRF(t *testing.T) {
	// Arrange.
	inventory := devicetypes.NewInventory()
	vrfID := uuid.New()
	inventory.VRFs[vrfID] = &devicetypes.CaniVRF{ID: vrfID, Name: "blue", Namespace: "tenant-a"}
	harness := cmdtest.New(t, NewCommand(), newPrefixCommand(), inventory)

	// Act.
	err := harness.Run(t, map[string][]string{"vrf": {"tenant-a/blue"}}, "10.0.0.0/24")

	// Assert.
	want := `VRF "blue" is in namespace "tenant-a", not "Global"`
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("add prefix --vrf tenant-a/blue error = %v, want it to contain %q", err, want)
	}
	if harness.Store.Saves != 0 {
		t.Errorf("Saves = %d, want 0 (nothing should persist on failure)", harness.Store.Saves)
	}
}
