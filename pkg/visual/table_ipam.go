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
package visual

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// Column widths for IPAM tables.
const (
	ColPrefix  = 20
	ColVID     = 6
	ColAddress = 20
	ColDNS     = 30
)

// PrintVLANTable renders VLANs as a fixed-width table.
func PrintVLANTable(vlans []*devicetypes.CaniVLAN, inv *devicetypes.Inventory) {
	header := Col("VID", ColVID) + "  " +
		Col("NAME", ColName) + "  " +
		Col("STATUS", ColStatus) + "  " +
		Col("LOCATION", ColLocation)
	sep := Col(strings.Repeat("-", ColVID), ColVID) + "  " +
		Col(strings.Repeat("-", ColName), ColName) + "  " +
		Col(strings.Repeat("-", ColStatus), ColStatus) + "  " +
		Col(strings.Repeat("-", ColLocation), ColLocation)

	fmt.Println(header)
	fmt.Println(sep)
	for _, v := range vlans {
		locName := ResolveLocationName(v.Location, inv)
		fmt.Println(
			Col(strconv.Itoa(v.VID), ColVID) + "  " +
				Col(v.Name, ColName) + "  " +
				Col(v.Status, ColStatus) + "  " +
				Col(locName, ColLocation),
		)
	}
	fmt.Printf("\nTotal: %d VLAN(s)\n", len(vlans))
}

// PrintPrefixTable renders prefixes as a fixed-width table.
func PrintPrefixTable(prefixes []*devicetypes.CaniPrefix, inv *devicetypes.Inventory) {
	header := Col("PREFIX", ColPrefix) + "  " +
		Col("TYPE", ColType) + "  " +
		Col("ROLE", ColRole) + "  " +
		Col("STATUS", ColStatus) + "  " +
		Col("VLAN", ColCount)
	sep := Col(strings.Repeat("-", ColPrefix), ColPrefix) + "  " +
		Col(strings.Repeat("-", ColType), ColType) + "  " +
		Col(strings.Repeat("-", ColRole), ColRole) + "  " +
		Col(strings.Repeat("-", ColStatus), ColStatus) + "  " +
		Col(strings.Repeat("-", ColCount), ColCount)

	fmt.Println(header)
	fmt.Println(sep)
	for _, p := range prefixes {
		vlanName := resolveVLANName(p.VLAN, inv)
		fmt.Println(
			Col(p.Prefix, ColPrefix) + "  " +
				Col(string(p.Type), ColType) + "  " +
				Col(p.Role, ColRole) + "  " +
				Col(p.Status, ColStatus) + "  " +
				Col(vlanName, ColCount),
		)
	}
	fmt.Printf("\nTotal: %d prefix(es)\n", len(prefixes))
}

// PrintPrefixTree renders prefixes as an indented hierarchy.
func PrintPrefixTree(prefixes []*devicetypes.CaniPrefix) {
	// Build parent-child map
	roots := make([]*devicetypes.CaniPrefix, 0)
	children := make(map[string][]*devicetypes.CaniPrefix)
	for _, p := range prefixes {
		if p.Parent.String() == "00000000-0000-0000-0000-000000000000" {
			roots = append(roots, p)
		} else {
			children[p.Parent.String()] = append(children[p.Parent.String()], p)
		}
	}
	for _, root := range roots {
		printPrefixNode(root, children, "", true)
	}
	fmt.Printf("\nTotal: %d prefix(es)\n", len(prefixes))
}

func printPrefixNode(p *devicetypes.CaniPrefix, children map[string][]*devicetypes.CaniPrefix, indent string, last bool) {
	connector := "├── "
	if last {
		connector = "└── "
	}
	label := p.Prefix
	if string(p.Type) != "" {
		label += " [" + string(p.Type) + "]"
	}
	if p.Role != "" {
		label += " (" + p.Role + ")"
	}
	fmt.Println(indent + connector + label)

	childIndent := indent + "│   "
	if last {
		childIndent = indent + "    "
	}
	kids := children[p.ID.String()]
	for i, kid := range kids {
		printPrefixNode(kid, children, childIndent, i == len(kids)-1)
	}
}

// PrintIPAddressTable renders IP addresses as a fixed-width table.
func PrintIPAddressTable(addrs []*devicetypes.CaniIPAddress, inv *devicetypes.Inventory) {
	header := Col("ADDRESS", ColAddress) + "  " +
		Col("STATUS", ColStatus) + "  " +
		Col("ROLE", ColRole) + "  " +
		Col("DNS NAME", ColDNS) + "  " +
		Col("PREFIX", ColPrefix)
	sep := Col(strings.Repeat("-", ColAddress), ColAddress) + "  " +
		Col(strings.Repeat("-", ColStatus), ColStatus) + "  " +
		Col(strings.Repeat("-", ColRole), ColRole) + "  " +
		Col(strings.Repeat("-", ColDNS), ColDNS) + "  " +
		Col(strings.Repeat("-", ColPrefix), ColPrefix)

	fmt.Println(header)
	fmt.Println(sep)
	for _, a := range addrs {
		parentPrefix := resolvePrefixCIDR(a.Parent, inv)
		role := string(a.IPRole)
		fmt.Println(
			Col(a.Address, ColAddress) + "  " +
				Col(a.Status, ColStatus) + "  " +
				Col(role, ColRole) + "  " +
				Col(a.DNSName, ColDNS) + "  " +
				Col(parentPrefix, ColPrefix),
		)
	}
	fmt.Printf("\nTotal: %d IP address(es)\n", len(addrs))
}

func resolveVLANName(id [16]byte, inv *devicetypes.Inventory) string {
	if id == [16]byte{} || inv == nil || inv.VLANs == nil {
		return ""
	}
	if v, ok := inv.VLANs[id]; ok {
		return v.Name
	}
	return ""
}

func resolvePrefixCIDR(id [16]byte, inv *devicetypes.Inventory) string {
	if id == [16]byte{} || inv == nil || inv.Prefixes == nil {
		return ""
	}
	if p, ok := inv.Prefixes[id]; ok {
		return p.Prefix
	}
	return ""
}
