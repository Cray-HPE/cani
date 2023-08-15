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

// AddBlade adds a blade to the inventory by crafting location paths from the given ordinals and generating a hardware buildout
func (d *Domain) AddBlade(ctx context.Context, deviceTypeSlug string, cabinetOrdinal, chassisOrdinal, bladeOrdinal int) (AddHardwareResult, error) {
	// Check if the cabinet exists
	cabinetPath := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: cabinetOrdinal},
	}
	var exists bool
	var err error
	exists, err = cabinetPath.Exists(d.datastore)
	if err != nil {
		return AddHardwareResult{}, errors.Join(
			fmt.Errorf("unable to check if %s exists at %s", hardwaretypes.Cabinet, cabinetPath),
			err,
		)
	}

	// error if the cabinet does not exit (cannot add a blade if no cabinet exists)
	if !exists {
		return AddHardwareResult{}, errors.Join(
			fmt.Errorf("unable to find %s at %s", hardwaretypes.Cabinet, cabinetPath),
			fmt.Errorf("try 'go run main.go alpha list cabinet'"),
		)
	}

	// Check if the chassis exists
	chassisPath := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: cabinetOrdinal},
		{HardwareType: hardwaretypes.Chassis, Ordinal: chassisOrdinal},
	}

	exists, err = chassisPath.Exists(d.datastore)
	if err != nil {
		return AddHardwareResult{}, errors.Join(
			fmt.Errorf("unable to check if %s exists at %s", hardwaretypes.Chassis, chassisPath),
			err,
		)
	}

	// error if no chassis exists (cannot add a blade if a chassis does not exist)
	if !exists {
		return AddHardwareResult{},
			errors.Join(
				fmt.Errorf("in order to add a %s, a %s is needed", hardwaretypes.NodeBlade, hardwaretypes.Chassis),
				fmt.Errorf("unable to find %s at %s", hardwaretypes.Chassis, chassisPath),
			)
	}

	// Check if the blade exists
	bladePath := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: cabinetOrdinal},
		{HardwareType: hardwaretypes.Chassis, Ordinal: chassisOrdinal},
		{HardwareType: hardwaretypes.NodeBlade, Ordinal: bladeOrdinal},
	}

	var existingHardware []inventory.Hardware
	existingBlade, err := d.datastore.GetAtLocation(bladePath)
	if errors.Is(err, inventory.ErrHardwareNotFound) {
		// Hardware does not exist, this is fine!
		log.Debug().Msgf("No %s exists at %s", hardwaretypes.NodeBlade, bladePath)
	} else if err != nil {
		return AddHardwareResult{}, errors.Join(
			fmt.Errorf("unable to check if %s exists at %s", hardwaretypes.NodeBlade, bladePath),
			err,
		)
	} else if existingBlade.Status != inventory.HardwareStatusEmpty {
		return AddHardwareResult{},
			errors.Join(
				fmt.Errorf("%s number %d is already in use", hardwaretypes.NodeBlade, bladeOrdinal),
				fmt.Errorf("please re-run the command with an available %s number", hardwaretypes.NodeBlade),
				fmt.Errorf("try 'cani alpha list blade'"),
			)
	} else {
		// Hardware exists in inventory as empty
		log.Debug().Msgf("%s exists at %s with status %s", hardwaretypes.NodeBlade, bladePath, existingBlade.Status)

		// Retrieve the child hardware of this blade
		existingChildHardware, err := d.datastore.GetDescendents(existingBlade.ID)
		if err != nil {
			return AddHardwareResult{}, errors.Join(
				fmt.Errorf("unable to retrieve descents for %s %s at %v", hardwaretypes.NodeBlade, existingBlade.ID, bladePath),
				err,
			)
		}

		// Build up slice of hardware!
		existingHardware = append(existingHardware, existingBlade)
		existingHardware = append(existingHardware, existingChildHardware...)

		// Need a LP to UUID lookup map to specify ID overrides for GetDefaultBuildOut

		//
	}

	// Verify the provided device type slug is a node blade
	deviceType, err := d.hardwareTypeLibrary.GetDeviceType(deviceTypeSlug)
	if err != nil {
		return AddHardwareResult{}, err
	}
	if deviceType.HardwareType != hardwaretypes.NodeBlade {
		return AddHardwareResult{}, fmt.Errorf("provided device hardware type (%s) is not a %s", deviceTypeSlug, hardwaretypes.NodeBlade)
	}

	// Get the chassis ID, since it is needed as an arg to the hardware buildout so the blade is added to the correct parent device
	chassis, err := chassisPath.Get(d.datastore)
	if err != nil {
		return AddHardwareResult{}, err
	}

	// Generate a hardware build out
	hardwareBuildOutItems, err := inventory.GenerateDefaultHardwareBuildOut(d.hardwareTypeLibrary, deviceTypeSlug, bladeOrdinal, chassis)
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

func (d *Domain) RemoveBlade(u uuid.UUID, recursion bool) error {
	err := d.datastore.Remove(u, recursion)
	if err != nil {
		return errors.Join(
			fmt.Errorf("unable to remove %s from inventory datastore", hardwaretypes.NodeBlade),
			err,
		)
	}

	return d.datastore.Flush()
}
