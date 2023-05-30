package domain

import (
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/rs/zerolog/log"
)

func (d *Domain) UpdateNode(cabinet, chassis, slot, bmc, node int, metadata map[string]interface{}) error {
	// Get the node object from the datastore
	locationPath := inventory.LocationPath{
		{HardwareType: hardwaretypes.Cabinet, Ordinal: cabinet},
		{HardwareType: hardwaretypes.Chassis, Ordinal: chassis},
		{HardwareType: hardwaretypes.NodeBlade, Ordinal: slot},
		{HardwareType: hardwaretypes.NodeCard, Ordinal: bmc}, // Yes I mean put the BMC location for the node card location
		{HardwareType: hardwaretypes.Node, Ordinal: node},
	}
	hw, err := d.datastore.GetAtLocation(locationPath)
	if err != nil {
		return err
	}

	log.Debug().Msgf("Found node at: %s with ID (%s)", locationPath, hw.ID)

	// Ask the inventory provider to craft a metadata object for this information
	if err := d.externalInventoryProvider.BuildHardwareMetadata(&hw, metadata); err != nil {
		return err
	}

	log.Debug().Any("metadata", hw.ProviderProperties).Msg("Provider Properties")

	// Push it back into the data store
	if err := d.datastore.Update(&hw); err != nil {
		return err
	}

	return d.datastore.Flush()
}
