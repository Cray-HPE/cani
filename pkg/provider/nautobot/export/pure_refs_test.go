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
	"testing"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/Cray-HPE/cani/pkg/provider/nautobot/transform"
	"github.com/google/uuid"
)

// ---------- refID / tenantRefID ----------

// TestRefID verifies refID extracts a UUID from a reference union, returns Nil
// for a nil pointer, and that tenantRefID delegates identically.
//
// Why it matters: Nautobot responses wrap related-object IDs in union types;
// diffing local vs remote devices depends on pulling the UUID back out, and a
// nil-unsafe extractor would panic on absent references.
// Inputs: nil, then a union built from a fixed UUID.
// Outputs: uuid.Nil for nil, the embedded UUID otherwise.
// Data choice: a fixed all-a UUID makes the round-trip assertion exact and the
// tenantRefID delegation check reuses the same union.
func TestRefID(t *testing.T) {
	// nil returns Nil
	if got := refID(nil); got != uuid.Nil {
		t.Errorf("refID(nil) = %s, want Nil", got)
	}

	// valid union
	id := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	var req nautobotapi.WritablePrefixRequest
	mustSetRefID(t, &req.Status, id)
	if got := refID(req.Status.Id); got != id {
		t.Errorf("refID() = %s, want %s", got, id)
	}

	// tenantRefID delegates to refID
	if got := tenantRefID(req.Status.Id); got != id {
		t.Errorf("tenantRefID() = %s, want %s", got, id)
	}
	if got := tenantRefID(nil); got != uuid.Nil {
		t.Errorf("tenantRefID(nil) = %s, want Nil", got)
	}
}

// TestRefIDNilInput verifies refID returns uuid.Nil for a nil reference union.
//
// Why it matters: remote Nautobot references are frequently nil; refID must
// degrade to uuid.Nil so diffing/mapping treats them as "absent" rather than
// panicking.
// Inputs: a nil *BulkWritableCableRequestStatusId. Outputs: uuid.Nil.
// Data choice: nil is the boundary input that the guard clause protects
// against.
func TestRefIDNilInput(t *testing.T) {
	result := refID(nil)
	if result != uuid.Nil {
		t.Errorf("expected uuid.Nil for nil input, got %s", result)
	}
}

// TestTenantRefIDNilInput verifies tenantRefID returns uuid.Nil for a nil
// reference union (it delegates to refID).
//
// Why it matters: tenant/rack references on remote devices may be nil; the
// helper must yield uuid.Nil so comparisons treat them as unset.
// Inputs: a nil *BulkWritableCableRequestStatusId. Outputs: uuid.Nil.
// Data choice: nil is the boundary input exercising the delegated guard.
func TestTenantRefIDNilInput(t *testing.T) {
	result := tenantRefID(nil)
	if result != uuid.Nil {
		t.Errorf("expected uuid.Nil for nil input, got %s", result)
	}
}

// ---------- setRefID ----------

// TestSetRefIDStatus verifies setRefID wraps a UUID into a status reference
// whose embedded Id round-trips back to the same UUID.
//
// Why it matters: device writes reference status/role/etc. by this union type;
// a malformed wrapper would send a bad or empty reference to Nautobot.
// Inputs: a fixed UUID. Outputs: a ref with non-nil Id decoding to that UUID.
// Data choice: an all-1s UUID makes the AsBulkWritableCableRequestStatusId0
// round-trip assertion unambiguous.
func TestSetRefIDStatus(t *testing.T) {
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	var req nautobotapi.WritablePrefixRequest
	mustSetRefID(t, &req.Status, id)
	if req.Status.Id == nil {
		t.Fatal("expected non-nil Id")
	}
	if got := transform.RefUUID(req.Status.Id); got != id {
		t.Errorf("expected %s, got %s", id, got)
	}
}

// TestSetRefIDLocation verifies setRefID wraps a UUID into a prefix-
// location reference whose Id round-trips back to the same UUID.
//
// Why it matters: prefix writes attach a location by this typed reference, so
// the UUID must survive wrapping to place the prefix correctly.
// Inputs: a fixed UUID. Outputs: a ref with non-nil Id decoding to that UUID.
// Data choice: an all-2s UUID distinguishes this case from the other setRefID
// tests while reusing the same round-trip check.
func TestSetRefIDLocation(t *testing.T) {
	id := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	var req nautobotapi.WritablePrefixRequest
	mustSetRefID(t, &req.Location, id)
	if req.Location == nil || req.Location.Id == nil {
		t.Fatal("expected non-nil Id")
	}
	if got := transform.RefUUID(req.Location.Id); got != id {
		t.Errorf("expected %s, got %s", id, got)
	}
}

// TestSetRefIDPrefixParent verifies setRefID wraps a UUID into a
// prefix-parent reference whose Id round-trips back to the same UUID.
//
// Why it matters: nested prefixes set their parent via this reference; a lost
// UUID would orphan child prefixes during export.
// Inputs: a fixed UUID. Outputs: a ref with non-nil Id decoding to that UUID.
// Data choice: an all-3s UUID keeps each setRefID test's fixture distinct for
// clearer failure diagnosis.
func TestSetRefIDPrefixParent(t *testing.T) {
	id := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	var req nautobotapi.WritablePrefixRequest
	mustSetRefID(t, &req.Parent, id)
	if req.Parent == nil || req.Parent.Id == nil {
		t.Fatal("expected non-nil Id")
	}
	if got := transform.RefUUID(req.Parent.Id); got != id {
		t.Errorf("expected %s, got %s", id, got)
	}
}

// TestSetRefIDIPParent verifies setRefID wraps a UUID into an IP-address
// parent (prefix) reference whose Id round-trips back to the same UUID.
//
// Why it matters: IP addresses link to their containing prefix via this typed
// reference, which export needs to set parentage correctly.
// Inputs: a fixed UUID. Outputs: a ref with non-nil Id decoding to that UUID.
// Data choice: an all-4s UUID mirrors the sibling setRefID tests with a unique
// value.
func TestSetRefIDIPParent(t *testing.T) {
	id := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	var req nautobotapi.IPAddressRequest
	mustSetRefID(t, &req.Parent, id)
	if req.Parent == nil || req.Parent.Id == nil {
		t.Fatal("expected non-nil Id")
	}
	if got := transform.RefUUID(req.Parent.Id); got != id {
		t.Errorf("expected %s, got %s", id, got)
	}
}

// TestSetRefIDIPNamespace verifies setRefID wraps a UUID into an IP
// namespace reference whose Id round-trips back to the same UUID.
//
// Why it matters: IP addresses and prefixes are scoped to a namespace in
// Nautobot; a broken namespace reference would file addresses in the wrong
// (or default) namespace.
// Inputs: a fixed UUID. Outputs: a ref with non-nil Id decoding to that UUID.
// Data choice: an all-5s UUID completes the distinct-per-test fixture set.
func TestSetRefIDIPNamespace(t *testing.T) {
	id := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	var req nautobotapi.IPAddressRequest
	mustSetRefID(t, &req.Namespace, id)
	if req.Namespace == nil || req.Namespace.Id == nil {
		t.Fatal("expected non-nil Id")
	}
	if got := transform.RefUUID(req.Namespace.Id); got != id {
		t.Errorf("expected %s, got %s", id, got)
	}
}

// ---------- refID with valid union ----------

// TestRefIDValidUnion verifies refID extracts the embedded UUID from a populated
// BulkWritableCableRequestStatusId union.
//
// Why it matters: Nautobot foreign keys arrive as oapi-codegen union types, and
// diffing/exporting relies on pulling the raw UUID back out correctly.
// Inputs: a union built via FromBulkWritableCableRequestStatusId0(expected).
// Outputs: the same UUID that was stored.
// Data choice: a freshly generated UUID round-trips the union to prove identity
// is preserved.
func TestRefIDValidUnion(t *testing.T) {
	expected := uuid.New()
	var req nautobotapi.WritablePrefixRequest
	mustSetRefID(t, &req.Status, expected)

	got := refID(req.Status.Id)
	if got != expected {
		t.Errorf("got %s, want %s", got, expected)
	}
}
