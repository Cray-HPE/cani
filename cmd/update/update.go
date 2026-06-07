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
package update

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the parent "update" command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update items in the inventory.",
		Long:  `Update items in the inventory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Help()
			return nil
		},
	}

	// Add noun-based subcommands
	cmd.AddCommand(newLocationCommand())
	cmd.AddCommand(newRackUpdateCommand())
	cmd.AddCommand(newDeviceCommand())
	cmd.AddCommand(newModuleCommand())
	cmd.AddCommand(newCableCommand())
	cmd.AddCommand(newInterfaceCommand())
	cmd.AddCommand(newOrphansCommand())

	cmd.PersistentFlags().StringArray("set", nil, "Set field value as key=value (repeatable)")
	cmd.PersistentFlags().StringArray("tag", nil, "Tag(s) to apply to the item (repeatable)")
	cmd.PersistentFlags().StringArray("metadata", nil, "Provider metadata key=value pairs (repeatable)")

	return cmd
}
