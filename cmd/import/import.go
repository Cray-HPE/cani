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

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/visual"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// Phase constants
const (
	PhaseExtract   = "e"
	PhaseTransform = "t"
	PhaseLoad      = "l"
)

// Package-level flags
var (
	phaseFlag   string
	noColorFlag bool
	stepFlag    bool
)

// ErrAborted is returned when user aborts during step mode
var ErrAborted = fmt.Errorf("aborted by user")

// etlContext encapsulates shared state for the ETL pipeline.
// Data flows through the pipeline as:
//
//	Extract (provider.Import) → Transform (provider.Transform) → Load (Datastore.Save)
//
// The inventory field is mutated by each phase and passed to the next.
type etlContext struct {
	cmd       *cobra.Command
	args      []string
	provider  provider.Provider
	inventory *devicetypes.Inventory
	opts      visual.ETLOptions
	debug     bool
}

// newETLContext creates and initializes the ETL context.
// It validates flags and propagates settings to the global config.
func newETLContext(cmd *cobra.Command, args []string, p provider.Provider) (*etlContext, error) {
	if err := validatePhaseFlag(); err != nil {
		return nil, err
	}

	// Enable debug mode if step flag is set
	if stepFlag && !config.Cfg.Debug {
		config.Cfg.Debug = true
		log.Println("Note: --step implies --debug, enabling debug mode")
	}

	// Propagate step/color flags to config for provider access
	config.Cfg.StepMode = stepFlag
	config.Cfg.NoColor = noColorFlag

	return &etlContext{
		cmd:      cmd,
		args:     args,
		provider: p,
		opts:     visual.ETLOptions{NoColor: noColorFlag},
		debug:    config.Cfg.Debug,
	}, nil
}

// ValidPhases returns the list of valid phase values
func ValidPhases() []string {
	return []string{"e", "et", "etl"}
}

// NewCommand creates the import command with provider subcommands
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import PROVIDER [flags]",
		Short: "Import assets into the inventory",
		Long:  `Import assets into the inventory from an external source using a provider.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Help()
			return nil
		},
	}

	// Add persistent flags for all import subcommands
	cmd.PersistentFlags().StringVar(&phaseFlag, "phase", "etl", "ETL phases to run: e (extract), et (+transform), etl (+load)")
	cmd.PersistentFlags().BoolVar(&noColorFlag, "no-color", false, "Disable colorized output")
	cmd.PersistentFlags().BoolVar(&stepFlag, "step", false, "Step through each item interactively (implies --debug)")

	// Add provider subcommands
	addProviderSubcommands(cmd)

	return cmd
}

// addProviderSubcommands adds a subcommand for each registered provider
func addProviderSubcommands(importCmd *cobra.Command) {
	for _, p := range provider.GetProviders() {
		// Get provider-specific import command (with flags)
		providerImportCmd, err := p.NewProviderCmd(importCmd)
		if err != nil || providerImportCmd == nil {
			// Provider doesn't support import, create a basic subcommand
			providerImportCmd = &cobra.Command{}
		}

		p := p // capture for closure
		providerImportCmd.Use = p.Slug()
		providerImportCmd.Short = fmt.Sprintf("Import assets using the %s provider", p.Slug())

		// Wrap the provider's RunE with the ETL logic
		origRunE := providerImportCmd.RunE
		providerImportCmd.RunE = func(cmd *cobra.Command, args []string) error {
			// 1) Run provider's validation/setup if it has one
			if origRunE != nil {
				if err := origRunE(cmd, args); err != nil {
					return err
				}
			}
			// 2) Run the ETL pipeline
			return runImportETL(cmd, args, p)
		}

		importCmd.AddCommand(providerImportCmd)
	}
}

// shouldRunPhase checks if a phase should be executed based on the --phase flag
func shouldRunPhase(phase string) bool {
	return strings.Contains(phaseFlag, phase)
}

// validatePhaseFlag validates the --phase flag value
func validatePhaseFlag() error {
	for _, valid := range ValidPhases() {
		if phaseFlag == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid phase %q: must be one of %v", phaseFlag, ValidPhases())
}

// runImportETL executes the Extract-Transform-Load pipeline for a provider.
// Data flows through three phases:
//  1. Extract: provider.Import() populates ctx.inventory with raw data
//  2. Transform: provider.Transform() converts data and merges into ctx.inventory
//  3. Load: Datastore.Save() persists ctx.inventory to disk
func runImportETL(cmd *cobra.Command, args []string, p provider.Provider) error {
	ctx, err := newETLContext(cmd, args, p)
	if err != nil {
		return err
	}

	if err := initializeDatastore(ctx); err != nil {
		return err
	}

	// Run ETL phases based on --phase flag
	if shouldRunPhase(PhaseExtract) {
		if err := runExtractPhase(ctx); err != nil {
			return err
		}
	}

	if shouldRunPhase(PhaseTransform) {
		if err := runTransformPhase(ctx); err != nil {
			return err
		}
	}

	if shouldRunPhase(PhaseLoad) {
		if err := runLoadPhase(ctx); err != nil {
			return err
		}
	}

	log.Printf("Import completed successfully using provider %s", p.Slug())
	return nil
}

// initializeDatastore sets up the datastore and loads existing inventory into ctx.
func initializeDatastore(ctx *etlContext) error {
	if ctx.debug {
		visual.PrintCaniOperation("Initializing inventory datastore", ctx.opts)
	}

	if err := datastores.SetDeviceStore(ctx.cmd, ctx.args); err != nil {
		return fmt.Errorf("init: failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("init: failed to load existing inventory: %w", err)
	}
	ctx.inventory = inventory

	// Warn if running transform/load without prior data
	if (phaseFlag == PhaseTransform || phaseFlag == PhaseLoad) && len(ctx.inventory.Devices) == 0 {
		visual.PrintWarning("No existing devices found - prior phases may not have been run", ctx.opts)
	}

	return nil
}

// runExtractPhase executes the Extract phase of the ETL pipeline.
// It calls provider.Import() which populates ctx.inventory with raw data from the source.
func runExtractPhase(ctx *etlContext) error {
	if ctx.debug {
		visual.PrintPhaseHeader("EXTRACT", ctx.opts)
		visual.PrintCaniOperation("Starting extract phase", ctx.opts)
	}

	if stepFlag {
		if err := promptPhaseStart("EXTRACT", ctx.opts); err != nil {
			return fmt.Errorf("extract phase: %w", err)
		}
	}

	importer, ok := ctx.provider.(provider.Importer)
	if !ok {
		if ctx.debug {
			visual.PrintPhaseComplete("EXTRACT", ctx.opts)
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
		visual.PrintPhaseComplete("EXTRACT", ctx.opts)
	}
	return nil
}

// runTransformPhase executes the Transform phase of the ETL pipeline.
// It calls provider.Transform() and merges the result into ctx.inventory.
func runTransformPhase(ctx *etlContext) error {
	if ctx.debug {
		visual.PrintPhaseHeader("TRANSFORM", ctx.opts)
		visual.PrintCaniOperation("Starting transform phase", ctx.opts)
	}

	if stepFlag {
		if err := promptPhaseStart("TRANSFORM", ctx.opts); err != nil {
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
		visual.PrintPhaseComplete("TRANSFORM", ctx.opts)
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
	mergeLocations(ctx, result.Locations)
	mergeRacks(ctx, result.Racks)
	mergeDevices(ctx, result.Devices)
	mergeModules(ctx, result.Modules)
	mergeCables(ctx, result.Cables)
	mergeFrus(ctx, result.Frus)

	// Single verify pass after all merges — avoids duplicate warnings
	// from per-entity verify calls.
	log.Printf("Verifying parent-child relationships")
	ctx.inventory.VerifyParentChildRelationships()
}

// mergeLocations adds transformed locations to the inventory.
func mergeLocations(ctx *etlContext, locations map[uuid.UUID]*devicetypes.CaniLocationType) {
	if len(locations) == 0 {
		return
	}
	if ctx.debug {
		visual.PrintCaniOperation(fmt.Sprintf("Merging %d locations into inventory", len(locations)), ctx.opts)
	}
	ctx.inventory.MergeLocations(locations)
}

// mergeRacks adds transformed racks to the inventory.
func mergeRacks(ctx *etlContext, racks map[uuid.UUID]*devicetypes.CaniRackType) {
	if len(racks) == 0 {
		return
	}
	if ctx.debug {
		visual.PrintCaniOperation(fmt.Sprintf("Merging %d racks into inventory", len(racks)), ctx.opts)
	}
	ctx.inventory.MergeRacks(racks)
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

	strict := config.Cfg.Strict
	skipped := ctx.inventory.MergeDevicesStrict(devices, strict)

	if len(skipped) == 0 {
		return
	}

	// In step mode, prompt user to classify each skipped device.
	if stepFlag {
		classifyOpts := devicetypes.ClassifyOptions{NoColor: noColorFlag}
		classified := 0
		for _, ud := range skipped {
			slug, err := devicetypes.PromptForDeviceType(ud, classifyOpts)
			if err != nil {
				log.Printf("  ! %s: classification error: %v", ud.Name, err)
				continue
			}
			device := devices[ud.ID]
			if device == nil {
				continue
			}
			if slug == "" {
				log.Printf("  - %s: skipped (no type selected)", ud.Name)
			} else {
				if err := devicetypes.ApplyDeviceType(device, slug); err != nil {
					log.Printf("  ! %s: failed to apply type %q: %v", ud.Name, slug, err)
				} else {
					classified++
				}
			}
			// Always merge the device so modules/FRUs can reference it as a parent.
			ctx.inventory.MergeDevices(map[uuid.UUID]*devicetypes.CaniDeviceType{ud.ID: device})
		}
		if classified > 0 {
			log.Printf("  Classified %d of %d unclassified devices", classified, len(skipped))
		}
		return
	}

	// Non-interactive: warn and suggest the classify command.
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

	// Always merge skipped devices so modules/FRUs can reference them as parents.
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

// runLoadPhase executes the Load phase of the ETL pipeline.
// It persists ctx.inventory to the local datastore.
func runLoadPhase(ctx *etlContext) error {
	if ctx.debug {
		visual.PrintPhaseHeader("LOAD", ctx.opts)
		visual.PrintCaniOperation("Starting load phase", ctx.opts)
	}

	if stepFlag {
		if err := promptPhaseStart("LOAD", ctx.opts); err != nil {
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
		visual.PrintPhaseComplete("LOAD", ctx.opts)
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
