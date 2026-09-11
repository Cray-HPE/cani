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

	"github.com/google/uuid"
)

func (inv *Inventory) validateDeviceRefs() []string {
	var errs []string
	for id, device := range inv.Devices {
		if device == nil {
			continue
		}
		if device.Parent != uuid.Nil && !inv.parentExists(device.Parent) {
			errs = append(errs, fmt.Sprintf("device %q (%s): parent %s not found", device.Name, id, device.Parent))
		}
		for _, childID := range device.Children {
			if _, ok := inv.Devices[childID]; !ok {
				errs = append(errs, fmt.Sprintf("device %q (%s): child %s not found", device.Name, id, childID))
			}
		}
	}
	return errs
}

func (inv *Inventory) validateLocationRefs() []string {
	var errs []string
	for id, location := range inv.Locations {
		if location == nil {
			continue
		}
		if location.Parent != uuid.Nil {
			if _, ok := inv.Locations[location.Parent]; !ok {
				errs = append(errs, fmt.Sprintf("location %q (%s): parent %s not found", location.Name, id, location.Parent))
			}
		}
		for _, rackID := range location.Racks {
			if _, ok := inv.Racks[rackID]; !ok {
				errs = append(errs, fmt.Sprintf("location %q (%s): rack %s not found", location.Name, id, rackID))
			}
		}
	}
	return errs
}

func (inv *Inventory) validateRackRefs() []string {
	var errs []string
	for id, rack := range inv.Racks {
		if rack != nil && rack.Location != uuid.Nil {
			if _, ok := inv.Locations[rack.Location]; !ok {
				errs = append(errs, fmt.Sprintf("rack %q (%s): location %s not found", rack.Name, id, rack.Location))
			}
		}
	}
	return errs
}

func (inv *Inventory) validateModuleRefs() []string {
	var errs []string
	for id, module := range inv.Modules {
		if module == nil {
			continue
		}
		if _, ok := inv.Devices[module.ParentDevice]; !ok {
			errs = append(errs, fmt.Sprintf("module %q (%s): parent device %s not found", module.Name, id, module.ParentDevice))
		}
	}
	return errs
}

func (inv *Inventory) validateCableRefs() []string {
	var errs []string
	for id, cable := range inv.Cables {
		if cable == nil {
			continue
		}
		if cable.TerminationADevice != uuid.Nil {
			if _, ok := inv.Devices[cable.TerminationADevice]; !ok {
				errs = append(errs, fmt.Sprintf("cable %q (%s): termination A device %s not found", cable.Label, id, cable.TerminationADevice))
			}
		}
		if cable.TerminationBDevice != uuid.Nil {
			if _, ok := inv.Devices[cable.TerminationBDevice]; !ok {
				errs = append(errs, fmt.Sprintf("cable %q (%s): termination B device %s not found", cable.Label, id, cable.TerminationBDevice))
			}
		}
	}
	return errs
}

func (inv *Inventory) validateFruRefs() []string {
	var errs []string
	for id, fru := range inv.Frus {
		if fru != nil && fru.Device != uuid.Nil {
			if _, ok := inv.Devices[fru.Device]; !ok {
				errs = append(errs, fmt.Sprintf("fru %q (%s): device %s not found", fru.Name, id, fru.Device))
			}
		}
	}
	return errs
}

// Validate checks referential integrity across the entire inventory.
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
