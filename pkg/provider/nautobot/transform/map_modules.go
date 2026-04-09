package transform

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// MapModules converts Nautobot Module objects to CANI modules.
// moduleBayMap maps module-bay UUID → (parent device UUID, bay name).
func MapModules(
	raw []nautobotapi.Module,
	moduleBayMap map[uuid.UUID]moduleBayRef,
	deviceMap map[uuid.UUID]uuid.UUID,
) map[uuid.UUID]*devicetypes.CaniModuleType {
	result := make(map[uuid.UUID]*devicetypes.CaniModuleType, len(raw))

	for _, mod := range raw {
		nbID := directUUID(mod.Id)
		if nbID == uuid.Nil {
			continue
		}
		caniID := uuid.New()

		m := &devicetypes.CaniModuleType{
			ID:         caniID,
			Serial:     strVal(mod.Serial),
			ObjectMeta: devicetypes.ObjectMeta{Status: strVal(mod.Status.Url)},
		}

		if mod.AssetTag != nil {
			m.AssetTag = *mod.AssetTag
		}

		// Resolve parent device and module bay name.
		if mod.ParentModuleBay != nil {
			bayNBID := tenantRefID(mod.ParentModuleBay)
			if ref, ok := moduleBayMap[bayNBID]; ok {
				if caniDevID, ok2 := deviceMap[ref.deviceID]; ok2 {
					m.ParentDevice = caniDevID
				}
				m.ModuleBayName = ref.name
			}
		}

		// Location.
		if mod.Location != nil {
			m.Location = tenantRefID(mod.Location)
		}

		if mod.Role != nil {
			roleID := tenantRefID(mod.Role)
			if roleID != uuid.Nil {
				m.Role = roleID.String()
			}
		}

		if mod.CustomFields != nil {
			m.CustomFields = *mod.CustomFields
		}

		result[caniID] = m
	}

	return result
}

// moduleBayRef holds resolved info about a module bay.
type moduleBayRef struct {
	deviceID uuid.UUID // Nautobot device UUID
	name     string
}

// BuildModuleBayMap creates a lookup from Nautobot module-bay UUID →
// (parent-device UUID, bay name).
func BuildModuleBayMap(bays []nautobotapi.ModuleBay) map[uuid.UUID]moduleBayRef {
	m := make(map[uuid.UUID]moduleBayRef, len(bays))
	for _, bay := range bays {
		bayID := directUUID(bay.Id)
		if bayID == uuid.Nil {
			continue
		}
		devID := uuid.Nil
		if bay.ParentDevice != nil {
			devID = tenantRefID(bay.ParentDevice)
		}
		m[bayID] = moduleBayRef{deviceID: devID, name: bay.Name}
	}
	return m
}
