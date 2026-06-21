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
package imprt

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// twoPageServer returns a fake Nautobot server that serves the given path as two
// pages: the first carries a non-empty "next" URL (forcing the fetcher to
// continue) and the second clears it (ending the loop). It records the number of
// requests in calls so a test can assert both pages were fetched.
func twoPageServer(t *testing.T, path string, page1, page2 []map[string]interface{}, calls *int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		page := atomic.AddInt32(calls, 1)
		w.Header().Set("Content-Type", "application/json")
		if page == 1 {
			next := fmt.Sprintf("http://%s%s?limit=100&offset=100", r.Host, path)
			_, _ = w.Write(paginatedResponse(2, page1, next))
			return
		}
		_, _ = w.Write(paginatedResponse(2, page2, ""))
	}))
}

// TestFetchDevices_Pagination verifies FetchDevices follows the "next" link and
// concatenates results across multiple API pages.
//
// Why it matters: a Nautobot import must not silently truncate at the first page
// boundary; large fleets span many pages, and a fetcher that stops early would
// import an incomplete inventory.
// Inputs: a fake server returning two single-item pages for /dcim/devices/, the
// first advertising a next URL. Outputs: a two-element slice and exactly two API
// calls. Data choice: two distinct device names ("dev1"/"dev2") confirm results
// from both pages are appended rather than one page being fetched twice.
func TestFetchDevices_Pagination(t *testing.T) {
	var calls int32
	srv := twoPageServer(t, "/dcim/devices/",
		[]map[string]interface{}{{"name": "dev1"}},
		[]map[string]interface{}{{"name": "dev2"}}, &calls)
	defer srv.Close()

	got, err := FetchDevices(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchDevices: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(got))
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 API calls, got %d", calls)
	}
}

// TestFetchDeviceTypes_Pagination verifies FetchDeviceTypes follows the "next"
// link and concatenates results across multiple API pages.
//
// Why it matters: device-type catalogs can exceed one page, and an import that
// dropped later pages would fail to resolve devices whose model lives beyond the
// first page.
// Inputs: a fake server returning two single-item pages for /dcim/device-types/,
// the first advertising a next URL. Outputs: a two-element slice and exactly two
// API calls. Data choice: two distinct model names exercise the append-and-
// continue path uniquely.
func TestFetchDeviceTypes_Pagination(t *testing.T) {
	var calls int32
	srv := twoPageServer(t, "/dcim/device-types/",
		[]map[string]interface{}{{"model": "dt1"}},
		[]map[string]interface{}{{"model": "dt2"}}, &calls)
	defer srv.Close()

	got, err := FetchDeviceTypes(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchDeviceTypes: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 device types, got %d", len(got))
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 API calls, got %d", calls)
	}
}

// TestFetchInterfaces_Pagination verifies FetchInterfaces follows the "next"
// link and concatenates results across multiple API pages.
//
// Why it matters: a single device can expose dozens of interfaces, so a fleet
// easily exceeds one page; truncating would import devices with missing ports.
// Inputs: a fake server returning two single-item pages for /dcim/interfaces/,
// the first advertising a next URL. Outputs: a two-element slice and exactly two
// API calls. Data choice: two interface names confirm both pages are merged.
func TestFetchInterfaces_Pagination(t *testing.T) {
	var calls int32
	srv := twoPageServer(t, "/dcim/interfaces/",
		[]map[string]interface{}{{"name": "eth0"}},
		[]map[string]interface{}{{"name": "eth1"}}, &calls)
	defer srv.Close()

	got, err := FetchInterfaces(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchInterfaces: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 interfaces, got %d", len(got))
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 API calls, got %d", calls)
	}
}

// TestFetchModules_Pagination verifies FetchModules follows the "next" link and
// concatenates results across multiple API pages.
//
// Why it matters: modular chassis report many modules; a fetcher that stopped at
// the first page would import a partial module tree.
// Inputs: a fake server returning two single-item pages for /dcim/modules/, the
// first advertising a next URL. Outputs: a two-element slice and exactly two API
// calls. Data choice: two module serials exercise the continuation branch.
func TestFetchModules_Pagination(t *testing.T) {
	var calls int32
	srv := twoPageServer(t, "/dcim/modules/",
		[]map[string]interface{}{{"serial": "m1"}},
		[]map[string]interface{}{{"serial": "m2"}}, &calls)
	defer srv.Close()

	got, err := FetchModules(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchModules: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(got))
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 API calls, got %d", calls)
	}
}

// TestFetchModuleBays_Pagination verifies FetchModuleBays follows the "next"
// link and concatenates results across multiple API pages.
//
// Why it matters: module bays define where modules slot in; missing later pages
// would leave imported modules unanchored to their bays.
// Inputs: a fake server returning two single-item pages for /dcim/module-bays/,
// the first advertising a next URL. Outputs: a two-element slice and exactly two
// API calls. Data choice: two bay names confirm cross-page accumulation.
func TestFetchModuleBays_Pagination(t *testing.T) {
	var calls int32
	srv := twoPageServer(t, "/dcim/module-bays/",
		[]map[string]interface{}{{"name": "bay1"}},
		[]map[string]interface{}{{"name": "bay2"}}, &calls)
	defer srv.Close()

	got, err := FetchModuleBays(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchModuleBays: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 module bays, got %d", len(got))
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 API calls, got %d", calls)
	}
}

// TestFetchInventoryItems_Pagination verifies FetchInventoryItems follows the
// "next" link and concatenates results across multiple API pages.
//
// Why it matters: inventory items (FRUs, sub-components) are numerous per device,
// so they routinely paginate; truncation would drop hardware provenance.
// Inputs: a fake server returning two single-item pages for
// /dcim/inventory-items/, the first advertising a next URL. Outputs: a
// two-element slice and exactly two API calls. Data choice: two item names verify
// both pages are gathered.
func TestFetchInventoryItems_Pagination(t *testing.T) {
	var calls int32
	srv := twoPageServer(t, "/dcim/inventory-items/",
		[]map[string]interface{}{{"name": "item1"}},
		[]map[string]interface{}{{"name": "item2"}}, &calls)
	defer srv.Close()

	got, err := FetchInventoryItems(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchInventoryItems: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 inventory items, got %d", len(got))
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 API calls, got %d", calls)
	}
}

// TestFetchStatuses_Pagination verifies FetchStatuses follows the "next" link and
// concatenates results across multiple API pages.
//
// Why it matters: status definitions are shared metadata the importer resolves
// devices against; missing a later page would make some statuses unresolvable.
// Inputs: a fake server returning two single-item pages for /extras/statuses/,
// each item carrying the required content_types, the first advertising a next
// URL. Outputs: a two-element slice and exactly two API calls. Data choice: two
// status names ("Active"/"Staged") with a device content type mirror the
// SinglePage fixture while exercising the continuation branch.
func TestFetchStatuses_Pagination(t *testing.T) {
	var calls int32
	srv := twoPageServer(t, "/extras/statuses/",
		[]map[string]interface{}{{"name": "Active", "content_types": []string{"dcim.device"}}},
		[]map[string]interface{}{{"name": "Staged", "content_types": []string{"dcim.device"}}}, &calls)
	defer srv.Close()

	got, err := FetchStatuses(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchStatuses: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(got))
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 API calls, got %d", calls)
	}
}

// TestFetchTransportErrors verifies every Fetch* helper surfaces a transport
// (connection) error rather than returning a nil error or panicking when the
// Nautobot endpoint is unreachable.
//
// Why it matters: import runs against a live Nautobot over the network; if a
// fetcher swallowed a dial failure it would report an empty-but-successful
// result and the caller would wipe or skip real inventory believing the server
// was simply empty. Every fetcher must propagate the error so the import aborts.
// Inputs: a client pointed at an httptest server that is closed before the call,
// guaranteeing connection-refused on the first request for each of the 11
// fetchers. Outputs: a non-nil error from each fetcher. Data choice: closing the
// server (rather than returning a 500) exercises the transport-level err!=nil
// branch specifically, which the existing per-fetcher ServerError tests (which
// hit the non-200 branch) do not cover.
func TestFetchTransportErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	client := newTestClient(t, srv.URL)
	srv.Close() // subsequent dials are refused

	ctx := context.Background()
	cases := []struct {
		name  string
		fetch func() error
	}{
		{"locations", func() error { _, err := FetchLocations(ctx, client); return err }},
		{"racks", func() error { _, err := FetchRacks(ctx, client); return err }},
		{"devices", func() error { _, err := FetchDevices(ctx, client); return err }},
		{"device-types", func() error { _, err := FetchDeviceTypes(ctx, client); return err }},
		{"interfaces", func() error { _, err := FetchInterfaces(ctx, client); return err }},
		{"modules", func() error { _, err := FetchModules(ctx, client); return err }},
		{"module-bays", func() error { _, err := FetchModuleBays(ctx, client); return err }},
		{"cables", func() error { _, err := FetchCables(ctx, client); return err }},
		{"inventory-items", func() error { _, err := FetchInventoryItems(ctx, client); return err }},
		{"statuses", func() error { _, err := FetchStatuses(ctx, client); return err }},
		{"roles", func() error { _, err := FetchRoles(ctx, client); return err }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fetch(); err == nil {
				t.Errorf("expected transport error for %s, got nil", tc.name)
			}
		})
	}
}
