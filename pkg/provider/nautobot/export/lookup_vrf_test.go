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
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// TestLookupCachedVRF_QualifiesAcrossNamespaces verifies a bare VRF name
// resolves while unique, becomes an error once a second namespace carries it,
// and a "namespace/name" reference always resolves exactly.
//
// Why it matters: interface specs reference VRFs by name; silently picking one
// of two same-named VRFs would attach interfaces to the wrong routing table.
// Inputs: "blue" cached under Global, then also under tenant-a. Outputs: the
// Global item for the bare name before the second cache, an error mentioning
// the qualifier after, and the tenant item for "tenant-a/blue".
// Data choice: caching the second copy mid-test isolates the transition from
// unique to ambiguous.
func TestLookupCachedVRF_QualifiesAcrossNamespaces(t *testing.T) {
	cache := NewLookupCache(nil)
	globalBlue, tenantBlue := uuid.New(), uuid.New()
	cache.CacheVRF("Global", "blue", &CachedItem{ID: globalBlue, Name: "blue"})

	if item, err := cache.LookupCachedVRF("blue"); err != nil || item == nil || item.ID != globalBlue {
		t.Fatalf("unique bare lookup = %v, %v; want Global blue", item, err)
	}
	cache.CacheVRF("tenant-a", "blue", &CachedItem{ID: tenantBlue, Name: "blue"})
	if _, err := cache.LookupCachedVRF("blue"); err == nil || !strings.Contains(err.Error(), "qualify it as namespace/blue") {
		t.Errorf("ambiguous bare lookup error = %v, want qualifier guidance", err)
	}
	if item, err := cache.LookupCachedVRF("tenant-a/blue"); err != nil || item == nil || item.ID != tenantBlue {
		t.Errorf("qualified lookup = %v, %v; want tenant blue", item, err)
	}
	if item, err := cache.LookupCachedVRF("tenant-b/blue"); err != nil || item != nil {
		t.Errorf("unknown qualified lookup = %v, %v; want nil, nil", item, err)
	}
}

// TestLookupVRF_KeepsASlashedNameApartFromAQualifiedOne verifies the cache
// distinguishes VRF "blue/green" in Global from VRF "green" in a namespace
// named "Global/blue".
//
// Why it matters: both Nautobot names may contain a slash, and the model's
// reference grammar allows a slashed namespace; a key that joined the two
// with a slash would make the second overwrite the first.
// Inputs: the two VRFs cached with distinct IDs. Outputs: LookupVRF returns
// each one under its own namespace and name.
// Data choice: the pair is the exact collision a joined "namespace/name" key
// produces.
func TestLookupVRF_KeepsASlashedNameApartFromAQualifiedOne(t *testing.T) {
	// Arrange.
	cache := NewLookupCache(nil)
	slashed, nested := uuid.New(), uuid.New()
	cache.CacheVRF("Global", "blue/green", &CachedItem{ID: slashed, Name: "blue/green"})
	cache.CacheVRF("Global/blue", "green", &CachedItem{ID: nested, Name: "green"})

	// Act.
	gotSlashed := cache.LookupVRF("Global", "blue/green")
	gotNested := cache.LookupVRF("Global/blue", "green")

	// Assert.
	if gotSlashed == nil || gotSlashed.ID != slashed {
		t.Errorf("LookupVRF(Global, blue/green) = %+v, want ID %s", gotSlashed, slashed)
	}
	if gotNested == nil || gotNested.ID != nested {
		t.Errorf("LookupVRF(Global/blue, green) = %+v, want ID %s", gotNested, nested)
	}
}

// TestFindVRF_FiltersByNamespaceAndRejectsMultiple verifies the remote VRF
// lookup carries the namespace filter and treats several results as an error.
//
// Why it matters: taking Results[0] of a name-only query is how a tenant VRF
// could be reused for another tenant's prefixes.
// Inputs: a server returning two VRFs for the query. Outputs: an error naming
// the count, and a request whose query includes name and namespace filters.
// Data choice: two results is the smallest ambiguous response.
func TestFindVRF_FiltersByNamespaceAndRejectsMultiple(t *testing.T) {
	var query string
	handler := func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"count":2,"results":[{"id":"`+uuid.New().String()+`","name":"blue"},{"id":"`+uuid.New().String()+`","name":"blue"}]}`)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	_, err := e.findVRF(context.Background(), "tenant-a", "blue")

	if err == nil || !strings.Contains(err.Error(), `2 VRFs named "blue" in namespace "tenant-a"`) {
		t.Errorf("findVRF error = %v, want multiple-match error", err)
	}
	if !strings.Contains(query, "name=blue") || !strings.Contains(query, "namespace=tenant-a") {
		t.Errorf("query = %q, want name and namespace filters", query)
	}
}
