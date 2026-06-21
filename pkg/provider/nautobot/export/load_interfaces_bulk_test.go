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

// bulkArrayServer answers the bulk interface POST (a JSON array body) with the
// supplied response. Any other request receives an empty result set.
func bulkArrayServer(postCalls *int, status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "dcim/interfaces") {
			*postCalls++
			w.WriteHeader(status)
			_, _ = w.Write([]byte(body))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"count":0,"results":[]}`))
	}
}

// bulkFallbackServer fails the bulk array POST but succeeds for the per-item
// object POSTs the exporter falls back to. It distinguishes the two by the
// shape of the request body: a JSON array starts with '['.
func bulkFallbackServer(individualBody string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "dcim/interfaces") {
			raw, _ := io.ReadAll(r.Body)
			if strings.HasPrefix(strings.TrimSpace(string(raw)), "[") {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"detail":"bulk create unsupported"}`))
				return
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(individualBody))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"count":0,"results":[]}`))
	}
}

// bulkItems builds n interface items, each on its own device, named eth0..ethN.
func bulkItems(n int) []bulkInterfaceItem {
	items := make([]bulkInterfaceItem, 0, n)
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("eth%d", i)
		items = append(items, bulkInterfaceItem{
			DeviceID:   uuid.New(),
			DeviceName: fmt.Sprintf("device-%d", i),
			Spec:       interfaceSpec{Name: name, Type: "1000base-t"},
		})
	}
	return items
}

// -----------------------------------------------------------------------------
// createInterfacesBulk
// -----------------------------------------------------------------------------

// TestCreateInterfacesBulk_CreatesWholeBatchAndCaches verifies the happy path:
// a batch is sent as a single bulk POST to /dcim/interfaces/, every returned
// interface is counted in result.IfacesCreated, and all are cached by
// device/name.
//
// Why it matters: bulk creation is the fast path for the interface phase — one
// round-trip per batch keeps large exports tractable and feeds later cabling.
// Inputs: a two-item batch; the server returns a JSON array of two interfaces
// with fresh IDs. Outputs: IfacesCreated == 2, exactly one POST, two cache hits.
// Data choice: two items on distinct devices confirm batching and per-device
// caching together.
func TestCreateInterfacesBulk_CreatesWholeBatchAndCaches(t *testing.T) {
	items := bulkItems(2)
	created := fmt.Sprintf(`[{"id":%q,"name":"eth0","display":"eth0"},{"id":%q,"name":"eth1","display":"eth1"}]`,
		uuid.NewString(), uuid.NewString())
	var postCalls int
	e, cleanup := newExporterWithServer(t, bulkArrayServer(&postCalls, http.StatusCreated, created))
	defer cleanup()
	seedActiveStatus(t, e)

	result := &LoadResult{}
	if err := e.createInterfacesBulk(context.Background(), items, result); err != nil {
		t.Fatalf("createInterfacesBulk() error = %v", err)
	}
	if result.IfacesCreated != 2 {
		t.Errorf("IfacesCreated = %d, want 2", result.IfacesCreated)
	}
	if postCalls != 1 {
		t.Errorf("expected a single bulk POST, got %d", postCalls)
	}

	// Both created interfaces should be cached by device/name for cable creation.
	for _, item := range items {
		e.Cache.interfacesMu.RLock()
		_, ok := e.Cache.interfaces[interfaceCacheKey(item.DeviceID, item.Spec.Name)]
		e.Cache.interfacesMu.RUnlock()
		if !ok {
			t.Errorf("expected interface %s on %s to be cached", item.Spec.Name, item.DeviceID)
		}
	}
}

// TestCreateInterfacesBulk_EmptyItemsNoOp verifies that a nil/empty batch
// performs no work: no HTTP POST and no change to result.IfacesCreated.
//
// Why it matters: the orchestrator may invoke the interface phase even when a
// run has no new interfaces, and doing so must not emit a spurious bulk request.
// Inputs: a nil items slice (Active status seeded so its resolve is a cache
// hit). Outputs: IfacesCreated == 0 and zero POSTs, tracked by a server counter.
// Data choice: nil is the simplest representation of "nothing to create" and
// exercises the early zero-batch path.
func TestCreateInterfacesBulk_EmptyItemsNoOp(t *testing.T) {
	var postCalls int
	e, cleanup := newExporterWithServer(t, bulkArrayServer(&postCalls, http.StatusCreated, `[]`))
	defer cleanup()
	seedActiveStatus(t, e)

	result := &LoadResult{}
	if err := e.createInterfacesBulk(context.Background(), nil, result); err != nil {
		t.Fatalf("createInterfacesBulk() error = %v", err)
	}
	if result.IfacesCreated != 0 || postCalls != 0 {
		t.Errorf("expected no work for an empty batch, got created=%d posts=%d", result.IfacesCreated, postCalls)
	}
}

// TestCreateInterfacesBulk_ErrorsWhenActiveStatusUnresolvable verifies the bulk
// path aborts with an error when the "Active" status cannot be resolved, before
// sending any interface.
//
// Why it matters: every interface POST must carry a valid status reference, so
// failing fast surfaces the misconfiguration instead of attempting status-less
// interfaces Nautobot would reject one by one.
// Inputs: a one-item batch with no seeded "Active" status and an empty status
// lookup. Outputs: a non-nil error from createInterfacesBulk.
// Data choice: deliberately omitting seedActiveStatus forces the status resolve
// at the top of the function to fail.
func TestCreateInterfacesBulk_ErrorsWhenActiveStatusUnresolvable(t *testing.T) {
	// No "Active" status is seeded and the lookup returns an empty list, so the
	// bulk path fails before any interface is sent.
	var postCalls int
	e, cleanup := newExporterWithServer(t, bulkArrayServer(&postCalls, http.StatusCreated, `[]`))
	defer cleanup()

	result := &LoadResult{}
	if err := e.createInterfacesBulk(context.Background(), bulkItems(1), result); err == nil {
		t.Fatal("expected an error when the Active status cannot be resolved")
	}
}

// TestCreateInterfacesBulk_FallsBackToIndividualOnBatchError verifies that when
// the bulk array POST fails, the exporter retries each item as an individual
// object POST and still records both as created.
//
// Why it matters: some Nautobot deployments reject bulk array creates, so the
// per-item fallback keeps the export working instead of losing the whole batch.
// Inputs: a two-item batch against a server that 500s the JSON-array POST but
// 201s object POSTs. Outputs: IfacesCreated == 2 with no returned error.
// Data choice: bulkFallbackServer branches on a leading '[' to tell the bulk
// array body apart from the fallback object bodies.
func TestCreateInterfacesBulk_FallsBackToIndividualOnBatchError(t *testing.T) {
	individual := fmt.Sprintf(`{"id":%q,"name":"eth0","display":"eth0"}`, uuid.NewString())
	e, cleanup := newExporterWithServer(t, bulkFallbackServer(individual))
	defer cleanup()
	seedActiveStatus(t, e)

	result := &LoadResult{}
	if err := e.createInterfacesBulk(context.Background(), bulkItems(2), result); err != nil {
		t.Fatalf("createInterfacesBulk() error = %v", err)
	}
	if result.IfacesCreated != 2 {
		t.Errorf("IfacesCreated = %d, want 2 (created individually after bulk failure)", result.IfacesCreated)
	}
}

// TestCreateInterfacesBulk_ReturnsErrorWhenIndividualAlsoFails verifies that
// when both the bulk POST and the per-item fallback fail, the aggregated errors
// are returned to the caller.
//
// Why it matters: a total failure of the interface phase must propagate so the
// run is reported as failed rather than silently dropping interfaces.
// Inputs: a two-item batch against a server that 500s every POST. Outputs: a
// non-nil error from createInterfacesBulk.
// Data choice: an unconditional 500 guarantees both the batch send and each
// individual retry fail, reaching the error-join branch.
func TestCreateInterfacesBulk_ReturnsErrorWhenIndividualAlsoFails(t *testing.T) {
	// Every POST fails, so the bulk send and the per-item fallback both error,
	// and the collected errors are returned.
	var postCalls int
	e, cleanup := newExporterWithServer(t, bulkArrayServer(&postCalls, http.StatusInternalServerError, `{"detail":"boom"}`))
	defer cleanup()
	seedActiveStatus(t, e)

	result := &LoadResult{}
	if err := e.createInterfacesBulk(context.Background(), bulkItems(2), result); err == nil {
		t.Fatal("expected an error when both bulk and individual creates fail")
	}
}
