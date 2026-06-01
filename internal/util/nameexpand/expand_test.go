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
package nameexpand

import (
	"reflect"
	"testing"
)

func TestExpand_NumericRanges(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    []string
	}{
		{
			name:    "simple numeric",
			pattern: "x370{1..4}",
			want:    []string{"x3701", "x3702", "x3703", "x3704"},
		},
		{
			name:    "zero padded",
			pattern: "x{01..04}",
			want:    []string{"x01", "x02", "x03", "x04"},
		},
		{
			name:    "zero padded crossing tens",
			pattern: "node{08..12}",
			want:    []string{"node08", "node09", "node10", "node11", "node12"},
		},
		{
			name:    "wide padding from start",
			pattern: "x{001..010}",
			want:    []string{"x001", "x002", "x003", "x004", "x005", "x006", "x007", "x008", "x009", "x010"},
		},
		{
			name:    "single value range",
			pattern: "rack{5..5}",
			want:    []string{"rack5"},
		},
		{
			name:    "prefix and suffix",
			pattern: "rack-{1..3}-prod",
			want:    []string{"rack-1-prod", "rack-2-prod", "rack-3-prod"},
		},
		{
			name:    "hundreds place",
			pattern: "sw{100..103}",
			want:    []string{"sw100", "sw101", "sw102", "sw103"},
		},
		{
			name:    "thousands crossing",
			pattern: "n{997..1002}",
			want:    []string{"n0997", "n0998", "n0999", "n1000", "n1001", "n1002"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Expand(tt.pattern)
			if err != nil {
				t.Fatalf("Expand(%q) unexpected error: %v", tt.pattern, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expand(%q)\n  got  %v\n  want %v", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestExpand_AlphaRanges(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    []string
	}{
		{
			name:    "lowercase letters",
			pattern: "rack-{a..d}",
			want:    []string{"rack-a", "rack-b", "rack-c", "rack-d"},
		},
		{
			name:    "uppercase letters",
			pattern: "RACK-{A..D}",
			want:    []string{"RACK-A", "RACK-B", "RACK-C", "RACK-D"},
		},
		{
			name:    "wrapping z to aa",
			pattern: "shelf-{y..ab}",
			want:    []string{"shelf-y", "shelf-z", "shelf-aa", "shelf-ab"},
		},
		{
			name:    "single letter range",
			pattern: "{a..a}",
			want:    []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Expand(tt.pattern)
			if err != nil {
				t.Fatalf("Expand(%q) unexpected error: %v", tt.pattern, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expand(%q)\n  got  %v\n  want %v", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestExpand_CartesianProduct(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    []string
	}{
		{
			name:    "alpha then numeric",
			pattern: "{a..b}{1..3}",
			want:    []string{"a1", "a2", "a3", "b1", "b2", "b3"},
		},
		{
			name:    "prefix alpha numeric suffix",
			pattern: "x{a..b}-{01..02}-z",
			want:    []string{"xa-01-z", "xa-02-z", "xb-01-z", "xb-02-z"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Expand(tt.pattern)
			if err != nil {
				t.Fatalf("Expand(%q) unexpected error: %v", tt.pattern, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expand(%q)\n  got  %v\n  want %v", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestExpand_CommaLists(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    []string
	}{
		{
			name:    "numeric comma list",
			pattern: "port{9,10,11,12}",
			want:    []string{"port9", "port10", "port11", "port12"},
		},
		{
			name:    "alpha comma list",
			pattern: "rack-{a,b,c}",
			want:    []string{"rack-a", "rack-b", "rack-c"},
		},
		{
			name:    "device name list",
			pattern: "GH-x3701u{34,26,18,10}",
			want:    []string{"GH-x3701u34", "GH-x3701u26", "GH-x3701u18", "GH-x3701u10"},
		},
		{
			name:    "mixed with range",
			pattern: "{a,b}{1..3}",
			want:    []string{"a1", "a2", "a3", "b1", "b2", "b3"},
		},
		{
			name:    "no zero padding",
			pattern: "{9,10,11,12}",
			want:    []string{"9", "10", "11", "12"},
		},
		{
			name:    "comma list with spaces trimmed",
			pattern: "x{1, 2, 3}",
			want:    []string{"x1", "x2", "x3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Expand(tt.pattern)
			if err != nil {
				t.Fatalf("Expand(%q) unexpected error: %v", tt.pattern, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expand(%q)\n  got  %v\n  want %v", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestExpand_CommaList_Errors(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
	}{
		{name: "empty element", pattern: "x{1,,3}"},
		{name: "trailing comma", pattern: "x{1,2,}"},
		{name: "leading comma", pattern: "x{,1,2}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Expand(tt.pattern)
			if err == nil {
				t.Errorf("Expand(%q) expected error, got nil", tt.pattern)
			}
		})
	}
}

func TestExpand_NoBraces(t *testing.T) {
	got, err := Expand("plainname")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"plainname"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExpand_Errors(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
	}{
		{name: "reversed numeric", pattern: "x{5..2}"},
		{name: "reversed alpha", pattern: "x{d..a}"},
		{name: "mixed types", pattern: "x{a..3}"},
		{name: "unclosed brace", pattern: "x{1..4"},
		{name: "no dots", pattern: "x{14}"},
		{name: "empty start", pattern: "x{..4}"},
		{name: "empty end", pattern: "x{1..}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Expand(tt.pattern)
			if err == nil {
				t.Errorf("Expand(%q) expected error, got nil", tt.pattern)
			}
		})
	}
}

func TestAlphaToIndex(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"a", 0},
		{"b", 1},
		{"z", 25},
		{"aa", 26},
		{"ab", 27},
		{"az", 51},
		{"ba", 52},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := alphaToIndex(tt.input)
			if got != tt.want {
				t.Errorf("alphaToIndex(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestIndexToAlpha(t *testing.T) {
	tests := []struct {
		idx   int
		upper bool
		want  string
	}{
		{0, false, "a"},
		{25, false, "z"},
		{26, false, "aa"},
		{27, false, "ab"},
		{0, true, "A"},
		{26, true, "AA"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := indexToAlpha(tt.idx, tt.upper)
			if got != tt.want {
				t.Errorf("indexToAlpha(%d, %v) = %q, want %q", tt.idx, tt.upper, got, tt.want)
			}
		})
	}
}
