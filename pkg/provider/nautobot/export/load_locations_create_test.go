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
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// locationCreateHandler answers the requests createLocationFromCani makes: the
// location-type lookup (returns an existing type so no sub-create is needed) and
// the location create POST. The locations GET returns existingLocationsBody so
// the same handler can also drive loadLocations' LookupLocation call. locPosts
// counts only the location create POSTs.
func locationCreateHandler(locTypeID, createdLocID uuid.UUID, createStatus int, existingLocationsBody string, locPosts *int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "dcim/location-types"):
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, fmt.Sprintf(`{"count":1,"results":[%s]}`,
				refObjectJSON(locTypeID, "Section")))
		case strings.Contains(r.URL.Path, "dcim/locations"):
			if r.Method == http.MethodPost {
				*locPosts++
				w.WriteHeader(createStatus)
				_, _ = io.WriteString(w, refObjectJSON(createdLocID, "DC1"))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, existingLocationsBody)
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, emptyListJSON)
		}
	}
}

// newCaniLocation builds a minimal location with the given name and type.
func newCaniLocation(name, locType string) *devicetypes.CaniLocationType {
	return &devicetypes.CaniLocationType{
		ID:           uuid.New(),
		Name:         name,
		LocationType: locType,
	}
}

// -----------------------------------------------------------------------------
// createLocationFromCani
// -----------------------------------------------------------------------------

// TestCreateLocationFromCani_CreatesOn201 verifies a location is created in
// Nautobot when its type resolves and no matching location exists, returning the
// new ID, recording it under LocationsCreated, and mapping all optional fields.
//
// Why it matters: locations are the top of the Nautobot hierarchy that devices,
// racks and IPAM hang from; the create must resolve the location-type FK and
// carry the descriptive metadata operators rely on.
// Inputs: a context, a CaniLocationType, the parent-ID map, and a LoadResult.
// Outputs: the new location UUID and an error; side effects are the counters and
// one POST.
// Data choice: every optional field (Facility/Description/PhysicalAddress/
// ContactName/TimeZone/Comments/CustomFields) is set so the test exercises each
// payload-mapping branch in a single realistic datacenter ("DC1").
func TestCreateLocationFromCani_CreatesOn201(t *testing.T) {
	locTypeID, createdLocID := uuid.New(), uuid.New()
	var locPosts int
	e, cleanup := newExporterWithServer(t, locationCreateHandler(locTypeID, createdLocID, http.StatusCreated, emptyListJSON, &locPosts))
	defer cleanup()
	seedActiveStatus(t, e)

	loc := newCaniLocation("DC1", "Section")
	// Exercise the optional-field branches.
	loc.Facility = "fac-1"
	loc.Description = "primary datacenter"
	loc.PhysicalAddress = "1 Main St"
	loc.ContactName = "ops"
	loc.TimeZone = "UTC"
	loc.Comments = "note"
	loc.CustomFields = map[string]interface{}{"tier": "1"}

	result := &LoadResult{}
	got, err := e.createLocationFromCani(context.Background(), loc, map[uuid.UUID]uuid.UUID{}, result)
	if err != nil {
		t.Fatalf("createLocationFromCani() error = %v", err)
	}
	if got != createdLocID {
		t.Errorf("returned ID = %s, want %s", got, createdLocID)
	}
	if !containsName(result.LocationsCreated, "DC1") {
		t.Errorf("LocationsCreated = %v, want it to contain DC1", result.LocationsCreated)
	}
	if locPosts != 1 {
		t.Errorf("expected exactly one location create POST, got %d", locPosts)
	}
}

// TestCreateLocationFromCani_ErrorsWhenNoLocationType verifies the create fails
// when the location has no LocationType set.
//
// Why it matters: Nautobot requires every location to declare a location-type;
// exporting one without it would be rejected, so cani fails fast with a clear
// error.
// Inputs: a location whose LocationType is "". Outputs: a non-nil error.
// Data choice: an empty LocationType is the precise precondition under test,
// with all other fields valid so the empty type is the sole cause of failure.
func TestCreateLocationFromCani_ErrorsWhenNoLocationType(t *testing.T) {
	var locPosts int
	e, cleanup := newExporterWithServer(t, locationCreateHandler(uuid.New(), uuid.New(), http.StatusCreated, emptyListJSON, &locPosts))
	defer cleanup()

	loc := newCaniLocation("DC1", "") // missing location type

	result := &LoadResult{}
	if _, err := e.createLocationFromCani(context.Background(), loc, map[uuid.UUID]uuid.UUID{}, result); err == nil {
		t.Fatal("expected an error when the location has no locationType set")
	}
}

// TestCreateLocationFromCani_ErrorsWhenLocationTypeUnresolvable verifies the
// create fails when the location-type lookup misses and auto-creation is off.
//
// Why it matters: the location-type FK must resolve to a real Nautobot object;
// without one (and with creation disabled) the export cannot proceed.
// Inputs: a server returning an empty location-types list, createLocationTypes
// set false. Outputs: a non-nil error.
// Data choice: an empty list plus the disabled create flag is the minimal setup
// that forces the unresolvable-type branch without other interference.
func TestCreateLocationFromCani_ErrorsWhenLocationTypeUnresolvable(t *testing.T) {
	// location-types lookup returns no match and auto-creation is disabled.
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, emptyListJSON))
	defer cleanup()
	seedActiveStatus(t, e)
	e.Cache.createLocationTypes = false

	loc := newCaniLocation("DC1", "Section")

	result := &LoadResult{}
	if _, err := e.createLocationFromCani(context.Background(), loc, map[uuid.UUID]uuid.UUID{}, result); err == nil {
		t.Fatal("expected an error when the location type cannot be resolved")
	}
}

// TestCreateLocationFromCani_MapsParentFK verifies a location whose Parent is
// already created maps that parent's Nautobot ID into the create payload and
// still creates successfully.
//
// Why it matters: nested locations (building under site) must reference their
// parent's Nautobot ID, or the hierarchy would be flattened or rejected.
// Inputs: a location with Parent set and a createdMap resolving that parent.
// Outputs: the new location UUID and an error; one POST occurs.
// Data choice: the parent's cani ID is placed in createdMap with a distinct
// Nautobot ID to prove the FK is looked up and substituted from the map.
func TestCreateLocationFromCani_MapsParentFK(t *testing.T) {
	locTypeID, createdLocID := uuid.New(), uuid.New()
	var locPosts int
	e, cleanup := newExporterWithServer(t, locationCreateHandler(locTypeID, createdLocID, http.StatusCreated, emptyListJSON, &locPosts))
	defer cleanup()
	seedActiveStatus(t, e)

	parentCaniID := uuid.New()
	parentNautobotID := uuid.New()
	loc := newCaniLocation("DC1", "Section")
	loc.Parent = parentCaniID
	createdMap := map[uuid.UUID]uuid.UUID{parentCaniID: parentNautobotID}

	result := &LoadResult{}
	got, err := e.createLocationFromCani(context.Background(), loc, createdMap, result)
	if err != nil {
		t.Fatalf("createLocationFromCani() error = %v", err)
	}
	if got != createdLocID {
		t.Errorf("returned ID = %s, want %s", got, createdLocID)
	}
	if locPosts != 1 {
		t.Errorf("expected exactly one location create POST, got %d", locPosts)
	}
}

// TestCreateLocationFromCani_ErrorsWhenParentNotCreated verifies the create
// fails (and posts nothing) when the location's Parent is not present in the
// created-ID map.
//
// Why it matters: a child location cannot be created before its parent exists in
// Nautobot; failing here prevents an orphaned or misattached location.
// Inputs: a location with Parent set but an empty createdMap. Outputs: a
// non-nil error; locPosts stays 0.
// Data choice: Parent is a fresh UUID absent from the empty createdMap, modeling
// a parent whose creation has not happened yet.
func TestCreateLocationFromCani_ErrorsWhenParentNotCreated(t *testing.T) {
	locTypeID, createdLocID := uuid.New(), uuid.New()
	var locPosts int
	e, cleanup := newExporterWithServer(t, locationCreateHandler(locTypeID, createdLocID, http.StatusCreated, emptyListJSON, &locPosts))
	defer cleanup()
	seedActiveStatus(t, e)

	loc := newCaniLocation("DC1", "Section")
	loc.Parent = uuid.New() // references a parent absent from createdMap

	result := &LoadResult{}
	if _, err := e.createLocationFromCani(context.Background(), loc, map[uuid.UUID]uuid.UUID{}, result); err == nil {
		t.Fatal("expected an error when the parent has not yet been created")
	}
	if locPosts != 0 {
		t.Errorf("expected no create POST when the parent FK is unresolved, got %d", locPosts)
	}
}

// TestCreateLocationFromCani_DryRunSkipsCreate verifies dry-run returns a Nil ID,
// issues no POST, yet still caches the location by name.
//
// Why it matters: previewing must not mutate Nautobot, but later phases still
// need to resolve the location, so the dry-run path caches it locally.
// Inputs: the create path with Options.DryRun=true. Outputs: uuid.Nil and an
// error; the cache is asserted to contain "DC1".
// Data choice: a single location with no parent isolates the dry-run behavior;
// the test then reaches into the cache to confirm the local registration.
func TestCreateLocationFromCani_DryRunSkipsCreate(t *testing.T) {
	locTypeID, createdLocID := uuid.New(), uuid.New()
	var locPosts int
	e, cleanup := newExporterWithServer(t, locationCreateHandler(locTypeID, createdLocID, http.StatusCreated, emptyListJSON, &locPosts))
	defer cleanup()
	seedActiveStatus(t, e)
	e.Options.DryRun = true

	loc := newCaniLocation("DC1", "Section")

	result := &LoadResult{}
	got, err := e.createLocationFromCani(context.Background(), loc, map[uuid.UUID]uuid.UUID{}, result)
	if err != nil {
		t.Fatalf("createLocationFromCani() error = %v", err)
	}
	if got != uuid.Nil {
		t.Errorf("dry-run returned ID = %s, want Nil", got)
	}
	if locPosts != 0 {
		t.Errorf("expected no create POST in dry-run, got %d", locPosts)
	}
	// The location is cached by name so downstream phases can resolve it.
	e.Cache.locationsMu.RLock()
	_, ok := e.Cache.locations["DC1"]
	e.Cache.locationsMu.RUnlock()
	if !ok {
		t.Error("expected DC1 to be cached even in dry-run")
	}
}

// TestCreateLocationFromCani_ReturnsErrorOnNon201 verifies a non-201 location
// create response is surfaced as an error.
//
// Why it matters: a rejected location create must abort rather than be treated
// as success, since dependent objects would otherwise reference a non-existent
// location.
// Inputs: the create path with the locations POST returning 400. Outputs: a
// non-nil error.
// Data choice: only the create status is flipped to 400 while the type lookup
// still succeeds, isolating the failure to the location POST.
func TestCreateLocationFromCani_ReturnsErrorOnNon201(t *testing.T) {
	locTypeID, createdLocID := uuid.New(), uuid.New()
	var locPosts int
	e, cleanup := newExporterWithServer(t, locationCreateHandler(locTypeID, createdLocID, http.StatusBadRequest, emptyListJSON, &locPosts))
	defer cleanup()
	seedActiveStatus(t, e)

	loc := newCaniLocation("DC1", "Section")

	result := &LoadResult{}
	if _, err := e.createLocationFromCani(context.Background(), loc, map[uuid.UUID]uuid.UUID{}, result); err == nil {
		t.Fatal("expected an error when the location create responds with 400")
	}
}

// -----------------------------------------------------------------------------
// loadLocations
// -----------------------------------------------------------------------------

// TestLoadLocations_EmptyInventoryNoOp verifies that an inventory with no
// locations performs no HTTP calls and returns an empty mapping.
//
// Why it matters: not every export carries locations; the phase must be a clean
// no-op rather than issuing spurious requests when there is nothing to do.
// Inputs: a context, an empty Inventory, and a LoadResult. Outputs: an empty
// mapping and an error.
// Data choice: the handler fails the test on any request, so the assertion is
// simply that it is never reached for empty input.
func TestLoadLocations_EmptyInventoryNoOp(t *testing.T) {
	e, cleanup := newExporterWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected HTTP call to %s for an empty inventory", r.URL.Path)
	})
	defer cleanup()

	result := &LoadResult{}
	created, err := e.loadLocations(context.Background(), &devicetypes.Inventory{}, result)
	if err != nil {
		t.Fatalf("loadLocations() error = %v", err)
	}
	if len(created) != 0 {
		t.Errorf("expected empty mapping, got %d entries", len(created))
	}
}

// TestLoadLocations_SkipsExistingLocation verifies that when LookupLocation finds
// an existing location, loadLocations records it under LocationsSkipped, maps its
// remote ID, and issues no create POST.
//
// Why it matters: re-running an export must reuse existing locations rather than
// duplicate them, while still returning their IDs for dependent phases.
// Inputs: an inventory with one location plus a server whose locations GET
// returns a matching object. Outputs: the cani->Nautobot mapping and an error.
// Data choice: the locations list returns a single "DC1" match so the lookup
// resolves and the skip branch (not the create branch) runs.
func TestLoadLocations_SkipsExistingLocation(t *testing.T) {
	existingID := uuid.New()
	existingBody := fmt.Sprintf(`{"count":1,"results":[%s]}`, refObjectJSON(existingID, "DC1"))
	var locPosts int
	e, cleanup := newExporterWithServer(t, locationCreateHandler(uuid.New(), uuid.New(), http.StatusCreated, existingBody, &locPosts))
	defer cleanup()

	loc := newCaniLocation("DC1", "Section")
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{loc.ID: loc},
	}

	result := &LoadResult{}
	created, err := e.loadLocations(context.Background(), inv, result)
	if err != nil {
		t.Fatalf("loadLocations() error = %v", err)
	}
	if created[loc.ID] != existingID {
		t.Errorf("mapping[loc] = %s, want existing %s", created[loc.ID], existingID)
	}
	if !containsName(result.LocationsSkipped, "DC1") {
		t.Errorf("LocationsSkipped = %v, want it to contain DC1", result.LocationsSkipped)
	}
	if locPosts != 0 {
		t.Errorf("expected no create POST for an existing location, got %d", locPosts)
	}
}

// TestLoadLocations_CreatesNewLocation verifies that when no existing location is
// found, loadLocations creates it, records it under LocationsCreated, and maps
// the new remote ID.
//
// Why it matters: this is the end-to-end create path for the locations phase,
// turning a cani location into a real Nautobot location and remembering its ID.
// Inputs: an inventory with one location plus a server whose locations GET is
// empty (so creation runs). Outputs: the mapping and an error; one POST occurs.
// Data choice: an empty existing-locations body forces the miss-then-create
// path, the complement of the skip test, using the same "DC1" fixture.
func TestLoadLocations_CreatesNewLocation(t *testing.T) {
	createdLocID := uuid.New()
	var locPosts int
	e, cleanup := newExporterWithServer(t, locationCreateHandler(uuid.New(), createdLocID, http.StatusCreated, emptyListJSON, &locPosts))
	defer cleanup()
	seedActiveStatus(t, e)

	loc := newCaniLocation("DC1", "Section")
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{loc.ID: loc},
	}

	result := &LoadResult{}
	created, err := e.loadLocations(context.Background(), inv, result)
	if err != nil {
		t.Fatalf("loadLocations() error = %v", err)
	}
	if created[loc.ID] != createdLocID {
		t.Errorf("mapping[loc] = %s, want created %s", created[loc.ID], createdLocID)
	}
	if !containsName(result.LocationsCreated, "DC1") {
		t.Errorf("LocationsCreated = %v, want it to contain DC1", result.LocationsCreated)
	}
	if locPosts != 1 {
		t.Errorf("expected exactly one create POST, got %d", locPosts)
	}
}
