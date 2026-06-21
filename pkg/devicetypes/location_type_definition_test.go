/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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

// TestLocationTypeDefinitionValidate verifies Validate accepts a definition
// with both Name and Slug and rejects one missing either field.
//
// Why it matters: LocationTypeDefinition maps 1:1 to a Nautobot LocationType,
// so a definition with no name or slug cannot be created upstream and must be
// caught at load time.
// Inputs: a complete definition, then ones missing Name and missing Slug.
// Outputs: nil error for the complete case; a non-nil error for each missing
// field.
// Data choice: toggling exactly one required field per case proves each guard
// is checked independently rather than as a single combined condition.
func TestLocationTypeDefinitionValidate(t *testing.T) {
	cases := []struct {
		name    string
		def     LocationTypeDefinition
		wantErr bool
	}{
		{"valid", LocationTypeDefinition{Name: "Site", Slug: "site"}, false},
		{"missing name", LocationTypeDefinition{Slug: "site"}, true},
		{"missing slug", LocationTypeDefinition{Name: "Site"}, true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.def.Validate()
			if tt.wantErr && err == nil {
				t.Errorf("Validate() = nil, want error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() = %v, want nil", err)
			}
		})
	}
}
