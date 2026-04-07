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

func TestTopologicalSortLocations(t *testing.T) {
	t.Run("parents come before children", func(t *testing.T) {
		parentID := uuid.New()
		childID := uuid.New()

		locs := map[uuid.UUID]*devicetypes.CaniLocationType{
			childID: {
				ID:     childID,
				Name:   "Floor-1",
				Parent: parentID,
			},
			parentID: {
				ID:   parentID,
				Name: "Building-A",
			},
		}

		ordered := topologicalSortLocations(locs)

		if len(ordered) != 2 {
			t.Fatalf("expected 2 locations, got %d", len(ordered))
		}

		parentIdx := -1
		childIdx := -1
		for i, loc := range ordered {
			if loc.ID == parentID {
				parentIdx = i
			}
			if loc.ID == childID {
				childIdx = i
			}
		}

		if parentIdx > childIdx {
			t.Errorf("parent (idx=%d) should come before child (idx=%d)", parentIdx, childIdx)
		}
	})

	t.Run("empty map returns empty slice", func(t *testing.T) {
		locs := map[uuid.UUID]*devicetypes.CaniLocationType{}

		ordered := topologicalSortLocations(locs)

		if len(ordered) != 0 {
			t.Errorf("expected 0 locations, got %d", len(ordered))
		}
	})

	t.Run("nil entries are skipped", func(t *testing.T) {
		id := uuid.New()
		locs := map[uuid.UUID]*devicetypes.CaniLocationType{
			id:         {ID: id, Name: "Site-A"},
			uuid.New(): nil,
		}

		ordered := topologicalSortLocations(locs)

		if len(ordered) != 1 {
			t.Errorf("expected 1 location (nil skipped), got %d", len(ordered))
		}
	})
}

func TestMakeTenantRef(t *testing.T) {
	t.Run("creates ref from valid UUID", func(t *testing.T) {
		id := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		ref := makeTenantRef(id)

		if ref == nil {
			t.Fatal("expected non-nil ref")
		}
		if ref.Id == nil {
			t.Fatal("expected ref.Id to be non-nil")
		}
		got, err := ref.Id.AsBulkWritableCableRequestStatusId0()
		if err != nil {
			t.Fatalf("unexpected error extracting UUID: %v", err)
		}
		if uuid.UUID(got) != id {
			t.Errorf("makeTenantRef() round-trip = %s, want %s", uuid.UUID(got), id)
		}
	})

	t.Run("creates ref from nil UUID", func(t *testing.T) {
		ref := makeTenantRef(uuid.Nil)
		if ref == nil {
			t.Fatal("expected non-nil ref even for nil UUID")
		}
	})
}
