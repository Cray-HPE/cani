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
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/internal/util/resolve"
	"github.com/Cray-HPE/cani/internal/util/store"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

const (
	flagAssignVLAN = "assign-vlan"
	flagBMCOf      = "bmc-of"
)

// newDeviceCommand creates the "update device" subcommand.
func newDeviceCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "device <uuid-or-name>",
		Short: "Update a device in the inventory.",
		Long:  "Update a device's fields by UUID or name.",
		Args:  cli.ExactArgs(1),
		RunE:  updateDevice,
	}

	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("status", "", "New status")
	cmd.Flags().String("role", "", "New role")
	cmd.Flags().String(flagDescription, "", "Description")
	cmd.Flags().Int(flagPosition, 0, "Rack U position")
	cmd.Flags().String("face", "", "Rack face (front, rear)")
	cmd.Flags().Bool("swap", false, "Swap position with the device occupying the target slot")
	cmd.Flags().String("parent", "", "Parent UUID or name (rack or device)")
	cmd.Flags().String("primary-ipv4", "", "Primary IPv4 address (CIDR or UUID)")
	cmd.Flags().String("primary-ipv6", "", "Primary IPv6 address (CIDR or UUID)")
	cmd.Flags().StringSlice(flagAssignVLAN, nil, "Assign a VLAN (VID or name) to this device (comma-separated or repeatable)")
	cmd.Flags().String(flagBMCOf, "", "Mark this device as the BMC of the given parent device (UUID or name)")

	// Let providers contribute their own device-update flags.
	for _, p := range provider.All() {
		if fp, ok := p.(provider.DeviceUpdateFlagProvider); ok {
			fp.RegisterDeviceUpdateFlags(cmd)
		}
	}

	return cmd
}

func updateDevice(cmd *cli.Command, args []string) error {
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
	if err := applyRelationshipFlags(cmd, inventory, device); err != nil {
		return err
	}
	if err := applySetToDevice(cmd, device); err != nil {
		return err
	}

	return finalizeDeviceUpdate(inventory, id, device)
}

// loadDeviceForUpdate sets the device store, loads the inventory, and resolves
// the target device from args.
func loadDeviceForUpdate(cmd *cli.Command, args []string) (*devicetypes.Inventory, uuid.UUID, *devicetypes.CaniDeviceType, error) {
	if err := store.Setup(cmd); err != nil {
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

// applyRelationshipFlags sets the assigned-VLAN and BMC-parent relationship
// fields from --assign-vlan and --bmc-of.
func applyRelationshipFlags(cmd *cli.Command, inventory *devicetypes.Inventory, device *devicetypes.CaniDeviceType) error {
	if cmd.Flags().Changed(flagAssignVLAN) {
		refs, _ := cmd.Flags().GetStringSlice(flagAssignVLAN)
		vlanIDs, err := resolveVLANRefs(inventory, refs)
		if err != nil {
			return err
		}
		device.AssignedVLANs = vlanIDs
	}
	if cmd.Flags().Changed(flagBMCOf) {
		parentRef, _ := cmd.Flags().GetString(flagBMCOf)
		parentID, err := resolve.Device(inventory, parentRef)
		if err != nil {
			return fmt.Errorf("--bmc-of: %w", err)
		}
		device.BMCParent = parentID
	}
	return nil
}

// resolveVLANRefs resolves VLAN references (VID or name) to VLAN UUIDs.
func resolveVLANRefs(inventory *devicetypes.Inventory, refs []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(refs))
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		if ref == "" {
			continue
		}
		id := findVLANByRef(inventory, ref)
		if id == uuid.Nil {
			return nil, fmt.Errorf("VLAN %q not found", ref)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// findVLANByRef returns the UUID of the VLAN matching ref (VID or name), or Nil.
func findVLANByRef(inventory *devicetypes.Inventory, ref string) uuid.UUID {
	vid, isNumeric := 0, false
	if n, err := strconv.Atoi(ref); err == nil {
		vid, isNumeric = n, true
	}
	for _, v := range inventory.VLANs {
		if v == nil {
			continue
		}
		if isNumeric && v.VID == vid {
			return v.ID
		}
		if v.Name == ref {
			return v.ID
		}
	}
	return uuid.Nil
}

// applyProviderDeviceFlags lets each registered provider apply its own
// device-update flags to the device.
func applyProviderDeviceFlags(cmd *cli.Command, device *devicetypes.CaniDeviceType) error {
	for _, p := range provider.All() {
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
