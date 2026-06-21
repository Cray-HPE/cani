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
	"net/http/httptest"
	"testing"
)

// TestTestConnection_SucceedsOn200 verifies a connectivity probe returns nil and
// issues exactly one request when Nautobot answers 200 OK.
//
// Why it matters: the exporter probes connectivity before pushing data, so a
// healthy Nautobot must be recognized as reachable without wasting extra round
// trips.
// Inputs: a context. Outputs: an error (nil on success); the request count is
// observed through the handler.
// Data choice: an empty result list is the smallest valid Nautobot list
// response, proving the probe checks only reachability, not payload contents.
func TestTestConnection_SucceedsOn200(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{"count":0,"results":[]}`))
	defer cleanup()

	if err := e.Client.TestConnection(context.Background()); err != nil {
		t.Fatalf("TestConnection() error = %v, want nil", err)
	}
	if calls != 1 {
		t.Errorf("expected exactly one connectivity request, got %d", calls)
	}
}

// TestTestConnection_ReturnsErrorOnNon200 verifies the probe reports an error
// when Nautobot rejects the request with a 403.
//
// Why it matters: bad credentials or insufficient permissions must abort an
// export early rather than let later writes fail unpredictably.
// Inputs: a context. Outputs: a non-nil error.
// Data choice: 403 Forbidden with a {"detail":"denied"} body mirrors how
// Nautobot rejects an unauthorized token, the most common connectivity failure.
func TestTestConnection_ReturnsErrorOnNon200(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusForbidden, `{"detail":"denied"}`))
	defer cleanup()

	if err := e.Client.TestConnection(context.Background()); err == nil {
		t.Fatal("expected an error when Nautobot responds with 403")
	}
}

// TestTestConnection_ReturnsErrorOnTransportFailure verifies the probe reports an
// error when the Nautobot host cannot be reached at all.
//
// Why it matters: network, DNS, or host-down failures must be surfaced clearly
// so an operator knows the export never started, distinct from an auth rejection.
// Inputs: a context. Outputs: a non-nil error.
// Data choice: the test server is closed before the request so the dial fails at
// the transport layer, simulating an unreachable Nautobot without relying on a
// real unroutable address.
func TestTestConnection_ReturnsErrorOnTransportFailure(t *testing.T) {
	// Build a client against a server that is immediately shut down so the
	// connection attempt fails at the transport layer.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	client, err := NewNautobotClient(srv.URL, "test-token")
	if err != nil {
		srv.Close()
		t.Fatalf("NewNautobotClient: %v", err)
	}
	srv.Close() // close before the request so the dial fails

	if err := client.TestConnection(context.Background()); err == nil {
		t.Fatal("expected a transport error when the server is unreachable")
	}
}
