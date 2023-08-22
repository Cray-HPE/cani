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

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/Cray-HPE/cani/pkg/pointers"
	"github.com/mitchellh/mapstructure"
)

const (
	ProviderMetadataVlanId  = "VlanID"
	ProviderMetadataRole    = "Role"
	ProviderMetadataSubRole = "SubRole"
	ProviderMetadataAlias   = "Alias"
	ProviderMetadataNID     = "NID"
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
	if ok {
		// Decode the Raw extra properties into the Metadata structure
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			DecodeHook: mapstructure.StringToIPHookFunc(),
			Result:     &result,
		})
		if err != nil {
			return Metadata{}, err
		}

		err = decoder.Decode(ProviderMetadataRaw)
		if err != nil {
			return result, err
		}
	}

	// Set initial values if not already present
	if result.Cabinet == nil && cHardware.Type == hardwaretypes.Cabinet {
		result.Cabinet = &CabinetMetadata{}
	} else if result.Node == nil && cHardware.Type == hardwaretypes.Node {
		result.Node = &NodeMetadata{}
	}

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
		if vlanIDRaw, exists := rawProperties[ProviderMetadataVlanId]; exists {
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
		if roleRaw, exists := rawProperties[ProviderMetadataRole]; exists {
			if roleRaw == nil {
				metadata.Node.Role = nil
			} else {
				metadata.Node.Role = pointers.StringPtr(roleRaw.(string))
			}
		}
		if subroleRaw, exists := rawProperties[ProviderMetadataSubRole]; exists {
			if subroleRaw == nil {
				metadata.Node.SubRole = nil
			} else {
				metadata.Node.SubRole = pointers.StringPtr(subroleRaw.(string))
			}
		}
		if nidRaw, exists := rawProperties[ProviderMetadataNID]; exists {
			if nidRaw == nil {
				metadata.Node.Nid = nil
			} else {
				metadata.Node.Nid = pointers.IntPtr(nidRaw.(int))
			}
		}
		if aliasRaw, exists := rawProperties[ProviderMetadataAlias]; exists {
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
