/*
 *
 *  MIT License
 *
 *  (C) Copyright 2024-2026 Hewlett Packard Enterprise Development LP
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
package visual

import (
	"sort"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// ResolveLocationChildren resolves a slice of location UUIDs into sorted location pointers.
func ResolveLocationChildren(ids []uuid.UUID, inv *devicetypes.Inventory) []*devicetypes.CaniLocationType {
	if inv == nil || inv.Locations == nil {
		return nil
	}
	result := make([]*devicetypes.CaniLocationType, 0, len(ids))
	for _, id := range ids {
		if loc, ok := inv.Locations[id]; ok {
			result = append(result, loc)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// ResolveRacks resolves a slice of rack UUIDs into sorted rack pointers.
func ResolveRacks(ids []uuid.UUID, inv *devicetypes.Inventory) []*devicetypes.CaniRackType {
	if inv == nil || inv.Racks == nil {
		return nil
	}
	result := make([]*devicetypes.CaniRackType, 0, len(ids))
	for _, id := range ids {
		if r, ok := inv.Racks[id]; ok {
			result = append(result, r)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// ResolveDevices resolves a slice of device UUIDs into sorted device pointers.
func ResolveDevices(ids []uuid.UUID, inv *devicetypes.Inventory) []*devicetypes.CaniDeviceType {
	if inv == nil || inv.Devices == nil {
		return nil
	}
	result := make([]*devicetypes.CaniDeviceType, 0, len(ids))
	for _, id := range ids {
		if d, ok := inv.Devices[id]; ok {
			result = append(result, d)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		pi, pj := result[i].RackPosition, result[j].RackPosition
		if pi != 0 && pj != 0 {
			return pi > pj
		}
		if pi != pj {
			return pi != 0
		}
		return result[i].Name < result[j].Name
	})
	return result
}

// ResolveDeviceChildren resolves child device UUIDs into sorted device pointers.
func ResolveDeviceChildren(ids []uuid.UUID, inv *devicetypes.Inventory) []*devicetypes.CaniDeviceType {
	return ResolveDevices(ids, inv)
}

// FindModulesForDevice returns all modules whose ParentDevice matches deviceID.
func FindModulesForDevice(deviceID uuid.UUID, inv *devicetypes.Inventory) []*devicetypes.CaniModuleType {
	if inv == nil || inv.Modules == nil {
		return nil
	}
	var result []*devicetypes.CaniModuleType
	for _, mod := range inv.Modules {
		if mod.ParentDevice == deviceID {
			result = append(result, mod)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// FindFrusForDevice returns all FRUs whose Device matches deviceID.
func FindFrusForDevice(deviceID uuid.UUID, inv *devicetypes.Inventory) []*devicetypes.CaniFruType {
	if inv == nil || inv.Frus == nil {
		return nil
	}
	var result []*devicetypes.CaniFruType
	for _, fru := range inv.Frus {
		if fru.Device == deviceID {
			result = append(result, fru)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// FindCablesForDevice returns all cables where TerminationADevice matches deviceID.
func FindCablesForDevice(deviceID uuid.UUID, inv *devicetypes.Inventory) []*devicetypes.CaniCableType {
	if inv == nil || inv.Cables == nil {
		return nil
	}
	var result []*devicetypes.CaniCableType
	for _, c := range inv.Cables {
		if c.TerminationADevice == deviceID {
			result = append(result, c)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Label < result[j].Label
	})
	return result
}

// FindCableForInterface returns the cable connected to iface on deviceID, or nil.
func FindCableForInterface(iface devicetypes.InterfaceSpec, deviceID uuid.UUID, inv *devicetypes.Inventory) *devicetypes.CaniCableType {
	if inv == nil || inv.Cables == nil {
		return nil
	}
	if iface.ConnectedCable != nil {
		if c, ok := inv.Cables[*iface.ConnectedCable]; ok {
			return c
		}
	}
	for _, c := range inv.Cables {
		if c.TerminationADevice == deviceID && c.TerminationAPort == iface.Name {
			return c
		}
		if c.TerminationBDevice == deviceID && c.TerminationBPort == iface.Name {
			return c
		}
	}
	return nil
}

// CableMatchesAnyInterface returns true if cable c is connected to one of
// dev's interfaces (meaning it is already shown inline on an interface line).
func CableMatchesAnyInterface(c *devicetypes.CaniCableType, dev *devicetypes.CaniDeviceType, inv *devicetypes.Inventory) bool {
	port := c.TerminationAPort
	if c.TerminationADevice != dev.ID {
		port = c.TerminationBPort
	}
	for _, iface := range dev.Interfaces {
		if iface.Name == port {
			return true
		}
	}
	return false
}

// CableAttachedToAnyDeviceInterface returns true if a cable's termination port
// matches an interface on either termination device.
func CableAttachedToAnyDeviceInterface(c *devicetypes.CaniCableType, inv *devicetypes.Inventory) bool {
	if inv == nil || inv.Devices == nil {
		return false
	}
	if dev, ok := inv.Devices[c.TerminationADevice]; ok {
		for _, iface := range dev.Interfaces {
			if iface.Name == c.TerminationAPort {
				return true
			}
		}
	}
	if dev, ok := inv.Devices[c.TerminationBDevice]; ok {
		for _, iface := range dev.Interfaces {
			if iface.Name == c.TerminationBPort {
				return true
			}
		}
	}
	return false
}

// CountLocationDescendants recursively counts racks and devices under a location.
func CountLocationDescendants(loc *devicetypes.CaniLocationType, inv *devicetypes.Inventory) (racks int, devices int) {
	if inv == nil {
		return 0, 0
	}
	racks += len(loc.Racks)
	for _, rackID := range loc.Racks {
		if rack, ok := inv.Racks[rackID]; ok {
			devices += len(rack.Devices)
		}
	}
	for _, childID := range loc.Children {
		if child, ok := inv.Locations[childID]; ok {
			r, d := CountLocationDescendants(child, inv)
			racks += r
			devices += d
		}
	}
	return racks, devices
}

// RemoteTermination returns the formatted device:port for the far end of a cable
// relative to localDevice.
func RemoteTermination(c *devicetypes.CaniCableType, localDevice uuid.UUID, inv *devicetypes.Inventory) string {
	if c.TerminationADevice == localDevice {
		return FormatTermination(c.TerminationBDevice, c.TerminationBPort, inv)
	}
	return FormatTermination(c.TerminationADevice, c.TerminationAPort, inv)
}
