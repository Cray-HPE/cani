package inventory

import (
	"fmt"
	"strings"

	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
)

type HardwareStatus string

const (
	HardwareStatusStaged        HardwareStatus = HardwareStatus("staged")
	HardwareStatusProvisioned   HardwareStatus = HardwareStatus("provisioned")
	HardwareStatusDecomissioned HardwareStatus = HardwareStatus("decomissioned")
)

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
	Parent             uuid.UUID                  `json:"Parent,omitempty" yaml:"Parent,omitempty" default:"00000000-0000-0000-0000-000000000000" usage:"Parent hardware"`
	Role               string                     `json:"Role,omitempty" yaml:"Role,omitempty" default:"" usage:"Role"`
	SubRole            string                     `json:"SubRole,omitempty" yaml:"SubRole,omitempty" default:"" usage:"SubRole"`
	Alias              string                     `json:"Alias,omitempty" yaml:"Alias,omitempty" default:"" usage:"Alias"`
	ProviderProperties map[string]interface{}     `json:"ProviderProperties,omitempty" yaml:"ProviderProperties,omitempty" default:"" usage:"ProviderProperties"`

	LocationOrdinal *int
}

type LocationToken struct {
	HardwareType hardwaretypes.HardwareType
	Ordinal      int
}

func (lt *LocationToken) String() string {
	return fmt.Sprintf("%s:%d", lt.HardwareType, lt.Ordinal)
}

type LocationPath []LocationToken

func (lp LocationPath) String() string {
	tokens := []string{}

	for _, token := range lp {
		tokens = append(tokens, token.String())
	}

	return strings.Join(tokens, "->")
}

func (lp LocationPath) GetHardwareTypePath() hardwaretypes.HardwareTypePath {
	result := hardwaretypes.HardwareTypePath{}
	for _, token := range lp {
		result = append(result, token.HardwareType)
	}

	return result
}

func (lp LocationPath) GetOrdinalPath() []int {
	result := []int{}
	for _, token := range lp {
		result = append(result, token.Ordinal)
	}

	return result
}

type SchemaVersion string

const (
	SchemaVersionV1Alpha1 = SchemaVersion("v1alpha1")
)

type ExternalInventoryProvider string

const (
	ExternalInventoryProviderCSM = ExternalInventoryProvider("csm")
)

type Inventory struct {
	SchemaVersion             SchemaVersion
	ExternalInventoryProvider ExternalInventoryProvider
	Hardware                  map[uuid.UUID]Hardware
}
