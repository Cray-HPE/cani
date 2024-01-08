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

	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// AddCabinet adds a cabinet to the inventory
func (d *Domain) AddCabinet(cmd *cobra.Command, args []string, recommendations provider.HardwareRecommendations) (AddHardwareResult, error) {
	// Validate provided cabinet exists
	// Craft the path to the cabinet
	cabinetLocationPath := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: recommendations.CabinetOrdinal},
	}

	// Check if the cabinet already exists
	exists, err := cabinetLocationPath.Exists(d.datastore)
	if err != nil {
		return AddHardwareResult{}, errors.Join(
			fmt.Errorf("unable to check if %s exists", hardwaretypes.Cabinet),
			err,
		)
	}
	// Fail if the cabinet already exists and provide actionable error message
	if exists {
		return AddHardwareResult{},
			errors.Join(
				fmt.Errorf("%s number %d is already in use", hardwaretypes.Cabinet, recommendations.CabinetOrdinal),
				fmt.Errorf("please re-run the command with an available %s number", hardwaretypes.Cabinet),
			)
	}

	system, err := d.datastore.GetSystemZero()
	if err != nil {
		return AddHardwareResult{}, errors.Join(
			fmt.Errorf("unable to get system zero"),
			err,
		)
	}

	deviceTypeSlug := args[0]
	// Verify the provided device type slug is a cabinet
	deviceType, err := d.hardwareTypeLibrary.GetDeviceType(deviceTypeSlug)
	if err != nil {
		return AddHardwareResult{}, err
	}
	if deviceType.HardwareType != hardwaretypes.Cabinet {
		return AddHardwareResult{}, fmt.Errorf("provided device hardware type (%s) is not a %s", deviceTypeSlug, hardwaretypes.Cabinet) // TODO better error message
	}

	// Generate a hardware build out using the system as a parent
	hardwareBuildOutItems, err := inventory.GenerateDefaultHardwareBuildOut(d.hardwareTypeLibrary, deviceTypeSlug, recommendations.CabinetOrdinal, system)
	if err != nil {
		return AddHardwareResult{}, errors.Join(
			fmt.Errorf("unable to build default hardware build out for %s", deviceTypeSlug),
			err,
		)
	}

	var result AddHardwareResult

	for _, hardwareBuildOut := range hardwareBuildOutItems {
		// Generate the CANI hardware inventory version of the hardware build out data
		hardware := inventory.NewHardwareFromBuildOut(hardwareBuildOut, inventory.HardwareStatusStaged)

		// Ask the inventory provider to craft a metadata object for this information
		if err := d.externalInventoryProvider.BuildHardwareMetadata(&hardware, cmd, args, recommendations); err != nil {
			return AddHardwareResult{}, err
		}

		log.Debug().Any("id", hardware.ID).Msg("Hardware")
		log.Debug().Str("path", hardwareBuildOut.LocationPath.String()).Msg("Hardware Build out")

		// Metadata is now set by the BuildHardwareMetadata so it can be added to the datastore
		if err := d.datastore.Add(&hardware); err != nil {
			return AddHardwareResult{}, errors.Join(
				fmt.Errorf("unable to add hardware to inventory datastore"),
				err,
			)
		}

		hlp := HardwareLocationPair{
			Hardware: hardware,
		}

		result.AddedHardware = append(result.AddedHardware, hlp)

		if d.Provider == taxonomy.CSM {
			hardwareLocation, err := d.datastore.GetLocation(hardware)
			if err != nil {
				panic(err)
			}
			hlp.Location = hardwareLocation
			log.Debug().Str("path", hardwareLocation.String()).Msg("Datastore")
		}
	}

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

	return result, d.datastore.Flush()
}

func (d *Domain) RemoveCabinet(u uuid.UUID, recursion bool) error {
	err := d.datastore.Remove(u, recursion)
	if err != nil {
		return errors.Join(
			fmt.Errorf("unable to remove %s from inventory datastore", hardwaretypes.Cabinet),
			err,
		)
	}

	return d.datastore.Flush()
}

func (d *Domain) Recommend(cmd *cobra.Command, args []string, auto bool) (recommendations provider.HardwareRecommendations, err error) {
	// Get the existing inventory
	inv, err := d.List()
	if err != nil {
		return recommendations, err
	}
	// Get recommendations from the CSM provider for the cabinet
	recommendations, err = d.externalInventoryProvider.RecommendHardware(inv, cmd, args, auto)
	if err != nil {
		return recommendations, err
	}
	return recommendations, nil
}
