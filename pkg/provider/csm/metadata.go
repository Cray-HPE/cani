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

	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/Cray-HPE/cani/pkg/pointers"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
)

const (
	ProviderMetadataVlanId  = "HMNVlan"
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

func (csm *CSM) NewHardwareMetadata(hw *inventory.Hardware, cmd *cobra.Command, args []string) (err error) {
	md, err := DecodeProviderMetadata(*hw)
	if err != nil {
		return err
	}
	// Get the flags and set the metdata accordingly
	role, _ := cmd.Flags().GetString("role")
	if cmd.Flags().Changed("role") {
		md.Node.Role = &role
	}
	subrole, _ := cmd.Flags().GetString("subrole")
	if cmd.Flags().Changed("subrole") {
		md.Node.SubRole = &subrole
	}
	nid, _ := cmd.Flags().GetInt("nid")
	if cmd.Flags().Changed("nid") {
		md.Node.Nid = &nid
	}
	alias, _ := cmd.Flags().GetStringSlice("alias")
	if cmd.Flags().Changed("alias") {
		md.Node.Alias = alias
	}

	metadata, err := EncodeProviderMetadata(md)
	if err != nil {
		return err
	}

	hw.SetProviderMetadata(inventory.CSMProvider, metadata)

	return nil
}

func (csm *CSM) BuildHardwareMetadata(cHardware *inventory.Hardware, cmd *cobra.Command, args []string, recommendations provider.HardwareRecommendations) error {
	if cHardware == nil {
		return fmt.Errorf("provided hardware is nil")
	}

	var metadata = map[string]interface{}{
		// string(hardwaretypes.Cabinet): map[string]interface{}{},
		// string(hardwaretypes.Node):    map[string]interface{}{},
	}
	switch cHardware.Type {
	case hardwaretypes.Cabinet:
		var vlan int
		// if the flag is set, get the vlan there
		if cmd.Flags().Changed("vlan-id") {
			vlan, _ = cmd.Flags().GetInt("vlan-id")
			// otherwise, get it from the recommendations
		} else {
			val, exists := recommendations.ProviderMetadata[ProviderMetadataVlanId]
			if exists {
				vlan = val.(int)
			}
		}

		// check for the vlan limit
		max, err := DetermineEndingVlanFromSlug(cHardware.DeviceTypeSlug, *csm.hardwareLibrary)
		if err != nil {
			return err
		}

		// if the VLAN is greater than the max, fail
		if vlan > max {
			return fmt.Errorf("VLAN exceeds the provider's maximum range (%d).  Please choose a valid VLAN", max)
		}

		metadata[string(hardwaretypes.Cabinet)] = recommendations.ProviderMetadata

		if cHardware.ProviderMetadata == nil {
			cHardware.ProviderMetadata = map[inventory.Provider]inventory.ProviderMetadataRaw{}
		}
		if cHardware.ProviderMetadata[taxonomy.CSM] == nil {
			cHardware.ProviderMetadata[taxonomy.CSM] = map[string]interface{}{}
		}
		// set the metadata
		cHardware.ProviderMetadata[taxonomy.CSM] = metadata

	case hardwaretypes.Node:
		role, _ := cmd.Flags().GetString("role")
		subrole, _ := cmd.Flags().GetString("subrole")
		nid, _ := cmd.Flags().GetInt("nid")
		alias, _ := cmd.Flags().GetStringSlice("alias")

		md := map[string]interface{}{}
		if role != "" {
			md["role"] = role
		}
		if subrole != "" {
			md["subrole"] = subrole
		}
		if cmd.Flags().Changed("nid") {
			md["nid"] = nid
		}
		if alias != nil {
			md["alias"] = alias
		}
		metadata[string(hardwaretypes.Node)] = md

	default:
		// This hardware type doesn't have metadata for it right now
		return nil
	}

	cHardware.SetProviderMetadata(inventory.CSMProvider, metadata)
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
