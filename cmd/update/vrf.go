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
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// newVRFCommand creates the "update vrf" subcommand.
func newVRFCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "vrf <name-or-uuid>",
		Short: "Update a VRF in the inventory.",
		Long: `Update a VRF's fields or add devices to it.

The VRF is matched by UUID, by name within --namespace (default Global), or
by a qualified "namespace/name".

Examples:
  cani alpha update vrf keepalive --add-device sw-spine-02
  cani alpha update vrf LEGACY --description "Updated description"
  cani alpha update vrf keepalive --remove-device sw-leaf-01
  cani alpha update vrf tenant-a/blue --rd 65000:10`,
		Args: cli.ExactArgs(1),
		RunE: updateVRF,
	}

	cmd.Flags().String(flagNamespace, devicetypes.DefaultNamespace, "IPAM namespace to search when the name is unqualified")
	cmd.Flags().String(flagDescription, "", "New description")
	cmd.Flags().String("rd", "", "Route distinguisher")
	cmd.Flags().StringArray(flagAddDevice, nil, "Add a device to this VRF (name or UUID, repeatable)")
	cmd.Flags().StringArray(flagRemoveDevice, nil, "Remove a device from this VRF (name or UUID, repeatable)")

	return cmd
}

func updateVRF(cmd *cli.Command, args []string) error {
	if err := store.Setup(cmd); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}
	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	namespace, _ := cmd.Flags().GetString(flagNamespace)
	vrf, err := inventory.ResolveVRFReference(namespace, args[0])
	if err != nil {
		return err
	}
	changed, err := applyVRFUpdates(cmd, inventory, vrf)
	if err != nil {
		return err
	}
	if !changed {
		return fmt.Errorf("no changes specified")
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}
	log.Printf("Updated VRF %q in namespace %s (%s)", vrf.Name, vrf.EffectiveNamespace(), vrf.ID)
	return nil
}

// applyVRFUpdates applies the field and device flags and reports whether
// anything changed.
func applyVRFUpdates(cmd *cli.Command, inventory *devicetypes.Inventory, vrf *devicetypes.CaniVRF) (bool, error) {
	changed := applyVRFFields(cmd, vrf)
	added, err := addVRFDevices(cmd, inventory, vrf)
	if err != nil {
		return false, err
	}
	removed, err := removeVRFDevices(cmd, inventory, vrf)
	if err != nil {
		return false, err
	}
	return changed || added || removed, nil
}

// applyVRFFields copies the description and RD flags when set.
func applyVRFFields(cmd *cli.Command, vrf *devicetypes.CaniVRF) bool {
	changed := false
	if cmd.Flags().Changed(flagDescription) {
		vrf.Description, _ = cmd.Flags().GetString(flagDescription)
		changed = true
	}
	if cmd.Flags().Changed("rd") {
		vrf.RD, _ = cmd.Flags().GetString("rd")
		changed = true
	}
	return changed
}

// addVRFDevices assigns each --add-device to the VRF through the model.
func addVRFDevices(cmd *cli.Command, inventory *devicetypes.Inventory, vrf *devicetypes.CaniVRF) (bool, error) {
	refs, _ := cmd.Flags().GetStringArray(flagAddDevice)
	changed := false
	for _, ref := range refs {
		id, err := resolve.Device(inventory, ref)
		if err != nil {
			return false, fmt.Errorf("resolving --%s %q: %w", flagAddDevice, ref, err)
		}
		added, err := inventory.AddVRFDevice(vrf, id)
		if err != nil {
			return false, err
		}
		changed = changed || added
	}
	return changed, nil
}

// removeVRFDevices drops each --remove-device assignment through the model.
func removeVRFDevices(cmd *cli.Command, inventory *devicetypes.Inventory, vrf *devicetypes.CaniVRF) (bool, error) {
	refs, _ := cmd.Flags().GetStringArray(flagRemoveDevice)
	changed := false
	for _, ref := range refs {
		id, err := resolve.Device(inventory, ref)
		if err != nil {
			return false, fmt.Errorf("resolving --%s %q: %w", flagRemoveDevice, ref, err)
		}
		changed = vrf.RemoveDevice(id) || changed
	}
	return changed, nil
}
