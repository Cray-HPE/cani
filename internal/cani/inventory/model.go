package inventory

import (
	hardware_type_library "github.com/Cray-HPE/cani/pkg/hardware-type-library"
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
	ID            uuid.UUID
	Name          string                             `json:"Name,omitempty" yaml:"Name,omitempty" default:"" usage:"Friendly name"`
	Type          hardware_type_library.HardwareType `json:"Type,omitempty" yaml:"Type,omitempty" default:"" usage:"Type"`
	Vendor        string                             `json:"Vendor,omitempty" yaml:"Vendor,omitempty" default:"" usage:"Vendor"`
	Architechture string                             `json:"Architechture,omitempty" yaml:"Architechture,omitempty" default:"" usage:"Architechture"`
	Model         string                             `json:"Model,omitempty" yaml:"Model,omitempty" default:"" usage:"Model"`
	Status        HardwareStatus                     `json:"Status,omitempty" yaml:"Status,omitempty" default:"Staged" usage:"Hardware can be [staged, provisioned, decomissioned]"`
	Properties    interface{}                        `json:"Properties,omitempty" yaml:"Properties,omitempty" default:"" usage:"Properties"`
	Parent        uuid.UUID                          `json:"Parent,omitempty" yaml:"Parent,omitempty" default:"00000000-0000-0000-0000-000000000000" usage:"Parent hardware"`
	// Children      []uuid.UUID `json:"Children,omitempty" yaml:"Children,omitempty" default:"" usage:"Child hardware"`

	LocationOrdinal *int
}

type LocationToken struct {
	HardwareType hardware_type_library.HardwareType
	Ordinal      int
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
