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

	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// NewCommand creates the parent "add" command.
// When called with a slug or part number argument, it searches all registries
// (rack, device, module, cable) and adds the matching hardware type.
// Subcommands restrict to their specific type and reject mismatches.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [slug-or-part-number]",
		Short: "Add items to the inventory",
		Long: `Add items to the inventory.

When called with a slug or part number, searches all hardware registries
(rack, device, module, cable) and automatically determines the type.

Use subcommands (rack, device, module, cable, location) to constrain
to a specific type; subcommands reject slugs that do not match their type.`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
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

	cmd.PersistentFlags().BoolP("auto", "a", false, "Automatically recommend values for parent hardware")
	cmd.PersistentFlags().BoolP("accept", "y", false, "Automatically accept recommended values.")
	cmd.PersistentFlags().BoolP("list-supported-types", "L", false, "List supported hardware types.")
	cmd.PersistentFlags().IntP("qty", "q", 1, "Quantity of items to add.")
	cmd.PersistentFlags().StringP("parent", "p", uuid.Nil.String(), "Parent item UUID.")

	// Let registered providers decorate the add command tree.
	for _, p := range provider.GetProviders() {
		p.NewProviderCmd(cmd)
	}

	return cmd
}

// listAllSupportedTypes prints all available hardware types from every registry.
func listAllSupportedTypes(cmd *cobra.Command) error {
	cmd.SetOut(os.Stderr)
	entries := devicetypes.ListAllAvailableTypes()
	printTypeTable(cmd, entries)
	os.Exit(0)
	return nil
}
