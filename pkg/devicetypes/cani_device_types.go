package devicetypes

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/google/uuid"
)

// CaniDeviceType is the single device representation used for both
// hardware-library templates (loaded from YAML) and inventory instances.
type CaniDeviceType struct {
	// Identity
	ID           uuid.UUID `json:"id" yaml:"id,omitempty"`
	Name         string    `json:"name" yaml:"name,omitempty"`
	Slug         string    `json:"slug" yaml:"slug,omitempty"`
	PartNumber   string    `json:"partNumber" yaml:"part_number,omitempty"`
	Manufacturer string    `json:"manufacturer" yaml:"manufacturer,omitempty"`
	Vendor       string    `json:"vendor,omitempty" yaml:"vendor,omitempty"`
	Model        string    `json:"model" yaml:"model,omitempty"`
	Description  string    `json:"description,omitempty" yaml:"description,omitempty"`
	Serial       string    `json:"serial,omitempty" yaml:"serial,omitempty"`
	AssetTag     string    `json:"assetTag,omitempty" yaml:"asset_tag,omitempty"`

	// Classification
	Type          Type   `json:"type,omitempty" yaml:"type,omitempty"`
	SubdeviceRole string `json:"subdeviceRole,omitempty" yaml:"subdevice_role,omitempty"` // parent or child (chassis/blade)

	// Physical
	UHeight     int     `json:"uHeight,omitempty" yaml:"u_height,omitempty"`
	IsFullDepth bool    `json:"isFullDepth,omitempty" yaml:"is_full_depth,omitempty"`
	Weight      float64 `json:"weight,omitempty" yaml:"weight,omitempty"`
	WeightUnit  string  `json:"weightUnit,omitempty" yaml:"weight_unit,omitempty"`
	Comments    string  `json:"comments,omitempty" yaml:"comments,omitempty"`

	// Hardware specs (from library YAML)
	Interfaces      []InterfaceSpec   `json:"interfaces,omitempty" yaml:"interfaces,omitempty"`
	ConsolePorts    []ConsolePortSpec `json:"consolePorts,omitempty" yaml:"console-ports,omitempty"`
	PowerPorts      []PowerPortSpec   `json:"powerPorts,omitempty" yaml:"power-ports,omitempty"`
	ModuleBays      []ModuleBaySpec   `json:"moduleBays,omitempty" yaml:"module-bays,omitempty"`
	DeviceBays      []DeviceBaySpec   `json:"deviceBays,omitempty" yaml:"device-bays,omitempty"`
	Identifications []Identification  `json:"identifications,omitempty" yaml:"identifications,omitempty"`

	// Shared metadata (status, role, tags, tenant, custom fields, external IDs, provider metadata)
	ObjectMeta `yaml:",inline"`

	// Inventory state
	Platform string      `json:"platform,omitempty" yaml:"platform,omitempty"` // OS/firmware platform
	Parent   uuid.UUID   `json:"parent" yaml:"parent,omitempty"`
	Children []uuid.UUID `json:"children,omitempty" yaml:"children,omitempty"`
	Frus     []uuid.UUID `json:"frus,omitempty" yaml:"frus,omitempty"`

	// Explicit FK fields for Nautobot export (derived from Parent at load time)
	Rack         uuid.UUID `json:"rack,omitempty" yaml:"rack,omitempty"`                  // FK to CaniRackType
	Location     uuid.UUID `json:"location,omitempty" yaml:"location,omitempty"`          // FK to CaniLocationType
	ParentDevice uuid.UUID `json:"parentDevice,omitempty" yaml:"parent_device,omitempty"` // FK to parent device (blade→chassis)

	// Rack placement
	RackPosition int    `json:"rackPosition,omitempty" yaml:"rack_position,omitempty"`
	Face         string `json:"face,omitempty" yaml:"face,omitempty"`

	// IPAM: primary IP addresses for this device
	PrimaryIPv4 uuid.UUID `json:"primaryIpv4,omitempty" yaml:"primary_ipv4,omitempty"`
	PrimaryIPv6 uuid.UUID `json:"primaryIpv6,omitempty" yaml:"primary_ipv6,omitempty"`

	// Source tracks where this type was loaded from (e.g. "builtin", "local:/path", "git:url").
	Source string `json:"-" yaml:"-"`
}

// IsCable returns true if this device is a cable type.
func (c *CaniDeviceType) IsCable() bool {
	if c == nil {
		return false
	}
	return c.Type == TypeCable
}

// GetVendor returns the vendor name, falling back to Manufacturer.
func (c *CaniDeviceType) GetVendor() string {
	if c == nil {
		return ""
	}
	if c.Vendor != "" {
		return c.Vendor
	}
	return c.Manufacturer
}

// GetType returns the hardware type as a Type constant.
func (c *CaniDeviceType) GetType() Type {
	if c == nil {
		return ""
	}
	return c.Type
}

// MergeProperties merges non-empty properties from another CaniDeviceType into this one.
// Returns true if any changes were made.
func (c *CaniDeviceType) MergeProperties(other *CaniDeviceType) bool {
	if c == nil || other == nil {
		return false
	}
	changed := false

	if other.Name != "" && c.Name != other.Name {
		c.Name = other.Name
		changed = true
	}
	if other.Slug != "" && c.Slug != other.Slug {
		c.Slug = other.Slug
		changed = true
	}
	if other.Manufacturer != "" && c.Manufacturer != other.Manufacturer {
		c.Manufacturer = other.Manufacturer
		changed = true
	}
	if other.Model != "" && c.Model != other.Model {
		c.Model = other.Model
		changed = true
	}
	if other.Status != "" && c.Status != other.Status {
		c.Status = other.Status
		changed = true
	}
	if other.Type != "" && c.Type != other.Type {
		c.Type = other.Type
		changed = true
	}
	if other.UHeight != 0 && c.UHeight != other.UHeight {
		c.UHeight = other.UHeight
		changed = true
	}
	if other.RackPosition != 0 && c.RackPosition != other.RackPosition {
		c.RackPosition = other.RackPosition
		changed = true
	}
	if other.Face != "" && c.Face != other.Face {
		c.Face = other.Face
		changed = true
	}
	// Nautobot requires rack, position, and face to be set together;
	// default face to "front" when a position is present.
	if c.RackPosition > 0 && c.Face == "" {
		c.Face = "front"
		changed = true
	}
	if other.Parent != uuid.Nil && c.Parent != other.Parent {
		c.Parent = other.Parent
		changed = true
	}
	if other.Role != "" && c.Role != other.Role {
		c.Role = other.Role
		changed = true
	}
	if other.Serial != "" && c.Serial != other.Serial {
		c.Serial = other.Serial
		changed = true
	}
	if other.Description != "" && c.Description != other.Description {
		c.Description = other.Description
		changed = true
	}

	// Merge ProviderMetadata: copy/overwrite provider sub-maps and
	// top-level keys from other into c.
	if len(other.ProviderMetadata) > 0 {
		if c.ProviderMetadata == nil {
			c.ProviderMetadata = make(map[string]any)
		}
		for k, v := range other.ProviderMetadata {
			otherSub, isMap := v.(map[string]any)
			if !isMap {
				// Top-level scalar key.
				if !reflect.DeepEqual(c.ProviderMetadata[k], v) {
					c.ProviderMetadata[k] = v
					changed = true
				}
				continue
			}
			// Merge provider sub-map key by key.
			existing, _ := c.ProviderMetadata[k].(map[string]any)
			if existing == nil {
				existing = make(map[string]any)
				c.ProviderMetadata[k] = existing
			}
			for sk, sv := range otherSub {
				if !reflect.DeepEqual(existing[sk], sv) {
					existing[sk] = sv
					changed = true
				}
			}
		}
	}

	return changed
}

// GetID returns the unique identifier.
func (c *CaniDeviceType) GetID() uuid.UUID {
	if c == nil {
		return uuid.Nil
	}
	return c.ID
}

// GetSlug returns the device type slug.
func (c *CaniDeviceType) GetSlug() string {
	if c == nil {
		return ""
	}
	return c.Slug
}

// GetStatus returns the current status.
func (c *CaniDeviceType) GetStatus() string {
	if c == nil {
		return ""
	}
	return c.Status
}

// Validate checks the device type for consistency and returns an error if invalid.
func (c *CaniDeviceType) Validate() error {
	if c == nil {
		return errors.New("cannot validate nil CaniDeviceType")
	}
	if c.Slug != "" {
		if _, ok := GetBySlug(c.Slug); !ok {
			return fmt.Errorf("device type slug %q not found in library", c.Slug)
		}
	}

	return nil
}

// InstantiateInterfaces creates CaniInterface entries from this device's InterfaceSpec definitions.
// Returns the instantiated interfaces for assignment to an inventory record.
func (c *CaniDeviceType) InstantiateInterfaces() []CaniInterface {
	if c == nil {
		return nil
	}
	instances := make([]CaniInterface, 0, len(c.Interfaces))
	for _, iface := range c.Interfaces {
		mgmtOnly := false
		if iface.MgmtOnly != nil {
			mgmtOnly = *iface.MgmtOnly
		}
		role := ResolveInterfaceRole(iface.Role, iface.Name, iface.Type, mgmtOnly)
		instances = append(instances, CaniInterface{
			ID:            uuid.New(),
			Name:          iface.Name,
			InterfaceType: iface.Type,
			DeviceID:      c.ID,
			ObjectMeta:    ObjectMeta{Status: string(StatusActive), Role: role},
			MgmtOnly:      mgmtOnly,
			MacAddress:    iface.MacAddress,
		})
	}
	return instances
}

// GetInterface returns the interface spec with the given name, or nil if not found.
func (c *CaniDeviceType) GetInterface(name string) *InterfaceSpec {
	if c == nil {
		return nil
	}
	for i := range c.Interfaces {
		if c.Interfaces[i].Name == name {
			return &c.Interfaces[i]
		}
	}
	return nil
}

// GetRackID returns the rack ID for this device. Prefers the explicit Rack FK;
// falls back to checking if Parent exists in the inventory's Racks collection.
// Returns uuid.Nil if not in a rack.
func (c *CaniDeviceType) GetRackID(inv *Inventory) uuid.UUID {
	if c == nil || inv == nil {
		return uuid.Nil
	}
	if c.Rack != uuid.Nil {
		return c.Rack
	}
	if _, ok := inv.Racks[c.Parent]; ok {
		return c.Parent
	}
	return uuid.Nil
}

// GetUHeight returns the U-height for this device.
// Returns 1 as default if not set.
func (c *CaniDeviceType) GetUHeight() int {
	if c == nil || c.UHeight < 1 {
		return 1
	}
	return c.UHeight
}

// GetIsFullDepth returns whether this device occupies full rack depth.
func (c *CaniDeviceType) GetIsFullDepth() bool {
	if c == nil {
		return false
	}
	return c.IsFullDepth
}
