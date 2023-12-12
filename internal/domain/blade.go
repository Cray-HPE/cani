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
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// AddBlade adds a blade to the inventory by crafting location paths from the given ordinals and generating a hardware buildout
func (d *Domain) AddBlade(cmd *cobra.Command, args []string, cabinetOrdinal, chassisOrdinal, bladeOrdinal int) (AddHardwareResult, error) {
	deviceTypeSlug := args[0]

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
			fmt.Errorf("try 'list cabinet'"),
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

	//
	// TODO EVERYTHING BELOW IS GENERIC CODE THAT SHOULD BE SHARED
	//

	var existingDescendantHardware []inventory.Hardware
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
		existingDescendantHardware, err = d.datastore.GetDescendants(existingBlade.ID)
		if err != nil {
			return AddHardwareResult{}, errors.Join(
				fmt.Errorf("unable to retrieve descents for %s %s at %v", hardwaretypes.NodeBlade, existingBlade.ID, bladePath),
				err,
			)
		}
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
	hardwareBuildOutItems, err := inventory.GenerateHardwareBuildOut(d.hardwareTypeLibrary, inventory.GenerateHardwareBuildOutOpts{
		DeviceTypeSlug: deviceTypeSlug,
		DeviceOrdinal:  bladeOrdinal,
		DeviceID:       existingBlade.ID, // If a existing piece of hardware existed this will be something other than uuid.Nil.

		ParentHardware: chassis,

		ExistingDescendantHardware: existingDescendantHardware,
	})
	if err != nil {
		return AddHardwareResult{}, errors.Join(
			fmt.Errorf("unable to build default hardware build out for %s", deviceTypeSlug),
			err,
		)
	}

	var result AddHardwareResult

	for _, hardwareBuildOut := range hardwareBuildOutItems {
		var hardware inventory.Hardware

		if hardwareBuildOut.ExistingHardware == nil {
			//
			// New hardware not present in the inventory
			//

			// Generate the CANI hardware inventory version of the hardware build out data
			hardware = inventory.NewHardwareFromBuildOut(hardwareBuildOut, inventory.HardwareStatusStaged)

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
		} else {
			//
			// Empty hardware is present in the inventory
			//

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

		hardwareLocation, err := d.datastore.GetLocation(hardware)
		if err != nil {
			panic(err)
		}

		pair := HardwareLocationPair{
			Hardware: hardware,
			Location: hardwareLocation,
		}
		pair.Hardware.LocationPath = hardwareLocation
		result.AddedHardware = append(result.AddedHardware, pair)
		log.Debug().Str("path", hardwareLocation.String()).Msg("Datastore")

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
