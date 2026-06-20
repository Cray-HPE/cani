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

// TestMapLocations verifies MapLocations converts locations, resolves status
// names and parent links to CANI UUIDs, builds the Nautobot->CANI UUID map, and
// skips nil-Id locations.
//
// Why it matters: locations are the first entity imported and seed the UUID map
// every later mapper (racks, devices) depends on; parent links must be rewritten
// from Nautobot to CANI UUIDs, and a parent outside the import set must resolve
// to uuid.Nil rather than dangle.
// Inputs: nil, a nil-Id location, a single location, a parent/child pair, a child
// with an unknown parent, and a location with custom fields. Outputs: the CANI
// location map and the Nautobot->CANI UUID map, including status name fields.
// Data choice: a two-element parent/child pair proves cross-referencing rewrites
// the child's parent to the parent's CANI UUID, and the unknown-parent case
// proves the second-pass resolution falls back to uuid.Nil.
func TestMapLocations(t *testing.T) {
	t.Run("empty input returns empty maps", func(t *testing.T) {
		locs, nbMap := MapLocations(nil, nil)
		if len(locs) != 0 {
			t.Errorf("expected 0 locations, got %d", len(locs))
		}
		if len(nbMap) != 0 {
			t.Errorf("expected 0 mappings, got %d", len(nbMap))
		}
	})

	t.Run("location with nil ID is skipped", func(t *testing.T) {
		raw := []nautobotapi.Location{
			{Name: "orphan", Id: nil, Status: makeStatusRefFromUUID(uuid.New()), LocationType: makeStatusRefFromUUID(uuid.New())},
		}
		locs, nbMap := MapLocations(raw, nil)
		if len(locs) != 0 {
			t.Errorf("expected 0 locations, got %d", len(locs))
		}
		if len(nbMap) != 0 {
			t.Errorf("expected 0 mappings, got %d", len(nbMap))
		}
	})

	t.Run("single location is mapped", func(t *testing.T) {
		nbID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
		oaID := openapi_types.UUID(nbID)
		desc := "Test datacenter"
		facility := "DC-1"
		statusID := uuid.MustParse("12121212-1212-1212-1212-121212121212")
		statusNameMap := map[uuid.UUID]string{statusID: "Active"}

		raw := []nautobotapi.Location{
			{
				Id:           &oaID,
				Name:         "Site-A",
				Description:  &desc,
				Facility:     &facility,
				Status:       makeStatusRefFromUUID(statusID),
				LocationType: makeStatusRefFromUUID(uuid.New()),
			},
		}

		locs, nbMap := MapLocations(raw, statusNameMap)
		if len(locs) != 1 {
			t.Fatalf("expected 1 location, got %d", len(locs))
		}
		if len(nbMap) != 1 {
			t.Fatalf("expected 1 mapping, got %d", len(nbMap))
		}

		caniID := nbMap[nbID]
		loc := locs[caniID]
		if loc == nil {
			t.Fatal("location not found by CANI ID")
		}
		if loc.Name != "Site-A" {
			t.Errorf("Name = %q, want %q", loc.Name, "Site-A")
		}
		if loc.Description != "Test datacenter" {
			t.Errorf("Description = %q, want %q", loc.Description, "Test datacenter")
		}
		if loc.Facility != "DC-1" {
			t.Errorf("Facility = %q, want %q", loc.Facility, "DC-1")
		}
		if loc.Status != "Active" {
			t.Errorf("Status = %q, want %q", loc.Status, "Active")
		}
		if loc.ObjectMeta.ExternalIDs["nautobot"] != nbID {
			t.Errorf("ExternalIDs[nautobot] = %s, want %s", loc.ObjectMeta.ExternalIDs["nautobot"], nbID)
		}
	})

	t.Run("parent resolution maps nautobot UUID to CANI UUID", func(t *testing.T) {
		parentNBID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
		childNBID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
		parentOA := openapi_types.UUID(parentNBID)
		childOA := openapi_types.UUID(childNBID)
		parentRef := makeTenantRefFromUUID(parentNBID)

		raw := []nautobotapi.Location{
			{
				Id:           &parentOA,
				Name:         "Region",
				Status:       makeStatusRefFromUUID(uuid.New()),
				LocationType: makeStatusRefFromUUID(uuid.New()),
			},
			{
				Id:           &childOA,
				Name:         "Site",
				Parent:       &parentRef,
				Status:       makeStatusRefFromUUID(uuid.New()),
				LocationType: makeStatusRefFromUUID(uuid.New()),
			},
		}

		locs, nbMap := MapLocations(raw, nil)
		if len(locs) != 2 {
			t.Fatalf("expected 2 locations, got %d", len(locs))
		}

		childCaniID := nbMap[childNBID]
		parentCaniID := nbMap[parentNBID]
		child := locs[childCaniID]

		if child.Parent != parentCaniID {
			t.Errorf("child.Parent = %s, want %s", child.Parent, parentCaniID)
		}
	})

	t.Run("parent not in import set is resolved to nil UUID", func(t *testing.T) {
		childNBID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
		unknownParent := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
		childOA := openapi_types.UUID(childNBID)
		parentRef := makeTenantRefFromUUID(unknownParent)

		raw := []nautobotapi.Location{
			{
				Id:           &childOA,
				Name:         "Orphan-Site",
				Parent:       &parentRef,
				Status:       makeStatusRefFromUUID(uuid.New()),
				LocationType: makeStatusRefFromUUID(uuid.New()),
			},
		}

		locs, nbMap := MapLocations(raw, nil)
		childCaniID := nbMap[childNBID]
		child := locs[childCaniID]

		if child.Parent != uuid.Nil {
			t.Errorf("child.Parent = %s, want Nil (unknown parent)", child.Parent)
		}
	})

	t.Run("custom fields are passed through", func(t *testing.T) {
		nbID := uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
		oaID := openapi_types.UUID(nbID)
		cf := map[string]interface{}{"env": "prod"}

		raw := []nautobotapi.Location{
			{
				Id:           &oaID,
				Name:         "CF-Site",
				CustomFields: &cf,
				Status:       makeStatusRefFromUUID(uuid.New()),
				LocationType: makeStatusRefFromUUID(uuid.New()),
			},
		}

		locs, nbMap := MapLocations(raw, nil)
		caniID := nbMap[nbID]
		loc := locs[caniID]

		if loc.CustomFields == nil {
			t.Fatal("expected CustomFields to be set")
		}
		if loc.CustomFields["env"] != "prod" {
			t.Errorf("CustomFields[env] = %v, want %q", loc.CustomFields["env"], "prod")
		}
	})
}
