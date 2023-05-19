package domain

import (
	"errors"
	"fmt"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
)

func (d *Domain) AddBlade(deviceTypeSlug string, cabinetOrdinal, chassisOrdinal, slotOrdinal int) error {
	// Validate provided cabinet exists
	// TODO

	// Validate provided chassis exists
	// TODO this is just a stand in, just for testing
	chassis := inventory.Hardware{
		ID:              uuid.New(),
		Type:            hardwaretypes.HardwareTypeChassis,
		Status:          inventory.HardwareStatusProvisioned,
		LocationOrdinal: &slotOrdinal,
	}
	if err := d.datastore.Add(&chassis); err != nil {
		return errors.Join(
			fmt.Errorf("unable to add chassis hardware"),
			err,
		)

	}

	// chassisLocationPath := []inventory.LocationToken{
	// 	{HardwareType: hardwaretypes.HardwareTypeCabinet, Ordinal: cabinetOrdinal},
	// 	{HardwareType: hardwaretypes.HardwareTypeChassis, Ordinal: chassisOrdinal},
	// }

	// Verify the provided device type slug is a node blade
	deviceType, err := d.hardwareTypeLibrary.GetDeviceType(deviceTypeSlug)
	if err != nil {
		return err
	}
	if deviceType.HardwareType != hardwaretypes.HardwareTypeNodeBlade {
		return fmt.Errorf("provided device hardware type (%s) is not a node blade", deviceTypeSlug) // TODO better error message
	}

	// Generate a hardware build out
	hardwareBuildOutItems, err := d.hardwareTypeLibrary.GetDefaultHardwareBuildOut(deviceTypeSlug, slotOrdinal, chassis.ID)
	if err != nil {
		return errors.Join(
			fmt.Errorf("unable to build default hardware build out for %s", deviceTypeSlug),
			err,
		)
	}

	for _, hardwareBuildOut := range hardwareBuildOutItems {
		// Generate the CANI hardware inventory version of the hardware build out data
		// TODO
		hardware := inventory.Hardware{
			ID:     hardwareBuildOut.ID,
			Parent: hardwareBuildOut.ParentID,
			Type:   hardwareBuildOut.DeviceType.HardwareType,
			Vendor: hardwareBuildOut.DeviceType.Manufacturer,
			Model:  hardwareBuildOut.DeviceType.Model,

			LocationOrdinal: &hardwareBuildOut.Ordinal,

			Status: inventory.HardwareStatusStaged,
		}

		// TODO need a check to see if all the needed information exists,
		// Things like role/subrole/nid/alias could be injected at a later time.
		// Not sure how hard it would be to specify at this point in time.
		// This command creates the physical information for a node, have another command for the logical part of the data
		if err := d.datastore.Add(&hardware); err != nil {
			return errors.Join(
				fmt.Errorf("unable to add hardware to inventory datastore"),
				err,
			)
		}
	}

	return d.datastore.Flush()
}

func (d *Domain) RemoveBlade(u uuid.UUID) error {
	err := d.datastore.Remove(u)
	if err != nil {
		return errors.Join(
			fmt.Errorf("unable to remove hardware from inventory datastore"),
			err,
		)
	}

	return d.datastore.Flush()
}
