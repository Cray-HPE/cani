package devicetypes

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// CaniCableType represents a physical cable, both as hardware-library
// template and inventory instance.
type CaniCableType struct {
	// Identity
	ID           uuid.UUID `json:"id" yaml:"id,omitempty"`
	Slug         string    `json:"slug" yaml:"slug,omitempty"`
	Label        string    `json:"label,omitempty" yaml:"label,omitempty"`
	PartNumber   string    `json:"partNumber,omitempty" yaml:"part_number,omitempty"`
	Manufacturer string    `json:"manufacturer,omitempty" yaml:"manufacturer,omitempty"`
	Model        string    `json:"model,omitempty" yaml:"model,omitempty"`
	Description  string    `json:"description,omitempty" yaml:"description,omitempty"`
	HardwareType string    `json:"hardwareType,omitempty" yaml:"hardware-type,omitempty"`

	// Cable specifics
	CableCategory string   `json:"cableCategory,omitempty" yaml:"cable_category,omitempty"`
	ConnectorType string   `json:"connectorType,omitempty" yaml:"connector_type,omitempty"`
	CableType     string   `json:"cableType,omitempty" yaml:"cable_type,omitempty"` // Nautobot cable type enum (cat3, cat5e, dac-passive, aoc, smf, etc.)
	Length        *float64 `json:"length,omitempty" yaml:"length,omitempty"`
	LengthUnit    string   `json:"lengthUnit,omitempty" yaml:"length_unit,omitempty"`
	Weight        float64  `json:"weight,omitempty" yaml:"weight,omitempty"`
	WeightUnit    string   `json:"weightUnit,omitempty" yaml:"weight_unit,omitempty"`
	Color         string   `json:"color,omitempty" yaml:"color,omitempty"`

	// Shared metadata (status, role, tags, tenant, custom fields, external IDs, provider metadata)
	ObjectMeta `yaml:",inline"`

	// Inventory state
	TerminationA     uuid.UUID `json:"terminationA,omitempty" yaml:"termination_a,omitempty"`
	TerminationAType string    `json:"terminationAType,omitempty" yaml:"termination_a_type,omitempty"`
	TerminationB     uuid.UUID `json:"terminationB,omitempty" yaml:"termination_b,omitempty"`
	TerminationBType string    `json:"terminationBType,omitempty" yaml:"termination_b_type,omitempty"`

	// Device-level termination references
	TerminationADevice uuid.UUID `json:"terminationADevice,omitempty" yaml:"termination_a_device,omitempty"`
	TerminationBDevice uuid.UUID `json:"terminationBDevice,omitempty" yaml:"termination_b_device,omitempty"`
	TerminationAPort   string    `json:"terminationAPort,omitempty" yaml:"termination_a_port,omitempty"`
	TerminationBPort   string    `json:"terminationBPort,omitempty" yaml:"termination_b_port,omitempty"`

	// Source tracks where this type was loaded from (e.g. "builtin", "local:/path", "git:url").
	Source string `json:"-" yaml:"-"`
}

// NewCable creates a new CaniCableType with a generated UUID.
func NewCable(cableType, label string) *CaniCableType {
	return &CaniCableType{
		ID:         uuid.New(),
		Slug:       cableType,
		Label:      label,
		ObjectMeta: ObjectMeta{Status: string(StatusConnected)},
	}
}

// Validate checks the cable connection for internal consistency.
func (c *CaniCableType) Validate() error {
	if c == nil {
		return errors.New("cannot validate nil CaniCableType")
	}
	if c.Slug == "" {
		return errors.New("cable type slug must not be empty")
	}
	return nil
}

// GetID returns the unique identifier.
func (c *CaniCableType) GetID() uuid.UUID {
	if c == nil {
		return uuid.Nil
	}
	return c.ID
}

// GetSlug returns the cable type slug.
func (c *CaniCableType) GetSlug() string {
	if c == nil {
		return ""
	}
	return c.Slug
}

// GetVendor returns the manufacturer name.
func (c *CaniCableType) GetVendor() string {
	if c == nil {
		return ""
	}
	return c.Manufacturer
}

// GetType returns the cable hardware type.
func (c *CaniCableType) GetType() Type {
	return TypeCable
}

// GetStatus returns the current status.
func (c *CaniCableType) GetStatus() string {
	if c == nil {
		return ""
	}
	return c.Status
}

// SetTerminations sets both interface terminations.
func (c *CaniCableType) SetTerminations(interfaceA, interfaceB uuid.UUID) {
	if c == nil {
		return
	}
	c.TerminationA = interfaceA
	c.TerminationB = interfaceB
}

// SetDeviceTerminations sets both device UUIDs and port names for the cable endpoints.
func (c *CaniCableType) SetDeviceTerminations(deviceA, deviceB uuid.UUID, portA, portB string) {
	if c == nil {
		return
	}
	c.TerminationADevice = deviceA
	c.TerminationBDevice = deviceB
	c.TerminationAPort = portA
	c.TerminationBPort = portB
}

// ValidateCable checks if a cable connection is valid by verifying:
// 1. Both termination interfaces exist in the inventory
// 2. Interface types are compatible (same type for copper connections)
// 3. Interfaces are not already connected to different cables
func ValidateCable(cable *CaniCableType, inv *Inventory) error {
	if cable == nil {
		return errors.New("cable is nil")
	}
	if inv == nil {
		return errors.New("inventory is nil")
	}

	// Check termination A interface exists
	ifaceA, deviceA := inv.GetInterfaceByID(cable.TerminationA)
	if ifaceA == nil {
		return fmt.Errorf("termination A interface %s not found in inventory", cable.TerminationA)
	}

	// Check termination B interface exists
	ifaceB, deviceB := inv.GetInterfaceByID(cable.TerminationB)
	if ifaceB == nil {
		return fmt.Errorf("termination B interface %s not found in inventory", cable.TerminationB)
	}

	// Check interface type compatibility
	if !areInterfacesCompatible(ifaceA.Type, ifaceB.Type) {
		return fmt.Errorf("interface type mismatch: %s (%s) cannot connect to %s (%s)",
			ifaceA.Name, ifaceA.Type, ifaceB.Name, ifaceB.Type)
	}

	// Check if interfaces are already connected to different cables
	if ifaceA.ConnectedCable != nil && *ifaceA.ConnectedCable != cable.ID {
		return fmt.Errorf("interface %s on %s is already connected to another cable", ifaceA.Name, deviceA.Name)
	}
	if ifaceB.ConnectedCable != nil && *ifaceB.ConnectedCable != cable.ID {
		return fmt.Errorf("interface %s on %s is already connected to another cable", ifaceB.Name, deviceB.Name)
	}

	return nil
}

// areInterfacesCompatible checks if two interface types can be connected via cable.
// Generally, interfaces of the same speed/type are compatible.
func areInterfacesCompatible(typeA, typeB InterfacesElemType) bool {
	// Same type is always compatible
	if typeA == typeB {
		return true
	}

	// Define compatible interface groups
	compatibleGroups := [][]InterfacesElemType{
		// 1GbE copper interfaces
		{InterfacesElemTypeA1000BaseT, InterfacesElemTypeA1000BaseKx},
		// 10GbE SFP+ interfaces
		{InterfacesElemTypeA10GbaseXSfpp, InterfacesElemTypeA10GbaseT},
	}

	for _, group := range compatibleGroups {
		aInGroup := false
		bInGroup := false
		for _, t := range group {
			if typeA == t {
				aInGroup = true
			}
			if typeB == t {
				bInGroup = true
			}
		}
		if aInGroup && bInGroup {
			return true
		}
	}

	return false
}
