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

func TestResolveNames_BraceExpansion(t *testing.T) {
	tests := []struct {
		name     string
		nameFlag string
		qty      int
		want     []string
	}{
		{
			name:     "numeric brace expansion",
			nameFlag: "x370{1..4}", qty: 4,
			want: []string{"x3701", "x3702", "x3703", "x3704"},
		},
		{
			name:     "alpha brace expansion",
			nameFlag: "rack-{a..c}", qty: 3,
			want: []string{"rack-a", "rack-b", "rack-c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveNames(tt.nameFlag, "", 1, 0, tt.qty)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveNames_Prefix(t *testing.T) {
	got, err := ResolveNames("", "x370", 1, 0, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"x3701", "x3702", "x3703", "x3704"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestResolveNames_PlainName(t *testing.T) {
	got, err := ResolveNames("myRack", "", 1, 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"myRack"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestResolveNames_NoFlags(t *testing.T) {
	got, err := ResolveNames("", "", 1, 0, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestResolveNamesTemplateDeferral(t *testing.T) {
	got, err := ResolveNames("gh-%{RACK}u%{U}", "", 1, 0, 12)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for template deferral, got %v", got)
	}
}

func TestResolveNames_Errors(t *testing.T) {
	tests := []struct {
		name     string
		nameFlag string
		prefix   string
		qty      int
	}{
		{
			name:     "mutual exclusion",
			nameFlag: "x{1..4}", prefix: "x", qty: 4,
		},
		{
			name:     "count mismatch",
			nameFlag: "x{1..4}", prefix: "", qty: 3,
		},
		{
			name:     "plain name with qty > 1",
			nameFlag: "rack", prefix: "", qty: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ResolveNames(tt.nameFlag, tt.prefix, 1, 0, tt.qty)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}
