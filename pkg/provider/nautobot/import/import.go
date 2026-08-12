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
	"errors"
	"fmt"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
)

// RawData holds all raw API responses fetched during import.
type RawData struct {
	Locations            []nautobotapi.Location
	Racks                []nautobotapi.Rack
	Devices              []nautobotapi.Device
	DeviceTypes          []nautobotapi.DeviceType
	Interfaces           []nautobotapi.Interface
	Modules              []nautobotapi.Module
	ModuleBays           []nautobotapi.ModuleBay
	Cables               []nautobotapi.Cable
	InventoryItems       []nautobotapi.InventoryItem
	Statuses             []nautobotapi.Status
	Roles                []nautobotapi.Role
	VLANs                []nautobotapi.VLAN
	Prefixes             []nautobotapi.Prefix
	IPAddresses          []nautobotapi.IPAddress
	VRFs                 []nautobotapi.VRF
	VRFDeviceAssignments []nautobotapi.VRFDeviceAssignment
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
func Import(cmd *cli.Command, args []string, inventory *devicetypes.Inventory) error {
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

	d.VLANs, err = FetchVLANs(ctx, client)
	if err != nil {
		return fmt.Errorf("fetching vlans: %w", err)
	}

	d.Prefixes, err = FetchPrefixes(ctx, client)
	if err != nil {
		return fmt.Errorf("fetching prefixes: %w", err)
	}

	d.IPAddresses, err = FetchIPAddresses(ctx, client)
	if err != nil {
		return fmt.Errorf("fetching ip addresses: %w", err)
	}

	d.VRFs, err = FetchVRFs(ctx, client)
	if err != nil {
		return fmt.Errorf("fetching vrfs: %w", err)
	}

	d.VRFDeviceAssignments, err = FetchVRFDeviceAssignments(ctx, client)
	if err != nil {
		return fmt.Errorf("fetching vrf-device-assignments: %w", err)
	}

	prov.ClearRawData()
	prov.SetRawData(d)

	return nil
}
