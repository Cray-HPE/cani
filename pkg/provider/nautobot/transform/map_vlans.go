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

// MapVLANs converts Nautobot VLAN objects to CANI VLANs.
// It requires:
//   - locationMap: Nautobot location UUID → CANI location UUID
//   - statusNameMap: Nautobot status UUID → status name
//   - roleNameMap: Nautobot role UUID → role name
//
// Returns the VLANs keyed by CANI UUID and a mapping from Nautobot VLAN
// UUID → CANI VLAN UUID, which prefixes use to resolve their VLAN association.
func MapVLANs(
	raw []nautobotapi.VLAN,
	locationMap map[uuid.UUID]uuid.UUID,
	statusNameMap map[uuid.UUID]string,
	roleNameMap map[uuid.UUID]string,
) (map[uuid.UUID]*devicetypes.CaniVLAN, map[uuid.UUID]uuid.UUID) {
	result := make(map[uuid.UUID]*devicetypes.CaniVLAN, len(raw))
	nbToCani := make(map[uuid.UUID]uuid.UUID, len(raw))

	for _, vlan := range raw {
		nbID := directUUID(vlan.Id)
		if nbID == uuid.Nil {
			continue
		}
		caniID := uuid.New()
		nbToCani[nbID] = caniID

		caniVLAN := &devicetypes.CaniVLAN{
			ID:          caniID,
			VID:         vlan.Vid,
			Name:        vlan.Name,
			Description: strVal(vlan.Description),
			Location:    firstLocation(vlan.Locations, locationMap),
			ObjectMeta: devicetypes.ObjectMeta{
				Status:      resolveRefName(vlan.Status, statusNameMap),
				ExternalIDs: map[string]uuid.UUID{"nautobot": nbID},
			},
		}
		if roleName := resolveTenantRefName(vlan.Role, roleNameMap); roleName != "" {
			caniVLAN.Role = roleName
		}
		if vlan.CustomFields != nil {
			caniVLAN.CustomFields = *vlan.CustomFields
		}

		result[caniID] = caniVLAN
	}

	return result, nbToCani
}
