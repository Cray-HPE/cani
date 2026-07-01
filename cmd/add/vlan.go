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
	"strconv"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/util/store"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func newVLANCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "vlan <vid>",
		Short: "Add a VLAN to the inventory.",
		Long: `Add a VLAN (layer-2 domain) to the inventory.

Examples:
  cani alpha add vlan 100 --name "Management" --status active
  cani alpha add vlan 200 --name "BMC" --status active --location "Zone-A"`,
		Args: cli.ExactArgs(1),
		RunE: addVLAN,
	}

	cmd.Flags().String("name", "", "VLAN name (required)")
	cmd.Flags().String("role", "", "VLAN role (e.g. AfcTransitVlan)")
	cmd.Flags().String(flagLocation, "", "Location UUID or name")
	cmd.Flags().String(flagDescription, "", "VLAN description")

	return cmd
}

func addVLAN(cmd *cli.Command, args []string) error {
	vid, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid VLAN ID %q: must be an integer", args[0])
	}
	if vid < 1 || vid > 4094 {
		return fmt.Errorf("VLAN ID must be between 1 and 4094, got %d", vid)
	}

	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		return fmt.Errorf("--name is required")
	}

	if err := store.Setup(cmd); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	vlan := &devicetypes.CaniVLAN{
		ID:   uuid.New(),
		VID:  vid,
		Name: name,
	}

	if cmd.Flags().Changed(flagDescription) {
		vlan.Description, _ = cmd.Flags().GetString(flagDescription)
	}
	if cmd.Flags().Changed(flagLocation) {
		locationArg, _ := cmd.Flags().GetString(flagLocation)
		vlan.Location = resolveLocation(inventory, locationArg)
	}
	if cmd.Flags().Changed("status") {
		vlan.Status, _ = cmd.Flags().GetString("status")
	}
	if cmd.Flags().Changed("role") {
		vlan.Role, _ = cmd.Flags().GetString("role")
	}
	tags, _ := cmd.Flags().GetStringArray("tag")
	if len(tags) > 0 {
		vlan.Tags = tags
	}

	if err := inventory.AddVLAN(vlan); err != nil {
		return fmt.Errorf("failed to add vlan: %w", err)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("Added VLAN %d %q (%s)", vlan.VID, vlan.Name, vlan.ID)
	return nil
}
