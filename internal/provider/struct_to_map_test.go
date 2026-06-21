/*
 *
 *  MIT License
 *
 *  (C) Copyright 2024-2026 Hewlett Packard Enterprise Development LP
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
package provider

import "testing"

type smInner struct {
	Value string `yaml:"value"`
}

type smSample struct {
	Named      string   `yaml:"named_tag,omitempty"` // comma-option tag
	DashTag    string   `yaml:"-"`                   // "-" falls back to field name
	NoTag      int      // no tag falls back to field name
	Nested     smInner  `yaml:"nested"`     // recursive struct
	PtrStruct  *smInner `yaml:"ptr_struct"` // non-nil pointer to struct
	PtrScalar  *int     `yaml:"ptr_scalar"` // non-nil pointer to scalar
	NilPtr     *smInner `yaml:"nil_ptr"`    // nil pointer (skipped)
	unexported string   // unexported (skipped)
}

// TestStructToMapAllNonStruct verifies StructToMapAll returns an empty map for
// inputs that are not structs (after optional pointer dereferencing).
//
// Why it matters: the converter is fed arbitrary provider values during YAML
// serialization, so a non-struct must yield an empty map rather than panic on a
// reflect call that assumes struct fields.
// Inputs: an int value and a nil *smInner pointer. Outputs: an empty map for
// each. Data choice: a bare scalar and a nil typed pointer cover both the
// "kind is not struct" return and the pointer path that dereferences to an
// invalid value.
func TestStructToMapAllNonStruct(t *testing.T) {
	if got := StructToMapAll(42); len(got) != 0 {
		t.Errorf("StructToMapAll(int) = %v, want empty map", got)
	}
	var nilPtr *smInner
	if got := StructToMapAll(nilPtr); len(got) != 0 {
		t.Errorf("StructToMapAll(nil ptr) = %v, want empty map", got)
	}
}

// TestStructToMapAll verifies StructToMapAll converts every exported field of a
// struct (reached via a pointer), honoring yaml tag names, recursing into nested
// and pointed-to structs, unwrapping scalar pointers, and skipping nil pointers
// and unexported fields.
//
// Why it matters: this function serializes provider option structs into the
// config map, so each reflection branch — tag parsing, struct recursion, pointer
// handling, and field-visibility filtering — must behave exactly, or config
// output would lose or mislabel fields.
// Inputs: a pointer to a struct populated with a comma-option tag, a "-" tag, an
// untagged field, a nested struct, a non-nil struct pointer, a non-nil scalar
// pointer, a nil pointer, and an unexported field. Outputs: a map keyed by the
// resolved tag/field names with nested structs rendered as sub-maps, the scalar
// pointer unwrapped, and the nil-pointer and unexported keys absent. Data
// choice: one struct carrying every field shape exercises all switch arms and
// both tag-fallback paths in a single deterministic pass.
func TestStructToMapAll(t *testing.T) {
	scalar := 7
	s := smSample{
		Named:      "n1",
		DashTag:    "d1",
		NoTag:      42,
		Nested:     smInner{Value: "nv"},
		PtrStruct:  &smInner{Value: "pv"},
		PtrScalar:  &scalar,
		NilPtr:     nil,
		unexported: "hidden",
	}

	got := StructToMapAll(&s) // pointer input exercises the top-level deref

	if got["named_tag"] != "n1" {
		t.Errorf("named_tag = %v, want n1 (comma option stripped)", got["named_tag"])
	}
	if got["DashTag"] != "d1" {
		t.Errorf("DashTag = %v, want d1 (\"-\" falls back to field name)", got["DashTag"])
	}
	if got["NoTag"] != 42 {
		t.Errorf("NoTag = %v, want 42 (untagged uses field name)", got["NoTag"])
	}

	nested, ok := got["nested"].(map[string]any)
	if !ok || nested["value"] != "nv" {
		t.Errorf("nested = %v, want sub-map with value=nv", got["nested"])
	}
	ptrStruct, ok := got["ptr_struct"].(map[string]any)
	if !ok || ptrStruct["value"] != "pv" {
		t.Errorf("ptr_struct = %v, want sub-map with value=pv", got["ptr_struct"])
	}
	if got["ptr_scalar"] != 7 {
		t.Errorf("ptr_scalar = %v, want 7 (scalar pointer unwrapped)", got["ptr_scalar"])
	}

	if _, present := got["nil_ptr"]; present {
		t.Error("nil_ptr should be omitted")
	}
	if _, present := got["unexported"]; present {
		t.Error("unexported field should be skipped")
	}
}
