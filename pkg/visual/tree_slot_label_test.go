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
package visual

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// TestRackSlotLabel verifies the tree label shows a rack U range, a single U,
// or the device bay of a child, and nothing for an unplaced device.
//
// Why it matters: children in a chassis have no rack position, so the bay is
// the only placement the tree view can show for them.
func TestRackSlotLabel(t *testing.T) {
	cases := []struct {
		dev  devicetypes.CaniDeviceType
		want string
	}{
		{devicetypes.CaniDeviceType{RackPosition: 10, UHeight: 2}, "U10-U9"},
		{devicetypes.CaniDeviceType{RackPosition: 15, UHeight: 1}, "U15"},
		{devicetypes.CaniDeviceType{ParentBay: "Bay 2"}, "[Bay 2]"},
		{devicetypes.CaniDeviceType{}, ""},
	}
	for _, c := range cases {
		if got := rackSlotLabel(&c.dev); got != c.want {
			t.Errorf("rackSlotLabel(%+v) = %q, want %q", c.dev, got, c.want)
		}
	}
}
