package inventory

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type HardwareBuildOut struct {
	ID             uuid.UUID
	ParentID       uuid.UUID
	DeviceTypeSlug string
	DeviceType     hardwaretypes.DeviceType
	DeviceOrdinal  int

	LocationPath LocationPath

	// TODO perhaps the OrdinalPath and HardwareTypePath should maybe become there down struct and be paired together.
}

func (hbo *HardwareBuildOut) GetOrdinal() int {
	ordinalPath := hbo.LocationPath.GetOrdinalPath()
	return ordinalPath[len(ordinalPath)-1]
}

// TODO make this should work the inventory data structure
func GetDefaultHardwareBuildOut(l *hardwaretypes.Library, deviceTypeSlug string, deviceOrdinal int, parentID uuid.UUID) (results []HardwareBuildOut, err error) {
	queue := []HardwareBuildOut{
		{
			ID:             uuid.New(),
			ParentID:       parentID,
			DeviceTypeSlug: deviceTypeSlug,
			DeviceOrdinal:  deviceOrdinal,
		},
	}

	for len(queue) != 0 {
		current := queue[0]
		queue = queue[1:]

		log.Debug().Msgf("Visiting: %s", current.DeviceTypeSlug)
		currentDeviceType, ok := l.DeviceTypes[current.DeviceTypeSlug]
		if !ok {
			return nil, fmt.Errorf("device type (%v) does not exist", current.DeviceTypeSlug)
		}

		// Retrieve the hardware type at this point in time, so we only lookup in the map once
		current.DeviceType = currentDeviceType
		current.LocationPath = append(current.LocationPath, LocationToken{
			HardwareType: current.DeviceType.HardwareType,
			Ordinal:      current.DeviceOrdinal,
		})

		for _, deviceBay := range currentDeviceType.DeviceBays {
			log.Debug().Msgf("  Device bay: %s", deviceBay.Name)
			if deviceBay.Default != nil {
				log.Debug().Msgf("    Default: %s", deviceBay.Default.Slug)

				// Extract the ordinal
				// This is one way of going about, but it assumes that each name has a number
				// There are two other ways to consider:
				// - Embed an actual ordinal number in the yaml files
				// - Get all of the device base with that type, and then sort them lexicographically. This is how HSM does it, but assumes the names can be sorted in a predictable order
				r := regexp.MustCompile(`\d+`)
				match := r.FindString(deviceBay.Name)

				var ordinal int
				if match != "" {
					ordinal, err = strconv.Atoi(match)
					if err != nil {
						return nil, errors.Join(
							fmt.Errorf("unable extract ordinal from device bay name (%s) from device type (%s)", deviceBay.Name, current.DeviceTypeSlug),
							err,
						)
					}
				}

				queue = append(queue, HardwareBuildOut{
					// Hardware type is deferred until when it is processed
					ID:             uuid.New(),
					ParentID:       current.ID,
					DeviceTypeSlug: deviceBay.Default.Slug,
					DeviceOrdinal:  ordinal,
					LocationPath:   current.LocationPath,
				})
			}
		}

		results = append(results, current)
	}

	return results, nil
}
