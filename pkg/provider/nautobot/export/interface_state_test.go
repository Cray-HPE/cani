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
package export

import (
	"encoding/json"
	"testing"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// TestInterfaceCurrentSkipsMatchingInterface verifies that an existing
// interface whose remote state equals what the spec would write is reported
// current, and that a differing type, a differing description, or unknown
// remote state is not.
//
// Why it matters: --merge used to PATCH every existing interface, so an
// idle export wrote to Nautobot; the comparison is what makes it idempotent.
// Inputs: an Active status seeded in the cache; a remote interface of type
// 1000base-t with description "uplink" and upper-case MAC; specs that match,
// change the type, change the description, and an item without state.
// Outputs: true, false, false, false.
// Data choice: the matching spec gives the MAC in lower case to show the
// comparison normalises it the way Nautobot stores it.
func TestInterfaceCurrentSkipsMatchingInterface(t *testing.T) {
	statusID := uuid.New()
	cache := NewLookupCache(nil)
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	e := &Exporter{Cache: cache}
	existing := &CachedItem{ID: uuid.New(), Name: "eth0", Iface: &interfaceState{
		Type: "1000base-t", Mac: "AA:BB:CC:DD:EE:FF", Description: "uplink", StatusID: statusID,
	}}
	matching := interfaceSpec{Name: "eth0", Type: "1000base-t", Mac: "aa:bb:cc:dd:ee:ff", Description: "uplink"}

	if !e.interfaceCurrent(existing, matching) {
		t.Fatal("an interface that already matches must be current")
	}
	retyped := matching
	retyped.Type = "10gbase-x-sfpp"
	if e.interfaceCurrent(existing, retyped) {
		t.Fatal("a differing type must not be current")
	}
	redescribed := matching
	redescribed.Description = ""
	if e.interfaceCurrent(existing, redescribed) {
		t.Fatal("an emptied description is written unconditionally, so it must not be current")
	}
	if e.interfaceCurrent(&CachedItem{ID: uuid.New(), Name: "eth0"}, matching) {
		t.Fatal("unknown remote state must not be current")
	}
}

// TestRemoteInterfaceStateReadsGeneratedShape verifies that the state
// captured from a generated Interface carries its type, mgmt_only flag,
// status and tag UUIDs, and reads a nil tags slice as no tags.
//
// Why it matters: the comparison is only as good as the snapshot taken from
// the list response; a field read from the wrong place would make every
// interface look stale or, worse, current.
// Inputs: an Interface decoded from JSON with two tags and mgmt_only true.
// Outputs: Type "1000base-t", MgmtOnly true, the status UUID, both tag UUIDs
// joined in sorted order; a second interface with no tags yields "".
// Data choice: JSON input mirrors the wire shape instead of naming the
// generated anonymous struct types.
func TestRemoteInterfaceStateReadsGeneratedShape(t *testing.T) {
	statusID, tagA, tagB := uuid.New(), uuid.New(), uuid.New()
	first, second := tagA.String(), tagB.String()
	if second < first {
		first, second = second, first
	}
	var iface nautobotapi.Interface
	body := `{"name":"eth0","type":{"value":"1000base-t","label":"1000BASE-T"},"mgmt_only":true,` +
		`"status":{"id":"` + statusID.String() + `"},"tags":[{"id":"` + tagB.String() + `"},{"id":"` + tagA.String() + `"}]}`
	if err := json.Unmarshal([]byte(body), &iface); err != nil {
		t.Fatal(err)
	}

	state := remoteInterfaceState(iface)

	if state.Type != "1000base-t" || !state.MgmtOnly || state.StatusID != statusID {
		t.Fatalf("state = %+v", *state)
	}
	if state.Tags != first+","+second {
		t.Fatalf("tags = %q, want %q", state.Tags, first+","+second)
	}
	if got := remoteInterfaceState(nautobotapi.Interface{Name: "eth1"}); got.Tags != "" {
		t.Fatalf("an interface without tags must yield no tag key, got %q", got.Tags)
	}
}
