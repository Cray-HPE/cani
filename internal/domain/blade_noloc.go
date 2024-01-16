/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// addBladeNoGeoLoc adds a blade without worrying about geolocation or ordinals
func (d *Domain) addBladeNoGeoLoc(cmd *cobra.Command, args []string) (result AddHardwareResult, err error) {
	deviceTypeSlug := args[0]
	// Verify the provided device type slug is a node blade
	deviceType, err := d.hardwareTypeLibrary.GetDeviceType(deviceTypeSlug)
	if err != nil {
		return AddHardwareResult{}, err
	}
	if deviceType.HardwareType != hardwaretypes.NodeBlade {
		return AddHardwareResult{}, fmt.Errorf("provided device hardware type (%s) is not a %s", deviceTypeSlug, hardwaretypes.NodeBlade)
	}

	// Generate a hardware build out
	hardwareBuildOutItems, err := inventory.GenerateHardwareBuildOut(d.hardwareTypeLibrary, inventory.GenerateHardwareBuildOutOpts{
		DeviceTypeSlug: deviceTypeSlug,
	})
	if err != nil {
		return result, err
	}

	for _, hardwareBuildOut := range hardwareBuildOutItems {
		var hardware inventory.Hardware

		if hardwareBuildOut.ExistingHardware == nil {
			// New hardware not present in the inventory
			hardware = inventory.NewHardwareFromBuildOut(hardwareBuildOut, inventory.HardwareStatusStaged)

			if err := d.datastore.Add(&hardware); err != nil {
				return AddHardwareResult{}, errors.Join(
					fmt.Errorf("unable to add hardware to inventory datastore"),
					err,
				)
			}
		} else {
			// Empty hardware is present in the inventory
			hardware = *hardwareBuildOut.ExistingHardware
			// Set hardware type information from build out
			hardware.DeviceTypeSlug = hardwareBuildOut.DeviceTypeSlug
			hardware.Type = hardwareBuildOut.DeviceType.HardwareType
			hardware.Vendor = hardwareBuildOut.DeviceType.Manufacturer
			hardware.Model = hardwareBuildOut.DeviceType.Model
			// The hardware is now staged, and not empty
			hardware.Status = inventory.HardwareStatusStaged

			if err := d.datastore.Update(&hardware); err != nil {
				return AddHardwareResult{}, errors.Join(
					fmt.Errorf("unable to add hardware to inventory datastore"),
					err,
				)
			}
		}
		pair := HardwareLocationPair{
			Hardware: hardware,
		}
		log.Debug().Msgf("Also adding %s--a %s--with parent %s", pair.Hardware.ID, pair.Hardware.Type, pair.Hardware.Parent)
		result.AddedHardware = append(result.AddedHardware, pair)
	}

	// Validate the CANI's datastore
	if failedValidations, err := d.datastore.Validate(); len(failedValidations) > 0 {
		result.DatastoreValidationErrors = failedValidations
	} else if err != nil {
		return AddHardwareResult{}, errors.Join(
			fmt.Errorf("failed to validate datastore inventory"),
			err,
		)
	}

	// Validate the current state of CANI's inventory data against the provider plugin
	// for provider specific data.
	if failedValidations, err := d.externalInventoryProvider.ValidateInternal(cmd, args, d.datastore, false); len(failedValidations) > 0 {
		result.ProviderValidationErrors = failedValidations
	} else if err != nil {
		return AddHardwareResult{}, errors.Join(
			fmt.Errorf("failed to validate inventory against inventory provider plugin"),
			err,
		)
	}

	if len(result.DatastoreValidationErrors) > 0 || len(result.ProviderValidationErrors) > 0 {
		return result, provider.ErrDataValidationFailure
	}

	return result, nil
}
