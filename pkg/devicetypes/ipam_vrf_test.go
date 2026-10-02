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
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// vrfWithDevice returns an inventory holding one device and one Global VRF,
// and both IDs.
func vrfWithDevice() (*Inventory, *CaniVRF, uuid.UUID) {
	inv := NewInventory()
	device := uuid.New()
	inv.Devices[device] = &CaniDeviceType{ID: device, Name: "cn-01"}
	vrf := &CaniVRF{ID: uuid.New(), Name: "blue"}
	inv.VRFs[vrf.ID] = vrf
	return inv, vrf, device
}

// TestAddVRFDeviceAssignsOnceAndRequiresTheDevice verifies AddVRFDevice adds
// an existing device once, reports a repeat as no change, and refuses a
// device the inventory does not hold.
//
// Why it matters: AddVRF validates a new VRF's devices; an assignment added
// later must meet the same rule or export fails on a device Nautobot lacks.
// Inputs: a VRF, an existing device added twice, then a random UUID.
// Outputs: (true, nil), (false, nil), then an error naming the device, with
// the assignment list holding the one device throughout.
// Data choice: the repeat and the unknown UUID each exercise one guard.
func TestAddVRFDeviceAssignsOnceAndRequiresTheDevice(t *testing.T) {
	// Arrange.
	inv, vrf, device := vrfWithDevice()
	unknown := uuid.New()

	// Act.
	first, firstErr := inv.AddVRFDevice(vrf, device)
	again, againErr := inv.AddVRFDevice(vrf, device)
	_, unknownErr := inv.AddVRFDevice(vrf, unknown)

	// Assert.
	if !first || firstErr != nil || again || againErr != nil {
		t.Fatalf("AddVRFDevice = (%v, %v) then (%v, %v); want (true, nil) then (false, nil)", first, firstErr, again, againErr)
	}
	if unknownErr == nil || !strings.Contains(unknownErr.Error(), unknown.String()) {
		t.Errorf("unknown device error = %v, want one naming %s", unknownErr, unknown)
	}
	if !reflect.DeepEqual(vrf.Devices, []uuid.UUID{device}) {
		t.Errorf("Devices = %v, want [%v]", vrf.Devices, device)
	}
}

// TestRemoveDeviceDropsAnAssignmentWhoseDeviceIsGone verifies RemoveDevice
// drops an assignment whether or not the device still exists and reports an
// absent assignment as no change.
//
// Why it matters: a device removed from the inventory can leave its VRF
// assignment behind; the VRF must be able to shed it.
// Inputs: a VRF assigned a UUID with no device; remove it, then remove it
// again.
// Outputs: true then false, with no assignments left.
// Data choice: no device exists for the UUID, so existence cannot be what
// the method checks.
func TestRemoveDeviceDropsAnAssignmentWhoseDeviceIsGone(t *testing.T) {
	// Arrange.
	gone := uuid.New()
	vrf := &CaniVRF{ID: uuid.New(), Name: "blue", Devices: []uuid.UUID{gone}}

	// Act.
	removed := vrf.RemoveDevice(gone)
	again := vrf.RemoveDevice(gone)

	// Assert.
	if !removed || again || len(vrf.Devices) != 0 {
		t.Errorf("RemoveDevice = %v then %v with %v left; want true, false, none", removed, again, vrf.Devices)
	}
}

// TestResolveVRFReferenceListsAmbiguousCandidates verifies an ambiguous name
// is refused with every candidate UUID, as the prefix and address resolvers do.
//
// Why it matters: an operator told only "use the UUID" has to run a second
// command to find it; the error should carry the choice.
// Inputs: two Global VRFs named blue; the reference "blue".
// Outputs: an error containing both UUIDs.
// Data choice: exact duplicates leave nothing but the UUIDs to choose by.
func TestResolveVRFReferenceListsAmbiguousCandidates(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	first, second := uuid.New(), uuid.New()
	inv.VRFs[first] = &CaniVRF{ID: first, Name: "blue"}
	inv.VRFs[second] = &CaniVRF{ID: second, Name: "blue"}

	// Act.
	_, err := inv.ResolveVRFReference("Global", "blue")

	// Assert.
	if err == nil || !strings.Contains(err.Error(), first.String()) || !strings.Contains(err.Error(), second.String()) {
		t.Errorf("error = %v, want both %s and %s", err, first, second)
	}
}
