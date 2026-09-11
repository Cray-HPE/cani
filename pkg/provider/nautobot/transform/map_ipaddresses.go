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
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// MapIPAddresses converts Nautobot IPAddress objects to CANI IP addresses.
// It requires:
//   - statusNameMap: Nautobot status UUID → status name
//   - roleNameMap: Nautobot role UUID → role name
//
// Interface assignments, NAT relationships, and the parent prefix are
// intentionally left unset: interfaces are re-linked separately and the
// parent prefix is recomputed from the address after import.
func MapIPAddresses(
	raw []nautobotapi.IPAddress,
	statusNameMap map[uuid.UUID]string,
	roleNameMap map[uuid.UUID]string,
) map[uuid.UUID]*devicetypes.CaniIPAddress {
	result := make(map[uuid.UUID]*devicetypes.CaniIPAddress, len(raw))

	for _, ip := range raw {
		nbID := directUUID(ip.Id)
		if nbID == uuid.Nil {
			continue
		}
		caniID := uuid.New()

		caniIP := &devicetypes.CaniIPAddress{
			ID:          caniID,
			Address:     ip.Address,
			Host:        hostFromAddress(ip.Address),
			MaskLength:  intVal(ip.MaskLength),
			IPVersion:   intVal(ip.IpVersion),
			DNSName:     strVal(ip.DnsName),
			Description: strVal(ip.Description),
			ObjectMeta: devicetypes.ObjectMeta{
				Status:      resolveRefName(ip.Status, statusNameMap),
				ExternalIDs: map[string]uuid.UUID{"nautobot": nbID},
			},
		}
		if ip.Type != nil {
			caniIP.Type = devicetypes.IPAddressType(string(*ip.Type))
		}
		if roleName := resolveTenantRefName(ip.Role, roleNameMap); roleName != "" {
			caniIP.IPRole = devicetypes.IPAddressRole(roleName)
		}
		if ip.CustomFields != nil {
			caniIP.CustomFields = convCustomFields(ip.CustomFields)
		}

		result[caniID] = caniIP
	}

	return result
}

// hostFromAddress returns the host portion of a CIDR address,
// e.g. "10.0.0.1/24" → "10.0.0.1". Addresses without a mask are
// returned unchanged.
func hostFromAddress(address string) string {
	if i := strings.IndexByte(address, '/'); i >= 0 {
		return address[:i]
	}
	return address
}
