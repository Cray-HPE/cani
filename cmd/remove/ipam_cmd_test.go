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
package remove

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// addPrefix stores a prefix directly, so tests control its parent.
func addPrefix(inventory *devicetypes.Inventory, cidr, namespace string, parent uuid.UUID) uuid.UUID {
	id := uuid.New()
	inventory.Prefixes[id] = &devicetypes.CaniPrefix{ID: id, Prefix: cidr, Namespace: namespace, Parent: parent}
	return id
}

// addAddress stores an address under a prefix.
func addAddress(inventory *devicetypes.Inventory, address string, parent uuid.UUID) uuid.UUID {
	id := uuid.New()
	host := strings.SplitN(address, "/", 2)[0]
	inventory.IPAddresses[id] = &devicetypes.CaniIPAddress{ID: id, Host: host, Address: address, Parent: parent}
	return id
}

// TestRemovePrefixByCIDRMovesItsChildrenUp verifies "remove prefix <cidr>"
// removes the Global prefix, hands its address to the parent prefix and
// leaves the same CIDR in another namespace alone.
//
// Why it matters: without the command, a duplicate or stale prefix could only
// be removed by editing the datastore.
// Inputs: Global 10.0.0.0/8 holding 10.0.0.0/16, which holds 10.0.0.5; and
// tenant-a 10.0.0.0/16.
// Outputs: one save; the Global /16 gone, the address under the /8, and the
// tenant-a /16 kept.
// Data choice: the tenant-a copy shows an unqualified CIDR is searched in
// Global only.
func TestRemovePrefixByCIDRMovesItsChildrenUp(t *testing.T) {
	// Arrange.
	inventory := devicetypes.NewInventory()
	slash8 := addPrefix(inventory, "10.0.0.0/8", "", uuid.Nil)
	global := addPrefix(inventory, "10.0.0.0/16", "", slash8)
	tenant := addPrefix(inventory, "10.0.0.0/16", "tenant-a", uuid.Nil)
	address := addAddress(inventory, "10.0.0.5/16", global)
	harness := cmdtest.New(t, NewCommand(), newPrefixCommand(), inventory)

	// Act.
	err := harness.Run(t, nil, "10.0.0.0/16")

	// Assert.
	if err != nil {
		t.Fatalf("remove prefix: %v", err)
	}
	got := harness.Inventory()
	if harness.Store.Saves != 1 || got.Prefixes[global] != nil || got.Prefixes[tenant] == nil {
		t.Fatalf("saves = %d, Global /16 kept = %v, tenant-a /16 kept = %v; want 1, false, true",
			harness.Store.Saves, got.Prefixes[global] != nil, got.Prefixes[tenant] != nil)
	}
	if parent := got.IPAddresses[address].Parent; parent != slash8 {
		t.Errorf("address parent = %v, want the /8 %v", parent, slash8)
	}
}

// TestRemovePrefixAcceptsANamespaceQualifiedCIDR verifies a CIDR written as
// "namespace/CIDR" removes the prefix in that namespace.
//
// Why it matters: the same CIDR exists once per namespace, and a qualified
// reference is how the operator picks one without looking up a UUID.
// Inputs: 10.0.0.0/24 in Global and in tenant-a; "tenant-a/10.0.0.0/24".
// Outputs: the tenant-a prefix is gone and the Global one is kept.
// Data choice: identical CIDRs leave only the namespace to choose by.
func TestRemovePrefixAcceptsANamespaceQualifiedCIDR(t *testing.T) {
	// Arrange.
	inventory := devicetypes.NewInventory()
	global := addPrefix(inventory, "10.0.0.0/24", "", uuid.Nil)
	tenant := addPrefix(inventory, "10.0.0.0/24", "tenant-a", uuid.Nil)
	harness := cmdtest.New(t, NewCommand(), newPrefixCommand(), inventory)

	// Act.
	err := harness.Run(t, nil, "tenant-a/10.0.0.0/24")

	// Assert.
	if err != nil {
		t.Fatalf("remove prefix: %v", err)
	}
	if got := harness.Inventory(); got.Prefixes[tenant] != nil || got.Prefixes[global] == nil {
		t.Errorf("tenant-a kept = %v, Global kept = %v; want false, true",
			got.Prefixes[tenant] != nil, got.Prefixes[global] != nil)
	}
}

// TestRemoveRootPrefixNeedsCascade verifies removing a root prefix that holds
// addresses fails, names --cascade and saves nothing, and that --cascade then
// removes the prefix and its addresses.
//
// Why it matters: removing the prefix alone would leave addresses that cannot
// be exported, and the error is the operator's only guide to the override.
// Inputs: a root Global 10.0.0.0/24 holding 10.0.0.5, removed without and
// then with --cascade.
// Outputs: first an error naming --cascade and no save; then one save with
// neither the prefix nor the address left.
// Data choice: a root prefix is the one case the model refuses.
func TestRemoveRootPrefixNeedsCascade(t *testing.T) {
	// Arrange.
	inventory := devicetypes.NewInventory()
	prefix := addPrefix(inventory, "10.0.0.0/24", "", uuid.Nil)
	address := addAddress(inventory, "10.0.0.5/24", prefix)
	harness := cmdtest.New(t, NewCommand(), newPrefixCommand(), inventory)

	// Act.
	refused := harness.Run(t, nil, "10.0.0.0/24")
	savesAfterRefusal := harness.Store.Saves
	cascaded := harness.Run(t, map[string][]string{flagCascade: {"true"}}, "10.0.0.0/24")

	// Assert.
	if refused == nil || !strings.Contains(refused.Error(), "rerun with --cascade") || savesAfterRefusal != 0 {
		t.Fatalf("without --cascade: error = %v, saves = %d; want an error naming --cascade and no save", refused, savesAfterRefusal)
	}
	if cascaded != nil {
		t.Fatalf("with --cascade: %v", cascaded)
	}
	if got := harness.Inventory(); got.Prefixes[prefix] != nil || got.IPAddresses[address] != nil {
		t.Errorf("prefix kept = %v, address kept = %v; want both removed", got.Prefixes[prefix] != nil, got.IPAddresses[address] != nil)
	}
}

// TestRemovePrefixListsDuplicatesByUUID verifies a CIDR held by two legacy
// records is refused with both UUIDs, and that removing one by UUID works.
//
// Why it matters: duplicates are the conflict the operator is told to
// resolve, and picking either record silently could remove the wrong one.
// Inputs: two Global 10.0.0.0/24 records; the CIDR, then the first UUID.
// Outputs: an error listing both UUIDs and no save; then one record left.
// Data choice: an exact duplicate cannot be told apart by any other field.
func TestRemovePrefixListsDuplicatesByUUID(t *testing.T) {
	// Arrange.
	inventory := devicetypes.NewInventory()
	first := addPrefix(inventory, "10.0.0.0/24", "", uuid.Nil)
	second := addPrefix(inventory, "10.0.0.0/24", "", uuid.Nil)
	harness := cmdtest.New(t, NewCommand(), newPrefixCommand(), inventory)

	// Act.
	ambiguous := harness.Run(t, nil, "10.0.0.0/24")
	byUUID := harness.Run(t, nil, first.String())

	// Assert.
	if ambiguous == nil || !strings.Contains(ambiguous.Error(), first.String()) || !strings.Contains(ambiguous.Error(), second.String()) {
		t.Errorf("by CIDR: error = %v, want both UUIDs", ambiguous)
	}
	if byUUID != nil || harness.Store.Saves != 1 {
		t.Fatalf("by UUID: error = %v, saves = %d; want nil, 1", byUUID, harness.Store.Saves)
	}
	if got := harness.Inventory(); got.Prefixes[first] != nil || got.Prefixes[second] == nil {
		t.Errorf("first kept = %v, second kept = %v; want false, true", got.Prefixes[first] != nil, got.Prefixes[second] != nil)
	}
}

// TestRemoveIPClearsThePrimaryIP verifies "remove ip" accepts the address as
// it was added, mask included, and clears the device primary IP that used it.
//
// Why it matters: a dangling primary IP fails validation, so the save would
// be refused and the address could not be removed at all.
// Inputs: a device whose primary IPv4 is 10.0.0.5 under Global 10.0.0.0/24;
// the reference "10.0.0.5/24".
// Outputs: one save, no address, and no primary IP.
// Data choice: the masked reference is how the address was written on add.
func TestRemoveIPClearsThePrimaryIP(t *testing.T) {
	// Arrange.
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	device := cmdtest.AddDevice(inventory, rackID, "cn-01", 10)
	address := addAddress(inventory, "10.0.0.5/24", addPrefix(inventory, "10.0.0.0/24", "", uuid.Nil))
	inventory.Devices[device].PrimaryIPv4 = address
	harness := cmdtest.New(t, NewCommand(), newIPCommand(), inventory)

	// Act.
	err := harness.Run(t, nil, "10.0.0.5/24")

	// Assert.
	if err != nil || harness.Store.Saves != 1 {
		t.Fatalf("remove ip: error = %v, saves = %d; want nil, 1", err, harness.Store.Saves)
	}
	if got := harness.Inventory(); got.IPAddresses[address] != nil || got.Devices[device].PrimaryIPv4 != uuid.Nil {
		t.Errorf("address kept = %v, primary IP = %v; want removed and cleared",
			got.IPAddresses[address] != nil, got.Devices[device].PrimaryIPv4)
	}
}

// TestRemoveVRFDropsItsPrefixMemberships verifies "remove vrf <name>
// --namespace tenant-a" removes that namespace's VRF and its prefix
// memberships.
//
// Why it matters: a membership naming a missing VRF fails validation, so the
// save would be refused.
// Inputs: tenant-a VRF red with tenant-a 10.0.0.0/24 as a member, and a
// Global VRF red; the name "red" with --namespace tenant-a.
// Outputs: one save; the tenant-a VRF and the membership gone, the Global VRF
// kept.
// Data choice: the Global namesake shows the flag picks the namespace.
func TestRemoveVRFDropsItsPrefixMemberships(t *testing.T) {
	// Arrange.
	inventory := devicetypes.NewInventory()
	tenantVRF, globalVRF := uuid.New(), uuid.New()
	inventory.VRFs[tenantVRF] = &devicetypes.CaniVRF{ID: tenantVRF, Name: "red", Namespace: "tenant-a"}
	inventory.VRFs[globalVRF] = &devicetypes.CaniVRF{ID: globalVRF, Name: "red"}
	member := addPrefix(inventory, "10.0.0.0/24", "tenant-a", uuid.Nil)
	inventory.Prefixes[member].VRFs = []uuid.UUID{tenantVRF}
	harness := cmdtest.New(t, NewCommand(), newVRFCommand(), inventory)

	// Act.
	err := harness.Run(t, map[string][]string{flagNamespace: {"tenant-a"}}, "red")

	// Assert.
	if err != nil || harness.Store.Saves != 1 {
		t.Fatalf("remove vrf: error = %v, saves = %d; want nil, 1", err, harness.Store.Saves)
	}
	got := harness.Inventory()
	if got.VRFs[tenantVRF] != nil || got.VRFs[globalVRF] == nil || len(got.Prefixes[member].VRFs) != 0 {
		t.Errorf("tenant VRF kept = %v, Global VRF kept = %v, memberships = %v; want false, true, none",
			got.VRFs[tenantVRF] != nil, got.VRFs[globalVRF] != nil, got.Prefixes[member].VRFs)
	}
}

// TestRemoveVRFListsTheInterfacesStillNamingIt verifies "remove vrf" reports
// the interfaces whose VRF setting names the removed VRF and leaves that
// setting as it was.
//
// Why it matters: interface VRF settings are names resolved at export, so
// nothing else tells the operator that an interface now names a VRF that is
// gone and that export will resolve or skip the name.
// Inputs: Global VRF blue; device node01 with eth0 naming "blue" and eth1
// naming "green".
// Outputs: one save; the log names 1 interface, node01/eth0; eth0 still
// reads "blue".
// Data choice: eth1 shows only the removed VRF's name is counted.
func TestRemoveVRFListsTheInterfacesStillNamingIt(t *testing.T) {
	// Arrange.
	inventory, rack := cmdtest.InventoryWithRack("x3000")
	device := cmdtest.AddDevice(inventory, rack, "node01", 1)
	inventory.Devices[device].Interfaces = []devicetypes.InterfaceSpec{
		{ID: uuid.New(), Name: "eth0", Type: "1000base-t", VRF: "blue"},
		{ID: uuid.New(), Name: "eth1", Type: "1000base-t", VRF: "green"},
	}
	blue := uuid.New()
	inventory.VRFs[blue] = &devicetypes.CaniVRF{ID: blue, Name: "blue"}
	harness := cmdtest.New(t, NewCommand(), newVRFCommand(), inventory)
	var logged bytes.Buffer
	log.SetOutput(&logged)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	// Act.
	err := harness.Run(t, nil, "blue")

	// Assert.
	if err != nil || harness.Store.Saves != 1 {
		t.Fatalf("remove vrf: error = %v, saves = %d; want nil, 1", err, harness.Store.Saves)
	}
	if !strings.Contains(logged.String(), "1 interface(s) still name it (node01/eth0)") {
		t.Errorf("log = %q, want it to list node01/eth0 as still naming the VRF", logged.String())
	}
	if got := harness.Inventory().Devices[device].Interfaces[0].VRF; got != "blue" {
		t.Errorf("eth0 VRF = %q, want the name left as \"blue\"", got)
	}
}
