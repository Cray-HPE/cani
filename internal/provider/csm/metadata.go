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

// func (nm *NodeMetadata) Validate() error {
// 	var validationErrs []error
// 	// Role checks
// 	if nm.Role == nil {
// 		validationErrs = append(validationErrs, fmt.Errorf("no role set"))
// 	}

// 	// Check to see if the role matches what is allowed by HSM
// 	// TODO

// 	// Check to see if the sub-role matches what is allowed by HSM

// 	// NID checks
// 	if nm.Nid == nil {
// 		validationErrs = append(validationErrs, fmt.Errorf("no nid set"))
// 	}
// 	if *nm.Nid < 0 {
// 		validationErrs = append(validationErrs, fmt.Errorf("a non-positive NID is set: %d", *nm.Nid))
// 	}

// 	// Alias checks
// 	if nm.Alias == nil {
// 		validationErrs = append(validationErrs, fmt.Errorf("no alias set"))
// 	}
// 	if strings.Contains(*nm.Alias, " ") {
// 		validationErrs = append(validationErrs, fmt.Errorf("alias contains spaces"))
// 	}

// 	if len(validationErrs) == 0 {
// 		return nil
// 	}

// 	return errors.Join(validationErrs...)
// }

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
