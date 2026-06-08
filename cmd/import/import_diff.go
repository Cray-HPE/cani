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
package imprt

import (
	"fmt"
	"log"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// fieldDiff captures a single field's pending change during import merge.
type fieldDiff struct {
	field  string
	oldVal string
	newVal string
}

// changedDevice groups all pending field changes for one device.
type changedDevice struct {
	name  string
	diffs []fieldDiff
}

// printImportDiff prints a summary of fields that will change when incoming
// devices are merged into the existing inventory.
func printImportDiff(ctx *etlContext, incoming map[uuid.UUID]*devicetypes.CaniDeviceType) {
	var changedDevices []changedDevice
	for _, dev := range incoming {
		if dev == nil || dev.Name == "" {
			continue
		}
		existing := findExistingByName(ctx.inventory, dev.Name)
		if existing == nil {
			continue
		}
		if diffs := computeDeviceDiffs(existing, dev); len(diffs) > 0 {
			changedDevices = append(changedDevices, changedDevice{dev.Name, diffs})
		}
	}
	printChangedDevices(changedDevices)
}

// findExistingByName returns the inventory device with the given name, or nil.
func findExistingByName(inv *devicetypes.Inventory, name string) *devicetypes.CaniDeviceType {
	for _, e := range inv.Devices {
		if e != nil && e.Name == name {
			return e
		}
	}
	return nil
}

// appendStrDiff records a string field change when the incoming value is
// non-empty and differs from the existing value.
func appendStrDiff(diffs []fieldDiff, field, oldVal, newVal string) []fieldDiff {
	if newVal != "" && newVal != oldVal {
		diffs = append(diffs, fieldDiff{field, oldVal, newVal})
	}
	return diffs
}

// computeDeviceDiffs returns the set of fields that differ between an existing
// device and its incoming counterpart, preserving field order. Only non-empty
// incoming values that differ from the existing value are reported.
func computeDeviceDiffs(existing, dev *devicetypes.CaniDeviceType) []fieldDiff {
	var diffs []fieldDiff
	if dev.RackPosition != 0 && dev.RackPosition != existing.RackPosition {
		diffs = append(diffs, fieldDiff{"position", fmt.Sprintf("%d", existing.RackPosition), fmt.Sprintf("%d", dev.RackPosition)})
	}
	diffs = appendStrDiff(diffs, "status", existing.Status, dev.Status)
	diffs = appendStrDiff(diffs, "role", existing.Role, dev.Role)
	diffs = appendStrDiff(diffs, "model", existing.Model, dev.Model)
	if dev.Parent != uuid.Nil && dev.Parent != existing.Parent {
		diffs = append(diffs, fieldDiff{"parent", existing.Parent.String(), dev.Parent.String()})
	}
	diffs = appendStrDiff(diffs, "face", existing.Face, dev.Face)
	diffs = appendStrDiff(diffs, "serial", existing.Serial, dev.Serial)
	return diffs
}

// printChangedDevices logs the pending field changes grouped by device.
func printChangedDevices(changed []changedDevice) {
	if len(changed) == 0 {
		return
	}
	log.Println()
	log.Println("=== Devices with pending changes ===")
	for _, cd := range changed {
		log.Printf("  %s: %d field(s) changed:", cd.name, len(cd.diffs))
		for _, d := range cd.diffs {
			log.Printf("      %s: %s --> %s", d.field, d.oldVal, d.newVal)
		}
	}
	log.Println()
}
