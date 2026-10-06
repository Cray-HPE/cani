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
	"slices"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// tenantVRFInventory holds VRF "blue" and a member prefix, both in tenant-a.
func tenantVRFInventory() *devicetypes.Inventory {
	inv := devicetypes.NewInventory()
	vrfID, prefixID := uuid.New(), uuid.New()
	inv.VRFs[vrfID] = &devicetypes.CaniVRF{ID: vrfID, Name: "blue", Namespace: "tenant-a"}
	inv.Prefixes[prefixID] = &devicetypes.CaniPrefix{
		ID: prefixID, Prefix: "10.0.0.0/24", PrefixLen: 24, Namespace: "tenant-a", VRFs: []uuid.UUID{vrfID},
	}
	return inv
}

// exportVRFsThenPrefixes runs the VRF and prefix phases in Load's order.
func exportVRFsThenPrefixes(t *testing.T, e *Exporter, inv *devicetypes.Inventory) *LoadResult {
	t.Helper()
	result := &LoadResult{}
	if err := e.loadVRFs(context.Background(), inv, result); err != nil {
		t.Fatalf("loadVRFs returned error: %v", err)
	}
	if _, err := e.loadPrefixes(context.Background(), inv, nil, nil, result); err != nil {
		t.Fatalf("loadPrefixes returned error: %v", err)
	}
	return result
}

// TestLoadVRFs_CreatesNamespaceBeforeLookup verifies the first export of a VRF
// in a namespace Nautobot lacks creates the namespace before looking the VRF
// up, so the VRF and its prefix membership export in a single run.
//
// Why it matters: Nautobot answers 400 to a VRF list filtered by an unknown
// namespace; looking up first failed the VRF and skipped its prefix
// assignment until a second export.
// Inputs: VRF "blue" and a member prefix in "tenant-a", run in Load's phase
// order against a server where only Global exists. Outputs: one VRF and one
// VRF-prefix assignment created, no errors, and the namespace POST ahead of
// the VRF lookup.
// Data choice: a namespace missing from the server is exactly the case the old
// lookup order could not pass.
func TestLoadVRFs_CreatesNamespaceBeforeLookup(t *testing.T) {
	// Arrange.
	server := newNamespaceServer()
	e, cleanup := newExporterWithServer(t, server.handler)
	defer cleanup()
	seedActiveStatus(t, e)

	// Act.
	result := exportVRFsThenPrefixes(t, e, tenantVRFInventory())

	// Assert.
	if result.VRFsCreated != 1 || result.VRFPrefixAssignmentsCreated != 1 || len(result.Errors) != 0 {
		t.Fatalf("VRFs created = %d, assignments created = %d, errors = %v; want 1, 1, none",
			result.VRFsCreated, result.VRFPrefixAssignmentsCreated, result.Errors)
	}
	created := slices.IndexFunc(server.requests, func(r string) bool { return strings.HasPrefix(r, "POST /ipam/namespaces/") })
	lookup := slices.IndexFunc(server.requests, func(r string) bool { return strings.HasPrefix(r, "GET /ipam/vrfs/") })
	if created < 0 || lookup < created {
		t.Errorf("requests = %v, want the namespace created before the VRF lookup", server.requests)
	}
}

// TestLoadVRFs_DryRunPlansNewNamespace verifies a dry run plans a VRF in a new
// namespace, its member prefix and the membership without writing or failing.
//
// Why it matters: --dry-run must preview the same work a real run would do,
// and it may neither create the namespace nor query inside it.
// Inputs: the tenant-a VRF and prefix with DryRun set, against a server where
// only Global exists. Outputs: one VRF, one prefix and one assignment planned,
// no errors, and no write request.
// Data choice: the same inventory as the real-run test, so the two runs differ
// only by the dry-run switch.
func TestLoadVRFs_DryRunPlansNewNamespace(t *testing.T) {
	// Arrange.
	server := newNamespaceServer()
	e, cleanup := newExporterWithServer(t, server.handler)
	defer cleanup()
	seedActiveStatus(t, e)
	e.Options.DryRun = true
	e.Cache.SetDryRun(true)

	// Act.
	result := exportVRFsThenPrefixes(t, e, tenantVRFInventory())

	// Assert.
	if writes := server.writes(); len(writes) != 0 {
		t.Errorf("dry run wrote to Nautobot: %v", writes)
	}
	if result.VRFsCreated != 1 || result.PrefixesCreated != 1 || result.VRFPrefixAssignmentsCreated != 1 ||
		len(result.Errors) != 0 {
		t.Errorf("planned VRFs = %d, prefixes = %d, assignments = %d, errors = %v; want 1, 1, 1, none",
			result.VRFsCreated, result.PrefixesCreated, result.VRFPrefixAssignmentsCreated, result.Errors)
	}
}
