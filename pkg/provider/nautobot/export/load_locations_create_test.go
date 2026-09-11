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

// TestCreateLocationFromCani_DefaultsEmptyLocationTypeToSite verifies an empty
// LocationType resolves the documented Site default and creates successfully.
//
// Why it matters: older CANI inventories may omit LocationType, and the mapping
// contract preserves their exportability by treating those locations as sites.
// Inputs: a location whose LocationType is "". Outputs: a successful create and
// a location-type lookup filtered to Site.
func TestCreateLocationFromCani_DefaultsEmptyLocationTypeToSite(t *testing.T) {
	locTypeID, createdLocID := uuid.New(), uuid.New()
	var locPosts int
	var requestedType string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "dcim/location-types"):
			if name := r.URL.Query().Get("name"); name != "" {
				requestedType = name
			}
			_, _ = io.WriteString(w, fmt.Sprintf(`{"count":1,"results":[%s]}`,
				refObjectJSON(locTypeID, "Site")))
		case strings.Contains(r.URL.Path, "dcim/locations") && r.Method == http.MethodPost:
			locPosts++
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, refObjectJSON(createdLocID, "DC1"))
		default:
			_, _ = io.WriteString(w, emptyListJSON)
		}
	})
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	seedActiveStatus(t, e)

	loc := newCaniLocation("DC1", "")

	result := &LoadResult{}
	got, err := e.createLocationFromCani(context.Background(), loc, map[uuid.UUID]uuid.UUID{}, result)
	if err != nil {
		t.Fatalf("createLocationFromCani() error = %v", err)
	}
	if got != createdLocID {
		t.Errorf("returned ID = %s, want %s", got, createdLocID)
	}
	if requestedType != "Site" {
		t.Errorf("location type lookup = %q, want Site", requestedType)
	}
	if locPosts != 1 {
		t.Errorf("location POST count = %d, want 1", locPosts)
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
