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
package remove

import (
	"fmt"
	"log"

	"github.com/Cray-HPE/cani/internal/util/resolve"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/spf13/cobra"
)

// newLocationCommand creates the "remove location" subcommand.
func newLocationCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "location <uuid-or-name>",
		Short: "Remove a location from the inventory.",
		Long:  "Remove a location by UUID or name.",
		Args:  cobra.ExactArgs(1),
		RunE:  removeLocation,
	}
}

func removeLocation(cmd *cobra.Command, args []string) error {
	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	id, err := resolve.Location(inventory, args[0])
	if err != nil {
		return err
	}

	if err := inventory.RemoveLocation(id); err != nil {
		return fmt.Errorf("failed to remove location: %w", err)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("Removed location %s", id)
	return nil
}
