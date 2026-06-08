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
	"strings"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/visual"
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
