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
	"fmt"
	"strings"

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

// GetInterfacesByDevice returns all InterfaceInstance entries belonging
// to the given device (including interfaces on the device's modules).
func (inv *Inventory) GetInterfacesByDevice(deviceID uuid.UUID) []*InterfaceInstance {
	var result []*InterfaceInstance
	for _, inst := range inv.Interfaces {
		if inst.DeviceID == deviceID {
			result = append(result, inst)
		}
	}
	return result
}

// --- Referential integrity ---

// validateDeviceRefs checks device parent and children references.
func (inv *Inventory) validateDeviceRefs() []string {
	var errs []string
	for id, device := range inv.Devices {
		if device == nil {
			continue
		}
		if device.Parent != uuid.Nil && !inv.parentExists(device.Parent) {
			errs = append(errs, fmt.Sprintf(
				"device %q (%s): parent %s not found", device.Name, id, device.Parent))
		}
		for _, childID := range device.Children {
			if _, ok := inv.Devices[childID]; !ok {
				errs = append(errs, fmt.Sprintf(
					"device %q (%s): child %s not found", device.Name, id, childID))
			}
		}
	}
	return errs
}

// validateLocationRefs checks location parent and rack references.
func (inv *Inventory) validateLocationRefs() []string {
	var errs []string
	for id, loc := range inv.Locations {
		if loc == nil {
			continue
		}
		if loc.Parent != uuid.Nil {
			if _, ok := inv.Locations[loc.Parent]; !ok {
				errs = append(errs, fmt.Sprintf(
					"location %q (%s): parent %s not found", loc.Name, id, loc.Parent))
			}
		}
		for _, rackID := range loc.Racks {
			if _, ok := inv.Racks[rackID]; !ok {
				errs = append(errs, fmt.Sprintf(
					"location %q (%s): rack %s not found", loc.Name, id, rackID))
			}
		}
	}
	return errs
}

// validateRackRefs checks rack location references.
func (inv *Inventory) validateRackRefs() []string {
	var errs []string
	for id, rack := range inv.Racks {
		if rack == nil {
			continue
		}
		if rack.Location != uuid.Nil {
			if _, ok := inv.Locations[rack.Location]; !ok {
				errs = append(errs, fmt.Sprintf(
					"rack %q (%s): location %s not found", rack.Name, id, rack.Location))
			}
		}
	}
	return errs
}

// validateModuleRefs checks module parent device references.
func (inv *Inventory) validateModuleRefs() []string {
	var errs []string
	for id, mod := range inv.Modules {
		if mod == nil {
			continue
		}
		if _, ok := inv.Devices[mod.ParentDevice]; !ok {
			errs = append(errs, fmt.Sprintf(
				"module %q (%s): parent device %s not found", mod.Name, id, mod.ParentDevice))
		}
	}
	return errs
}

// validateCableRefs checks cable termination device references.
func (inv *Inventory) validateCableRefs() []string {
	var errs []string
	for id, cable := range inv.Cables {
		if cable == nil {
			continue
		}
		if cable.TerminationADevice != uuid.Nil {
			if _, ok := inv.Devices[cable.TerminationADevice]; !ok {
				errs = append(errs, fmt.Sprintf(
					"cable %q (%s): termination A device %s not found",
					cable.Label, id, cable.TerminationADevice))
			}
		}
		if cable.TerminationBDevice != uuid.Nil {
			if _, ok := inv.Devices[cable.TerminationBDevice]; !ok {
				errs = append(errs, fmt.Sprintf(
					"cable %q (%s): termination B device %s not found",
					cable.Label, id, cable.TerminationBDevice))
			}
		}
	}
	return errs
}

// validateFruRefs checks FRU device references.
func (inv *Inventory) validateFruRefs() []string {
	var errs []string
	for id, fru := range inv.Frus {
		if fru == nil {
			continue
		}
		if fru.Device != uuid.Nil {
			if _, ok := inv.Devices[fru.Device]; !ok {
				errs = append(errs, fmt.Sprintf(
					"fru %q (%s): device %s not found", fru.Name, id, fru.Device))
			}
		}
	}
	return errs
}

// Validate checks referential integrity across the entire inventory.
// It returns an error describing all broken references, or nil if valid.
func (inv *Inventory) Validate() error {
	if inv == nil {
		return fmt.Errorf("inventory is nil")
	}

	var errs []string
	errs = append(errs, inv.validateDeviceRefs()...)
	errs = append(errs, inv.validateLocationRefs()...)
	errs = append(errs, inv.validateRackRefs()...)
	errs = append(errs, inv.validateModuleRefs()...)
	errs = append(errs, inv.validateCableRefs()...)
	errs = append(errs, inv.validateFruRefs()...)

	if len(errs) > 0 {
		return fmt.Errorf("inventory validation failed:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}

// parentExists checks if a UUID exists as a device, rack, or location.
func (inv *Inventory) parentExists(id uuid.UUID) bool {
	if _, ok := inv.Devices[id]; ok {
		return true
	}
	if _, ok := inv.Modules[id]; ok {
		return true
	}
	if _, ok := inv.Racks[id]; ok {
		return true
	}
	if _, ok := inv.Locations[id]; ok {
		return true
	}
	return false
}
