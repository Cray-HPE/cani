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
	"net/http"
	"testing"

	"github.com/google/uuid"
)

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
