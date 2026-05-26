/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
package nautobot

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/nautobot/transform"
)

// Transform transforms raw Nautobot data into CANI's format.
// When raw data has been fetched (via Import()), it is passed to the
// transform package for full entity mapping.  Otherwise the legacy
// copy-existing-devices path is used.
func (p *Nautobot) Transform(existing devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	var raw *transform.RawData
	if len(p.rawDevices) > 0 || len(p.rawLocations) > 0 {
		raw = &transform.RawData{
			Locations:      p.rawLocations,
			Racks:          p.rawRacks,
			Devices:        p.rawDevices,
			DeviceTypes:    p.rawDeviceTypes,
			Interfaces:     p.rawInterfaces,
			Modules:        p.rawModules,
			ModuleBays:     p.rawModuleBays,
			Cables:         p.rawCables,
			InventoryItems: p.rawInventoryItems,
			Statuses:       p.rawStatuses,
			Roles:          p.rawRoles,
		}
	}
	return transform.Transform(existing, raw)
}
