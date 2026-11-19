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
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// portRef names one cable end: a device and one of its port names.
type portRef struct {
	dev  uuid.UUID
	port string
}

// htmlIndex holds the cable and IP joins shared by every card of the page.
// Cables are joined on device + port name because InterfaceSpec.ConnectedCable
// is not populated by the loaders; interface UUIDs are the fallback.
type htmlIndex struct {
	inv    *devicetypes.Inventory
	pairs  [][2]portRef
	cables map[portRef]portRef
	ips    map[uuid.UUID][]*devicetypes.CaniIPAddress
}

// newHTMLIndex joins every cable to both of its ends and every IP to its interfaces.
func newHTMLIndex(inv *devicetypes.Inventory) *htmlIndex {
	idx := &htmlIndex{inv: inv, cables: map[portRef]portRef{}, ips: map[uuid.UUID][]*devicetypes.CaniIPAddress{}}
	for _, cable := range inv.Cables {
		if cable == nil {
			continue
		}
		a := cableTermination(inv, cable.TerminationADevice, cable.TerminationAPort, cable.TerminationA)
		b := cableTermination(inv, cable.TerminationBDevice, cable.TerminationBPort, cable.TerminationB)
		if a.dev == uuid.Nil || b.dev == uuid.Nil {
			continue
		}
		idx.pairs = append(idx.pairs, [2]portRef{a, b})
		idx.cables[a] = b
		idx.cables[b] = a
	}
	for _, ip := range inv.IPAddresses {
		if ip == nil {
			continue
		}
		for _, ifaceID := range ip.Interfaces {
			idx.ips[ifaceID] = append(idx.ips[ifaceID], ip)
		}
	}
	for _, list := range idx.ips {
		sort.Slice(list, func(i, j int) bool { return ipHost(list[i]) < ipHost(list[j]) })
	}
	return idx
}

// cableTermination resolves one cable end to a device + port name, falling
// back to the interface UUID through the inventory's interface index.
func cableTermination(inv *devicetypes.Inventory, dev uuid.UUID, port string, iface uuid.UUID) portRef {
	if dev != uuid.Nil && port != "" {
		return portRef{dev: dev, port: port}
	}
	if inst := inv.Interfaces[iface]; inst != nil {
		return portRef{dev: inst.DeviceID, port: inst.Name}
	}
	return portRef{}
}

// peer returns the far end of the cable plugged into (dev, port).
func (idx *htmlIndex) peer(dev uuid.UUID, port string) (portRef, bool) {
	far, ok := idx.cables[portRef{dev: dev, port: port}]
	return far, ok
}

// resolve returns the device and interface spec behind a cable end; either
// may be nil when the inventory no longer holds them.
func (idx *htmlIndex) resolve(ref portRef) (*devicetypes.CaniDeviceType, *devicetypes.InterfaceSpec) {
	dev := idx.inv.Devices[ref.dev]
	return dev, dev.GetInterface(ref.port)
}

// ipOf returns the first address bound to an interface, IPv4 before IPv6.
func (idx *htmlIndex) ipOf(iface uuid.UUID) *devicetypes.CaniIPAddress {
	list := idx.ips[iface]
	for _, ip := range list {
		if !strings.Contains(ipHost(ip), ":") {
			return ip
		}
	}
	if len(list) > 0 {
		return list[0]
	}
	return nil
}

// ipHost returns the address without its mask.
func ipHost(ip *devicetypes.CaniIPAddress) string {
	if ip.Host != "" {
		return ip.Host
	}
	host, _, _ := strings.Cut(ip.Address, "/")
	return host
}

// cableCount counts the cables with at least one end on one of the devices.
func (idx *htmlIndex) cableCount(devices map[uuid.UUID]bool) int {
	n := 0
	for _, pair := range idx.pairs {
		if devices[pair[0].dev] || devices[pair[1].dev] {
			n++
		}
	}
	return n
}

// buildPorts lists the interfaces of dev that have a cable, an IP or a MAC,
// in device order, and returns the number of cabled ones.
func buildPorts(idx *htmlIndex, dev *devicetypes.CaniDeviceType) ([]htmlPort, int) {
	var ports []htmlPort
	cabled := 0
	for i := range dev.Interfaces {
		spec := &dev.Interfaces[i]
		far, hasCable := idx.peer(dev.ID, spec.Name)
		ip := idx.ipOf(spec.ID)
		if !hasCable && ip == nil && spec.MacAddress == "" {
			continue
		}
		port := htmlPort{Name: spec.Name, MAC: spec.MacAddress, Cabled: hasCable}
		peerDev, peerSpec := idx.resolve(far)
		if hasCable {
			cabled++
			port.PeerID, port.PeerName, port.PeerPort = peerIdentity(peerDev, far)
			port.VLAN = portVLAN(spec, peerSpec)
		} else {
			chassisUplink(idx, dev, spec, &port)
		}
		port.Hint = speedHint(spec, peerDev, peerSpec)
		if ip != nil {
			port.IP, port.IPNote = ipHost(ip), ipNote(ip)
			port.IPURL = idx.mgmtURL(dev, spec, ip)
		}
		ports = append(ports, port)
	}
	return ports, cabled
}

// peerIdentity names the far end of a cable; a dangling cable keeps the port
// name but gets no popover target.
func peerIdentity(peer *devicetypes.CaniDeviceType, far portRef) (id, name, port string) {
	if peer == nil {
		return "", "unknown device", far.port
	}
	return peer.ID.String(), peer.Name, far.port
}

// ipNote marks addresses that are reserved rather than live.
func ipNote(ip *devicetypes.CaniIPAddress) string {
	if strings.EqualFold(ip.Status, string(devicetypes.StatusReserved)) {
		return "(reserved)"
	}
	return ""
}

// vlanLabel renders the switchport configuration of one interface: the
// access VLAN id, "trunk" for tagged modes, "" when the port has no mode.
func vlanLabel(spec *devicetypes.InterfaceSpec) string {
	if spec == nil {
		return ""
	}
	switch spec.Mode {
	case devicetypes.InterfaceModeAccess:
		if spec.UntaggedVLAN > 0 {
			return strconv.Itoa(spec.UntaggedVLAN)
		}
	case devicetypes.InterfaceModeTagged, devicetypes.InterfaceModeTaggedAll:
		return "trunk"
	}
	return ""
}

// portVLAN is the VLAN text of a ports-table row, taken from whichever end
// of the cable is the configured switchport (the local end first).
func portVLAN(local, peer *devicetypes.InterfaceSpec) string {
	label := vlanLabel(local)
	if label == "" {
		label = vlanLabel(peer)
	}
	if label == "" || label == "trunk" {
		return label
	}
	return "VLAN " + label
}

// speedHint derives the muted port hint: "1G mgmt", "25G", "200G IB".
func speedHint(spec *devicetypes.InterfaceSpec, peer *devicetypes.CaniDeviceType, peerSpec *devicetypes.InterfaceSpec) string {
	var parts []string
	if speed := speedOf(spec.Type); speed != "" {
		parts = append(parts, speed)
	}
	if isMgmtPort(spec) {
		parts = append(parts, "mgmt")
	}
	if isInfiniband(spec, peer, peerSpec) {
		parts = append(parts, "IB")
	}
	return strings.Join(parts, " ")
}

// speedOf maps an interface type onto its nominal speed: 1000base-t -> 1G,
// 25gbase-x-sfp28 -> 25G; virtual and lag types have none.
func speedOf(t devicetypes.InterfacesElemType) string {
	prefix, _, found := strings.Cut(string(t), "base")
	if !found || prefix == "" {
		return ""
	}
	if strings.HasSuffix(prefix, "g") {
		return strings.ToUpper(prefix)
	}
	mbit, err := strconv.Atoi(prefix)
	if err != nil {
		return ""
	}
	if mbit%1000 == 0 {
		return fmt.Sprintf("%dG", mbit/1000)
	}
	return fmt.Sprintf("%dM", mbit)
}

// isMgmtOnly reports whether the interface is flagged as management-only.
func isMgmtOnly(spec *devicetypes.InterfaceSpec) bool {
	return spec.MgmtOnly != nil && *spec.MgmtOnly
}

// isInfiniband reports an IB port by role, or by its peer being an HSN switch
// because role data is inconsistent across the inventory; a management port
// on either end is never a fabric link.
func isInfiniband(spec *devicetypes.InterfaceSpec, peer *devicetypes.CaniDeviceType, peerSpec *devicetypes.InterfaceSpec) bool {
	role := strings.ToLower(spec.Role)
	if strings.Contains(role, "infiniband") || role == devicetypes.InterfaceRoleHSN {
		return true
	}
	if isMgmtPort(spec) || (peerSpec != nil && isMgmtPort(peerSpec)) {
		return false
	}
	return peer != nil && peer.Type == devicetypes.TypeHsnSwitch
}
