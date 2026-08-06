package devicetypes

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ValidateVID verifies vid is within the valid 802.1Q range (1-4094).
func ValidateVID(vid int) error {
	if vid < 1 || vid > 4094 {
		return fmt.Errorf("invalid VLAN ID %d: must be between 1 and 4094", vid)
	}
	return nil
}

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

// Validate checks the VLAN for internal consistency.
func (v *CaniVLAN) Validate() error {
	if v == nil {
		return errors.New("cannot validate nil CaniVLAN")
	}
	if err := ValidateVID(v.VID); err != nil {
		return fmt.Errorf("VLAN %s: %w", v.ID, err)
	}
	return nil
}
