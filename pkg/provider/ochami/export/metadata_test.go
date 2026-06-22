package export

import (
	"reflect"
	"testing"
)

type stringerValue string

func (s stringerValue) String() string { return string(s) }

// TestExtractString verifies metadata string extraction accepts strings and
// fmt.Stringer values while rejecting missing or non-string values.
//
// Why it matters: Ochami export builds OpenCHAMI xname, mac, and ip fields from
// untyped provider metadata, so only deliberate string-like values should become
// YAML fields.
// Inputs: metadata maps with present, fmt.Stringer, missing, nil, and wrong-type
// xname values. Outputs: the extracted string or an empty string.
// Data choice: xname is the primary OpenCHAMI identifier and a representative
// string metadata key.
func TestExtractString(t *testing.T) {
	tests := []struct {
		name string
		meta map[string]any
		key  string
		want string
	}{
		{"present", map[string]any{"xname": "x3000c0s1b0"}, "xname", "x3000c0s1b0"},
		{"stringer", map[string]any{"xname": stringerValue("x3000c0s2b0")}, "xname", "x3000c0s2b0"},
		{"missing", map[string]any{"other": "val"}, "xname", ""},
		{"nil map", nil, "xname", ""},
		{"wrong type", map[string]any{"xname": 123}, "xname", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractString(tt.meta, tt.key)
			if got != tt.want {
				t.Errorf("extractString(%v, %q) = %q, want %q", tt.meta, tt.key, got, tt.want)
			}
		})
	}
}

// TestExtractIntPtr verifies metadata integer pointer extraction accepts int,
// float64, and numeric string values while rejecting absent or invalid values.
//
// Why it matters: older Ochami export metadata included numeric node IDs in an
// untyped map, and this helper should preserve only parseable values.
// Inputs: metadata maps with int, float64, numeric string, bad string, missing,
// nil, and wrong-type values. Outputs: an integer pointer for parseable values or
// nil.
// Data choice: the cases mirror common YAML/JSON round-trip representations for
// numeric fields.
func TestExtractIntPtr(t *testing.T) {
	tests := []struct {
		name    string
		meta    map[string]any
		key     string
		wantNil bool
		wantVal int
	}{
		{"int", map[string]any{"nid": 42}, "nid", false, 42},
		{"float64", map[string]any{"nid": float64(99)}, "nid", false, 99},
		{"string", map[string]any{"nid": "7"}, "nid", false, 7},
		{"bad string", map[string]any{"nid": "abc"}, "nid", true, 0},
		{"missing", map[string]any{}, "nid", true, 0},
		{"nil map", nil, "nid", true, 0},
		{"wrong type", map[string]any{"nid": []string{"7"}}, "nid", true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractIntPtr(tt.meta, tt.key)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %d", *got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil, got nil")
			}
			if *got != tt.wantVal {
				t.Errorf("got %d, want %d", *got, tt.wantVal)
			}
		})
	}
}

// TestExtractStringSlice verifies metadata string-slice extraction accepts native
// string slices and YAML/JSON round-tripped []any values.
//
// Why it matters: older Ochami export metadata included host aliases as an
// untyped list, and callers should get deterministic []string values only from
// list-like metadata.
// Inputs: metadata maps with []string, []any, missing, nil, and wrong-type
// host_aliases values. Outputs: a string slice or nil.
// Data choice: host_aliases is the representative list field and exercises both
// native and round-tripped collection shapes.
func TestExtractStringSlice(t *testing.T) {
	tests := []struct {
		name string
		meta map[string]any
		key  string
		want []string
	}{
		{"[]string", map[string]any{"host_aliases": []string{"a", "b"}}, "host_aliases", []string{"a", "b"}},
		{"[]any", map[string]any{"host_aliases": []any{"c", "d"}}, "host_aliases", []string{"c", "d"}},
		{"missing", map[string]any{}, "host_aliases", nil},
		{"nil map", nil, "host_aliases", nil},
		{"wrong type", map[string]any{"host_aliases": 123}, "host_aliases", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractStringSlice(tt.meta, tt.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractStringSlice(%v, %q) = %v, want %v", tt.meta, tt.key, got, tt.want)
			}
		})
	}
}
