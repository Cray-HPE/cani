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

func (d *Domain) AddBlade(ctx context.Context, deviceTypeSlug string, cabinetOrdinal, chassisOrdinal, slotOrdinal int) (AddHardwarePassback, error) {
	// Validate provided cabinet exists
	// TODO

	// Validate provided chassis exists

	// TODO this is just a stand in, just for testing
	cabinet := inventory.Hardware{
		ID:              uuid.New(),
		Type:            hardwaretypes.Cabinet,
		Status:          inventory.HardwareStatusProvisioned,
		LocationOrdinal: &cabinetOrdinal,
	}
	if err := d.datastore.Add(&cabinet); err != nil {
		return AddHardwarePassback{}, errors.Join(
			fmt.Errorf("unable to add cabinet hardware"),
			err,
		)

	}

	// TODO this is just a stand in, just for testing
	chassis := inventory.Hardware{
		Parent:          cabinet.ID,
		ID:              uuid.New(),
		Type:            hardwaretypes.Chassis,
		Status:          inventory.HardwareStatusProvisioned,
		LocationOrdinal: &slotOrdinal,
	}
	if err := d.datastore.Add(&chassis); err != nil {
		return AddHardwarePassback{}, errors.Join(
			fmt.Errorf("unable to add chassis hardware"),
			err,
		)

	}

	// chassisLocationPath := []inventory.LocationToken{
	// 	{HardwareType: hardwaretypes.Cabinet, Ordinal: cabinetOrdinal},
	// 	{HardwareType: hardwaretypes.Chassis, Ordinal: chassisOrdinal},
	// }

	// Verify the provided device type slug is a node blade
	deviceType, err := d.hardwareTypeLibrary.GetDeviceType(deviceTypeSlug)
	if err != nil {
		return AddHardwarePassback{}, err
	}
	if deviceType.HardwareType != hardwaretypes.NodeBlade {
		return AddHardwarePassback{}, fmt.Errorf("provided device hardware type (%s) is not a node blade", deviceTypeSlug) // TODO better error message
	}

	// Generate a hardware build out
	hardwareBuildOutItems, err := d.hardwareTypeLibrary.GetDefaultHardwareBuildOut(deviceTypeSlug, slotOrdinal, chassis.ID)
	if err != nil {
		return AddHardwarePassback{}, errors.Join(
			fmt.Errorf("unable to build default hardware build out for %s", deviceTypeSlug),
			err,
		)
	}

	var passback AddHardwarePassback

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
			return AddHardwarePassback{}, errors.Join(
				fmt.Errorf("unable to add hardware to inventory datastore"),
				err,
			)
		}

		hardwareLocation, err := d.datastore.GetLocation(hardware)
		if err != nil {
			panic(err)
		}

		passback.AddedHardware = append(passback.AddedHardware, HardwareLocationPair{
			Hardware: hardware,
			Location: hardwareLocation,
		})
		log.Debug().Str("path", hardwareLocation.String()).Msg("Datastore")

	}

	// Validate the current state of CANI's inventory data against the provider plugin
	// for provider specific data.
	if failedValidations, err := d.externalInventoryProvider.ValidateInternal(ctx, d.datastore, false); len(failedValidations) > 0 {
		passback.ProviderValidationErrors = failedValidations
		return passback, provider.ErrDataValidationFailure
	} else if err != nil {
		return AddHardwarePassback{}, errors.Join(
			fmt.Errorf("failed to validate inventory against inventory provider plugin"),
			err,
		)
	}

	return passback, d.datastore.Flush()
}

func (d *Domain) RemoveBlade(u uuid.UUID, recursion bool) error {
	err := d.datastore.Remove(u, recursion)
	if err != nil {
		return errors.Join(
			fmt.Errorf("unable to remove hardware from inventory datastore"),
			err,
		)
	}

	return d.datastore.Flush()
}
