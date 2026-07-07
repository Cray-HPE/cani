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

// TestValidateInterfaceMode verifies switchport-mode normalization and rejection
// of unsupported values.
//
// Why it matters: interface mode is exported to Nautobot as a constrained enum;
// accepting an out-of-range value would produce an API error at export time, so
// the CLI must reject it up front and normalize case/whitespace.
// Inputs: the three valid modes (with mixed case/spacing), the empty string, and
// a bogus value. Outputs: normalized lowercase mode for valid inputs, empty for
// empty, and an error for the bogus value.
// Data choice: covers each accepted mode plus the empty (clear) and invalid
// branches so every switch arm is exercised.
func TestValidateInterfaceMode(t *testing.T) {
	valid := map[string]string{
		"access":       "access",
		"Tagged":       "tagged",
		" tagged-all ": "tagged-all",
		"":             "",
	}
	for in, want := range valid {
		got, err := ValidateInterfaceMode(in)
		if err != nil {
			t.Errorf("ValidateInterfaceMode(%q) unexpected error: %v", in, err)
		}
		if got != want {
			t.Errorf("ValidateInterfaceMode(%q) = %q, want %q", in, got, want)
		}
	}
	if _, err := ValidateInterfaceMode("trunk"); err == nil {
		t.Error(`ValidateInterfaceMode("trunk") expected error, got nil`)
	}
}
