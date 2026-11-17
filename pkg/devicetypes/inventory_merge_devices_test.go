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
	"testing"

	"github.com/google/uuid"
)

// Test coverage for the device merge's identity contract in
// inventory_merge_devices.go: providerIdentityCompatible, IdentityValue and
// the keys MergeTransformResult reads from TransformResult.IdentityKeys.

// ---------- providerIdentityCompatible ----------

// redfishKeys are the identity keys the Redfish provider declares.
var redfishKeys = []string{"bmc_fqdn", "bmc_hostname"}

func TestProviderIdentityCompatibleMatch(t *testing.T) {
	a := &CaniDeviceType{
		Name: "server-1",
		ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{
			"redfish": map[string]any{"bmc_fqdn": "bmc1.example.com"},
		}},
	}
	b := &CaniDeviceType{
		Name: "server-1",
		ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{
			"redfish": map[string]any{"bmc_fqdn": "bmc1.example.com"},
		}},
	}

	if !providerIdentityCompatible(a, b, redfishKeys) {
		t.Error("expected compatible when bmc_fqdn values match")
	}
}

func TestProviderIdentityCompatibleConflict(t *testing.T) {
	a := &CaniDeviceType{
		Name: "server-1",
		ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{
			"redfish": map[string]any{"bmc_fqdn": "bmc1.example.com"},
		}},
	}
	b := &CaniDeviceType{
		Name: "server-1",
		ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{
			"redfish": map[string]any{"bmc_fqdn": "bmc2.example.com"},
		}},
	}

	if providerIdentityCompatible(a, b, redfishKeys) {
		t.Error("expected incompatible when bmc_fqdn values differ")
	}
}

// TestProviderIdentityComparesDeclaredKeysOnly verifies that only the keys
// the provider declared decide whether two devices of one name are one, and
// that with no keys declared any two are.
//
// Why it matters: the model does not know which keys identify a device; a
// provider says so on its TransformResult. A key it did not declare, such
// as a serial number both sides happen to record, must not keep two records
// of one node apart, and a provider that declares nothing merges by name as
// every provider did before the contract existed.
// Inputs: two devices of one name whose Redfish buckets hold different
// bmc_fqdn values and different serial numbers, compared with the Redfish
// keys, with a key neither declares, and with none.
// Outputs: incompatible, compatible, compatible.
// Data choice: the serial differs too, so the second result shows the
// undeclared key is ignored rather than compared and found equal.
func TestProviderIdentityComparesDeclaredKeysOnly(t *testing.T) {
	// Arrange.
	device := func(fqdn string) *CaniDeviceType {
		return &CaniDeviceType{Name: "server-1", ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{
			"redfish": map[string]any{"bmc_fqdn": fqdn, "serial": fqdn + "-sn"},
		}}}
	}
	a, b := device("bmc1.example.com"), device("bmc2.example.com")

	// Act.
	declared := providerIdentityCompatible(a, b, redfishKeys)
	undeclared := providerIdentityCompatible(a, b, []string{"redfish_uuid"})
	none := providerIdentityCompatible(a, b, nil)

	// Assert.
	if declared || !undeclared || !none {
		t.Errorf("compatible: declared keys = %t, undeclared key = %t, no keys = %t; want false, true, true", declared, undeclared, none)
	}
}

// TestProviderIdentityComparesAcrossProvidersBuckets verifies an identity
// key is compared whichever provider's bucket holds it on each side, and
// that a value outside a bucket is not an identity.
//
// Why it matters: two imports of one node by different providers must not
// be merged when they disagree on its xname, although each keeps the xname
// in its own bucket. Declaring xname asserts it means the same in every
// bucket; this is the one place that still looks in every bucket, and it
// must not start reading values no provider owns.
// Inputs: a device whose CSM bucket holds an xname, compared with devices
// whose OChaMi bucket holds the same xname, a different one, and with a
// device that holds a different xname outside any bucket, xname declared.
// Outputs: compatible, incompatible, compatible.
// Data choice: the third device's value would make it incompatible if a
// value outside a bucket still counted.
func TestProviderIdentityComparesAcrossProvidersBuckets(t *testing.T) {
	// Arrange.
	device := func(provider map[string]any) *CaniDeviceType {
		return &CaniDeviceType{Name: "nid000001", ObjectMeta: ObjectMeta{ProviderMetadata: provider}}
	}
	fromCSM := device(map[string]any{"csm": map[string]any{"xname": "x3000c0s1b0n0"}})
	xname := []string{"xname"}

	// Act.
	same := providerIdentityCompatible(fromCSM, device(map[string]any{"ochami": map[string]any{"xname": "x3000c0s1b0n0"}}), xname)
	other := providerIdentityCompatible(fromCSM, device(map[string]any{"ochami": map[string]any{"xname": "x3000c0s2b0n0"}}), xname)
	unowned := providerIdentityCompatible(fromCSM, device(map[string]any{"xname": "x3000c0s2b0n0"}), xname)

	// Assert.
	if !same || other || !unowned {
		t.Errorf("compatible: same xname = %t, other xname = %t, xname outside a bucket = %t; want true, false, true", same, other, unowned)
	}
}

// TestMergeTransformResultKeepsDevicesApartByDeclaredIdentity verifies the
// merge of a transform result reads the identity keys the result declares:
// an incoming device of an existing name with a different value for a
// declared key is inserted beside the record, not merged onto it.
//
// Why it matters: the keys only matter if the import path consults them; a
// result that declared xname and was merged by name alone would fold two
// nodes that share a name into one record.
// Inputs: an inventory with one node and a result holding a node of the same
// name with a different xname, once with xname declared and once without.
// Outputs: two devices after the first merge, one after the second.
// Data choice: the same result is merged both ways, so the declaration alone
// makes the difference.
func TestMergeTransformResultKeepsDevicesApartByDeclaredIdentity(t *testing.T) {
	// Arrange.
	node := func(xname string) *CaniDeviceType {
		return &CaniDeviceType{ID: uuid.New(), Name: "nid000001", Type: TypeNode, ObjectMeta: ObjectMeta{
			Status: string(StatusActive), ProviderMetadata: map[string]any{"csm": map[string]any{"xname": xname}},
		}}
	}
	result := func(keys []string) *TransformResult {
		incoming := node("x3000c0s2b0n0")
		return &TransformResult{Devices: map[uuid.UUID]*CaniDeviceType{incoming.ID: incoming}, IdentityKeys: keys}
	}
	declared, undeclared := NewInventory(), NewInventory()
	for _, inv := range []*Inventory{declared, undeclared} {
		existing := node("x3000c0s1b0n0")
		inv.Devices[existing.ID] = existing
	}

	// Act.
	if _, err := declared.MergeTransformResult(result([]string{"xname"})); err != nil {
		t.Fatalf("MergeTransformResult(declared): %v", err)
	}
	if _, err := undeclared.MergeTransformResult(result(nil)); err != nil {
		t.Fatalf("MergeTransformResult(undeclared): %v", err)
	}

	// Assert.
	if len(declared.Devices) != 2 || len(undeclared.Devices) != 1 {
		t.Errorf("devices after merge: xname declared = %d, none declared = %d; want 2 and 1", len(declared.Devices), len(undeclared.Devices))
	}
}
