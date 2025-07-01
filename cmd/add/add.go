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
	"log"

	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// NewCommand creates the parent "add" command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add",
		Short:   "Add items to the inventory",
		Long:    `Add items to the inventory.`,
		PreRunE: provider.GetActiveProvider,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Help()
			return nil
		},
	}

	// Add all subcommands
	cmd.AddCommand(newRackCommand())
	cmd.AddCommand(newBladeCommand())

	cmd.PersistentFlags().BoolP("auto", "a", false, "Automatically recommend values for parent hardware")
	cmd.PersistentFlags().BoolP("accept", "y", false, "Automatically accept recommended values.")
	cmd.PersistentFlags().BoolP("list-supported-types", "L", false, "List supported hardware types.")
	cmd.MarkFlagsRequiredTogether("list-supported-types")
	cmd.PersistentFlags().IntP("qty", "q", 1, "Quantity of device types to add.")
	cmd.PersistentFlags().StringP("parent", "p", uuid.Nil.String(), "Parent device ID.")
	cmd.MarkFlagsMutuallyExclusive("auto")

	return cmd
}

// add is the main entry point for the add command.
func add(cmd *cobra.Command, args []string) error {
	var devicetype devicetypes.DeviceType

	switch cmd.Name() {
	case "blade":
		devicetype = devicetypes.Blades()[args[0]]
	case "rack":
		devicetype = devicetypes.Racks()[args[0]]
	default:
		return fmt.Errorf("unknown device type: %s", args[0])
	}

	devicesToAdd, err := provider.ActiveProvider.Add(cmd, args, devicetype)
	if err != nil {
		return err
	}

	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}
	if err := datastores.Datastore.Create(devicesToAdd); err != nil {
		return fmt.Errorf("failed to create devices in datastore: %w", err)
	}

	log.Println("")
	for _, device := range devicesToAdd {
		if device == nil {
			log.Printf("No devices to add for %s", devicetype.Type)
			continue
		}
		if device.Parent == uuid.Nil && device.Type != devicetypes.Rack {
			log.Printf("Added %s (%s) without a parent", device.ID, device.Name)
			continue
		}
		log.Printf("Added %s (%s) with parent %s", device.ID, device.Name, device.Parent)
	}
	log.Println("")
	log.Printf("%d %s added to the inventory", len(devicesToAdd), devicetype.Type)
	return nil
}
