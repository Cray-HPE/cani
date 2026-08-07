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

// locationMergeHandler simulates Nautobot for the merge flow:
// GET /dcim/locations/{id}/ returns a location with old values,
// PATCH /dcim/locations/{id}/ accepts the merge.
func locationMergeHandler(locID uuid.UUID, remoteJSON string, patchCalls *int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "dcim/locations/"+locID.String()) {
			if r.Method == http.MethodPatch {
				*patchCalls++
				w.WriteHeader(http.StatusOK)
				_, _ = io.WriteString(w, `{}`)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, remoteJSON)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}
}

func TestMergeLocation_PatchesWhenDrifted(t *testing.T) {
	locID := uuid.New()
	remoteJSON := fmt.Sprintf(`{"id":%q,"name":"DC1","description":"old","location_type":{"id":"%s"},"status":{"id":"%s"}}`,
		locID, uuid.New(), uuid.New())
	var patchCalls int
	e, cleanup := newExporterWithServer(t, locationMergeHandler(locID, remoteJSON, &patchCalls))
	defer cleanup()
	e.Options.Merge = true

	loc := &devicetypes.CaniLocationType{
		ID:          uuid.New(),
		Name:        "DC1",
		Description: "new",
	}

	updated, err := e.mergeLocation(context.Background(), loc, locID)
	if err != nil {
		t.Fatalf("mergeLocation() error = %v", err)
	}
	if !updated {
		t.Error("expected updated=true when description drifted")
	}
	if patchCalls != 1 {
		t.Errorf("expected 1 PATCH, got %d", patchCalls)
	}
}

func TestMergeLocation_SkipsWhenNoDrift(t *testing.T) {
	locID := uuid.New()
	remoteJSON := fmt.Sprintf(`{"id":%q,"name":"DC1","location_type":{"id":"%s"},"status":{"id":"%s"}}`,
		locID, uuid.New(), uuid.New())
	var patchCalls int
	e, cleanup := newExporterWithServer(t, locationMergeHandler(locID, remoteJSON, &patchCalls))
	defer cleanup()
	e.Options.Merge = true

	loc := &devicetypes.CaniLocationType{
		ID:   uuid.New(),
		Name: "DC1",
	}

	updated, err := e.mergeLocation(context.Background(), loc, locID)
	if err != nil {
		t.Fatalf("mergeLocation() error = %v", err)
	}
	if updated {
		t.Error("expected updated=false when no drift")
	}
	if patchCalls != 0 {
		t.Errorf("expected 0 PATCH calls, got %d", patchCalls)
	}
}

func TestMergeLocation_DryRunDoesNotPatch(t *testing.T) {
	locID := uuid.New()
	remoteJSON := fmt.Sprintf(`{"id":%q,"name":"DC1","description":"old","location_type":{"id":"%s"},"status":{"id":"%s"}}`,
		locID, uuid.New(), uuid.New())
	var patchCalls int
	e, cleanup := newExporterWithServer(t, locationMergeHandler(locID, remoteJSON, &patchCalls))
	defer cleanup()
	e.Options.Merge = true
	e.Options.DryRun = true

	loc := &devicetypes.CaniLocationType{
		ID:          uuid.New(),
		Name:        "DC1",
		Description: "new",
	}

	updated, err := e.mergeLocation(context.Background(), loc, locID)
	if err != nil {
		t.Fatalf("mergeLocation() error = %v", err)
	}
	if !updated {
		t.Error("dry-run should still report updated=true")
	}
	if patchCalls != 0 {
		t.Errorf("dry-run should not PATCH, got %d", patchCalls)
	}
}

// loadLocationsHandler handles the full loadLocations flow: the location lookup
// (returns existing), the retrieve (returns remote state), and the PATCH.
func loadLocationsHandler(locID uuid.UUID, existingJSON, retrieveJSON string, patchCalls *int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "dcim/locations/"+locID.String()):
			if r.Method == http.MethodPatch {
				*patchCalls++
				w.WriteHeader(http.StatusOK)
				_, _ = io.WriteString(w, `{}`)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, retrieveJSON)
		case strings.Contains(r.URL.Path, "dcim/locations"):
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, existingJSON)
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, emptyListJSON)
		}
	}
}

func TestLoadLocations_MergeUpdatesWhenDrifted(t *testing.T) {
	locID := uuid.New()
	existingJSON := fmt.Sprintf(`{"count":1,"results":[{"id":%q,"name":"DC1","display":"DC1"}]}`, locID)
	retrieveJSON := fmt.Sprintf(`{"id":%q,"name":"DC1","description":"old","location_type":{"id":"%s"},"status":{"id":"%s"}}`,
		locID, uuid.New(), uuid.New())
	var patchCalls int
	e, cleanup := newExporterWithServer(t, loadLocationsHandler(locID, existingJSON, retrieveJSON, &patchCalls))
	defer cleanup()
	e.Options.Merge = true

	caniID := uuid.New()
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			caniID: {
				ID:           caniID,
				Name:         "DC1",
				LocationType: "Section",
				Description:  "new",
			},
		},
	}
	result := &LoadResult{
		LocationsCreated: make([]string, 0),
		LocationsUpdated: make([]string, 0),
		LocationsSkipped: make([]string, 0),
	}

	_, err := e.loadLocations(context.Background(), inv, result)
	if err != nil {
		t.Fatalf("loadLocations() error = %v", err)
	}
	if len(result.LocationsUpdated) != 1 || result.LocationsUpdated[0] != "DC1" {
		t.Errorf("LocationsUpdated = %v, want [DC1]", result.LocationsUpdated)
	}
	if len(result.LocationsSkipped) != 0 {
		t.Errorf("LocationsSkipped = %v, want empty", result.LocationsSkipped)
	}
	if patchCalls != 1 {
		t.Errorf("expected 1 PATCH, got %d", patchCalls)
	}
}

func TestLoadLocations_MergeSkipsWhenIdentical(t *testing.T) {
	locID := uuid.New()
	existingJSON := fmt.Sprintf(`{"count":1,"results":[{"id":%q,"name":"DC1","display":"DC1"}]}`, locID)
	retrieveJSON := fmt.Sprintf(`{"id":%q,"name":"DC1","location_type":{"id":"%s"},"status":{"id":"%s"}}`,
		locID, uuid.New(), uuid.New())
	var patchCalls int
	e, cleanup := newExporterWithServer(t, loadLocationsHandler(locID, existingJSON, retrieveJSON, &patchCalls))
	defer cleanup()
	e.Options.Merge = true

	caniID := uuid.New()
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			caniID: {
				ID:           caniID,
				Name:         "DC1",
				LocationType: "Section",
			},
		},
	}
	result := &LoadResult{
		LocationsCreated: make([]string, 0),
		LocationsUpdated: make([]string, 0),
		LocationsSkipped: make([]string, 0),
	}

	_, err := e.loadLocations(context.Background(), inv, result)
	if err != nil {
		t.Fatalf("loadLocations() error = %v", err)
	}
	if len(result.LocationsUpdated) != 0 {
		t.Errorf("LocationsUpdated = %v, want empty", result.LocationsUpdated)
	}
	if len(result.LocationsSkipped) != 1 {
		t.Errorf("LocationsSkipped = %v, want [DC1]", result.LocationsSkipped)
	}
	if patchCalls != 0 {
		t.Errorf("expected 0 PATCH calls, got %d", patchCalls)
	}
}

func TestLoadLocations_MergeDryRunIssuesNoMutations(t *testing.T) {
	locID := uuid.New()
	existingJSON := fmt.Sprintf(`{"count":1,"results":[{"id":%q,"name":"DC1","display":"DC1"}]}`, locID)
	retrieveJSON := fmt.Sprintf(`{"id":%q,"name":"DC1","description":"old","location_type":{"id":"%s"},"status":{"id":"%s"}}`,
		locID, uuid.New(), uuid.New())
	var patchCalls int
	e, cleanup := newExporterWithServer(t, loadLocationsHandler(locID, existingJSON, retrieveJSON, &patchCalls))
	defer cleanup()
	e.Options.Merge = true
	e.Options.DryRun = true

	caniID := uuid.New()
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			caniID: {
				ID:           caniID,
				Name:         "DC1",
				LocationType: "Section",
				Description:  "new",
			},
		},
	}
	result := &LoadResult{
		LocationsCreated: make([]string, 0),
		LocationsUpdated: make([]string, 0),
		LocationsSkipped: make([]string, 0),
	}

	_, err := e.loadLocations(context.Background(), inv, result)
	if err != nil {
		t.Fatalf("loadLocations() error = %v", err)
	}
	if patchCalls != 0 {
		t.Errorf("dry-run should issue 0 PATCH calls, got %d", patchCalls)
	}
	if len(result.LocationsUpdated) != 1 {
		t.Errorf("LocationsUpdated = %v, want [DC1]", result.LocationsUpdated)
	}
}
