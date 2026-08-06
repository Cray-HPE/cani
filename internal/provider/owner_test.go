/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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
package provider

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// TestForInventoryReturnsOnlyTheOwningProvider verifies that an inventory which
// records an owner resolves to that provider alone.
//
// Why it matters: generic CRUD offers optional hooks (MetadataApplier,
// RackPostAddHook, DeviceStager) to providers, and those hooks mutate the
// inventory. Running every registered provider's hooks files a user's metadata
// under a provider that is not in use and lets one provider's naming scheme
// overwrite objects owned by another.
// Inputs: two registered fakeProviders and an inventory whose Provider names
// the second. Outputs: a single-element slice holding that provider. Data
// choice: registering two providers proves the result is filtered rather than
// simply reflecting a registry that happens to hold one entry.
func TestForInventoryReturnsOnlyTheOwningProvider(t *testing.T) {
	for _, name := range []string{"fake-owner-a", "fake-owner-b"} {
		Register(name, fakeProvider{slug: name})
		t.Cleanup(func() { delete(providers, name) })
	}

	inv := devicetypes.NewInventory()
	inv.Provider = "fake-owner-b"

	got := ForInventory(inv)
	if len(got) != 1 {
		t.Fatalf("ForInventory() returned %d providers, want 1", len(got))
	}
	if got[0].Slug() != "fake-owner-b" {
		t.Errorf("ForInventory() = %q, want %q", got[0].Slug(), "fake-owner-b")
	}
}

// TestForInventoryFallsBackWhenOwnerUnrecorded verifies that an inventory with
// no recorded owner resolves to every registered provider.
//
// Why it matters: inventories written before the importer began stamping an
// owner, and inventories built purely with `cani add`, carry no provider name.
// Those must keep receiving provider hooks or existing workflows would silently
// stop staging devices and applying metadata.
// Inputs: a freshly constructed inventory with an empty Provider, plus one
// registered fakeProvider. Outputs: a slice containing that provider. Data
// choice: asserting the test provider is present rather than an exact length
// keeps the test independent of other registrations.
func TestForInventoryFallsBackWhenOwnerUnrecorded(t *testing.T) {
	const name = "fake-unowned"
	Register(name, fakeProvider{slug: name})
	t.Cleanup(func() { delete(providers, name) })

	var found bool
	for _, p := range ForInventory(devicetypes.NewInventory()) {
		if p.Slug() == name {
			found = true
		}
	}
	if !found {
		t.Errorf("ForInventory() on an unowned inventory omitted %q", name)
	}
}

// TestForInventoryReturnsNoneForUnregisteredOwner verifies that an inventory
// owned by a provider that is not compiled in resolves to no providers.
//
// Why it matters: the owner is recorded in the datastore, so it can name a
// provider that a later build no longer registers. Falling back to every
// provider there would let an unrelated provider mutate data it does not own,
// which is exactly the failure the ownership check exists to prevent.
// Inputs: an inventory naming a provider that was never registered. Outputs: an
// empty slice. Data choice: a name that no init() could plausibly claim.
func TestForInventoryReturnsNoneForUnregisteredOwner(t *testing.T) {
	inv := devicetypes.NewInventory()
	inv.Provider = "no-such-provider-xyz"

	if got := ForInventory(inv); len(got) != 0 {
		t.Errorf("ForInventory() returned %d providers, want 0", len(got))
	}
}
