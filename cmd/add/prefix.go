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

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/util/store"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func newPrefixCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "prefix <cidr>",
		Short: "Add an IP prefix (subnet) to the inventory.",
		Long: `Add an IP prefix to the inventory in CIDR notation.

Examples:
  cani alpha add prefix 10.0.0.0/16 --type container --role infrastructure
  cani alpha add prefix 10.0.1.0/24 --type network --role management --vlan "Management"
  cani alpha add prefix 10.0.1.128/25 --type pool --role dhcp-pool`,
		Args: cli.ExactArgs(1),
		RunE: addPrefix,
	}

	cmd.Flags().String("type", "", "Prefix type: container, network, or pool")
	cmd.Flags().String("role", "", "Prefix role (e.g. management, bmc, infrastructure)")
	cmd.Flags().String("vlan", "", "Associated VLAN name or UUID")
	cmd.Flags().String("vrf", "", "VRF name")
	cmd.Flags().String(flagLocation, "", "Location UUID or name")
	cmd.Flags().String(flagDescription, "", "Prefix description")

	return cmd
}

func addPrefix(cmd *cli.Command, args []string) error {
	cidr := args[0]

	if err := store.Setup(cmd); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	prefix := &devicetypes.CaniPrefix{
		ID:     uuid.New(),
		Prefix: cidr,
	}

	if cmd.Flags().Changed("type") {
		t, _ := cmd.Flags().GetString("type")
		prefix.Type = devicetypes.PrefixType(t)
	}
	if cmd.Flags().Changed(flagDescription) {
		prefix.Description, _ = cmd.Flags().GetString(flagDescription)
	}
	if cmd.Flags().Changed("vrf") {
		prefix.VRF, _ = cmd.Flags().GetString("vrf")
	}
	if cmd.Flags().Changed(flagLocation) {
		locationArg, _ := cmd.Flags().GetString(flagLocation)
		prefix.Location = resolveLocation(inventory, locationArg)
	}
	if cmd.Flags().Changed("vlan") {
		vlanArg, _ := cmd.Flags().GetString("vlan")
		prefix.VLAN = resolveVLAN(inventory, vlanArg)
	}
	if cmd.Flags().Changed("status") {
		prefix.Status, _ = cmd.Flags().GetString("status")
	}
	if cmd.Flags().Changed("role") {
		prefix.Role, _ = cmd.Flags().GetString("role")
	}
	tags, _ := cmd.Flags().GetStringArray("tag")
	if len(tags) > 0 {
		prefix.Tags = tags
	}

	if err := inventory.AddPrefix(prefix); err != nil {
		return fmt.Errorf("failed to add prefix: %w", err)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("Added prefix %s (%s)", prefix.Prefix, prefix.ID)
	return nil
}

// resolveVLAN looks up a VLAN by name or UUID string. Returns uuid.Nil if not found.
func resolveVLAN(inv *devicetypes.Inventory, ref string) uuid.UUID {
	// Try as UUID first
	if id, err := uuid.Parse(ref); err == nil {
		if _, ok := inv.VLANs[id]; ok {
			return id
		}
	}
	// Try as name
	for _, v := range inv.VLANs {
		if v.Name == ref {
			return v.ID
		}
	}
	return uuid.Nil
}
