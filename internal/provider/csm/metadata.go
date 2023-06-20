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

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
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
		if vlanIDRaw, exists := rawProperties["vlanID"]; exists {
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
