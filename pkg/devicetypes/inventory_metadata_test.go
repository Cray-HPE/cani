/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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
package devicetypes

import "testing"

// TestAddMetadataEachKind verifies AddMetadata appends an entry to the correct
// catalog slice for each supported kind and lazily allocates the metadata
// container.
//
// Why it matters: the metadata catalog backs role/status/tag definitions that
// inventory items reference, so each kind must land in its own slice and a nil
// Metadata pointer must be initialized rather than panicking.
// Inputs: a fresh inventory and one entry added for "roles", "statuses", and
// "tags". Outputs: nil errors and each ListMetadata(kind) returning the entry.
// Data choice: a distinct entry name per kind proves the switch routes to the
// right slice instead of writing them all to one.
func TestAddMetadataEachKind(t *testing.T) {
	inv := NewInventory()
	inv.Metadata = nil // force lazy allocation path

	kinds := []string{"roles", "statuses", "tags"}
	for _, kind := range kinds {
		entry := MetadataEntry{Name: kind + "-entry", Color: "ff0000"}
		if err := inv.AddMetadata(kind, entry); err != nil {
			t.Fatalf("AddMetadata(%q) unexpected error: %v", kind, err)
		}
		got := inv.ListMetadata(kind)
		if len(got) != 1 || got[0].Name != kind+"-entry" {
			t.Errorf("ListMetadata(%q) = %+v, want one entry named %q", kind, got, kind+"-entry")
		}
	}
}

// TestAddMetadataDuplicate verifies AddMetadata rejects a second entry with a
// name that already exists for the same kind.
//
// Why it matters: catalog names are identifiers, so duplicates must be refused
// to keep references unambiguous.
// Inputs: two "roles" entries both named "leader". Outputs: nil error on the
// first add and a non-nil error on the second.
// Data choice: identical names with the same kind isolate the uniqueness check
// from the kind-routing logic.
func TestAddMetadataDuplicate(t *testing.T) {
	inv := NewInventory()
	if err := inv.AddMetadata("roles", MetadataEntry{Name: "leader"}); err != nil {
		t.Fatalf("first AddMetadata() error: %v", err)
	}
	if err := inv.AddMetadata("roles", MetadataEntry{Name: "leader"}); err == nil {
		t.Error("AddMetadata(duplicate name) should return an error")
	}
}

// TestAddMetadataUnknownKind verifies AddMetadata rejects an unrecognized kind.
//
// Why it matters: only roles, statuses, and tags are valid catalogs, so an
// unknown kind must error rather than silently dropping the entry.
// Inputs: an entry added under kind "widgets". Outputs: a non-nil error.
// Data choice: "widgets" is clearly outside the supported set, exercising the
// default branch of the switch.
func TestAddMetadataUnknownKind(t *testing.T) {
	inv := NewInventory()
	if err := inv.AddMetadata("widgets", MetadataEntry{Name: "x"}); err == nil {
		t.Error("AddMetadata(unknown kind) should return an error")
	}
}

// TestListMetadataNilAndUnknown verifies ListMetadata returns nil when the
// catalog is unallocated and when the kind is unknown.
//
// Why it matters: read paths must be nil-safe so callers can range over the
// result without first checking allocation or kind validity.
// Inputs: a kind query against a nil Metadata pointer, then an unknown kind
// against an allocated catalog. Outputs: nil slices.
// Data choice: the two queries cover both early-return branches of
// listMetadataSlice (nil container and default case).
func TestListMetadataNilAndUnknown(t *testing.T) {
	inv := NewInventory()
	inv.Metadata = nil
	if got := inv.ListMetadata("roles"); got != nil {
		t.Errorf("ListMetadata on nil catalog = %+v, want nil", got)
	}

	inv.Metadata = &InventoryMetadata{}
	if got := inv.ListMetadata("widgets"); got != nil {
		t.Errorf("ListMetadata(unknown kind) = %+v, want nil", got)
	}
}
