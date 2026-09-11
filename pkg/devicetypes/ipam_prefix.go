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

// PrefixType classifies a prefix's function within the IP hierarchy.
type PrefixType string

const (
	PrefixTypeContainer PrefixType = "container"
	PrefixTypeNetwork   PrefixType = "network"
	PrefixTypePool      PrefixType = "pool"
)

// CaniPrefix represents an IPv4 or IPv6 network prefix in CIDR notation.
// Prefixes form a hierarchy: a more-specific prefix is a child of a
// less-specific one that contains it.
type CaniPrefix struct {
	// Identity
	ID          uuid.UUID  `json:"id" yaml:"id"`
	Prefix      string     `json:"prefix" yaml:"prefix"`                           // CIDR notation, e.g. "10.0.0.0/24"
	Network     string     `json:"network,omitempty" yaml:"network,omitempty"`     // Network address (derived)
	Broadcast   string     `json:"broadcast,omitempty" yaml:"broadcast,omitempty"` // Broadcast address (derived)
	PrefixLen   int        `json:"prefixLength" yaml:"prefix_length"`              // Mask bits
	IPVersion   int        `json:"ipVersion" yaml:"ip_version"`                    // 4 or 6
	Type        PrefixType `json:"type,omitempty" yaml:"type,omitempty"`           // container, network, pool
	Description string     `json:"description,omitempty" yaml:"description,omitempty"`

	// Relationships
	Location uuid.UUID `json:"location,omitempty" yaml:"location,omitempty"` // Optional location scope
	VLAN     uuid.UUID `json:"vlan,omitempty" yaml:"vlan,omitempty"`         // Optional VLAN association
	VRF      string    `json:"vrf,omitempty" yaml:"vrf,omitempty"`           // Optional VRF name (string, not FK)
	Parent   uuid.UUID `json:"parent,omitempty" yaml:"parent,omitempty"`     // Parent prefix (auto-computed)

	// Shared metadata (status, role, tags, tenant, custom fields, external IDs, provider metadata)
	ObjectMeta `yaml:",inline"`
}

// Validate checks that the prefix is valid CIDR without mutating the receiver.
func (p *CaniPrefix) Validate() error {
	if p == nil {
		return fmt.Errorf("cannot validate nil CaniPrefix")
	}
	copy := *p
	return ParsePrefix(&copy)
}

// GetID returns the unique identifier.
func (p *CaniPrefix) GetID() uuid.UUID {
	if p == nil {
		return uuid.Nil
	}
	return p.ID
}

// GetSlug returns the CIDR as the prefix's portable natural key.
func (p *CaniPrefix) GetSlug() string {
	if p == nil {
		return ""
	}
	return p.Prefix
}

// GetStatus returns the current status.
func (p *CaniPrefix) GetStatus() string {
	if p == nil {
		return ""
	}
	return p.Status
}
