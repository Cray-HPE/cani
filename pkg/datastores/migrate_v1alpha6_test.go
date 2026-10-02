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
package datastores

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// TestMigrateV1Alpha6StampsNamespaceScope verifies the v1alpha6-to-v1alpha7
// migration defaults omitted scope to Global, folds a uniquely resolvable
// legacy VRF name into VRF memberships, and preserves anything it cannot
// resolve or deduplicate.
//
// Why it matters: legacy single-namespace inventories must keep their meaning
// (Global) and their object UUIDs, while ambiguous VRF names and duplicate
// CIDRs are reported for explicit resolution instead of being guessed away.
// Inputs: a VRF and two prefixes without a namespace, one prefix naming the
// unique VRF "blue", one naming "red" which exists twice, plus a tenant-a VRF
// that must keep its name. Outputs: Global stamped only where blank, blue
// folded into VRFs with the name cleared, red retained as a name, both
// duplicate prefixes still present, and the v1alpha7 schema version.
// Data choice: the duplicate CIDR pair and the ambiguous VRF pair are the two
// legacy shapes the ticket forbids collapsing.
func TestMigrateV1Alpha6StampsNamespaceScope(t *testing.T) {
	blue, redOne, redTwo, tenant := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	pBlue, pRed := uuid.New(), uuid.New()
	inv := devicetypes.NewInventory()
	inv.SchemaVersion = devicetypes.SchemaVersionV1Alpha6
	inv.VRFs[blue] = &devicetypes.CaniVRF{ID: blue, Name: "blue"}
	inv.VRFs[redOne] = &devicetypes.CaniVRF{ID: redOne, Name: "red"}
	inv.VRFs[redTwo] = &devicetypes.CaniVRF{ID: redTwo, Name: "red"}
	inv.VRFs[tenant] = &devicetypes.CaniVRF{ID: tenant, Name: "blue", Namespace: "tenant-a"}
	inv.Prefixes[pBlue] = &devicetypes.CaniPrefix{ID: pBlue, Prefix: "10.0.0.0/24", VRF: "blue"}
	inv.Prefixes[pRed] = &devicetypes.CaniPrefix{ID: pRed, Prefix: "10.0.0.0/24", VRF: "red"}

	migrateV1Alpha6(inv)

	if inv.SchemaVersion != devicetypes.SchemaVersionV1Alpha7 {
		t.Errorf("SchemaVersion = %q, want %q", inv.SchemaVersion, devicetypes.SchemaVersionV1Alpha7)
	}
	if got := inv.VRFs[blue].Namespace; got != devicetypes.DefaultNamespace {
		t.Errorf("blue VRF namespace = %q, want %q", got, devicetypes.DefaultNamespace)
	}
	if got := inv.VRFs[tenant].Namespace; got != "tenant-a" {
		t.Errorf("tenant VRF namespace = %q, want tenant-a (must not be rewritten)", got)
	}
	if len(inv.Prefixes) != 2 {
		t.Fatalf("prefixes = %d, want both duplicate CIDRs preserved", len(inv.Prefixes))
	}
	resolved := inv.Prefixes[pBlue]
	if resolved.Namespace != devicetypes.DefaultNamespace || resolved.VRF != "" ||
		len(resolved.VRFs) != 1 || resolved.VRFs[0] != blue {
		t.Errorf("blue prefix = {ns %q, vrf %q, vrfs %v}, want Global, cleared name, [blue]",
			resolved.Namespace, resolved.VRF, resolved.VRFs)
	}
	ambiguous := inv.Prefixes[pRed]
	if ambiguous.VRF != "red" || len(ambiguous.VRFs) != 0 {
		t.Errorf("red prefix = {vrf %q, vrfs %v}, want the legacy name retained and no membership",
			ambiguous.VRF, ambiguous.VRFs)
	}
}

// TestMigrateV1Alpha6IsIdempotent verifies re-running the migration on an
// already scoped prefix neither duplicates the membership nor changes scope.
//
// Why it matters: the legacy chain applies every generation in order, so a
// prefix that already carries the membership must not accumulate copies.
// Inputs: a Global prefix whose VRFs already contain "blue" while the legacy
// name is still set. Outputs: one membership and a cleared name.
// Data choice: the pre-populated membership isolates the duplicate guard.
func TestMigrateV1Alpha6IsIdempotent(t *testing.T) {
	blue, prefixID := uuid.New(), uuid.New()
	inv := devicetypes.NewInventory()
	inv.VRFs[blue] = &devicetypes.CaniVRF{ID: blue, Name: "blue", Namespace: "Global"}
	inv.Prefixes[prefixID] = &devicetypes.CaniPrefix{
		ID: prefixID, Prefix: "10.0.0.0/24", Namespace: "Global", VRF: "blue", VRFs: []uuid.UUID{blue},
	}

	migrateV1Alpha6(inv)

	prefix := inv.Prefixes[prefixID]
	if prefix.VRF != "" || len(prefix.VRFs) != 1 {
		t.Errorf("prefix = {vrf %q, vrfs %v}, want cleared name and a single membership", prefix.VRF, prefix.VRFs)
	}
}
