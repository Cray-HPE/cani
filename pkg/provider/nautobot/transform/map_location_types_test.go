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
package transform

import (
	"testing"

	openapi_types "github.com/Cray-HPE/cani/internal/openapi/types"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// TestLocationTypeKeyResolvesRegistryNamesAndSlugs verifies LocationTypeKey
// maps a Nautobot location type name back to the registered cani key by
// display name or slug, case-insensitively, and slugifies unknown names.
//
// Why it matters: export writes the definition's display name ("Data Center")
// for the key ("dc"); import must reverse it or a re-export cannot resolve the
// type and the datastore holds an API URL instead of a key.
// Inputs: "Data Center", "DC", "section", "Campus Building". Outputs: "dc",
// "dc", "section", "campus-building".
// Data choice: the bundled dc/section definitions cover both match branches;
// the unknown two-word name covers the slug fallback.
func TestLocationTypeKeyResolvesRegistryNamesAndSlugs(t *testing.T) {
	cases := map[string]string{
		"Data Center":     "dc",
		"DC":              "dc",
		"section":         "section",
		"Campus Building": "campus-building",
	}
	for name, want := range cases {
		if got := LocationTypeKey(name); got != want {
			t.Errorf("LocationTypeKey(%q) = %q, want %q", name, got, want)
		}
	}
}

// TestBuildLocationTypeKeyMapKeysByNautobotID verifies BuildLocationTypeKeyMap
// keys resolved cani keys by the Nautobot UUID and ignores entries without one.
//
// Why it matters: MapLocations resolves each location's LocationType reference
// through this map; a wrong key or a nil-ID crash would break every import.
// Inputs: two location types with IDs ("Data Center", "Level") and one with a
// nil ID. Outputs: a two-entry map with "dc" and "level".
// Data choice: the nil-ID entry models a malformed API object.
func TestBuildLocationTypeKeyMapKeysByNautobotID(t *testing.T) {
	dcID := openapi_types.UUID(uuid.New())
	levelID := openapi_types.UUID(uuid.New())
	raw := []nautobotapi.LocationType{
		{Id: &dcID, Name: "Data Center"},
		{Id: &levelID, Name: "Level"},
		{Name: "orphan"},
	}

	keys := BuildLocationTypeKeyMap(raw)

	if len(keys) != 2 || keys[uuid.UUID(dcID)] != "dc" || keys[uuid.UUID(levelID)] != "level" {
		t.Fatalf("keys = %v, want dc and level keyed by their IDs", keys)
	}
}
