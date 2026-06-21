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
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// ifaceResultJSON builds one entry of a /dcim/interfaces/ list response.
func ifaceResultJSON(id uuid.UUID, name string) string {
	return fmt.Sprintf(`{"id":%q,"name":%q,"display":%q}`, id.String(), name, name)
}

// interfaceListServer returns a handler that replies to any GET on
// /dcim/interfaces/ with the supplied result entries, counting invocations so
// tests can assert caching behavior.
func interfaceListServer(calls *int, results ...string) http.HandlerFunc {
	body := fmt.Sprintf(`{"count":%d,"results":[%s]}`, len(results), strings.Join(results, ","))
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "dcim/interfaces") {
			*calls++
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, body)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	}
}

// -----------------------------------------------------------------------------
// GetInterfacesByDevice — single GET returning every interface on a device.
// -----------------------------------------------------------------------------

// TestGetInterfacesByDevice_MapsResults verifies that GetInterfacesByDevice
// issues one list call and maps each Nautobot interface into a CachedItem
// carrying its id and name.
//
// Why it matters: interfaces belong to a device; fetching them in bulk lets the
// export resolve and cache a device's ports efficiently before linking cables.
// Inputs: a device UUID and a one-entry list ("eth0"). Outputs: a []*CachedItem
// with the mapped id and name.
// Data choice: a single eth0 entry is the minimal fixture that proves field
// mapping from the list response.
func TestGetInterfacesByDevice_MapsResults(t *testing.T) {
	ifaceID := uuid.New()
	var calls int
	e, cleanup := newExporterWithServer(t, interfaceListServer(&calls, ifaceResultJSON(ifaceID, "eth0")))
	defer cleanup()

	items, err := e.Cache.GetInterfacesByDevice(uuid.New())
	if err != nil {
		t.Fatalf("GetInterfacesByDevice() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != ifaceID || items[0].Name != "eth0" {
		t.Errorf("unexpected interfaces: %+v", items)
	}
}

// TestGetInterfacesByDevice_ExtractsAttachedCableID verifies that
// GetInterfacesByDevice populates CachedItem.CableID when the interface JSON
// carries a nested cable reference.
//
// Why it matters: cable IDs let the export reconcile existing physical
// connections so cabling is not duplicated or lost on re-export.
// Inputs: an interface result embedding "cable":{"id":...}. Outputs: a
// CachedItem whose CableID is set to that UUID.
// Data choice: hand-built JSON with a nested cable object exercises the
// optional cable-extraction branch that the helper's plain entries omit.
func TestGetInterfacesByDevice_ExtractsAttachedCableID(t *testing.T) {
	ifaceID := uuid.New()
	cableID := uuid.New()
	result := fmt.Sprintf(`{"id":%q,"name":"eth0","display":"eth0","cable":{"id":%q}}`,
		ifaceID.String(), cableID.String())
	var calls int
	e, cleanup := newExporterWithServer(t, interfaceListServer(&calls, result))
	defer cleanup()

	items, err := e.Cache.GetInterfacesByDevice(uuid.New())
	if err != nil {
		t.Fatalf("GetInterfacesByDevice() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(items))
	}
	if items[0].CableID != cableID {
		t.Errorf("CableID = %s, want %s", items[0].CableID, cableID)
	}
}

// TestGetInterfacesByDevice_ReturnsErrorOnNon200 verifies that
// GetInterfacesByDevice returns an error when the list responds with 500.
//
// Why it matters: a failed interface fetch must surface so the export does not
// proceed against an incomplete view of a device's ports.
// Inputs: a 500 response. Outputs: a non-nil error.
// Data choice: a 500 models a server-side failure on the list endpoint.
func TestGetInterfacesByDevice_ReturnsErrorOnNon200(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusInternalServerError, `{}`))
	defer cleanup()

	if _, err := e.Cache.GetInterfacesByDevice(uuid.New()); err == nil {
		t.Fatal("expected an error when the interface list responds with 500")
	}
}

// -----------------------------------------------------------------------------
// PrefetchInterfacesForDevice — fetch once, then serve from the local cache.
// -----------------------------------------------------------------------------

// TestPrefetchInterfacesForDevice_FetchesOnceThenShortCircuits verifies that
// PrefetchInterfacesForDevice fetches a device's interfaces once, caches them,
// and that a second call is a no-op (still one HTTP call) while the prefetched
// interface resolves via GetInterfaceByDeviceAndName.
//
// Why it matters: prefetching avoids per-interface queries (which break on
// names containing "/") and redundant re-fetching, keeping large exports fast
// and correct.
// Inputs: the same device UUID twice and a one-entry list. Outputs: calls == 1
// and the cached interface is resolvable by name.
// Data choice: calling prefetch twice and asserting calls == 1 directly proves
// the prefetched-flag short-circuit and the cache population.
func TestPrefetchInterfacesForDevice_FetchesOnceThenShortCircuits(t *testing.T) {
	deviceID := uuid.New()
	var calls int
	e, cleanup := newExporterWithServer(t, interfaceListServer(&calls, ifaceResultJSON(uuid.New(), "eth0")))
	defer cleanup()

	if err := e.Cache.PrefetchInterfacesForDevice(deviceID); err != nil {
		t.Fatalf("first PrefetchInterfacesForDevice() error = %v", err)
	}
	// A second prefetch must be a no-op because the device is flagged as done.
	if err := e.Cache.PrefetchInterfacesForDevice(deviceID); err != nil {
		t.Fatalf("second PrefetchInterfacesForDevice() error = %v", err)
	}
	if calls != 1 {
		t.Errorf("expected exactly 1 HTTP fetch, got %d", calls)
	}

	// The prefetched interface must now be resolvable from the cache.
	item, err := e.Cache.GetInterfaceByDeviceAndName(deviceID, "eth0")
	if err != nil {
		t.Fatalf("GetInterfaceByDeviceAndName() error = %v", err)
	}
	if item == nil {
		t.Error("expected the prefetched interface to be cached")
	}
}

// TestPrefetchInterfacesForDevice_PropagatesError verifies that
// PrefetchInterfacesForDevice propagates the error from the underlying list
// when it fails with 500.
//
// Why it matters: if the bulk fetch fails the cache must stay empty and
// unflagged so later lookups do not treat a device as fully prefetched.
// Inputs: a 500 response. Outputs: a non-nil error.
// Data choice: a 500 reuses the standard server-error fixture to drive the
// failure path.
func TestPrefetchInterfacesForDevice_PropagatesError(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusInternalServerError, `{}`))
	defer cleanup()

	if err := e.Cache.PrefetchInterfacesForDevice(uuid.New()); err == nil {
		t.Fatal("expected an error to propagate from the underlying interface list")
	}
}

// -----------------------------------------------------------------------------
// GetInterfaceByDeviceAndNameFuzzy — exact first, then normalized matching.
// -----------------------------------------------------------------------------

// TestGetInterfaceByDeviceAndNameFuzzy_FindsExactMatch verifies that
// GetInterfaceByDeviceAndNameFuzzy returns the interface whose name matches
// exactly, before any fuzzy logic runs.
//
// Why it matters: most cani interface names map 1:1 to Nautobot, so the exact
// path must win to avoid mis-associating ports during the export.
// Inputs: a device UUID and search "eth0" against a remote "eth0". Outputs: the
// matching *CachedItem.
// Data choice: identical search and remote names isolate the exact-match
// branch from the normalization fallbacks.
func TestGetInterfaceByDeviceAndNameFuzzy_FindsExactMatch(t *testing.T) {
	deviceID := uuid.New()
	ifaceID := uuid.New()
	var calls int
	e, cleanup := newExporterWithServer(t, interfaceListServer(&calls, ifaceResultJSON(ifaceID, "eth0")))
	defer cleanup()

	item, err := e.Cache.GetInterfaceByDeviceAndNameFuzzy(deviceID, "eth0")
	if err != nil {
		t.Fatalf("GetInterfaceByDeviceAndNameFuzzy() error = %v", err)
	}
	if item == nil || item.ID != ifaceID {
		t.Errorf("expected the exact interface %s, got %+v", ifaceID, item)
	}
}

// TestGetInterfaceByDeviceAndNameFuzzy_FindsNormalizedMatch verifies that, when
// no exact match exists, the fuzzy lookup connects names that normalize equally
// ("eth0" and "Gig-E 0" both reduce to "0").
//
// Why it matters: cani and Nautobot often label the same port differently;
// normalized matching lets the export reconcile them instead of creating
// duplicate interfaces.
// Inputs: search "eth0" against a remote "Gig-E 0". Outputs: the remote
// interface *CachedItem.
// Data choice: "Gig-E 0" vs "eth0" specifically exercises the prefix-stripping
// normalization, since both collapse to "0".
func TestGetInterfaceByDeviceAndNameFuzzy_FindsNormalizedMatch(t *testing.T) {
	deviceID := uuid.New()
	ifaceID := uuid.New()
	// The remote interface is named "Gig-E 0"; both it and the search term
	// "eth0" normalize to "0", so fuzzy matching must connect them.
	var calls int
	e, cleanup := newExporterWithServer(t, interfaceListServer(&calls, ifaceResultJSON(ifaceID, "Gig-E 0")))
	defer cleanup()

	item, err := e.Cache.GetInterfaceByDeviceAndNameFuzzy(deviceID, "eth0")
	if err != nil {
		t.Fatalf("GetInterfaceByDeviceAndNameFuzzy() error = %v", err)
	}
	if item == nil || item.ID != ifaceID {
		t.Errorf("expected a fuzzy match to %s, got %+v", ifaceID, item)
	}
}

// TestGetInterfaceByDeviceAndNameFuzzy_ReturnsNilWhenNoMatch verifies that the
// fuzzy lookup returns (nil, nil) — not an error — when neither exact,
// normalized, nor numeric port matching finds the interface.
//
// Why it matters: a not-found interface is an expected, non-fatal outcome the
// export handles (e.g. skip or create), so it must stay distinguishable from a
// real error.
// Inputs: search "Ethernet99" against only a remote "mgmt". Outputs: a nil item
// and a nil error.
// Data choice: unrelated names with no shared port number ensure all three
// matching strategies miss.
func TestGetInterfaceByDeviceAndNameFuzzy_ReturnsNilWhenNoMatch(t *testing.T) {
	deviceID := uuid.New()
	var calls int
	e, cleanup := newExporterWithServer(t, interfaceListServer(&calls, ifaceResultJSON(uuid.New(), "mgmt")))
	defer cleanup()

	item, err := e.Cache.GetInterfaceByDeviceAndNameFuzzy(deviceID, "Ethernet99")
	if err != nil {
		t.Fatalf("GetInterfaceByDeviceAndNameFuzzy() error = %v", err)
	}
	if item != nil {
		t.Errorf("expected no match, got %+v", item)
	}
}
