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
package show

import (
	"fmt"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// findLocationByNameOrUUID looks up a location by UUID string or exact name (case-insensitive).
func findLocationByNameOrUUID(arg string, inv *devicetypes.Inventory) (*devicetypes.CaniLocationType, error) {
	if id, err := uuid.Parse(arg); err == nil {
		if loc, ok := inv.Locations[id]; ok {
			return loc, nil
		}
		return nil, fmt.Errorf("location with UUID %q not found", arg)
	}
	for _, loc := range inv.Locations {
		if strings.EqualFold(loc.Name, arg) {
			return loc, nil
		}
	}
	return nil, fmt.Errorf("location %q not found", arg)
}

// findRackByNameOrUUID looks up a rack by UUID string or exact name (case-insensitive).
func findRackByNameOrUUID(arg string, inv *devicetypes.Inventory) (*devicetypes.CaniRackType, error) {
	if id, err := uuid.Parse(arg); err == nil {
		if r, ok := inv.Racks[id]; ok {
			return r, nil
		}
		return nil, fmt.Errorf("rack with UUID %q not found", arg)
	}
	for _, r := range inv.Racks {
		if strings.EqualFold(r.Name, arg) {
			return r, nil
		}
	}
	return nil, fmt.Errorf("rack %q not found", arg)
}

// findDeviceByNameOrUUID looks up a device by UUID string or exact name (case-insensitive).
func findDeviceByNameOrUUID(arg string, inv *devicetypes.Inventory) (*devicetypes.CaniDeviceType, error) {
	if id, err := uuid.Parse(arg); err == nil {
		if d, ok := inv.Devices[id]; ok {
			return d, nil
		}
		return nil, fmt.Errorf("device with UUID %q not found", arg)
	}
	for _, d := range inv.Devices {
		if strings.EqualFold(d.Name, arg) {
			return d, nil
		}
	}
	return nil, fmt.Errorf("device %q not found", arg)
}

// findModuleByNameOrUUID looks up a module by UUID string or exact name (case-insensitive).
func findModuleByNameOrUUID(arg string, inv *devicetypes.Inventory) (*devicetypes.CaniModuleType, error) {
	if id, err := uuid.Parse(arg); err == nil {
		if m, ok := inv.Modules[id]; ok {
			return m, nil
		}
		return nil, fmt.Errorf("module with UUID %q not found", arg)
	}
	for _, m := range inv.Modules {
		if strings.EqualFold(m.Name, arg) {
			return m, nil
		}
	}
	return nil, fmt.Errorf("module %q not found", arg)
}

// findCableByLabelOrUUID looks up a cable by UUID string or exact label (case-insensitive).
func findCableByLabelOrUUID(arg string, inv *devicetypes.Inventory) (*devicetypes.CaniCableType, error) {
	if id, err := uuid.Parse(arg); err == nil {
		if c, ok := inv.Cables[id]; ok {
			return c, nil
		}
		return nil, fmt.Errorf("cable with UUID %q not found", arg)
	}
	for _, c := range inv.Cables {
		if strings.EqualFold(c.Label, arg) {
			return c, nil
		}
	}
	return nil, fmt.Errorf("cable %q not found", arg)
}

// findFruByNameOrUUID looks up a FRU by UUID string or exact name (case-insensitive).
func findFruByNameOrUUID(arg string, inv *devicetypes.Inventory) (*devicetypes.CaniFruType, error) {
	if id, err := uuid.Parse(arg); err == nil {
		if f, ok := inv.Frus[id]; ok {
			return f, nil
		}
		return nil, fmt.Errorf("fru with UUID %q not found", arg)
	}
	for _, f := range inv.Frus {
		if strings.EqualFold(f.Name, arg) {
			return f, nil
		}
	}
	return nil, fmt.Errorf("fru %q not found", arg)
}

// findInterfaceByNameOrUUID looks up an interface by UUID string or exact name (case-insensitive).
func findInterfaceByNameOrUUID(arg string, inv *devicetypes.Inventory) (*devicetypes.InterfaceInstance, error) {
	if id, err := uuid.Parse(arg); err == nil {
		if iface, ok := inv.Interfaces[id]; ok {
			return iface, nil
		}
		return nil, fmt.Errorf("interface with UUID %q not found", arg)
	}
	for _, iface := range inv.Interfaces {
		if strings.EqualFold(iface.Name, arg) {
			return iface, nil
		}
	}
	return nil, fmt.Errorf("interface %q not found", arg)
}
