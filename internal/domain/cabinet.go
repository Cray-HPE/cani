package domain

import (
	"errors"
	"fmt"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// AddCabinet adds a cabinet to the inventory
func (d *Domain) AddCabinet(deviceTypeSlug string, cabinetOrdinal int) ([]AddHardwareResult, error) {
	// Validate provided cabinet exists
	// TODO

	// TODO this is just a stand in, just for testing
	cabinet := inventory.Hardware{
		ID:              uuid.New(),
		Type:            hardwaretypes.Cabinet,
		Status:          inventory.HardwareStatusProvisioned,
		LocationOrdinal: &cabinetOrdinal,
	}
	if err := d.datastore.Add(&cabinet); err != nil {
		return nil, errors.Join(
			fmt.Errorf("unable to add cabinet hardware"),
			err,
		)

	}

	// Verify the provided device type slug is a node blade
	deviceType, err := d.hardwareTypeLibrary.GetDeviceType(deviceTypeSlug)
	if err != nil {
		return nil, err
	}
	if deviceType.HardwareType != hardwaretypes.Cabinet {
		return nil, fmt.Errorf("provided device hardware type (%s) is not a %s", deviceTypeSlug, hardwaretypes.Cabinet) // TODO better error message
	}

	// Generate a hardware build out
	// FIXME: no parent
	hardwareBuildOutItems, err := d.hardwareTypeLibrary.GetDefaultHardwareBuildOut(deviceTypeSlug, 0, uuid.UUID{})
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("unable to build default hardware build out for %s", deviceTypeSlug),
			err,
		)
	}

	var results []AddHardwareResult

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
			return nil, errors.Join(
				fmt.Errorf("unable to add hardware to inventory datastore"),
				err,
			)
		}

		hardwareLocation, err := d.datastore.GetLocation(hardware)
		if err != nil {
			panic(err)
		}

		results = append(results, AddHardwareResult{
			Hardware: hardware,
			Location: hardwareLocation,
		})
		log.Debug().Str("path", hardwareLocation.String()).Msg("Datastore")

	}

	return results, d.datastore.Flush()
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
