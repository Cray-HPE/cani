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
package datastores

import (
	"path/filepath"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// TestJSONStoreRoundTripsNamespaceScope verifies Save then Load preserves a
// prefix's namespace and VRF memberships and an address's intended namespace,
// and that the reloaded inventory validates without conflicts.
//
// Why it matters: these fields are what makes two namespaces distinct on
// disk; a wrong tag or an omitempty slip would silently fold a tenant back
// into Global on the next load.
// Inputs: a tenant-a VRF, a tenant-a 10.0.0.0/24 that belongs to it, and
// 10.0.0.5/24 added with tenant-a intent. Outputs: the same namespace,
// membership, and parent after Load, no relationship errors and no
// unresolved conflicts.
// Data choice: a non-Global namespace is the value a defaulting bug would
// erase; the address is added without a parent so its scope must be derived.
func TestJSONStoreRoundTripsNamespaceScope(t *testing.T) {
	// Arrange.
	path := filepath.Join(t.TempDir(), "inventory.json")
	inv := devicetypes.NewInventory()
	vrf := &devicetypes.CaniVRF{ID: uuid.New(), Name: "blue", Namespace: "tenant-a"}
	prefix := &devicetypes.CaniPrefix{ID: uuid.New(), Prefix: "10.0.0.0/24", Namespace: "tenant-a", VRFs: []uuid.UUID{vrf.ID}}
	addr := &devicetypes.CaniIPAddress{ID: uuid.New(), Address: "10.0.0.5/24", Namespace: "tenant-a"}
	for _, err := range []error{inv.AddVRF(vrf), inv.AddPrefix(prefix), inv.AddIPAddress(addr)} {
		if err != nil {
			t.Fatalf("building inventory: %v", err)
		}
	}
	store := &JSONStore{Path: path}

	// Act.
	if err := store.Save(inv); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := store.Load()

	// Assert.
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	assertScopePreserved(t, loaded, vrf.ID, prefix.ID, addr.ID)
}

// assertScopePreserved checks the reloaded VRF, prefix and address still carry
// the tenant-a scope they were saved with, and that the inventory validates.
func assertScopePreserved(t *testing.T, loaded *devicetypes.Inventory, vrfID, prefixID, addrID uuid.UUID) {
	t.Helper()
	if got := loaded.VRFs[vrfID]; got == nil || got.Namespace != "tenant-a" {
		t.Errorf("loaded VRF = %+v, want namespace tenant-a", got)
	}
	prefix := loaded.Prefixes[prefixID]
	if prefix == nil || prefix.Namespace != "tenant-a" || len(prefix.VRFs) != 1 || prefix.VRFs[0] != vrfID || prefix.VRF != "" {
		t.Errorf("loaded prefix = %+v, want namespace tenant-a, VRFs [%v], no legacy name", prefix, vrfID)
	}
	addr := loaded.IPAddresses[addrID]
	if addr == nil || addr.Namespace != "tenant-a" || addr.Parent != prefixID || loaded.IPAddressNamespace(addr) != "tenant-a" {
		t.Errorf("loaded address = %+v, want namespace tenant-a parented to %v", addr, prefixID)
	}
	if rel := loaded.RebuildDerivedState(); len(rel.Errors) != 0 || len(rel.Unresolved) != 0 {
		t.Errorf("reloaded inventory: errors = %v, unresolved = %v; want none", rel.Errors, rel.Unresolved)
	}
}
