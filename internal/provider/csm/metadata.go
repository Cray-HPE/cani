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
	"fmt"
	"sort"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
)

const (
	ProviderPropertyVlanId = "vlanID"
)

type NodeMetadata struct {
	Role                 *string
	SubRole              *string
	Nid                  *int
	Alias                []string
	AdditionalProperties map[string]interface{}
}

type CabinetMetadata struct {
	HMNVlan *int
}

// TODO this might need a better home
func StringPtr(s string) *string {
	return &s
}
func IntPtr(i int) *int {
	return &i
}

func GetProviderMetadata(cHardware inventory.Hardware) (result interface{}, err error) {
	providerPropertiesRaw, ok := cHardware.ProviderProperties["csm"]
	if !ok {
		log.Debug().Any("id", cHardware.ID).Msgf("GetProviderMetadata: No CSM provider properties found")
		return nil, nil // This should be ok, as its possible as not all hardware inventory items may have CSM specific data
	}

	switch cHardware.Type {
	case hardwaretypes.Node:
		result = NodeMetadata{}
	case hardwaretypes.Cabinet:
		result = CabinetMetadata{}
	default:
		// This may be caused if new metadata structs are added, but not to this switch case
		return nil, fmt.Errorf("hardware object (%s) has unexpected provider metadata", cHardware.ID)
	}

	// Decode the Raw extra properties into a give structure
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.StringToIPHookFunc(),
		Result:     &result,
	})
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(providerPropertiesRaw)

	return result, err
}

func GetProviderMetadataT[T any](cHardware inventory.Hardware) (*T, error) {
	metadataRaw, err := GetProviderMetadata(cHardware)
	if err != nil {
		return nil, err
	}

	if metadataRaw == nil {
		log.Debug().Any("id", cHardware.ID).Msgf("GetProviderMetadataT: No metadata returned from GetProviderMetadata")
		return nil, nil
	}

	metadata, ok := metadataRaw.(T)
	if !ok {
		var expectedType T
		return nil, fmt.Errorf("unexpected provider metadata type (%T) expected (%T)", metadataRaw, expectedType)
	}
	return &metadata, nil
}

func (csm *CSM) BuildHardwareMetadata(cHardware *inventory.Hardware, rawProperties map[string]interface{}) error {
	if cHardware.ProviderProperties == nil {
		cHardware.ProviderProperties = map[string]interface{}{}
	}

	switch cHardware.Type {
	case hardwaretypes.Cabinet:
		properties := CabinetMetadata{}
		if _, exists := cHardware.ProviderProperties["csm"]; exists {
			// If one exists set it.
			if err := mapstructure.Decode(cHardware.ProviderProperties["csm"], &properties); err != nil {
				return err
			}
		}

		// Make changes to the node metadata
		// The keys of rawProperties need to match what is defined in ./cmd/cabinet/add_cabinet.go
		if vlanIDRaw, exists := rawProperties[ProviderPropertyVlanId]; exists {
			if vlanIDRaw == nil {
				properties.HMNVlan = nil
			} else {
				properties.HMNVlan = IntPtr(vlanIDRaw.(int))
			}
		}

		cHardware.ProviderProperties["csm"] = properties

		return nil
	case hardwaretypes.Node:
		// TODO do something interesting with the raw data, and convert it/validate it
		properties := NodeMetadata{} // Create an empty one
		if _, exists := cHardware.ProviderProperties["csm"]; exists {
			// If one exists set it.
			if err := mapstructure.Decode(cHardware.ProviderProperties["csm"], &properties); err != nil {
				return err
			}
		}
		// Make changes to the node metadata
		// The keys of rawProperties need to match what is defined in ./cmd/node/update_node.go
		if roleRaw, exists := rawProperties["role"]; exists {
			if roleRaw == nil {
				properties.Role = nil
			} else {
				properties.Role = StringPtr(roleRaw.(string))
			}
		}
		if subroleRaw, exists := rawProperties["subrole"]; exists {
			if subroleRaw == nil {
				properties.SubRole = nil
			} else {
				properties.SubRole = StringPtr(subroleRaw.(string))
			}
		}
		if nidRaw, exists := rawProperties["nid"]; exists {
			if nidRaw == nil {
				properties.Nid = nil
			} else {
				properties.Nid = IntPtr(nidRaw.(int))
			}
		}
		if aliasRaw, exists := rawProperties["alias"]; exists {
			if aliasRaw == nil {
				properties.Alias = nil
			} else {
				properties.Alias = []string{aliasRaw.(string)}
			}
		}

		cHardware.ProviderProperties["csm"] = properties

		return nil
	default:
		// This hardware type doesn't have metadata for it right now
		return nil
	}

}

func (csm *CSM) RecommendCabinet(inv inventory.Inventory, deviceTypeSlug string) (recommended provider.HardwareRecommendations, err error) {
	// defined by csi
	const riverStartingVlan = 1513
	const riverEndingVlan = 1769
	// var riverVlanRange = []int16{riverStartingVlan, riverEndingVlan}
	// defined by csi
	const mountainStartingVlan = 3000
	const mountainEndingVlan = 3999
	// var mountainVlanRange = []int16{mountainStartingVlan, mountainEndingVlan}

	// slice to track existing vlans
	var existingVlans = []int{}
	// slice to track existing cabinets
	var existingCabinets = []int{}

	// loop through the existing inventory to check for vlans
	log.Debug().Msg("Checking existing hardware to find recommendations")
	for _, cHardware := range inv.Hardware {
		if cHardware.ProviderProperties == nil {
			cHardware.ProviderProperties = map[string]interface{}{}
		}

		switch cHardware.Type {
		case hardwaretypes.Cabinet:
			log.Debug().Msgf("Checking %s (%s)", cHardware.Type, cHardware.ID.String())
			properties := CabinetMetadata{}
			if _, exists := cHardware.ProviderProperties[string(inventory.CSMProvider)]; exists {
				// If one exists set it.
				log.Debug().Msgf("Decoding csm properties %+v", cHardware.ProviderProperties)
				if err := mapstructure.Decode(cHardware.ProviderProperties[string(inventory.CSMProvider)], &properties); err != nil {
					return recommended, err
				}
			}

			// Make changes to the node metadata
			// The keys of rawProperties need to match what is defined in ./cmd/cabinet/add_cabinet.go
			if properties.HMNVlan != nil {
				// add it to the slice that tracks existing vlans
				existingVlans = append(existingVlans, *properties.HMNVlan)
			}

			// add the ordinal to the existing cabinets slice for choosing a new one later
			existingCabinets = append(existingCabinets, *cHardware.LocationOrdinal)

		default:
			// This function only handles cabinets
			continue
		}
	}

	var chosenOrdinal int
	chosenOrdinal, err = DetermineStartingOrdinalFromSlug(deviceTypeSlug, *csm.hardwareLibrary)
	if err != nil {
		return recommended, err
	}
	log.Debug().Msgf("chosenOrdinal %d (%s)", chosenOrdinal, deviceTypeSlug)
	// Set the cabinet location
	if len(existingCabinets) == 0 {
		// there are no cabinets yet, so set it to the provider default
		recommended.LocationOrdinal = chosenOrdinal
		log.Debug().Msgf("No cabinets found, using %d", recommended.LocationOrdinal)
	} else {
		// set the recommended cabinet number
		recommended.LocationOrdinal = nextAvailableInt(existingCabinets, chosenOrdinal)
		log.Debug().Msgf("Existing cabinets found (%v), using %d", existingCabinets, recommended.LocationOrdinal)
	}

	// Determine the hardware class based off the slug
	// This is needed to assign an approriate VLAN from the ranges defined above
	class, err := DetermineHardwareClassFromSlug(deviceTypeSlug, *csm.hardwareLibrary)
	if err != nil {
		return recommended, err
	}

	// Set the metadata vlan
	var chosenVlan int
	if len(existingCabinets) == 0 {
		// choose a starting vlan based on the class
		if class == sls_client.HardwareClassMountain || class == sls_client.HardwareClassHill {
			chosenVlan = mountainStartingVlan
		}
		if class == sls_client.HardwareClassRiver {
			chosenVlan = riverStartingVlan
		}
		log.Debug().Msgf("No cabinet VLANs found, using %d for %s %s", chosenVlan, class, hardwaretypes.Cabinet)
	} else {
		// set the recommended vlan by finding an available one from the existing
		if class == sls_client.HardwareClassMountain || class == sls_client.HardwareClassHill {
			chosenVlan = nextAvailableInt(existingVlans, mountainStartingVlan)
		}
		if class == sls_client.HardwareClassRiver {
			chosenVlan = nextAvailableInt(existingVlans, riverStartingVlan)
		}

	}

	// set the provider metadata
	recommended.ProviderMetadata = map[string]interface{}{
		// there are no vlans yet, and presumably no cabinets, so set it to 1
		ProviderPropertyVlanId: chosenVlan,
	}

	// return the recommendations
	return recommended, nil
}

func nextAvailableInt(s []int, offset int) int {
	// slice must be sorted in order to work properly
	sort.Ints(s)

	// If the slice is empty, return the offset
	if len(s) == 0 || offset < s[0] {
		return offset
	}
	// Check if the offset is in the slice
	i := sort.Search(len(s), func(i int) bool { return s[i] >= offset })
	if i < len(s) && s[i] == offset {
		for ; i < len(s); i++ {
			// if it is the last element or there is a gap to the next one
			if i == len(s)-1 || s[i+1]-s[i] > 1 {
				return s[i] + 1
			}
		}
	}
	// the offset is not in the slice, so return it
	return offset
}
