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
package devicetypes

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// TestValidateReportsObjectErrorsFromEveryCollection verifies that
// Inventory.Validate surfaces per-object Validate failures from the IPAM and
// interface collections, not just the six DCIM ones.
//
// Why it matters: Validate previously performed only referential-integrity
// checks over locations, racks, devices, modules, cables and FRUs. Each entity
// had its own Validate that nothing ever called, and prefixes, IP addresses,
// VLANs, VRFs and interfaces were not examined at all, so malformed IPAM could
// be saved and later exported to Nautobot.
//
// Inputs: an inventory holding one invalid entity in each of the five
// previously unchecked collections. Output: an error naming all five.
//
// Data choice: each entity breaks a different rule its own Validate enforces
// (out-of-range VID, non-CIDR prefix, non-IP host, empty VRF name, interface
// with no device), which proves the walk reaches each collection and delegates
// rather than re-implementing the checks.
func TestValidateReportsObjectErrorsFromEveryCollection(t *testing.T) {
	inv := NewInventory()

	vlanID, prefixID, ipID, vrfID, ifaceID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.VLANs[vlanID] = &CaniVLAN{ID: vlanID, VID: 9999, Name: "bad-vid"}
	inv.Prefixes[prefixID] = &CaniPrefix{ID: prefixID, Prefix: "not-a-cidr"}
	inv.IPAddresses[ipID] = &CaniIPAddress{ID: ipID, Host: "999.999.999.999"}
	inv.VRFs[vrfID] = &CaniVRF{ID: vrfID}
	inv.Interfaces[ifaceID] = &CaniInterface{ID: ifaceID, Name: "eth0"}

	err := inv.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want an error naming the invalid objects")
	}

	for _, want := range []string{
		"vlan " + vlanID.String(),
		"prefix " + prefixID.String(),
		"ip address " + ipID.String(),
		"vrf " + vrfID.String(),
		"interface " + ifaceID.String(),
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Validate() error missing %q; got:\n%v", want, err)
		}
	}
}

// TestValidateObjectErrorsAreOrdered verifies that repeated Validate calls on
// the same inventory produce byte-identical error text.
//
// Why it matters: the object walk ranges Go maps, whose iteration order is
// randomized. Without sorting, the same broken inventory would report its
// problems in a different order on every run, which makes the message hard to
// diff and unusable as a test fixture or CI artifact.
//
// Inputs: an inventory with three invalid VLANs. Output: the same error string
// from every call. Data choice: three entities in one collection is the
// smallest input that can expose map-order instability.
func TestValidateObjectErrorsAreOrdered(t *testing.T) {
	inv := NewInventory()
	for range 3 {
		id := uuid.New()
		inv.VLANs[id] = &CaniVLAN{ID: id, VID: 0}
	}

	first := inv.Validate()
	if first == nil {
		t.Fatal("Validate() = nil, want an error")
	}
	for range 10 {
		if got := inv.Validate(); got.Error() != first.Error() {
			t.Fatalf("Validate() error text is unstable:\nfirst: %v\nlater: %v", first, got)
		}
	}
}

// TestValidatePassesForValidObjects verifies that a well-formed inventory
// containing IPAM and interface entities validates cleanly.
//
// Why it matters: the object walk is additive, so it must not reject data that
// was previously accepted; a false positive here would block saves and imports
// for every existing user.
//
// Inputs: one valid entity in each newly covered collection. Output: nil.
// Data choice: minimal but complete instances that satisfy each type's own
// Validate, mirroring what the transform stage produces.
func TestValidatePassesForValidObjects(t *testing.T) {
	inv := NewInventory()

	deviceID := uuid.New()
	inv.Devices[deviceID] = &CaniDeviceType{ID: deviceID, Name: "node1"}

	vlanID, prefixID, ipID, vrfID, ifaceID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.VLANs[vlanID] = &CaniVLAN{ID: vlanID, VID: 100, Name: "hmn"}
	inv.Prefixes[prefixID] = &CaniPrefix{ID: prefixID, Prefix: "10.0.0.0/24"}
	inv.IPAddresses[ipID] = &CaniIPAddress{ID: ipID, Host: "10.0.0.1", Address: "10.0.0.1/24"}
	inv.VRFs[vrfID] = &CaniVRF{ID: vrfID, Name: "default"}
	inv.Interfaces[ifaceID] = &CaniInterface{ID: ifaceID, Name: "eth0", DeviceID: deviceID}

	if err := inv.Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}
