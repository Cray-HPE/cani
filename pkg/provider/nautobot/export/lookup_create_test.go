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

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// createdItemJSON is the canonical 201 body Nautobot returns for a freshly
// created reference object (status, role, manufacturer, rack): an id, a name,
// and the human-friendly display string the exporter dereferences.
func createdItemJSON(id uuid.UUID, name string) string {
	return fmt.Sprintf(`{"id":%q,"name":%q,"display":%q}`, id.String(), name, name)
}

// -----------------------------------------------------------------------------
// CreateStatus — single POST to /extras/statuses/.
// -----------------------------------------------------------------------------

// TestCreateStatus_CachesCreatedStatus verifies that a single POST to
// /extras/statuses/ returns the created status with the server-assigned ID and
// stores it in the name-keyed status cache.
//
// Why it matters: statuses gate the lifecycle state of every device, rack,
// module, and IPAM object the export writes; caching the created status lets
// later lookups reuse it instead of POSTing a duplicate.
// Inputs: the status name "Burned-In". Outputs: a *CachedItem plus an entry in
// c.statuses keyed by that name.
// Data choice: "Burned-In" is a cani-specific state absent from Nautobot's
// stock defaults, so the create path (not a pre-existing match) is exercised.
func TestCreateStatus_CachesCreatedStatus(t *testing.T) {
	id := uuid.New()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "extras/statuses") {
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, createdItemJSON(id, "Burned-In"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	item, err := e.Cache.CreateStatus("Burned-In")
	if err != nil {
		t.Fatalf("CreateStatus() error = %v", err)
	}
	if item.ID != id {
		t.Errorf("status ID = %s, want %s", item.ID, id)
	}
	// The created status must be cached under its name for later lookups.
	if cached := e.Cache.statuses["Burned-In"]; cached == nil || cached.ID != id {
		t.Errorf("status was not cached under its name: %+v", cached)
	}
}

// TestCreateStatus_ReturnsErrorOnNon201 verifies that CreateStatus returns an
// error when Nautobot answers the POST with a non-201 status.
//
// Why it matters: a failed status create must surface so the export aborts
// rather than caching a phantom status and later attaching devices to an ID
// that was never persisted.
// Inputs: the status name "Burned-In". Outputs: a non-nil error and no cache
// entry.
// Data choice: a 400 with {"name":["already exists"]} mimics Nautobot's real
// duplicate-name rejection, the most likely non-201 in practice.
func TestCreateStatus_ReturnsErrorOnNon201(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"name":["already exists"]}`)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	if _, err := e.Cache.CreateStatus("Burned-In"); err == nil {
		t.Fatal("expected an error when the create responds with 400")
	}
}

// -----------------------------------------------------------------------------
// CreateRole — single POST to /extras/roles/.
// -----------------------------------------------------------------------------

// TestCreateRole_CachesCreatedRole verifies that a single POST to /extras/roles/
// returns the created role with the server-assigned ID and stores it in the
// name-keyed role cache.
//
// Why it matters: roles classify exported devices (leaf, spine, compute);
// caching the created role prevents a duplicate POST the next time a device of
// the same role is exported.
// Inputs: the role name "Leaf". Outputs: a *CachedItem plus an entry in c.roles
// keyed by that name.
// Data choice: "Leaf" is a realistic switch role that would not already exist,
// exercising the create-and-cache path.
func TestCreateRole_CachesCreatedRole(t *testing.T) {
	id := uuid.New()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "extras/roles") {
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, createdItemJSON(id, "Leaf"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	item, err := e.Cache.CreateRole("Leaf")
	if err != nil {
		t.Fatalf("CreateRole() error = %v", err)
	}
	if item.ID != id {
		t.Errorf("role ID = %s, want %s", item.ID, id)
	}
	if cached := e.Cache.roles["Leaf"]; cached == nil || cached.ID != id {
		t.Errorf("role was not cached under its name: %+v", cached)
	}
}

// TestCreateRole_ReturnsErrorOnNon201 verifies that CreateRole returns an error
// when Nautobot answers the POST with a non-201 status.
//
// Why it matters: propagating the failure stops the export from referencing a
// role ID that was never created when assigning devices.
// Inputs: the role name "Leaf". Outputs: a non-nil error and no cache entry.
// Data choice: a 500 simulates a server-side fault (distinct from the 400
// duplicate case used for statuses), covering the generic non-201 branch.
func TestCreateRole_ReturnsErrorOnNon201(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{}`)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	if _, err := e.Cache.CreateRole("Leaf"); err == nil {
		t.Fatal("expected an error when the create responds with 500")
	}
}

// -----------------------------------------------------------------------------
// GetOrCreateManufacturer — GET to find, then POST to create on a miss.
// -----------------------------------------------------------------------------

// TestGetOrCreateManufacturer_ReturnsCachedWithoutHTTP verifies that a
// manufacturer already present in the cache is returned without any HTTP call.
//
// Why it matters: a large export resolves the same manufacturer for many device
// types; the cache must short-circuit to avoid redundant round-trips to
// Nautobot.
// Inputs: name "HPE", pre-seeded in c.manufacturers. Outputs: the cached
// *CachedItem with its original ID.
// Data choice: the cache is built on a nil client so any HTTP attempt would
// panic, proving the lookup never touches the network.
func TestGetOrCreateManufacturer_ReturnsCachedWithoutHTTP(t *testing.T) {
	cache := NewLookupCache(nil) // nil client: any HTTP would panic
	id := uuid.New()
	cache.manufacturersMu.Lock()
	cache.manufacturers["HPE"] = &CachedItem{ID: id, Name: "HPE"}
	cache.manufacturersMu.Unlock()

	item, err := cache.GetOrCreateManufacturer("HPE")
	if err != nil {
		t.Fatalf("GetOrCreateManufacturer() error = %v", err)
	}
	if item.ID != id {
		t.Errorf("manufacturer ID = %s, want cached %s", item.ID, id)
	}
}

// TestGetOrCreateManufacturer_ReturnsExistingFromList verifies that when the GET
// finds an existing manufacturer, its ID is returned and no create POST is sent.
//
// Why it matters: re-exporting must be idempotent — an existing manufacturer in
// Nautobot must be reused, never duplicated.
// Inputs: name "HPE". Outputs: the existing *CachedItem; a posted flag asserts
// the POST handler was not reached.
// Data choice: the GET returns one match while the POST handler would supply a
// different ID, so reusing the GET's ID (and posted==false) proves no create
// occurred.
func TestGetOrCreateManufacturer_ReturnsExistingFromList(t *testing.T) {
	id := uuid.New()
	var posted bool
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "dcim/manufacturers"):
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `{"count":1,"results":[%s]}`, createdItemJSON(id, "HPE"))
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "dcim/manufacturers"):
			posted = true
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, createdItemJSON(uuid.New(), "HPE"))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `{}`)
		}
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	item, err := e.Cache.GetOrCreateManufacturer("HPE")
	if err != nil {
		t.Fatalf("GetOrCreateManufacturer() error = %v", err)
	}
	if item.ID != id {
		t.Errorf("manufacturer ID = %s, want the existing %s", item.ID, id)
	}
	if posted {
		t.Error("must not POST a create when the manufacturer already exists")
	}
}

// TestGetOrCreateManufacturer_CreatesOnMiss verifies that when the GET finds no
// manufacturer, a create POST is sent and the created object is returned.
//
// Why it matters: device types reference a manufacturer, so a missing one must
// be auto-created or the device-type export would fail.
// Inputs: name "Acme". Outputs: the created *CachedItem carrying the
// server-assigned ID.
// Data choice: an empty results list ({"count":0}) forces the create branch,
// and "Acme" is an arbitrary new vendor not present in Nautobot.
func TestGetOrCreateManufacturer_CreatesOnMiss(t *testing.T) {
	createdID := uuid.New()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "dcim/manufacturers"):
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `{"count":0,"results":[]}`)
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "dcim/manufacturers"):
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, createdItemJSON(createdID, "Acme"))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `{}`)
		}
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	item, err := e.Cache.GetOrCreateManufacturer("Acme")
	if err != nil {
		t.Fatalf("GetOrCreateManufacturer() error = %v", err)
	}
	if item.ID != createdID {
		t.Errorf("manufacturer ID = %s, want the created %s", item.ID, createdID)
	}
}

// TestGetOrCreateManufacturer_ReturnsErrorOnListFailure verifies that a failed
// lookup GET returns an error instead of falling through to a create.
//
// Why it matters: treating a transport/server failure as "not found" would
// wrongly POST a new manufacturer and risk duplicates in Nautobot.
// Inputs: name "Acme". Outputs: a non-nil error and no create attempt.
// Data choice: the server returns 500 for every request, so the very first
// lookup call fails before any create is considered.
func TestGetOrCreateManufacturer_ReturnsErrorOnListFailure(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{}`)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	if _, err := e.Cache.GetOrCreateManufacturer("Acme"); err == nil {
		t.Fatal("expected an error when the manufacturer lookup responds with 500")
	}
}

// TestGetOrCreateManufacturer_ReturnsErrorWhenCreateFails verifies that an error
// is returned when the lookup misses but the subsequent create POST fails.
//
// Why it matters: the failure must surface so the export does not cache or
// reference a manufacturer that was never persisted.
// Inputs: name "Acme". Outputs: a non-nil error.
// Data choice: the GET returns an empty list to reach the create branch, then
// every other request (the POST) answers 400 to fail it.
func TestGetOrCreateManufacturer_ReturnsErrorWhenCreateFails(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "dcim/manufacturers"):
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `{"count":0,"results":[]}`)
		default:
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{}`)
		}
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	if _, err := e.Cache.GetOrCreateManufacturer("Acme"); err == nil {
		t.Fatal("expected an error when the manufacturer create responds with 400")
	}
}

// -----------------------------------------------------------------------------
// GetRackByName + resolveRackNautobotID — rack lookups have no dedicated cache.
// -----------------------------------------------------------------------------

// TestGetRackByName_ReturnsMatch verifies that GetRackByName returns the rack's
// Nautobot ID when the name filter matches exactly one rack.
//
// Why it matters: a device's placement in Nautobot is anchored to its parent
// rack, so resolving the rack name to its UUID is a prerequisite for exporting
// device positions.
// Inputs: rack name "Rack-1". Outputs: a *CachedItem with the matching ID.
// Data choice: a single-result list is the normal success case; racks have no
// dedicated cache, so the call always queries the API.
func TestGetRackByName_ReturnsMatch(t *testing.T) {
	id := uuid.New()
	var calls int
	body := fmt.Sprintf(`{"count":1,"results":[%s]}`, createdItemJSON(id, "Rack-1"))
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, body))
	defer cleanup()

	item, err := e.Cache.GetRackByName("Rack-1")
	if err != nil {
		t.Fatalf("GetRackByName() error = %v", err)
	}
	if item == nil || item.ID != id {
		t.Errorf("rack = %+v, want ID %s", item, id)
	}
}

// TestGetRackByName_ReturnsNilWithoutErrorWhenAbsent verifies that an empty
// result set yields (nil, nil) rather than an error.
//
// Why it matters: a not-yet-exported rack is an expected condition, not a
// failure; callers treat nil as "rack absent" and continue instead of aborting.
// Inputs: rack name "Missing". Outputs: a nil item and a nil error.
// Data choice: a {"count":0} body models the absent-rack case that
// distinguishes "missing" from a genuine API error.
func TestGetRackByName_ReturnsNilWithoutErrorWhenAbsent(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{"count":0,"results":[]}`))
	defer cleanup()

	item, err := e.Cache.GetRackByName("Missing")
	if err != nil {
		t.Fatalf("GetRackByName() error = %v, want nil for an absent rack", err)
	}
	if item != nil {
		t.Errorf("expected nil item for an absent rack, got %+v", item)
	}
}

// TestGetRackByName_ReturnsErrorOnNon200 verifies that a non-200 response from
// the rack lookup produces an error.
//
// Why it matters: a real API failure must be distinguishable from an absent
// rack so the export does not silently skip valid placements.
// Inputs: rack name "Rack-1". Outputs: a non-nil error.
// Data choice: a 500 represents a server-side failure, the error case paired
// with the nil-without-error absent case above.
func TestGetRackByName_ReturnsErrorOnNon200(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusInternalServerError, `{}`))
	defer cleanup()

	if _, err := e.Cache.GetRackByName("Rack-1"); err == nil {
		t.Fatal("expected an error when the rack lookup responds with 500")
	}
}

// TestResolveRackNautobotID_ResolvesViaInventoryAndCache verifies that a
// device's parent UUID is mapped through the cani inventory to a rack name, then
// resolved via the lookup cache's GetRackByName to the rack's Nautobot UUID.
//
// Why it matters: swap/placement logic needs the Nautobot rack ID, which lives
// in a different ID space than the cani inventory's rack UUID; this two-step
// bridge connects them.
// Inputs: a device whose Parent points at an inventory rack named "Rack-1".
// Outputs: the Nautobot rack UUID returned by the lookup.
// Data choice: the cani rack UUID and the Nautobot rack UUID are deliberately
// different so the result proves resolution reached the Nautobot side.
func TestResolveRackNautobotID_ResolvesViaInventoryAndCache(t *testing.T) {
	rackID := uuid.New()
	nautobotRackID := uuid.New()
	var calls int
	body := fmt.Sprintf(`{"count":1,"results":[%s]}`, createdItemJSON(nautobotRackID, "Rack-1"))
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, body))
	defer cleanup()

	device := &devicetypes.CaniDeviceType{Name: "node-1", Parent: rackID}
	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackID: {ID: rackID, Name: "Rack-1"},
		},
	}

	got := e.resolveRackNautobotID(device, inv)
	if got != nautobotRackID {
		t.Errorf("resolveRackNautobotID = %s, want %s", got, nautobotRackID)
	}
}

// TestResolveRackNautobotID_ReturnsNilWhenDeviceHasNoParent verifies that a
// device with no parent rack resolves to uuid.Nil without any HTTP call.
//
// Why it matters: unracked devices are valid and must short-circuit cheaply
// instead of triggering a needless rack lookup during export.
// Inputs: a device with a zero-value Parent. Outputs: uuid.Nil.
// Data choice: the Exporter is built on a nil client, so reaching the API would
// panic — proving the no-parent guard returns before any request.
func TestResolveRackNautobotID_ReturnsNilWhenDeviceHasNoParent(t *testing.T) {
	// No parent means no rack to resolve; a nil client proves no HTTP occurs.
	e := &Exporter{Cache: NewLookupCache(nil)}
	device := &devicetypes.CaniDeviceType{Name: "orphan"}

	if got := e.resolveRackNautobotID(device, &devicetypes.Inventory{}); got != uuid.Nil {
		t.Errorf("resolveRackNautobotID = %s, want uuid.Nil", got)
	}
}

// TestResolveRackNautobotID_ReturnsNilWhenRackMissingFromInventory verifies that
// a device whose parent UUID is absent from the inventory resolves to uuid.Nil.
//
// Why it matters: a dangling parent reference must degrade gracefully to "no
// rack" rather than crash or issue a lookup for a rack name it cannot find.
// Inputs: a device with a random Parent UUID and an empty Inventory. Outputs:
// uuid.Nil.
// Data choice: an empty Inventory plus a nil client guarantees the missing-rack
// branch is taken before any HTTP would occur.
func TestResolveRackNautobotID_ReturnsNilWhenRackMissingFromInventory(t *testing.T) {
	e := &Exporter{Cache: NewLookupCache(nil)}
	device := &devicetypes.CaniDeviceType{Name: "node-1", Parent: uuid.New()}

	if got := e.resolveRackNautobotID(device, &devicetypes.Inventory{}); got != uuid.Nil {
		t.Errorf("resolveRackNautobotID = %s, want uuid.Nil for an unknown rack", got)
	}
}
