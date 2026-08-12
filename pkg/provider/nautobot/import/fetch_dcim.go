/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package imprt

import (
	"context"
	"fmt"
	"net/http"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
)

// FetchLocations retrieves all locations from the Nautobot API.
func FetchLocations(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Location, error) {
	return paginate(ctx, "locations", func(ctx context.Context, offset int) (pageResult[nautobotapi.Location], error) {
		resp, err := client.DcimLocationsListWithResponse(ctx, &nautobotapi.DcimLocationsListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return pageResult[nautobotapi.Location]{}, err
		}
		return stdPage(resp.StatusCode(), resp.JSON200, "locations",
			func(b *nautobotapi.PaginatedLocationList) ([]nautobotapi.Location, *string) {
				return b.Results, b.Next
			})
	})
}

// FetchRacks retrieves all racks from the Nautobot API.
func FetchRacks(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Rack, error) {
	return paginate(ctx, "racks", func(ctx context.Context, offset int) (pageResult[nautobotapi.Rack], error) {
		resp, err := client.DcimRacksListWithResponse(ctx, &nautobotapi.DcimRacksListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return pageResult[nautobotapi.Rack]{}, err
		}
		return stdPage(resp.StatusCode(), resp.JSON200, "racks",
			func(b *nautobotapi.PaginatedRackList) ([]nautobotapi.Rack, *string) {
				return b.Results, b.Next
			})
	})
}

// FetchDevices retrieves all devices from the Nautobot API.
func FetchDevices(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Device, error) {
	return paginate(ctx, "devices", func(ctx context.Context, offset int) (pageResult[nautobotapi.Device], error) {
		resp, err := client.DcimDevicesListWithResponse(ctx, &nautobotapi.DcimDevicesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return pageResult[nautobotapi.Device]{}, err
		}
		return stdPage(resp.StatusCode(), resp.JSON200, "devices",
			func(b *nautobotapi.PaginatedDeviceList) ([]nautobotapi.Device, *string) {
				return b.Results, b.Next
			})
	})
}

// FetchDeviceTypes retrieves all device types from the Nautobot API.
func FetchDeviceTypes(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.DeviceType, error) {
	return paginate(ctx, "device types", func(ctx context.Context, offset int) (pageResult[nautobotapi.DeviceType], error) {
		resp, err := client.DcimDeviceTypesListWithResponse(ctx, &nautobotapi.DcimDeviceTypesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return pageResult[nautobotapi.DeviceType]{}, err
		}
		return stdPage(resp.StatusCode(), resp.JSON200, "device types",
			func(b *nautobotapi.PaginatedDeviceTypeList) ([]nautobotapi.DeviceType, *string) {
				return b.Results, b.Next
			})
	})
}

// FetchInterfaces retrieves all interfaces from the Nautobot API.
func FetchInterfaces(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Interface, error) {
	return paginate(ctx, "interfaces", func(ctx context.Context, offset int) (pageResult[nautobotapi.Interface], error) {
		resp, err := client.DcimInterfacesListWithResponse(ctx, &nautobotapi.DcimInterfacesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return pageResult[nautobotapi.Interface]{}, err
		}
		return stdPage(resp.StatusCode(), resp.JSON200, "interfaces",
			func(b *nautobotapi.PaginatedInterfaceList) ([]nautobotapi.Interface, *string) {
				return b.Results, b.Next
			})
	})
}

// FetchModules retrieves all modules from the Nautobot API.
func FetchModules(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Module, error) {
	return paginate(ctx, "modules", func(ctx context.Context, offset int) (pageResult[nautobotapi.Module], error) {
		resp, err := client.DcimModulesListWithResponse(ctx, &nautobotapi.DcimModulesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return pageResult[nautobotapi.Module]{}, err
		}
		return stdPage(resp.StatusCode(), resp.JSON200, "modules",
			func(b *nautobotapi.PaginatedModuleList) ([]nautobotapi.Module, *string) {
				return b.Results, b.Next
			})
	})
}

// FetchModuleBays retrieves all module bays from the Nautobot API.
func FetchModuleBays(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.ModuleBay, error) {
	return paginate(ctx, "module bays", func(ctx context.Context, offset int) (pageResult[nautobotapi.ModuleBay], error) {
		resp, err := client.DcimModuleBaysListWithResponse(ctx, &nautobotapi.DcimModuleBaysListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return pageResult[nautobotapi.ModuleBay]{}, err
		}
		return stdPage(resp.StatusCode(), resp.JSON200, "module bays",
			func(b *nautobotapi.PaginatedModuleBayList) ([]nautobotapi.ModuleBay, *string) {
				return b.Results, b.Next
			})
	})
}

// FetchCables retrieves all cables from the Nautobot API.
func FetchCables(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Cable, error) {
	return paginate(ctx, "cables", func(ctx context.Context, offset int) (pageResult[nautobotapi.Cable], error) {
		resp, err := client.DcimCablesListWithResponse(ctx, &nautobotapi.DcimCablesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return pageResult[nautobotapi.Cable]{}, err
		}
		return stdPage(resp.StatusCode(), resp.JSON200, "cables",
			func(b *nautobotapi.PaginatedCableList) ([]nautobotapi.Cable, *string) {
				return b.Results, b.Next
			})
	})
}

// FetchInventoryItems retrieves all inventory items (FRUs) from the Nautobot API.
func FetchInventoryItems(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.InventoryItem, error) {
	return paginate(ctx, "inventory items", func(ctx context.Context, offset int) (pageResult[nautobotapi.InventoryItem], error) {
		resp, err := client.DcimInventoryItemsListWithResponse(ctx, &nautobotapi.DcimInventoryItemsListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return pageResult[nautobotapi.InventoryItem]{}, err
		}
		return stdPage(resp.StatusCode(), resp.JSON200, "inventory items",
			func(b *nautobotapi.PaginatedInventoryItemList) ([]nautobotapi.InventoryItem, *string) {
				return b.Results, b.Next
			})
	})
}

// stdPage converts a status code and typed paginated body into a pageResult.
func stdPage[T any, B any](status int, body *B, noun string, extract func(*B) ([]T, *string)) (pageResult[T], error) {
	if status != http.StatusOK {
		return pageResult[T]{}, fmt.Errorf("list %s: status %d", noun, status)
	}
	if body == nil {
		return pageResult[T]{Done: true}, nil
	}
	items, next := extract(body)
	done := next == nil || *next == ""
	return pageResult[T]{Items: items, Done: done}, nil
}
