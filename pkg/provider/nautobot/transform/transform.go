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
package transform

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/Cray-HPE/cani/pkg/provider/nautobot/logcolor"
	"github.com/google/uuid"
)

var clog = logcolor.New("[nautobot] ", false)

// RawData holds all raw API responses fetched during Import.
type RawData struct {
	Locations      []nautobotapi.Location
	Racks          []nautobotapi.Rack
	Devices        []nautobotapi.Device
	DeviceTypes    []nautobotapi.DeviceType
	Interfaces     []nautobotapi.Interface
	Modules        []nautobotapi.Module
	ModuleBays     []nautobotapi.ModuleBay
	Cables         []nautobotapi.Cable
	InventoryItems []nautobotapi.InventoryItem
	Statuses       []nautobotapi.Status
	Roles          []nautobotapi.Role
}

// Transform converts raw Nautobot API data into a TransformResult.
// When raw is nil (legacy path), it falls back to copying existing
// devices from the inventory.
func Transform(existing devicetypes.Inventory, raw *RawData) (*devicetypes.TransformResult, error) {
	if raw == nil {
		return transformLegacy(existing)
	}
	return transformRaw(raw)
}

// transformRaw runs all entity mappers against raw API data.
func transformRaw(raw *RawData) (*devicetypes.TransformResult, error) {
	statusNameMap := BuildStatusNameMap(raw.Statuses)
	roleNameMap := BuildRoleNameMap(raw.Roles)

	// 1. Locations – also produces Nautobot→CANI UUID map.
	locations, locationMap := MapLocations(raw.Locations, statusNameMap)
	clog.Detail("  Transformed %d locations", len(locations))

	// 2. Racks.
	racks, rackMap := MapRacks(raw.Racks, locationMap, statusNameMap, roleNameMap)
	clog.Detail("  Transformed %d racks", len(racks))

	// 3. Pre-build lookup tables needed by devices, cables, modules.
	ifacesByDevice := GroupInterfacesByDevice(raw.Interfaces)
	deviceTypeMap := BuildDeviceTypeMap(raw.DeviceTypes)
	ifaceMap := BuildInterfaceMap(raw.Interfaces)
	moduleBayMap := BuildModuleBayMap(raw.ModuleBays)

	// 4. Devices.
	devices, deviceMap := MapDevices(raw.Devices, rackMap, locationMap, deviceTypeMap, ifacesByDevice, statusNameMap, roleNameMap)
	clog.Detail("  Transformed %d devices", len(devices))

	// 5. Cables.
	cables := MapCables(raw.Cables, deviceMap, ifaceMap, statusNameMap)
	clog.Detail("  Transformed %d cables", len(cables))

	// 6. Modules.
	modules := MapModules(raw.Modules, moduleBayMap, deviceMap, locationMap, statusNameMap, roleNameMap)
	clog.Detail("  Transformed %d modules", len(modules))

	// 7. FRUs (inventory items).
	frus := MapFrus(raw.InventoryItems, deviceMap)
	clog.Detail("  Transformed %d FRUs", len(frus))

	return &devicetypes.TransformResult{
		Locations: locations,
		Racks:     racks,
		Devices:   devices,
		Modules:   modules,
		Cables:    cables,
		Frus:      frus,
	}, nil
}

// transformLegacy is the original path: copy existing devices.
func transformLegacy(existing devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	transformed := make(map[uuid.UUID]*devicetypes.CaniDeviceType, len(existing.Devices))
	for _, device := range existing.Devices {
		d := device
		transformed[d.ID] = d
	}

	clog.Detail("  %d devices existing in current inventory", len(existing.Devices))
	clog.Detail("  %d devices Transformed (not yet Loaded)", len(transformed))

	return &devicetypes.TransformResult{
		Devices: transformed,
	}, nil
}
