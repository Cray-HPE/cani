package transform

import (
	"fmt"
	"log"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/google/uuid"
)

// transformSystemDevices creates devices from system CSV device records.
func transformSystemDevices(data *import_.SystemCSV, result *devicetypes.TransformResult, inv *devicetypes.Inventory, racksByName map[string]uuid.UUID) error {
	for _, rec := range data.Devices {
		rec = data.ApplyDefaults(rec)
		if rec.PartNumber == "" {
			return fmt.Errorf("device %q missing PartNumber (slug or part number)", rec.Name)
		}

		for i := 0; i < rec.Qty; i++ {
			id := uuid.New()
			name := rec.Name
			if rec.Qty > 1 && name != "" {
				name = fmt.Sprintf("%s-%d", name, i+1)
			}

			device := &devicetypes.CaniDeviceType{
				ID:         id,
				Name:       name,
				Slug:       rec.PartNumber,
				PartNumber: rec.PartNumber,
				Serial:     rec.Serial,
				ObjectMeta: devicetypes.ObjectMeta{
					Status: rec.Status,
					Role:   rec.Role,
				},
			}

			// Populate from device type library
			if dt, ok := devicetypes.GetBySlug(rec.PartNumber); ok {
				populateFromDeviceType(device, &dt)
			} else if dt, ok := devicetypes.GetByPartNumber(rec.PartNumber); ok {
				populateFromDeviceType(device, &dt)
			}

			// Place in rack
			if rec.Rack != "" {
				rackID, ok := racksByName[rec.Rack]
				if !ok {
					// Try existing inventory
					rackID, ok = findRackByName(inv, rec.Rack)
				}
				if !ok {
					return fmt.Errorf("device %q references unknown rack %q", name, rec.Rack)
				}
				device.Parent = rackID
				device.Rack = rackID

				if rec.Position > 0 {
					device.RackPosition = rec.Position
					face := rec.Face
					if face == "" {
						face = devicetypes.FaceFront
					}
					device.Face = face

					if rack, ok := inv.Racks[rackID]; ok {
						height := device.UHeight
						if height < 1 {
							height = 1
						}
						if !rack.PlaceDevice(id, rec.Position, height, face, device.IsFullDepth) {
							log.Printf("WARN: device %q cannot fit at U%d in rack %q", name, rec.Position, rec.Rack)
						}
					}
				}
			}

			result.Devices[id] = device
			inv.Devices[id] = device
		}
	}

	return nil
}
