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
package update

import (
	"fmt"
	"log"

	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/internal/util/resolve"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// newDeviceCommand creates the "update device" subcommand.
func newDeviceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "device <uuid-or-name>",
		Short: "Update a device in the inventory.",
		Long:  "Update a device's fields by UUID or name.",
		Args:  cobra.ExactArgs(1),
		RunE:  updateDevice,
	}

	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("status", "", "New status")
	cmd.Flags().String("role", "", "New role")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().Int("position", 0, "Rack U position")
	cmd.Flags().String("face", "", "Rack face (front, rear)")
	cmd.Flags().Bool("swap", false, "Swap position with the device occupying the target slot")
	cmd.Flags().String("parent", "", "Parent UUID or name (rack or device)")
	cmd.Flags().String("primary-ipv4", "", "Primary IPv4 address (CIDR or UUID)")
	cmd.Flags().String("primary-ipv6", "", "Primary IPv6 address (CIDR or UUID)")

	// Let providers contribute their own device-update flags.
	for _, p := range provider.GetProviders() {
		if fp, ok := p.(provider.DeviceUpdateFlagProvider); ok {
			fp.RegisterDeviceUpdateFlags(cmd)
		}
	}

	return cmd
}

func updateDevice(cmd *cobra.Command, args []string) error {
	inventory, id, device, err := loadDeviceForUpdate(cmd, args)
	if err != nil {
		return err
	}

	applyScalarFields(cmd, device)

	if err := applyPositionUpdate(cmd, inventory, id, device); err != nil {
		return err
	}
	if err := applyParentUpdate(cmd, inventory, device); err != nil {
		return err
	}
	if err := applyTagsAndMetadata(cmd, device); err != nil {
		return err
	}
	if err := applyProviderDeviceFlags(cmd, device); err != nil {
		return err
	}
	if err := applyPrimaryIPs(cmd, inventory, device); err != nil {
		return err
	}
	if err := applySetToDevice(cmd, device); err != nil {
		return err
	}

	return finalizeDeviceUpdate(inventory, id, device)
}

// loadDeviceForUpdate sets the device store, loads the inventory, and resolves
// the target device from args.
func loadDeviceForUpdate(cmd *cobra.Command, args []string) (*devicetypes.Inventory, uuid.UUID, *devicetypes.CaniDeviceType, error) {
	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return nil, uuid.Nil, nil, fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return nil, uuid.Nil, nil, fmt.Errorf("failed to load inventory: %w", err)
	}

	id, err := resolve.Device(inventory, args[0])
	if err != nil {
		return nil, uuid.Nil, nil, err
	}

	return inventory, id, inventory.Devices[id], nil
}

// finalizeDeviceUpdate rebuilds derived relationships and persists the inventory.
func finalizeDeviceUpdate(inventory *devicetypes.Inventory, id uuid.UUID, device *devicetypes.CaniDeviceType) error {
	// Rebuild relationships so derived fields are updated.
	result := inventory.VerifyParentChildRelationships()
	if result.HasErrors() {
		return fmt.Errorf("relationship errors: %v", result.Errors)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("Updated device %s (%s)", id, device.Name)
	return nil
}

// applyProviderDeviceFlags lets each registered provider apply its own
// device-update flags to the device.
func applyProviderDeviceFlags(cmd *cobra.Command, device *devicetypes.CaniDeviceType) error {
	for _, p := range provider.GetProviders() {
		fp, ok := p.(provider.DeviceUpdateFlagProvider)
		if !ok {
			continue
		}
		if err := fp.ApplyDeviceUpdateFlags(cmd, device); err != nil {
			return err
		}
	}
	return nil
}
