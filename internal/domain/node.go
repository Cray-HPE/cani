package domain

import (
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
)

func (d *Domain) UpdateNode(cabinet, chassis, slot, bmc, node int, metadata map[string]interface{}) error {
	// Get the node object from the datastore
	cHardware, err := d.datastore.GetAtLocation(inventory.LocationPath{
		{hardwaretypes.HardwareTypeCabinet, cabinet},
		{hardwaretypes.HardwareTypeChassis, chassis},
		{hardwaretypes.HardwareTypeNodeBlade, slot},
		{hardwaretypes.HardwareTypeNodeCard, bmc}, // Yes I mean put the BMC location for the node card location
		{hardwaretypes.HardwareTypeNode, node},
	})

	if err != nil {
		return err
	}

	// Ask the inventory provider to craft a metadata object for this information
	if err := d.externalInventoryProvider.BuildHardwareMetadata(&cHardware, metadata); err != nil {
		return err
	}

	// Push it back into the data store
	if err := d.datastore.Update(&cHardware); err != nil {
		return err
	}

	return nil
}
