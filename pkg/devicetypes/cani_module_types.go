package devicetypes

import (
	"errors"

	"github.com/google/uuid"
)

// CaniModuleType represents a module installed in a device (GPU, NIC, PSU, memory, etc.).
// Serves as both hardware-library template and inventory instance.
type CaniModuleType struct {
	// Identity
	ID           uuid.UUID `json:"id" yaml:"id,omitempty"`
	Name         string    `json:"name" yaml:"name,omitempty"`
	Slug         string    `json:"slug" yaml:"slug,omitempty"`
	PartNumber   string    `json:"partNumber,omitempty" yaml:"part_number,omitempty"`
	Manufacturer string    `json:"manufacturer,omitempty" yaml:"manufacturer,omitempty"`
	Model        string    `json:"model,omitempty" yaml:"model,omitempty"`
	Description  string    `json:"description,omitempty" yaml:"description,omitempty"`
	HardwareType string    `json:"hardwareType,omitempty" yaml:"hardware-type,omitempty"`

	// Physical
	Weight     float64 `json:"weight,omitempty" yaml:"weight,omitempty"`
	WeightUnit string  `json:"weightUnit,omitempty" yaml:"weight_unit,omitempty"`
	Comments   string  `json:"comments,omitempty" yaml:"comments,omitempty"`

	// Hardware specs (from library YAML)
	Interfaces []InterfaceSpec `json:"interfaces,omitempty" yaml:"interfaces,omitempty"`

	// Inventory state
	ParentDevice  uuid.UUID   `json:"parentDevice,omitempty" yaml:"parent_device,omitempty"`
	ModuleBayName string      `json:"moduleBayName,omitempty" yaml:"module_bay_name,omitempty"`
	Serial        string      `json:"serial,omitempty" yaml:"serial,omitempty"`
	AssetTag      string      `json:"assetTag,omitempty" yaml:"asset_tag,omitempty"`
	Status        string      `json:"status" yaml:"status,omitempty"`
	Role          string      `json:"role,omitempty" yaml:"role,omitempty"`
	Location      uuid.UUID   `json:"location,omitempty" yaml:"location,omitempty"` // optional; inherits from parent device if unset
	Frus          []uuid.UUID `json:"frus,omitempty" yaml:"frus,omitempty"`

	// Multi-tenancy and metadata
	Tenant       string         `json:"tenant,omitempty" yaml:"tenant,omitempty"`
	Tags         []string       `json:"tags,omitempty" yaml:"tags,omitempty"`
	CustomFields map[string]any `json:"customFields,omitempty" yaml:"custom_fields,omitempty"`

	// Source tracks where this type was loaded from (e.g. "builtin", "local:/path", "git:url").
	Source string `json:"-" yaml:"-"`
}

// Validate checks the module for internal consistency.
func (m *CaniModuleType) Validate() error {
	if m == nil {
		return errors.New("cannot validate nil CaniModuleType")
	}
	return nil
}

// GetID returns the unique identifier.
func (m *CaniModuleType) GetID() uuid.UUID {
	if m == nil {
		return uuid.Nil
	}
	return m.ID
}

// GetSlug returns the module type slug.
func (m *CaniModuleType) GetSlug() string {
	if m == nil {
		return ""
	}
	return m.Slug
}

// GetStatus returns the current status.
func (m *CaniModuleType) GetStatus() string {
	if m == nil {
		return ""
	}
	return m.Status
}

// GetVendor returns the manufacturer name.
func (m *CaniModuleType) GetVendor() string {
	if m == nil {
		return ""
	}
	return m.Manufacturer
}

// GetType returns the hardware type as a Type constant.
func (m *CaniModuleType) GetType() Type {
	if m == nil {
		return ""
	}
	if m.HardwareType != "" {
		return Type(m.HardwareType)
	}
	return TypeModule
}

// InstantiateInterfaces creates InterfaceInstance entries from this module's specs.
func (m *CaniModuleType) InstantiateInterfaces() []InterfaceInstance {
	if m == nil {
		return nil
	}
	instances := make([]InterfaceInstance, 0, len(m.Interfaces))
	for _, iface := range m.Interfaces {
		mgmtOnly := false
		if iface.MgmtOnly != nil {
			mgmtOnly = *iface.MgmtOnly
		}
		instances = append(instances, InterfaceInstance{
			ID:            uuid.New(),
			Name:          iface.Name,
			InterfaceType: iface.Type,
			DeviceID:      m.ID,
			Status:        "active",
			MgmtOnly:      mgmtOnly,
		})
	}
	return instances
}

// GetInterface returns the interface spec with the given name, or nil if not found.
func (m *CaniModuleType) GetInterface(name string) *InterfaceSpec {
	if m == nil {
		return nil
	}
	for i := range m.Interfaces {
		if m.Interfaces[i].Name == name {
			return &m.Interfaces[i]
		}
	}
	return nil
}

// GetInterfaceByName is an alias for GetInterface.
func (m *CaniModuleType) GetInterfaceByName(name string) *InterfaceSpec {
	return m.GetInterface(name)
}
