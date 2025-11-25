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

	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// newDeviceCommand creates the "add device" subcommand.
func newDeviceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "device <slug-or-part-number>",
		Short: "Add device(s) to the inventory.",
		Long:  "Add one or more devices to the inventory by slug or part number.",
		Args:  validSlugOrPartNumber(NounDevice),
		RunE:  addDevice,
	}

	cmd.Flags().String("rack", "", "Parent rack UUID or name")
	cmd.Flags().Int("position", 0, "Rack U position")
	cmd.Flags().String("face", "", "Rack face (front, rear)")

	return cmd
}

func addDevice(cmd *cobra.Command, args []string) error {
	qty, _ := cmd.Flags().GetInt("qty")
	if qty < 1 {
		qty = 1
	}

	result, err := lookupBySlugOrPart(NounDevice, args[0])
	if err != nil {
		return err
	}

	parentArg, _ := cmd.Flags().GetString("parent")
	rackArg, _ := cmd.Flags().GetString("rack")
	position, _ := cmd.Flags().GetInt("position")
	face, _ := cmd.Flags().GetString("face")

	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	locationID := inventory.EnsureLocation()
	inventory.AssignRacksToLocation(locationID)

	devicesToAdd := make(map[uuid.UUID]*devicetypes.CaniDeviceType)
	for range qty {
		device := *result.Device // shallow copy
		device.ID = uuid.New()

		if rackArg != "" {
			if rackID, rerr := uuid.Parse(rackArg); rerr == nil {
				device.Parent = rackID
			}
		} else if parentArg != "" {
			if pid, perr := uuid.Parse(parentArg); perr == nil {
				device.Parent = pid
			}
		}

		device.RackPosition = position
		device.Face = face

		devicesToAdd[device.ID] = &device
	}

	if err := inventory.AddDevices(devicesToAdd); err != nil {
		return fmt.Errorf("failed to add devices: %w", err)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	for _, d := range devicesToAdd {
		log.Printf("Added device %s (%s)", d.ID, d.Name)
	}
	log.Printf("%d device(s) added", qty)
	return nil
}
