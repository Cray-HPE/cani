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

func strPtr(s string) *string { return &s }
func intPtr(v int) *int       { return &v }

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

func TestRefIDVal(t *testing.T) {
	id := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	ref := makeStatusRefFromUUID(id)

	got := refIDVal(ref)
	if got != id {
		t.Errorf("refIDVal() = %s, want %s", got, id)
	}
}

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
