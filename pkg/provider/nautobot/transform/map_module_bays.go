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
	"slices"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// ApplyModuleBays records on each imported device the module bays Nautobot
// declares for it, so the merge resolves a module's bay against the source
// system's names and the library template supplies bays only for a device
// that declares none (devicetypes.moduleBaysOf). Without this, a bay the
// library spells differently is cleared on import and the module cannot be
// exported again. A bay nested in a module has no parent device and is
// skipped; a device's bays are sorted by name so the datastore is stable
// across imports. Returns the number of bays recorded.
func ApplyModuleBays(
	bays []nautobotapi.ModuleBay,
	devices map[uuid.UUID]*devicetypes.CaniDeviceType,
	deviceMap map[uuid.UUID]uuid.UUID,
) int {
	recorded := 0
	touched := map[*devicetypes.CaniDeviceType]bool{}
	for _, bay := range bays {
		if bay.ParentDevice == nil {
			continue
		}
		device := devices[deviceMap[tenantRefID(bay.ParentDevice)]]
		if device == nil || declaresModuleBay(device, bay.Name) {
			continue
		}
		device.ModuleBays = append(device.ModuleBays, devicetypes.ModuleBaySpec{Name: bay.Name, Position: strVal(bay.Position)})
		touched[device] = true
		recorded++
	}
	for device := range touched {
		slices.SortFunc(device.ModuleBays, func(a, b devicetypes.ModuleBaySpec) int { return strings.Compare(a.Name, b.Name) })
	}
	return recorded
}

// declaresModuleBay reports whether the device instance already names the bay.
func declaresModuleBay(device *devicetypes.CaniDeviceType, name string) bool {
	return slices.ContainsFunc(device.ModuleBays, func(spec devicetypes.ModuleBaySpec) bool {
		return strings.EqualFold(spec.Name, name)
	})
}
