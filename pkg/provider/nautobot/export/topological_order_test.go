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
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func fruNames(frus []*devicetypes.CaniFruType) string {
	names := make([]string, 0, len(frus))
	for _, f := range frus {
		names = append(names, f.Name)
	}
	return strings.Join(names, ",")
}

// TestTopologicalOrderIsStableAcrossRuns verifies that roots and siblings come
// out by name no matter how the input map iterates.
//
// Why it matters: FRU ordering used to follow Go's randomized map order, so two
// exports of the same inventory could create nested items in different
// sequences; a stable order makes export logs and dry runs comparable.
// Inputs: two roots and three children of one root, inserted in reverse name
// order. Outputs: the same name-sorted, parents-first sequence on every run.
// Data choice: 20 repetitions make a map-order dependence show up in practice.
func TestTopologicalOrderIsStableAcrossRuns(t *testing.T) {
	root := uuid.New()
	frus := map[uuid.UUID]*devicetypes.CaniFruType{
		root:       {ID: root, Name: "chassis-b"},
		uuid.New(): {ID: uuid.New(), Name: "chassis-a"},
	}
	for _, name := range []string{"psu-3", "psu-2", "psu-1"} {
		id := uuid.New()
		frus[id] = &devicetypes.CaniFruType{ID: id, Name: name, Parent: root}
	}
	const want = "chassis-a,chassis-b,psu-1,psu-2,psu-3"
	for i := 0; i < 20; i++ {
		ordered, unreachable := topologicalOrder(frus, fruHierarchy)
		if got := fruNames(ordered); got != want {
			t.Fatalf("run %d: order = %s, want %s", i, got, want)
		}
		if len(unreachable) != 0 {
			t.Fatalf("run %d: unexpected unreachable items %v", i, fruHierarchy.labels(unreachable))
		}
	}
}

// TestTopologicalOrderReportsUnreachable verifies that items in a parent cycle
// or under a parent that does not exist are returned as unreachable, by name,
// and left out of the ordered result.
//
// Why it matters: the loaders log these so an operator learns why an item was
// never created instead of seeing it vanish from the export.
// Inputs: one root, a two-item cycle and an orphan pointing at a missing parent.
// Outputs: only the root is ordered; the three others are listed by name.
// Data choice: the cycle and the orphan are the two ways a chain misses a root.
func TestTopologicalOrderReportsUnreachable(t *testing.T) {
	root, cycleA, cycleB := uuid.New(), uuid.New(), uuid.New()
	frus := map[uuid.UUID]*devicetypes.CaniFruType{
		root:       {ID: root, Name: "root"},
		cycleA:     {ID: cycleA, Name: "loop-a", Parent: cycleB},
		cycleB:     {ID: cycleB, Name: "loop-b", Parent: cycleA},
		uuid.New(): {Name: "orphan", Parent: uuid.New()},
	}
	ordered, unreachable := topologicalOrder(frus, fruHierarchy)
	if got := fruNames(ordered); got != "root" {
		t.Errorf("ordered = %s, want root", got)
	}
	if got := fruNames(unreachable); got != "loop-a,loop-b,orphan" {
		t.Errorf("unreachable = %s, want loop-a,loop-b,orphan", got)
	}
	labels := fruHierarchy.labels(unreachable[:1])
	if len(labels) != 1 || !strings.HasPrefix(labels[0], "loop-a (") {
		t.Errorf("labels = %v, want one \"loop-a (<id>)\" entry", labels)
	}
}
