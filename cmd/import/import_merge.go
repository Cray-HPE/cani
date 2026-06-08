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

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/visual"
	"github.com/google/uuid"
)

// mergeMetadata adds roles, statuses, and tags from the transform result
// into the inventory metadata catalog. Duplicates are silently ignored.
func mergeMetadata(ctx *etlContext, meta *devicetypes.InventoryMetadata) {
	if meta == nil {
		return
	}
	for _, role := range meta.Roles {
		_ = ctx.inventory.AddMetadata("roles", role)
	}
	for _, status := range meta.Statuses {
		_ = ctx.inventory.AddMetadata("statuses", status)
	}
	for _, tag := range meta.Tags {
		_ = ctx.inventory.AddMetadata("tags", tag)
	}
}

// remapDeviceParents rewrites device Parent fields using the UUID remap
// maps returned by MergeLocations and MergeRacks. This ensures devices
// point to existing inventory UUIDs rather than ephemeral transform UUIDs.
func remapDeviceParents(
	devices map[uuid.UUID]*devicetypes.CaniDeviceType,
	locationRemap, rackRemap map[uuid.UUID]uuid.UUID,
) {
	for _, dev := range devices {
		if dev == nil || dev.Parent == uuid.Nil {
			continue
		}
		if mapped, ok := rackRemap[dev.Parent]; ok {
			dev.Parent = mapped
		} else if mapped, ok := locationRemap[dev.Parent]; ok {
			dev.Parent = mapped
		}
	}
}

// mergeLocations adds transformed locations to the inventory.
func mergeLocations(ctx *etlContext, locations map[uuid.UUID]*devicetypes.CaniLocationType) map[uuid.UUID]uuid.UUID {
	if len(locations) == 0 {
		return nil
	}
	if ctx.debug {
		visual.PrintCaniOperation(fmt.Sprintf("Merging %d locations into inventory", len(locations)), ctx.opts)
	}
	return ctx.inventory.MergeLocations(locations)
}

// mergeRacks adds transformed racks to the inventory.
func mergeRacks(ctx *etlContext, racks map[uuid.UUID]*devicetypes.CaniRackType) map[uuid.UUID]uuid.UUID {
	if len(racks) == 0 {
		return nil
	}
	if ctx.debug {
		visual.PrintCaniOperation(fmt.Sprintf("Merging %d racks into inventory", len(racks)), ctx.opts)
	}
	return ctx.inventory.MergeRacks(racks)
}

// mergeDevices adds transformed devices to the inventory.
// In strict mode, unclassified devices (no slug/model) are rejected.
// In step mode, the user is prompted to interactively classify them.
func mergeDevices(ctx *etlContext, devices map[uuid.UUID]*devicetypes.CaniDeviceType) {
	if len(devices) == 0 {
		return
	}
	if ctx.debug {
		visual.PrintCaniOperation(fmt.Sprintf("Merging %d transformed devices into inventory", len(devices)), ctx.opts)
	}

	skipped := ctx.inventory.MergeDevicesStrict(devices, config.Cfg.Strict)
	if len(skipped) == 0 {
		return
	}

	if stepFlag {
		classifySkippedDevices(ctx, devices, skipped)
		return
	}

	warnUnclassifiedDevices(ctx, devices, skipped)
}

// classifySkippedDevices prompts the user to classify each unclassified device
// (step mode) and always merges it so modules/FRUs can reference it as a parent.
func classifySkippedDevices(
	ctx *etlContext,
	devices map[uuid.UUID]*devicetypes.CaniDeviceType,
	skipped []devicetypes.UnclassifiedDevice,
) {
	classifyOpts := devicetypes.ClassifyOptions{NoColor: noColorFlag}
	classified := 0
	for _, ud := range skipped {
		if classifyOneDevice(ctx, devices, ud, classifyOpts) {
			classified++
		}
	}
	if classified > 0 {
		log.Printf("  Classified %d of %d unclassified devices", classified, len(skipped))
	}
}

// classifyOneDevice prompts for a single device's type, applies it, and merges
// the device. It returns true when a type was successfully applied.
func classifyOneDevice(
	ctx *etlContext,
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
	classified := applyClassification(device, ud, slug)
	// Always merge the device so modules/FRUs can reference it as a parent.
	ctx.inventory.MergeDevices(map[uuid.UUID]*devicetypes.CaniDeviceType{ud.ID: device})
	return classified
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
func warnUnclassifiedDevices(
	ctx *etlContext,
	devices map[uuid.UUID]*devicetypes.CaniDeviceType,
	skipped []devicetypes.UnclassifiedDevice,
) {
	log.Printf("")
	log.Printf("  ⚠ %d devices rejected (no device type slug or model):", len(skipped))
	for _, ud := range skipped {
		log.Printf("    - %s", ud.Name)
	}
	log.Printf("")
	log.Printf("  To assign types interactively, run:")
	log.Printf("    cani alpha classify")
	log.Printf("  Or re-import with --step to classify inline.")
	log.Printf("  To allow unclassified devices, use --strict=false")
	log.Printf("")

	for _, ud := range skipped {
		device := devices[ud.ID]
		if device == nil {
			continue
		}
		ctx.inventory.MergeDevices(map[uuid.UUID]*devicetypes.CaniDeviceType{ud.ID: device})
	}
}

// mergeModules adds transformed modules to the inventory.
func mergeModules(ctx *etlContext, modules map[uuid.UUID]*devicetypes.CaniModuleType) {
	if len(modules) == 0 {
		return
	}
	if ctx.debug {
		visual.PrintCaniOperation(fmt.Sprintf("Merging %d modules into inventory", len(modules)), ctx.opts)
	}
	ctx.inventory.MergeModules(modules)
}

// mergeCables adds transformed cables to the inventory.
func mergeCables(ctx *etlContext, cables map[uuid.UUID]*devicetypes.CaniCableType) {
	if len(cables) == 0 {
		return
	}
	if ctx.debug {
		visual.PrintCaniOperation(fmt.Sprintf("Merging %d cables into inventory", len(cables)), ctx.opts)
	}
	ctx.inventory.MergeCables(cables)
}

// mergeFrus adds transformed FRUs to the inventory.
func mergeFrus(ctx *etlContext, frus map[uuid.UUID]*devicetypes.CaniFruType) {
	if len(frus) == 0 {
		return
	}
	if ctx.debug {
		visual.PrintCaniOperation(fmt.Sprintf("Merging %d FRUs into inventory", len(frus)), ctx.opts)
	}
	ctx.inventory.MergeFrus(frus)
}
