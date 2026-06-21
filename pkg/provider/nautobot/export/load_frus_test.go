/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// TestTopologicalSortFrus verifies FRUs are ordered parents-before-children,
// that an empty map yields an empty slice, and that nil entries are skipped.
//
// Why it matters: FRUs (field-replaceable units) nest via parent references and
// Nautobot rejects a child inventory-item whose parent does not yet exist, so
// the export must create parents first.
// Inputs: a map of FRU ID -> *CaniFruType. Outputs: a slice in creation order.
// Data choice: a deliberately out-of-order map (child listed before parent)
// proves the sort reorders rather than echoing map order; the nil entry guards
// against panics from sparse inventory.
func TestTopologicalSortFrus(t *testing.T) {
	t.Run("parents come before children", func(t *testing.T) {
		parentID := uuid.New()
		childID := uuid.New()

		frus := map[uuid.UUID]*devicetypes.CaniFruType{
			childID: {
				ID:     childID,
				Name:   "child-fru",
				Parent: parentID,
			},
			parentID: {
				ID:   parentID,
				Name: "parent-fru",
			},
		}

		ordered := topologicalSortFrus(frus)

		if len(ordered) != 2 {
			t.Fatalf("expected 2 FRUs, got %d", len(ordered))
		}

		parentIdx := -1
		childIdx := -1
		for i, fru := range ordered {
			if fru.ID == parentID {
				parentIdx = i
			}
			if fru.ID == childID {
				childIdx = i
			}
		}

		if parentIdx > childIdx {
			t.Errorf("parent (idx=%d) should come before child (idx=%d)", parentIdx, childIdx)
		}
	})

	t.Run("empty map returns empty slice", func(t *testing.T) {
		frus := map[uuid.UUID]*devicetypes.CaniFruType{}

		ordered := topologicalSortFrus(frus)

		if len(ordered) != 0 {
			t.Errorf("expected 0 FRUs, got %d", len(ordered))
		}
	})

	t.Run("nil entries are skipped", func(t *testing.T) {
		id := uuid.New()
		frus := map[uuid.UUID]*devicetypes.CaniFruType{
			id:         {ID: id, Name: "real-fru"},
			uuid.New(): nil,
		}

		ordered := topologicalSortFrus(frus)

		if len(ordered) != 1 {
			t.Errorf("expected 1 FRU (nil skipped), got %d", len(ordered))
		}
	})
}
