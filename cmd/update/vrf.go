/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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
	"fmt"
	"log"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/util/resolve"
	"github.com/Cray-HPE/cani/internal/util/store"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/google/uuid"
)

// newVRFCommand creates the "update vrf" subcommand.
func newVRFCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "vrf <name-or-uuid>",
		Short: "Update a VRF in the inventory.",
		Long: `Update a VRF's fields or add devices to it.

Examples:
  cani alpha update vrf keepalive --add-device sw-spine-02
  cani alpha update vrf LEGACY --description "Updated description"
  cani alpha update vrf keepalive --remove-device sw-leaf-01`,
		Args: cli.ExactArgs(1),
		RunE: updateVRF,
	}

	cmd.Flags().String("description", "", "New description")
	cmd.Flags().String("rd", "", "Route distinguisher")
	cmd.Flags().StringArray("add-device", nil, "Add a device to this VRF (name or UUID, repeatable)")
	cmd.Flags().StringArray("remove-device", nil, "Remove a device from this VRF (name or UUID, repeatable)")

	return cmd
}

func updateVRF(cmd *cli.Command, args []string) error {
	ref := args[0]

	if err := store.Setup(cmd); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	vrf, err := inventory.FindVRFByNameOrID(ref)
	if err != nil {
		return err
	}

	changed := false

	if cmd.Flags().Changed("description") {
		vrf.Description, _ = cmd.Flags().GetString("description")
		changed = true
	}
	if cmd.Flags().Changed("rd") {
		vrf.RD, _ = cmd.Flags().GetString("rd")
		changed = true
	}

	addDevices, _ := cmd.Flags().GetStringArray("add-device")
	for _, ref := range addDevices {
		id, err := resolve.Device(inventory, ref)
		if err != nil {
			return fmt.Errorf("resolving --add-device %q: %w", ref, err)
		}
		if !containsUUID(vrf.Devices, id) {
			vrf.Devices = append(vrf.Devices, id)
			changed = true
		}
	}

	removeDevices, _ := cmd.Flags().GetStringArray("remove-device")
	for _, ref := range removeDevices {
		id, err := resolve.Device(inventory, ref)
		if err != nil {
			return fmt.Errorf("resolving --remove-device %q: %w", ref, err)
		}
		if removed := removeUUID(&vrf.Devices, id); removed {
			changed = true
		}
	}

	if !changed {
		return fmt.Errorf("no changes specified")
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("Updated VRF %q (%s)", vrf.Name, vrf.ID)
	return nil
}

func containsUUID(slice []uuid.UUID, id uuid.UUID) bool {
	for _, v := range slice {
		if v == id {
			return true
		}
	}
	return false
}

func removeUUID(slice *[]uuid.UUID, id uuid.UUID) bool {
	for i, v := range *slice {
		if v == id {
			*slice = append((*slice)[:i], (*slice)[i+1:]...)
			return true
		}
	}
	return false
}
