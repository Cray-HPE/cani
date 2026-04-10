package transform

import (
	"strconv"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// MapRacks converts Nautobot Rack objects to CANI racks.
// locationMap maps Nautobot location UUID → CANI location UUID.
// Returns the racks and a mapping from Nautobot rack UUID → CANI rack UUID.
func MapRacks(raw []nautobotapi.Rack, locationMap map[uuid.UUID]uuid.UUID) (map[uuid.UUID]*devicetypes.CaniRackType, map[uuid.UUID]uuid.UUID) {
	result := make(map[uuid.UUID]*devicetypes.CaniRackType, len(raw))
	nbToCani := make(map[uuid.UUID]uuid.UUID, len(raw))

	for _, rack := range raw {
		nbID := directUUID(rack.Id)
		if nbID == uuid.Nil {
			continue
		}
		caniID := uuid.New()
		nbToCani[nbID] = caniID

		caniRack := &devicetypes.CaniRackType{
			ID:         caniID,
			Name:       rack.Name,
			ObjectMeta: devicetypes.ObjectMeta{Status: strVal(rack.Status.Url), ExternalIDs: map[string]uuid.UUID{"nautobot": nbID}},
			Serial:     strVal(rack.Serial),
			FacilityId: strVal(rack.FacilityId),
			UHeight:    intVal(rack.UHeight),
			OuterWidth: intVal(rack.OuterWidth),
			OuterDepth: intVal(rack.OuterDepth),
			Comments:   strVal(rack.Comments),
		}

		if rack.AssetTag != nil {
			caniRack.AssetTag = *rack.AssetTag
		}

		// Resolve location.
		nbLocID := refIDVal(rack.Location)
		if caniLocID, ok := locationMap[nbLocID]; ok {
			caniRack.Location = caniLocID
		}

		// Map role from tenant reference.
		if rack.Role != nil {
			caniRack.Role = strVal(rack.Role.Url)
		}

		// Map outer unit.
		if rack.OuterUnit != nil && rack.OuterUnit.Value != nil {
			caniRack.OuterUnit = string(*rack.OuterUnit.Value)
		}

		// Map width.
		if rack.Width != nil && rack.Width.Value != nil {
			caniRack.Width = strconv.Itoa(int(*rack.Width.Value))
		}

		// Map rack type.
		if rack.Type != nil && rack.Type.Value != nil {
			caniRack.RackType = string(*rack.Type.Value)
		}

		if rack.CustomFields != nil {
			caniRack.CustomFields = *rack.CustomFields
		}

		result[caniID] = caniRack
	}

	return result, nbToCani
}
