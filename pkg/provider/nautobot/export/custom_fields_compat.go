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
package export

// Nautobot 3.2 changed custom_fields from map[string]interface{} to
// map[string]*interface{} (pointer values). These helpers bridge that shape
// against the flat maps CANI uses internally.

// derefCustomFields converts a Nautobot custom-fields map (pointer values) into
// a flat map[string]interface{} for comparison and formatting.
func derefCustomFields(cf *map[string]*interface{}) map[string]interface{} {
	if cf == nil {
		return nil
	}
	out := make(map[string]interface{}, len(*cf))
	for k, v := range *cf {
		if v != nil {
			out[k] = *v
		} else {
			out[k] = nil
		}
	}
	return out
}

// toNautobotCustomFields converts a flat custom-fields map into the Nautobot
// 3.2 request shape (pointer values). Returns nil for a nil map.
func toNautobotCustomFields(cf map[string]interface{}) *map[string]*interface{} {
	if cf == nil {
		return nil
	}
	out := make(map[string]*interface{}, len(cf))
	for k, v := range cf {
		v := v
		out[k] = &v
	}
	return &out
}
