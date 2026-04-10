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

	for _, item := range raw {
		nbID := directUUID(item.Id)
		if nbID == uuid.Nil {
			continue
		}
		caniID := uuid.New()

		fru := &devicetypes.CaniFruType{
			ID:          caniID,
			Name:        item.Name,
			Label:       strVal(item.Label),
			Serial:      strVal(item.Serial),
			Description: strVal(item.Description),
			PartNumber:  strVal(item.PartId),
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
			fru.CustomFields = *item.CustomFields
		}

		result[caniID] = fru
	}

	return result
}
