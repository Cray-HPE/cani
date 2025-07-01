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
package hpcm

import (
	"fmt"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func (hpcm *Hpcm) TranslateCmdb() (translated map[uuid.UUID]*inventory.Hardware, err error) {
	translated = make(map[uuid.UUID]*inventory.Hardware, 0)
	for _, systemgroup := range hpcm.Cmdb.SystemGroups {
		log.Debug().Msgf("Translating HPCM systemgroup: %+v", systemgroup.Name)
		for _, node := range systemgroup.Nodes {
			log.Debug().Msgf("Translating HPCM node: %+v", node.Name)
			hw := &inventory.Hardware{}
			hw.Name = node.Name

			// convert the hpcm uuid string to a uuid type
			u, err := uuid.Parse(node.Uuid)
			if err != nil {
				return translated, fmt.Errorf("unable to translate HPCM UUID to CANI UUID: %v", err)
			}
			hw.ID = u

			// translate the hpcm type to a hardwaretypes library type
			translatedType, err := hpcmTypeToCaniHardwareType(node.Type_)
			if err != nil {
				return translated, err
			}
			hw.Type = translatedType

			hw.Vendor = getVendor(node)
			// FIXME: does not conform to existing slugs
			hw.DeviceTypeSlug = getModel(node)

			var lp inventory.LocationPath
			lp, err = xnameToLocationPath(hw.Name)
			if err != nil {
				return translated, err
			}

			if lp == nil {
				// convert hpcm geolocation to cani geolocation
				lp, err = hpcmLocToCaniLoc(hw.Type, node.Location)
				if err != nil {
					return translated, err
				}
			}
			hw.LocationPath = lp
			// architecture is a string in both, so map it directly
			hw.Architecture = node.Platform.Architecture

			if hw.ProviderMetadata == nil {
				hw.ProviderMetadata = make(map[inventory.Provider]inventory.ProviderMetadataRaw, 0)
			}

			_, exists := hw.ProviderMetadata["hpcm"]
			if !exists {
				hw.ProviderMetadata["hpcm"] = make(inventory.ProviderMetadataRaw, 0)
			}

			providerMetadata, err := extractProviderMetadata(node)
			if err != nil {
				return translated, err
			}

			hw.ProviderMetadata["hpcm"] = providerMetadata

			_, ok := translated[hw.ID]
			if !ok {
				translated[hw.ID] = hw
			}
		}
	}
	return translated, nil
}
