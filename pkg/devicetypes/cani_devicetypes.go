package devicetypes

import (
	"fmt"

	"github.com/google/uuid"
)

type CaniDeviceType struct {
	ID               uuid.UUID      `json:"ID" yaml:"ID" default:"" usage:"Unique Identifier"`
	Name             string         `json:"Name,omitempty" yaml:"Name,omitempty" default:"" usage:"Friendly name"`
	Type             Type           `json:"Type,omitempty" yaml:"Type,omitempty" default:"" usage:"Type"`
	DeviceTypeSlug   string         `json:"DeviceTypeSlug,omitempty" yaml:"DeviceTypeSlug,omitempty" default:"" usage:"Hardware Type Library Device slug"`
	Vendor           string         `json:"Vendor,omitempty" yaml:"Vendor,omitempty" default:"" usage:"Vendor"`
	Architecture     string         `json:"Architecture,omitempty" yaml:"Architecture,omitempty" default:"" usage:"Architecture"`
	Model            string         `json:"Model,omitempty" yaml:"Model,omitempty" default:"" usage:"Model"`
	Status           string         `json:"Status,omitempty" yaml:"Status,omitempty" default:"Staged" usage:"Hardware can be [staged, provisioned, decomissioned]"`
	Properties       map[string]any `json:"Properties,omitempty" yaml:"Properties,omitempty" default:"" usage:"Properties"`
	ProviderMetadata map[string]any `json:"ProviderMetadata,omitempty" yaml:"ProviderMetadata,omitempty" default:"" usage:"ProviderMetadata"`
	Parent           uuid.UUID      `json:"Parent,omitempty" yaml:"Parent,omitempty" default:"00000000-0000-0000-0000-000000000000" usage:"Parent hardware"`
	Children         []uuid.UUID    `json:"Children,omitempty" yaml:"Children,omitempty"` // derived from Parent
	// LocationPath     LocationPath           `json:"LocationPath,omitempty" yaml:"LocationPath,omitempty"` // derived from Parent
	// LocationOrdinal *int `json:"LocationOrdinal,omitempty" yaml:"LocationOrdinal,omitempty" default:"" usage:"LocationOrdinal"`
}

// Merge merges the fields of another CaniDeviceType into this one.
// If the ID is empty, it generates a new UUID.
func (c *CaniDeviceType) Merge(*CaniDeviceType) error {
	if c == nil {
		return fmt.Errorf("cannot merge into nil CaniDeviceType")
	}
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}

	// Merge fields
	if c.Name == "" {
		c.Name = "Unnamed Device"
	}
	if c.DeviceTypeSlug == "" {
		c.DeviceTypeSlug = "unknown"
	}
	if c.Vendor == "" {
		c.Vendor = "Unknown Vendor"
	}
	if c.Architecture == "" {
		c.Architecture = "Unknown Architecture"
	}
	if c.Model == "" {
		c.Model = "Unknown Model"
	}

	return nil
}
