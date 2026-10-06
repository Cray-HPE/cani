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
	"github.com/Cray-HPE/cani/pkg/provider/nautobot/transform"
)

// RawData is the snapshot of API collections Import hands to the transform
// stage. It is the transform package's input type, so the two stages share
// one definition and a new collection is declared once.
type RawData = transform.RawData

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
	d, err := fetchAll(prov.GetContext(), prov.GetClient())
	if err != nil {
		return err
	}

	prov.ClearRawData()
	prov.SetRawData(d)

	return nil
}

// fetchAll fetches every collection the transform needs, in order. The first
// failure aborts the import before any raw data is stored.
func fetchAll(ctx context.Context, client *nautobotapi.ClientWithResponses) (RawData, error) {
	var d RawData
	steps := []struct {
		name  string
		fetch func() error
	}{
		{"locations", into(ctx, client, &d.Locations, FetchLocations)},
		{"location types", into(ctx, client, &d.LocationTypes, FetchLocationTypes)},
		{"racks", into(ctx, client, &d.Racks, FetchRacks)},
		{"devices", into(ctx, client, &d.Devices, FetchDevices)},
		{"device types", into(ctx, client, &d.DeviceTypes, FetchDeviceTypes)},
		{"interfaces", into(ctx, client, &d.Interfaces, FetchInterfaces)},
		{"modules", into(ctx, client, &d.Modules, FetchModules)},
		{"module bays", into(ctx, client, &d.ModuleBays, FetchModuleBays)},
		{"cables", into(ctx, client, &d.Cables, FetchCables)},
		{"inventory items", into(ctx, client, &d.InventoryItems, FetchInventoryItems)},
		{"statuses", into(ctx, client, &d.Statuses, FetchStatuses)},
		{"roles", into(ctx, client, &d.Roles, FetchRoles)},
		{"vlans", into(ctx, client, &d.VLANs, FetchVLANs)},
		{"VLAN location assignments", into(ctx, client, &d.VLANLocationAssignments, FetchVLANLocationAssignments)},
		{"prefixes", into(ctx, client, &d.Prefixes, FetchPrefixes)},
		{"prefix location assignments", into(ctx, client, &d.PrefixLocationAssignments, FetchPrefixLocationAssignments)},
		{"ip addresses", into(ctx, client, &d.IPAddresses, FetchIPAddresses)},
		{"vrfs", into(ctx, client, &d.VRFs, FetchVRFs)},
		{"vrf-device-assignments", into(ctx, client, &d.VRFDeviceAssignments, FetchVRFDeviceAssignments)},
		{"vrf-prefix-assignments", into(ctx, client, &d.VRFPrefixAssignments, FetchVRFPrefixAssignments)},
		{"namespaces", into(ctx, client, &d.Namespaces, FetchNamespaces)},
	}
	for _, step := range steps {
		if err := step.fetch(); err != nil {
			return RawData{}, fmt.Errorf("fetching %s: %w", step.name, err)
		}
	}
	return d, nil
}

// into binds a collection fetcher to the RawData field it fills.
func into[T any](
	ctx context.Context,
	client *nautobotapi.ClientWithResponses,
	dst *[]T,
	fetch func(context.Context, *nautobotapi.ClientWithResponses) ([]T, error),
) func() error {
	return func() (err error) {
		*dst, err = fetch(ctx, client)
		return err
	}
}
