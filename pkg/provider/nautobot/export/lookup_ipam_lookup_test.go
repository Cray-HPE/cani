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
	"net/http"
	"testing"

	"github.com/google/uuid"
)

// -----------------------------------------------------------------------------
// LookupVLAN — fetch from API on a cache miss; never auto-creates.
// -----------------------------------------------------------------------------

// TestLookupVLAN_FetchesAndCachesFromAPI verifies LookupVLAN fetches a VLAN from
// the API on a cache miss and serves the second call from cache (one HTTP call).
//
// Why it matters: IPAM export resolves VLANs by VID+location repeatedly; caching
// avoids hammering Nautobot, and LookupVLAN must never auto-create.
// Inputs: a VID and location. Outputs: a *CachedItem and an error; the call
// counter proves caching.
// Data choice: VID 100 in "DC1" with a single matching result is the canonical
// hit; the doubled call asserts the cache short-circuit via calls==1.
func TestLookupVLAN_FetchesAndCachesFromAPI(t *testing.T) {
	resetIPAMCaches()
	vlanNID := uuid.New()
	var calls int
	body := `{"count":1,"results":[{"id":"` + vlanNID.String() + `","vid":100,"name":"vlan100","display":"vlan100"}]}`
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, body))
	defer cleanup()

	item, err := e.Cache.LookupVLAN(100, "DC1")
	if err != nil {
		t.Fatalf("LookupVLAN() error = %v", err)
	}
	if item == nil || item.ID != vlanNID {
		t.Fatalf("expected VLAN %s, got %+v", vlanNID, item)
	}

	// A second call must be served from the cache without another HTTP request.
	if _, err := e.Cache.LookupVLAN(100, "DC1"); err != nil {
		t.Fatalf("LookupVLAN() second call error = %v", err)
	}
	if calls != 1 {
		t.Errorf("expected exactly one HTTP call (second served from cache), got %d", calls)
	}
}

// TestLookupVLAN_ReturnsNilWhenNotFound verifies LookupVLAN returns (nil, nil)
// when no VLAN matches, rather than an error.
//
// Why it matters: callers use a nil result to decide whether to create the VLAN;
// conflating "absent" with "error" would break that create/skip logic.
// Inputs: a VID/location with no match. Outputs: nil item and nil error.
// Data choice: VID 999 against an empty result list models a clean miss.
func TestLookupVLAN_ReturnsNilWhenNotFound(t *testing.T) {
	resetIPAMCaches()
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, emptyListJSON))
	defer cleanup()

	item, err := e.Cache.LookupVLAN(999, "DC1")
	if err != nil {
		t.Fatalf("LookupVLAN() error = %v", err)
	}
	if item != nil {
		t.Errorf("expected nil item when the VLAN is not found, got %+v", item)
	}
}

// TestLookupVLAN_ReturnsErrorOnNon200 verifies a non-200 VLAN list response is
// surfaced as an error.
//
// Why it matters: a server error during lookup must not be mistaken for "VLAN
// absent", which would trigger an erroneous create.
// Inputs: a VID/location with the server returning 500. Outputs: a non-nil
// error.
// Data choice: a 500 with a {"detail":"boom"} body represents a transient
// Nautobot failure distinct from an empty (not-found) result.
func TestLookupVLAN_ReturnsErrorOnNon200(t *testing.T) {
	resetIPAMCaches()
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusInternalServerError, `{"detail":"boom"}`))
	defer cleanup()

	if _, err := e.Cache.LookupVLAN(100, "DC1"); err == nil {
		t.Fatal("expected an error when the VLAN list responds with 500")
	}
}

// -----------------------------------------------------------------------------
// LookupIPAddress — raw fetch on a cache miss; never auto-creates.
// -----------------------------------------------------------------------------

// TestLookupIPAddress_FetchesAndCachesFromAPI verifies LookupIPAddress fetches an
// IP on a cache miss, parses id/address, and serves the second call from cache.
//
// Why it matters: assigning IPs to interfaces requires resolving existing IPs by
// address; caching keeps the export efficient and this path must never create.
// Inputs: an address string. Outputs: a *CachedItem (ID + display) and an error;
// calls==1 proves the cache.
// Data choice: "10.0.0.1/24" with a single result exercises the raw JSON parse
// of id/address/display and the cache key by address.
func TestLookupIPAddress_FetchesAndCachesFromAPI(t *testing.T) {
	resetIPAMCaches()
	ipNID := uuid.New()
	var calls int
	body := `{"results":[{"id":"` + ipNID.String() + `","address":"10.0.0.1/24","display":"10.0.0.1/24"}]}`
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, body))
	defer cleanup()

	item, err := e.Cache.LookupIPAddress("10.0.0.1/24")
	if err != nil {
		t.Fatalf("LookupIPAddress() error = %v", err)
	}
	if item == nil || item.ID != ipNID || item.Name != "10.0.0.1/24" {
		t.Fatalf("expected IP %s, got %+v", ipNID, item)
	}

	// The second lookup is served from the cache.
	if _, err := e.Cache.LookupIPAddress("10.0.0.1/24"); err != nil {
		t.Fatalf("LookupIPAddress() second call error = %v", err)
	}
	if calls != 1 {
		t.Errorf("expected exactly one HTTP call (second served from cache), got %d", calls)
	}
}

// TestLookupIPAddress_ReturnsNilWhenNotFound verifies LookupIPAddress returns
// (nil, nil) when no IP matches.
//
// Why it matters: a nil result tells the caller the IP must be created;
// returning an error instead would block the assignment flow.
// Inputs: an unmatched address. Outputs: nil item and nil error.
// Data choice: "10.0.0.9/24" against an empty results list models a clean miss.
func TestLookupIPAddress_ReturnsNilWhenNotFound(t *testing.T) {
	resetIPAMCaches()
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{"results":[]}`))
	defer cleanup()

	item, err := e.Cache.LookupIPAddress("10.0.0.9/24")
	if err != nil {
		t.Fatalf("LookupIPAddress() error = %v", err)
	}
	if item != nil {
		t.Errorf("expected nil item when the IP is not found, got %+v", item)
	}
}

// TestLookupIPAddress_ReturnsErrorOnNon200 verifies a non-200 IP list response is
// surfaced as an error.
//
// Why it matters: a server error must not be misread as "IP absent" and cause a
// duplicate create.
// Inputs: an address with the server returning 500. Outputs: a non-nil error.
// Data choice: a 500 with a detail body represents a transient Nautobot failure.
func TestLookupIPAddress_ReturnsErrorOnNon200(t *testing.T) {
	resetIPAMCaches()
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusInternalServerError, `{"detail":"boom"}`))
	defer cleanup()

	if _, err := e.Cache.LookupIPAddress("10.0.0.1/24"); err == nil {
		t.Fatal("expected an error when the IP list responds with 500")
	}
}

// TestLookupIPAddress_ReturnsErrorOnMalformedJSON verifies a malformed list body
// is surfaced as an error.
//
// Why it matters: LookupIPAddress parses the response by hand; a decode failure
// must propagate rather than silently yield a wrong or empty result.
// Inputs: an address with the server returning invalid JSON. Outputs: a non-nil
// error.
// Data choice: a truncated "{not valid json" body deterministically fails the
// JSON decoder.
func TestLookupIPAddress_ReturnsErrorOnMalformedJSON(t *testing.T) {
	resetIPAMCaches()
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{not valid json`))
	defer cleanup()

	if _, err := e.Cache.LookupIPAddress("10.0.0.1/24"); err == nil {
		t.Fatal("expected an error when the IP list body is malformed JSON")
	}
}

// TestLookupIPAddress_ReturnsErrorOnBadUUID verifies an IP whose returned ID is
// not a valid UUID is surfaced as an error.
//
// Why it matters: the parsed ID becomes a foreign key in later writes; an
// unparseable ID must fail loudly instead of corrupting references.
// Inputs: an address whose result carries id "not-a-uuid". Outputs: a non-nil
// error.
// Data choice: an obviously invalid "not-a-uuid" id isolates the UUID-parse
// branch from the JSON-decode branch.
func TestLookupIPAddress_ReturnsErrorOnBadUUID(t *testing.T) {
	resetIPAMCaches()
	var calls int
	body := `{"results":[{"id":"not-a-uuid","address":"10.0.0.1/24","display":"10.0.0.1/24"}]}`
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, body))
	defer cleanup()

	if _, err := e.Cache.LookupIPAddress("10.0.0.1/24"); err == nil {
		t.Fatal("expected an error when the returned IP ID is not a valid UUID")
	}
}
