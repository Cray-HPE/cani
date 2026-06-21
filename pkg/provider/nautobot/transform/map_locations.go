package transform

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// MapLocations converts Nautobot Location objects to CANI locations.
// It builds a mapping from Nautobot location UUID → CANI location UUID.
func MapLocations(raw []nautobotapi.Location, statusNameMap map[uuid.UUID]string) (map[uuid.UUID]*devicetypes.CaniLocationType, map[uuid.UUID]uuid.UUID) {
	result := make(map[uuid.UUID]*devicetypes.CaniLocationType, len(raw))
	// nbToCani maps Nautobot UUID → CANI UUID for cross-referencing.
	nbToCani := make(map[uuid.UUID]uuid.UUID, len(raw))

	for _, loc := range raw {
		nbID := directUUID(loc.Id)
		if nbID == uuid.Nil {
			continue
		}
		caniID := uuid.New()
		nbToCani[nbID] = caniID

		caniLoc := &devicetypes.CaniLocationType{
			ID:              caniID,
			Name:            loc.Name,
			LocationType:    strVal(loc.LocationType.Url), // will be enriched below
			ObjectMeta:      devicetypes.ObjectMeta{Status: resolveRefName(loc.Status, statusNameMap), ExternalIDs: map[string]uuid.UUID{"nautobot": nbID}},
			Description:     strVal(loc.Description),
			Facility:        strVal(loc.Facility),
			PhysicalAddress: strVal(loc.PhysicalAddress),
			Latitude:        strVal(loc.Latitude),
			Longitude:       strVal(loc.Longitude),
			ContactName:     strVal(loc.ContactName),
			ContactEmail:    strVal(loc.ContactEmail),
			ContactPhone:    strVal(loc.ContactPhone),
			TimeZone:        strVal(loc.TimeZone),
			Asn:             loc.Asn,
			Comments:        strVal(loc.Comments),
		}

		// Extract location type display name if available.
		if loc.LocationType.Id != nil {
			// Best effort: use the URL as-is or just store the name.
			// The name isn't directly in BulkWritableCableRequestStatus,
			// but the fixture data provides it.
		}

		// Map parent reference.
		if loc.Parent != nil {
			parentNBID := tenantRefID(loc.Parent)
			caniLoc.Parent = parentNBID // temporary; resolved below
		}

		if loc.CustomFields != nil {
			caniLoc.CustomFields = *loc.CustomFields
		}

		result[caniID] = caniLoc
	}

	// Resolve parent UUIDs from Nautobot → CANI mapping.
	for _, loc := range result {
		if loc.Parent != uuid.Nil {
			if caniParent, ok := nbToCani[loc.Parent]; ok {
				loc.Parent = caniParent
			} else {
				loc.Parent = uuid.Nil // parent not in our import set
			}
		}
	}

	return result, nbToCani
}
