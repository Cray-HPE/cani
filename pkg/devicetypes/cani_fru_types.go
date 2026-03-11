package devicetypes

import (
	"errors"

	"github.com/google/uuid"
)

// CaniFruType represents a Field Replaceable Unit (spare part or replacement component).
// Serves as both hardware-library template and inventory instance.
type CaniFruType struct {
	// Identity
	ID           uuid.UUID `json:"id" yaml:"id,omitempty"`
	Name         string    `json:"name" yaml:"name,omitempty"`
	Slug         string    `json:"slug" yaml:"slug,omitempty"`
	PartNumber   string    `json:"partNumber" yaml:"part_number,omitempty"`
	Manufacturer string    `json:"manufacturer,omitempty" yaml:"manufacturer,omitempty"`
	Model        string    `json:"model,omitempty" yaml:"model,omitempty"`
	Description  string    `json:"description,omitempty" yaml:"description,omitempty"`
	HardwareType string    `json:"hardwareType,omitempty" yaml:"hardware_type,omitempty"`

	// Physical
	Weight     float64 `json:"weight,omitempty" yaml:"weight,omitempty"`
	WeightUnit string  `json:"weightUnit,omitempty" yaml:"weight_unit,omitempty"`

	// Inventory state
	Label      string    `json:"label,omitempty" yaml:"label,omitempty"`
	Serial     string    `json:"serial,omitempty" yaml:"serial,omitempty"`
	AssetTag   string    `json:"assetTag,omitempty" yaml:"asset_tag,omitempty"`
	Role       string    `json:"role,omitempty" yaml:"role,omitempty"`
	Status     string    `json:"status" yaml:"status,omitempty"`
	Device     uuid.UUID `json:"device,omitempty" yaml:"device,omitempty"`
	Parent     uuid.UUID `json:"parent,omitempty" yaml:"parent,omitempty"`
	Discovered bool      `json:"discovered,omitempty" yaml:"discovered,omitempty"`

	// Multi-tenancy and metadata
	Tags         []string       `json:"tags,omitempty" yaml:"tags,omitempty"`
	CustomFields map[string]any `json:"customFields,omitempty" yaml:"custom_fields,omitempty"`

	// Source tracks where this type was loaded from (e.g. "builtin", "local:/path", "git:url").
	Source string `json:"-" yaml:"-"`
}

// Validate checks the FRU for internal consistency.
func (f *CaniFruType) Validate() error {
	if f == nil {
		return errors.New("cannot validate nil CaniFruType")
	}
	return nil
}

// GetID returns the unique identifier.
func (f *CaniFruType) GetID() uuid.UUID {
	if f == nil {
		return uuid.Nil
	}
	return f.ID
}

// GetSlug returns the FRU type slug.
func (f *CaniFruType) GetSlug() string {
	if f == nil {
		return ""
	}
	return f.Slug
}

// GetStatus returns the current status.
func (f *CaniFruType) GetStatus() string {
	if f == nil {
		return ""
	}
	return f.Status
}

// GetVendor returns the manufacturer name.
func (f *CaniFruType) GetVendor() string {
	if f == nil {
		return ""
	}
	return f.Manufacturer
}

// GetType returns the hardware type as a Type constant.
func (f *CaniFruType) GetType() Type {
	if f == nil {
		return ""
	}
	if f.HardwareType != "" {
		return Type(f.HardwareType)
	}
	return TypeFru
}
