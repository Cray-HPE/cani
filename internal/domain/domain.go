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
)

// Domain is the logic that drives the application
// It contains the hardware types, the datastore, and the inventory provider
type Domain struct {
	hardwareTypeLibrary       *hardwaretypes.Library
	datastore                 inventory.Datastore
	externalInventoryProvider provider.InventoryProvider
	configOptions             provider.ConfigOptions
}

// NewOpts are the options for creating a new Domain
type NewOpts struct {
	DatastorePath string      `yaml:"datastore_path"`
	LogFilePath   string      `yaml:"log_file_path"`
	Provider      string      `yaml:"provider"`
	CsmOptions    csm.NewOpts `yaml:"csm_options"`
}

// New returns a new Domain using the provided options
func New(opts *NewOpts) (*Domain, error) {
	var err error
	domain := &Domain{}

	// Load the hardware type library
	// TODO make this be able to be loaded from a directory
	domain.hardwareTypeLibrary, err = hardwaretypes.NewEmbeddedLibrary()
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("failed to load embedded hardware type library"),
			err,
		)
	}

	// Load the datastore
	domain.datastore, err = inventory.NewDatastoreJSON(opts.DatastorePath, opts.LogFilePath, inventory.Provider(opts.Provider))
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("failed to load inventory datastore from file"),
			err,
		)
	}

	// Setup External inventory provider
	inventoryProvider, err := domain.datastore.InventoryProvider()
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("failed to retrieve external inventory provider type"),
			err,
		)
	}

	// Determine which external inventory provider to use
	switch inventoryProvider {
	case inventory.CSMProvider:
		domain.externalInventoryProvider, err = csm.New(&opts.CsmOptions, domain.hardwareTypeLibrary)
		if err != nil {
			return nil, errors.Join(
				fmt.Errorf("failed to initialize CSM external inventory provider"),
				err,
			)
		}
		domain.configOptions.ValidRoles = opts.CsmOptions.ValidRoles
		domain.configOptions.ValidSubRoles = opts.CsmOptions.ValidSubRoles
		domain.configOptions.K8sPodsCidr = opts.CsmOptions.K8sPodsCidr
		domain.configOptions.K8sServicesCidr = opts.CsmOptions.K8sServicesCidr
	default:
		return nil, fmt.Errorf("unknown external inventory provider provided (%s)", inventoryProvider)
	}
	return domain, nil
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
