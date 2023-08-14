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
	"context"
	"errors"
	"fmt"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// AddCabinet adds a cabinet to the inventory
func (d *Domain) AddCabinet(ctx context.Context, deviceTypeSlug string, cabinetOrdinal int, metadata map[string]interface{}) (AddHardwareResult, error) {
	// Validate provided cabinet exists
	// Craft the path to the cabinet
	cabinetLocationPath := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: cabinetOrdinal},
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
				fmt.Errorf("%s number %d is already in use", hardwaretypes.Cabinet, cabinetOrdinal),
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

	// Verify the provided device type slug is a cabinet
	deviceType, err := d.hardwareTypeLibrary.GetDeviceType(deviceTypeSlug)
	if err != nil {
		return AddHardwareResult{}, err
	}
	if deviceType.HardwareType != hardwaretypes.Cabinet {
		return AddHardwareResult{}, fmt.Errorf("provided device hardware type (%s) is not a %s", deviceTypeSlug, hardwaretypes.Cabinet) // TODO better error message
	}

	// Generate a hardware build out using the system as a parent
	hardwareBuildOutItems, err := inventory.GetDefaultHardwareBuildOut(d.hardwareTypeLibrary, deviceTypeSlug, cabinetOrdinal, system.ID)
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
		if err := d.externalInventoryProvider.BuildHardwareMetadata(&hardware, metadata); err != nil {
			return AddHardwareResult{}, err
		}

		log.Debug().Any("id", hardware.ID).Msg("Hardware")
		log.Debug().Str("path", hardwareBuildOut.LocationPath.String()).Msg("Hardware Build out")

		// TODO need a check to see if all the needed information exists,
		// Things like role/subrole/nid/alias could be injected at a later time.
		// Not sure how hard it would be to specify at this point in time.
		// This command creates the physical information for a node, have another command for the logical part of the data
		if err := d.datastore.Add(&hardware); err != nil {
			return AddHardwareResult{}, errors.Join(
				fmt.Errorf("unable to add hardware to inventory datastore"),
				err,
			)
		}

		hardwareLocation, err := d.datastore.GetLocation(hardware)
		if err != nil {
			panic(err)
		}

		result.AddedHardware = append(result.AddedHardware, HardwareLocationPair{
			Hardware: hardware,
			Location: hardwareLocation,
		})
		log.Debug().Str("path", hardwareLocation.String()).Msg("Datastore")

	}

	// Validate the current state of CANI's inventory data against the provider plugin
	// for provider specific data.
	if failedValidations, err := d.externalInventoryProvider.ValidateInternal(ctx, d.datastore, false); len(failedValidations) > 0 {
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

func (d *Domain) Recommend(deviceTypeSlug string) (recommendations provider.HardwareRecommendations, err error) {
	// Get the existing inventory
	inv, err := d.List()
	if err != nil {
		return recommendations, err
	}
	// Get recommendations from the CSM provider for the cabinet
	recommendations, err = d.externalInventoryProvider.RecommendCabinet(inv, deviceTypeSlug)
	if err != nil {
		return recommendations, err
	}
	return recommendations, nil
}
