package imprt

import (
	"context"
	"fmt"
	"net/http"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
)

// pageSize is the number of items to request per API page.
const pageSize = 100

func intPtr(v int) *int { return &v }

// FetchLocations retrieves all locations from the Nautobot API.
func FetchLocations(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Location, error) {
	var all []nautobotapi.Location
	offset := 0
	for {
		resp, err := client.DcimLocationsListWithResponse(ctx, &nautobotapi.DcimLocationsListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list locations: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list locations: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchRacks retrieves all racks from the Nautobot API.
func FetchRacks(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Rack, error) {
	var all []nautobotapi.Rack
	offset := 0
	for {
		resp, err := client.DcimRacksListWithResponse(ctx, &nautobotapi.DcimRacksListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list racks: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list racks: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchDevices retrieves all devices from the Nautobot API.
func FetchDevices(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Device, error) {
	var all []nautobotapi.Device
	offset := 0
	for {
		resp, err := client.DcimDevicesListWithResponse(ctx, &nautobotapi.DcimDevicesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list devices: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list devices: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchDeviceTypes retrieves all device types from the Nautobot API.
func FetchDeviceTypes(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.DeviceType, error) {
	var all []nautobotapi.DeviceType
	offset := 0
	for {
		resp, err := client.DcimDeviceTypesListWithResponse(ctx, &nautobotapi.DcimDeviceTypesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list device types: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list device types: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchInterfaces retrieves all interfaces from the Nautobot API.
func FetchInterfaces(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Interface, error) {
	var all []nautobotapi.Interface
	offset := 0
	for {
		resp, err := client.DcimInterfacesListWithResponse(ctx, &nautobotapi.DcimInterfacesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list interfaces: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list interfaces: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchModules retrieves all modules from the Nautobot API.
func FetchModules(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Module, error) {
	var all []nautobotapi.Module
	offset := 0
	for {
		resp, err := client.DcimModulesListWithResponse(ctx, &nautobotapi.DcimModulesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list modules: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list modules: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchModuleBays retrieves all module bays from the Nautobot API.
func FetchModuleBays(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.ModuleBay, error) {
	var all []nautobotapi.ModuleBay
	offset := 0
	for {
		resp, err := client.DcimModuleBaysListWithResponse(ctx, &nautobotapi.DcimModuleBaysListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list module bays: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list module bays: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchCables retrieves all cables from the Nautobot API.
func FetchCables(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Cable, error) {
	var all []nautobotapi.Cable
	offset := 0
	for {
		resp, err := client.DcimCablesListWithResponse(ctx, &nautobotapi.DcimCablesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list cables: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list cables: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchInventoryItems retrieves all inventory items (FRUs) from the Nautobot API.
func FetchInventoryItems(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.InventoryItem, error) {
	var all []nautobotapi.InventoryItem
	offset := 0
	for {
		resp, err := client.DcimInventoryItemsListWithResponse(ctx, &nautobotapi.DcimInventoryItemsListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list inventory items: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list inventory items: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchStatuses retrieves all statuses from the Nautobot API.
func FetchStatuses(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Status, error) {
	var all []nautobotapi.Status
	offset := 0
	for {
		resp, err := client.ExtrasStatusesListWithResponse(ctx, &nautobotapi.ExtrasStatusesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list statuses: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list statuses: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchRoles retrieves all roles from the Nautobot API.
func FetchRoles(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Role, error) {
	var all []nautobotapi.Role
	offset := 0
	for {
		resp, err := client.ExtrasRolesListWithResponse(ctx, &nautobotapi.ExtrasRolesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list roles: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list roles: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}
