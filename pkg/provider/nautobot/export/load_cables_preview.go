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
package export

import (
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// previewPlannedCable reports, under a dry run, a cable with an end on a
// device the run has only planned: its interfaces do not exist yet, so the
// create is previewed here rather than failing the device lookup. An end
// that does exist in Nautobot is still checked for a cable already attached,
// so the preview skips the same conflicts the real run would.
func (e *Exporter) previewPlannedCable(cable *devicetypes.CaniCableType, deviceA, deviceB *devicetypes.CaniDeviceType, result *LoadResult) bool {
	if !e.Options.DryRun {
		return false
	}
	plannedA, plannedB := e.Cache.DevicePlanned(deviceA.Name), e.Cache.DevicePlanned(deviceB.Name)
	if !plannedA && !plannedB {
		return false
	}
	if (!plannedA && e.interfaceCabled(deviceA.Name, cable.TerminationAPort)) ||
		(!plannedB && e.interfaceCabled(deviceB.Name, cable.TerminationBPort)) {
		result.CablesConflicted++
		return true
	}
	clog.DryRun("Would create cable: %s (%s:%s <-> %s:%s)",
		cable.Label, deviceA.Name, cable.TerminationAPort, deviceB.Name, cable.TerminationBPort)
	result.CablesCreated++
	return true
}

// interfaceCabled reports whether the named interface of a device that is
// already in Nautobot carries a cable, logging the skip the create path logs
// for the same conflict. A device or interface the lookup cannot find is not
// a conflict here; the create path reports those itself.
func (e *Exporter) interfaceCabled(deviceName, port string) bool {
	device, err := e.Cache.GetDeviceByName(deviceName)
	if err != nil || device == nil {
		return false
	}
	iface, err := e.Cache.GetInterfaceByDeviceAndNameFuzzy(device.ID, port)
	if err != nil || iface == nil || iface.CableID == uuid.Nil {
		return false
	}
	clog.Skipped("Skipped cable (interface already cabled): %s:%s", deviceName, port)
	return true
}

// cableStatusName is the status the cable phase gives a cable: Planned when
// the cable says so, Connected otherwise.
func cableStatusName(cable *devicetypes.CaniCableType) string {
	if strings.EqualFold(cable.Status, "planned") {
		return string(devicetypes.StatusPlanned)
	}
	return string(devicetypes.StatusConnected)
}
