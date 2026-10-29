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
func (e *Exporter) previewPlannedCable(cable *devicetypes.CaniCableType, a, b cableEnd, result *LoadResult) bool {
	if !e.Options.DryRun {
		return false
	}
	plannedA, plannedB := e.endPlanned(a), e.endPlanned(b)
	if !plannedA && !plannedB {
		return false
	}
	if (!plannedA && e.interfaceCabled(a)) || (!plannedB && e.interfaceCabled(b)) {
		result.CablesConflicted++
		return true
	}
	clog.DryRun("Would create cable: %s (%s:%s <-> %s:%s)",
		cable.Label, a.device.Name, cable.TerminationAPort, b.device.Name, cable.TerminationBPort)
	result.CablesCreated++
	return true
}

// endPlanned reports whether the dry run stands in for the end's device or,
// for a module-owned port, for its module: neither exists in Nautobot yet,
// so the interface cannot be looked up and the cable is previewed instead.
func (e *Exporter) endPlanned(end cableEnd) bool {
	if e.Cache.DevicePlanned(end.device.Name) {
		return true
	}
	return end.module != nil && e.Cache.ModulePlanned(end.module.ID)
}

// interfaceCabled reports whether the end's interface, on a device that is
// already in Nautobot, carries a cable, logging the skip the create path logs
// for the same conflict. The interface is resolved as the create path
// resolves it, so a module port is looked up on its module. A device or
// interface the lookup cannot find is not a conflict here; the create path
// reports those itself.
func (e *Exporter) interfaceCabled(end cableEnd) bool {
	device, err := e.Cache.GetDeviceByName(end.device.Name)
	if err != nil || device == nil {
		return false
	}
	iface, err := e.resolveOwnedInterface(end.moduleID, device.ID, end.port)
	if err != nil || iface == nil || iface.CableID == uuid.Nil {
		return false
	}
	clog.Skipped("Skipped cable (interface already cabled): %s:%s", end.device.Name, end.port)
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
