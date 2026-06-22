/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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
	"os"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// NewCommand creates the parent "add" command.
// When called with a slug or part number argument, it searches all registries
// (rack, device, module, cable) and adds the matching hardware type.
// Subcommands restrict to their specific type and reject mismatches.
func NewCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "add [slug-or-part-number]",
		Short: "Add items to the inventory",
		Long: `Add items to the inventory.

When called with a slug or part number, searches all hardware registries
(rack, device, module, cable) and automatically determines the type.

Use subcommands (rack, device, module, cable, location) to constrain
to a specific type; subcommands reject slugs that do not match their type.`,
		Args: cli.ArbitraryArgs,
		RunE: func(cmd *cli.Command, args []string) error {
			if cmd.Flags().Changed("list-supported-types") {
				return listAllSupportedTypes(cmd)
			}
			if len(args) > 0 {
				return addAny(cmd, args)
			}
			fmt.Fprintln(cmd.ErrOrStderr(), "Provide a slug or part number, or use a subcommand.")
			return cmd.Help()
		},
	}

	// Add noun-based subcommands
	cmd.AddCommand(newLocationCommand())
	cmd.AddCommand(newRackAddCommand())
	cmd.AddCommand(newDeviceCommand())
	cmd.AddCommand(newModuleCommand())
	cmd.AddCommand(newCableCommand())
	cmd.AddCommand(newMetadataCommand())
	cmd.AddCommand(newConnectionsCommand())
	cmd.AddCommand(newVLANCommand())
	cmd.AddCommand(newPrefixCommand())
	cmd.AddCommand(newIPCommand())

	cmd.PersistentFlags().BoolP("auto", "a", false, "Automatically recommend values for parent hardware")
	cmd.PersistentFlags().BoolP("accept", "y", false, "Automatically accept recommended values.")
	cmd.PersistentFlags().BoolP("list-supported-types", "L", false, "List supported hardware types.")
	cmd.PersistentFlags().IntP("qty", "q", 1, "Quantity of items to add.")
	cmd.PersistentFlags().StringP("parent", "p", uuid.Nil.String(), "Parent item UUID.")
	cmd.PersistentFlags().String("prefix", "", "Name prefix for sequential naming (used with --qty).")
	cmd.PersistentFlags().Int("start", 1, "Starting number for sequential names (used with --prefix).")
	cmd.PersistentFlags().Int("pad-width", 0, "Zero-pad width for sequential names (0 = auto).")
	cmd.PersistentFlags().StringArray("tag", nil, "Tag(s) to apply to the item (repeatable)")
	cmd.PersistentFlags().StringArray("metadata", nil, "Provider metadata key=value pairs (repeatable)")
	cmd.PersistentFlags().String("status", "", "Status (Active, Available, Connected, Decommissioned, Decommissioning, Deprecated, Deprovisioning, Down, End-of-Life, Extended Support, Failed, Inventory, Maintenance, Offline, Planned, Primary, Provisioning, Reserved, Retired, Secondary, Staging, or any custom status)")
	cmd.PersistentFlags().String("serial", "", "Serial number")

	// Flags for slug-based adds (not persistent – subcommands define their own).
	cmd.Flags().String("location", "", "Parent location UUID or name")
	cmd.Flags().String("name", "", "Name for the added item")

	return cmd
}

// listAllSupportedTypes prints all available hardware types from every registry.
func listAllSupportedTypes(cmd *cli.Command) error {
	cmd.SetOut(os.Stderr)
	entries := devicetypes.ListAllAvailableTypes()
	printTypeTable(cmd, entries)
	os.Exit(0)
	return nil
}
