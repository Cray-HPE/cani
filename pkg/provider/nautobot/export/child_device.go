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

import "github.com/Cray-HPE/cani/pkg/devicetypes"

// subdeviceRoleChild is the device-type subdevice_role for devices that are
// installed in a parent device's bay rather than mounted in a rack.
const subdeviceRoleChild = "child"

// isChildDevice reports whether device sits inside a parent device (for
// example a half-width switch in a chassis) or has a child device type.
// Nautobot rejects rack position and face on child devices because those
// belong to the parent.
func isChildDevice(device *devicetypes.CaniDeviceType, inv *devicetypes.Inventory) bool {
	if device.SubdeviceRole == subdeviceRoleChild {
		return true
	}
	if inv == nil {
		return false
	}
	parent, ok := inv.Devices[device.Parent]
	return ok && parent != nil && parent.Type != devicetypes.Rack
}

// setRackSlot sets the rack position and face on a device request, leaving
// both unset for child devices.
func setRackSlot(device *devicetypes.CaniDeviceType, inv *devicetypes.Inventory, position **int, face any) {
	if isChildDevice(device, inv) {
		return
	}
	if device.RackPosition > 0 {
		pos := device.RackPosition
		*position = &pos
	}
	setDeviceFace(face, device.Face)
}
