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

// sameNamedVRFs seeds a VRF named blue in each given namespace and returns
// their IDs in that order.
func sameNamedVRFs(inventory *devicetypes.Inventory, namespaces ...string) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(namespaces))
	for _, namespace := range namespaces {
		id := uuid.New()
		inventory.VRFs[id] = &devicetypes.CaniVRF{ID: id, Name: "blue", Namespace: namespace}
		ids = append(ids, id)
	}
	return ids
}

// TestUpdateVRFResolvesTheNamespaceQualifiedName verifies "namespace/name"
// picks the VRF in that namespace and leaves its Global namesake alone.
//
// Why it matters: a bare name used to match any namespace's VRF, so a tenant
// VRF could only be told apart from the Global one by UUID.
// Inputs: VRFs named blue in Global and tenant-a; "tenant-a/blue --rd 65000:10".
// Outputs: one save, the tenant-a RD set and the Global RD still empty.
// Data choice: identical names leave only the namespace to choose by.
func TestUpdateVRFResolvesTheNamespaceQualifiedName(t *testing.T) {
	// Arrange.
	inventory := devicetypes.NewInventory()
	ids := sameNamedVRFs(inventory, "Global", "tenant-a")
	harness := cmdtest.New(t, NewCommand(), newVRFCommand(), inventory)

	// Act.
	err := harness.Run(t, map[string][]string{"rd": {"65000:10"}}, "tenant-a/blue")

	// Assert.
	if err != nil || harness.Store.Saves != 1 {
		t.Fatalf("update vrf: error = %v, saves = %d; want nil, 1", err, harness.Store.Saves)
	}
	got := harness.Inventory()
	if got.VRFs[ids[1]].RD != "65000:10" || got.VRFs[ids[0]].RD != "" {
		t.Errorf("RDs = {tenant-a %q, Global %q}; want 65000:10 and empty", got.VRFs[ids[1]].RD, got.VRFs[ids[0]].RD)
	}
}

// TestUpdateVRFSearchesTheNamespaceFlag verifies a bare name is searched in
// --namespace rather than in Global.
//
// Why it matters: --namespace is the form scripts use when the name cannot be
// rewritten; it must scope the lookup the way the qualified form does.
// Inputs: VRFs named blue in Global and tenant-a; "blue --namespace tenant-a
// --description Tenant".
// Outputs: the tenant-a description set and the Global one untouched.
// Data choice: as above, only the namespace distinguishes the two.
func TestUpdateVRFSearchesTheNamespaceFlag(t *testing.T) {
	// Arrange.
	inventory := devicetypes.NewInventory()
	ids := sameNamedVRFs(inventory, "Global", "tenant-a")
	harness := cmdtest.New(t, NewCommand(), newVRFCommand(), inventory)

	// Act.
	err := harness.Run(t, map[string][]string{flagNamespace: {"tenant-a"}, flagDescription: {"Tenant"}}, "blue")

	// Assert.
	if err != nil {
		t.Fatalf("update vrf --namespace tenant-a: %v", err)
	}
	got := harness.Inventory()
	if got.VRFs[ids[1]].Description != "Tenant" || got.VRFs[ids[0]].Description != "" {
		t.Errorf("descriptions = {tenant-a %q, Global %q}; want Tenant and empty",
			got.VRFs[ids[1]].Description, got.VRFs[ids[0]].Description)
	}
}

// TestUpdateVRFReportsAnAmbiguousName verifies a name held by two VRFs in the
// namespace is refused with both UUIDs and nothing is saved.
//
// Why it matters: the old lookup returned the first match, so an update could
// land on the wrong legacy duplicate without a word.
// Inputs: two Global VRFs named blue; "blue --description x".
// Outputs: an error naming both UUIDs and no save.
// Data choice: exact duplicates cannot be told apart by any other field.
func TestUpdateVRFReportsAnAmbiguousName(t *testing.T) {
	// Arrange.
	inventory := devicetypes.NewInventory()
	ids := sameNamedVRFs(inventory, "Global", "Global")
	harness := cmdtest.New(t, NewCommand(), newVRFCommand(), inventory)

	// Act.
	err := harness.Run(t, map[string][]string{flagDescription: {"x"}}, "blue")

	// Assert.
	if err == nil || !strings.Contains(err.Error(), "ambiguous") || harness.Store.Saves != 0 {
		t.Fatalf("error = %v, saves = %d; want an ambiguity error and no save", err, harness.Store.Saves)
	}
	for _, id := range ids {
		if !strings.Contains(err.Error(), id.String()) {
			t.Errorf("error %v does not name candidate %s", err, id)
		}
	}
}

// TestUpdateVRFAddsAndRemovesDevicesThroughTheModel verifies --add-device
// assigns a device once and --remove-device drops it, each counted as a change.
//
// Why it matters: device assignments are model state; the command must report
// a no-op as "no changes" rather than rewrite the datastore.
// Inputs: a rack with device cn-01 and a Global VRF blue; add cn-01 twice in
// one call, then remove it.
// Outputs: one device after the add, none after the remove, one save each.
// Data choice: repeating the device in one call proves the model dedupes.
func TestUpdateVRFAddsAndRemovesDevicesThroughTheModel(t *testing.T) {
	// Arrange.
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	device := cmdtest.AddDevice(inventory, rackID, "cn-01", 10)
	vrf := sameNamedVRFs(inventory, "Global")[0]
	harness := cmdtest.New(t, NewCommand(), newVRFCommand(), inventory)

	// Act.
	added := harness.Run(t, map[string][]string{flagAddDevice: {"cn-01", "cn-01"}}, "blue")
	afterAdd := append([]uuid.UUID(nil), harness.Inventory().VRFs[vrf].Devices...)
	removed := harness.Run(t, map[string][]string{flagRemoveDevice: {"cn-01"}}, "blue")

	// Assert.
	if added != nil || removed != nil || harness.Store.Saves != 2 {
		t.Fatalf("add error = %v, remove error = %v, saves = %d; want nil, nil, 2", added, removed, harness.Store.Saves)
	}
	if !reflect.DeepEqual(afterAdd, []uuid.UUID{device}) || len(harness.Inventory().VRFs[vrf].Devices) != 0 {
		t.Errorf("devices after add = %v, after remove = %v; want [%v] then none", afterAdd, harness.Inventory().VRFs[vrf].Devices, device)
	}
}

// TestUpdateVRFRemovingAnUnassignedDeviceIsNoChange verifies --remove-device
// for a device the VRF does not hold fails as "no changes specified".
//
// Why it matters: a silent success would hide a mistyped device name.
// Inputs: device cn-01 that is not in VRF blue; "--remove-device cn-01".
// Outputs: "no changes specified" and no save.
// Data choice: the device exists, so only the missing assignment can fail.
func TestUpdateVRFRemovingAnUnassignedDeviceIsNoChange(t *testing.T) {
	// Arrange.
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	cmdtest.AddDevice(inventory, rackID, "cn-01", 10)
	sameNamedVRFs(inventory, "Global")
	harness := cmdtest.New(t, NewCommand(), newVRFCommand(), inventory)

	// Act.
	err := harness.Run(t, map[string][]string{flagRemoveDevice: {"cn-01"}}, "blue")

	// Assert.
	if err == nil || !strings.Contains(err.Error(), "no changes specified") || harness.Store.Saves != 0 {
		t.Errorf("error = %v, saves = %d; want no changes specified and no save", err, harness.Store.Saves)
	}
}
