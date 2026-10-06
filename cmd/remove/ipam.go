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
package remove

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/util/store"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

const (
	flagNamespace = "namespace"
	flagCascade   = "cascade"
)

// newPrefixCommand creates the "remove prefix" subcommand.
func newPrefixCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "prefix <uuid | [namespace/]cidr>",
		Short: "Remove a prefix from the inventory.",
		Long: `Remove a prefix by UUID or CIDR. A CIDR is searched in --namespace
(default Global) unless written as "namespace/CIDR".

The prefix's child prefixes and IP addresses move up to its parent prefix, as
in Nautobot. A prefix without a parent that still holds IP addresses is kept
unless --cascade is given, which removes those addresses with it.

Examples:
  cani alpha remove prefix 10.0.0.0/24
  cani alpha remove prefix tenant-a/10.0.0.0/24 --cascade`,
		Args: cli.ExactArgs(1),
		RunE: removePrefix,
	}
	cmd.Flags().String(flagNamespace, devicetypes.DefaultNamespace, "IPAM namespace to search when the CIDR is unqualified")
	cmd.Flags().Bool(flagCascade, false, "Also remove the IP addresses held by a prefix that has no parent")
	return cmd
}

// newIPCommand creates the "remove ip" subcommand.
func newIPCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "ip <uuid | [namespace/]address>",
		Short: "Remove an IP address from the inventory.",
		Long: `Remove an IP address by UUID or address, with or without its mask. An
address is searched in --namespace (default Global) unless written as
"namespace/address". Device primary IPs, NAT inside references and interface
assignments that point at it are cleared.

Examples:
  cani alpha remove ip 10.0.0.5
  cani alpha remove ip tenant-a/10.0.0.5/24`,
		Args: cli.ExactArgs(1),
		RunE: removeIP,
	}
	cmd.Flags().String(flagNamespace, devicetypes.DefaultNamespace, "IPAM namespace to search when the address is unqualified")
	return cmd
}

// newVRFCommand creates the "remove vrf" subcommand.
func newVRFCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "vrf <uuid | [namespace/]name>",
		Short: "Remove a VRF from the inventory.",
		Long: `Remove a VRF by UUID or name. A name is searched in --namespace (default
Global) unless written as "namespace/name". Prefixes in the VRF lose that
membership; interfaces that name the VRF keep the name, which export then
resolves against the remaining VRFs or skips, so they are listed.

Examples:
  cani alpha remove vrf keepalive
  cani alpha remove vrf tenant-a/blue`,
		Args: cli.ExactArgs(1),
		RunE: removeVRF,
	}
	cmd.Flags().String(flagNamespace, devicetypes.DefaultNamespace, "IPAM namespace to search when the name is unqualified")
	return cmd
}

// removeIPAM loads the inventory, lets remove resolve and delete one record
// using the --namespace value, saves, and logs what remove reports.
func removeIPAM(cmd *cli.Command, remove func(*devicetypes.Inventory, string) (string, error)) error {
	if err := store.Setup(cmd); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}
	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}
	namespace, _ := cmd.Flags().GetString(flagNamespace)
	removed, err := remove(inventory, namespace)
	if err != nil {
		return err
	}
	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}
	log.Printf("Removed %s", removed)
	return nil
}

func removePrefix(cmd *cli.Command, args []string) error {
	cascade, _ := cmd.Flags().GetBool(flagCascade)
	return removeIPAM(cmd, func(inventory *devicetypes.Inventory, namespace string) (string, error) {
		prefix, err := inventory.ResolvePrefixReference(namespace, args[0])
		if err != nil {
			return "", err
		}
		removed := fmt.Sprintf("prefix %s in namespace %s (%s)", prefix.Prefix, prefix.EffectiveNamespace(), prefix.ID)
		if !cascade {
			return removed, prefixRemovalError(inventory.RemovePrefix(prefix.ID))
		}
		addresses, err := inventory.RemovePrefixWithAddresses(prefix.ID)
		if err != nil {
			return "", prefixRemovalError(err)
		}
		return fmt.Sprintf("%s and %d IP address(es) it held", removed, addresses), nil
	})
}

// prefixRemovalError names the --cascade override when the model refuses to
// leave addresses without a parent prefix.
func prefixRemovalError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, devicetypes.ErrPrefixHoldsAddresses):
		return fmt.Errorf("failed to remove prefix: %w, or rerun with --%s to remove them with it", err, flagCascade)
	default:
		return fmt.Errorf("failed to remove prefix: %w", err)
	}
}

func removeIP(cmd *cli.Command, args []string) error {
	return removeIPAM(cmd, func(inventory *devicetypes.Inventory, namespace string) (string, error) {
		address, err := inventory.ResolveIPAddressReference(namespace, args[0])
		if err != nil {
			return "", err
		}
		removed := fmt.Sprintf("IP address %s in namespace %s (%s)",
			address.Address, inventory.IPAddressNamespace(address), address.ID)
		if err := inventory.RemoveIPAddress(address.ID); err != nil {
			return "", fmt.Errorf("failed to remove IP address: %w", err)
		}
		return removed, nil
	})
}

func removeVRF(cmd *cli.Command, args []string) error {
	return removeIPAM(cmd, func(inventory *devicetypes.Inventory, namespace string) (string, error) {
		vrf, err := inventory.ResolveVRFReference(namespace, args[0])
		if err != nil {
			return "", err
		}
		named := inventory.InterfacesNamingVRF(vrf)
		if err := inventory.RemoveVRF(vrf.ID); err != nil {
			return "", fmt.Errorf("failed to remove VRF: %w", err)
		}
		removed := fmt.Sprintf("VRF %q in namespace %s (%s)", vrf.Name, vrf.EffectiveNamespace(), vrf.ID)
		if len(named) > 0 {
			removed += fmt.Sprintf("; %d interface(s) still name it (%s) and export resolves or skips the name",
				len(named), interfaceLabels(inventory, named))
		}
		return removed, nil
	})
}

// interfaceLabels renders interfaces as device/name, comma-separated.
func interfaceLabels(inventory *devicetypes.Inventory, ifaces []*devicetypes.CaniInterface) string {
	labels := make([]string, 0, len(ifaces))
	for _, iface := range ifaces {
		owner := iface.DeviceID.String()
		if device := inventory.Devices[iface.DeviceID]; device != nil {
			owner = device.Name
		}
		labels = append(labels, owner+"/"+iface.Name)
	}
	return strings.Join(labels, ", ")
}
