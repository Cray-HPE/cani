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
package devicetypes

import "testing"

// TestValidateVID verifies VLAN-ID range enforcement.
//
// Why it matters: VLAN IDs outside 1-4094 are invalid in 802.1Q and rejected by
// Nautobot; validating locally gives the user a clear error before export.
// Inputs: the inclusive boundaries (1 and 4094) and out-of-range values (0,
// 4095, -1). Outputs: nil for in-range, error for out-of-range.
// Data choice: boundary values catch off-by-one errors in the range check.
func TestValidateVID(t *testing.T) {
	for _, vid := range []int{1, 4094} {
		if err := ValidateVID(vid); err != nil {
			t.Errorf("ValidateVID(%d) unexpected error: %v", vid, err)
		}
	}
	for _, vid := range []int{0, 4095, -1} {
		if err := ValidateVID(vid); err == nil {
			t.Errorf("ValidateVID(%d) expected error, got nil", vid)
		}
	}
}
