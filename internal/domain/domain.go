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
	"os"

	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/internal/provider/csm"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// Domain is the logic that drives the application
// It contains the hardware types, the datastore, and the inventory provider
type Domain struct {
	hardwareTypeLibrary       *hardwaretypes.Library
	datastore                 inventory.Datastore
	externalInventoryProvider provider.InventoryProvider
	Active                    bool        `yaml:"active"`
	DatastorePath             string      `yaml:"datastore_path"`
	LogFilePath               string      `yaml:"log_file_path"`
	CustomHardwareTypesDir    string      `yaml:"custom_hardware_types_dir"`
	Provider                  string      `yaml:"provider"`
	Options                   interface{} `yaml:"options"`
}

// SessionInitCmd is used by the cmd package, but the provider sets all the command options/flags/etc.
var SessionInitCmd *cobra.Command

// New returns a new Domain
func New(cmd *cobra.Command, args []string) (d *Domain, err error) {
	d = &Domain{}
	d.Provider = args[0]
	return d, nil
}

// SetupDomain sets the provider options for the domain
func (d *Domain) SetupDomain(cmd *cobra.Command, args []string, configDomains map[string]*Domain) (err error) {
	// The domain needs a minimum of three things to start:
	//   1. a datastore to save the inventory to
	//   2. a hardware type library to know compatible hardware
	//   3. a provider interface object

	for _, sessionDomain := range configDomains {
		if sessionDomain.Active {
			d.Active = true
			log.Debug().Msgf("Setting top level Domain to Active=true")
		}
	}

	// active sessions should have a datastore
	if _, err := os.Stat(d.DatastorePath); os.IsNotExist(err) {
		if d.Active {
			return fmt.Errorf("Datastore '%s' does not exist.  Run 'session init' to begin", d.DatastorePath)
		}
	}

	log.Debug().Msgf("Setting up datastore interface: %s", d.DatastorePath)
	// Load the datastore.  Different providers have different storage needs
	switch d.Provider {
	case taxonomy.CSM:
		d.datastore, err = inventory.NewDatastoreJSONCSM(d.DatastorePath, d.LogFilePath, inventory.Provider(d.Provider))
	default:
		log.Warn().Msgf("using default provider datastore")
		d.datastore, err = inventory.NewDatastoreJSON(d.DatastorePath, d.LogFilePath, inventory.Provider(d.Provider))
	}
	if err != nil {
		return errors.Join(
			fmt.Errorf("failed to load %s inventory datastore from source", taxonomy.App),
			err,
		)
	}

	log.Debug().Msgf("loading embedded and custom hardwaretypes: %s", d.CustomHardwareTypesDir)
	// Load the hardware type library.  Supported hardware is embedded
	// this also loads any custom-deinfed hardware at the given path
	d.hardwareTypeLibrary, err = hardwaretypes.NewEmbeddedLibrary(d.CustomHardwareTypesDir)
	if err != nil {
		return errors.Join(
			fmt.Errorf("failed to load embedded hardware type library"),
			err,
		)
	}

	log.Debug().Msgf("getting plugin settings for provider %s", d.Provider)
	// Instantiate the provider interface object
	switch d.Provider {
	case taxonomy.CSM:
		d.externalInventoryProvider, err = csm.New(cmd, args, d.hardwareTypeLibrary, d.Options)

	default:
		return fmt.Errorf("unknown external inventory provider provided (%s)", d.Provider)
	}

	if err != nil {
		return errors.Join(
			fmt.Errorf("failed to initialize %s external inventory provider", d.Provider),
			err,
		)
	}

	if cmd.Name() == "init" {
		err := d.externalInventoryProvider.SetProviderOptions(cmd, args)
		if err != nil {
			return err
		}
	}

	if d.Options == nil {
		opts, err := d.externalInventoryProvider.GetProviderOptions()
		if err != nil {
			return err
		}

		d.Options = opts
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

func GetProviders() []provider.InventoryProvider {
	supportedProviders := []provider.InventoryProvider{
		&csm.CSM{},
	}
	return supportedProviders
}
