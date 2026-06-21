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

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// makeStatusRefFromUUID creates a BulkWritableCableRequestStatus with the given UUID.
func makeStatusRefFromUUID(id uuid.UUID) nautobotapi.BulkWritableCableRequestStatus {
	var idUnion nautobotapi.BulkWritableCableRequestStatusId
	_ = idUnion.FromBulkWritableCableRequestStatusId0(openapi_types.UUID(id))
	return nautobotapi.BulkWritableCableRequestStatus{Id: &idUnion}
}

// makeTenantRefFromUUID creates a BulkWritableCircuitRequestTenant with the given UUID.
func makeTenantRefFromUUID(id uuid.UUID) nautobotapi.BulkWritableCircuitRequestTenant {
	var idUnion nautobotapi.BulkWritableCableRequestStatusId
	_ = idUnion.FromBulkWritableCableRequestStatusId0(openapi_types.UUID(id))
	return nautobotapi.BulkWritableCircuitRequestTenant{Id: &idUnion}
}

// makeInvalidIDUnion builds an Id union holding the integer variant. The UUID
// accessor (AsBulkWritableCableRequestStatusId0) then fails to decode it,
// exercising the error path in refID/tenantRefID.
func makeInvalidIDUnion() nautobotapi.BulkWritableCableRequestStatusId {
	var idUnion nautobotapi.BulkWritableCableRequestStatusId
	_ = idUnion.FromBulkWritableCableRequestStatusId1(123)
	return idUnion
}

func strPtr(s string) *string { return &s }
func intPtr(v int) *int       { return &v }

// TestRefID verifies refID extracts a UUID from a cable/status reference,
// returning the nil UUID for a nil reference or a nil inner Id.
//
// Why it matters: Nautobot references arrive as oapi-codegen unions; the import
// transform dereferences them to CANI UUIDs, so a missing reference must degrade
// to uuid.Nil rather than panic or fabricate a foreign key.
// Inputs: nil, a ref with a nil Id, and a ref wrapping a known UUID. Outputs: the
// extracted uuid.UUID for each case.
// Data choice: the two nil shapes cover both guard clauses, and a fixed UUID
// proves the happy path returns the exact value that was wrapped.
func TestRefID(t *testing.T) {
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	tests := []struct {
		name     string
		ref      *nautobotapi.BulkWritableCableRequestStatus
		expected uuid.UUID
	}{
		{
			name:     "nil ref returns nil UUID",
			ref:      nil,
			expected: uuid.Nil,
		},
		{
			name:     "ref with nil Id returns nil UUID",
			ref:      &nautobotapi.BulkWritableCableRequestStatus{Id: nil},
			expected: uuid.Nil,
		},
		{
			name: "ref with valid UUID returns it",
			ref: func() *nautobotapi.BulkWritableCableRequestStatus {
				r := makeStatusRefFromUUID(id)
				return &r
			}(),
			expected: id,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := refID(tt.ref)
			if got != tt.expected {
				t.Errorf("refID() = %s, want %s", got, tt.expected)
			}
		})
	}
}

// TestRefIDVal verifies refIDVal extracts the UUID from a non-pointer
// cable/status reference.
//
// Why it matters: several mappers hold references by value (e.g. device status,
// rack location); refIDVal lets them resolve a CANI foreign key without first
// taking an address, so it must mirror refID's UUID extraction.
// Inputs: a value reference wrapping a known UUID. Outputs: the extracted
// uuid.UUID.
// Data choice: a single fixed UUID is the minimal case proving the value form
// unwraps to exactly the UUID it was built from.
func TestRefIDVal(t *testing.T) {
	id := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	ref := makeStatusRefFromUUID(id)

	got := refIDVal(ref)
	if got != id {
		t.Errorf("refIDVal() = %s, want %s", got, id)
	}
}

// TestTenantRefID verifies tenantRefID extracts a UUID from a tenant-style
// reference, returning the nil UUID for a nil reference or nil inner Id.
//
// Why it matters: tenant references model parent/owner links (rack parents,
// device racks, module locations, manufacturers); the transform must resolve a
// present link and skip an absent one instead of emitting a bogus foreign key.
// Inputs: nil, a ref with a nil Id, and a ref wrapping a known UUID. Outputs: the
// extracted uuid.UUID for each case.
// Data choice: both nil shapes exercise the guard clauses, and a fixed UUID
// confirms the populated path returns the wrapped value unchanged.
func TestTenantRefID(t *testing.T) {
	id := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	tests := []struct {
		name     string
		ref      *nautobotapi.BulkWritableCircuitRequestTenant
		expected uuid.UUID
	}{
		{
			name:     "nil ref returns nil UUID",
			ref:      nil,
			expected: uuid.Nil,
		},
		{
			name:     "ref with nil Id returns nil UUID",
			ref:      &nautobotapi.BulkWritableCircuitRequestTenant{Id: nil},
			expected: uuid.Nil,
		},
		{
			name: "ref with valid UUID returns it",
			ref: func() *nautobotapi.BulkWritableCircuitRequestTenant {
				r := makeTenantRefFromUUID(id)
				return &r
			}(),
			expected: id,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tenantRefID(tt.ref)
			if got != tt.expected {
				t.Errorf("tenantRefID() = %s, want %s", got, tt.expected)
			}
		})
	}
}

// TestDirectUUID verifies directUUID converts an openapi_types.UUID pointer to a
// uuid.UUID, returning the nil UUID for a nil pointer.
//
// Why it matters: every Nautobot object's primary Id is an optional
// openapi_types.UUID pointer; the mappers use directUUID to derive the source
// key that anchors the Nautobot->CANI UUID map, so a nil Id must collapse to
// uuid.Nil and be skipped.
// Inputs: a nil pointer and a pointer to a known UUID. Outputs: the converted
// uuid.UUID.
// Data choice: the nil case covers the skip guard the mappers rely on; a fixed
// UUID proves the conversion preserves the value.
func TestDirectUUID(t *testing.T) {
	id := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	tests := []struct {
		name     string
		input    *openapi_types.UUID
		expected uuid.UUID
	}{
		{
			name:     "nil returns nil UUID",
			input:    nil,
			expected: uuid.Nil,
		},
		{
			name: "valid pointer returns UUID",
			input: func() *openapi_types.UUID {
				u := openapi_types.UUID(id)
				return &u
			}(),
			expected: id,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := directUUID(tt.input)
			if got != tt.expected {
				t.Errorf("directUUID() = %s, want %s", got, tt.expected)
			}
		})
	}
}

// TestStrVal verifies strVal safely dereferences a *string, returning the empty
// string when the pointer is nil.
//
// Why it matters: most Nautobot string fields are optional pointers copied into
// CANI structs during import, so a nil field must become "" rather than
// dereference-panic.
// Inputs: nil, a pointer to "hello", and a pointer to "". Outputs: the
// dereferenced string.
// Data choice: a non-empty value and an empty-string pointer together prove the
// function distinguishes "absent" (nil) from "present but empty", both yielding
// "" via different branches.
func TestStrVal(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{name: "nil returns empty", input: nil, expected: ""},
		{name: "non-nil returns value", input: strPtr("hello"), expected: "hello"},
		{name: "empty string ptr returns empty", input: strPtr(""), expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strVal(tt.input)
			if got != tt.expected {
				t.Errorf("strVal() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestIntVal verifies intVal safely dereferences a *int, returning 0 when the
// pointer is nil.
//
// Why it matters: optional numeric Nautobot fields (rack U-height, outer
// dimensions) are int pointers copied into CANI structs, so a nil field must
// become 0 rather than panic during import.
// Inputs: nil, a pointer to 42, and a pointer to 0. Outputs: the dereferenced
// int.
// Data choice: 42 proves a non-zero value is preserved, while the zero pointer
// exercises the non-nil branch for a value indistinguishable in result from an
// absent field.
func TestIntVal(t *testing.T) {
	tests := []struct {
		name     string
		input    *int
		expected int
	}{
		{name: "nil returns 0", input: nil, expected: 0},
		{name: "non-nil returns value", input: intPtr(42), expected: 42},
		{name: "zero pointer returns 0", input: intPtr(0), expected: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := intVal(tt.input)
			if got != tt.expected {
				t.Errorf("intVal() = %d, want %d", got, tt.expected)
			}
		})
	}
}

// TestRefDisplay verifies refDisplay returns a reference's Url field, yielding
// the empty string for a nil reference or a nil Url.
//
// Why it matters: when a human-readable display name is unavailable, the import
// falls back to the reference URL; refDisplay centralizes that fallback so
// callers never dereference a nil reference.
// Inputs: nil, a ref with no Url, and a ref carrying a URL string. Outputs: the
// display string.
// Data choice: the two empty cases cover both guard clauses, and a realistic API
// URL proves the populated path returns the URL verbatim.
func TestRefDisplay(t *testing.T) {
	tests := []struct {
		name     string
		ref      *nautobotapi.BulkWritableCableRequestStatus
		expected string
	}{
		{
			name:     "nil ref returns empty",
			ref:      nil,
			expected: "",
		},
		{
			name:     "ref with nil Url returns empty",
			ref:      &nautobotapi.BulkWritableCableRequestStatus{},
			expected: "",
		},
		{
			name:     "ref with Url returns it",
			ref:      &nautobotapi.BulkWritableCableRequestStatus{Url: strPtr("http://example.com/api/status/1/")},
			expected: "http://example.com/api/status/1/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := refDisplay(tt.ref)
			if got != tt.expected {
				t.Errorf("refDisplay() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestBuildStatusNameMap verifies BuildStatusNameMap indexes statuses by UUID to
// name and skips entries with a nil Id.
//
// Why it matters: devices reference status by UUID, but CANI stores the
// human-readable status name; this lookup lets MapDevices resolve "Active"
// instead of a URL, so a nil-Id status must be dropped rather than corrupt the
// map.
// Inputs: nil, two valid statuses, and a mix of one nil-Id and one valid status.
// Outputs: the UUID->name map, asserted by length and per-key value.
// Data choice: two distinct named statuses prove independent keys map correctly,
// and the nil-Id "Orphan" proves the skip guard excludes it from the result.
func TestBuildStatusNameMap(t *testing.T) {
	id1 := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	id2 := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	oaID1 := openapi_types.UUID(id1)
	oaID2 := openapi_types.UUID(id2)

	tests := []struct {
		name     string
		statuses []nautobotapi.Status
		expected map[uuid.UUID]string
	}{
		{
			name:     "empty input returns empty map",
			statuses: nil,
			expected: map[uuid.UUID]string{},
		},
		{
			name: "statuses with IDs are mapped",
			statuses: []nautobotapi.Status{
				{Id: &oaID1, Name: "Active"},
				{Id: &oaID2, Name: "Planned"},
			},
			expected: map[uuid.UUID]string{
				id1: "Active",
				id2: "Planned",
			},
		},
		{
			name: "status with nil ID is skipped",
			statuses: []nautobotapi.Status{
				{Id: nil, Name: "Orphan"},
				{Id: &oaID1, Name: "Active"},
			},
			expected: map[uuid.UUID]string{
				id1: "Active",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildStatusNameMap(tt.statuses)
			if len(got) != len(tt.expected) {
				t.Fatalf("BuildStatusNameMap() len = %d, want %d", len(got), len(tt.expected))
			}
			for k, v := range tt.expected {
				if got[k] != v {
					t.Errorf("BuildStatusNameMap()[%s] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

// TestBuildRoleNameMap verifies BuildRoleNameMap indexes roles by UUID to name
// and skips entries with a nil Id.
//
// Why it matters: devices reference role by UUID; CANI stores the role name, so
// this lookup lets MapDevices resolve "Compute" from a reference, and a nil-Id
// role must be excluded rather than added under uuid.Nil.
// Inputs: nil, a single valid role, and a single nil-Id role. Outputs: the
// UUID->name map, asserted by length and per-key value.
// Data choice: one valid role proves the happy path, and a separate nil-Id role
// isolates the skip guard producing an empty map.
func TestBuildRoleNameMap(t *testing.T) {
	id1 := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	oaID1 := openapi_types.UUID(id1)

	tests := []struct {
		name     string
		roles    []nautobotapi.Role
		expected map[uuid.UUID]string
	}{
		{
			name:     "empty input returns empty map",
			roles:    nil,
			expected: map[uuid.UUID]string{},
		},
		{
			name: "role with ID is mapped",
			roles: []nautobotapi.Role{
				{Id: &oaID1, Name: "Compute"},
			},
			expected: map[uuid.UUID]string{
				id1: "Compute",
			},
		},
		{
			name: "role with nil ID is skipped",
			roles: []nautobotapi.Role{
				{Id: nil, Name: "Orphan"},
			},
			expected: map[uuid.UUID]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildRoleNameMap(tt.roles)
			if len(got) != len(tt.expected) {
				t.Fatalf("BuildRoleNameMap() len = %d, want %d", len(got), len(tt.expected))
			}
			for k, v := range tt.expected {
				if got[k] != v {
					t.Errorf("BuildRoleNameMap()[%s] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

// TestResolveRefName verifies resolveRefName returns the mapped name for a
// reference's UUID and falls back to the reference URL when the UUID is absent
// from the map.
//
// Why it matters: device status and role are imported as names; resolveRefName
// turns a UUID reference into a name via the prebuilt lookup, degrading to the
// URL when the referenced object was not imported, so neither path is lost.
// Inputs: a name map plus references that hit the map, miss it without a URL,
// miss it with a URL, and carry a nil Id with a URL. Outputs: the resolved
// string.
// Data choice: the four cases cover map-hit, map-miss-empty, map-miss-URL, and
// nil-Id-URL, exhausting both the lookup success and every URL-fallback branch.
func TestResolveRefName(t *testing.T) {
	id := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	nameMap := map[uuid.UUID]string{id: "Active"}

	tests := []struct {
		name     string
		ref      nautobotapi.BulkWritableCableRequestStatus
		nameMap  map[uuid.UUID]string
		expected string
	}{
		{
			name:     "found in map returns name",
			ref:      makeStatusRefFromUUID(id),
			nameMap:  nameMap,
			expected: "Active",
		},
		{
			name:     "not in map falls back to URL",
			ref:      makeStatusRefFromUUID(uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")),
			nameMap:  nameMap,
			expected: "",
		},
		{
			name: "not in map with URL returns URL",
			ref: func() nautobotapi.BulkWritableCableRequestStatus {
				r := makeStatusRefFromUUID(uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"))
				r.Url = strPtr("http://example.com/status/unknown/")
				return r
			}(),
			nameMap:  nameMap,
			expected: "http://example.com/status/unknown/",
		},
		{
			name:     "nil ID ref returns URL",
			ref:      nautobotapi.BulkWritableCableRequestStatus{Url: strPtr("http://fallback/")},
			nameMap:  nameMap,
			expected: "http://fallback/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveRefName(tt.ref, tt.nameMap)
			if got != tt.expected {
				t.Errorf("resolveRefName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestResolveTenantRefName verifies resolveTenantRefName returns the mapped name
// for a tenant-style reference's UUID and falls back to the reference URL or an
// empty string when no name is available.
//
// Why it matters: racks and modules carry role references as tenant-style
// objects, but CANI stores role names; resolving these references keeps role
// parity with device mapping while preserving a deterministic fallback.
// Inputs: a name map plus references that hit the map, miss it with a URL, miss
// it without a URL, and a nil reference. Outputs: the resolved string.
// Data choice: one fixed role UUID proves the map-hit path, while URL, empty,
// and nil cases cover every fallback branch the mappers rely on.
func TestResolveTenantRefName(t *testing.T) {
	id := uuid.MustParse("12121212-1212-1212-1212-121212121212")
	nameMap := map[uuid.UUID]string{id: "Network"}

	tests := []struct {
		name     string
		ref      *nautobotapi.BulkWritableCircuitRequestTenant
		nameMap  map[uuid.UUID]string
		expected string
	}{
		{
			name: "found in map returns name",
			ref: func() *nautobotapi.BulkWritableCircuitRequestTenant {
				r := makeTenantRefFromUUID(id)
				return &r
			}(),
			nameMap:  nameMap,
			expected: "Network",
		},
		{
			name: "not in map with URL returns URL",
			ref: func() *nautobotapi.BulkWritableCircuitRequestTenant {
				r := makeTenantRefFromUUID(uuid.MustParse("34343434-3434-3434-3434-343434343434"))
				r.Url = strPtr("http://example.com/roles/unknown/")
				return &r
			}(),
			nameMap:  nameMap,
			expected: "http://example.com/roles/unknown/",
		},
		{
			name: "not in map without URL returns empty",
			ref: func() *nautobotapi.BulkWritableCircuitRequestTenant {
				r := makeTenantRefFromUUID(uuid.MustParse("56565656-5656-5656-5656-565656565656"))
				return &r
			}(),
			nameMap:  nameMap,
			expected: "",
		},
		{
			name:     "nil ref returns empty",
			ref:      nil,
			nameMap:  nameMap,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveTenantRefName(tt.ref, tt.nameMap)
			if got != tt.expected {
				t.Errorf("resolveTenantRefName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestRefID_UnionDecodeError verifies refID returns the nil UUID when the Id
// union holds a non-UUID variant that fails to decode.
//
// Why it matters: Nautobot references arrive as oapi-codegen unions; the
// transform layer dereferences them to CANI UUIDs, so a malformed reference
// must degrade to uuid.Nil instead of panicking or returning garbage that would
// corrupt downstream device/location/cable wiring.
// Inputs: a reference whose Id union was populated with the integer variant.
// Outputs: the returned uuid.UUID (expected uuid.Nil).
// Data choice: the integer variant is the other arm of the generated union, so
// the UUID accessor's json.Unmarshal genuinely fails — proving the err != nil
// branch rather than the nil-pointer guard already covered elsewhere.
func TestRefID_UnionDecodeError(t *testing.T) {
	idUnion := makeInvalidIDUnion()
	ref := &nautobotapi.BulkWritableCableRequestStatus{Id: &idUnion}

	if got := refID(ref); got != uuid.Nil {
		t.Errorf("refID() = %s, want %s for undecodable union", got, uuid.Nil)
	}
}

// TestTenantRefID_UnionDecodeError verifies tenantRefID returns the nil UUID
// when the tenant reference's Id union cannot be decoded as a UUID.
//
// Why it matters: tenant-style references (rack/device parents, locations,
// roles, manufacturers) flow through tenantRefID; a reference that cannot be
// decoded must resolve to uuid.Nil so the mappers skip it rather than emit a
// bogus foreign key.
// Inputs: a tenant reference whose Id union holds the integer variant.
// Outputs: the returned uuid.UUID (expected uuid.Nil).
// Data choice: the integer variant forces the UUID accessor to error, isolating
// the decode-failure branch that the nil-ref and nil-Id cases cannot reach.
func TestTenantRefID_UnionDecodeError(t *testing.T) {
	idUnion := makeInvalidIDUnion()
	ref := &nautobotapi.BulkWritableCircuitRequestTenant{Id: &idUnion}

	if got := tenantRefID(ref); got != uuid.Nil {
		t.Errorf("tenantRefID() = %s, want %s for undecodable union", got, uuid.Nil)
	}
}
