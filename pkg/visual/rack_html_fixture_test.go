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
	"bytes"
	"fmt"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// Cabling fixture: one 20U rack with a server cabled to two switches (iLO on
// an access port, NIC on a breakout access port, one uncabled IB port), a
// switch-to-switch trunk, an idle switch, a device with hostile characters in
// its name and interface label, a chassis with one child, and two devices
// exercising the bmc_url override (one valid, one hostile).
var (
	cabRackID    = uuid.MustParse("00000000-0000-0000-0010-000000000001")
	cabServerID  = uuid.MustParse("00000000-0000-0000-0011-000000000001")
	cab6300ID    = uuid.MustParse("00000000-0000-0000-0011-000000000002")
	cab8325ID    = uuid.MustParse("00000000-0000-0000-0011-000000000003")
	cabIdleID    = uuid.MustParse("00000000-0000-0000-0011-000000000004")
	cabEvilID    = uuid.MustParse("00000000-0000-0000-0011-000000000005")
	cabChassisID = uuid.MustParse("00000000-0000-0000-0011-000000000006")
	cabChildID   = uuid.MustParse("00000000-0000-0000-0011-000000000007")
	cabBMCID     = uuid.MustParse("00000000-0000-0000-0011-000000000008")
	cabEvilBMCID = uuid.MustParse("00000000-0000-0000-0011-000000000009")
)

func cabIface(n int) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0012-%012d", n))
}

func cabIP(n int) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0013-%012d", n))
}

func boolPtr(b bool) *bool { return &b }

func cabSpec(id int, name string, t devicetypes.InterfacesElemType) devicetypes.InterfaceSpec {
	return devicetypes.InterfaceSpec{ID: cabIface(id), Name: name, Type: t}
}

func cabAccess(id int, name, label string, vlan int) devicetypes.InterfaceSpec {
	spec := cabSpec(id, name, devicetypes.InterfacesElemTypeA1000BaseT)
	spec.Label, spec.Mode, spec.UntaggedVLAN = label, devicetypes.InterfaceModeAccess, vlan
	return spec
}

func cabMgmt(id int, name, mac string) devicetypes.InterfaceSpec {
	spec := cabSpec(id, name, devicetypes.InterfacesElemTypeA1000BaseT)
	spec.MgmtOnly, spec.MacAddress = boolPtr(true), mac
	return spec
}

func cabCable(inv *devicetypes.Inventory, n int, devA uuid.UUID, portA string, devB uuid.UUID, portB string) {
	id := uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0014-%012d", n))
	inv.Cables[id] = &devicetypes.CaniCableType{ID: id, Slug: "generic-cat6-patch",
		TerminationADevice: devA, TerminationAPort: portA, TerminationBDevice: devB, TerminationBPort: portB}
}

func cabAddress(inv *devicetypes.Inventory, n int, host, status string, iface uuid.UUID) {
	inv.IPAddresses[cabIP(n)] = &devicetypes.CaniIPAddress{ID: cabIP(n), Host: host, Address: host + "/24",
		Interfaces: []uuid.UUID{iface}, ObjectMeta: devicetypes.ObjectMeta{Status: status}}
}

func cabFru(inv *devicetypes.Inventory, n int, dev uuid.UUID, t devicetypes.Type, name, model, serial string, fields map[string]any) {
	id := uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0015-%012d", n))
	inv.Frus[id] = &devicetypes.CaniFruType{ID: id, Name: name, Model: model, Manufacturer: "Acme", Type: t, Serial: serial,
		PartNumber: "PN-" + model, Device: dev, ObjectMeta: devicetypes.ObjectMeta{CustomFields: fields}}
}

func cablingInventory() *devicetypes.Inventory {
	inv := devicetypes.NewInventory()
	inv.Racks[cabRackID] = &devicetypes.CaniRackType{ID: cabRackID, Name: "Rack-X", UHeight: 20,
		Devices: []uuid.UUID{cabServerID, cab6300ID, cab8325ID, cabIdleID, cabEvilID, cabChassisID, cabBMCID, cabEvilBMCID}}
	ib := cabSpec(3, "Slot 7 Port 1", devicetypes.InterfacesElemTypeA200GbaseXQsfp56)
	ib.Role, ib.MacAddress = "infiniband", "aa:bb:cc:00:00:03"
	nic := cabSpec(2, "Slot 10 Port 1", devicetypes.InterfacesElemTypeA25GbaseXSfp28)
	nic.MacAddress = "aa:bb:cc:00:00:02"
	inv.Devices[cabServerID] = &devicetypes.CaniDeviceType{ID: cabServerID, Name: "srv", Model: "DL380", Type: devicetypes.TypeNode,
		UHeight: 2, IsFullDepth: true, RackPosition: 3, Parent: cabRackID, ObjectMeta: devicetypes.ObjectMeta{Status: "Active"},
		Description: "Issue: fan 3 reports degraded.",
		Interfaces:  []devicetypes.InterfaceSpec{cabMgmt(1, "iLO", "aa:bb:cc:00:00:01"), nic, ib}}
	uplink := cabSpec(14, "1/1/49", devicetypes.InterfacesElemTypeA25GbaseXSfp28)
	uplink.Label, uplink.Mode = "uplink sw8325", devicetypes.InterfaceModeTaggedAll
	inv.Devices[cab6300ID] = &devicetypes.CaniDeviceType{ID: cab6300ID, Name: "sw6300", Model: "6300M", Type: devicetypes.TypeSwitch,
		UHeight: 1, RackPosition: 10, Parent: cabRackID, ObjectMeta: devicetypes.ObjectMeta{Status: "Active"},
		Interfaces: []devicetypes.InterfaceSpec{cabAccess(10, "1/1/1", "srv-ilo", 1084), cabAccess(11, "1/1/2", "<b>x</b>", 1083),
			cabAccess(12, "1/1/3", "", 1083), cabAccess(13, "1/1/4", "", 1083), uplink, cabMgmt(15, "mgmt", "")}}
	trunk := cabSpec(25, "1/1/2", devicetypes.InterfacesElemTypeA100GbaseXQsfp28)
	trunk.Mode = devicetypes.InterfaceModeTaggedAll
	inv.Devices[cab8325ID] = &devicetypes.CaniDeviceType{ID: cab8325ID, Name: "sw8325", Model: "8325", Type: devicetypes.TypeSwitch,
		UHeight: 1, RackPosition: 12, Parent: cabRackID, ObjectMeta: devicetypes.ObjectMeta{Status: "Active"},
		Interfaces: []devicetypes.InterfaceSpec{cabSpec(20, "1/1/1", devicetypes.InterfacesElemTypeA100GbaseXQsfp28),
			cabAccess(21, "1/1/1:1", "srv-eth", 1083), cabSpec(22, "1/1/1:2", devicetypes.InterfacesElemTypeA25GbaseXSfp28),
			cabSpec(23, "1/1/1:3", devicetypes.InterfacesElemTypeA25GbaseXSfp28), cabSpec(24, "1/1/1:4", devicetypes.InterfacesElemTypeA25GbaseXSfp28),
			trunk, cabMgmt(26, "mgmt", "")}}
	inv.Devices[cabIdleID] = &devicetypes.CaniDeviceType{ID: cabIdleID, Name: "sw-idle", Type: devicetypes.TypeSwitch,
		UHeight: 1, RackPosition: 14, Parent: cabRackID, ObjectMeta: devicetypes.ObjectMeta{Status: "Planned"},
		Interfaces: []devicetypes.InterfaceSpec{cabAccess(30, "1/1/1", "", 1), cabAccess(31, "1/1/2", "", 1)}}
	evilNIC := cabSpec(40, "eth0", devicetypes.InterfacesElemTypeA1000BaseT)
	evilNIC.Label, evilNIC.MacAddress = "<b>x</b>", "aa:bb:cc:00:00:40"
	inv.Devices[cabEvilID] = &devicetypes.CaniDeviceType{ID: cabEvilID, Name: `evil"]<`, Type: devicetypes.TypeNode,
		UHeight: 1, RackPosition: 16, Parent: cabRackID, ObjectMeta: devicetypes.ObjectMeta{Status: "Active"},
		Interfaces: []devicetypes.InterfaceSpec{evilNIC}}
	inv.Devices[cabChassisID] = &devicetypes.CaniDeviceType{ID: cabChassisID, Name: "chas", Type: devicetypes.TypeChassis,
		UHeight: 2, IsFullDepth: true, RackPosition: 18, Parent: cabRackID, Children: []uuid.UUID{cabChildID},
		ObjectMeta: devicetypes.ObjectMeta{Status: "Active"}}
	inv.Devices[cabChildID] = &devicetypes.CaniDeviceType{ID: cabChildID, Name: "cn1", Type: devicetypes.TypeBlade, Parent: cabChassisID,
		ObjectMeta: devicetypes.ObjectMeta{Status: "Active"}, Interfaces: []devicetypes.InterfaceSpec{cabMgmt(50, "iLO", "")}}
	inv.Devices[cabBMCID] = &devicetypes.CaniDeviceType{ID: cabBMCID, Name: "riscv-box", Type: "other",
		UHeight: 1, RackPosition: 7, Parent: cabRackID, Interfaces: []devicetypes.InterfaceSpec{cabMgmt(60, "mgmt", "")},
		ObjectMeta: devicetypes.ObjectMeta{Status: "Active", CustomFields: map[string]any{"bmc_url": "http://10.0.84.14:8000/"}}}
	inv.Devices[cabEvilBMCID] = &devicetypes.CaniDeviceType{ID: cabEvilBMCID, Name: "evil-bmc", Type: devicetypes.TypeNode,
		UHeight: 1, RackPosition: 8, Parent: cabRackID, Interfaces: []devicetypes.InterfaceSpec{cabMgmt(70, "mgmt", "")},
		Description: "Not verified: no BMC answered.",
		ObjectMeta:  devicetypes.ObjectMeta{Status: "Active", Tags: []string{"unverified"}, CustomFields: map[string]any{"bmc_url": "javascript:alert(1)"}}}
	cabCable(inv, 1, cabServerID, "iLO", cab6300ID, "1/1/1")
	cabCable(inv, 2, cabServerID, "Slot 10 Port 1", cab8325ID, "1/1/1:1")
	cabCable(inv, 3, cab6300ID, "1/1/49", cab8325ID, "1/1/2")
	cabCable(inv, 4, cabEvilID, "eth0", cab6300ID, "1/1/2")
	cabAddress(inv, 1, "10.0.84.5", "Active", cabIface(1))
	cabAddress(inv, 2, "10.0.83.5", "Reserved", cabIface(2))
	cabAddress(inv, 3, "10.0.83.240", "Active", cabIface(15))
	cabAddress(inv, 4, "10.0.84.9", "Active", cabIface(50))
	cabAddress(inv, 5, "10.0.84.14", "Active", cabIface(60))
	cabAddress(inv, 6, "10.0.84.70", "Active", cabIface(70))
	addCablingHardware(inv)
	return inv
}

// addCablingHardware gives srv 2 CPUs, 16 + 1 DIMMs, one drive and two GPU
// modules; CustomFields mix int and float64 as JSON decoding would produce.
func addCablingHardware(inv *devicetypes.Inventory) {
	for i := 1; i <= 2; i++ {
		cabFru(inv, i, cabServerID, devicetypes.TypeCPU, fmt.Sprintf("srv CPU %d", i), "EPYC 7543 32-Core", fmt.Sprintf("CPU%d", i),
			map[string]any{"cores": 32, "threads": 64, "socket": fmt.Sprintf("Proc %d", i)})
	}
	for i := 1; i <= 16; i++ {
		cabFru(inv, 10+i, cabServerID, devicetypes.TypeMemory, fmt.Sprintf("srv DIMM %d", i), "32 GiB DDR4", fmt.Sprintf("DIMM%d", i),
			map[string]any{"capacity_mib": float64(32768), "speed_mhz": float64(3200), "slot": fmt.Sprintf("PROC 1 DIMM %d", i)})
	}
	cabFru(inv, 30, cabServerID, devicetypes.TypeMemory, "srv DIMM 17", "16 GiB DDR4", "DIMM17",
		map[string]any{"capacity_mib": 16384, "speed_mhz": 3200, "slot": "PROC 2 DIMM 1"})
	cabFru(inv, 31, cabServerID, "disk", "srv disk 1", "PM9A3", "DISK1",
		map[string]any{"capacity_bytes": float64(1920383410176), "media": "SSD", "protocol": "NVMe", "slot": "Box 1 Bay 1"})
	for i, bay := range []string{"PCIe1", "PCIe3"} {
		id := uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0016-%012d", i+1))
		inv.Modules[id] = &devicetypes.CaniModuleType{ID: id, Name: "srv-" + bay, Manufacturer: "NVIDIA", Model: "A100 80GB", Type: devicetypes.TypeGPU,
			ParentDevice: cabServerID, ModuleBayName: bay, Serial: fmt.Sprintf("GPU%d", i+1), PartNumber: "900-21001"}
	}
}

func renderCabling(t *testing.T) string {
	t.Helper()
	var buf bytes.Buffer
	if err := RenderRackHTML(&buf, cablingInventory(), CompactRenderOptions{}); err != nil {
		t.Fatalf("RenderRackHTML: %v", err)
	}
	return buf.String()
}

func cablingDevice(t *testing.T, inv *devicetypes.Inventory, id uuid.UUID) *devicetypes.CaniDeviceType {
	t.Helper()
	dev := inv.Devices[id]
	if dev == nil {
		t.Fatalf("fixture device %s missing", id)
	}
	return dev
}
