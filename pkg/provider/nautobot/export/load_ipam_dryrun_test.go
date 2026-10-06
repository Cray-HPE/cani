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
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// TestDryRunPlansIPAMInNewNamespaceWithoutWriting verifies a dry run of a
// prefix and an address in a namespace Nautobot does not have yet sends no
// write and reports both as planned.
//
// Why it matters: --dry-run must never write to Nautobot, yet the new namespace
// cannot be queried either, because Nautobot answers 400 to a list filtered by
// an unknown namespace; the run must treat everything in it as absent.
// Inputs: 10.0.0.0/24 and its host 10.0.0.5/24 in "tenant-new", against a
// server that knows no namespace and rejects unknown namespace filters the way
// Nautobot 3.2 does. Outputs: one planned prefix and one planned address, no
// errors, no write request, and a single namespace lookup.
// Data choice: the 400 is what a guard that only suppressed the POST would
// still trip over.
func TestDryRunPlansIPAMInNewNamespaceWithoutWriting(t *testing.T) {
	// Arrange.
	var writes []string
	namespaceLookups := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method != http.MethodGet:
			writes = append(writes, r.Method+" "+r.URL.Path)
			w.WriteHeader(http.StatusCreated)
		case strings.Contains(r.URL.Path, "ipam/namespaces"):
			namespaceLookups++
			_, _ = io.WriteString(w, emptyListJSON)
		case r.URL.Query().Get("namespace") != "":
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"namespace":["Select a valid choice."]}`)
		default:
			_, _ = io.WriteString(w, emptyListJSON)
		}
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	seedActiveStatus(t, e)
	e.Options.DryRun = true
	e.Cache.SetDryRun(true)

	inv := devicetypes.NewInventory()
	prefixID, addressID := uuid.New(), uuid.New()
	inv.Prefixes[prefixID] = &devicetypes.CaniPrefix{ID: prefixID, Prefix: "10.0.0.0/24", PrefixLen: 24, Namespace: "tenant-new"}
	inv.IPAddresses[addressID] = &devicetypes.CaniIPAddress{ID: addressID, Host: "10.0.0.5", Address: "10.0.0.5/24", Parent: prefixID}
	result := &LoadResult{}

	// Act.
	created, err := e.loadPrefixes(context.Background(), inv, nil, nil, result)
	if err != nil {
		t.Fatalf("loadPrefixes returned error: %v", err)
	}
	if _, err := e.loadIPAddresses(context.Background(), inv, created, nil, result); err != nil {
		t.Fatalf("loadIPAddresses returned error: %v", err)
	}

	// Assert.
	if len(writes) != 0 {
		t.Errorf("dry run wrote to Nautobot: %v", writes)
	}
	if result.PrefixesCreated != 1 || result.IPAddressesCreated != 1 || len(result.Errors) != 0 {
		t.Errorf("planned prefixes = %d, addresses = %d, errors = %v; want 1, 1, none",
			result.PrefixesCreated, result.IPAddressesCreated, result.Errors)
	}
	if namespaceLookups != 1 {
		t.Errorf("namespace lookups = %d, want 1 (the plan is remembered)", namespaceLookups)
	}
}
