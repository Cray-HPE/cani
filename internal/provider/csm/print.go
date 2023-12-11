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
package csm

import (
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/rs/zerolog/log"
)

func (csm *CSM) PrintHardware(hw *inventory.Hardware) {
	switch hw.Type {
	case hardwaretypes.NodeBlade:
		log.Info().Msgf("UUID: %s", hw.ID)

		cabinet, _ := hw.LocationPath.GetOrdinal(hardwaretypes.Cabinet)
		log.Info().Msgf("Cabinet: %d", cabinet)

		chassis, _ := hw.LocationPath.GetOrdinal(hardwaretypes.Chassis)
		log.Info().Msgf("Chassis: %d", chassis)

		blade, _ := hw.LocationPath.GetOrdinal(hardwaretypes.NodeBlade)
		log.Info().Msgf("Blade: %d", blade)

	case hardwaretypes.Cabinet:
		log.Info().Msgf("UUID: %s", hw.ID)
		log.Info().Msgf("Cabinet Number: %d", *hw.LocationOrdinal)
		log.Info().Msgf("VLAN ID: %d", hw.ProviderMetadata["csm"]["Cabinet"].(map[string]interface{})["HMNVlan"])

	default:
		log.Info().Msgf("UUID: %s", hw.ID)
		log.Info().Msgf("Type: %s", hw.Type)
	}
}
