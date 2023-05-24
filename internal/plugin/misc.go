package plugin

import (
	"github.com/Cray-HPE/cani/internal/inventory"
)

// List returns the inventory
func (p *Plugin) List() (inventory.Inventory, error) {
	inv, err := p.datastore.List()
	if err != nil {
		return inventory.Inventory{}, err
	}

	return inv, nil
}
