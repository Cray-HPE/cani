package transform

import (
	"fmt"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/google/uuid"
)

// transformDcimRacks creates racks from DCIM CSV rack records.
// Returns a map of rack name → UUID for device parenting.
func transformDcimRacks(data *import_.DcimCSV, result *devicetypes.TransformResult, inv *devicetypes.Inventory, locationsByName map[string]uuid.UUID) (map[string]uuid.UUID, error) {
	racksByName := make(map[string]uuid.UUID)

	for _, rec := range data.Racks {
		rec = data.ApplyDefaults(rec)
		if rec.PartNumber == "" {
			return nil, fmt.Errorf("rack %q missing PartNumber (slug or part number)", rec.Name)
		}

		// Lookup rack type from library
		slug := rec.PartNumber
		uHeight := 48
		var partNumber, manufacturer, model string

		if rt, ok := devicetypes.GetRackTypeBySlug(rec.PartNumber); ok {
			slug = rt.Slug
			uHeight = rt.UHeight
			partNumber = rt.PartNumber
			manufacturer = rt.Manufacturer
			model = rt.Model
		} else if rt, ok := devicetypes.GetRackTypeByPartNumber(rec.PartNumber); ok {
			slug = rt.Slug
			uHeight = rt.UHeight
			partNumber = rt.PartNumber
			manufacturer = rt.Manufacturer
			model = rt.Model
		}

		for i := 0; i < rec.Qty; i++ {
			id := uuid.New()
			name := rec.Name
			if rec.Qty > 1 && name != "" {
				name = fmt.Sprintf("%s-%d", name, i+1)
			}

			rack := &devicetypes.CaniRackType{
				ID:           id,
				Name:         name,
				Slug:         slug,
				PartNumber:   partNumber,
				Manufacturer: manufacturer,
				Model:        model,
				UHeight:      uHeight,
				ObjectMeta:   devicetypes.ObjectMeta{Status: rec.Status},
				Devices:      []uuid.UUID{},
			}

			// Assign location
			if rec.Location != "" {
				locID, ok := locationsByName[rec.Location]
				if !ok {
					locID, ok = findLocationByName(inv, rec.Location)
				}
				if ok {
					rack.Location = locID
				}
			}

			result.Racks[id] = rack
			inv.Racks[id] = rack
			if name != "" {
				racksByName[name] = id
			}
		}
	}

	return racksByName, nil
}

// findRackByName searches the inventory for a rack with the given name.
func findRackByName(inv *devicetypes.Inventory, name string) (uuid.UUID, bool) {
	for id, rack := range inv.Racks {
		if rack.Name == name {
			return id, true
		}
	}
	return uuid.Nil, false
}
