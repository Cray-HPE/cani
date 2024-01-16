/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
	"fmt"

	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func GenerateDefaultHardwareBuildOutNoGeoloc(l *hardwaretypes.Library, opts GenerateHardwareBuildOutOpts) (results []HardwareBuildOut, err error) {

	queue := []HardwareBuildOut{
		{
			ID:             uuid.New(),
			DeviceTypeSlug: opts.DeviceTypeSlug,
		},
	}

	for len(queue) != 0 {
		current := queue[0]
		queue = queue[1:]

		log.Trace().Msgf("Visiting: %s", current.DeviceTypeSlug)
		currentDeviceType, ok := l.DeviceTypes[current.DeviceTypeSlug]
		if !ok {
			return nil, fmt.Errorf("device type (%v) does not exist", current.DeviceTypeSlug)
		}

		// Retrieve the hardware type at this point in time, so we only lookup in the map once
		current.DeviceType = currentDeviceType

		for _, deviceBay := range currentDeviceType.DeviceBays {
			log.Trace().Msgf("  Device bay: %s", deviceBay.Name)
			if deviceBay.Default != nil {
				log.Trace().Msgf("    Default: %s", deviceBay.Default.Slug)

				queue = append(queue, HardwareBuildOut{
					ID: uuid.New(),
				})
			}
		}

		results = append(results, current)
	}

	return results, nil
}
