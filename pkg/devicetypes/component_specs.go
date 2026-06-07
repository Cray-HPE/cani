package devicetypes

import "github.com/google/uuid"

// InterfaceSpec defines an interface template in a device/module type.
// When used in inventory, ID and ConnectedCable are populated.
type InterfaceSpec struct {
	ID             uuid.UUID          `yaml:"id,omitempty" json:"id,omitempty"`
	Name           string             `yaml:"name" json:"name"`
	Type           InterfacesElemType `yaml:"type" json:"type"`
	Label          string             `yaml:"label,omitempty" json:"label,omitempty"`
	Role           string             `yaml:"role,omitempty" json:"role,omitempty"`
	MacAddress     string             `yaml:"mac_address,omitempty" json:"macAddress,omitempty"`
	MgmtOnly       *bool              `yaml:"mgmt_only,omitempty" json:"mgmt_only,omitempty"`
	ConnectedCable *uuid.UUID         `yaml:"connected_cable,omitempty" json:"connectedCable,omitempty"`
}

// CaniInterface represents an instantiated interface on a specific device.
type CaniInterface struct {
	ID            uuid.UUID          `json:"id" yaml:"id"`
	Name          string             `json:"name" yaml:"name"`
	InterfaceType InterfacesElemType `json:"interfaceType" yaml:"interface_type"`
	DeviceID      uuid.UUID          `json:"deviceId" yaml:"device_id"`

	// Shared metadata (status, role, tags, tenant, custom fields, external IDs, provider metadata)
	ObjectMeta `yaml:",inline"`

	MgmtOnly       bool        `json:"mgmtOnly,omitempty" yaml:"mgmt_only,omitempty"`
	Label          string      `json:"label,omitempty" yaml:"label,omitempty"`
	MacAddress     string      `json:"macAddress,omitempty" yaml:"mac_address,omitempty"`
	ConnectedCable *uuid.UUID  `json:"connectedCable,omitempty" yaml:"connected_cable,omitempty"`
	ContentType    string      `json:"contentType,omitempty" yaml:"content_type,omitempty"` // For cable terminations (e.g., "dcim.interface")
	IPAddresses    []uuid.UUID `json:"ipAddresses,omitempty" yaml:"ip_addresses,omitempty"` // Assigned IP address IDs
}

// ConsolePortSpec defines a console port in a device type.
type ConsolePortSpec struct {
	Name string `yaml:"name" json:"name"`
	Type string `yaml:"type" json:"type"`
}

// PowerPortSpec defines a power port in a device type.
type PowerPortSpec struct {
	Name          string `yaml:"name" json:"name"`
	Type          string `yaml:"type" json:"type"`
	MaximumDraw   int    `yaml:"maximum_draw,omitempty" json:"maximum_draw,omitempty"`
	AllocatedDraw int    `yaml:"allocated_draw,omitempty" json:"allocated_draw,omitempty"`
}

// ModuleBaySpec defines a module bay/slot in a device type.
type ModuleBaySpec struct {
	Name     string `yaml:"name" json:"name"`
	Position string `yaml:"position,omitempty" json:"position,omitempty"`
}

// DeviceBaySlugRef is a slug or list of slugs used in allowed/default device-bay fields.
type DeviceBaySlugRef struct {
	Slug  interface{} `yaml:"slug" json:"slug"`
	Types interface{} `yaml:"types,omitempty" json:"types,omitempty"`
}

// Slugs returns the referenced slugs as a string slice.
func (r DeviceBaySlugRef) Slugs() []string {
	switch v := r.Slug.(type) {
	case string:
		return []string{v}
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

// AllowedTypes returns the allowed hardware-type strings as a slice.
func (r DeviceBaySlugRef) AllowedTypes() []string {
	switch v := r.Types.(type) {
	case string:
		return []string{v}
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

// DeviceBaySpec defines a device bay (U-slot) in a rack or device type.
// Provider-specific YAML keys (e.g. "ordinal") are captured in Extra.
type DeviceBaySpec struct {
	Name     string            `yaml:"name" json:"name"`
	Position string            `yaml:"position,omitempty" json:"position,omitempty"`
	Allowed  *DeviceBaySlugRef `yaml:"allowed,omitempty" json:"allowed,omitempty"`
	Default  *DeviceBaySlugRef `yaml:"default,omitempty" json:"default,omitempty"`
	Extra    map[string]any    `yaml:",inline" json:"extra,omitempty"`
}

// Identification provides alternate manufacturer/model combinations for lookup.
type Identification struct {
	Manufacturer string `yaml:"manufacturer" json:"manufacturer"`
	Model        string `yaml:"model" json:"model"`
}

// InterfacesElemType represents the type of network interface.
type InterfacesElemType string

const (
	InterfacesElemTypeA1000BaseT       InterfacesElemType = "1000base-t"
	InterfacesElemTypeA1000BaseKx      InterfacesElemType = "1000base-kx"
	InterfacesElemTypeA10GbaseT        InterfacesElemType = "10gbase-t"
	InterfacesElemTypeA10GbaseXSfpp    InterfacesElemType = "10gbase-x-sfpp"
	InterfacesElemTypeA25GbaseXSfp28   InterfacesElemType = "25gbase-x-sfp28"
	InterfacesElemTypeA40GbaseXQsfpp   InterfacesElemType = "40gbase-x-qsfpp"
	InterfacesElemTypeA100GbaseXQsfp28 InterfacesElemType = "100gbase-x-qsfp28"
	InterfacesElemTypeA200GbaseXQsfp56 InterfacesElemType = "200gbase-x-qsfp56"
	InterfacesElemTypeA400GbaseXQsfpdd InterfacesElemType = "400gbase-x-qsfpdd"
	InterfacesElemTypeA400GbaseXOsfp   InterfacesElemType = "400gbase-x-osfp"
	InterfacesElemTypeVirtual          InterfacesElemType = "virtual"
	InterfacesElemTypeLag              InterfacesElemType = "lag"
)
