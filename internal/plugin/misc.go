package plugin

import (
	"github.com/Cray-HPE/cani/internal/inventory"
)

// List returns the inventory
func (d *Domain) List() (inventory.Inventory, error) {
	inv, err := d.datastore.List()
	if err != nil {
		return inventory.Inventory{}, err
	}

	return inv, nil
}
