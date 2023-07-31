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
	"errors"
	"fmt"
	"sort"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/Cray-HPE/cani/pkg/pointers"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
)

const (
	ProviderPropertyVlanId  = "VlanID"
	ProviderPropertyRole    = "Role"
	ProviderPropertySubRole = "SubRole"
	ProviderPropertyAlias   = "Alias"
	ProviderPropertyNID     = "NID"
)

// NOTE: When adding new Metadata structure make sure to add them to the MetadataStructTagSuite test suite
// in metadata_test.go

type Metadata struct {
	Cabinet *CabinetMetadata `json:"Cabinet,omitempty" mapstructure:"Cabinet,omitempty"`
	Node    *NodeMetadata    `json:"Node,omitempty" mapstructure:"Node,omitempty"`
}

type NodeMetadata struct {
	Role                 *string                `json:"Role" mapstructure:"Role"`
	SubRole              *string                `json:"SubRole" mapstructure:"SubRole"`
	Nid                  *int                   `json:"Nid" mapstructure:"Nid"`
	Alias                []string               `json:"Alias" mapstructure:"Alias"`
	AdditionalProperties map[string]interface{} `json:"AdditionalProperties" mapstructure:"AdditionalProperties"`
}

type NodeMetadataStrings struct {
	Role                 string
	SubRole              string
	Nid                  string
	Alias                []string
	AdditionalProperties map[string]interface{}
}

type CabinetMetadata struct {
	HMNVlan *int `json:"HMNVlan" mapstructure:"HMNVlan"`
}

// DecodeProviderMetadata return a Metadata structure from the given hardwares CSM Provider properties.
// If the hardware doesn't have any metadata set an empty Metadata struct will be returned.
func DecodeProviderMetadata(cHardware inventory.Hardware) (result Metadata, err error) {
	ProviderMetadataRaw, ok := cHardware.ProviderMetadata[inventory.CSMProvider]
	if !ok {
		log.Debug().Any("id", cHardware.ID).Msgf("GetProviderMetadata: No CSM provider properties found")
		return Metadata{}, nil // This should be ok, as its possible as not all hardware inventory items may have CSM specific data
	}

	// Decode the Raw extra properties into the Metadata structure
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.StringToIPHookFunc(),
		Result:     &result,
	})
	if err != nil {
		return Metadata{}, err
	}
	err = decoder.Decode(ProviderMetadataRaw)

	return result, err
}

func EncodeProviderMetadata(metadata Metadata) (result map[string]interface{}, err error) {
	// Encode the Metadata struct into map[string]interface{}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.StringToIPHookFunc(),
		Result:     &result,
	})
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(metadata)
	if err != nil {
		return nil, err
	}

	return result, err
}

func (csm *CSM) BuildHardwareMetadata(cHardware *inventory.Hardware, rawProperties map[string]interface{}) error {
	if cHardware == nil {
		return fmt.Errorf("provided hardware is nil")
	}

	metadata := Metadata{}
	if cHardware.ProviderMetadata != nil {
		var err error
		metadata, err = DecodeProviderMetadata(*cHardware)

		if err != nil {
			return errors.Join(fmt.Errorf("failed to decode CSM metadata from hardware (%v)", cHardware.ID), err)
		}
	}

	switch cHardware.Type {
	case hardwaretypes.Cabinet:
		if metadata.Cabinet == nil {
			// Create an cabinet metadata object it does not exist
			metadata.Cabinet = &CabinetMetadata{}
		}

		// Make changes to the node metadata
		// The keys of rawProperties need to match what is defined in ./cmd/cabinet/add_cabinet.go
		if vlanIDRaw, exists := rawProperties[ProviderPropertyVlanId]; exists {
			// Check if the VLAN exceeds the valid range for the hardware
			max, err := DetermineEndingVlanFromSlug(cHardware.DeviceTypeSlug, *csm.hardwareLibrary)
			if err != nil {
				return err
			}
			// if the VLAN is greater than the max, fail
			if vlanIDRaw.(int) > max {
				return fmt.Errorf("VLAN exceeds the provider's maximum range (%d).  Please choose a valid VLAN", max)
			}
			if vlanIDRaw == nil {
				metadata.Cabinet.HMNVlan = nil
			} else {
				metadata.Cabinet.HMNVlan = pointers.IntPtr(vlanIDRaw.(int))
			}
		}
	case hardwaretypes.Node:
		if metadata.Node == nil {
			// Create an cabinet metadata object it does not exist
			metadata.Node = &NodeMetadata{}
		}

		// Make changes to the node metadata
		// The keys of rawProperties need to match what is defined in ./cmd/node/update_node.go
		if roleRaw, exists := rawProperties[ProviderPropertyRole]; exists {
			if roleRaw == nil {
				metadata.Node.Role = nil
			} else {
				metadata.Node.Role = pointers.StringPtr(roleRaw.(string))
			}
		}
		if subroleRaw, exists := rawProperties[ProviderPropertySubRole]; exists {
			if subroleRaw == nil {
				metadata.Node.SubRole = nil
			} else {
				metadata.Node.SubRole = pointers.StringPtr(subroleRaw.(string))
			}
		}
		if nidRaw, exists := rawProperties[ProviderPropertyNID]; exists {
			if nidRaw == nil {
				metadata.Node.Nid = nil
			} else {
				metadata.Node.Nid = pointers.IntPtr(nidRaw.(int))
			}
		}
		if aliasRaw, exists := rawProperties[ProviderPropertyAlias]; exists {
			if aliasRaw == nil {
				metadata.Node.Alias = nil
			} else {
				metadata.Node.Alias = []string{aliasRaw.(string)}
			}
		}

	default:
		// This hardware type doesn't have metadata for it right now
		return nil
	}

	// Set the hardware metadata
	metadataRaw, err := EncodeProviderMetadata(metadata)
	if err != nil {
		return errors.Join(fmt.Errorf("failed to encoder CSM Metadata for hardware (%v)", cHardware.ID), err)
	}

	cHardware.SetProviderMetadata(inventory.CSMProvider, metadataRaw)
	return nil
}

func (csm *CSM) RecommendCabinet(inv inventory.Inventory, deviceTypeSlug string) (recommended provider.HardwareRecommendations, err error) {
	// slice to track existing vlans
	var existingVlans = []int{}
	// slice to track existing cabinets
	var existingCabinets = []int{}

	// loop through the existing inventory to check for vlans
	log.Debug().Msg("Checking existing hardware to find recommendations")
	for _, cHardware := range inv.Hardware {
		switch cHardware.Type {
		case hardwaretypes.Cabinet:
			log.Debug().Msgf("Checking %s (%s)", cHardware.Type, cHardware.ID.String())
			log.Debug().Msgf("Decoding csm properties %+v", cHardware.ProviderMetadata)

			metadata, err := DecodeProviderMetadata(cHardware)
			if err != nil {
				return provider.HardwareRecommendations{}, errors.Join(fmt.Errorf("failed to decode CSM metadata from cabinet (%v)", cHardware.ID), err)
			}

			if metadata.Cabinet == nil {
				// There is no existing cabinet metadata
				continue
			}

			if metadata.Cabinet.HMNVlan != nil {
				// add it to the slice that tracks existing vlans
				existingVlans = append(existingVlans, *metadata.Cabinet.HMNVlan)
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
	var startingVlan, chosenVlan int
	startingVlan, err = DetermineStartingVlanFromSlug(deviceTypeSlug, *csm.hardwareLibrary)
	if err != nil {
		return recommended, err
	}
	if len(existingCabinets) == 0 {
		// choose a starting vlan based on the class
		chosenVlan = startingVlan
		log.Debug().Msgf("No cabinet VLANs found, using %d for %s %s", chosenVlan, class, hardwaretypes.Cabinet)
	} else {
		// set the recommended vlan by finding an available one from the existing
		chosenVlan = nextAvailableInt(existingVlans, startingVlan)

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

func (nm *NodeMetadata) Pretty() (prettyNm NodeMetadataStrings) {
	alias := []string{}
	if nm.Alias != nil {
		alias = nm.Alias
	}

	return NodeMetadataStrings{
		Role:    pointers.StrPtrToStr(nm.Role),
		SubRole: pointers.StrPtrToStr(nm.SubRole),
		Alias:   alias,
		Nid:     pointers.IntPtrToStr(nm.Nid),
	}
}
