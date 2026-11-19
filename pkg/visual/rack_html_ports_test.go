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
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func TestBuildPortsListsServerInterfacesInDeviceOrder(t *testing.T) {
	inv := cablingInventory()
	ports, cabled := buildPorts(newHTMLIndex(inv), cablingDevice(t, inv, cabServerID))
	if len(ports) != 3 || cabled != 2 {
		t.Fatalf("expected 3 rows with 2 cabled, got %d rows, %d cabled", len(ports), cabled)
	}
	want := []htmlPort{
		{Name: "iLO", Hint: "1G mgmt", PeerID: cab6300ID.String(), PeerName: "sw6300", PeerPort: "1/1/1", VLAN: "VLAN 1084",
			IP: "10.0.84.5", IPURL: "https://10.0.84.5/", MAC: "aa:bb:cc:00:00:01", Cabled: true},
		{Name: "Slot 10 Port 1", Hint: "25G", PeerID: cab8325ID.String(), PeerName: "sw8325", PeerPort: "1/1/1:1", VLAN: "VLAN 1083",
			IP: "10.0.83.5", IPNote: "(reserved)", MAC: "aa:bb:cc:00:00:02", Cabled: true},
		{Name: "Slot 7 Port 1", Hint: "200G IB", MAC: "aa:bb:cc:00:00:03"},
	}
	for i, w := range want {
		if ports[i] != w {
			t.Errorf("row %d:\n got %+v\nwant %+v", i, ports[i], w)
		}
	}
}

func TestBuildPortsBothEndsListEachOther(t *testing.T) {
	inv := cablingInventory()
	ports, cabled := buildPorts(newHTMLIndex(inv), cablingDevice(t, inv, cab6300ID))
	if cabled != 3 || len(ports) != 4 {
		t.Fatalf("switch should list 3 cabled ports plus its mgmt port with an IP, got %d rows, %d cabled", len(ports), cabled)
	}
	byName := map[string]htmlPort{}
	for _, p := range ports {
		byName[p.Name] = p
	}
	if p := byName["1/1/1"]; p.PeerID != cabServerID.String() || p.PeerPort != "iLO" || p.VLAN != "VLAN 1084" {
		t.Errorf("switch side of the iLO cable: %+v", p)
	}
	if p := byName["1/1/49"]; p.VLAN != "trunk" || p.PeerName != "sw8325" || p.PeerPort != "1/1/2" || p.Hint != "25G" {
		t.Errorf("switch-to-switch trunk: %+v", p)
	}
	if p := byName["mgmt"]; !strings.Contains(p.Hint, "mgmt") || p.IP != "10.0.83.240" || p.IPURL != "https://10.0.83.240/" || p.Cabled {
		t.Errorf("uncabled mgmt port with an address: %+v", p)
	}
}

func TestPortVLANFallsBackToTheSwitchEnd(t *testing.T) {
	access := &devicetypes.InterfaceSpec{Mode: devicetypes.InterfaceModeAccess, UntaggedVLAN: 1084}
	trunk := &devicetypes.InterfaceSpec{Mode: devicetypes.InterfaceModeTaggedAll}
	plain := &devicetypes.InterfaceSpec{}
	cases := []struct {
		local, peer *devicetypes.InterfaceSpec
		want        string
	}{
		{plain, access, "VLAN 1084"},
		{access, plain, "VLAN 1084"},
		{trunk, trunk, "trunk"},
		{plain, nil, ""},
		{nil, nil, ""},
	}
	for _, c := range cases {
		if got := portVLAN(c.local, c.peer); got != c.want {
			t.Errorf("portVLAN(%+v, %+v) = %q, want %q", c.local, c.peer, got, c.want)
		}
	}
}

func TestSpeedHint(t *testing.T) {
	hsn := &devicetypes.CaniDeviceType{Type: devicetypes.TypeHsnSwitch}
	cases := []struct {
		spec devicetypes.InterfaceSpec
		peer *devicetypes.CaniDeviceType
		want string
	}{
		{devicetypes.InterfaceSpec{Type: devicetypes.InterfacesElemTypeA1000BaseT, MgmtOnly: boolPtr(true)}, nil, "1G mgmt"},
		{devicetypes.InterfaceSpec{Type: devicetypes.InterfacesElemTypeA10GbaseT}, nil, "10G"},
		{devicetypes.InterfaceSpec{Type: devicetypes.InterfacesElemTypeA25GbaseXSfp28}, nil, "25G"},
		{devicetypes.InterfaceSpec{Type: devicetypes.InterfacesElemTypeA100GbaseXQsfp28}, nil, "100G"},
		{devicetypes.InterfaceSpec{Type: devicetypes.InterfacesElemTypeA200GbaseXQsfp56}, hsn, "200G IB"},
		{devicetypes.InterfaceSpec{Type: devicetypes.InterfacesElemTypeA200GbaseXQsfp56, Role: "infiniband"}, nil, "200G IB"},
		{devicetypes.InterfaceSpec{Type: devicetypes.InterfacesElemTypeVirtual}, nil, ""},
		{devicetypes.InterfaceSpec{Type: "100base-tx"}, nil, "100M"},
		{devicetypes.InterfaceSpec{Type: devicetypes.InterfacesElemTypeA1000BaseT, Name: "Management"}, hsn, "1G mgmt"},
	}
	for _, c := range cases {
		if got := speedHint(&c.spec, c.peer, nil); got != c.want {
			t.Errorf("speedHint(%q, role %q) = %q, want %q", c.spec.Type, c.spec.Role, got, c.want)
		}
	}
}

func TestRenderedPortsTableUsesShowOnlyPeerButtons(t *testing.T) {
	out := renderCabling(t)
	peer := `<button type="button" class="peer" popovertarget="p-` + cab6300ID.String() + `" popovertargetaction="show"><span class="peer-dev">sw6300</span> <span class="peer-port mono">1/1/1</span></button>`
	for _, want := range []string{peer, `<span class="vlan">VLAN 1084</span>`, `<span class="vlan">VLAN 1083</span>`, `<span class="weak">not cabled</span>`, `<span class="weak">(reserved)</span>`, `<span class="port-hint">200G IB</span>`} {
		if !strings.Contains(out, want) {
			t.Errorf("expected ports table to contain %q", want)
		}
	}
	if strings.Contains(out, `popovertargetaction="toggle"`) {
		t.Errorf("peer navigation must use popovertargetaction=show so revisiting the previous card never hides both")
	}
}

func TestCableIndexFallsBackToInterfaceUUIDs(t *testing.T) {
	inv := cablingInventory()
	for _, cable := range inv.Cables {
		cable.TerminationA = inv.FindInterfaceIDByPort(cable.TerminationADevice, cable.TerminationAPort)
		cable.TerminationB = inv.FindInterfaceIDByPort(cable.TerminationBDevice, cable.TerminationBPort)
		cable.TerminationADevice, cable.TerminationBDevice = uuid.Nil, uuid.Nil
		cable.TerminationAPort, cable.TerminationBPort = "", ""
	}
	inv.RebuildDerivedState()
	idx := newHTMLIndex(inv)
	far, ok := idx.peer(cabServerID, "iLO")
	if !ok || far.dev != cab6300ID || far.port != "1/1/1" {
		t.Fatalf("expected the iLO cable to resolve through interface UUIDs, got %+v (%v)", far, ok)
	}
	if n := idx.cableCount(map[uuid.UUID]bool{cabServerID: true}); n != 2 {
		t.Errorf("expected 2 cables on the server, got %d", n)
	}
}

func TestBuildPortsShowsChassisUplinkForUncabledNodeILO(t *testing.T) {
	inv := cablingInventory()
	chassis := cablingDevice(t, inv, cabChassisID)
	chassis.Interfaces = append(chassis.Interfaces, cabMgmt(80, "iLO port 1", ""))
	cabCable(inv, 80, cabChassisID, "iLO port 1", cab6300ID, "1/1/3")

	ports, cabled := buildPorts(newHTMLIndex(inv), cablingDevice(t, inv, cabChildID))
	if len(ports) != 1 || cabled != 0 {
		t.Fatalf("expected the child's iLO as the only, uncabled row; got %d rows, %d cabled", len(ports), cabled)
	}
	got := ports[0]
	if got.Cabled || got.ViaID != cabChassisID.String() || got.ViaName != "chas" || got.ViaPort != "iLO port 1" ||
		got.PeerID != cab6300ID.String() || got.PeerPort != "1/1/3" || got.VLAN != "VLAN 1083" {
		t.Fatalf("unexpected via row: %+v", got)
	}
}

func TestBuildPortsKeepsNotCabledWithoutChassisUplink(t *testing.T) {
	inv := cablingInventory()
	ports, _ := buildPorts(newHTMLIndex(inv), cablingDevice(t, inv, cabChildID))
	if len(ports) != 1 || ports[0].ViaID != "" || ports[0].PeerID != "" {
		t.Fatalf("expected a plain not-cabled row, got %+v", ports)
	}
}
