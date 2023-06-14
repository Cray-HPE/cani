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

	exists, err = bladePath.Exists(d.datastore)
	if err != nil {
		return AddHardwareResult{}, errors.Join(
			fmt.Errorf("unable to check if %s exists at %s", hardwaretypes.NodeBlade, bladePath),
			err,
		)
	}

	// error if it exists because a blade cannot be added if one is already in place
	if exists {
		return AddHardwareResult{},
			errors.Join(
				fmt.Errorf("%s number %d is already in use", hardwaretypes.NodeBlade, bladeOrdinal),
				fmt.Errorf("please re-run the command with an available %s number", hardwaretypes.NodeBlade),
				fmt.Errorf("try 'cani alpha list blade'"),
			)
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
	chassisID, err := chassisPath.GetUUID(d.datastore)
	if err != nil {
		return AddHardwareResult{}, err
	}

	// Generate a hardware build out
	hardwareBuildOutItems, err := d.hardwareTypeLibrary.GetDefaultHardwareBuildOut(deviceTypeSlug, bladeOrdinal, chassisID)
	if err != nil {
		return AddHardwareResult{}, errors.Join(
			fmt.Errorf("unable to build default hardware build out for %s", deviceTypeSlug),
			err,
		)
	}

	var result AddHardwareResult

	for _, hardwareBuildOut := range hardwareBuildOutItems {
		// Generate the CANI hardware inventory version of the hardware build out data
		// TODO

		locationOrdinal := hardwareBuildOut.OrdinalPath[len(hardwareBuildOut.OrdinalPath)-1]

		hardware := inventory.Hardware{
			ID:     hardwareBuildOut.ID,
			Parent: hardwareBuildOut.ParentID,
			Type:   hardwareBuildOut.DeviceType.HardwareType,
			Vendor: hardwareBuildOut.DeviceType.Manufacturer,
			Model:  hardwareBuildOut.DeviceType.Model,

			LocationOrdinal: &locationOrdinal,

			Status: inventory.HardwareStatusStaged,
		}

		log.Debug().Any("id", hardware.ID).Msg("Hardware")
		log.Debug().Str("path", hardwareBuildOut.LocationPathString()).Msg("Hardware Build out")

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
