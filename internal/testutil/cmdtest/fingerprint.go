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

package cmdtest

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// Fingerprint renders the meaningful content of an inventory as sorted, stable
// text.
//
// Identifiers are generated fresh on every run, so two inventories that are
// semantically identical never compare equal byte-for-byte. Fingerprint drops
// the UUIDs and keeps what the portable model actually asserts — names, slugs,
// placement, parentage by name — so the same operation run under different
// conditions can be compared.
func Fingerprint(inventory *devicetypes.Inventory) string {
	if inventory == nil {
		return "<nil inventory>"
	}

	lines := make([]string, 0, len(inventory.Devices)+len(inventory.Racks)+len(inventory.Locations))

	for _, location := range inventory.Locations {
		lines = append(lines, fmt.Sprintf("location name=%s slug=%s kind=%s",
			location.Name, location.Slug, location.LocationType))
	}

	for _, rack := range inventory.Racks {
		lines = append(lines, fmt.Sprintf("rack name=%s slug=%s uheight=%d status=%s",
			rack.Name, rack.Slug, rack.UHeight, rack.Status))
	}

	for _, device := range inventory.Devices {
		lines = append(lines, fmt.Sprintf("device name=%s slug=%s parent=%s u=%d face=%s type=%s status=%s serial=%s tags=%s",
			device.Name, device.Slug, parentName(inventory, device),
			device.RackPosition, device.Face, device.Type, device.Status,
			device.Serial, strings.Join(sortedCopy(device.Tags), "|")))
	}

	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

// parentName resolves a device's parent FK to a name so the fingerprint records
// the relationship rather than a per-run identifier.
func parentName(inventory *devicetypes.Inventory, device *devicetypes.CaniDeviceType) string {
	if rack, ok := inventory.Racks[device.Parent]; ok {
		return "rack:" + rack.Name
	}
	if parent, ok := inventory.Devices[device.Parent]; ok {
		return "device:" + parent.Name
	}
	return "none"
}

func sortedCopy(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}
