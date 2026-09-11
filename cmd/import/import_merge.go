/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024, 2026 Hewlett Packard Enterprise Development LP
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
	"log"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func prepareUnclassifiedDevices(devices map[uuid.UUID]*devicetypes.CaniDeviceType) {
	if config.Cfg == nil || !config.Cfg.Strict || len(devices) == 0 {
		return
	}
	skipped := devicetypes.FindUnclassifiedDevices(&devicetypes.Inventory{Devices: devices})
	if stepFlag {
		classifySkippedDevices(devices, skipped)
		return
	}
	warnUnclassifiedDevices(skipped)
}

// classifySkippedDevices prompts the user to classify each unclassified device
// in step mode before the model-owned merge transaction begins.
func classifySkippedDevices(
	devices map[uuid.UUID]*devicetypes.CaniDeviceType,
	skipped []devicetypes.UnclassifiedDevice,
) {
	classifyOpts := devicetypes.ClassifyOptions{NoColor: noColorFlag}
	classified := 0
	for _, ud := range skipped {
		if classifyOneDevice(devices, ud, classifyOpts) {
			classified++
		}
	}
	if classified > 0 {
		log.Printf("  Classified %d of %d unclassified devices", classified, len(skipped))
	}
}

// classifyOneDevice prompts for a single device's type and applies it.
func classifyOneDevice(
	devices map[uuid.UUID]*devicetypes.CaniDeviceType,
	ud devicetypes.UnclassifiedDevice,
	opts devicetypes.ClassifyOptions,
) bool {
	slug, err := devicetypes.PromptForDeviceType(ud, opts)
	if err != nil {
		log.Printf("  ! %s: classification error: %v", ud.Name, err)
		return false
	}
	device := devices[ud.ID]
	if device == nil {
		return false
	}
	return applyClassification(device, ud, slug)
}

// applyClassification applies the chosen slug to a device, logging the outcome.
// It returns true only when a non-empty slug was applied without error.
func applyClassification(device *devicetypes.CaniDeviceType, ud devicetypes.UnclassifiedDevice, slug string) bool {
	if slug == "" {
		log.Printf("  - %s: skipped (no type selected)", ud.Name)
		return false
	}
	if err := devicetypes.ApplyDeviceType(device, slug); err != nil {
		log.Printf("  ! %s: failed to apply type %q: %v", ud.Name, slug, err)
		return false
	}
	return true
}

// warnUnclassifiedDevices warns about rejected devices (non-interactive mode)
// and still merges them so modules/FRUs can reference them as parents.
func warnUnclassifiedDevices(skipped []devicetypes.UnclassifiedDevice) {
	log.Printf("")
	log.Printf("  ⚠ %d devices are unclassified (no device type slug or model):", len(skipped))
	for _, ud := range skipped {
		log.Printf("    - %s", ud.Name)
	}
	log.Printf("")
	log.Printf("  To assign types interactively, run:")
	log.Printf("    cani alpha classify")
	log.Printf("  Or re-import with --step to classify inline.")
	log.Printf("  To allow unclassified devices, use --strict=false")
	log.Printf("")
}
