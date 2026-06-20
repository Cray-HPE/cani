package devicetypes

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// CaniLocationType represents a physical location in the inventory hierarchy.
// Locations can contain child locations (building → floor → room) and racks.
type CaniLocationType struct {
	// Identity
	ID           uuid.UUID   `json:"id" yaml:"id"`
	Name         string      `json:"name" yaml:"name"`
	Slug         string      `json:"slug,omitempty" yaml:"slug,omitempty"`
	LocationType string      `json:"locationType" yaml:"location_type"` // site, building, floor, room, etc.
	Parent       uuid.UUID   `json:"parent,omitempty" yaml:"parent,omitempty"`
	Children     []uuid.UUID `json:"-" yaml:"-"` // reverse index; rebuilt from child Location.Parent at load time
	Racks        []uuid.UUID `json:"-" yaml:"-"` // reverse index; rebuilt from CaniRackType.Location at load time

	// Shared metadata (status, role, tags, tenant, custom fields, external IDs, provider metadata)
	ObjectMeta `yaml:",inline"`

	Nestable     bool     `json:"nestable,omitempty" yaml:"nestable,omitempty"`
	ContentTypes []string `json:"contentTypes,omitempty" yaml:"content_types,omitempty"`
	Source       string   `json:"source,omitempty" yaml:"-"`

	// Nautobot-equivalent fields
	Facility        string `json:"facility,omitempty" yaml:"facility,omitempty"`
	Description     string `json:"description,omitempty" yaml:"description,omitempty"`
	PhysicalAddress string `json:"physicalAddress,omitempty" yaml:"physical_address,omitempty"`
	ShippingAddress string `json:"shippingAddress,omitempty" yaml:"shipping_address,omitempty"`
	Latitude        string `json:"latitude,omitempty" yaml:"latitude,omitempty"`
	Longitude       string `json:"longitude,omitempty" yaml:"longitude,omitempty"`
	ContactName     string `json:"contactName,omitempty" yaml:"contact_name,omitempty"`
	ContactPhone    string `json:"contactPhone,omitempty" yaml:"contact_phone,omitempty"`
	ContactEmail    string `json:"contactEmail,omitempty" yaml:"contact_email,omitempty"`
	TimeZone        string `json:"timeZone,omitempty" yaml:"time_zone,omitempty"`
	Asn             *int64 `json:"asn,omitempty" yaml:"asn,omitempty"`
	Comments        string `json:"comments,omitempty" yaml:"comments,omitempty"`
}

// Validate checks the location for internal consistency.
func (l *CaniLocationType) Validate() error {
	if l == nil {
		return errors.New("cannot validate nil CaniLocationType")
	}
	if l.LocationType == "" {
		return errors.New("location type must not be empty")
	}
	return nil
}

// GetID returns the unique identifier.
func (l *CaniLocationType) GetID() uuid.UUID {
	if l == nil {
		return uuid.Nil
	}
	return l.ID
}

// GetSlug returns the location type string (e.g., "site", "building", "room").
func (l *CaniLocationType) GetSlug() string {
	if l == nil {
		return ""
	}
	return l.LocationType
}

// GetStatus returns the current status.
func (l *CaniLocationType) GetStatus() string {
	if l == nil {
		return ""
	}
	return l.Status
}

// ValidateContentType checks whether the given content type is allowed.
// An empty ContentTypes list permits everything (backwards compatibility).
func (l *CaniLocationType) ValidateContentType(ct string) error {
	if l == nil || len(l.ContentTypes) == 0 {
		return nil
	}
	for _, allowed := range l.ContentTypes {
		if allowed == ct {
			return nil
		}
	}
	return fmt.Errorf("content type %q is not permitted at location %s (%s); allowed: %v",
		ct, l.Name, l.LocationType, l.ContentTypes)
}

// AddRack adds a rack UUID to this location's rack list.
func (l *CaniLocationType) AddRack(rackID uuid.UUID) {
	if l == nil {
		return
	}
	for _, id := range l.Racks {
		if id == rackID {
			return // already present
		}
	}
	l.Racks = append(l.Racks, rackID)
}

// AddChild adds a child location UUID to this location.
func (l *CaniLocationType) AddChild(childID uuid.UUID) {
	if l == nil {
		return
	}
	for _, id := range l.Children {
		if id == childID {
			return // already present
		}
	}
	l.Children = append(l.Children, childID)
}
