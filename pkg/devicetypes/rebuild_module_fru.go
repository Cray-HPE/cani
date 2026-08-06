/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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

	"github.com/google/uuid"
)

// validateModuleRelationships verifies every module's ParentDevice
// exists in the device map.
func (inv *Inventory) validateModuleRelationships() *RelationshipResult {
	res := &RelationshipResult{}
	for id, mod := range inv.Modules {
		if mod == nil {
			continue
		}
		if mod.ParentDevice == uuid.Nil {
			res.Warnings = append(res.Warnings,
				fmt.Sprintf("module %q (%s): no parent device assigned",
					mod.Name, id))
			continue
		}
		if _, ok := inv.Devices[mod.ParentDevice]; !ok {
			res.Errors = append(res.Errors,
				fmt.Errorf("module %q (%s): parent device %s not found",
					mod.Name, id, mod.ParentDevice))
		}
	}
	return res
}

// rebuildFruRelationships clears and rebuilds Device.Frus and Module.Frus
// from each FRU's Device/Parent field, and validates references.
func (inv *Inventory) rebuildFruRelationships() *RelationshipResult {
	res := &RelationshipResult{}

	// Clear existing reverse lists.
	for _, dev := range inv.Devices {
		if dev != nil {
			dev.Frus = nil
		}
	}
	for _, mod := range inv.Modules {
		if mod != nil {
			mod.Frus = nil
		}
	}

	for id, fru := range inv.Frus {
		if fru == nil {
			continue
		}
		if fru.Device == uuid.Nil && fru.Parent == uuid.Nil {
			res.Warnings = append(res.Warnings,
				fmt.Sprintf("fru %q (%s): no device or parent assigned",
					fru.Name, id))
			continue
		}

		// Validate and link by Device field.
		if fru.Device != uuid.Nil {
			if dev, ok := inv.Devices[fru.Device]; ok {
				if !containsUUID(dev.Frus, id) {
					dev.Frus = append(dev.Frus, id)
				}
			} else if mod, ok := inv.Modules[fru.Device]; ok {
				if !containsUUID(mod.Frus, id) {
					mod.Frus = append(mod.Frus, id)
				}
			} else {
				res.Errors = append(res.Errors,
					fmt.Errorf("fru %q (%s): device %s not found in devices or modules",
						fru.Name, id, fru.Device))
			}
		}

		// Validate Parent field.
		if fru.Parent != uuid.Nil && !inv.parentExists(fru.Parent) {
			res.Errors = append(res.Errors,
				fmt.Errorf("fru %q (%s): parent %s not found",
					fru.Name, id, fru.Parent))
		}
	}
	return res
}
