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
	"fmt"
	"strings"
)

// GenerateDeviceNames assigns a unique cani-prefixed name to every
// device-category device that has none, built from the best available
// identifier: serial number, slug, model, or the first eight characters of
// the cani UUID. Providers whose targets require a name call it before export.
func (inv *Inventory) GenerateDeviceNames() {
	for _, device := range inv.Devices {
		if device == nil || device.Name != "" {
			continue
		}
		category := ClassifyForNautobot(string(device.Type))
		if category != CategoryDevice {
			continue
		}

		var base string
		switch {
		case device.Serial != "":
			base = device.Serial
		case device.Slug != "":
			base = device.Slug
		case device.Model != "":
			base = strings.ToLower(strings.ReplaceAll(device.Model, " ", "-"))
		default:
			base = device.ID.String()[:8]
		}
		device.Name = "cani-" + base
	}
}

// DisambiguateDeviceNames makes duplicate device-category names unique by
// appending what an operator can see on the hardware: the serial, the rack
// and U position, or the first eight characters of the cani UUID. Nautobot
// enforces name uniqueness per location and tenant, so duplicates must be
// resolved before export. TransformResult.EnsureUniqueDeviceNames is the
// import-time counterpart: it numbers an incoming batch against the names the
// inventory already holds.
func (inv *Inventory) DisambiguateDeviceNames() {
	for _, devices := range inv.deviceCategoryByName() {
		if len(devices) <= 1 {
			continue
		}
		for _, device := range devices {
			device.Name += inv.deviceNameSuffix(device)
		}
	}
}

// deviceCategoryByName groups the named device-category devices by name.
func (inv *Inventory) deviceCategoryByName() map[string][]*CaniDeviceType {
	byName := make(map[string][]*CaniDeviceType)
	for _, device := range inv.Devices {
		if device == nil || device.Name == "" || ClassifyForNautobot(string(device.Type)) != CategoryDevice {
			continue
		}
		byName[device.Name] = append(byName[device.Name], device)
	}
	return byName
}

// deviceNameSuffix labels a duplicate by its serial, its rack and U position,
// or the first eight characters of its UUID.
func (inv *Inventory) deviceNameSuffix(device *CaniDeviceType) string {
	switch {
	case device.Serial != "":
		return " (" + device.Serial + ")"
	case device.RackPosition > 0:
		rackLabel := ""
		if rack, ok := inv.Racks[device.GetRackID(inv)]; ok && rack != nil {
			rackLabel = rack.Name + " "
		}
		return fmt.Sprintf(" (%sU%d)", rackLabel, device.RackPosition)
	default:
		return fmt.Sprintf(" (%s)", device.ID.String()[:8])
	}
}
