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
package devicetypes

import (
	"sort"
	"strings"

	"github.com/google/uuid"
)

// FindDeviceByNameOrID resolves a device by UUID or exact name.
func (inv *Inventory) FindDeviceByNameOrID(ref string) *CaniDeviceType {
	if inv == nil || ref == "" {
		return nil
	}
	if id, err := uuid.Parse(ref); err == nil {
		if device, ok := inv.Devices[id]; ok {
			return device
		}
	}
	for _, device := range inv.Devices {
		if device != nil && device.Name == ref {
			return device
		}
	}
	return nil
}

// FindConnectableByNameOrID resolves a cable endpoint device or module.
func (inv *Inventory) FindConnectableByNameOrID(ref string) uuid.UUID {
	if device := inv.FindDeviceByNameOrID(ref); device != nil {
		return device.ID
	}
	if module := inv.FindModuleByName(ref); module != nil {
		return module.ID
	}
	return uuid.Nil
}

// DevicesBySlug returns matching devices in deterministic name order.
func (inv *Inventory) DevicesBySlug(slug string) []*CaniDeviceType {
	if inv == nil || slug == "" {
		return nil
	}
	lower := strings.ToLower(slug)
	var result []*CaniDeviceType
	for _, device := range inv.Devices {
		if device != nil && strings.ToLower(device.Slug) == lower {
			result = append(result, device)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

// OccupiedModuleBays maps occupied bay names to module IDs.
func (inv *Inventory) OccupiedModuleBays(deviceID uuid.UUID) map[string]uuid.UUID {
	result := make(map[string]uuid.UUID)
	if inv == nil {
		return result
	}
	for _, module := range inv.Modules {
		if module != nil && module.ParentDevice == deviceID && module.ModuleBayName != "" {
			result[module.ModuleBayName] = module.ID
		}
	}
	return result
}

// AvailableModuleBays returns unoccupied bays matching an optional filter.
func (inv *Inventory) AvailableModuleBays(deviceID uuid.UUID, bayFilter string) []ModuleBaySpec {
	if inv == nil || inv.Devices[deviceID] == nil {
		return nil
	}
	occupied := inv.OccupiedModuleBays(deviceID)
	filter := strings.ToLower(bayFilter)
	var result []ModuleBaySpec
	for _, bay := range inv.Devices[deviceID].ModuleBays {
		if _, taken := occupied[bay.Name]; taken {
			continue
		}
		if filter == "" || strings.Contains(strings.ToLower(bay.Name), filter) {
			result = append(result, bay)
		}
	}
	return result
}

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
	if _, ok := inv.Frus[id]; ok {
		return true
	}
	return false
}
