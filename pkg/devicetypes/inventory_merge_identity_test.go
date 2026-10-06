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
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// TestMergePrefixesFreesTheOldKeyOfAMovedRecord verifies a record its source
// moved to another namespace stops answering to its old key within the same
// batch, so a new object at that old key becomes a record of its own.
//
// Why it matters: indexing natural keys as they stood before the batch let
// the new object land on the moved record and overwrite it, losing the move.
// Inputs: Global 10.0.0.0/24 with source ID S; incoming S now in tenant-a, and
// a Global 10.0.0.0/24 with no source ID. Outputs: two records, the original
// UUID in tenant-a and the new object under its own UUID in Global.
// Data choice: the moved object sorts first, the order that exposed the stale
// key.
func TestMergePrefixesFreesTheOldKeyOfAMovedRecord(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	original, sourceID := uuid.New(), uuid.New()
	inv.Prefixes[original] = &CaniPrefix{ID: original, Prefix: "10.0.0.0/24", Namespace: "Global",
		ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": sourceID}}}
	moved := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	fresh := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	incoming := map[uuid.UUID]*CaniPrefix{
		moved: {ID: moved, Prefix: "10.0.0.0/24", Namespace: "tenant-a",
			ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": sourceID}}},
		fresh: {ID: fresh, Prefix: "10.0.0.0/24", Namespace: "Global"},
	}

	// Act.
	remap, err := inv.MergePrefixes(incoming)

	// Assert.
	if err != nil {
		t.Fatalf("MergePrefixes returned error: %v", err)
	}
	if remap[moved] != original || remap[fresh] != fresh || len(inv.Prefixes) != 2 {
		t.Fatalf("remap = %v with %d records, want moved -> %v, fresh kept, 2 records", remap, len(inv.Prefixes), original)
	}
	if got := inv.Prefixes[original].Namespace; got != "tenant-a" {
		t.Errorf("moved record namespace = %q, want tenant-a", got)
	}
}

// TestMergePrefixesRejectsTakeoverOfAnOccupiedKey verifies a source match
// that moves a record onto a natural key another record holds is an error
// and leaves the inventory unchanged.
//
// Why it matters: a locally authored 10.0.0.0/24 and an imported 10.1.0.0/24
// that the source later renumbers to 10.0.0.0/24 would otherwise merge into
// two records under one key, which Nautobot rejects on export and nothing
// reported without --debug.
// Inputs: local Global 10.0.0.0/24 without a source ID, imported Global
// 10.1.0.0/24 with source ID S, and incoming S now at 10.0.0.0/24.
// Outputs: an error naming both records; both records keep their CIDRs.
// Data choice: the local record has no source ID, so only the key can
// collide.
func TestMergePrefixesRejectsTakeoverOfAnOccupiedKey(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	local, imported, sourceID := uuid.New(), uuid.New(), uuid.New()
	inv.Prefixes[local] = &CaniPrefix{ID: local, Prefix: "10.0.0.0/24", Namespace: "Global"}
	inv.Prefixes[imported] = &CaniPrefix{ID: imported, Prefix: "10.1.0.0/24", Namespace: "Global",
		ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": sourceID}}}
	incoming := map[uuid.UUID]*CaniPrefix{uuid.New(): {Prefix: "10.0.0.0/24", Namespace: "Global",
		ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": sourceID}}}}

	// Act.
	_, err := inv.MergePrefixes(incoming)

	// Assert.
	want := fmt.Sprintf(`prefix 10.0.0.0/24 in namespace "Global": record %s now holds the natural key of record %s`, imported, local)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("MergePrefixes error = %v, want containing %q", err, want)
	}
	if inv.Prefixes[imported].Prefix != "10.1.0.0/24" || inv.Prefixes[local].Prefix != "10.0.0.0/24" || len(inv.Prefixes) != 2 {
		t.Errorf("collection changed on error: %d prefixes, imported = %q", len(inv.Prefixes), inv.Prefixes[imported].Prefix)
	}
}

// TestMergePrefixesAllowsRecordsToSwapKeys verifies two records whose source
// exchanged their CIDRs in one batch both move, with no takeover error.
//
// Why it matters: the occupied-key check must look at the index after the
// whole source pass; checking as each record moves would see the first move
// land on a key its partner has not yet left.
// Inputs: Global 10.0.0.0/24 (source A) and 10.1.0.0/24 (source B); incoming
// A at 10.1.0.0/24 and B at 10.0.0.0/24.
// Outputs: no error, the same two UUIDs, each now holding the other's CIDR.
// Data choice: a pure swap is the smallest batch in which every intermediate
// state has a duplicate key but the final state has none.
func TestMergePrefixesAllowsRecordsToSwapKeys(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	first, second, sourceA, sourceB := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.Prefixes[first] = &CaniPrefix{ID: first, Prefix: "10.0.0.0/24", Namespace: "Global",
		ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": sourceA}}}
	inv.Prefixes[second] = &CaniPrefix{ID: second, Prefix: "10.1.0.0/24", Namespace: "Global",
		ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": sourceB}}}
	incoming := map[uuid.UUID]*CaniPrefix{
		uuid.New(): {Prefix: "10.1.0.0/24", Namespace: "Global", ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": sourceA}}},
		uuid.New(): {Prefix: "10.0.0.0/24", Namespace: "Global", ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": sourceB}}},
	}

	// Act.
	_, err := inv.MergePrefixes(incoming)

	// Assert.
	if err != nil {
		t.Fatalf("MergePrefixes returned error for a key swap: %v", err)
	}
	if len(inv.Prefixes) != 2 || inv.Prefixes[first].Prefix != "10.1.0.0/24" || inv.Prefixes[second].Prefix != "10.0.0.0/24" {
		t.Errorf("after swap: %d records, first = %q, second = %q; want 2 records with exchanged CIDRs",
			len(inv.Prefixes), inv.Prefixes[first].Prefix, inv.Prefixes[second].Prefix)
	}
}

// TestMergePrefixesRejectsSourceIDsNamingTwoRecords verifies an incoming
// object whose provider IDs point at two different records is an error that
// names both, and leaves the inventory unchanged.
//
// Why it matters: the two records are one object the sources never merged
// locally; taking the lower UUID would update one and leave the other
// claiming the same source ID, a guess the merge otherwise never makes.
// Inputs: 10.0.0.0/24 known to nautobot as X and 10.1.0.0/24 known to csm as
// Y; incoming 10.2.0.0/24 carrying both X and Y.
// Outputs: an error naming the object and both records; both CIDRs kept.
// Data choice: two providers on two records is the only way one object can
// match more than one record by source ID.
func TestMergePrefixesRejectsSourceIDsNamingTwoRecords(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	byNautobot := uuid.MustParse("00000000-0000-0000-0000-00000000000a")
	byCSM := uuid.MustParse("00000000-0000-0000-0000-00000000000b")
	x, y := uuid.New(), uuid.New()
	inv.Prefixes[byNautobot] = &CaniPrefix{ID: byNautobot, Prefix: "10.0.0.0/24",
		ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": x}}}
	inv.Prefixes[byCSM] = &CaniPrefix{ID: byCSM, Prefix: "10.1.0.0/24",
		ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"csm": y}}}
	incoming := map[uuid.UUID]*CaniPrefix{uuid.New(): {Prefix: "10.2.0.0/24",
		ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": x, "csm": y}}}}

	// Act.
	_, err := inv.MergePrefixes(incoming)

	// Assert.
	want := fmt.Sprintf(`prefix 10.2.0.0/24 in namespace "Global": source IDs name 2 existing records ([%s %s])`, byNautobot, byCSM)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("MergePrefixes error = %v, want containing %q", err, want)
	}
	if len(inv.Prefixes) != 2 || inv.Prefixes[byNautobot].Prefix != "10.0.0.0/24" || inv.Prefixes[byCSM].Prefix != "10.1.0.0/24" {
		t.Errorf("collection changed on error: %d prefixes, %q and %q", len(inv.Prefixes), inv.Prefixes[byNautobot].Prefix, inv.Prefixes[byCSM].Prefix)
	}
}

// TestMergePrefixesFollowsTwoSourcesOnOneRecord verifies an incoming object
// whose provider IDs all point at the same record takes that record over.
//
// Why it matters: a record imported from two providers legitimately carries
// both IDs; only IDs that disagree about which record they name are an error.
// Inputs: 10.0.0.0/24 known to nautobot as X and to csm as Y; incoming
// 10.2.0.0/24 carrying both. Outputs: the record keeps its UUID and takes the
// new CIDR; one record in total.
// Data choice: the CIDR changes so the match can only have come from the IDs.
func TestMergePrefixesFollowsTwoSourcesOnOneRecord(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	record, x, y := uuid.New(), uuid.New(), uuid.New()
	inv.Prefixes[record] = &CaniPrefix{ID: record, Prefix: "10.0.0.0/24",
		ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": x, "csm": y}}}
	incomingID := uuid.New()
	incoming := map[uuid.UUID]*CaniPrefix{incomingID: {ID: incomingID, Prefix: "10.2.0.0/24",
		ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": x, "csm": y}}}}

	// Act.
	remap, err := inv.MergePrefixes(incoming)

	// Assert.
	if err != nil {
		t.Fatalf("MergePrefixes returned error: %v", err)
	}
	if remap[incomingID] != record || len(inv.Prefixes) != 1 || inv.Prefixes[record].Prefix != "10.2.0.0/24" {
		t.Errorf("remap = %v with %d records, prefix %q; want %v, 1 record at 10.2.0.0/24",
			remap, len(inv.Prefixes), inv.Prefixes[record].Prefix, record)
	}
}

// reimportFixture returns n addresses already in an inventory and the same n
// as a Nautobot re-import delivers them: fresh UUIDs, same hosts and source IDs.
func reimportFixture(n int) (*Inventory, map[uuid.UUID]*CaniIPAddress) {
	inv := NewInventory()
	incoming := make(map[uuid.UUID]*CaniIPAddress, n)
	for i := 0; i < n; i++ {
		host := fmt.Sprintf("10.%d.%d.%d", i/65536, (i/256)%256, i%256)
		source := map[string]uuid.UUID{"nautobot": uuid.New()}
		existingID, incomingID := uuid.New(), uuid.New()
		inv.IPAddresses[existingID] = &CaniIPAddress{ID: existingID, Host: host, Address: host + "/8",
			ObjectMeta: ObjectMeta{ExternalIDs: source}}
		incoming[incomingID] = &CaniIPAddress{ID: incomingID, Host: host, Address: host + "/8",
			ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": source["nautobot"]}}}
	}
	return inv, incoming
}

// mergeAllocatedBytes reports the bytes allocated while merging a re-import of
// n addresses.
func mergeAllocatedBytes(t *testing.T, n int) uint64 {
	t.Helper()
	inv, incoming := reimportFixture(n)
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	if _, err := inv.MergeIPAddresses(incoming); err != nil {
		t.Fatalf("MergeIPAddresses(%d addresses): %v", n, err)
	}
	runtime.ReadMemStats(&after)
	return after.TotalAlloc - before.TotalAlloc
}

// TestMergeIPAddressesReimportScalesLinearly verifies re-importing four times
// as many addresses allocates far less than sixteen times the memory.
//
// Why it matters: the Nautobot import mints fresh UUIDs on every run, so every
// address takes the identity lookup; re-sorting every record per address made
// a 4,000-address re-import allocate 263 MB and take 2.6 seconds.
// Inputs: re-imports of 1,000 and 4,000 addresses that all match by source
// ID. Outputs: the larger merge allocates under 8x the smaller one.
// Data choice: 4x the input separates linear growth (about 4x) from the old
// quadratic growth (about 16x) with room for noise.
func TestMergeIPAddressesReimportScalesLinearly(t *testing.T) {
	small, large := mergeAllocatedBytes(t, 1000), mergeAllocatedBytes(t, 4000)

	if large > 8*small {
		t.Errorf("allocated %d bytes for 4,000 addresses and %d for 1,000 (%.1fx), want under 8x",
			large, small, float64(large)/float64(small))
	}
}

func BenchmarkMergeIPAddressesReimport(b *testing.B) {
	for _, n := range []int{1000, 4000} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				inv, incoming := reimportFixture(n)
				b.StartTimer()
				if _, err := inv.MergeIPAddresses(incoming); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
