package devicetypes

import "github.com/google/uuid"

// CaniVLAN represents a layer-2 VLAN domain.
type CaniVLAN struct {
	// Identity
	ID          uuid.UUID `json:"id" yaml:"id"`
	VID         int       `json:"vid" yaml:"vid"` // VLAN ID (1-4094)
	Name        string    `json:"name" yaml:"name"`
	Description string    `json:"description,omitempty" yaml:"description,omitempty"`

	// Relationships
	Location uuid.UUID `json:"location,omitempty" yaml:"location,omitempty"` // Optional location scope

	// Shared metadata (status, role, tags, tenant, custom fields, external IDs, provider metadata)
	ObjectMeta `yaml:",inline"`
}

// GetID returns the unique identifier.
func (v *CaniVLAN) GetID() uuid.UUID {
	if v == nil {
		return uuid.Nil
	}
	return v.ID
}
