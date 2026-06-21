package transform

import (
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// MapDevices converts Nautobot Device objects to CANI devices.
// It requires:
//   - rackMap: Nautobot rack UUID → CANI rack UUID
//   - locationMap: Nautobot location UUID → CANI location UUID
//   - deviceTypeMap: Nautobot device-type UUID → DeviceType object
//   - ifacesByDevice: Nautobot device UUID → list of interfaces
//   - statusNameMap: Nautobot status UUID → status name
//   - roleNameMap: Nautobot role UUID → role name
//
// Returns devices and a mapping from Nautobot device UUID → CANI device UUID.
func MapDevices(
	raw []nautobotapi.Device,
	rackMap map[uuid.UUID]uuid.UUID,
	locationMap map[uuid.UUID]uuid.UUID,
	deviceTypeMap map[uuid.UUID]nautobotapi.DeviceType,
	ifacesByDevice map[uuid.UUID][]nautobotapi.Interface,
	statusNameMap map[uuid.UUID]string,
	roleNameMap map[uuid.UUID]string,
) (map[uuid.UUID]*devicetypes.CaniDeviceType, map[uuid.UUID]uuid.UUID) {
	result := make(map[uuid.UUID]*devicetypes.CaniDeviceType, len(raw))
	nbToCani := make(map[uuid.UUID]uuid.UUID, len(raw))

	for _, dev := range raw {
		nbID := directUUID(dev.Id)
		if nbID == uuid.Nil {
			continue
		}
		caniID := uuid.New()
		nbToCani[nbID] = caniID

		caniDev := &devicetypes.CaniDeviceType{
			ID:         caniID,
			Name:       strVal(dev.Name),
			ObjectMeta: devicetypes.ObjectMeta{Status: resolveRefName(dev.Status, statusNameMap), ExternalIDs: map[string]uuid.UUID{"nautobot": nbID}},
			Serial:     strVal(dev.Serial),
			Comments:   strVal(dev.Comments),
		}
		if dev.AssetTag != nil {
			caniDev.AssetTag = *dev.AssetTag
		}

		// Resolve device type → slug, model, manufacturer, uHeight, isFullDepth.
		dtNBID := refIDVal(dev.DeviceType)
		if dt, ok := deviceTypeMap[dtNBID]; ok {
			caniDev.Model = dt.Model
			if dt.UHeight != nil {
				caniDev.UHeight = *dt.UHeight
			}
			if dt.IsFullDepth != nil {
				caniDev.IsFullDepth = *dt.IsFullDepth
			}
			// Resolve slug: try NaturalSlug first, then fall back to model-based lookup.
			caniDev.Slug = resolveDeviceSlug(dt)
		}

		// Resolve rack → Parent.
		if dev.Rack != nil {
			rackNBID := tenantRefID(dev.Rack)
			if caniRackID, ok := rackMap[rackNBID]; ok {
				caniDev.Parent = caniRackID
			}
		}

		// Resolve explicit location.
		locNBID := refIDVal(dev.Location)
		if caniLocID, ok := locationMap[locNBID]; ok {
			caniDev.Location = caniLocID
		}

		// Rack position and face.
		// Nautobot requires rack, position, and face to be set together;
		// default face to "front" when a position is present.
		if dev.Position != nil {
			caniDev.RackPosition = *dev.Position
		}
		if dev.Face != nil && dev.Face.Value != nil {
			caniDev.Face = string(*dev.Face.Value)
		}
		if caniDev.RackPosition > 0 && caniDev.Face == "" {
			caniDev.Face = "front"
		}

		// Map interfaces from the pre-grouped interface data.
		if ifaces, ok := ifacesByDevice[nbID]; ok {
			for _, iface := range ifaces {
				mgmt := false
				if iface.MgmtOnly != nil {
					mgmt = *iface.MgmtOnly
				}
				ifaceID := directUUID(iface.Id)
				if ifaceID == uuid.Nil {
					ifaceID = uuid.New()
				}
				ifaceType := ""
				if iface.Type.Value != nil {
					ifaceType = string(*iface.Type.Value)
				}
				spec := devicetypes.InterfaceSpec{
					ID:         ifaceID,
					Name:       iface.Name,
					Type:       devicetypes.InterfacesElemType(ifaceType),
					Label:      strVal(iface.Label),
					MacAddress: strVal(iface.MacAddress),
					MgmtOnly:   &mgmt,
				}
				caniDev.Interfaces = append(caniDev.Interfaces, spec)
			}
		}

		if roleName := resolveRefName(dev.Role, roleNameMap); roleName != "" {
			caniDev.Role = roleName
		}

		if dev.CustomFields != nil {
			caniDev.CustomFields = *dev.CustomFields
		}

		result[caniID] = caniDev
	}

	return result, nbToCani
}

// GroupInterfacesByDevice groups interfaces by their parent device Nautobot UUID.
func GroupInterfacesByDevice(ifaces []nautobotapi.Interface) map[uuid.UUID][]nautobotapi.Interface {
	result := make(map[uuid.UUID][]nautobotapi.Interface)
	for _, iface := range ifaces {
		if iface.Device == nil {
			continue
		}
		devID := tenantRefID(iface.Device)
		if devID != uuid.Nil {
			result[devID] = append(result[devID], iface)
		}
	}
	return result
}

// BuildDeviceTypeMap creates a lookup from Nautobot device-type UUID → DeviceType.
func BuildDeviceTypeMap(dts []nautobotapi.DeviceType) map[uuid.UUID]nautobotapi.DeviceType {
	m := make(map[uuid.UUID]nautobotapi.DeviceType, len(dts))
	for _, dt := range dts {
		id := directUUID(dt.Id)
		if id != uuid.Nil {
			m[id] = dt
		}
	}
	return m
}

// resolveDeviceSlug attempts to find a matching slug in the cani device type library.
// It tries: 1) NaturalSlug as-is, 2) model-based lookup across all library entries.
// Returns the library slug if found, or empty string if no match.
func resolveDeviceSlug(dt nautobotapi.DeviceType) string {
	// Try NaturalSlug directly.
	if dt.NaturalSlug != nil {
		if _, ok := devicetypes.GetBySlug(*dt.NaturalSlug); ok {
			return *dt.NaturalSlug
		}
	}

	// Try matching by model name (case-insensitive) across all library entries.
	modelLower := strings.ToLower(dt.Model)
	for _, libDT := range devicetypes.All() {
		if strings.ToLower(libDT.Model) == modelLower {
			return libDT.Slug
		}
	}

	return ""
}
