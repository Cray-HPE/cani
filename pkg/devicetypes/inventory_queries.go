/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023, 2026 Hewlett Packard Enterprise Development LP
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
	"sort"

	"github.com/google/uuid"
)

// --- Location lookups ---

// FindLocationByName returns the first location matching the given name, or nil.
func (inv *Inventory) FindLocationByName(name string) *CaniLocationType {
	if inv == nil {
		return nil
	}
	for _, loc := range inv.Locations {
		if loc != nil && loc.Name == name {
			return loc
		}
	}
	return nil
}

// LocationExists returns true if a location with the given name exists.
func (inv *Inventory) LocationExists(name string) bool {
	return inv.FindLocationByName(name) != nil
}

// --- Location helpers ---

// FindLocationByNameOrID tries to parse ref as a UUID; if that fails it
// falls back to a name lookup. Returns nil if nothing matches.
func (inv *Inventory) FindLocationByNameOrID(ref string) *CaniLocationType {
	if inv == nil || ref == "" {
		return nil
	}
	if id, err := uuid.Parse(ref); err == nil {
		if loc, ok := inv.Locations[id]; ok {
			return loc
		}
	}
	return inv.FindLocationByName(ref)
}

// --- Rack lookups ---

// FindRackByName returns the first rack matching the given name, or nil.
func (inv *Inventory) FindRackByName(name string) *CaniRackType {
	if inv == nil {
		return nil
	}
	for _, rack := range inv.Racks {
		if rack != nil && rack.Name == name {
			return rack
		}
	}
	return nil
}

// RackExists returns true if a rack with the given name exists.
func (inv *Inventory) RackExists(name string) bool {
	return inv.FindRackByName(name) != nil
}

// RacksByLocation returns all racks at the given location, sorted
// alphabetically by name for deterministic ordering.
func (inv *Inventory) RacksByLocation(locationID uuid.UUID) []*CaniRackType {
	if inv == nil {
		return nil
	}
	var result []*CaniRackType
	for _, rack := range inv.Racks {
		if rack != nil && rack.Location == locationID {
			result = append(result, rack)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// AllRacks returns every rack in the inventory, sorted alphabetically
// by name for deterministic ordering.
func (inv *Inventory) AllRacks() []*CaniRackType {
	if inv == nil {
		return nil
	}
	result := make([]*CaniRackType, 0, len(inv.Racks))
	for _, rack := range inv.Racks {
		if rack != nil {
			result = append(result, rack)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// --- Module lookups ---

// FindModuleByName returns the first module matching the given name, or nil.
func (inv *Inventory) FindModuleByName(name string) *CaniModuleType {
	if inv == nil {
		return nil
	}
	for _, mod := range inv.Modules {
		if mod != nil && mod.Name == name {
			return mod
		}
	}
	return nil
}

// ModuleExists returns true if a module with the given name exists.
func (inv *Inventory) ModuleExists(name string) bool {
	return inv.FindModuleByName(name) != nil
}

// --- FRU lookups ---

// FindFruByName returns the first FRU matching the given name, or nil.
func (inv *Inventory) FindFruByName(name string) *CaniFruType {
	if inv == nil {
		return nil
	}
	for _, fru := range inv.Frus {
		if fru != nil && fru.Name == name {
			return fru
		}
	}
	return nil
}

// FruExists returns true if a FRU with the given name exists.
func (inv *Inventory) FruExists(name string) bool {
	return inv.FindFruByName(name) != nil
}

// --- Cable lookups ---

// FindCableByLabel returns the first cable matching the given label, or nil.
func (inv *Inventory) FindCableByLabel(label string) *CaniCableType {
	if inv == nil {
		return nil
	}
	for _, cable := range inv.Cables {
		if cable != nil && cable.Label == label {
			return cable
		}
	}
	return nil
}

// --- Cross-reference queries ---

// GetDevicesInRack returns all devices whose Parent matches the given rack UUID.
func (inv *Inventory) GetDevicesInRack(rackID uuid.UUID) []*CaniDeviceType {
	if inv == nil {
		return nil
	}
	var result []*CaniDeviceType
	for _, device := range inv.Devices {
		if device != nil && device.Parent == rackID {
			result = append(result, device)
		}
	}
	return result
}

// GetCablesForDevice returns all cables where either termination references the device.
func (inv *Inventory) GetCablesForDevice(deviceID uuid.UUID) []*CaniCableType {
	if inv == nil {
		return nil
	}
	var result []*CaniCableType
	for _, cable := range inv.Cables {
		if cable == nil {
			continue
		}
		if cable.TerminationADevice == deviceID || cable.TerminationBDevice == deviceID {
			result = append(result, cable)
		}
	}
	return result
}

// GetModulesForDevice returns all modules whose ParentDevice matches the device UUID.
func (inv *Inventory) GetModulesForDevice(deviceID uuid.UUID) []*CaniModuleType {
	if inv == nil {
		return nil
	}
	var result []*CaniModuleType
	for _, mod := range inv.Modules {
		if mod != nil && mod.ParentDevice == deviceID {
			result = append(result, mod)
		}
	}
	return result
}

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
