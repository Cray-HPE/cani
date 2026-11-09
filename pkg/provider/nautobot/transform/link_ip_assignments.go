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
	"fmt"
	"sort"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// LinkIPAssignments sets each imported address's Interfaces from the
// assignment rows, after the addresses and every port, device-owned or
// module-owned, exist in the result. An imported interface spec keeps its
// Nautobot UUID as its ID, so an assignment resolves to the spec the result
// placed. A row for a VM interface, for an address the import left out, or
// for a port no device or module carries is left out and reported in the
// same form as the other records the transform cannot place. The seven
// is_* flags of a row are not modelled: the assignment is kept and every
// flag the row sets is named in a warning, never silently dropped.
func LinkIPAssignments(result *devicetypes.TransformResult, rows []nautobotapi.IPAddressToInterface, ifaceMap map[uuid.UUID]ifaceRef) []string {
	addresses := nautobotIDIndex(result.IPAddresses, addressExternalIDs)
	placed := placedInterfaces(result)
	var warnings []string
	for i := range rows {
		row := &rows[i]
		if row.VmInterface != nil && row.Interface == nil {
			warnings = append(warnings, fmt.Sprintf("ip assignment %s: virtual machine interfaces are not modelled", rowLabel(row)))
			continue
		}
		address := result.IPAddresses[addresses[refID(row.IpAddress)]]
		if address == nil {
			warnings = append(warnings, fmt.Sprintf("ip assignment %s: address %s not imported", rowLabel(row), refID(row.IpAddress)))
			continue
		}
		ifaceID := refID(row.Interface)
		if !placed[ifaceID] {
			warnings = append(warnings, fmt.Sprintf("ip assignment %s: interface %s not imported", rowLabel(row), ifaceMap[ifaceID].name))
			continue
		}
		address.Interfaces = appendUnique(address.Interfaces, ifaceID)
		if flags := setAssignmentFlags(row); len(flags) > 0 {
			warnings = append(warnings, fmt.Sprintf("ip assignment %s: %s on %s: %s not modelled; the assignment is kept without them",
				rowLabel(row), address.Address, ifaceMap[ifaceID].name, strings.Join(flags, ", ")))
		}
	}
	sortAddressInterfaces(result.IPAddresses)
	return warnings
}

// setAssignmentFlags names, in Nautobot's field order, the is_* flags a row
// sets; CANI carries none of them.
func setAssignmentFlags(row *nautobotapi.IPAddressToInterface) []string {
	var set []string
	for _, flag := range []struct {
		name string
		on   *bool
	}{
		{"is_source", row.IsSource}, {"is_destination", row.IsDestination}, {"is_default", row.IsDefault},
		{"is_preferred", row.IsPreferred}, {"is_primary", row.IsPrimary}, {"is_secondary", row.IsSecondary},
		{"is_standby", row.IsStandby},
	} {
		if flag.on != nil && *flag.on {
			set = append(set, flag.name)
		}
	}
	return set
}

// placedInterfaces indexes the ID of every interface spec a mapped device or
// module carries.
func placedInterfaces(result *devicetypes.TransformResult) map[uuid.UUID]bool {
	placed := make(map[uuid.UUID]bool)
	for _, device := range result.Devices {
		if device != nil {
			indexSpecIDs(placed, device.Interfaces)
		}
	}
	for _, module := range result.Modules {
		if module != nil {
			indexSpecIDs(placed, module.Interfaces)
		}
	}
	return placed
}

func indexSpecIDs(into map[uuid.UUID]bool, specs []devicetypes.InterfaceSpec) {
	for i := range specs {
		into[specs[i].ID] = true
	}
}

// rowLabel names an assignment row by its Nautobot UUID for a warning.
func rowLabel(row *nautobotapi.IPAddressToInterface) string {
	return directUUID(row.Id).String()
}

// appendUnique adds id unless the slice already holds it.
func appendUnique(ids []uuid.UUID, id uuid.UUID) []uuid.UUID {
	for _, have := range ids {
		if have == id {
			return ids
		}
	}
	return append(ids, id)
}

// sortAddressInterfaces orders every address's assignments so the saved
// inventory does not depend on the order Nautobot listed the rows.
func sortAddressInterfaces(addresses map[uuid.UUID]*devicetypes.CaniIPAddress) {
	for _, address := range addresses {
		if address == nil || len(address.Interfaces) < 2 {
			continue
		}
		sort.Slice(address.Interfaces, func(i, j int) bool {
			return address.Interfaces[i].String() < address.Interfaces[j].String()
		})
	}
}
