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
