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
package remove

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the parent "remove" command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove items from the inventory",
		Long:  `Remove items from the inventory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Help()
			return nil
		},
	}

	// Add noun-based subcommands
	cmd.AddCommand(newLocationCommand())
	cmd.AddCommand(newRackRemoveCommand())
	cmd.AddCommand(newDeviceCommand())
	cmd.AddCommand(newModuleCommand())
	cmd.AddCommand(newCableCommand())

	cmd.PersistentFlags().BoolP("force", "f", false, "Remove items without confirmation.")

	return cmd
}
