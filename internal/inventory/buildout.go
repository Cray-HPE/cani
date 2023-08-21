/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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

	ExistingHardware *Hardware
}

func (hbo *HardwareBuildOut) GetOrdinal() int {
	ordinalPath := hbo.LocationPath.GetOrdinalPath()
	return ordinalPath[len(ordinalPath)-1]
}

func GenerateDefaultHardwareBuildOut(l *hardwaretypes.Library, deviceTypeSlug string, deviceOrdinal int, parentHardware Hardware) (results []HardwareBuildOut, err error) {
	return GenerateHardwareBuildOut(l, GenerateHardwareBuildOutOpts{
		DeviceTypeSlug: deviceTypeSlug,
		DeviceOrdinal:  deviceOrdinal,
		DeviceID:       uuid.Nil, // Generate one, TODO maybe allocate the UUID here?
		ParentHardware: parentHardware,
	})
}

type GenerateHardwareBuildOutOpts struct {
	DeviceTypeSlug string
	DeviceOrdinal  int
	DeviceID       uuid.UUID // Optional: If specified use this for the top level hardware object created, otherwise if the UUID is uuid.Nil an UUID is generated if

	ParentHardware Hardware

	ExistingDescendentHardware []Hardware
}

// TODO make this should work the inventory data structure
func GenerateHardwareBuildOut(l *hardwaretypes.Library, opts GenerateHardwareBuildOutOpts) (results []HardwareBuildOut, err error) {
	//
	// Build up existing hardware lookup map
	//
	existingDescendentHardware := map[string]Hardware{}
	for _, hardware := range opts.ExistingDescendentHardware {
		existingDescendentHardware[hardware.LocationPath.String()] = hardware
	}

	//
	// Build out hardware
	//
	var topLevelHardwareID uuid.UUID
	if opts.DeviceID != uuid.Nil {
		topLevelHardwareID = opts.DeviceID
	} else {
		topLevelHardwareID = uuid.New()
	}

	queue := []HardwareBuildOut{
		{
			ID:             topLevelHardwareID,
			ParentID:       opts.ParentHardware.ID,
			DeviceTypeSlug: opts.DeviceTypeSlug,
			DeviceOrdinal:  opts.DeviceOrdinal,

			LocationPath: opts.ParentHardware.LocationPath, // The loop below will add on the required location token for this devices path.
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

		// Override hardware ID if there is a piece of hardware already exists
		// This override should be ok to do here, as no child hardware in the queue should have added
		// yet, as that happens in the loop below.
		if existingHardware, exists := existingDescendentHardware[current.LocationPath.String()]; exists {
			current.ID = existingHardware.ID
			current.ExistingHardware = &existingHardware
		}

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
