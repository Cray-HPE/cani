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

// MapFrus converts Nautobot InventoryItem objects to CANI FRUs.
func MapFrus(
	raw []nautobotapi.InventoryItem,
	deviceMap map[uuid.UUID]uuid.UUID,
) map[uuid.UUID]*devicetypes.CaniFruType {
	result := make(map[uuid.UUID]*devicetypes.CaniFruType, len(raw))
	nbToCani := make(map[uuid.UUID]uuid.UUID, len(raw))

	for _, item := range raw {
		nbID := directUUID(item.Id)
		if nbID == uuid.Nil {
			continue
		}
		caniID := uuid.New()
		nbToCani[nbID] = caniID

		fru := &devicetypes.CaniFruType{
			ID:          caniID,
			Name:        item.Name,
			Label:       strVal(item.Label),
			Serial:      strVal(item.Serial),
			Description: strVal(item.Description),
			PartNumber:  strVal(item.PartId),
			ObjectMeta: devicetypes.ObjectMeta{
				ExternalIDs: map[string]uuid.UUID{"nautobot": nbID},
			},
		}

		if item.Discovered != nil {
			fru.Discovered = *item.Discovered
		}

		if item.AssetTag != nil {
			fru.AssetTag = *item.AssetTag
		}

		// Resolve device reference.
		devNBID := refIDVal(item.Device)
		if devNBID != uuid.Nil {
			if caniDevID, ok := deviceMap[devNBID]; ok {
				fru.Device = caniDevID
			}
		}

		// Resolve parent (another inventory item).
		if item.Parent != nil {
			parentNBID := tenantRefID(item.Parent)
			if parentNBID != uuid.Nil {
				fru.Parent = parentNBID // will be resolved in a second pass if needed
			}
		}

		// Manufacturer.
		if item.Manufacturer != nil {
			mfgID := tenantRefID(item.Manufacturer)
			if mfgID != uuid.Nil {
				fru.Manufacturer = mfgID.String()
			}
		}

		if item.CustomFields != nil {
			fru.CustomFields = convCustomFields(item.CustomFields)
		}

		result[caniID] = fru
	}

	for _, fru := range result {
		if mapped, ok := nbToCani[fru.Parent]; ok {
			fru.Parent = mapped
		} else if fru.Parent != uuid.Nil {
			fru.Parent = uuid.Nil
		}
	}

	return result
}
