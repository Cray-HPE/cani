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

	"github.com/google/uuid"
)

// deviceObjectJSON renders a single Nautobot device object (as returned by the
// retrieve endpoint).
func deviceObjectJSON(id uuid.UUID, name string) string {
	return fmt.Sprintf(`{"id":%q,"name":%q,"display":%q}`, id.String(), name, name)
}

// -----------------------------------------------------------------------------
// fetchFullDeviceByID — GET /dcim/devices/{id}/
// -----------------------------------------------------------------------------

// TestFetchFullDeviceByID_ReturnsDevice verifies fetchFullDeviceByID returns the
// decoded Device when the Nautobot retrieve endpoint answers 200.
//
// Why it matters: the --merge diff path fetches the full remote device by UUID
// before comparing fields, and looking up by ID (not name) avoids acting on the
// wrong record when several Nautobot devices share a name.
// Inputs: a device UUID; the fake server returns one device object. Outputs: a
// *nautobotapi.Device whose Name is read back from the JSON200 payload.
// Data choice: name "compute-001" is an ordinary export target, confirming the
// happy path threads the response body through to the caller.
func TestFetchFullDeviceByID_ReturnsDevice(t *testing.T) {
	id := uuid.New()
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, deviceObjectJSON(id, "compute-001")))
	defer cleanup()

	dev, err := e.fetchFullDeviceByID(context.Background(), id)
	if err != nil {
		t.Fatalf("fetchFullDeviceByID() error = %v", err)
	}
	if dev == nil || ptrStr(dev.Name) != "compute-001" {
		t.Errorf("expected device compute-001, got %+v", dev)
	}
}

// TestFetchFullDeviceByID_ReturnsErrorOnNon200 verifies fetchFullDeviceByID
// returns an error when the retrieve endpoint answers 404.
//
// Why it matters: a missing remote device must not be treated as an empty match
// during a merge diff; surfacing the error stops the exporter from mutating or
// skipping data based on a record that isn't there.
// Inputs: a random UUID; the fake server replies 404 with an empty body.
// Outputs: a non-nil error and no Device.
// Data choice: 404 models the common "device was deleted in Nautobot" case,
// distinct from a transient server fault.
func TestFetchFullDeviceByID_ReturnsErrorOnNon200(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusNotFound, `{}`))
	defer cleanup()

	if _, err := e.fetchFullDeviceByID(context.Background(), uuid.New()); err == nil {
		t.Fatal("expected an error when device retrieve responds with 404")
	}
}

// TestFetchFullDeviceByID_ReturnsErrorOnServerError verifies fetchFullDeviceByID
// returns an error when the retrieve endpoint answers 500.
//
// Why it matters: a Nautobot outage during a merge diff must abort the lookup
// rather than silently proceed, so the export never overwrites remote state from
// incomplete data.
// Inputs: a random UUID; the fake server replies 500 with an empty body.
// Outputs: a non-nil error and no Device.
// Data choice: 500 represents a transient server fault, the complementary
// failure mode to the 404 "not found" case.
func TestFetchFullDeviceByID_ReturnsErrorOnServerError(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusInternalServerError, `{}`))
	defer cleanup()

	if _, err := e.fetchFullDeviceByID(context.Background(), uuid.New()); err == nil {
		t.Fatal("expected an error when device retrieve responds with 500")
	}
}

// -----------------------------------------------------------------------------
// fetchFullDevice — GET /dcim/devices/?name=
// -----------------------------------------------------------------------------

// TestFetchFullDevice_ReturnsFirstMatch verifies fetchFullDevice returns the
// first result from the name-filtered device list endpoint.
//
// Why it matters: when only a device name is known, the merge diff fetches the
// full remote object via /dcim/devices/?name=; returning results[0] gives the
// comparison a concrete record to diff against.
// Inputs: name "spine-01"; the fake server returns a one-element results list.
// Outputs: a *nautobotapi.Device whose Name matches the requested name.
// Data choice: a single-match list isolates the first-result selection without
// the complication of duplicate names.
func TestFetchFullDevice_ReturnsFirstMatch(t *testing.T) {
	id := uuid.New()
	body := fmt.Sprintf(`{"count":1,"results":[%s]}`, deviceObjectJSON(id, "spine-01"))
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, body))
	defer cleanup()

	dev, err := e.fetchFullDevice(context.Background(), "spine-01")
	if err != nil {
		t.Fatalf("fetchFullDevice() error = %v", err)
	}
	if dev == nil || ptrStr(dev.Name) != "spine-01" {
		t.Errorf("expected device spine-01, got %+v", dev)
	}
}

// TestFetchFullDevice_ReturnsErrorWhenAbsent verifies fetchFullDevice returns an
// error when the name filter matches zero devices.
//
// Why it matters: an empty result set means the device is not in Nautobot, and
// the merge diff must report that instead of dereferencing an empty list or
// diffing against nothing.
// Inputs: name "ghost"; the fake server returns count:0 with an empty results
// array. Outputs: a non-nil error and no Device.
// Data choice: an explicit empty list (not an HTTP error) exercises the
// zero-length guard separately from transport failures.
func TestFetchFullDevice_ReturnsErrorWhenAbsent(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{"count":0,"results":[]}`))
	defer cleanup()

	if _, err := e.fetchFullDevice(context.Background(), "ghost"); err == nil {
		t.Fatal("expected an error when no device matches the name")
	}
}

// TestFetchFullDevice_ReturnsErrorOnNon200 verifies fetchFullDevice returns an
// error when the device list endpoint answers 500.
//
// Why it matters: a failed list query during a merge diff must propagate so the
// exporter doesn't mistake an outage for "device absent" and take the wrong
// branch.
// Inputs: name "spine-01"; the handler returns 500 only for dcim/devices paths
// and 200 elsewhere. Outputs: a non-nil error and no Device.
// Data choice: the path-scoped handler returns 500 just for the list call,
// proving the error originates from that specific request.
func TestFetchFullDevice_ReturnsErrorOnNon200(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "dcim/devices") {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = io.WriteString(w, `{}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	if _, err := e.fetchFullDevice(context.Background(), "spine-01"); err == nil {
		t.Fatal("expected an error when the device list responds with 500")
	}
}
