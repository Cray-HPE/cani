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

// refObjectJSON renders a generic Nautobot reference object with id/name/display.
func refObjectJSON(id uuid.UUID, name string) string {
	return fmt.Sprintf(`{"id":%q,"name":%q,"display":%q}`, id.String(), name, name)
}

// locationServer routes the requests CreateLocation makes: the location-type
// lookup (which returns an existing "Section" type) and the location create.
func locationServer(locTypeID, locID uuid.UUID) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "dcim/location-types"):
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, fmt.Sprintf(`{"count":1,"results":[%s]}`,
				refObjectJSON(locTypeID, "Section")))
		case strings.Contains(r.URL.Path, "dcim/locations"):
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, refObjectJSON(locID, "DC1"))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `{}`)
		}
	}
}

// -----------------------------------------------------------------------------
// createLocationType — POST /dcim/location-types/
// -----------------------------------------------------------------------------

// TestCreateLocationType_CreatesWithDefaults verifies that createLocationType
// POSTs a new location type and returns the created item (201) when no
// definition is supplied, so the default content types are applied.
//
// Why it matters: auto-creating location types lets the exporter build the
// location hierarchy in an empty Nautobot without manual setup, which is a
// prerequisite for placing racks and devices.
// Inputs: name "section", nil *LocationTypeDefinition. Outputs: a *CachedItem
// with the new ID/name.
// Data choice: "section" is the default type CreateLocation relies on, and the
// nil definition exercises the default (dcim.device/rack) content-types branch.
func TestCreateLocationType_CreatesWithDefaults(t *testing.T) {
	ltID := uuid.New()
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusCreated, refObjectJSON(ltID, "section")))
	defer cleanup()

	item, err := e.Cache.createLocationType("section", nil)
	if err != nil {
		t.Fatalf("createLocationType() error = %v", err)
	}
	if item == nil || item.ID != ltID || item.Name != "section" {
		t.Errorf("expected location type %s, got %+v", ltID, item)
	}
}

// TestCreateLocationType_ReturnsErrorOnNon201 verifies that a non-201 (400)
// response from the location-type create is surfaced as an error.
//
// Why it matters: a failed location-type creation must abort the export rather
// than leave locations without a valid type to anchor them to.
// Inputs: name "section", nil definition; server replies 400. Outputs: a non-nil
// error.
// Data choice: 400 with a `{"detail":"bad"}` body mimics a Nautobot validation
// rejection; only the status code drives the error path.
func TestCreateLocationType_ReturnsErrorOnNon201(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusBadRequest, `{"detail":"bad"}`))
	defer cleanup()

	if _, err := e.Cache.createLocationType("section", nil); err == nil {
		t.Fatal("expected an error when location-type create responds with 400")
	}
}

// -----------------------------------------------------------------------------
// CreateLocation — resolves a location type + status, then POSTs the location.
// -----------------------------------------------------------------------------

// TestCreateLocation_CreatesAndCaches verifies that CreateLocation resolves a
// location type and an Active status, POSTs the location, returns it, and caches
// it by name.
//
// Why it matters: locations are the root dependency for racks and devices, so
// caching the created ID prevents duplicate locations on later resolves during
// the same export.
// Inputs: name "DC1"; the server routes the type lookup (existing "Section") and
// the location create (201). Outputs: a *CachedItem also stored in the locations
// cache.
// Data choice: seedActiveStatus supplies the required status and locationServer
// returns an existing type, focusing the test on the create-and-cache path.
func TestCreateLocation_CreatesAndCaches(t *testing.T) {
	locTypeID := uuid.New()
	locID := uuid.New()
	e, cleanup := newExporterWithServer(t, locationServer(locTypeID, locID))
	defer cleanup()
	seedActiveStatus(t, e)

	item, err := e.Cache.CreateLocation("DC1")
	if err != nil {
		t.Fatalf("CreateLocation() error = %v", err)
	}
	if item == nil || item.ID != locID || item.Name != "DC1" {
		t.Errorf("expected location %s, got %+v", locID, item)
	}

	e.Cache.locationsMu.RLock()
	cached, ok := e.Cache.locations["DC1"]
	e.Cache.locationsMu.RUnlock()
	if !ok || cached.ID != locID {
		t.Errorf("expected location DC1 cached, got %+v (ok=%v)", cached, ok)
	}
}

// TestCreateLocation_ReturnsErrorWhenLocationCreateFails verifies that a 400 from
// the location POST is surfaced as an error even though the location type
// resolved successfully.
//
// Why it matters: a failed location create must surface so dependent racks and
// devices are not exported against a parent location that does not exist.
// Inputs: name "DC1"; the type lookup succeeds and the location POST replies 400.
// Outputs: a non-nil error.
// Data choice: the handler resolves the type then 400s the location, and an
// Active status is seeded, so the failure is attributable solely to the location
// POST.
func TestCreateLocation_ReturnsErrorWhenLocationCreateFails(t *testing.T) {
	locTypeID := uuid.New()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "dcim/location-types") {
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, fmt.Sprintf(`{"count":1,"results":[%s]}`,
				refObjectJSON(locTypeID, "Section")))
			return
		}
		// The location POST fails.
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"detail":"bad"}`)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	seedActiveStatus(t, e)

	if _, err := e.Cache.CreateLocation("DC1"); err == nil {
		t.Fatal("expected an error when the location create responds with 400")
	}
}

// TestCreateLocation_ReturnsErrorWhenLocationTypeUnresolvable verifies that
// CreateLocation fails early when the location type cannot be resolved and
// auto-creation of location types is disabled.
//
// Why it matters: a location cannot be created without a location type, so
// failing fast avoids emitting a malformed location into the source-of-truth.
// Inputs: name "DC1"; the type list returns empty and createLocationTypes is
// false. Outputs: a non-nil error returned before any location POST.
// Data choice: an empty list plus disabled createLocationTypes forces the
// unresolved-type path, and seeding the Active status ensures the error is
// attributable to the type, not a missing status.
func TestCreateLocation_ReturnsErrorWhenLocationTypeUnresolvable(t *testing.T) {
	// The location-type lookup returns no match and auto-creation is disabled,
	// so the location type cannot be resolved and CreateLocation fails early.
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{"count":0,"results":[]}`))
	defer cleanup()
	seedActiveStatus(t, e)
	e.Cache.createLocationTypes = false

	if _, err := e.Cache.CreateLocation("DC1"); err == nil {
		t.Fatal("expected an error when the location type cannot be resolved")
	}
}
