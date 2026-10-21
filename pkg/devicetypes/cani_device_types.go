package devicetypes

import (
	"errors"
	"fmt"
	"reflect"
	"slices"

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
	Platform  string      `json:"platform,omitempty" yaml:"platform,omitempty"` // OS/firmware platform
	Parent    uuid.UUID   `json:"parent" yaml:"parent,omitempty"`
	ParentBay string      `json:"parentBay,omitempty" yaml:"parent_bay,omitempty"` // device bay occupied in the parent device
	Children  []uuid.UUID `json:"-" yaml:"-"`                                      // reverse index; rebuilt from child Device.Parent at load time
	Frus      []uuid.UUID `json:"frus,omitempty" yaml:"frus,omitempty"`

	// Explicit FK fields for Nautobot export. Rack and ParentDevice are
	// derived from Parent at load time and never persisted. Location is
	// derived the same way when the device has a Parent; a device placed
	// directly in a location (no rack) keeps the persisted value.
	Rack         uuid.UUID `json:"-" yaml:"-"`                                   // FK to CaniRackType
	Location     uuid.UUID `json:"location,omitempty" yaml:"location,omitempty"` // FK to CaniLocationType
	ParentDevice uuid.UUID `json:"-" yaml:"-"`                                   // FK to parent device (blade→chassis)

	// Rack placement
	RackPosition int    `json:"rackPosition,omitempty" yaml:"rack_position,omitempty"`
	Face         string `json:"face,omitempty" yaml:"face,omitempty"`

	// IPAM: primary IP addresses for this device
	PrimaryIPv4 uuid.UUID `json:"primaryIpv4,omitempty" yaml:"primary_ipv4,omitempty"`
	PrimaryIPv6 uuid.UUID `json:"primaryIpv6,omitempty" yaml:"primary_ipv6,omitempty"`

	// Relationships (Nautobot relationship associations)
	AssignedVLANs []uuid.UUID `json:"assignedVlans,omitempty" yaml:"assigned_vlans,omitempty"` // VLANs configured on this switch
	BMCParent     uuid.UUID   `json:"bmcParent,omitempty" yaml:"bmc_parent,omitempty"`         // Parent device this BMC belongs to

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

// MergeProperties merges set properties from another CaniDeviceType into this
// one and reports whether anything changed. An empty incoming value never
// clears an existing one.
func (c *CaniDeviceType) MergeProperties(other *CaniDeviceType) bool {
	if c == nil || other == nil {
		return false
	}
	// Every merge must run, so the results are collected rather than
	// short-circuited.
	changes := []bool{
		c.mergeStrings(other),
		mergeSet(&c.UHeight, other.UHeight),
		c.mergePlacement(other),
		mergeSet(&c.Parent, other.Parent),
		mergeSet(&c.ParentBay, other.ParentBay),
		mergeSet(&c.Location, other.Location),
		c.mergeMetadata(other.ProviderMetadata),
	}
	return slices.Contains(changes, true)
}

// mergeStrings carries every set string field of other; Type joins the table
// through its underlying string.
func (c *CaniDeviceType) mergeStrings(other *CaniDeviceType) bool {
	changed := false
	for _, f := range []struct {
		dst *string
		src string
	}{
		{&c.Name, other.Name}, {&c.Slug, other.Slug}, {&c.Manufacturer, other.Manufacturer},
		{&c.Model, other.Model}, {&c.Status, other.Status}, {(*string)(&c.Type), string(other.Type)},
		{&c.Role, other.Role}, {&c.Serial, other.Serial}, {&c.Description, other.Description},
		{&c.Comments, other.Comments}, {&c.AssetTag, other.AssetTag},
	} {
		changed = mergeSet(f.dst, f.src) || changed
	}
	return changed
}

// mergePlacement carries the rack position and face. Nautobot requires rack,
// position and face to be set together, so a position without a face
// defaults the face to "front".
func (c *CaniDeviceType) mergePlacement(other *CaniDeviceType) bool {
	changed := mergeSet(&c.RackPosition, other.RackPosition)
	changed = mergeSet(&c.Face, other.Face) || changed
	if c.RackPosition > 0 && c.Face == "" {
		c.Face = "front"
		changed = true
	}
	return changed
}

// mergeMetadata copies provider sub-maps key by key and top-level scalar keys
// from incoming into the device's ProviderMetadata.
func (c *CaniDeviceType) mergeMetadata(incoming map[string]any) bool {
	if len(incoming) == 0 {
		return false
	}
	if c.ProviderMetadata == nil {
		c.ProviderMetadata = make(map[string]any)
	}
	changed := false
	for k, v := range incoming {
		changed = c.mergeMetadataKey(k, v) || changed
	}
	return changed
}

func (c *CaniDeviceType) mergeMetadataKey(k string, v any) bool {
	incomingSub, isMap := v.(map[string]any)
	if !isMap {
		if reflect.DeepEqual(c.ProviderMetadata[k], v) {
			return false
		}
		c.ProviderMetadata[k] = v
		return true
	}
	existing, _ := c.ProviderMetadata[k].(map[string]any)
	if existing == nil {
		existing = make(map[string]any)
		c.ProviderMetadata[k] = existing
	}
	changed := false
	for sk, sv := range incomingSub {
		if !reflect.DeepEqual(existing[sk], sv) {
			existing[sk] = sv
			changed = true
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

// GetIsFullDepth returns whether this device occupies full rack depth.
func (c *CaniDeviceType) GetIsFullDepth() bool {
	if c == nil {
		return false
	}
	return c.IsFullDepth
}
