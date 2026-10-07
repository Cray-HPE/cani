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

// requireAdded fails the test when arranging a record fails.
func requireAdded(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("arrange: %v", err)
	}
}

// TestAddIPAMRejectsANaturalKeyDuplicate verifies adding a prefix, address or
// VRF whose natural key another record holds fails, naming that record.
//
// Why it matters: Nautobot keeps one object per key in a namespace, so the
// second record could never be exported, and every later import would find
// the key ambiguous.
// Inputs: Global 10.0.0.0/24, 10.0.0.5/24 and VRF red; then the prefix
// 10.0.0.9/24, the address 10.0.0.5/25, and red with Global spelled out.
// Outputs: an error naming the existing record for each; nothing is added.
// Data choice: each duplicate is spelled differently from the original, so
// only the canonical natural key can catch it.
func TestAddIPAMRejectsANaturalKeyDuplicate(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	prefix, address, vrf := uuid.New(), uuid.New(), uuid.New()
	requireAdded(t, inv.AddPrefix(&CaniPrefix{ID: prefix, Prefix: "10.0.0.0/24"}))
	requireAdded(t, inv.AddIPAddress(&CaniIPAddress{ID: address, Address: "10.0.0.5/24"}))
	requireAdded(t, inv.AddVRF(&CaniVRF{ID: vrf, Name: "red"}))
	duplicates := []struct {
		name   string
		add    func() error
		holder uuid.UUID
	}{
		{"prefix", func() error { return inv.AddPrefix(&CaniPrefix{ID: uuid.New(), Prefix: "10.0.0.9/24"}) }, prefix},
		{"address", func() error { return inv.AddIPAddress(&CaniIPAddress{ID: uuid.New(), Address: "10.0.0.5/25"}) }, address},
		{"vrf", func() error { return inv.AddVRF(&CaniVRF{ID: uuid.New(), Name: "red", Namespace: "Global"}) }, vrf},
	}
	for _, tc := range duplicates {
		t.Run(tc.name, func(t *testing.T) {
			// Act.
			err := tc.add()

			// Assert.
			want := `already exists in namespace "Global" as ` + tc.holder.String()
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Errorf("error = %v, want it to contain %q", err, want)
			}
		})
	}
	if len(inv.Prefixes) != 1 || len(inv.IPAddresses) != 1 || len(inv.VRFs) != 1 {
		t.Errorf("records = %d prefixes, %d addresses, %d VRFs; want 1 each",
			len(inv.Prefixes), len(inv.IPAddresses), len(inv.VRFs))
	}
}

// TestAddIPAMAcceptsTheSameKeyElsewhere verifies uniqueness is scoped by
// namespace and compares VRF names exactly.
//
// Why it matters: namespaces exist to hold the same network twice, and
// Nautobot treats "red" and "Red" as different VRFs.
// Inputs: Global 10.0.0.0/24, 10.0.0.5/24 and VRF red; then the same prefix,
// address and VRF in tenant-a, and VRF Red in Global.
// Outputs: every add succeeds.
// Data choice: each record repeats a Global key except for its namespace or
// the case of its name.
func TestAddIPAMAcceptsTheSameKeyElsewhere(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	requireAdded(t, inv.AddPrefix(&CaniPrefix{ID: uuid.New(), Prefix: "10.0.0.0/24"}))
	requireAdded(t, inv.AddIPAddress(&CaniIPAddress{ID: uuid.New(), Address: "10.0.0.5/24"}))
	requireAdded(t, inv.AddVRF(&CaniVRF{ID: uuid.New(), Name: "red"}))

	// Act.
	errs := []error{
		inv.AddPrefix(&CaniPrefix{ID: uuid.New(), Prefix: "10.0.0.0/24", Namespace: "tenant-a"}),
		inv.AddIPAddress(&CaniIPAddress{ID: uuid.New(), Address: "10.0.0.5/24", Namespace: "tenant-a"}),
		inv.AddVRF(&CaniVRF{ID: uuid.New(), Name: "red", Namespace: "tenant-a"}),
		inv.AddVRF(&CaniVRF{ID: uuid.New(), Name: "Red"}),
	}

	// Assert.
	for i, err := range errs {
		if err != nil {
			t.Errorf("add %d returned error: %v", i, err)
		}
	}
	if len(inv.Prefixes) != 2 || len(inv.IPAddresses) != 2 || len(inv.VRFs) != 3 {
		t.Errorf("records = %d prefixes, %d addresses, %d VRFs; want 2, 2, 3",
			len(inv.Prefixes), len(inv.IPAddresses), len(inv.VRFs))
	}
}
