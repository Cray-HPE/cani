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

import "github.com/google/uuid"

// GetRackID returns the rack ID for this device. Prefers the explicit Rack FK;
// falls back to checking if Parent exists in the inventory's Racks collection.
// Returns uuid.Nil if not in a rack.
func (c *CaniDeviceType) GetRackID(inv *Inventory) uuid.UUID {
	if c == nil || inv == nil {
		return uuid.Nil
	}
	if c.Rack != uuid.Nil {
		return c.Rack
	}
	if _, ok := inv.Racks[c.Parent]; ok {
		return c.Parent
	}
	return uuid.Nil
}

// GetUHeight returns the U-height for this device.
// Returns 1 as default if not set.
func (c *CaniDeviceType) GetUHeight() int {
	if c == nil || c.UHeight < 1 {
		return 1
	}
	return c.UHeight
}
