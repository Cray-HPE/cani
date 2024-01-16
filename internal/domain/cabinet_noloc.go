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
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func (d *Domain) addCabinetNoGeoLoc(cmd *cobra.Command, args []string, recommendations provider.HardwareRecommendations) (result AddHardwareResult, err error) {
	// Verify the provided device type slug is a node blade
	deviceTypeSlug := args[0]
	system, err := d.datastore.GetSystemZero()
	if err != nil {
		return AddHardwareResult{}, errors.Join(
			fmt.Errorf("unable to get system zero"),
			err,
		)
	}

	// Generate a hardware build out using the system as a parent
	hardwareBuildOutItems, err := inventory.GenerateDefaultHardwareBuildOutNoLoc(d.hardwareTypeLibrary, deviceTypeSlug, system)
	if err != nil {
		return AddHardwareResult{}, errors.Join(
			fmt.Errorf("unable to build default hardware build out for %s", deviceTypeSlug),
			err,
		)
	}

	for _, hardwareBuildOut := range hardwareBuildOutItems {
		// Generate the CANI hardware inventory version of the hardware build out data
		hardware := inventory.NewHardwareFromBuildOut(hardwareBuildOut, inventory.HardwareStatusStaged)

		// Ask the inventory provider to craft a metadata object for this information
		if err := d.externalInventoryProvider.BuildHardwareMetadata(&hardware, cmd, args, recommendations); err != nil {
			return AddHardwareResult{}, err
		}

		// Metadata is now set by the BuildHardwareMetadata so it can be added to the datastore
		if err := d.datastore.Add(&hardware); err != nil {
			return AddHardwareResult{}, errors.Join(
				fmt.Errorf("unable to add hardware to inventory datastore"),
				err,
			)
		}

		pair := HardwareLocationPair{
			Hardware: hardware,
		}

		log.Debug().Msgf("Also adding %s--a %s--with parent %s", pair.Hardware.ID, pair.Hardware.Type, pair.Hardware.Parent)
		result.AddedHardware = append(result.AddedHardware, pair)

		// Validate the current state of CANI's inventory data against the provider plugin
		// for provider specific data.
		if failedValidations, err := d.externalInventoryProvider.ValidateInternal(cmd, args, d.datastore, false); len(failedValidations) > 0 {
			result.ProviderValidationErrors = failedValidations
			return result, provider.ErrDataValidationFailure
		} else if err != nil {
			return AddHardwareResult{}, errors.Join(
				fmt.Errorf("failed to validate inventory against inventory provider plugin"),
				err,
			)
		}
	}

	return result, d.datastore.Flush()
}
