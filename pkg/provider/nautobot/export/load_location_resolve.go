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
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// resolveContentLocation walks the location tree from locID downward to find
// the deepest descendant whose LocationType allows the given content type
// ("rack", "device", or "module"). If the starting location itself supports
// the content type, it is returned immediately.
//
// This is needed because Nautobot location types enforce which content types
// can live at each level. For example a "Data Center" type does not allow
// racks or devices — those must be placed at a "Section" or similar child
// location type that includes "rack" in its content_types.
//
// If the direct tree walk finds no suitable descendant, it searches other
// locations with the same name in case there are multiple instances (e.g.,
// two "DC-01" entries where only one has a complete hierarchy).
//
// Returns the resolved location name or empty string if no suitable location
// is found, allowing the caller to fall back to its default.
func resolveContentLocation(locID uuid.UUID, contentType string, inv *devicetypes.Inventory) string {
	if inv == nil || locID == uuid.Nil {
		return ""
	}

	loc, ok := inv.Locations[locID]
	if !ok || loc == nil {
		return ""
	}

	// If this location's type supports the content type, use it directly.
	if locationTypeSupports(loc.LocationType, contentType) {
		return loc.Name
	}

	// Walk children depth-first looking for a descendant that supports it.
	if name := findContentChild(loc, contentType, inv); name != "" {
		clog.Detail("[location] Resolved %s (%s) to descendant %s for content type %q",
			loc.Name, loc.LocationType, name, contentType)
		return name
	}

	// Direct tree walk failed. Search other locations with the same name
	// in case there are multiple instances with different hierarchies.
	for otherID, other := range inv.Locations {
		if otherID == locID || other == nil || other.Name != loc.Name {
			continue
		}
		if name := findContentChild(other, contentType, inv); name != "" {
			clog.Detail("[location] Resolved %s (%s) via sibling %s to descendant %s for content type %q",
				loc.Name, loc.LocationType, otherID, name, contentType)
			return name
		}
	}

	// No descendant found — return empty so caller can fall back to default.
	clog.Detail("[location] Location %s (%s) does not support %q and has no suitable descendant",
		loc.Name, loc.LocationType, contentType)
	return ""
}

// findContentChild recursively searches children of loc for the deepest
// location whose type supports contentType.
func findContentChild(loc *devicetypes.CaniLocationType, contentType string, inv *devicetypes.Inventory) string {
	for _, childID := range loc.Children {
		child, ok := inv.Locations[childID]
		if !ok || child == nil {
			continue
		}

		// Recurse first: prefer deeper descendants (depth-first).
		if deeper := findContentChild(child, contentType, inv); deeper != "" {
			return deeper
		}

		// If this child supports it, use it.
		if locationTypeSupports(child.LocationType, contentType) {
			return child.Name
		}
	}
	return ""
}

// locationTypeSupports checks whether a location type slug has the given
// content type in its ContentTypes list. It consults the in-memory
// LocationTypeDefinition registry loaded from YAML.
func locationTypeSupports(locTypeSlug string, contentType string) bool {
	lt, ok := devicetypes.GetLocationTypeBySlug(locTypeSlug)
	if !ok {
		return false
	}
	for _, ct := range lt.ContentTypes {
		if ct == contentType {
			return true
		}
	}
	return false
}
