/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// These tests exercise the Nautobot IPAM export paths (Phases 7-9) end-to-end
// against an httptest server. They assert the exact JSON payload the exporter
// PUTs on the wire so the test fails if the mapping to the generated Nautobot
// bindings (request field names, FK reference shape) ever drifts. This is the
// "how the data should look after export" contract.

// -----------------------------------------------------------------------------
// Shared helpers (defined here, reused by the other IPAM load_* test files in
// this package).
// -----------------------------------------------------------------------------

// wireIDRef decodes a Nautobot foreign-key reference as it appears on the wire.
// Every FK helper (makeIDRef / makeTenantRef / makeIPParentRef / ...) marshals
// to the shape {"id":"<uuid>"}, so this single struct covers all of them.
type wireIDRef struct {
	ID string `json:"id"`
}

// resetIPAMCaches clears the package-level IPAM lookup caches. They are global
// (shared across LookupCache instances), so each HTTP test starts from a known
// empty state to remain independent and order-insensitive.
func resetIPAMCaches() {
	namespacesMu.Lock()
	for k := range namespaces {
		delete(namespaces, k)
	}
	namespacesMu.Unlock()

	vlansMu.Lock()
	for k := range vlans {
		delete(vlans, k)
	}
	vlansMu.Unlock()

	prefixesMu.Lock()
	for k := range prefixes {
		delete(prefixes, k)
	}
	prefixesMu.Unlock()

	ipAddressesMu.Lock()
	for k := range ipAddresses {
		delete(ipAddresses, k)
	}
	ipAddressesMu.Unlock()
}

// seedActiveStatus pre-populates the "Active" status in the cache so the
// create paths resolve it without an extra HTTP round-trip. Returns the
// Nautobot status ID the exporter is expected to reference.
func seedActiveStatus(t *testing.T, e *Exporter) uuid.UUID {
	t.Helper()
	id := uuid.New()
	e.Cache.statuses["Active"] = &CachedItem{ID: id, Name: "Active"}
	return id
}

// seedGlobalNamespace pre-populates the "Global" namespace so prefix/IP loads
// skip the namespace resolve round-trip.
func seedGlobalNamespace(id uuid.UUID) {
	namespacesMu.Lock()
	namespaces["Global"] = &CachedItem{ID: id, Name: "Global"}
	namespacesMu.Unlock()
}

// emptyListJSON is the canonical "no results" body for any best-effort lookup
// the exporter performs that the test does not care about.
const emptyListJSON = `{"count":0,"results":[]}`

// sentPrefix is the minimal view of the WritablePrefixRequest body the exporter
// POSTs to /ipam/prefixes/.
type sentPrefix struct {
	Prefix      string     `json:"prefix"`
	Status      wireIDRef  `json:"status"`
	Namespace   wireIDRef  `json:"namespace"`
	Parent      *wireIDRef `json:"parent"`
	Type        string     `json:"type"`
	Description string     `json:"description"`
}

// -----------------------------------------------------------------------------
// loadPrefixes — happy path (PASSING scenario)
// -----------------------------------------------------------------------------

// TestLoadPrefixes_SendsCorrectPayloadAndResolvesParent verifies prefixes are
// created shortest-mask-first so a child can reference the freshly-assigned
// Nautobot ID of its parent, and that every field maps onto the generated
// request type exactly.
//
// Why it matters: Nautobot models prefix containment by parent ID, so a wrong
// order or a mis-threaded in-run ID map would orphan child prefixes or mis-nest
// the IPAM tree.
// Inputs: a /16 container and a nested /24 network. Outputs: two POSTs in mask
// order, a cani-ID→Nautobot-ID map, and external IDs stamped back.
// Data choice: a container+network pair is the minimal fixture that proves both
// ordering and parent-ID threading.
func TestLoadPrefixes_SendsCorrectPayloadAndResolvesParent(t *testing.T) {
	// Arrange.
	resetIPAMCaches()

	parentNID := uuid.New() // Nautobot ID returned for the /16 container
	childNID := uuid.New()  // Nautobot ID returned for the /24 network
	createdIDs := []uuid.UUID{parentNID, childNID}
	createIdx := 0

	var prefixPosts [][]byte
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "ipam/prefixes"):
			body, _ := io.ReadAll(r.Body)
			prefixPosts = append(prefixPosts, body)
			id := createdIDs[createIdx]
			createIdx++
			w.WriteHeader(http.StatusCreated)
			_, _ = fmt.Fprintf(w, `{"id":%q}`, id.String())
		default:
			// Prefix lookup (GET) and any other best-effort call: not found.
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, emptyListJSON)
		}
	}

	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	statusID := seedActiveStatus(t, e)
	nsID := uuid.New()
	seedGlobalNamespace(nsID)

	parentID := uuid.New()
	childID := uuid.New()
	inv := &devicetypes.Inventory{
		Prefixes: map[uuid.UUID]*devicetypes.CaniPrefix{
			parentID: {
				ID:          parentID,
				Prefix:      "10.0.0.0/16",
				PrefixLen:   16,
				Type:        devicetypes.PrefixTypeContainer,
				Description: "site container",
				ObjectMeta:  devicetypes.ObjectMeta{Status: "Active"},
			},
			childID: {
				ID:         childID,
				Prefix:     "10.0.1.0/24",
				PrefixLen:  24,
				Type:       devicetypes.PrefixTypeNetwork,
				Parent:     parentID,
				ObjectMeta: devicetypes.ObjectMeta{Status: "Active"},
			},
		},
	}

	result := &LoadResult{}

	// Act.
	created, err := e.loadPrefixes(
		context.Background(),
		inv,
		map[uuid.UUID]uuid.UUID{}, // locationMap (unused: prefixes omit location)
		map[uuid.UUID]uuid.UUID{}, // vlanMap (no VLAN association)
		result,
	)

	// Assert.
	if err != nil {
		t.Fatalf("loadPrefixes returned error: %v", err)
	}
	if len(prefixPosts) != 2 {
		t.Fatalf("expected 2 prefix POSTs, got %d", len(prefixPosts))
	}

	// Parent is created first (shortest mask) and carries no parent reference.
	parentSent := decodeSentPrefix(t, prefixPosts[0])
	if parentSent.Prefix != "10.0.0.0/16" {
		t.Errorf("parent prefix = %q, want 10.0.0.0/16", parentSent.Prefix)
	}
	if parentSent.Status.ID != statusID.String() {
		t.Errorf("parent status.id = %q, want %q", parentSent.Status.ID, statusID)
	}
	if parentSent.Namespace.ID != nsID.String() {
		t.Errorf("parent namespace.id = %q, want %q", parentSent.Namespace.ID, nsID)
	}
	if parentSent.Type != string(nautobotapi.PrefixTypeChoicesContainer) {
		t.Errorf("parent type = %q, want %q", parentSent.Type, nautobotapi.PrefixTypeChoicesContainer)
	}
	if parentSent.Description != "site container" {
		t.Errorf("parent description = %q, want 'site container'", parentSent.Description)
	}
	if parentSent.Parent != nil {
		t.Errorf("parent prefix should not reference a parent, got %+v", parentSent.Parent)
	}

	// Child is created second and references the parent's *Nautobot* ID, proving
	// the in-run ID map (cani UUID -> Nautobot UUID) is threaded correctly.
	childSent := decodeSentPrefix(t, prefixPosts[1])
	if childSent.Prefix != "10.0.1.0/24" {
		t.Errorf("child prefix = %q, want 10.0.1.0/24", childSent.Prefix)
	}
	if childSent.Type != string(nautobotapi.PrefixTypeChoicesNetwork) {
		t.Errorf("child type = %q, want %q", childSent.Type, nautobotapi.PrefixTypeChoicesNetwork)
	}
	if childSent.Parent == nil {
		t.Fatal("child prefix must reference its parent")
	}
	if childSent.Parent.ID != parentNID.String() {
		t.Errorf("child parent.id = %q, want parent Nautobot ID %q", childSent.Parent.ID, parentNID)
	}

	// Return value maps cani IDs to the Nautobot IDs the server handed back.
	if created[parentID] != parentNID {
		t.Errorf("created[parent] = %s, want %s", created[parentID], parentNID)
	}
	if created[childID] != childNID {
		t.Errorf("created[child] = %s, want %s", created[childID], childNID)
	}
	if result.PrefixesCreated != 2 {
		t.Errorf("PrefixesCreated = %d, want 2", result.PrefixesCreated)
	}

	// External IDs are stamped back onto the inventory records for round-tripping.
	if got := inv.Prefixes[childID].ExternalIDs[externalIDKeyNautobot]; got != childNID {
		t.Errorf("child ExternalIDs[nautobot] = %s, want %s", got, childNID)
	}
}

// -----------------------------------------------------------------------------
// loadPrefixes — idempotency (skip already-existing)
// -----------------------------------------------------------------------------

// TestLoadPrefixes_SkipsExistingPrefix verifies that when a prefix already
// exists in Nautobot the exporter records it as skipped, reuses the remote ID,
// and issues no create POST.
//
// Why it matters: exports are re-run repeatedly, so the prefix phase must be
// idempotent — re-creating existing prefixes would duplicate IPAM objects or
// fail on uniqueness.
// Inputs: one CaniPrefix (172.16.0.0/24) whose GET lookup returns a match.
// Outputs: PrefixesSkipped == 1, PrefixesCreated == 0, and the returned map and
// external IDs point at the existing remote ID.
// Data choice: the lookup handler returns a populated results array to simulate
// a prior run having created the prefix.
func TestLoadPrefixes_SkipsExistingPrefix(t *testing.T) {
	// Arrange.
	resetIPAMCaches()

	existingNID := uuid.New()
	var createCalls int
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "ipam/prefixes"):
			createCalls++
			w.WriteHeader(http.StatusCreated)
			_, _ = fmt.Fprintf(w, `{"id":%q}`, uuid.New().String())
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "ipam/prefixes"):
			// Lookup reports the prefix already exists.
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w,
				`{"results":[{"id":%q,"prefix":"172.16.0.0/24","display":"172.16.0.0/24"}]}`,
				existingNID.String())
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, emptyListJSON)
		}
	}

	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	seedActiveStatus(t, e)
	seedGlobalNamespace(uuid.New())

	prefixID := uuid.New()
	inv := &devicetypes.Inventory{
		Prefixes: map[uuid.UUID]*devicetypes.CaniPrefix{
			prefixID: {ID: prefixID, Prefix: "172.16.0.0/24", PrefixLen: 24, ObjectMeta: devicetypes.ObjectMeta{Status: "Active"}},
		},
	}
	result := &LoadResult{}

	// Act.
	created, err := e.loadPrefixes(context.Background(), inv,
		map[uuid.UUID]uuid.UUID{}, map[uuid.UUID]uuid.UUID{}, result)

	// Assert.
	if err != nil {
		t.Fatalf("loadPrefixes returned error: %v", err)
	}
	if createCalls != 0 {
		t.Errorf("expected no create POSTs for an existing prefix, got %d", createCalls)
	}
	if result.PrefixesSkipped != 1 {
		t.Errorf("PrefixesSkipped = %d, want 1", result.PrefixesSkipped)
	}
	if result.PrefixesCreated != 0 {
		t.Errorf("PrefixesCreated = %d, want 0", result.PrefixesCreated)
	}
	if created[prefixID] != existingNID {
		t.Errorf("created[prefix] = %s, want existing remote ID %s", created[prefixID], existingNID)
	}
	if got := inv.Prefixes[prefixID].ExternalIDs[externalIDKeyNautobot]; got != existingNID {
		t.Errorf("ExternalIDs[nautobot] = %s, want %s", got, existingNID)
	}
}

// -----------------------------------------------------------------------------
// createPrefix — error path (the SUT-failure scenario)
// -----------------------------------------------------------------------------

// TestCreatePrefix_ReturnsErrorWhenStatusUnresolvable verifies the create path
// surfaces an error (and returns uuid.Nil) when the referenced status cannot be
// resolved and status auto-creation is disabled.
//
// Why it matters: a prefix without a valid status reference is invalid in
// Nautobot, so failing loudly beats silently POSTing a malformed object or
// fabricating a status.
// Inputs: a CaniPrefix with Status "Active" but no seeded status and an empty
// status lookup. Outputs: a non-nil error mentioning "status" and id == Nil.
// Data choice: intentionally not seeding "Active" with createStatuses off is
// the only way to drive GetStatus to fail.
func TestCreatePrefix_ReturnsErrorWhenStatusUnresolvable(t *testing.T) {
	// Arrange: server reports no matching status, and createStatuses is off, so
	// GetStatus fails.
	resetIPAMCaches()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	// Intentionally DO NOT seed the "Active" status.

	prefix := &devicetypes.CaniPrefix{
		ID:         uuid.New(),
		Prefix:     "10.9.0.0/24",
		PrefixLen:  24,
		ObjectMeta: devicetypes.ObjectMeta{Status: "Active"},
	}
	result := &LoadResult{}

	// Act.
	id, err := e.createPrefix(
		context.Background(),
		prefix,
		uuid.New(),                // namespaceID
		map[uuid.UUID]uuid.UUID{}, // vlanMap
		map[uuid.UUID]uuid.UUID{}, // createdPrefixes
		result,
	)

	// Assert.
	if err == nil {
		t.Fatal("expected an error when status cannot be resolved")
	}
	if id != uuid.Nil {
		t.Errorf("expected uuid.Nil on error, got %s", id)
	}
	if !strings.Contains(err.Error(), "status") {
		t.Errorf("error should mention the unresolved status, got: %v", err)
	}
}

// decodeSentPrefix unmarshals a captured prefix-create body, failing the test
// on malformed JSON.
func decodeSentPrefix(t *testing.T, body []byte) sentPrefix {
	t.Helper()
	var p sentPrefix
	if err := json.Unmarshal(body, &p); err != nil {
		t.Fatalf("decode prefix payload: %v\nbody: %s", err, body)
	}
	return p
}
