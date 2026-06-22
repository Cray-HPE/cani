package imprt

import (
	"context"
	"errors"
	"fmt"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/spf13/cobra"
)

// RawData holds all raw API responses fetched during import.
type RawData struct {
	Locations      []nautobotapi.Location
	Racks          []nautobotapi.Rack
	Devices        []nautobotapi.Device
	DeviceTypes    []nautobotapi.DeviceType
	Interfaces     []nautobotapi.Interface
	Modules        []nautobotapi.Module
	ModuleBays     []nautobotapi.ModuleBay
	Cables         []nautobotapi.Cable
	InventoryItems []nautobotapi.InventoryItem
	Statuses       []nautobotapi.Status
	Roles          []nautobotapi.Role
}

// providerGetter is used to get the Nautobot singleton from the parent package.
// Set by the parent package's init() to break the import cycle.
var providerGetter func() interface {
	ClearRawData()
	SetRawData(RawData)
	GetClient() *nautobotapi.ClientWithResponses
	GetContext() context.Context
}

// SetProviderGetter allows the parent package to provide access to the singleton.
func SetProviderGetter(getter func() interface {
	ClearRawData()
	SetRawData(RawData)
	GetClient() *nautobotapi.ClientWithResponses
	GetContext() context.Context
}) {
	providerGetter = getter
}

// GetProvider returns the Nautobot singleton via the registered getter.
// It returns an error when the parent package has not registered a getter.
func GetProvider() (interface {
	ClearRawData()
	SetRawData(RawData)
	GetClient() *nautobotapi.ClientWithResponses
	GetContext() context.Context
}, error) {
	if providerGetter == nil {
		return nil, errors.New("providerGetter not set; ensure nautobot package init() calls SetProviderGetter")
	}
	return providerGetter(), nil
}

// Import fetches all entity types from the Nautobot API and stores
// the raw responses on the provider struct via the setter.
func Import(cmd *cobra.Command, args []string, inventory *devicetypes.Inventory) error {
	prov, err := GetProvider()
	if err != nil {
		return err
	}
	ctx := prov.GetContext()
	client := prov.GetClient()

	var d RawData

	d.Locations, err = FetchLocations(ctx, client)
	if err != nil {
		return fmt.Errorf("fetching locations: %w", err)
	}

	d.Racks, err = FetchRacks(ctx, client)
	if err != nil {
		return fmt.Errorf("fetching racks: %w", err)
	}

	d.Devices, err = FetchDevices(ctx, client)
	if err != nil {
		return fmt.Errorf("fetching devices: %w", err)
	}

	d.DeviceTypes, err = FetchDeviceTypes(ctx, client)
	if err != nil {
		return fmt.Errorf("fetching device types: %w", err)
	}

	d.Interfaces, err = FetchInterfaces(ctx, client)
	if err != nil {
		return fmt.Errorf("fetching interfaces: %w", err)
	}

	d.Modules, err = FetchModules(ctx, client)
	if err != nil {
		return fmt.Errorf("fetching modules: %w", err)
	}

	d.ModuleBays, err = FetchModuleBays(ctx, client)
	if err != nil {
		return fmt.Errorf("fetching module bays: %w", err)
	}

	d.Cables, err = FetchCables(ctx, client)
	if err != nil {
		return fmt.Errorf("fetching cables: %w", err)
	}

	d.InventoryItems, err = FetchInventoryItems(ctx, client)
	if err != nil {
		return fmt.Errorf("fetching inventory items: %w", err)
	}

	d.Statuses, err = FetchStatuses(ctx, client)
	if err != nil {
		return fmt.Errorf("fetching statuses: %w", err)
	}

	d.Roles, err = FetchRoles(ctx, client)
	if err != nil {
		return fmt.Errorf("fetching roles: %w", err)
	}

	prov.ClearRawData()
	prov.SetRawData(d)

	return nil
}
