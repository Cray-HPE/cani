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
	Alias                *string
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
				properties.Alias = StringPtr(aliasRaw.(string))
			}
		}

		cHardware.ProviderProperties["csm"] = properties

		return nil
	default:
		// This hardware type doesn't have metadata for it right now
		return nil
	}

}
