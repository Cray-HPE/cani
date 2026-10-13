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
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// TestCanRemovePrefixReportsWhatRemovePrefixWouldRefuse verifies the check
// gives the same answer as the removal without performing it.
//
// Why it matters: a caller that names a removal as a remedy, such as an import
// conflict message, must know whether the plain removal will be refused so it
// can name the cascade instead; the check must not itself change anything.
// Inputs: a root 10.0.0.0/16 holding 10.0.0.5, a child 10.0.1.0/24 holding
// nothing, and a random UUID.
// Outputs: ErrPrefixHoldsAddresses for the root, nil for the child, a not-found
// error for the UUID; both prefixes and the address are still present.
// Data choice: the child shows the refusal is about parentless addresses, not
// about holding addresses at all.
func TestCanRemovePrefixReportsWhatRemovePrefixWouldRefuse(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	root, child, address := uuid.New(), uuid.New(), uuid.New()
	inv.Prefixes[root] = &CaniPrefix{ID: root, Prefix: "10.0.0.0/16"}
	inv.Prefixes[child] = &CaniPrefix{ID: child, Prefix: "10.0.1.0/24", Parent: root}
	inv.IPAddresses[address] = &CaniIPAddress{ID: address, Address: "10.0.0.5/16", Parent: root}

	// Act.
	rootErr := inv.CanRemovePrefix(root)
	childErr := inv.CanRemovePrefix(child)
	unknownErr := inv.CanRemovePrefix(uuid.New())

	// Assert.
	if !errors.Is(rootErr, ErrPrefixHoldsAddresses) {
		t.Errorf("CanRemovePrefix(root) = %v, want ErrPrefixHoldsAddresses", rootErr)
	}
	if childErr != nil {
		t.Errorf("CanRemovePrefix(child) = %v, want nil", childErr)
	}
	if unknownErr == nil || !strings.Contains(unknownErr.Error(), "not found") {
		t.Errorf("CanRemovePrefix(unknown) = %v, want a not-found error", unknownErr)
	}
	if len(inv.Prefixes) != 2 || inv.IPAddresses[address].Parent != root {
		t.Error("the check changed the inventory")
	}
}

// TestInterfacesNamingVRFListsBareAndQualifiedReferences verifies the query
// finds the interfaces whose VRF setting names a VRF, written bare or as
// namespace/name, and ignores other names.
//
// Why it matters: RemoveVRF leaves interface VRF names in place, so what still
// refers to a removed VRF is only known through this query; export later
// resolves or skips those names silently.
// Inputs: VRF blue in tenant-a; indexed interfaces naming "blue",
// "tenant-a/blue", "green" and nothing.
// Outputs: the two blue interfaces in ID order; the others absent.
// Data choice: the qualified spelling is the one export's reference grammar
// accepts, so both spellings must count.
func TestInterfacesNamingVRFListsBareAndQualifiedReferences(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	blue, bare, qualified, other, none := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.VRFs[blue] = &CaniVRF{ID: blue, Name: "blue", Namespace: "tenant-a"}
	inv.Interfaces[bare] = &CaniInterface{ID: bare, Name: "eth0", VRF: "blue"}
	inv.Interfaces[qualified] = &CaniInterface{ID: qualified, Name: "eth1", VRF: "tenant-a/blue"}
	inv.Interfaces[other] = &CaniInterface{ID: other, Name: "eth2", VRF: "green"}
	inv.Interfaces[none] = &CaniInterface{ID: none, Name: "eth3"}

	// Act.
	named := inv.InterfacesNamingVRF(inv.VRFs[blue])

	// Assert.
	got := make([]uuid.UUID, 0, len(named))
	for _, iface := range named {
		got = append(got, iface.ID)
	}
	want := []uuid.UUID{bare, qualified}
	if bytes.Compare(qualified[:], bare[:]) < 0 {
		want = []uuid.UUID{qualified, bare}
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("InterfacesNamingVRF = %v, want %v", got, want)
	}
}
