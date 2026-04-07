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

func TestSequence(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		start    int
		count    int
		padWidth int
		want     []string
	}{
		{
			name:   "xname style",
			prefix: "x370", start: 1, count: 4, padWidth: 0,
			want: []string{"x3701", "x3702", "x3703", "x3704"},
		},
		{
			name:   "explicit pad width",
			prefix: "node-", start: 1, count: 3, padWidth: 3,
			want: []string{"node-001", "node-002", "node-003"},
		},
		{
			name:   "auto pad from large end",
			prefix: "sw", start: 98, count: 5, padWidth: 0,
			want: []string{"sw098", "sw099", "sw100", "sw101", "sw102"},
		},
		{
			name:   "single item",
			prefix: "rack-", start: 42, count: 1, padWidth: 0,
			want: []string{"rack-42"},
		},
		{
			name:   "thousands",
			prefix: "cab", start: 9998, count: 4, padWidth: 0,
			want: []string{"cab09998", "cab09999", "cab10000", "cab10001"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Sequence(tt.prefix, tt.start, tt.count, tt.padWidth)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sequence(%q, %d, %d, %d)\n  got  %v\n  want %v",
					tt.prefix, tt.start, tt.count, tt.padWidth, got, tt.want)
			}
		})
	}
}

func TestSequenceAlpha(t *testing.T) {
	tests := []struct {
		name        string
		prefix      string
		startLetter string
		count       int
		want        []string
	}{
		{
			name:   "simple lowercase",
			prefix: "rack-", startLetter: "a", count: 4,
			want: []string{"rack-a", "rack-b", "rack-c", "rack-d"},
		},
		{
			name:   "wrapping z to aa",
			prefix: "shelf-", startLetter: "y", count: 4,
			want: []string{"shelf-y", "shelf-z", "shelf-aa", "shelf-ab"},
		},
		{
			name:   "uppercase",
			prefix: "R", startLetter: "A", count: 3,
			want: []string{"RA", "RB", "RC"},
		},
		{
			name:   "multi char start",
			prefix: "bay-", startLetter: "aa", count: 3,
			want: []string{"bay-aa", "bay-ab", "bay-ac"},
		},
		{
			name:   "single item",
			prefix: "", startLetter: "z", count: 1,
			want: []string{"z"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SequenceAlpha(tt.prefix, tt.startLetter, tt.count)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SequenceAlpha(%q, %q, %d)\n  got  %v\n  want %v",
					tt.prefix, tt.startLetter, tt.count, got, tt.want)
			}
		})
	}
}
