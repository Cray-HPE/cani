/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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
package devicetypes

import (
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

// Validate checks the VLAN identifier and required name.
func (v *CaniVLAN) Validate() error {
	if v == nil {
		return fmt.Errorf("cannot validate nil CaniVLAN")
	}
	if err := ValidateVID(v.VID); err != nil {
		return err
	}
	if v.Name == "" {
		return fmt.Errorf("VLAN name must not be empty")
	}
	return nil
}

// GetID returns the unique identifier.
func (v *CaniVLAN) GetID() uuid.UUID {
	if v == nil {
		return uuid.Nil
	}
	return v.ID
}

// GetSlug returns the VLAN name as its portable natural key.
func (v *CaniVLAN) GetSlug() string {
	if v == nil {
		return ""
	}
	return v.Name
}

// GetStatus returns the current status.
func (v *CaniVLAN) GetStatus() string {
	if v == nil {
		return ""
	}
	return v.Status
}
