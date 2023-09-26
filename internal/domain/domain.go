/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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
package domain

import (
	"errors"
	"fmt"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/internal/provider/csm"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// Domain is the logic that drives the application
// It contains the hardware types, the datastore, and the inventory provider
type Domain struct {
	hardwareTypeLibrary       *hardwaretypes.Library
	datastore                 inventory.Datastore
	externalInventoryProvider provider.InventoryProvider
	Active                    bool         `yaml:"active"`
	DatastorePath             string       `yaml:"datastore_path"`
	LogFilePath               string       `yaml:"log_file_path"`
	CustomHardwareTypesDir    string       `yaml:"custom_hardware_types_dir"`
	Provider                  string       `yaml:"provider"`
	CsmOptions                *csm.CsmOpts `yaml:"csm_options"`
}

// New returns a new Domain
func New(cmd *cobra.Command, args []string) (d *Domain, err error) {
	d = &Domain{}
	if cmd.Name() == "init" {
		// the only arg is the provider if the command is 'init'
		d.Provider = args[0]
	}
	return d, nil
}

// SetupDomain sets the provider options for the domain
func (d *Domain) SetupDomain(cmd *cobra.Command, args []string) (err error) {
	// Load the hardware type library
	d.hardwareTypeLibrary, err = hardwaretypes.NewEmbeddedLibrary(d.CustomHardwareTypesDir)
	if err != nil {
		return errors.Join(
			fmt.Errorf("failed to load embedded hardware type library"),
			err,
		)
	}

	// Load the datastore
	d.datastore, err = inventory.NewDatastoreJSON(d.DatastorePath, d.LogFilePath, inventory.Provider(d.Provider))
	if err != nil {
		return errors.Join(
			fmt.Errorf("failed to load inventory datastore from file"),
			err,
		)
	}

	// Setup External inventory provider
	inventoryProvider, err := d.datastore.InventoryProvider()
	if err != nil {
		return errors.Join(
			fmt.Errorf("failed to retrieve external inventory provider type"),
			err,
		)
	}

	// Determine which external inventory provider to use
	switch inventoryProvider {
	case inventory.CSMProvider:
		// Create a new provider object
		d.externalInventoryProvider, err = csm.New(cmd, args, d.hardwareTypeLibrary)
		if err != nil {
			return errors.Join(
				fmt.Errorf("failed to initialize CSM external inventory provider"),
				err,
			)
		}

		if cmd.Name() == "init" {
			// Set the additional options
			err := d.externalInventoryProvider.SetProviderOptions(cmd, args)
			if err != nil {
				return err
			}

			po, err := d.externalInventoryProvider.GetProviderOptions()
			if err != nil {
				return err
			}

			csmOpts := po.(*csm.CsmOpts)
			d.CsmOptions = csmOpts
		} else {
			err = d.externalInventoryProvider.SetProviderOptionsInterface(d.CsmOptions)
			if err != nil {
				return err
			}
		}

		if cmd.Name() == "apply" {
			// Set the additional options
			err := d.externalInventoryProvider.SetProviderOptions(cmd, args)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unknown external inventory provider provided (%s)", inventoryProvider)

	}
	return nil
}

type HardwareLocationPair struct {
	Hardware inventory.Hardware
	Location inventory.LocationPath
}

type AddHardwareResult struct {
	AddedHardware             []HardwareLocationPair
	DatastoreValidationErrors map[uuid.UUID]inventory.ValidateResult
	ProviderValidationErrors  map[uuid.UUID]provider.HardwareValidationResult
}

type UpdatedHardwareResult struct {
	// UpdatedHardware          []HardwareLocationPair
	DatastoreValidationErrors map[uuid.UUID]inventory.ValidateResult // TODO
	ProviderValidationErrors  map[uuid.UUID]provider.HardwareValidationResult
}
