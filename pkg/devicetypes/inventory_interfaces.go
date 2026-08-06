/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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

import (
	"strings"

	"github.com/google/uuid"
)

// findInterfaceOnDevice returns the InterfaceSpec matching ifaceID from the device, or nil.
func findInterfaceOnDevice(device *CaniDeviceType, ifaceID uuid.UUID) *InterfaceSpec {
	for i := range device.Interfaces {
		if device.Interfaces[i].ID == ifaceID {
			return &device.Interfaces[i]
		}
	}
	return nil
}

// findInterfaceInModules searches all modules for an InterfaceSpec matching ifaceID.
// Returns the spec and the module's parent device, or nil if not found.
func (inv *Inventory) findInterfaceInModules(ifaceID uuid.UUID) (*InterfaceSpec, *CaniDeviceType) {
	for _, mod := range inv.Modules {
		if mod == nil {
			continue
		}
		for i := range mod.Interfaces {
			if mod.Interfaces[i].ID == ifaceID {
				return &mod.Interfaces[i], inv.Devices[mod.ParentDevice]
			}
		}
	}
	return nil, nil
}

// GetInterfaceByID finds an interface by UUID using the Interfaces index.
// Returns the interface spec and the owning device (nil for module-owned interfaces).
func (inv *Inventory) GetInterfaceByID(ifaceID uuid.UUID) (*InterfaceSpec, *CaniDeviceType) {
	if inv == nil {
		return nil, nil
	}

	inst, ok := inv.Interfaces[ifaceID]
	if !ok {
		return nil, nil
	}

	if device, exists := inv.Devices[inst.DeviceID]; exists {
		if spec := findInterfaceOnDevice(device, ifaceID); spec != nil {
			return spec, device
		}
	}

	return inv.findInterfaceInModules(ifaceID)
}

// GetInterfacesByDevice returns all CaniInterface entries belonging
// to the given device (including interfaces on the device's modules).
func (inv *Inventory) GetInterfacesByDevice(deviceID uuid.UUID) []*CaniInterface {
	var result []*CaniInterface
	for _, inst := range inv.Interfaces {
		if inst.DeviceID == deviceID {
			result = append(result, inst)
		}
	}
	return result
}

// --- Module bay occupancy ---

// OccupiedModuleBays returns a map of bay-name → module-ID for all modules
// installed in the given device. This mirrors how Nautobot tracks module
// bay occupancy: each Module references a ModuleBay (by name) on a
// parent Device, and occupancy is implicit from the Module records.
func (inv *Inventory) OccupiedModuleBays(deviceID uuid.UUID) map[string]uuid.UUID {
	result := make(map[string]uuid.UUID)
	if inv == nil {
		return result
	}
	for _, mod := range inv.Modules {
		if mod != nil && mod.ParentDevice == deviceID && mod.ModuleBayName != "" {
			result[mod.ModuleBayName] = mod.ID
		}
	}
	return result
}

// AvailableModuleBays returns the device's module-bay specs that are not
// yet occupied by a module. If bayFilter is non-empty, only bays whose
// name contains the filter string (case-insensitive) are returned.
func (inv *Inventory) AvailableModuleBays(deviceID uuid.UUID, bayFilter string) []ModuleBaySpec {
	if inv == nil {
		return nil
	}
	dev, ok := inv.Devices[deviceID]
	if !ok || dev == nil {
		return nil
	}
	occupied := inv.OccupiedModuleBays(deviceID)
	filter := strings.ToLower(bayFilter)

	var result []ModuleBaySpec
	for _, bay := range dev.ModuleBays {
		if _, taken := occupied[bay.Name]; taken {
			continue
		}
		if filter != "" && !strings.Contains(strings.ToLower(bay.Name), filter) {
			continue
		}
		result = append(result, bay)
	}
	return result
}
