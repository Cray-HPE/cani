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
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// rackWithID returns a rack that stores the given Nautobot ID.
func rackWithID(id uuid.UUID) *devicetypes.CaniRackType {
	return &devicetypes.CaniRackType{
		Name:       "rack-01",
		ObjectMeta: devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{externalIDKeyNautobot: id}},
	}
}

// TestPruneStaleRackID_DropsAnIDTheDestinationLacks verifies that a stored
// rack ID the destination answers 404 for is removed after one retrieve of
// that rack.
//
// Why it matters: an inventory imported from one Nautobot carries that
// Nautobot's rack IDs; trusted on a fresh destination, Phase 1 skips every
// rack as already exported and every device placed in one fails to create.
// Inputs: a rack storing an ID; a Nautobot that answers 404 and records the
// path it was asked for.
// Outputs: the ID is gone, and the one request was the rack's retrieve.
// Data choice: the recorded path pins that the rack endpoint, not the device
// one, decides.
func TestPruneStaleRackID_DropsAnIDTheDestinationLacks(t *testing.T) {
	stale := uuid.New()
	var paths []string
	e, cleanup := newExporterWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	})
	defer cleanup()
	rack := rackWithID(stale)

	e.pruneStaleRackID(context.Background(), rack)

	if _, ok := rack.ExternalIDs[externalIDKeyNautobot]; ok {
		t.Error("the stored ID survived a 404; Phase 1 would skip the rack as already exported")
	}
	if len(paths) != 1 || !strings.HasSuffix(paths[0], "/dcim/racks/"+stale.String()+"/") {
		t.Errorf("requests = %v, want one retrieve of the rack", paths)
	}
}

// TestPruneStaleRackID_KeepsAnIDTheDestinationHoldsOrCannotConfirm verifies
// that a stored rack ID survives a 200 and a 500, and that a rack without an
// ID asks nothing.
//
// Why it matters: a rack Nautobot holds must keep its ID so a repeated export
// reuses it, and a transient failure must never read as absence, which would
// re-resolve by name and could duplicate it.
// Inputs: racks storing an ID against servers answering 200 and 500, and a
// rack with no ID against a server that fails the test if called.
// Outputs: both IDs unchanged; no request for the rack without one.
// Data choice: 500 stands in for any status outside 200 and 404.
func TestPruneStaleRackID_KeepsAnIDTheDestinationHoldsOrCannotConfirm(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusInternalServerError} {
		var calls int
		e, cleanup := newExporterWithServer(t, jsonHandler(&calls, status, `{}`))
		id := uuid.New()
		rack := rackWithID(id)

		e.pruneStaleRackID(context.Background(), rack)

		cleanup()
		if rack.ExternalIDs[externalIDKeyNautobot] != id {
			t.Errorf("status %d: the stored ID was dropped", status)
		}
	}

	e, cleanup := newExporterWithServer(t, func(http.ResponseWriter, *http.Request) {
		t.Error("a rack with no stored ID made a request")
	})
	defer cleanup()
	e.pruneStaleRackID(context.Background(), &devicetypes.CaniRackType{Name: "rack-02"})
}
