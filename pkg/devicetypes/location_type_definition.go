package devicetypes

import "errors"

// LocationTypeDefinition is the YAML-loadable template for a Nautobot LocationType.
// Each YAML file in location-types/ maps 1:1 to a Nautobot LocationType.
type LocationTypeDefinition struct {
	Name         string   `json:"name" yaml:"name"`
	Slug         string   `json:"slug" yaml:"slug"`
	Description  string   `json:"description,omitempty" yaml:"description,omitempty"`
	Nestable     bool     `json:"nestable,omitempty" yaml:"nestable,omitempty"`
	ContentTypes []string `json:"content_types,omitempty" yaml:"content_types,omitempty"`
	Parent       string   `json:"parent,omitempty" yaml:"parent,omitempty"` // parent type slug
	Source       string   `json:"source,omitempty" yaml:"-"`
}

// Validate checks the definition for required fields.
func (d *LocationTypeDefinition) Validate() error {
	if d.Name == "" {
		return errors.New("location type name must not be empty")
	}
	if d.Slug == "" {
		return errors.New("location type slug must not be empty")
	}
	return nil
}
