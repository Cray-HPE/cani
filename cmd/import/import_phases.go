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
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/visual"
	"github.com/google/uuid"
)

// ETL phase display labels used in headers, prompts, and completion messages.
const (
	phaseNameExtract   = "EXTRACT"
	phaseNameTransform = "TRANSFORM"
	phaseNameLoad      = "LOAD"
)

// runExtractPhase executes the Extract phase of the ETL pipeline.
// It calls provider.Import() which populates ctx.inventory with raw data from the source.
func runExtractPhase(ctx *etlContext) error {
	if ctx.debug {
		visual.PrintPhaseHeader(phaseNameExtract, ctx.opts)
		visual.PrintCaniOperation("Starting extract phase", ctx.opts)
	}

	if stepFlag {
		if err := promptPhaseStart(phaseNameExtract, ctx.opts); err != nil {
			return fmt.Errorf("extract phase: %w", err)
		}
	}

	importer, ok := ctx.provider.(provider.Importer)
	if !ok {
		if ctx.debug {
			visual.PrintPhaseComplete(phaseNameExtract, ctx.opts)
		}
		return nil
	}

	if ctx.debug {
		visual.PrintProviderOperation(fmt.Sprintf("Importing data from provider %s", ctx.provider.Slug()), ctx.opts)
	}

	if err := importer.Import(ctx.cmd, ctx.args, ctx.inventory); err != nil {
		return fmt.Errorf("extract phase: failed to import from %s: %w", ctx.provider.Slug(), err)
	}

	if ctx.debug {
		visual.PrintPhaseComplete(phaseNameExtract, ctx.opts)
	}
	return nil
}

// runTransformPhase executes the Transform phase of the ETL pipeline.
// It calls provider.Transform() and merges the result into ctx.inventory.
func runTransformPhase(ctx *etlContext) error {
	if ctx.debug {
		visual.PrintPhaseHeader(phaseNameTransform, ctx.opts)
		visual.PrintCaniOperation("Starting transform phase", ctx.opts)
	}

	if stepFlag {
		if err := promptPhaseStart(phaseNameTransform, ctx.opts); err != nil {
			return fmt.Errorf("transform phase: %w", err)
		}
	}

	if ctx.debug {
		visual.PrintProviderOperation(fmt.Sprintf("Transforming data from provider %s", ctx.provider.Slug()), ctx.opts)
	}

	result, err := ctx.provider.Transform(*ctx.inventory)
	if err != nil {
		return fmt.Errorf("transform phase: failed to transform data from %s: %w", ctx.provider.Slug(), err)
	}

	displayTransformSummary(ctx, result)
	mergeTransformResult(ctx, result)

	if ctx.debug {
		visual.PrintPhaseComplete(phaseNameTransform, ctx.opts)
	}
	return nil
}

// displayTransformSummary shows the transform results based on current mode.
func displayTransformSummary(ctx *etlContext, result *devicetypes.TransformResult) {
	if stepFlag {
		displayBriefSummary(result, ctx.opts)
	} else if ctx.debug {
		summary := buildImportSummary(result.Racks, result.Devices, result.Cables)
		visual.PrintImportSummary(summary, ctx.opts, false)
	}
}

// mergeTransformResult merges all transformed entities into ctx.inventory.
func mergeTransformResult(ctx *etlContext, result *devicetypes.TransformResult) {
	result.EnsureUniqueDeviceNames()
	mergeMetadata(ctx, result.Metadata)
	locationRemap := mergeLocations(ctx, result.Locations)
	rackRemap := mergeRacks(ctx, result.Racks)
	remapDeviceParents(result.Devices, locationRemap, rackRemap)
	printImportDiff(ctx, result.Devices)
	mergeDevices(ctx, result.Devices)
	mergeModules(ctx, result.Modules)
	mergeCables(ctx, result.Cables)
	mergeFrus(ctx, result.Frus)

	// Single verify pass after all merges — avoids duplicate warnings
	// from per-entity verify calls.
	log.Printf("Verifying parent-child relationships")
	ctx.inventory.VerifyParentChildRelationships()
}

// runLoadPhase executes the Load phase of the ETL pipeline.
// It persists ctx.inventory to the local datastore.
func runLoadPhase(ctx *etlContext) error {
	if ctx.debug {
		visual.PrintPhaseHeader(phaseNameLoad, ctx.opts)
		visual.PrintCaniOperation("Starting load phase", ctx.opts)
	}

	if stepFlag {
		if err := promptPhaseStart(phaseNameLoad, ctx.opts); err != nil {
			return fmt.Errorf("load phase: %w", err)
		}
	}

	if ctx.debug {
		visual.PrintCaniOperation("Saving inventory to local datastore", ctx.opts)
	}

	if err := datastores.Datastore.Save(ctx.inventory); err != nil {
		return fmt.Errorf("load phase: failed to save inventory: %w", err)
	}

	if ctx.debug {
		visual.PrintPhaseComplete(phaseNameLoad, ctx.opts)
	}
	return nil
}

// buildImportSummary creates a summary from transformed racks, devices, and cables.
func buildImportSummary(
	racks map[uuid.UUID]*devicetypes.CaniRackType,
	devices map[uuid.UUID]*devicetypes.CaniDeviceType,
	cables map[uuid.UUID]*devicetypes.CaniCableType,
) visual.ImportSummary {
	summary := visual.ImportSummary{
		RackNames:     []string{},
		DevicesByRack: make(map[string][]string),
		Cables:        []visual.CableSummary{},
	}

	// Build rack ID to name lookup
	racksByID := make(map[uuid.UUID]string)
	for id, rack := range racks {
		summary.RackNames = append(summary.RackNames, rack.Name)
		racksByID[id] = rack.Name
	}

	// Group devices by their ancestor rack (walk parent chain)
	for _, device := range devices {
		rackName := findAncestorRack(device, devices, racksByID)
		summary.DevicesByRack[rackName] = append(summary.DevicesByRack[rackName], device.Name)
	}

	// Add cable count
	for _, cable := range cables {
		summary.Cables = append(summary.Cables, visual.CableSummary{
			CableType: cable.Slug,
		})
	}

	return summary
}

// findAncestorRack walks the parent chain of a device to find its ancestor rack.
// Returns the rack name if found, or empty string if the device is standalone.
func findAncestorRack(
	device *devicetypes.CaniDeviceType,
	devices map[uuid.UUID]*devicetypes.CaniDeviceType,
	racksByID map[uuid.UUID]string,
) string {
	visited := make(map[uuid.UUID]bool)
	current := device.Parent
	for current != uuid.Nil && !visited[current] {
		visited[current] = true
		if name, ok := racksByID[current]; ok {
			return name
		}
		parent, ok := devices[current]
		if !ok {
			break
		}
		current = parent.Parent
	}
	return ""
}

// promptPhaseStart displays a visual checkpoint and waits for Enter to begin a phase
func promptPhaseStart(phase string, opts visual.ETLOptions) error {
	cyan, yellow, gray := visual.ColorCyan, visual.ColorYellow, visual.ColorGray
	reset := visual.ColorReset
	if opts.NoColor {
		cyan, yellow, gray, reset = "", "", "", ""
	}

	// Calculate padding for phase name to align box
	phaseText := fmt.Sprintf("▶ Ready to begin %s phase", phase)
	boxWidth := 50
	padding := boxWidth - len(phaseText) - 2 // -2 for leading spaces
	if padding < 0 {
		padding = 0
	}

	fmt.Println()
	fmt.Printf("%s┌%s┐%s\n", cyan, strings.Repeat("─", boxWidth), reset)
	fmt.Printf("%s│%s  %s%s%s%s%s│%s\n", cyan, reset, yellow, phaseText, reset, strings.Repeat(" ", padding), cyan, reset)
	fmt.Printf("%s│%s%s%s│%s\n", cyan, reset, strings.Repeat(" ", boxWidth), cyan, reset)
	fmt.Printf("%s│%s  %sPress Enter to continue, or Ctrl+C to abort...%s  %s│%s\n", cyan, reset, gray, reset, cyan, reset)
	fmt.Printf("%s└%s┘%s\n", cyan, strings.Repeat("─", boxWidth), reset)

	reader := bufio.NewReader(os.Stdin)
	_, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	return nil
}

// displayBriefSummary shows a summary of transformed items
func displayBriefSummary(result *devicetypes.TransformResult, opts visual.ETLOptions) error {
	summary := buildImportSummary(result.Racks, result.Devices, result.Cables)
	visual.PrintImportSummary(summary, opts, true)

	fmt.Printf("\nTransformed: %d racks, %d devices, %d cables\n",
		len(result.Racks), len(result.Devices), len(result.Cables))
	return nil
}
