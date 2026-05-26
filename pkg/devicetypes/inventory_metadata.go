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

import "fmt"

// MetadataEntry represents a role, status, or tag definition stored
// in the inventory metadata catalog.
type MetadataEntry struct {
	Name         string   `json:"name" yaml:"name"`
	Color        string   `json:"color,omitempty" yaml:"color,omitempty"`
	Description  string   `json:"description,omitempty" yaml:"description,omitempty"`
	ContentTypes []string `json:"contentTypes,omitempty" yaml:"content_types,omitempty"`
	Weight       int      `json:"weight,omitempty" yaml:"weight,omitempty"`
}

// InventoryMetadata holds the catalog of metadata definitions (roles,
// statuses, tags) that can be referenced by individual inventory items.
// It lives at the top level of the Inventory struct.
type InventoryMetadata struct {
	Roles    []MetadataEntry `json:"roles,omitempty"    yaml:"roles,omitempty"`
	Statuses []MetadataEntry `json:"statuses,omitempty" yaml:"statuses,omitempty"`
	Tags     []MetadataEntry `json:"tags,omitempty"     yaml:"tags,omitempty"`
}

// AddMetadata stores a metadata definition (role, status, or tag) in the
// inventory metadata catalog. kind must be "roles", "statuses", or "tags".
// Returns an error if an entry with the same name already exists.
func (inv *Inventory) AddMetadata(kind string, entry MetadataEntry) error {
	if inv.Metadata == nil {
		inv.Metadata = &InventoryMetadata{}
	}

	entries := inv.listMetadataSlice(kind)

	for _, e := range entries {
		if e.Name == entry.Name {
			return fmt.Errorf("%s %q already exists", kind, entry.Name)
		}
	}

	switch kind {
	case "roles":
		inv.Metadata.Roles = append(inv.Metadata.Roles, entry)
	case "statuses":
		inv.Metadata.Statuses = append(inv.Metadata.Statuses, entry)
	case "tags":
		inv.Metadata.Tags = append(inv.Metadata.Tags, entry)
	default:
		return fmt.Errorf("unknown metadata kind %q", kind)
	}
	return nil
}

// ListMetadata returns all metadata entries for a given kind.
func (inv *Inventory) ListMetadata(kind string) []MetadataEntry {
	return inv.listMetadataSlice(kind)
}

// listMetadataSlice returns the slice pointer for the given kind.
func (inv *Inventory) listMetadataSlice(kind string) []MetadataEntry {
	if inv.Metadata == nil {
		return nil
	}
	switch kind {
	case "roles":
		return inv.Metadata.Roles
	case "statuses":
		return inv.Metadata.Statuses
	case "tags":
		return inv.Metadata.Tags
	default:
		return nil
	}
}
