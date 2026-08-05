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
package add

import (
	"fmt"
	"log"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/util/resolve"
	"github.com/Cray-HPE/cani/internal/util/store"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// newVRFCommand creates the "add vrf" subcommand.
func newVRFCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "vrf <name>",
		Short: "Add a VRF to the inventory.",
		Long: `Add a VRF (virtual routing and forwarding instance) to the inventory.

Examples:
  cani alpha add vrf LEGACY --rd 65000:2000 --description "Legacy CSM CMN routing"
  cani alpha add vrf keepalive --description "VSX keepalive" --device sw-leaf-01 --device sw-spine-01`,
		Args: cli.ExactArgs(1),
		RunE: addVRF,
	}

	cmd.Flags().String("rd", "", "Route distinguisher (RFC 4364, e.g. 65000:100)")
	cmd.Flags().String("namespace", "", "IPAM namespace (defaults to Global)")
	cmd.Flags().String(flagDescription, "", "VRF description")
	cmd.Flags().StringArray("device", nil, "Device name or UUID to assign this VRF to (repeatable)")

	return cmd
}

func addVRF(cmd *cli.Command, args []string) error {
	name := args[0]
	if name == "" {
		return fmt.Errorf("VRF name is required")
	}

	if err := store.Setup(cmd); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	vrf := &devicetypes.CaniVRF{
		ID:   uuid.New(),
		Name: name,
	}
	applyVRFFlags(cmd, vrf)

	if err := resolveVRFDevices(cmd, inventory, vrf); err != nil {
		return err
	}

	if err := inventory.AddVRF(vrf); err != nil {
		return fmt.Errorf("failed to add vrf: %w", err)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("Added VRF %q (%s)", vrf.Name, vrf.ID)
	return nil
}

func applyVRFFlags(cmd *cli.Command, vrf *devicetypes.CaniVRF) {
	if cmd.Flags().Changed("rd") {
		vrf.RD, _ = cmd.Flags().GetString("rd")
	}
	if cmd.Flags().Changed("namespace") {
		vrf.Namespace, _ = cmd.Flags().GetString("namespace")
	}
	if cmd.Flags().Changed(flagDescription) {
		vrf.Description, _ = cmd.Flags().GetString(flagDescription)
	}
	if cmd.Flags().Changed("status") {
		vrf.Status, _ = cmd.Flags().GetString("status")
	}
	tags, _ := cmd.Flags().GetStringArray("tag")
	if len(tags) > 0 {
		vrf.Tags = tags
	}
}

func resolveVRFDevices(cmd *cli.Command, inventory *devicetypes.Inventory, vrf *devicetypes.CaniVRF) error {
	deviceRefs, _ := cmd.Flags().GetStringArray("device")
	for _, ref := range deviceRefs {
		id, err := resolve.Device(inventory, ref)
		if err != nil {
			return fmt.Errorf("resolving --device %q: %w", ref, err)
		}
		vrf.Devices = append(vrf.Devices, id)
	}
	return nil
}
