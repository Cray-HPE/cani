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
	"net/http"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

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
