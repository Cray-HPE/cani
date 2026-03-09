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
package nautobot

import (
	"context"
	"fmt"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/nautobot/export"
	"github.com/spf13/cobra"
)

// Export implements the provider.Exporter interface.
// It syncs the local inventory to Nautobot.
func (p *Nautobot) Export(cmd *cobra.Command, args []string, inventory *devicetypes.Inventory) error {
	if err := p.setupExportFromConfig(cmd); err != nil {
		return err
	}
	return p.exporter.Load(inventory)
}

// setupExportFromConfig initialises the Nautobot client, lookup cache and
// exporter from config values and command-line flag overrides.
func (p *Nautobot) setupExportFromConfig(cmd *cobra.Command) error {
	if err := p.loadOptionsFromConfig(); err != nil {
		return fmt.Errorf("failed to load nautobot config: %w", err)
	}

	p.applyFlagOverrides(cmd)

	if p.Options.URL == "" {
		return fmt.Errorf("nautobot URL not configured. Use --url flag or run 'cani alpha session init nautobot' first")
	}
	if p.Options.Token == "" {
		return fmt.Errorf("nautobot token not configured. Use --token flag or run 'cani alpha session init nautobot' first")
	}

	client, err := export.NewNautobotClient(p.Options.URL, p.Options.Token)
	if err != nil {
		return fmt.Errorf("failed to create Nautobot client: %w", err)
	}

	ctx := context.Background()
	if err := client.TestConnection(ctx); err != nil {
		return fmt.Errorf("failed to connect to Nautobot: %w", err)
	}
	clog.Info("Successfully connected to Nautobot")

	cache := export.NewLookupCache(client)
	cache.SetContext(ctx)

	p.applyCacheFlags(cmd, cache)

	if err := p.promptForDefaults(cache); err != nil {
		return err
	}

	p.exporter = export.NewExporter(client, cache, &export.ExporterOpts{
		DefaultLocation:   p.Options.DefaultLocation,
		DefaultRole:       p.Options.DefaultRole,
		DefaultStatus:     p.Options.DefaultStatus,
		Merge:             p.Options.Export.Merge,
		DryRun:            p.Options.Export.DryRun,
		Strict:            false,
		CreateDeviceTypes: p.Options.Export.CreateDeviceTypes,
		CreateLocations:   p.Options.Export.CreateLocations,
		CreateStatuses:    p.Options.Export.CreateStatuses,
		CreateRoles:       p.Options.Export.CreateRoles,
	})

	return nil
}

// applyFlagOverrides applies command-line flag overrides to Options.
func (p *Nautobot) applyFlagOverrides(cmd *cobra.Command) {
	if cmd.Flags().Changed("url") {
		p.Options.URL, _ = cmd.Flags().GetString("url")
	}
	if cmd.Flags().Changed("token") {
		p.Options.Token, _ = cmd.Flags().GetString("token")
	}
	if cmd.Flags().Changed("default-location") {
		p.Options.DefaultLocation, _ = cmd.Flags().GetString("default-location")
	}
	if cmd.Flags().Changed("default-role") {
		p.Options.DefaultRole, _ = cmd.Flags().GetString("default-role")
	}
	if cmd.Flags().Changed("default-status") {
		p.Options.DefaultStatus, _ = cmd.Flags().GetString("default-status")
	}
	if p.Options.Export == nil {
		p.Options.Export = &NautobotExportOpts{}
	}
	if cmd.Flags().Changed("merge") {
		p.Options.Export.Merge, _ = cmd.Flags().GetBool("merge")
	}
	if cmd.Flags().Changed("dry-run") {
		p.Options.Export.DryRun, _ = cmd.Flags().GetBool("dry-run")
	}
}

// applyCacheFlags applies create_* flags from config and CLI overrides to the cache.
func (p *Nautobot) applyCacheFlags(cmd *cobra.Command, cache *export.LookupCache) {
	// Apply config values first
	if p.Options.Export != nil {
		if p.Options.Export.CreateDeviceTypes {
			cache.SetCreateDeviceTypes(true)
		}
		if p.Options.Export.CreateLocations {
			cache.SetCreateLocations(true)
		}
		if p.Options.Export.CreateStatuses {
			cache.SetCreateStatuses(true)
		}
		if p.Options.Export.CreateRoles {
			cache.SetCreateRoles(true)
		}
	}

	// CLI overrides
	if cmd.Flags().Changed("create-device-types") {
		if v, _ := cmd.Flags().GetBool("create-device-types"); v {
			cache.SetCreateDeviceTypes(true)
		}
	}
	if cmd.Flags().Changed("create-statuses") {
		if v, _ := cmd.Flags().GetBool("create-statuses"); v {
			cache.SetCreateStatuses(true)
		}
	}
	if cmd.Flags().Changed("create-roles") {
		if v, _ := cmd.Flags().GetBool("create-roles"); v {
			cache.SetCreateRoles(true)
		}
	}
	if cmd.Flags().Changed("create-locations") {
		if v, _ := cmd.Flags().GetBool("create-locations"); v {
			cache.SetCreateLocations(true)
		}
	}
}

// promptForDefaults prompts the user interactively for any missing defaults
// unless auto-creation is enabled for the corresponding resource type.
func (p *Nautobot) promptForDefaults(cache *export.LookupCache) error {
	createLocations := p.Options.Export != nil && p.Options.Export.CreateLocations
	createRoles := p.Options.Export != nil && p.Options.Export.CreateRoles
	createStatuses := p.Options.Export != nil && p.Options.Export.CreateStatuses

	if p.Options.DefaultLocation == "" && !createLocations {
		loc, err := promptForSelection("location", func() ([]*export.CachedItem, error) {
			return cache.ListLocations()
		})
		if err != nil {
			return fmt.Errorf("failed to select location: %w", err)
		}
		p.Options.DefaultLocation = loc
	}

	if p.Options.DefaultRole == "" && !createRoles {
		role, err := promptForSelection("role", func() ([]*export.CachedItem, error) {
			return cache.ListRoles()
		})
		if err != nil {
			return fmt.Errorf("failed to select role: %w", err)
		}
		p.Options.DefaultRole = role
	}

	if p.Options.DefaultStatus == "" && !createStatuses {
		status, err := promptForSelection("status", func() ([]*export.CachedItem, error) {
			return cache.ListStatuses()
		})
		if err != nil {
			return fmt.Errorf("failed to select status: %w", err)
		}
		p.Options.DefaultStatus = status
	}

	return nil
}
