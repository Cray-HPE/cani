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
	"fmt"

	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/spf13/cobra"
)

// NewCommand creates the parent "add" command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove",
		Short:   "Remove items from the inventory",
		Long:    `Remove items from the inventory.`,
		PreRunE: provider.GetActiveProvider,
		Args:    cobra.ArbitraryArgs,
		RunE:    remove,
	}

	// Add all subcommands
	cmd.RemoveCommand(newRackCommand())
	cmd.RemoveCommand(newBladeCommand())

	cmd.PersistentFlags().BoolP("force", "f", false, "Remove devices without confirmation.")
	cmd.MarkFlagsMutuallyExclusive("force")

	return cmd
}

// remove is the main entry point for the remove command.
func remove(cmd *cobra.Command, args []string) error {
	devicesToRemove, err := provider.ActiveProvider.Remove(cmd, args)
	if err != nil {
		return err
	}

	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}
	if err := datastores.Datastore.Delete(devicesToRemove); err != nil {
		return fmt.Errorf("failed to delete devices from datastore: %w", err)
	}

	return nil
}
