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
package export

import (
	"fmt"
	"log"

	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/spf13/cobra"
)

// NewCommand creates the parent "export" command with provider subcommands
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export PROVIDER [flags]",
		Short: "Export inventory to an external provider",
		Long:  `Export the CANI inventory to an external provider using a provider.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Help()
			return nil
		},
	}

	// Add common persistent flags for all export subcommands
	cmd.PersistentFlags().Bool("merge", false, "Update existing devices instead of skipping conflicts")
	cmd.PersistentFlags().Bool("dry-run", false, "Preview changes without making API calls")

	// Add provider subcommands
	addProviderSubcommands(cmd)

	// Add non-provider subcommands
	cmd.AddCommand(newConnectionsCommand())

	return cmd
}

// addProviderSubcommands adds a subcommand for each registered provider
func addProviderSubcommands(exportCmd *cobra.Command) {
	for _, p := range provider.GetProviders() {
		// Get provider-specific export command (with flags)
		providerExportCmd, err := p.NewProviderCmd(exportCmd)
		if err != nil || providerExportCmd == nil {
			// Provider doesn't support export, create a basic subcommand
			providerExportCmd = &cobra.Command{}
		}

		p := p // capture for closure
		providerExportCmd.Use = p.Slug()
		providerExportCmd.Short = fmt.Sprintf("Export inventory using the %s provider", p.Slug())

		// Wrap the provider's RunE with the export logic
		origRunE := providerExportCmd.RunE
		providerExportCmd.RunE = func(cmd *cobra.Command, args []string) error {
			// 1) Run provider's validation/setup if it has one
			if origRunE != nil {
				if err := origRunE(cmd, args); err != nil {
					return err
				}
			}
			// 2) Run the export pipeline
			return runExport(cmd, args, p)
		}

		exportCmd.AddCommand(providerExportCmd)
	}
}

// runExport is the main entry point for the export command.
func runExport(cmd *cobra.Command, args []string, p provider.Provider) error {
	// Load the inventory from the datastore
	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inv, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	if inv == nil || (len(inv.Locations) == 0 && len(inv.Racks) == 0 && len(inv.Devices) == 0 && len(inv.Modules) == 0 && len(inv.Cables) == 0 && len(inv.Frus) == 0) {
		return fmt.Errorf("inventory is empty, nothing to export")
	}

	// Check if the provider implements Exporter
	exporter, ok := p.(provider.Exporter)
	if !ok {
		return fmt.Errorf("provider '%s' does not support export", p.Slug())
	}

	log.Printf("Exporting inventory to %s (locations=%d, racks=%d, devices=%d, modules=%d)...",
		p.Slug(), len(inv.Locations), len(inv.Racks), len(inv.Devices), len(inv.Modules))

	// Call the provider's Export method to sync to external system
	if err := exporter.Export(cmd.Context(), cmd, args, inv); err != nil {
		// Save inventory even on error so external IDs are persisted
		_ = datastores.Datastore.Save(inv)
		return fmt.Errorf("export failed: %w", err)
	}

	// Persist external IDs (and any other mutations) back to the datastore
	if err := datastores.Datastore.Save(inv); err != nil {
		return fmt.Errorf("failed to save inventory after export: %w", err)
	}

	log.Println("Export completed successfully")
	return nil
}
