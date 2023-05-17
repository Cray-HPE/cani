package domain

import (
	"fmt"
	"sort"

	"github.com/Cray-HPE/cani/internal/cani/inventory"
	hardware_type_library "github.com/Cray-HPE/cani/pkg/hardware-type-library"
)

// ListSupportedTypes prints a list of supported hardware models
func (d *Domain) ListSupportedTypes(hwtype hardware_type_library.HardwareType) {
	// Extract the model names into a slice of strings
	models := []string{}

	for _, k := range d.hardwareTypeLibrary.GetDeviceTypesByHardwareType(hwtype) {
		models = append(models, k.Slug)
	}

	// Sort the models slice alphabetically
	sort.Strings(models)

	for _, model := range models {
		fmt.Printf("- %+v\n", model)
	}
}

// List returns the inventory
func (d *Domain) List() (inventory.Inventory, error) {
	inv, err := d.datastore.List()
	if err != nil {
		return inventory.Inventory{}, err
	}

	return inv, nil
}
