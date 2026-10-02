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
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// memberPrefixInventory holds VRF "blue" and a Global prefix that is a member
// of it.
func memberPrefixInventory() (*devicetypes.Inventory, uuid.UUID) {
	inv := devicetypes.NewInventory()
	vrfID, prefixID := uuid.New(), uuid.New()
	inv.VRFs[vrfID] = &devicetypes.CaniVRF{ID: vrfID, Name: "blue"}
	inv.Prefixes[prefixID] = &devicetypes.CaniPrefix{ID: prefixID, Prefix: "10.0.0.0/24", PrefixLen: 24, VRFs: []uuid.UUID{vrfID}}
	return inv, prefixID
}

// assignmentJSON renders one VRF-prefix assignment the way Nautobot lists it,
// with the prefix as a nested reference.
func assignmentJSON(prefixID uuid.UUID) string {
	return `{"id":"` + uuid.New().String() + `","prefix":{"id":"` + prefixID.String() + `"}}`
}

// remotePrefixJSON renders a prefix list holding one prefix with the given ID.
func remotePrefixJSON(id uuid.UUID, cidr string) string {
	return `{"count":1,"results":[{"id":"` + id.String() + `","prefix":"` + cidr + `","display":"` + cidr + `"}]}`
}

// TestLoadPrefixes_DryRunReportsAnExistingVRFAssignmentAsSkipped verifies a
// dry run against a prefix and VRF that both exist in Nautobot asks whether
// their assignment exists and reports it as skipped, not as a planned create.
//
// Why it matters: the dry-run summary is the operator's preview; counting an
// assignment Nautobot already holds as "would create" misstates the run.
// Inputs: VRF "blue" cached as found, a prefix the server reports as existing,
// and a server that lists that prefix among the VRF's assignments; DryRun on.
// Outputs: VRFPrefixAssignmentsSkipped == 1, no creates, and no write request.
// Data choice: an existing prefix is the only case in which the assignment can
// already exist, so it is the only case the read-only check is worth.
func TestLoadPrefixes_DryRunReportsAnExistingVRFAssignmentAsSkipped(t *testing.T) {
	// Arrange.
	server := newNamespaceServer()
	remotePrefix := uuid.New()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "vrf-prefix-assignments"):
			_, _ = io.WriteString(w, `{"count":1,"results":[`+assignmentJSON(remotePrefix)+`]}`)
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "ipam/prefixes"):
			_, _ = io.WriteString(w, remotePrefixJSON(remotePrefix, "10.0.0.0/24"))
		default:
			server.handler(w, r)
		}
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	seedActiveStatus(t, e)
	e.Options.DryRun = true
	e.Cache.SetDryRun(true)
	e.Cache.CacheVRF("Global", "blue", &CachedItem{ID: uuid.New(), Name: "blue"})
	inv, _ := memberPrefixInventory()
	result := &LoadResult{}

	// Act.
	if _, err := e.loadPrefixes(context.Background(), inv, nil, nil, result); err != nil {
		t.Fatalf("loadPrefixes returned error: %v", err)
	}

	// Assert.
	if result.VRFPrefixAssignmentsSkipped != 1 || result.VRFPrefixAssignmentsCreated != 0 || len(result.Errors) != 0 {
		t.Errorf("skipped = %d, created = %d, errors = %v; want 1, 0, none",
			result.VRFPrefixAssignmentsSkipped, result.VRFPrefixAssignmentsCreated, result.Errors)
	}
	if writes := server.requestsMatching("POST "); len(writes) != 0 {
		t.Errorf("dry run wrote to Nautobot: %v", writes)
	}
}

// TestLoadPrefixes_DryRunPlansTheAssignmentOfAPlannedPrefix verifies a dry run
// counts the assignment of a prefix it would create as planned without asking
// Nautobot about it.
//
// Why it matters: nothing can be assigned to a prefix that does not exist yet,
// and Nautobot answers 400 to an assignment filter naming an unknown prefix,
// so the read-only check must be skipped for a planned end.
// Inputs: VRF "blue" cached as found and a prefix the server does not hold;
// DryRun on. Outputs: VRFPrefixAssignmentsCreated == 1, no assignment lookup,
// and no write request.
// Data choice: a found VRF isolates the prefix as the planned end.
func TestLoadPrefixes_DryRunPlansTheAssignmentOfAPlannedPrefix(t *testing.T) {
	// Arrange.
	server := newNamespaceServer()
	e, cleanup := newExporterWithServer(t, server.handler)
	defer cleanup()
	seedActiveStatus(t, e)
	e.Options.DryRun = true
	e.Cache.SetDryRun(true)
	e.Cache.CacheVRF("Global", "blue", &CachedItem{ID: uuid.New(), Name: "blue"})
	inv, _ := memberPrefixInventory()
	result := &LoadResult{}

	// Act.
	if _, err := e.loadPrefixes(context.Background(), inv, nil, nil, result); err != nil {
		t.Fatalf("loadPrefixes returned error: %v", err)
	}

	// Assert.
	if result.VRFPrefixAssignmentsCreated != 1 || result.VRFPrefixAssignmentsSkipped != 0 || len(result.Errors) != 0 {
		t.Errorf("created = %d, skipped = %d, errors = %v; want 1, 0, none",
			result.VRFPrefixAssignmentsCreated, result.VRFPrefixAssignmentsSkipped, result.Errors)
	}
	if lookups := server.requestsMatching("vrf-prefix-assignments"); len(lookups) != 0 {
		t.Errorf("assignment lookups = %v, want none for a planned prefix", lookups)
	}
	if writes := server.requestsMatching("POST "); len(writes) != 0 {
		t.Errorf("dry run wrote to Nautobot: %v", writes)
	}
}

// TestLoadPrefixes_ListsAVRFsAssignmentsOncePerExport verifies the existing
// assignments of a VRF are read once, through every page, and that each
// member prefix is then decided from that listing.
//
// Why it matters: one lookup per (VRF, prefix) pair made a large export pay a
// round trip per membership; one paginated listing per VRF replaces them, and
// a page boundary must not hide an assignment on the next page, even when the
// server returns fewer results per page than were requested.
// Inputs: VRF "blue" cached as found; three member prefixes the server holds;
// a server that lists blue's assignments as two one-result pages,
// 10.0.1.0/24 on the first and 10.0.2.0/24 on the second. Outputs: exactly
// two listing requests, both filtered by the VRF alone, the second starting
// at offset 1; two memberships skipped; one POST for 10.0.3.0/24.
// Data choice: a one-result first page against a 100-result request shows
// the client advances by what it received; the assignment on the second page
// is the one a listing that stopped at the first page would re-create.
func TestLoadPrefixes_ListsAVRFsAssignmentsOncePerExport(t *testing.T) {
	// Arrange.
	server := newNamespaceServer()
	vrfNID := uuid.New()
	remote := map[string]uuid.UUID{"10.0.1.0/24": uuid.New(), "10.0.2.0/24": uuid.New(), "10.0.3.0/24": uuid.New()}
	var listings []string
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "vrf-prefix-assignments"):
			listings = append(listings, r.URL.RawQuery)
			if r.URL.Query().Get("offset") == "0" {
				_, _ = io.WriteString(w, `{"count":2,"next":"http://nautobot/api/ipam/vrf-prefix-assignments/?offset=1","results":[`+assignmentJSON(remote["10.0.1.0/24"])+`]}`)
				return
			}
			if r.URL.Query().Get("offset") != "1" {
				t.Errorf("second listing offset = %q, want 1, the number of results the first page returned", r.URL.Query().Get("offset"))
			}
			_, _ = io.WriteString(w, `{"count":2,"results":[`+assignmentJSON(remote["10.0.2.0/24"])+`]}`)
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "ipam/prefixes"):
			cidr := r.URL.Query().Get("prefix")
			_, _ = io.WriteString(w, remotePrefixJSON(remote[cidr], cidr))
		default:
			server.handler(w, r)
		}
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	seedActiveStatus(t, e)
	e.Cache.CacheVRF("Global", "blue", &CachedItem{ID: vrfNID, Name: "blue"})
	inv := devicetypes.NewInventory()
	blue := uuid.New()
	inv.VRFs[blue] = &devicetypes.CaniVRF{ID: blue, Name: "blue"}
	for cidr := range remote {
		id := uuid.New()
		inv.Prefixes[id] = &devicetypes.CaniPrefix{ID: id, Prefix: cidr, PrefixLen: 24, VRFs: []uuid.UUID{blue}}
	}
	result := &LoadResult{}

	// Act.
	if _, err := e.loadPrefixes(context.Background(), inv, nil, nil, result); err != nil {
		t.Fatalf("loadPrefixes returned error: %v", err)
	}

	// Assert.
	if len(listings) != 2 {
		t.Fatalf("assignment listings = %v, want one request per page of the VRF", listings)
	}
	for _, query := range listings {
		if !strings.Contains(query, "vrf="+vrfNID.String()) || strings.Contains(query, "prefix=") {
			t.Errorf("listing %q should filter by the VRF alone", query)
		}
	}
	if result.VRFPrefixAssignmentsSkipped != 2 || result.VRFPrefixAssignmentsCreated != 1 || len(result.Errors) != 0 {
		t.Errorf("skipped = %d, created = %d, errors = %v; want 2, 1, none",
			result.VRFPrefixAssignmentsSkipped, result.VRFPrefixAssignmentsCreated, result.Errors)
	}
	posts := server.posts["/ipam/vrf-prefix-assignments/"]
	if len(posts) != 1 {
		t.Fatalf("assignment POSTs = %d, want 1", len(posts))
	}
	var sent struct {
		Prefix wireIDRef `json:"prefix"`
	}
	if err := json.Unmarshal(posts[0], &sent); err != nil {
		t.Fatalf("decoding assignment: %v", err)
	}
	if sent.Prefix.ID != remote["10.0.3.0/24"].String() {
		t.Errorf("assigned prefix = %s, want the one the listing did not hold (%s)", sent.Prefix.ID, remote["10.0.3.0/24"])
	}
}

// TestLoadPrefixes_AssignsAVRFWhoseNameContainsASlash verifies a prefix's
// membership in a VRF named with a slash is exported.
//
// Why it matters: the loader holds the VRF record, so it must not go through
// the "namespace/name" reference grammar, which would read "Global/blue/green"
// as VRF "green" in namespace "Global/blue" and report the VRF as not exported.
// Inputs: VRF "blue/green" in Global cached as found; a prefix that is a member
// of it; a server holding no assignments. Outputs: one assignment POST naming
// the VRF's ID and no error.
// Data choice: a slash in the name is the one character the reference grammar
// gives meaning to.
func TestLoadPrefixes_AssignsAVRFWhoseNameContainsASlash(t *testing.T) {
	// Arrange.
	server := newNamespaceServer()
	e, cleanup := newExporterWithServer(t, server.handler)
	defer cleanup()
	seedActiveStatus(t, e)
	vrfNID := uuid.New()
	e.Cache.CacheVRF("Global", "blue/green", &CachedItem{ID: vrfNID, Name: "blue/green"})
	inv := devicetypes.NewInventory()
	vrfID, prefixID := uuid.New(), uuid.New()
	inv.VRFs[vrfID] = &devicetypes.CaniVRF{ID: vrfID, Name: "blue/green"}
	inv.Prefixes[prefixID] = &devicetypes.CaniPrefix{ID: prefixID, Prefix: "10.0.0.0/24", PrefixLen: 24, VRFs: []uuid.UUID{vrfID}}
	result := &LoadResult{}

	// Act.
	if _, err := e.loadPrefixes(context.Background(), inv, nil, nil, result); err != nil {
		t.Fatalf("loadPrefixes returned error: %v", err)
	}

	// Assert.
	if len(result.Errors) != 0 || result.VRFPrefixAssignmentsCreated != 1 {
		t.Fatalf("created = %d, errors = %v; want 1 and none", result.VRFPrefixAssignmentsCreated, result.Errors)
	}
	posts := server.posts["/ipam/vrf-prefix-assignments/"]
	if len(posts) != 1 {
		t.Fatalf("assignment POSTs = %d, want 1", len(posts))
	}
	var sent struct {
		Vrf wireIDRef `json:"vrf"`
	}
	if err := json.Unmarshal(posts[0], &sent); err != nil {
		t.Fatalf("decoding assignment: %v", err)
	}
	if sent.Vrf.ID != vrfNID.String() {
		t.Errorf("assignment VRF = %s, want %s", sent.Vrf.ID, vrfNID)
	}
}

// TestLoadPrefixes_AssignsADuplicatedMembershipOnce verifies a VRF listed
// twice in a prefix's memberships yields one assignment.
//
// Why it matters: Nautobot holds one assignment per (VRF, prefix) and rejects
// a second; a link this run created must count as existing for the rest of
// the run, not only after the next listing.
// Inputs: VRF "blue" cached as found and a prefix whose VRFs name it twice;
// a server holding no assignments. Outputs: one POST, one skipped, no error.
// Data choice: the duplicate is the only way one run asks about the same pair
// twice.
func TestLoadPrefixes_AssignsADuplicatedMembershipOnce(t *testing.T) {
	// Arrange.
	server := newNamespaceServer()
	e, cleanup := newExporterWithServer(t, server.handler)
	defer cleanup()
	seedActiveStatus(t, e)
	e.Cache.CacheVRF("Global", "blue", &CachedItem{ID: uuid.New(), Name: "blue"})
	inv, prefixID := memberPrefixInventory()
	inv.Prefixes[prefixID].VRFs = append(inv.Prefixes[prefixID].VRFs, inv.Prefixes[prefixID].VRFs[0])
	result := &LoadResult{}

	// Act.
	if _, err := e.loadPrefixes(context.Background(), inv, nil, nil, result); err != nil {
		t.Fatalf("loadPrefixes returned error: %v", err)
	}

	// Assert.
	if posts := server.posts["/ipam/vrf-prefix-assignments/"]; len(posts) != 1 {
		t.Errorf("assignment POSTs = %d, want 1", len(posts))
	}
	if result.VRFPrefixAssignmentsCreated != 1 || result.VRFPrefixAssignmentsSkipped != 1 || len(result.Errors) != 0 {
		t.Errorf("created = %d, skipped = %d, errors = %v; want 1, 1, none",
			result.VRFPrefixAssignmentsCreated, result.VRFPrefixAssignmentsSkipped, result.Errors)
	}
}
