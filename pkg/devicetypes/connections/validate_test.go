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
package connections

import (
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// TestFilterInterfaceConflicts_DropsSecondCableOnSamePort verifies that when two
// connections claim the same (device, port) endpoint, the first in input order
// is kept and the second is returned as a conflict naming the contested port and
// the winning connection.
//
// Why it matters: a physical port hosts at most one cable. Without this filter,
// over-subscribed connection maps (e.g. one NAS port wired to two switches) push
// impossible cables into the inventory that the exporter later skips with a
// confusing "already exists" message.
// Inputs: two resolved connections that share endpoint A (nas:e0e), wired to two
// different leaf ports. Outputs: kept==[first], dropped==[second] with
// ConflictDevice/Port = nas:e0e and Winner = first.
// Data choice: identical A endpoint but distinct B endpoints isolates the
// shared-port rule and makes a wrong winner or wrong drop obvious.
func TestFilterInterfaceConflicts_DropsSecondCableOnSamePort(t *testing.T) {
	nas := uuid.New()
	leaf1 := uuid.New()
	leaf2 := uuid.New()

	first := ResolvedConnection{ADevice: nas, APort: "e0e", BDevice: leaf1, BPort: "1/1/26"}
	second := ResolvedConnection{ADevice: nas, APort: "e0e", BDevice: leaf2, BPort: "1/1/26"}

	kept, dropped := FilterInterfaceConflicts([]ResolvedConnection{first, second})

	if len(kept) != 1 || kept[0].BDevice != leaf1 {
		t.Fatalf("kept = %+v, want exactly the first connection (B=leaf1)", kept)
	}
	if len(dropped) != 1 {
		t.Fatalf("dropped = %d, want 1", len(dropped))
	}
	got := dropped[0]
	if got.ConflictDevice != nas || got.ConflictPort != "e0e" {
		t.Errorf("conflict endpoint = %s:%s, want %s:e0e", got.ConflictDevice, got.ConflictPort, nas)
	}
	if got.Dropped.BDevice != leaf2 {
		t.Errorf("dropped connection B = %s, want leaf2 %s", got.Dropped.BDevice, leaf2)
	}
	if got.Winner.BDevice != leaf1 {
		t.Errorf("winner B = %s, want leaf1 %s", got.Winner.BDevice, leaf1)
	}
}

// TestFilterInterfaceConflicts_DetectsReuseOnBSide verifies that the filter
// claims both endpoints of a kept connection, so a later connection reusing a
// port that first appeared on the B side is also dropped.
//
// Why it matters: port reuse can occur on either termination (e.g. a leaf
// uplink port reused by a storage cable); guarding only the A side would let
// half the conflicts through.
// Inputs: connection 1 claims leaf:1/1/28 on its B side; connection 2 reuses
// leaf:1/1/28 on its A side. Outputs: kept==[conn1], dropped==[conn2] with the
// contested endpoint = leaf:1/1/28.
// Data choice: placing the shared port on opposite sides (B then A) proves the
// claim is side-agnostic.
func TestFilterInterfaceConflicts_DetectsReuseOnBSide(t *testing.T) {
	spine := uuid.New()
	leaf := uuid.New()
	nas := uuid.New()

	uplink := ResolvedConnection{ADevice: spine, APort: "1/1/1", BDevice: leaf, BPort: "1/1/28"}
	storage := ResolvedConnection{ADevice: leaf, APort: "1/1/28", BDevice: nas, BPort: "e0f"}

	kept, dropped := FilterInterfaceConflicts([]ResolvedConnection{uplink, storage})

	if len(kept) != 1 || kept[0].ADevice != spine {
		t.Fatalf("kept = %+v, want only the uplink", kept)
	}
	if len(dropped) != 1 {
		t.Fatalf("dropped = %d, want 1", len(dropped))
	}
	if dropped[0].ConflictDevice != leaf || dropped[0].ConflictPort != "1/1/28" {
		t.Errorf("conflict endpoint = %s:%s, want leaf:1/1/28",
			dropped[0].ConflictDevice, dropped[0].ConflictPort)
	}
}

// TestFilterInterfaceConflicts_KeepsDistinctEndpoints verifies that connections
// using entirely distinct endpoints are all kept in input order with no
// conflicts reported.
//
// Why it matters: the common, valid case (every port used once) must pass
// through untouched; a false positive here would silently drop real cabling.
// Inputs: three connections on six distinct device:port endpoints. Outputs:
// kept has all three in order, dropped is empty.
// Data choice: fully disjoint endpoints guarantee the no-conflict path is the
// only one exercised.
func TestFilterInterfaceConflicts_KeepsDistinctEndpoints(t *testing.T) {
	a, b, c, d := uuid.New(), uuid.New(), uuid.New(), uuid.New()

	conns := []ResolvedConnection{
		{ADevice: a, APort: "p1", BDevice: b, BPort: "p1"},
		{ADevice: a, APort: "p2", BDevice: c, BPort: "p1"},
		{ADevice: b, APort: "p2", BDevice: d, BPort: "p1"},
	}

	kept, dropped := FilterInterfaceConflicts(conns)

	if len(dropped) != 0 {
		t.Fatalf("dropped = %d, want 0", len(dropped))
	}
	if len(kept) != 3 {
		t.Fatalf("kept = %d, want 3", len(kept))
	}
	for i := range conns {
		if kept[i].APort != conns[i].APort || kept[i].BPort != conns[i].BPort {
			t.Errorf("kept[%d] = %s/%s, want %s/%s", i,
				kept[i].APort, kept[i].BPort, conns[i].APort, conns[i].BPort)
		}
	}
}

// TestInterfaceConflict_Describe verifies that Describe resolves device UUIDs to
// names and reports the dropped cable, the contested port, and the winning peer.
//
// Why it matters: the dropped-cable warning is the operator's only signal that
// their connection map over-subscribed a port; it must name the real devices,
// not raw UUIDs, to be actionable.
// Inputs: an inventory naming nas->"NAS-01" and two leaves, plus a conflict where
// NAS-01:e0e is reused. Outputs: a string containing the dropped endpoints, the
// contested "NAS-01:e0e", and the winning "leaf-1" peer.
// Data choice: human-readable names distinct from UUIDs make a failed lookup or
// wrong field interpolation immediately visible.
func TestInterfaceConflict_Describe(t *testing.T) {
	inv := devicetypes.NewInventory()
	nas, leaf1, leaf2 := uuid.New(), uuid.New(), uuid.New()
	inv.Devices[nas] = &devicetypes.CaniDeviceType{ID: nas, Name: "NAS-01"}
	inv.Devices[leaf1] = &devicetypes.CaniDeviceType{ID: leaf1, Name: "leaf-1"}
	inv.Devices[leaf2] = &devicetypes.CaniDeviceType{ID: leaf2, Name: "leaf-2"}

	conflict := InterfaceConflict{
		Dropped:        ResolvedConnection{ADevice: nas, APort: "e0e", BDevice: leaf2, BPort: "1/1/26"},
		Winner:         ResolvedConnection{ADevice: nas, APort: "e0e", BDevice: leaf1, BPort: "1/1/26"},
		ConflictDevice: nas,
		ConflictPort:   "e0e",
	}

	msg := conflict.Describe(inv)

	for _, want := range []string{"NAS-01:e0e", "leaf-2:1/1/26", "leaf-1:1/1/26", "already cabled"} {
		if !strings.Contains(msg, want) {
			t.Errorf("Describe() = %q, want it to contain %q", msg, want)
		}
	}
}
