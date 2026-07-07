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
package update

import "testing"

// TestParseVIDs verifies that string VLAN IDs are converted and validated, that
// blank entries are skipped, and that bad values are rejected.
//
// Why it matters: the --tagged-vlan flag arrives as strings; converting and
// range-checking them here keeps invalid trunk memberships from reaching the
// exporter and gives the user an immediate, clear error.
// Inputs: a mix of valid IDs with surrounding/blank whitespace, then a
// non-numeric value and an out-of-range value. Outputs: the parsed int slice for
// valid input, and errors for the bad inputs.
// Data choice: an empty element proves blanks are skipped; the two failure cases
// cover both error branches (non-numeric and out-of-range).
func TestParseVIDs(t *testing.T) {
	got, err := parseVIDs([]string{"1", "", " 2000 "})
	if err != nil {
		t.Fatalf("parseVIDs valid: unexpected error: %v", err)
	}
	want := []int{1, 2000}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("parseVIDs = %v, want %v", got, want)
	}
	if _, err := parseVIDs([]string{"abc"}); err == nil {
		t.Error("parseVIDs(non-numeric) expected error, got nil")
	}
	if _, err := parseVIDs([]string{"5000"}); err == nil {
		t.Error("parseVIDs(out-of-range) expected error, got nil")
	}
}
