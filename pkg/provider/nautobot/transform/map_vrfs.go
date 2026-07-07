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

// MapVRFs converts Nautobot VRF objects to CANI VRFs.
// statusNameMap resolves the Nautobot status UUID to its name.
func MapVRFs(
	raw []nautobotapi.VRF,
	statusNameMap map[uuid.UUID]string,
) map[uuid.UUID]*devicetypes.CaniVRF {
	result := make(map[uuid.UUID]*devicetypes.CaniVRF, len(raw))

	for _, vrf := range raw {
		nbID := directUUID(vrf.Id)
		if nbID == uuid.Nil {
			continue
		}
		caniID := uuid.New()

		caniVRF := &devicetypes.CaniVRF{
			ID:          caniID,
			Name:        vrf.Name,
			RD:          strVal(vrf.Rd),
			Description: strVal(vrf.Description),
			ObjectMeta: devicetypes.ObjectMeta{
				Status:      resolveTenantRefName(vrf.Status, statusNameMap),
				ExternalIDs: map[string]uuid.UUID{"nautobot": nbID},
			},
		}
		if vrf.CustomFields != nil {
			caniVRF.CustomFields = *vrf.CustomFields
		}

		result[caniID] = caniVRF
	}

	return result
}
