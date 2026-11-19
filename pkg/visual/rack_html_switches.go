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
package visual

import (
	"sort"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// isSwitchKind reports whether a device gets a per-switch cabling table.
func isSwitchKind(t devicetypes.Type) bool {
	return t == devicetypes.TypeSwitch || t == devicetypes.TypeMgmtSwitch || t == devicetypes.TypeHsnSwitch
}

// buildSwitchTable builds the printable cabling table of one switch in strict
// port order, or nil when the switch has no cable or no patchable port. The
// cabled count and the free list cover the patchable ports; a cabled
// management port is still listed as a row.
func buildSwitchTable(idx *htmlIndex, dev *devicetypes.CaniDeviceType, where string) *htmlSwitchTable {
	rows := switchRows(idx, dev)
	ports := switchPorts(dev)
	if len(rows) == 0 || len(ports) == 0 {
		return nil
	}
	cabled := map[string]bool{}
	for _, row := range rows {
		cabled[row.Port] = true
	}
	var free []string
	cabledPorts := 0
	for _, port := range ports {
		if cabled[port] {
			cabledPorts++
		} else {
			free = append(free, port)
		}
	}
	return &htmlSwitchTable{
		ID: dev.ID.String(), Name: dev.Name, Where: where,
		Cabled: cabledPorts, Total: len(ports), HasVLAN: dev.Type != devicetypes.TypeHsnSwitch,
		Rows: rows, Free: portRanges(free), FreeCount: len(free),
	}
}

// switchRows lists every cable plugged into the switch, one row per local
// port, with the far end and the far end's address.
func switchRows(idx *htmlIndex, dev *devicetypes.CaniDeviceType) []htmlSwitchRow {
	var rows []htmlSwitchRow
	for local, far := range idx.cables {
		if local.dev != dev.ID {
			continue
		}
		spec := dev.GetInterface(local.port)
		peerDev, peerSpec := idx.resolve(far)
		row := htmlSwitchRow{Port: local.port, VLAN: vlanLabel(spec)}
		if spec != nil {
			row.Description = spec.Label
		}
		row.PeerID, row.PeerName, row.PeerPort = peerIdentity(peerDev, far)
		if peerSpec != nil {
			row.PeerIP, row.PeerIPURL = peerAddress(idx, peerDev, peerSpec)
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return naturalLess(rows[i].Port, rows[j].Port) })
	return rows
}

// peerAddress returns the far port's address and, for management ports, its link.
func peerAddress(idx *htmlIndex, dev *devicetypes.CaniDeviceType, spec *devicetypes.InterfaceSpec) (string, string) {
	ip := idx.ipOf(spec.ID)
	if ip == nil {
		return "", ""
	}
	return ipHost(ip), idx.mgmtURL(dev, spec, ip)
}

// switchPorts lists the patchable front ports: management ports, virtual
// interfaces and the parents of breakout ports (1/1/1 when 1/1/1:1 exists)
// are not counted.
func switchPorts(dev *devicetypes.CaniDeviceType) []string {
	names := map[string]bool{}
	for _, spec := range dev.Interfaces {
		names[spec.Name] = true
	}
	var ports []string
	for i := range dev.Interfaces {
		spec := &dev.Interfaces[i]
		if isMgmtOnly(spec) || strings.EqualFold(spec.Name, "mgmt") || names[spec.Name+":1"] {
			continue
		}
		if spec.Type == devicetypes.InterfacesElemTypeVirtual || spec.Type == devicetypes.InterfacesElemTypeLag {
			continue
		}
		ports = append(ports, spec.Name)
	}
	sort.Slice(ports, func(i, j int) bool { return naturalLess(ports[i], ports[j]) })
	return ports
}

// portRanges collapses sorted port names into "1/1/25–1/1/48, 1/1/51–1/1/52".
func portRanges(ports []string) string {
	var ranges []string
	for i := 0; i < len(ports); {
		j := i
		for j+1 < len(ports) && consecutivePorts(ports[j], ports[j+1]) {
			j++
		}
		if i == j {
			ranges = append(ranges, ports[i])
		} else {
			ranges = append(ranges, ports[i]+"\u2013"+ports[j])
		}
		i = j + 1
	}
	return strings.Join(ranges, ", ")
}

// consecutivePorts reports whether b is the next port after a in the same
// group: same prefix and a trailing number one higher.
func consecutivePorts(a, b string) bool {
	prefixA, numA, okA := splitTrailingInt(a)
	prefixB, numB, okB := splitTrailingInt(b)
	return okA && okB && prefixA == prefixB && numB == numA+1
}

// splitTrailingInt splits "1/1/25" into "1/1/" and 25.
func splitTrailingInt(s string) (string, int, bool) {
	end := len(s)
	start := end
	for start > 0 && isDigit(s[start-1]) {
		start--
	}
	if start == end {
		return s, 0, false
	}
	n, _ := leadingInt(s[start:])
	return s[:start], n, true
}

// naturalLess orders strings with embedded numbers numerically, so that
// 1/1/1:3 < 1/1/1:4 < 1/1/2 < 1/1/49 and DIMM 2 < DIMM 10.
func naturalLess(a, b string) bool {
	for a != "" && b != "" {
		if isDigit(a[0]) && isDigit(b[0]) {
			na, restA := leadingInt(a)
			nb, restB := leadingInt(b)
			if na != nb {
				return na < nb
			}
			a, b = restA, restB
			continue
		}
		if a[0] != b[0] {
			return a[0] < b[0]
		}
		a, b = a[1:], b[1:]
	}
	return len(a) < len(b)
}

// leadingInt parses the digit run at the start of s and returns the rest.
func leadingInt(s string) (int, string) {
	n, i := 0, 0
	for i < len(s) && isDigit(s[i]) {
		n = n*10 + int(s[i]-'0')
		i++
	}
	return n, s[i:]
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}
