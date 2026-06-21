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
package uuidutil

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// -----------------------------------------------------------------------------
// Join — deterministic, sorted, comma-free joining of UUIDs
// -----------------------------------------------------------------------------

// TestJoin_SortsIDsForDeterministicOutput covers the happy path: regardless of
// input order, Join returns the IDs sorted lexicographically and separated by
// the requested separator. Sorting is what makes the output stable for use as
// a map key or diff-friendly identifier.
func TestJoin_SortsIDsForDeterministicOutput(t *testing.T) {
	a := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	b := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	c := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	// Supply the IDs out of order to prove Join sorts them.
	got := Join([]uuid.UUID{c, a, b}, ",")
	want := strings.Join([]string{a.String(), b.String(), c.String()}, ",")
	if got != want {
		t.Errorf("Join unsorted input = %q, want %q", got, want)
	}
}

// TestJoin_OmitsIgnoredIDs covers the ignore-list behaviour: any ID passed as
// an ignore argument must not appear in the output.
func TestJoin_OmitsIgnoredIDs(t *testing.T) {
	keep := uuid.MustParse("00000000-0000-0000-0000-0000000000aa")
	drop := uuid.MustParse("00000000-0000-0000-0000-0000000000bb")

	got := Join([]uuid.UUID{keep, drop}, ",", drop)
	if got != keep.String() {
		t.Errorf("Join with ignore = %q, want %q", got, keep.String())
	}
	if strings.Contains(got, drop.String()) {
		t.Errorf("Join output %q must not contain ignored ID %s", got, drop)
	}
}

// TestJoin_EmptyInputReturnsEmptyString covers the boundary where there is
// nothing to join, which must yield an empty string rather than a stray
// separator.
func TestJoin_EmptyInputReturnsEmptyString(t *testing.T) {
	if got := Join(nil, ","); got != "" {
		t.Errorf("Join(nil) = %q, want empty string", got)
	}
}
