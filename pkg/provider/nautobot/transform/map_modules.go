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

// MapModules converts Nautobot Module objects to CANI modules.
// moduleBayMap maps module-bay UUID → (parent device UUID, bay name).
func MapModules(
	raw []nautobotapi.Module,
	moduleBayMap map[uuid.UUID]moduleBayRef,
	deviceMap map[uuid.UUID]uuid.UUID,
	locationMap map[uuid.UUID]uuid.UUID,
	statusNameMap map[uuid.UUID]string,
	roleNameMap map[uuid.UUID]string,
) map[uuid.UUID]*devicetypes.CaniModuleType {
	result := make(map[uuid.UUID]*devicetypes.CaniModuleType, len(raw))

	for _, mod := range raw {
		nbID := directUUID(mod.Id)
		if nbID == uuid.Nil {
			continue
		}
		caniID := uuid.New()

		m := &devicetypes.CaniModuleType{
			ID:     caniID,
			Name:   strVal(mod.Display),
			Serial: strVal(mod.Serial),
			ObjectMeta: devicetypes.ObjectMeta{
				Status:      resolveRefName(mod.Status, statusNameMap),
				ExternalIDs: map[string]uuid.UUID{"nautobot": nbID},
			},
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
				if m.Name == "" {
					m.Name = ref.name
				}
			}
		}

		// Location.
		if mod.Location != nil {
			locNBID := tenantRefID(mod.Location)
			if caniLocID, ok := locationMap[locNBID]; ok {
				m.Location = caniLocID
			}
		}

		if mod.Role != nil {
			m.Role = resolveTenantRefName(mod.Role, roleNameMap)
		}

		if mod.CustomFields != nil {
			m.CustomFields = convCustomFields(mod.CustomFields)
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
