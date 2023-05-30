package domain

import (
	"context"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/rs/zerolog/log"
)

// List returns the inventory
func (d *Domain) List() (inventory.Inventory, error) {
	inv, err := d.datastore.List()
	if err != nil {
		return inventory.Inventory{}, err
	}

	return inv, nil
}

func (d *Domain) Validate() error {
	err := d.externalInventoryProvider.ValidateExternal(context.Background())
	if err != nil {
		return err
	}

	log.Info().Msg("Validated external inventory provider")
	return nil
}
