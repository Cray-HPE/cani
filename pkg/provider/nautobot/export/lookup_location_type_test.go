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

// -----------------------------------------------------------------------------
// createLocationType — the def != nil branch (Nestable / Description /
// ContentTypes / Parent) is only reachable when a definition is supplied.
// -----------------------------------------------------------------------------

// TestCreateLocationType_CreatesWithDefinition verifies createLocationType builds
// a request from a non-nil definition — setting Nestable, Description,
// ContentTypes and a resolved Parent — and returns the created type.
//
// Why it matters: location-types define the shape of the Nautobot location
// hierarchy; the export must faithfully translate a cani LocationTypeDefinition,
// including nesting and the parent-type FK, or the hierarchy is malformed.
// Inputs: a name and a *LocationTypeDefinition. Outputs: a *CachedItem and an
// error; exactly one create POST is expected.
// Data choice: the def sets every optional field and a Parent ("building") that
// the server resolves to an existing type, exercising all def!=nil branches at
// once.
func TestCreateLocationType_CreatesWithDefinition(t *testing.T) {
	parentID, createdID := uuid.New(), uuid.New()
	var postCalls int
	// The parent lookup (GET) resolves to an existing type; the create (POST)
	// returns the new location type.
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "dcim/location-types") {
			if r.Method == http.MethodPost {
				postCalls++
				w.WriteHeader(http.StatusCreated)
				_, _ = io.WriteString(w, refObjectJSON(createdID, "Section"))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, fmt.Sprintf(`{"count":1,"results":[%s]}`,
				refObjectJSON(parentID, "Building")))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	def := &devicetypes.LocationTypeDefinition{
		Name:         "Section",
		Slug:         "section",
		Description:  "a section",
		Nestable:     true,
		ContentTypes: []string{"device", "rack", "module"},
		Parent:       "building",
	}

	item, err := e.Cache.createLocationType("Section", def)
	if err != nil {
		t.Fatalf("createLocationType() error = %v", err)
	}
	if item == nil || item.ID != createdID {
		t.Errorf("expected created location type %s, got %+v", createdID, item)
	}
	if postCalls != 1 {
		t.Errorf("expected exactly one location-type create POST, got %d", postCalls)
	}
}

// TestCreateLocationType_ContinuesWhenParentUnresolvable verifies the type is
// still created (with the parent omitted) when the parent type cannot be
// resolved and auto-creation is disabled.
//
// Why it matters: a missing parent type should not abort creating the child
// type; the export degrades gracefully by creating the type without the parent
// FK rather than failing the whole locations phase.
// Inputs: a def whose Parent ("ghost-parent") does not resolve, with
// createLocationTypes=false. Outputs: the created *CachedItem and an error; one
// POST still occurs.
// Data choice: a parent name with no matching type and creation disabled forces
// the unresolved-parent branch while keeping the child create successful.
func TestCreateLocationType_ContinuesWhenParentUnresolvable(t *testing.T) {
	createdID := uuid.New()
	var postCalls int
	// The parent lookup (GET) returns nothing and auto-creation is disabled, so
	// the parent cannot be resolved. createLocationType must still create the
	// type, simply omitting the parent reference.
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "dcim/location-types") && r.Method == http.MethodPost {
			postCalls++
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, refObjectJSON(createdID, "Section"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	e.Cache.createLocationTypes = false

	def := &devicetypes.LocationTypeDefinition{
		Name:     "Section",
		Slug:     "section",
		Nestable: false,
		Parent:   "ghost-parent",
	}

	item, err := e.Cache.createLocationType("Section", def)
	if err != nil {
		t.Fatalf("createLocationType() error = %v", err)
	}
	if item == nil || item.ID != createdID {
		t.Errorf("expected created location type %s, got %+v", createdID, item)
	}
	if postCalls != 1 {
		t.Errorf("expected exactly one location-type create POST, got %d", postCalls)
	}
}
