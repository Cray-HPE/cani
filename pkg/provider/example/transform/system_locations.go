package transform

import (
	"fmt"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/google/uuid"
)

// transformSystemLocations creates locations from system CSV location records.
// Returns a map of location name → UUID for rack parenting.
func transformSystemLocations(data *import_.SystemCSV, result *devicetypes.TransformResult, inv *devicetypes.Inventory) (map[string]uuid.UUID, error) {
	locationsByName := make(map[string]uuid.UUID)

	for _, rec := range data.Locations {
		rec = data.ApplyDefaults(rec)
		if rec.Name == "" {
			return nil, fmt.Errorf("location record missing Name")
		}
		locType := rec.LocationType
		if locType == "" {
			locType = rec.Role
		}
		if locType == "" {
			return nil, fmt.Errorf("location %q missing LocationType (e.g. dc, level, section)", rec.Name)
		}

		id := uuid.New()
		contentTypes := parseContentTypes(rec.ContentTypes)

		loc := &devicetypes.CaniLocationType{
			ID:           id,
			Name:         rec.Name,
			LocationType: locType,
			ContentTypes: contentTypes,
			ObjectMeta:   devicetypes.ObjectMeta{Status: rec.Status},
		}

		// Resolve parent by name
		if rec.Location != "" {
			parentID, ok := locationsByName[rec.Location]
			if !ok {
				parentID, ok = findLocationByName(inv, rec.Location)
			}
			if !ok {
				return nil, fmt.Errorf("location %q references unknown parent %q", rec.Name, rec.Location)
			}
			loc.Parent = parentID
		}

		result.Locations[id] = loc
		inv.Locations[id] = loc
		locationsByName[rec.Name] = id
	}

	return locationsByName, nil
}

// findLocationByName searches existing inventory for a location by name.
func findLocationByName(inv *devicetypes.Inventory, name string) (uuid.UUID, bool) {
	for id, loc := range inv.Locations {
		if loc.Name == name {
			return id, true
		}
	}
	return uuid.Nil, false
}
