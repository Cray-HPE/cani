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
package transform

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// assignmentRow builds an ip-address-to-interface row.
func assignmentRow(ipNB, ifaceNB uuid.UUID) nautobotapi.IPAddressToInterface {
	row := nautobotapi.IPAddressToInterface{Id: oaPtr(uuid.New())}
	setNBRef(&row.IpAddress, ipNB)
	setNBRef(&row.Interface, ifaceNB)
	return row
}

// vmAssignmentRow builds a row that assigns the address to a VM interface.
func vmAssignmentRow(ipNB uuid.UUID) nautobotapi.IPAddressToInterface {
	row := nautobotapi.IPAddressToInterface{Id: oaPtr(uuid.New())}
	setNBRef(&row.IpAddress, ipNB)
	setNBRef(&row.VmInterface, uuid.New())
	return row
}

// TestLinkIPAssignments_LinksDeviceAndModulePortsManyToMany verifies that
// one address assigned to a device port and a module port ends up with both
// interface IDs, that a duplicate row adds nothing, and that the list is
// sorted.
//
// Why it matters: Nautobot 2+ lets one address sit on several interfaces
// (anycast), and module ports are the ones the primary-IP phase most often
// needs; an import that linked only device ports would lose the rest.
// Inputs: an address, a device carrying eth0, a module carrying xe-0/0/0,
// three rows (eth0, xe-0/0/0, eth0 again).
// Outputs: Interfaces holds exactly the two spec IDs in sorted order; no
// warnings.
// Data choice: the duplicate row is the same assignment listed twice, the
// shape a paginated listing can produce.
func TestLinkIPAssignments_LinksDeviceAndModulePortsManyToMany(t *testing.T) {
	ipNB, eth0, xe0 := uuid.New(), uuid.New(), uuid.New()
	devID, modID, ipID := uuid.New(), uuid.New(), uuid.New()
	result := &devicetypes.TransformResult{
		Devices:     map[uuid.UUID]*devicetypes.CaniDeviceType{devID: {ID: devID, Interfaces: []devicetypes.InterfaceSpec{{ID: eth0, Name: "eth0"}}}},
		Modules:     map[uuid.UUID]*devicetypes.CaniModuleType{modID: {ID: modID, Interfaces: []devicetypes.InterfaceSpec{{ID: xe0, Name: "xe-0/0/0"}}}},
		IPAddresses: map[uuid.UUID]*devicetypes.CaniIPAddress{ipID: {ID: ipID, Address: "10.0.0.1/24", ObjectMeta: external(ipNB)}},
	}
	rows := []nautobotapi.IPAddressToInterface{assignmentRow(ipNB, eth0), assignmentRow(ipNB, xe0), assignmentRow(ipNB, eth0)}

	warnings := LinkIPAssignments(result, rows, nil)

	want := []uuid.UUID{eth0, xe0}
	if eth0.String() > xe0.String() {
		want = []uuid.UUID{xe0, eth0}
	}
	if got := result.IPAddresses[ipID].Interfaces; !reflect.DeepEqual(got, want) {
		t.Errorf("Interfaces = %v, want %v", got, want)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
}

// TestLinkIPAssignments_ReportsRowsItCannotPlace verifies that a VM
// interface row, a row for an address the import left out, and a row for a
// port no device or module carries are skipped and each named once.
//
// Why it matters: the Story requires every value the import cannot carry to
// be named rather than dropped silently; a VM assignment is the one kind of
// row CANI does not model at all.
// Inputs: one imported address; rows for a VM interface, an unknown
// address, and an unplaced port named in the interface map.
// Outputs: the address keeps no assignments; three warnings, the third
// naming the port.
// Data choice: the three distinct reasons a row can fail, one each.
func TestLinkIPAssignments_ReportsRowsItCannotPlace(t *testing.T) {
	ipNB, strayNB, ipID := uuid.New(), uuid.New(), uuid.New()
	result := &devicetypes.TransformResult{
		IPAddresses: map[uuid.UUID]*devicetypes.CaniIPAddress{ipID: {ID: ipID, ObjectMeta: external(ipNB)}},
	}
	rows := []nautobotapi.IPAddressToInterface{
		vmAssignmentRow(ipNB), assignmentRow(uuid.New(), strayNB), assignmentRow(ipNB, strayNB),
	}

	warnings := LinkIPAssignments(result, rows, map[uuid.UUID]ifaceRef{strayNB: {name: "ge-0/0/9"}})

	if got := result.IPAddresses[ipID].Interfaces; len(got) != 0 {
		t.Errorf("Interfaces = %v, want none", got)
	}
	if len(warnings) != 3 {
		t.Fatalf("warnings = %v, want 3", warnings)
	}
	if !strings.Contains(warnings[0], "virtual machine") || !strings.Contains(warnings[1], "address") || !strings.Contains(warnings[2], "ge-0/0/9") {
		t.Errorf("warnings = %v, want VM, missing address, unplaced port in order", warnings)
	}
}

// TestLinkIPAssignments_KeepsAnAssignmentAndNamesTheFlagsItDrops verifies
// that a row setting any of the seven is_* flags is still linked, and that
// the set flags are named in one warning in Nautobot's field order, while a
// row with every flag false or unset warns about nothing.
//
// Why it matters: CANI does not model the per-assignment flags; the Story
// requires a value the import cannot carry to be named, not dropped
// silently, and the assignment itself must survive because the address's
// placement is what the export recreates.
// Inputs: one address on one device port; a row with is_primary and
// is_standby true and is_source explicitly false, and a row with no flags.
// Outputs: the address is linked once; exactly one warning, naming the port
// and "is_primary, is_standby" and not "is_source".
// Data choice: one flag from each end of the field order plus an explicit
// false show that only true flags are reported and that order is Nautobot's.
func TestLinkIPAssignments_KeepsAnAssignmentAndNamesTheFlagsItDrops(t *testing.T) {
	ipNB, eth0, devID, ipID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	result := &devicetypes.TransformResult{
		Devices:     map[uuid.UUID]*devicetypes.CaniDeviceType{devID: {ID: devID, Interfaces: []devicetypes.InterfaceSpec{{ID: eth0, Name: "eth0"}}}},
		IPAddresses: map[uuid.UUID]*devicetypes.CaniIPAddress{ipID: {ID: ipID, Address: "10.0.0.1/24", ObjectMeta: external(ipNB)}},
	}
	on, off := true, false
	flagged := assignmentRow(ipNB, eth0)
	flagged.IsPrimary, flagged.IsStandby, flagged.IsSource = &on, &on, &off
	rows := []nautobotapi.IPAddressToInterface{flagged, assignmentRow(ipNB, eth0)}

	warnings := LinkIPAssignments(result, rows, map[uuid.UUID]ifaceRef{eth0: {name: "eth0"}})

	if got := result.IPAddresses[ipID].Interfaces; !reflect.DeepEqual(got, []uuid.UUID{eth0}) {
		t.Errorf("Interfaces = %v, want [%s]", got, eth0)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings = %v, want exactly one", warnings)
	}
	for _, want := range []string{"10.0.0.1/24", "eth0", "is_primary, is_standby", "not modelled"} {
		if !strings.Contains(warnings[0], want) {
			t.Errorf("warning = %q, want it to contain %q", warnings[0], want)
		}
	}
	if strings.Contains(warnings[0], "is_source") {
		t.Errorf("warning = %q names a flag the row left false", warnings[0])
	}
}

// TestTransformRaw_CarriesIPAssignmentsAndPrimaryIP verifies the pipeline
// end to end: a device with one interface, an address assigned to it and
// set as the device's primary IPv4 come out linked on both sides.
//
// Why it matters: the primary-IP export is refused by Nautobot unless the
// address is assigned to one of the device's interfaces, so the two links
// must travel together or the seed fails on the destination.
// Inputs: raw data with one status, one device type, one device, one
// interface, one namespace, one prefix, one address, one assignment row,
// and the device's primary_ip4 naming the address.
// Outputs: the mapped address's Interfaces is the interface's UUID and the
// device's PrimaryIPv4 is the mapped address's ID.
// Data choice: the minimum raw set that reaches both link steps.
func TestTransformRaw_CarriesIPAssignmentsAndPrimaryIP(t *testing.T) {
	devNB, ifaceNB, ipNB, statusNB, dtNB, nsNB := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	raw := &RawData{
		Statuses:    []nautobotapi.Status{{Id: oaPtr(statusNB), Name: "Active"}},
		DeviceTypes: []nautobotapi.DeviceType{{Id: oaPtr(dtNB), Model: "DL380"}},
		Devices:     []nautobotapi.Device{{Id: oaPtr(devNB), Name: strPtr("compute-001")}},
		Interfaces:  []nautobotapi.Interface{{Id: oaPtr(ifaceNB), Name: "eth0"}},
		Namespaces:  []nautobotapi.Namespace{{Id: oaPtr(nsNB), Name: "Global"}},
		IPAddresses: []nautobotapi.IPAddress{{Id: oaPtr(ipNB), Address: "10.0.0.1/24"}},
	}
	setNBRef(&raw.Devices[0].Status, statusNB)
	setNBRef(&raw.Devices[0].DeviceType, dtNB)
	setNBRef(&raw.Devices[0].PrimaryIp4, ipNB)
	setNBRef(&raw.Interfaces[0].Device, devNB)
	setNBRef(&raw.IPAddresses[0].Status, statusNB)
	raw.IPAddressAssignments = []nautobotapi.IPAddressToInterface{assignmentRow(ipNB, ifaceNB)}

	result, err := Transform(devicetypes.Inventory{}, raw)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}

	address := onlyValue(t, result.IPAddresses)
	if !reflect.DeepEqual(address.Interfaces, []uuid.UUID{ifaceNB}) {
		t.Errorf("address.Interfaces = %v, want [%s]", address.Interfaces, ifaceNB)
	}
	if device := onlyValue(t, result.Devices); device.PrimaryIPv4 != address.ID {
		t.Errorf("PrimaryIPv4 = %s, want %s", device.PrimaryIPv4, address.ID)
	}
}
