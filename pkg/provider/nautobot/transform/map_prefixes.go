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
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// MapPrefixes converts Nautobot Prefix objects to CANI prefixes.
// It requires:
//   - locationMap: Nautobot location UUID → CANI location UUID
//   - vlanMap: Nautobot VLAN UUID → CANI VLAN UUID
//   - statusNameMap: Nautobot status UUID → status name
//   - roleNameMap: Nautobot role UUID → role name
//
// Parent prefixes are intentionally left unset: CANI recomputes the prefix
// hierarchy from the CIDR values after import.
func MapPrefixes(
	raw []nautobotapi.Prefix,
	locationMap map[uuid.UUID]uuid.UUID,
	vlanMap map[uuid.UUID]uuid.UUID,
	statusNameMap map[uuid.UUID]string,
	roleNameMap map[uuid.UUID]string,
) map[uuid.UUID]*devicetypes.CaniPrefix {
	result := make(map[uuid.UUID]*devicetypes.CaniPrefix, len(raw))

	for _, prefix := range raw {
		nbID := directUUID(prefix.Id)
		if nbID == uuid.Nil {
			continue
		}
		caniID := uuid.New()

		caniPrefix := &devicetypes.CaniPrefix{
			ID:          caniID,
			Prefix:      prefix.Prefix,
			PrefixLen:   intVal(prefix.PrefixLength),
			IPVersion:   intVal(prefix.IpVersion),
			Description: strVal(prefix.Description),
			// TODO(nautobot-3.2): Prefix responses no longer expose a location;
			// resolve via a separate location lookup if needed.
			Location: uuid.Nil,
			ObjectMeta: devicetypes.ObjectMeta{
				Status:      resolveRefName(prefix.Status, statusNameMap),
				ExternalIDs: map[string]uuid.UUID{"nautobot": nbID},
			},
		}
		if prefix.Type != nil && prefix.Type.Value != nil {
			caniPrefix.Type = devicetypes.PrefixType(string(*prefix.Type.Value))
		}
		if vlanNBID := tenantRefID(prefix.Vlan); vlanNBID != uuid.Nil {
			if caniVLANID, ok := vlanMap[vlanNBID]; ok {
				caniPrefix.VLAN = caniVLANID
			}
		}
		if roleName := resolveTenantRefName(prefix.Role, roleNameMap); roleName != "" {
			caniPrefix.Role = roleName
		}
		if prefix.CustomFields != nil {
			caniPrefix.CustomFields = convCustomFields(prefix.CustomFields)
		}

		result[caniID] = caniPrefix
	}

	return result
}
