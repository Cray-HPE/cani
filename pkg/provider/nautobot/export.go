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

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/nautobot/export"
)

// Export implements the provider.Exporter interface.
// It syncs the local inventory to Nautobot.
func (p *Nautobot) Export(ctx context.Context, cmd *cli.Command, args []string, inventory *devicetypes.Inventory) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := p.setupExportFromConfig(cmd); err != nil {
		return err
	}
	return p.exporter.Load(inventory)
}

// ExportDryRun reports the dry-run setting the last Export resolved from the
// config file and the --dry-run flag, so the command layer can skip saving an
// inventory that was only planned against.
func (p *Nautobot) ExportDryRun() bool {
	return p.Options != nil && p.Options.Export.DryRun
}

// setupExportFromConfig initialises the Nautobot client, lookup cache and
// exporter from config values and command-line flag overrides.
func (p *Nautobot) setupExportFromConfig(cmd *cli.Command) error {
	if err := p.loadOptionsFromConfig(); err != nil {
		return fmt.Errorf("failed to load nautobot config: %w", err)
	}

	p.applyFlagOverrides(cmd)

	if p.Options.URL == "" {
		return fmt.Errorf("nautobot URL not configured; set nautobot.url in the config file")
	}
	if p.Options.Token == "" {
		return fmt.Errorf("nautobot token not configured; set nautobot.token in the config file")
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

	p.applyCacheFlags(cache)

	if err := p.promptForDefaults(cache); err != nil {
		return err
	}

	p.exporter = export.NewExporter(client, cache, &export.ExporterOpts{
		DefaultLocation:     p.Options.DefaultLocation,
		DefaultRole:         p.Options.DefaultRole,
		DefaultStatus:       p.Options.DefaultStatus,
		Merge:               p.Options.Export.Merge,
		DryRun:              p.Options.Export.DryRun,
		Strict:              false,
		CreateDeviceTypes:   p.Options.Export.CreateDeviceTypes,
		CreateLocationTypes: p.Options.Export.CreateLocationTypes,
		CreateModuleTypes:   p.Options.Export.CreateModuleTypes,
		CreateLocations:     p.Options.Export.CreateLocations,
		CreateStatuses:      p.Options.Export.CreateStatuses,
		CreateRoles:         p.Options.Export.CreateRoles,
	})

	return nil
}

// applyFlagOverrides applies command-line flag overrides to Options. A flag
// given on the command line wins over the config file in either direction;
// one left at its default changes nothing.
func (p *Nautobot) applyFlagOverrides(cmd *cli.Command) {
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
	for _, flag := range []struct {
		name string
		dst  *bool
	}{
		{"merge", &p.Options.Export.Merge},
		{"dry-run", &p.Options.Export.DryRun},
		{"create-device-types", &p.Options.Export.CreateDeviceTypes},
		{"create-location-types", &p.Options.Export.CreateLocationTypes},
		{"create-module-types", &p.Options.Export.CreateModuleTypes},
		{"create-locations", &p.Options.Export.CreateLocations},
		{"create-statuses", &p.Options.Export.CreateStatuses},
		{"create-roles", &p.Options.Export.CreateRoles},
	} {
		if cmd.Flags().Changed(flag.name) {
			*flag.dst, _ = cmd.Flags().GetBool(flag.name)
		}
	}
}

// applyCacheFlags mirrors the create_* options, config and flags already
// merged, into the lookup cache.
func (p *Nautobot) applyCacheFlags(cache *export.LookupCache) {
	opts := p.Options.Export
	cache.SetCreateDeviceTypes(opts.CreateDeviceTypes)
	cache.SetCreateLocationTypes(opts.CreateLocationTypes)
	cache.SetCreateModuleTypes(opts.CreateModuleTypes)
	cache.SetCreateLocations(opts.CreateLocations)
	cache.SetCreateStatuses(opts.CreateStatuses)
	cache.SetCreateRoles(opts.CreateRoles)
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
