package inventory

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type HardwareBuildOut struct {
	ID               uuid.UUID
	ParentID         uuid.UUID
	DeviceTypeString string
	DeviceType       hardwaretypes.DeviceType
	OrdinalPath      []int
	HardwareTypePath hardwaretypes.HardwareTypePath

	// TODO perhaps the OrdinalPath and HardwareTypePath should maybe become there down struct and be paired together.
}

func (hbo *HardwareBuildOut) GetOrdinal() int {
	return hbo.OrdinalPath[len(hbo.OrdinalPath)-1]
}

func (hbo *HardwareBuildOut) LocationPathString() string {
	tokens := []string{}

	for i, token := range hbo.HardwareTypePath {
		tokens = append(tokens, fmt.Sprintf("%s:%d", token, hbo.OrdinalPath[i]))
	}

	return strings.Join(tokens, "->")
}

// TODO make this should work the inventory data structure
func GetDefaultHardwareBuildOut(l *hardwaretypes.Library, deviceTypeString string, deviceOrdinal int, parentID uuid.UUID) (results []HardwareBuildOut, err error) {
	queue := []HardwareBuildOut{
		{
			ID:               uuid.New(),
			ParentID:         parentID,
			DeviceTypeString: deviceTypeString,
			OrdinalPath:      []int{deviceOrdinal},
		},
	}

	for len(queue) != 0 {
		current := queue[0]
		queue = queue[1:]

		log.Debug().Msgf("Visiting: %s", current.DeviceTypeString)
		currentDeviceType, ok := l.DeviceTypes[current.DeviceTypeString]
		if !ok {
			return nil, fmt.Errorf("device type (%v) does not exist", current.DeviceTypeString)
		}

		// Retrieve the hardware type at this point in time, so we only lookup in the map once
		current.DeviceType = currentDeviceType
		current.HardwareTypePath = append(current.HardwareTypePath, current.DeviceType.HardwareType)

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
							fmt.Errorf("unable extract ordinal from device bay name (%s) from device type (%s)", deviceBay.Name, current.DeviceTypeString),
							err,
						)
					}
				}

				queue = append(queue, HardwareBuildOut{
					// Hardware type is deferred until when it is processed
					ID:               uuid.New(),
					ParentID:         current.ID,
					DeviceTypeString: deviceBay.Default.Slug,
					OrdinalPath:      append(current.OrdinalPath, ordinal),
					HardwareTypePath: current.HardwareTypePath,
				})
			}
		}

		results = append(results, current)
	}

	return results, nil
}
