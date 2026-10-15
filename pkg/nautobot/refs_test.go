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
package nautobot

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// rawRef is a stand-in reference union that marshals to whatever bytes the
// test hands it, so each decode branch can be driven without a generated type.
type rawRef struct {
	b   []byte
	err error
}

func (r rawRef) MarshalJSON() ([]byte, error) { return r.b, r.err }

// TestRefUUIDCheckedRoundTripsGeneratedUnion verifies that a UUID written with
// SetRefUUID into a generated per-field union is read back by RefUUIDChecked.
//
// Why it matters: every provider reference (status, role, location, parent) is
// carried through these unions; the write and read halves must agree on the
// member that holds the UUID or exports would silently reference nothing.
// Inputs: a fixed UUID set on Device_Status_Id. Outputs: the same UUID, no error.
// Data choice: Device_Status_Id is the union the transform stage reads most.
func TestRefUUIDCheckedRoundTripsGeneratedUnion(t *testing.T) {
	want := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	var ref Device_Status_Id
	if err := SetRefUUID(&ref, want); err != nil {
		t.Fatalf("SetRefUUID: %v", err)
	}
	got, err := RefUUIDChecked(ref)
	if err != nil {
		t.Fatalf("RefUUIDChecked: %v", err)
	}
	if got != want {
		t.Errorf("RefUUIDChecked = %s, want %s", got, want)
	}
}

// TestRefUUIDCheckedAbsentVariants verifies that a nil interface, a typed-nil
// pointer, a JSON null and an integer id all decode to uuid.Nil without error.
//
// Why it matters: Nautobot omits references freely and may emit integer ids;
// callers treat uuid.Nil as "absent" and must not see spurious shape errors.
// Inputs: the four absent/integer shapes. Outputs: uuid.Nil and a nil error.
// Data choice: the typed-nil pointer case is the one that would panic inside
// a value-receiver MarshalJSON if the guard were missing.
func TestRefUUIDCheckedAbsentVariants(t *testing.T) {
	var typedNil *Device_Status_Id
	cases := map[string]interface{ MarshalJSON() ([]byte, error) }{
		"nil interface": nil,
		"typed nil":     typedNil,
		"json null":     rawRef{b: []byte("null")},
		"integer id":    rawRef{b: []byte("42")},
	}
	for name, m := range cases {
		got, err := RefUUIDChecked(m)
		if err != nil {
			t.Errorf("%s: unexpected error %v", name, err)
		}
		if got != uuid.Nil {
			t.Errorf("%s: got %s, want uuid.Nil", name, got)
		}
	}
}

// TestRefUUIDCheckedReportsMalformedOutput verifies that marshal failures,
// empty output, non-JSON bytes, unexpected JSON types and bad UUID strings are
// returned as errors naming the union type.
//
// Why it matters: generated-shape drift after a client regeneration must be
// loud at the first decode rather than degrade every reference to uuid.Nil.
// Inputs: five malformed marshal outputs. Outputs: a non-nil error for each,
// and the marshal error is wrapped so errors.Is still finds it.
// Data choice: the JSON object case models a union that marshals its struct
// instead of its member, the most likely regeneration regression.
func TestRefUUIDCheckedReportsMalformedOutput(t *testing.T) {
	marshalErr := errors.New("boom")
	cases := map[string]rawRef{
		"marshal error": {err: marshalErr},
		"empty output":  {b: []byte{}},
		"not json":      {b: []byte("{")},
		"json object":   {b: []byte(`{"id":1}`)},
		"bad uuid":      {b: []byte(`"not-a-uuid"`)},
	}
	for name, m := range cases {
		got, err := RefUUIDChecked(m)
		if err == nil {
			t.Errorf("%s: expected an error", name)
			continue
		}
		if got != uuid.Nil {
			t.Errorf("%s: got %s, want uuid.Nil on error", name, got)
		}
		if !strings.Contains(err.Error(), "rawRef") {
			t.Errorf("%s: error should name the union type, got %q", name, err)
		}
	}
	if _, err := RefUUIDChecked(rawRef{err: marshalErr}); !errors.Is(err, marshalErr) {
		t.Errorf("marshal error not wrapped: %v", err)
	}
}
