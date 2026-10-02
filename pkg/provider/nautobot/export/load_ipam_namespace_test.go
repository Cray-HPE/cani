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
	"strconv"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// namespaceServer answers /ipam/namespaces/ lookups with a stable ID per name
// and records every other request's path and query for assertions.
type namespaceServer struct {
	ids      map[string]uuid.UUID
	requests []string
	posts    map[string][][]byte
}

func newNamespaceServer() *namespaceServer {
	return &namespaceServer{ids: map[string]uuid.UUID{}, posts: map[string][][]byte{}}
}

func (s *namespaceServer) id(name string) uuid.UUID {
	if id, ok := s.ids[name]; ok {
		return id
	}
	s.ids[name] = uuid.New()
	return s.ids[name]
}

// handler serves namespaces by name, records POST bodies by path, returns a
// fresh ID for every create, and answers all other GETs with an empty list.
func (s *namespaceServer) handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s.requests = append(s.requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
	switch {
	case strings.Contains(r.URL.Path, "ipam/namespaces"):
		name := r.URL.Query().Get("name")
		_, _ = io.WriteString(w, `{"count":1,"results":[{"id":"`+s.id(name).String()+`","name":"`+name+`","display":"`+name+`"}]}`)
	case r.Method == http.MethodPost:
		body, _ := io.ReadAll(r.Body)
		s.posts[r.URL.Path] = append(s.posts[r.URL.Path], body)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id":"`+uuid.New().String()+`"}`)
	default:
		_, _ = io.WriteString(w, emptyListJSON)
	}
}

func (s *namespaceServer) requestsMatching(substr string) []string {
	var matches []string
	for _, request := range s.requests {
		if strings.Contains(request, substr) {
			matches = append(matches, request)
		}
	}
	return matches
}

// TestLoadPrefixes_ScopesEachPrefixByItsNamespace verifies the same CIDR in two
// namespaces is looked up and created inside each namespace rather than once
// under Global.
//
// Why it matters: the pre-namespace exporter resolved Global once and keyed
// the lookup cache by bare CIDR, so the second tenant's prefix was reported as
// "already exists" and never created.
// Inputs: 10.0.0.0/24 in "tenant-a" and in "tenant-b"; a server that answers
// namespace lookups by name and finds no existing prefixes. Outputs: two
// prefix GETs each carrying its namespace filter, two POSTs whose namespace
// IDs differ and match the names, PrefixesCreated == 2.
// Data choice: identical CIDR text is the only way to prove the namespace is
// doing the separating.
func TestLoadPrefixes_ScopesEachPrefixByItsNamespace(t *testing.T) {
	server := newNamespaceServer()
	e, cleanup := newExporterWithServer(t, server.handler)
	defer cleanup()
	seedActiveStatus(t, e)

	inv := devicetypes.NewInventory()
	for _, namespace := range []string{"tenant-a", "tenant-b"} {
		id := uuid.New()
		inv.Prefixes[id] = &devicetypes.CaniPrefix{ID: id, Prefix: "10.0.0.0/24", PrefixLen: 24, Namespace: namespace}
	}
	result := &LoadResult{}

	created, err := e.loadPrefixes(context.Background(), inv, nil, nil, result)
	if err != nil {
		t.Fatalf("loadPrefixes returned error: %v", err)
	}

	if result.PrefixesCreated != 2 || len(created) != 2 || len(result.Errors) != 0 {
		t.Fatalf("created = %d, mapped = %d, errors = %v; want 2, 2, none", result.PrefixesCreated, len(created), result.Errors)
	}
	lookups := server.requestsMatching("GET /ipam/prefixes/")
	if len(lookups) != 2 {
		t.Fatalf("prefix lookups = %v, want 2 namespace-scoped GETs", lookups)
	}
	for _, lookup := range lookups {
		if !strings.Contains(lookup, "namespace=tenant-") {
			t.Errorf("prefix lookup %q lacks a namespace filter", lookup)
		}
	}
	seen := map[string]bool{}
	for _, body := range server.posts["/ipam/prefixes/"] {
		sent := decodeSentPrefix(t, body)
		seen[sent.Namespace.ID] = true
	}
	if !seen[server.id("tenant-a").String()] || !seen[server.id("tenant-b").String()] {
		t.Errorf("POSTed namespace IDs = %v, want both tenant IDs", seen)
	}
}

// TestLoadPrefixes_CreatesVRFPrefixAssignmentsIdempotently verifies a prefix's
// VRF memberships become VRF-prefix assignment records, that an existing
// assignment is skipped, and that an unresolved legacy VRF name is reported.
//
// Why it matters: prefix-to-VRF associations were never exported before, and a
// repeated export must not duplicate the join records.
// Inputs: a Global VRF "blue" pre-cached as exported; a prefix the server
// holds that is a member of blue; a second prefix carrying only a legacy VRF
// name. The server lists the assignments it has been sent, so the first run
// sees none and a second run, made with a fresh cache as a second export
// invocation would be, sees the one the first created. Outputs: one
// assignment POST referencing the VRF and prefix IDs, one skipped on re-run,
// and an error naming the legacy VRF name.
// Data choice: re-running against a server that now reports the assignment is
// the idempotency proof the ticket asks for; a fresh cache makes the server,
// not this run's memory, the thing that proves it.
func TestLoadPrefixes_CreatesVRFPrefixAssignmentsIdempotently(t *testing.T) {
	vrfNID, blue := uuid.New(), uuid.New()
	remotePrefix := uuid.New()
	server := newNamespaceServer()
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "vrf-prefix-assignments"):
			w.Header().Set("Content-Type", "application/json")
			var listed []string
			for _, body := range server.posts["/ipam/vrf-prefix-assignments/"] {
				var sent struct {
					Prefix wireIDRef `json:"prefix"`
				}
				_ = json.Unmarshal(body, &sent)
				listed = append(listed, assignmentJSON(uuid.MustParse(sent.Prefix.ID)))
			}
			_, _ = io.WriteString(w, `{"count":`+strconv.Itoa(len(listed))+`,"results":[`+strings.Join(listed, ",")+`]}`)
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "ipam/prefixes") && r.URL.Query().Get("prefix") == "10.0.0.0/24":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, remotePrefixJSON(remotePrefix, "10.0.0.0/24"))
		default:
			server.handler(w, r)
		}
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	seedActiveStatus(t, e)
	e.Cache.CacheVRF("Global", "blue", &CachedItem{ID: vrfNID, Name: "blue"})

	inv := devicetypes.NewInventory()
	inv.VRFs[blue] = &devicetypes.CaniVRF{ID: blue, Name: "blue"}
	member, legacy := uuid.New(), uuid.New()
	inv.Prefixes[member] = &devicetypes.CaniPrefix{ID: member, Prefix: "10.0.0.0/24", PrefixLen: 24, VRFs: []uuid.UUID{blue}}
	inv.Prefixes[legacy] = &devicetypes.CaniPrefix{ID: legacy, Prefix: "10.1.0.0/24", PrefixLen: 24, VRF: "stale"}
	result := &LoadResult{}

	created, err := e.loadPrefixes(context.Background(), inv, nil, nil, result)
	if err != nil {
		t.Fatalf("loadPrefixes returned error: %v", err)
	}

	posts := server.posts["/ipam/vrf-prefix-assignments/"]
	if len(posts) != 1 || result.VRFPrefixAssignmentsCreated != 1 {
		t.Fatalf("assignment POSTs = %d, created = %d; want 1 and 1", len(posts), result.VRFPrefixAssignmentsCreated)
	}
	var sent struct {
		Vrf    wireIDRef `json:"vrf"`
		Prefix wireIDRef `json:"prefix"`
	}
	if err := json.Unmarshal(posts[0], &sent); err != nil {
		t.Fatalf("decoding assignment: %v", err)
	}
	if sent.Vrf.ID != vrfNID.String() || sent.Prefix.ID != created[member].String() {
		t.Errorf("assignment = vrf %s prefix %s, want vrf %s prefix %s", sent.Vrf.ID, sent.Prefix.ID, vrfNID, created[member])
	}
	if len(result.Errors) != 1 || !strings.Contains(result.Errors[0], `legacy VRF name "stale"`) {
		t.Errorf("errors = %v, want one legacy VRF name report", result.Errors)
	}

	// Second run with a fresh cache: the server now lists the assignment the
	// first run created, and the member prefix is found rather than created.
	second, cleanupSecond := newExporterWithServer(t, handler)
	defer cleanupSecond()
	seedActiveStatus(t, second)
	second.Cache.CacheVRF("Global", "blue", &CachedItem{ID: vrfNID, Name: "blue"})
	rerun := &LoadResult{}
	if _, err := second.loadPrefixes(context.Background(), inv, nil, nil, rerun); err != nil {
		t.Fatalf("second loadPrefixes returned error: %v", err)
	}
	if rerun.VRFPrefixAssignmentsSkipped != 1 || rerun.VRFPrefixAssignmentsCreated != 0 {
		t.Errorf("rerun skipped = %d, created = %d; want 1 and 0", rerun.VRFPrefixAssignmentsSkipped, rerun.VRFPrefixAssignmentsCreated)
	}
	if len(server.posts["/ipam/vrf-prefix-assignments/"]) != 1 {
		t.Errorf("assignment POSTs after rerun = %d, want still 1", len(server.posts["/ipam/vrf-prefix-assignments/"]))
	}
}

// TestLoadIPAddresses_UsesParentNamespaceAndRejectsOrphans verifies an address
// is filed in its parent prefix's namespace and an address without a
// containing prefix is reported instead of exported under Global.
//
// Why it matters: Nautobot derives an address's namespace from its parent;
// sending Global for a tenant address would seed the wrong network, and an
// orphan draft must not look like a valid seed.
// Inputs: a tenant-a prefix with a parented 10.0.0.5/24 and an orphan
// 192.168.0.1/24 with no parent. Outputs: one IP POST whose namespace is the
// tenant-a ID and whose lookup was filtered by host and namespace, plus one
// error for the orphan.
// Data choice: the orphan carries explicit namespace intent so the error can
// name the scope the operator needs to add a prefix to.
func TestLoadIPAddresses_UsesParentNamespaceAndRejectsOrphans(t *testing.T) {
	server := newNamespaceServer()
	e, cleanup := newExporterWithServer(t, server.handler)
	defer cleanup()
	seedActiveStatus(t, e)

	inv := devicetypes.NewInventory()
	prefixID, parented, orphan := uuid.New(), uuid.New(), uuid.New()
	inv.Prefixes[prefixID] = &devicetypes.CaniPrefix{ID: prefixID, Prefix: "10.0.0.0/24", Namespace: "tenant-a"}
	inv.IPAddresses[parented] = &devicetypes.CaniIPAddress{ID: parented, Host: "10.0.0.5", Address: "10.0.0.5/24", Parent: prefixID}
	inv.IPAddresses[orphan] = &devicetypes.CaniIPAddress{ID: orphan, Host: "192.168.0.1", Address: "192.168.0.1/24", Namespace: "tenant-b"}
	result := &LoadResult{}

	if err := e.loadIPAddresses(context.Background(), inv, map[uuid.UUID]uuid.UUID{prefixID: uuid.New()}, nil, result); err != nil {
		t.Fatalf("loadIPAddresses returned error: %v", err)
	}

	posts := server.posts["/ipam/ip-addresses/"]
	if len(posts) != 1 || result.IPAddressesCreated != 1 {
		t.Fatalf("IP POSTs = %d, created = %d; want 1 and 1", len(posts), result.IPAddressesCreated)
	}
	if ip := decodeSentIP(t, posts[0]); ip.Namespace.ID != server.id("tenant-a").String() {
		t.Errorf("namespace.id = %s, want tenant-a %s", ip.Namespace.ID, server.id("tenant-a"))
	}
	lookups := server.requestsMatching("GET /ipam/ip-addresses/")
	if len(lookups) != 1 {
		t.Fatalf("IP lookups = %v, want exactly one", lookups)
	}
	query := lookups[0][strings.IndexByte(lookups[0], '?')+1:]
	if !strings.Contains(query, "address=10.0.0.5") || strings.Contains(query, "address=10.0.0.5%2F") ||
		!strings.Contains(query, "namespace=tenant-a") {
		t.Errorf("IP lookup query = %q, want bare host and namespace filters", query)
	}
	if len(result.Errors) != 1 || !strings.Contains(result.Errors[0], `no containing prefix in namespace "tenant-b"`) {
		t.Errorf("errors = %v, want one orphan report naming tenant-b", result.Errors)
	}
}

// TestLoadPrefixes_FailsThePrefixWhenItsNamespaceCannotBeResolved verifies a
// namespace lookup failure is reported against the prefix and stops its
// export instead of letting Nautobot file it under Global.
//
// Why it matters: the pre-namespace exporter aborted the whole phase on a
// namespace failure, and a create sent without a namespace would land in
// Global, which is the silent redirect this change removes.
// Inputs: one prefix in "tenant-a" against a server that answers 500 to the
// namespace lookup. Outputs: one error naming the prefix and the namespace
// resolution, PrefixesCreated == 0, and no POST.
// Data choice: a 500 is a failure distinct from "not found", so it cannot be
// mistaken for a namespace to create.
func TestLoadPrefixes_FailsThePrefixWhenItsNamespaceCannotBeResolved(t *testing.T) {
	// Arrange.
	var writes []string
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method != http.MethodGet:
			writes = append(writes, r.Method+" "+r.URL.Path)
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"id":"`+uuid.New().String()+`"}`)
		case strings.Contains(r.URL.Path, "ipam/namespaces"):
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = io.WriteString(w, `{"detail":"boom"}`)
		default:
			_, _ = io.WriteString(w, emptyListJSON)
		}
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	seedActiveStatus(t, e)
	inv := devicetypes.NewInventory()
	prefixID := uuid.New()
	inv.Prefixes[prefixID] = &devicetypes.CaniPrefix{ID: prefixID, Prefix: "10.0.0.0/24", PrefixLen: 24, Namespace: "tenant-a"}
	result := &LoadResult{}

	// Act.
	created, err := e.loadPrefixes(context.Background(), inv, nil, nil, result)

	// Assert.
	if err != nil {
		t.Fatalf("loadPrefixes returned error: %v", err)
	}
	if len(result.Errors) != 1 || !strings.Contains(result.Errors[0], "prefix tenant-a/10.0.0.0/24: resolve namespace") {
		t.Errorf("errors = %v, want one namespace resolution error naming the prefix", result.Errors)
	}
	if result.PrefixesCreated != 0 || len(created) != 0 || len(writes) != 0 {
		t.Errorf("created = %d, mapped = %d, writes = %v; want nothing exported", result.PrefixesCreated, len(created), writes)
	}
}
