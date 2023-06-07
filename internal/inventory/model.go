package inventory

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Inventory is the top level object that represents the entire inventory
// This is what cani uses to represent the inventory
type Inventory struct {
	SchemaVersion SchemaVersion
	Provider      Provider
	Hardware      map[uuid.UUID]Hardware
}

// Hardware is the smallest unit of inventory
// It has all the potential fields that hardware can have
type Hardware struct {
	ID                 uuid.UUID
	Name               string                     `json:"Name,omitempty" yaml:"Name,omitempty" default:"" usage:"Friendly name"`
	Type               hardwaretypes.HardwareType `json:"Type,omitempty" yaml:"Type,omitempty" default:"" usage:"Type"`
	Vendor             string                     `json:"Vendor,omitempty" yaml:"Vendor,omitempty" default:"" usage:"Vendor"`
	Architecture       string                     `json:"Architecture,omitempty" yaml:"Architecture,omitempty" default:"" usage:"Architecture"`
	Model              string                     `json:"Model,omitempty" yaml:"Model,omitempty" default:"" usage:"Model"`
	Status             HardwareStatus             `json:"Status,omitempty" yaml:"Status,omitempty" default:"Staged" usage:"Hardware can be [staged, provisioned, decomissioned]"`
	Properties         interface{}                `json:"Properties,omitempty" yaml:"Properties,omitempty" default:"" usage:"Properties"`
	Role               string                     `json:"Role,omitempty" yaml:"Role,omitempty" default:"" usage:"Role"`
	SubRole            string                     `json:"SubRole,omitempty" yaml:"SubRole,omitempty" default:"" usage:"SubRole"`
	Alias              string                     `json:"Alias,omitempty" yaml:"Alias,omitempty" default:"" usage:"Alias"`
	ProviderProperties map[string]interface{}     `json:"ProviderProperties,omitempty" yaml:"ProviderProperties,omitempty" default:"" usage:"ProviderProperties"`

	Parent uuid.UUID `json:"Parent,omitempty" yaml:"Parent,omitempty" default:"00000000-0000-0000-0000-000000000000" usage:"Parent hardware"`
	// The following are derived from Parent
	Children     []uuid.UUID  `json:"Children,omitempty" yaml:"Children,omitempty"`
	LocationPath LocationPath `json:"LocationPath,omitempty" yaml:"LocationPath,omitempty"`

	LocationOrdinal *int
}

// HardwareStatus is the current state of the hardware
// Using a status allows for the hardware to be tracked through its lifecycle
// and allows for historical tracking of the hardware even if it is replaced or removed
type HardwareStatus string

// SchemaVersion is the version of the inventory schema
type SchemaVersion string

// Provider is the name of the external inventory provider
type Provider string

const (
	// Define constants for lifecyle states
	HardwareStatusStaged        = HardwareStatus("staged")
	HardwareStatusProvisioned   = HardwareStatus("provisioned")
	HardwareStatusDecomissioned = HardwareStatus("decomissioned")
	// Schema and proivider names are constant
	SchemaVersionV1Alpha1 = SchemaVersion("v1alpha1")
	CSMProvider           = Provider("csm")
)

type LocationToken struct {
	HardwareType hardwaretypes.HardwareType
	Ordinal      int
}

func (lt *LocationToken) String() string {
	return fmt.Sprintf("%s:%d", lt.HardwareType, lt.Ordinal)
}

type LocationPath []LocationToken

// String returns a string representation of the location path
func (lp LocationPath) String() string {
	tokens := []string{}

	for _, token := range lp {
		tokens = append(tokens, token.String())
	}

	return strings.Join(tokens, "->")
}

// GetHardwareTypePath returns the hardware type path of the location path
func (lp LocationPath) GetHardwareTypePath() hardwaretypes.HardwareTypePath {
	result := hardwaretypes.HardwareTypePath{}
	for _, token := range lp {
		result = append(result, token.HardwareType)
	}

	return result
}

// GetOrdinalPath returns the ordinal of the location path
func (lp LocationPath) GetOrdinalPath() []int {
	result := []int{}
	for _, token := range lp {
		result = append(result, token.Ordinal)
	}

	return result
}

// Exists returns true if the hardware exists in the datastore
func (lp LocationPath) Exists(ds Datastore) (bool, error) {
	log.Debug().Msgf("Checking if hardware exists at location %s", lp.String())
	hw, err := ds.GetAtLocation(lp)
	if err == nil {
		log.Warn().Msgf("%s %s exists at location %s", hw.Type, hw.ID.String(), lp.String())
		// Hardware found
		return true, nil
	} else if errors.Is(err, ErrHardwareNotFound) {
		// Hardware not found
		return false, nil
	} else {
		// Oops something happened
		return false, err
	}
}
