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
package add

import (
	"fmt"
	"log"
	"strings"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

const flagContentTypes = "content-types"

// newLocationCommand creates the "add location" subcommand.
func newLocationCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "location <slug>",
		Short: "Add a location to the inventory.",
		Long: `Add a location to the inventory using a registered location type slug.

Examples:
  cani alpha add location dc --name "Green Nitrogen"
  cani alpha add location level --name "Level 1"
  cani alpha add location section --name "Section A"

Use -L to list available location type slugs.`,
		Args: func(cmd *cli.Command, args []string) error {
			if cmd.Flags().Changed("list-supported-types") {
				return nil
			}
			if len(args) != 1 {
				return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
			}
			return nil
		},
		RunE: addLocation,
	}

	cmd.Flags().String("name", "", "Location name (required)")
	cmd.Flags().String("parent", "", "Parent location UUID or name")
	cmd.Flags().String(flagDescription, "", "Location description")
	cmd.Flags().String(flagContentTypes, "", "Comma-separated content types (e.g. device,module,rack)")

	return cmd
}

func addLocation(cmd *cli.Command, args []string) error {
	if cmd.Flags().Changed("list-supported-types") {
		return listTypesForNoun(cmd, NounLocation)
	}

	loc, err := buildLocation(cmd, args)
	if err != nil {
		return err
	}

	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	resolveParent(cmd, loc, inventory)

	if err := inventory.AddLocation(loc); err != nil {
		return fmt.Errorf("failed to add location: %w", err)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("Added location %s (%s, type=%s)", loc.ID, loc.Name, loc.LocationType)
	return nil
}

// buildLocation creates a CaniLocationType from a slug and required --name flag.
func buildLocation(cmd *cli.Command, args []string) (*devicetypes.CaniLocationType, error) {
	result, _ := lookupBySlugOrPart(NounLocation, args[0])
	if result == nil || result.Location == nil {
		return nil, fmt.Errorf("unknown location type slug: %s (use -L to list available types)", args[0])
	}
	loc := result.Location

	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		return nil, fmt.Errorf("--name is required")
	}
	loc.Name = name

	applyLocationFlags(cmd, loc)
	return loc, nil
}

// applyLocationFlags overrides location fields from CLI flags when set.
func applyLocationFlags(cmd *cli.Command, loc *devicetypes.CaniLocationType) {
	if cmd.Flags().Changed(flagDescription) {
		loc.Description, _ = cmd.Flags().GetString(flagDescription)
	}
	if cmd.Flags().Changed(flagContentTypes) {
		raw, _ := cmd.Flags().GetString(flagContentTypes)
		if raw != "" {
			loc.ContentTypes = strings.Split(raw, ",")
		}
	}
}

// resolveParent sets the parent UUID from the --parent flag.
// Accepts a UUID directly or looks up a location by name.
func resolveParent(cmd *cli.Command, loc *devicetypes.CaniLocationType, inv *devicetypes.Inventory) {
	parentArg, _ := cmd.Flags().GetString("parent")
	if parentArg == "" {
		return
	}
	if pid, err := uuid.Parse(parentArg); err == nil {
		loc.Parent = pid
		return
	}
	// Try name lookup.
	for _, existing := range inv.Locations {
		if existing.Name == parentArg {
			loc.Parent = existing.ID
			return
		}
	}
}
