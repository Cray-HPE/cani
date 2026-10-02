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
package show

import (
	"sort"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/visual"
)

// flagNamespace limits show prefix, ip and vrf to one IPAM namespace; empty
// means every namespace.
const flagNamespace = "namespace"

// addNamespaceFlag registers the --namespace filter on an IPAM show command.
func addNamespaceFlag(cmd *cli.Command) *cli.Command {
	cmd.Flags().String(flagNamespace, "", "Show only objects in this IPAM namespace")
	return cmd
}

// inNamespace reports whether got matches the --namespace filter: no filter
// admits every namespace.
func inNamespace(cmd *cli.Command, got string) bool {
	want, _ := cmd.Flags().GetString(flagNamespace)
	return want == "" || devicetypes.NamespaceOrDefault(want) == got
}

// newVRFShowCommand creates the "show vrf" subcommand.
func newVRFShowCommand() *cli.Command {
	return addNamespaceFlag(&cli.Command{
		Use:     "vrf",
		Aliases: []string{"vrfs"},
		Short:   "List VRFs in the inventory.",
		Args:    cli.NoArgs,
		RunE:    showVRFs,
	})
}

func showVRFs(cmd *cli.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
	}

	vrfs := make([]*devicetypes.CaniVRF, 0, len(inv.VRFs))
	for _, v := range inv.VRFs {
		if inNamespace(cmd, v.EffectiveNamespace()) {
			vrfs = append(vrfs, v)
		}
	}
	sort.Slice(vrfs, func(i, j int) bool {
		return vrfs[i].Name < vrfs[j].Name
	})

	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		visual.PrintVRFTable(vrfs, inv)
		return nil
	default:
		return marshalAndPrint(vrfs)
	}
}

// newVLANShowCommand creates the "show vlan" subcommand.
func newVLANShowCommand() *cli.Command {
	return &cli.Command{
		Use:     "vlan",
		Aliases: []string{"vlans"},
		Short:   "List VLANs in the inventory.",
		Args:    cli.NoArgs,
		RunE:    showVLANs,
	}
}

func showVLANs(cmd *cli.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
	}

	vlans := make([]*devicetypes.CaniVLAN, 0, len(inv.VLANs))
	for _, v := range inv.VLANs {
		vlans = append(vlans, v)
	}
	sort.Slice(vlans, func(i, j int) bool {
		return vlans[i].VID < vlans[j].VID
	})

	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		visual.PrintVLANTable(vlans, inv)
		return nil
	default:
		return marshalAndPrint(vlans)
	}
}

// newPrefixShowCommand creates the "show prefix" subcommand.
func newPrefixShowCommand() *cli.Command {
	cmd := &cli.Command{
		Use:     "prefix",
		Aliases: []string{"prefixes"},
		Short:   "List IP prefixes in the inventory.",
		Args:    cli.NoArgs,
		RunE:    showPrefixes,
	}
	cmd.Flags().Bool("tree", false, "Display prefixes as a hierarchy tree")
	return addNamespaceFlag(cmd)
}

func showPrefixes(cmd *cli.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
	}

	prefixes := make([]*devicetypes.CaniPrefix, 0, len(inv.Prefixes))
	for _, p := range inv.Prefixes {
		if inNamespace(cmd, p.EffectiveNamespace()) {
			prefixes = append(prefixes, p)
		}
	}
	sort.Slice(prefixes, func(i, j int) bool {
		return prefixes[i].Prefix < prefixes[j].Prefix
	})

	tree, _ := cmd.Flags().GetBool("tree")
	if tree {
		visual.PrintPrefixTree(prefixes)
		return nil
	}

	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		visual.PrintPrefixTable(prefixes, inv)
		return nil
	default:
		return marshalAndPrint(prefixes)
	}
}

// newIPShowCommand creates the "show ip" subcommand.
func newIPShowCommand() *cli.Command {
	cmd := &cli.Command{
		Use:     "ip",
		Aliases: []string{"ips", "ipaddress"},
		Short:   "List IP addresses in the inventory.",
		Args:    cli.NoArgs,
		RunE:    showIPAddresses,
	}
	cmd.Flags().String("prefix", "", "Filter to IPs within a specific prefix (CIDR)")
	return addNamespaceFlag(cmd)
}

func showIPAddresses(cmd *cli.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
	}

	addrs := selectIPAddresses(cmd, inv)
	sort.Slice(addrs, func(i, j int) bool {
		return addrs[i].Address < addrs[j].Address
	})

	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		visual.PrintIPAddressTable(addrs, inv)
		return nil
	default:
		return marshalAndPrint(addrs)
	}
}

// selectIPAddresses returns the addresses that pass the --prefix and
// --namespace filters.
func selectIPAddresses(cmd *cli.Command, inv *devicetypes.Inventory) []*devicetypes.CaniIPAddress {
	prefixFilter, _ := cmd.Flags().GetString("prefix")
	addrs := make([]*devicetypes.CaniIPAddress, 0, len(inv.IPAddresses))
	for _, a := range inv.IPAddresses {
		if prefixFilter != "" && findPrefixCIDR(a.Parent, inv) != prefixFilter {
			continue
		}
		if inNamespace(cmd, inv.IPAddressNamespace(a)) {
			addrs = append(addrs, a)
		}
	}
	return addrs
}

func findPrefixCIDR(id [16]byte, inv *devicetypes.Inventory) string {
	if id == [16]byte{} {
		return ""
	}
	if p, ok := inv.Prefixes[id]; ok {
		return p.Prefix
	}
	return ""
}
