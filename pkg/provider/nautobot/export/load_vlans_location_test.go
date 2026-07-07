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

// TestCreateVLAN_ScopesToLocation verifies that a known Nautobot location UUID is
// attached to the VLAN create payload.
//
// Why it matters: VLANs are scoped to a location in Nautobot; dropping the
// location would place the VLAN at the global scope and break per-location VID
// uniqueness the fabric relies on.
// Inputs: a VLAN plus a non-nil Nautobot location UUID. Outputs: a create POST
// whose body contains the location UUID.
// Data choice: passing the resolved location ID directly isolates the payload
// wiring from location-cache resolution.
func TestCreateVLAN_ScopesToLocation(t *testing.T) {
	resetIPAMCaches()
	vlanNID := uuid.New()
	var captured []byte
	e, cleanup := newExporterWithServer(t, vlanCreateServer(http.StatusCreated, vlanNID, &captured))
	defer cleanup()
	seedActiveStatus(t, e)

	locID := uuid.New()
	vlan := &devicetypes.CaniVLAN{VID: 100, Name: "vlan100"}
	vlan.Status = "Active"

	result := &LoadResult{}
	if _, err := e.createVLAN(context.Background(), vlan, locID, result); err != nil {
		t.Fatalf("createVLAN() error = %v", err)
	}
	if !strings.Contains(string(captured), locID.String()) {
		t.Errorf("VLAN payload missing location id %s:\n%s", locID, captured)
	}
}

// TestCreateVLAN_RetriesWithoutLocationOn400 verifies the graceful fallback: when
// Nautobot rejects the location-scoped create with 400, the VLAN is re-created
// without the location.
//
// Why it matters: a location type that lacks the ipam.vlan content type would
// otherwise fail the whole VLAN; the fallback keeps the export working (minus
// scoping) and warns instead of aborting.
// Inputs: a handler that 400s any VLAN POST carrying a location and 201s one
// without. Outputs: the VLAN's ID, exactly two POST attempts, the first with and
// the second without a location.
// Data choice: switching on the literal "location" in the body mirrors the real
// rejection and lets the test assert the retry dropped the field.
func TestCreateVLAN_RetriesWithoutLocationOn400(t *testing.T) {
	resetIPAMCaches()
	vlanNID := uuid.New()
	locID := uuid.New()
	var bodies [][]byte
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "ipam/vlans") {
			body, _ := io.ReadAll(r.Body)
			bodies = append(bodies, body)
			if strings.Contains(string(body), locID.String()) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = io.WriteString(w, `{"location":["not allowed for this location type"]}`)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, fmt.Sprintf(`{"id":%q,"vid":100,"name":"vlan100"}`, vlanNID.String()))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	seedActiveStatus(t, e)

	vlan := &devicetypes.CaniVLAN{VID: 100, Name: "vlan100"}
	vlan.Status = "Active"

	result := &LoadResult{}
	got, err := e.createVLAN(context.Background(), vlan, locID, result)
	if err != nil {
		t.Fatalf("createVLAN() error = %v", err)
	}
	if got != vlanNID {
		t.Errorf("returned ID = %s, want %s", got, vlanNID)
	}
	if len(bodies) != 2 {
		t.Fatalf("expected 2 POST attempts (with then without location), got %d", len(bodies))
	}
	if !strings.Contains(string(bodies[0]), locID.String()) {
		t.Error("first attempt should include the location")
	}
	if strings.Contains(string(bodies[1]), locID.String()) {
		t.Error("retry should omit the location")
	}
}

// TestCreateVLAN_RetriesWithoutLocationOn500 verifies the fallback also triggers
// when Nautobot answers the location-scoped create with a 500 rather than a 400.
//
// Why it matters: Nautobot's location-type validation can crash (HTTP 500)
// instead of rejecting cleanly; without treating 500 as a location rejection the
// VLAN would fail the whole export. The retry keeps the VLAN exporting unscoped.
// Inputs: a handler that 500s any VLAN POST carrying a location and 201s one
// without. Outputs: the VLAN's ID and exactly two POST attempts, the second
// dropping the location.
// Data choice: 500 mirrors the observed server-side crash for VLAN/prefix
// locations whose type lacks the ipam content type.
func TestCreateVLAN_RetriesWithoutLocationOn500(t *testing.T) {
	resetIPAMCaches()
	vlanNID := uuid.New()
	locID := uuid.New()
	var bodies [][]byte
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "ipam/vlans") {
			body, _ := io.ReadAll(r.Body)
			bodies = append(bodies, body)
			if strings.Contains(string(body), locID.String()) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = io.WriteString(w, `<!DOCTYPE html><html>Server Error</html>`)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, fmt.Sprintf(`{"id":%q,"vid":100,"name":"vlan100"}`, vlanNID.String()))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	seedActiveStatus(t, e)

	vlan := &devicetypes.CaniVLAN{VID: 100, Name: "vlan100"}
	vlan.Status = "Active"

	result := &LoadResult{}
	got, err := e.createVLAN(context.Background(), vlan, locID, result)
	if err != nil {
		t.Fatalf("createVLAN() error = %v", err)
	}
	if got != vlanNID {
		t.Errorf("returned ID = %s, want %s", got, vlanNID)
	}
	if len(bodies) != 2 {
		t.Fatalf("expected 2 POST attempts (with then without location), got %d", len(bodies))
	}
	if strings.Contains(string(bodies[1]), locID.String()) {
		t.Error("retry should omit the location")
	}
}
