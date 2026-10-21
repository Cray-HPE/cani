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
package transform

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// ApplyDeviceBays installs imported child devices into their parent's bay.
// For each Nautobot device bay with an installed device, the bay's device
// becomes the child's parent and the bay name its ParentBay. A child takes
// its rack placement from the parent, so its own position and face are
// cleared. deviceMap maps Nautobot device UUIDs to CANI device UUIDs.
// Returns the number of children placed.
func ApplyDeviceBays(
	bays []nautobotapi.DeviceBay,
	devices map[uuid.UUID]*devicetypes.CaniDeviceType,
	deviceMap map[uuid.UUID]uuid.UUID,
) int {
	placed := 0
	for _, bay := range bays {
		if bay.InstalledDevice == nil {
			continue
		}
		child := devices[deviceMap[refID(bay.InstalledDevice)]]
		parentID, ok := deviceMap[refID(bay.Device)]
		if child == nil || !ok {
			continue
		}
		child.Parent, child.ParentBay = parentID, bay.Name
		child.RackPosition, child.Face = 0, ""
		placed++
	}
	return placed
}
