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
func (d *Domain) AddCabinet(ctx context.Context, deviceTypeSlug string, cabinetOrdinal int) (AddHardwareResult, error) {
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
			fmt.Errorf("unable to check if cabinet exists"),
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

	// Verify the provided device type slug is a node blade
	deviceType, err := d.hardwareTypeLibrary.GetDeviceType(deviceTypeSlug)
	if err != nil {
		return AddHardwareResult{}, err
	}
	if deviceType.HardwareType != hardwaretypes.Cabinet {
		return AddHardwareResult{}, fmt.Errorf("provided device hardware type (%s) is not a %s", deviceTypeSlug, hardwaretypes.Cabinet) // TODO better error message
	}

	// Generate a hardware build out using the system as a parent
	hardwareBuildOutItems, err := d.hardwareTypeLibrary.GetDefaultHardwareBuildOut(deviceTypeSlug, cabinetOrdinal, system.ID)
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
			ID:             hardwareBuildOut.ID,
			Parent:         hardwareBuildOut.ParentID,
			Type:           hardwareBuildOut.DeviceType.HardwareType,
			DeviceTypeSlug: hardwareBuildOut.DeviceType.Slug,
			Vendor:         hardwareBuildOut.DeviceType.Manufacturer,
			Model:          hardwareBuildOut.DeviceType.Model,

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

func (d *Domain) RemoveCabinet(u uuid.UUID, recursion bool) error {
	err := d.datastore.Remove(u, recursion)
	if err != nil {
		return errors.Join(
			fmt.Errorf("unable to remove hardware from inventory datastore"),
			err,
		)
	}

	return d.datastore.Flush()
}
