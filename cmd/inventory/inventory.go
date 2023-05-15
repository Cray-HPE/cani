/*
MIT License

(C) Copyright 2022 Hewlett Packard Enterprise Development LP

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/
package inventory

import (
	"errors"
	"fmt"

	hardware_type_library "github.com/Cray-HPE/cani/pkg/hardware-type-library"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// The Inventory is a flat map of UUIDs to hardware
type Inventory map[uuid.UUID]Hardware

// Hardware is the smallest unit of inventory
// It has all the potential fields that hardware can have
type Hardware struct {
	Name          string                             `json:"Name,omitempty" yaml:"Name,omitempty" default:"" usage:"Friendly name"`
	Type          hardware_type_library.HardwareType `json:"Type,omitempty" yaml:"Type,omitempty" default:"" usage:"Type"`
	Vendor        string                             `json:"Vendor,omitempty" yaml:"Vendor,omitempty" default:"" usage:"Vendor"`
	Architechture string                             `json:"Architechture,omitempty" yaml:"Architechture,omitempty" default:"" usage:"Architechture"`
	Model         string                             `json:"Model,omitempty" yaml:"Model,omitempty" default:"" usage:"Model"`
	Status        string                             `json:"Status,omitempty" yaml:"Status,omitempty" default:"Staged" usage:"Hardware can be [staged, provisioned, decomissioned]"`
	Properties    interface{}                        `json:"Properties,omitempty" yaml:"Properties,omitempty" default:"" usage:"Properties"`
	Parent        uuid.UUID                          `json:"Parent,omitempty" yaml:"Parent,omitempty" default:"00000000-0000-0000-0000-000000000000" usage:"Parent hardware"`
	// Children      []uuid.UUID `json:"Children,omitempty" yaml:"Children,omitempty" default:"" usage:"Child hardware"`
}

type Properties struct {
}

// NewInventory creates a new Inventory instance
func NewInventory() *Inventory {
	inv := make(Inventory, 0)
	return &inv
}

// NewCabinet creates a new Cabinet instance
func NewCabinet(args []string) Hardware {
	hw := Hardware{
		Type: "Cabinet",
	}
	return hw
}

// NewChassis creates a new Chassis instance
func NewChassis(cmd *cobra.Command, args []string) Hardware {
	var hw Hardware
	var u uuid.UUID
	var err error
	uu := cmd.Flags().Lookup("cabinet").Value.String()
	if uu != "" {
		u, err = uuid.Parse(uu)
		if err != nil {
			return Hardware{}
		}
		hw = Hardware{
			Type:   "Chassis",
			Parent: u,
		}
	} else {
		hw = Hardware{
			Type: "Chassis",
		}
		return hw
	}
	return hw
}

// NewBlade creates a new Blade instance
func NewBlade(cmd *cobra.Command, args []string) (Hardware, error) {
	var hw Hardware
	var chassis string
	var parent uuid.UUID
	var err error

	// Create the library
	library, err := hardware_type_library.NewEmbeddedLibrary()
	if err != nil {
		panic(err)
	}

	// Get the list of supported blades
	blades := library.GetDeviceTypesByHardwareType(hardware_type_library.HardwareTypeNodeBlade)
	// For each arg, check if it matches a blade
	for _, arg := range args {
		for _, blade := range blades {
			// if the slug matches the arg, then we have a match and a blade to add
			if blade.Slug == arg {
				// Set all the hardware properties the inventory expects
				hw.Vendor = blade.Manufacturer
				hw.Model = blade.Model
				hw.Type = blade.HardwareType
				hw.Name = blade.Slug
			}
		}
	}

	// Check the chassis flag to see if we need to set the parent
	// TODO: Make this required
	chassis = cmd.Flags().Lookup("chassis").Value.String()
	if chassis != "" {
		// Check that it is a valid UUID
		parent, err = uuid.Parse(chassis)
		if err != nil {
			return Hardware{}, errors.New(fmt.Sprintf("Error parsing UUID: %v", err))
		}
		// Set that as the parent
		hw.Parent = parent
	}

	// return the constructed hardware
	return hw, nil
}

// NewPdu creates a new Pdu instance
func NewPdu(args []string) Hardware {
	hw := Hardware{
		Type: "PDU",
	}
	return hw
}

// NewSwitch creates a new Switch instance
func NewSwitch(args []string) Hardware {
	hw := Hardware{
		Type: "Switch",
	}
	return hw
}

// NewHsn creates a new Hsn instance
func NewHsn(args []string) Hardware {
	hw := Hardware{
		Type: "HSN",
	}
	return hw
}

// NewNode creates a new Node instance
func NewNode(args []string) Hardware {
	hw := Hardware{
		Type: "Node",
	}
	return hw
}

// Add adds hardware to the inventory
func Add(cmd *cobra.Command, args []string) (*Inventory, error) {
	// Get the database path from the root command
	dbpath := cmd.Root().PersistentFlags().Lookup("database").Value.String()
	// Instantiate the database
	db := GetInstance(dbpath)

	var hw Hardware
	var err error
	switch cmd.Name() {
	case "cabinet":
		hw = NewCabinet(args) // Create a new cabinet
	case "chassis":
		hw = NewChassis(cmd, args) // Create a new chassis
	case "blade":
		hw, err = NewBlade(cmd, args) // Create a new blade
	case "pdu":
		hw = NewPdu(args) // Create a new pdu
	case "switch":
		hw = NewSwitch(args) // Create a new switch
	case "hsn":
		hw = NewHsn(args) // Create a new hsn
	case "node":
		hw = NewNode(args) // Create a new node
	default:
		return db.Inventory, errors.New(fmt.Sprintf("Unknown hardware type: %s", cmd.Name()))
	}

	var u uuid.UUID
	uu := cmd.Flags().Lookup("uuid").Value.String()
	if uu != "" {
		u, err = uuid.Parse(uu)
		if err != nil {
			return db.Inventory, errors.New(fmt.Sprintf("Error parsing UUID: %v", err))
		}
	} else {
		u = uuid.New() // Create a new UUID
	}

	// Add the hardware to the database
	err = db.Set(u, hw)
	if err != nil {
		return db.Inventory, err
	}

	return db.Inventory, nil
}

// Remove removes hardware from the inventory
func Remove(cmd *cobra.Command, args []string) (*Inventory, error) {
	// Get the database path from the root command
	dbpath := cmd.Root().PersistentFlags().Lookup("database").Value.String()
	// Instantiate the database
	db := GetInstance(dbpath)

	keys := []uuid.UUID{}
	for _, u := range args {
		// Convert the string to a UUID
		u, err := uuid.Parse(u)
		if err != nil {
			return db.Inventory, errors.New(fmt.Sprintf("Error parsing UUID: %v", err))
		}
		keys = append(keys, u)
	}

	err := db.Delete(keys)
	if err != nil {
		return db.Inventory, err
	}

	return db.Inventory, nil
}

// List returns the inventory read from the database
func List(cmd *cobra.Command, args []string) (Inventory, error) {
	// Get the database path from the root command
	dbpath := cmd.Root().PersistentFlags().Lookup("database").Value.String()
	// Instantiate the database
	db := GetInstance(dbpath)

	keys := []uuid.UUID{}
	for _, u := range args {
		// Convert the string to a UUID
		u, err := uuid.Parse(u)
		if err != nil {
			return Inventory{}, errors.New(fmt.Sprintf("Error parsing UUID: %v", err))
		}
		keys = append(keys, u)
	}

	inv, err := db.Get(keys)
	if err != nil {
		return Inventory{}, err
	}

	return inv, nil
}
