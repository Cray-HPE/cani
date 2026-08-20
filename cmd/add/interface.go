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
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/util/resolve"
	"github.com/Cray-HPE/cani/internal/util/store"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func newInterfaceCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "interface <name>",
		Short: "Add an interface to a device.",
		Long: `Add a standalone interface (e.g. a LAG) to an existing device.

Use this to create interfaces that are not part of a device-type template,
such as LAG interfaces used to aggregate physical members.

After creating a LAG, assign members with:
  cani update interface --device <device> --name <member> --lag <lag-name>

Examples:
  cani alpha add interface lag256 --device switch-01 --type lag
  cani alpha add interface lag256 --device switch-01 --type lag --mode tagged --tagged-vlan 100,200
  cani alpha add interface lag256 --device switch-01 --type lag --mode access --untagged-vlan 10`,
		Args: cli.ExactArgs(1),
		RunE: addInterface,
	}

	cmd.Flags().String("device", "", "Device name or UUID (required)")
	cmd.Flags().String("type", "lag", "Interface type (e.g. lag, virtual, 1000base-t)")
	cmd.Flags().String("role", "", "Interface role")
	cmd.Flags().String("label", "", "Interface label")
	cmd.Flags().String("mac", "", "MAC address (e.g. aa:bb:cc:dd:ee:ff)")
	cmd.Flags().String("mode", "", "802.1Q mode: access, tagged, or tagged-all")
	cmd.Flags().Int("untagged-vlan", 0, "Untagged (native) VLAN ID")
	cmd.Flags().StringSlice("tagged-vlan", nil, "Tagged VLAN IDs (comma-separated or repeatable)")
	cmd.Flags().String("vrf", "", "VRF name")
	cmd.Flags().String(flagDescription, "", "Interface description")

	return cmd
}

func addInterface(cmd *cli.Command, args []string) error {
	name := args[0]

	deviceRef, _ := cmd.Flags().GetString("device")
	if deviceRef == "" {
		return fmt.Errorf("--device is required")
	}

	ifaceType, _ := cmd.Flags().GetString("type")
	if !devicetypes.IsValidInterfaceType(ifaceType) {
		return fmt.Errorf("unknown interface type %q", ifaceType)
	}

	if err := store.Setup(cmd); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	deviceID, err := resolve.Device(inventory, deviceRef)
	if err != nil {
		return fmt.Errorf("resolving --device: %w", err)
	}

	iface := &devicetypes.CaniInterface{
		ID:            uuid.New(),
		Name:          name,
		InterfaceType: devicetypes.InterfacesElemType(ifaceType),
		DeviceID:      deviceID,
		ContentType:   "dcim.interface",
	}

	if cmd.Flags().Changed("role") {
		iface.Role, _ = cmd.Flags().GetString("role")
	}
	if cmd.Flags().Changed("label") {
		iface.Label, _ = cmd.Flags().GetString("label")
	}
	if cmd.Flags().Changed("mac") {
		mac, _ := cmd.Flags().GetString("mac")
		normalized, merr := devicetypes.NormalizeMAC(mac)
		if merr != nil {
			return merr
		}
		iface.MacAddress = normalized
	}
	if cmd.Flags().Changed("mode") {
		mode, _ := cmd.Flags().GetString("mode")
		normalized, merr := devicetypes.ValidateInterfaceMode(mode)
		if merr != nil {
			return merr
		}
		iface.Mode = normalized
	}
	if cmd.Flags().Changed("untagged-vlan") {
		vid, _ := cmd.Flags().GetInt("untagged-vlan")
		if err := devicetypes.ValidateVID(vid); err != nil {
			return err
		}
		iface.UntaggedVLAN = vid
	}
	if cmd.Flags().Changed("tagged-vlan") {
		taggedStrs, _ := cmd.Flags().GetStringSlice("tagged-vlan")
		vids, verr := parseTaggedVLANs(taggedStrs)
		if verr != nil {
			return verr
		}
		iface.TaggedVLANs = vids
	}
	if cmd.Flags().Changed("vrf") {
		iface.VRF, _ = cmd.Flags().GetString("vrf")
	}
	if cmd.Flags().Changed(flagDescription) {
		iface.Description, _ = cmd.Flags().GetString(flagDescription)
	}
	if cmd.Flags().Changed("status") {
		iface.Status, _ = cmd.Flags().GetString("status")
	}
	tags, _ := cmd.Flags().GetStringArray("tag")
	if len(tags) > 0 {
		iface.Tags = tags
	}
	if meta := collectProviderMetadata(cmd); len(meta) > 0 {
		applyProviderMetadataMap(&iface.ProviderMetadata, meta)
	}

	if err := inventory.AddInterface(iface); err != nil {
		return fmt.Errorf("failed to add interface: %w", err)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("Added interface %q (type=%s) on device %s (%s)", iface.Name, ifaceType, deviceRef, iface.ID)
	return nil
}

// parseTaggedVLANs converts comma-separated or repeated VLAN ID strings to ints.
func parseTaggedVLANs(strs []string) ([]int, error) {
	var vids []int
	for _, s := range strs {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		vid, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("invalid tagged VLAN %q: must be an integer", s)
		}
		if err := devicetypes.ValidateVID(vid); err != nil {
			return nil, err
		}
		vids = append(vids, vid)
	}
	return vids, nil
}
