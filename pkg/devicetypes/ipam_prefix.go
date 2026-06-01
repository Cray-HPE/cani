package devicetypes

import "github.com/google/uuid"

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

// GetID returns the unique identifier.
func (p *CaniPrefix) GetID() uuid.UUID {
	if p == nil {
		return uuid.Nil
	}
	return p.ID
}
