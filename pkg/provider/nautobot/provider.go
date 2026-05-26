/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
package nautobot

import (
	"context"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/Cray-HPE/cani/pkg/provider/nautobot/export"
	imprt "github.com/Cray-HPE/cani/pkg/provider/nautobot/import"
)

// Nautobot is the struct for the Nautobot provider
type Nautobot struct {
	Options  *NautobotOpts
	exporter *export.Exporter

	// ctx and client are populated during importNautobot() setup.
	ctx    context.Context
	client *export.NautobotClient
	cache  *export.LookupCache

	// Raw API responses stored during Import() for use by Transform().
	rawLocations      []nautobotapi.Location
	rawRacks          []nautobotapi.Rack
	rawDevices        []nautobotapi.Device
	rawDeviceTypes    []nautobotapi.DeviceType
	rawInterfaces     []nautobotapi.Interface
	rawModules        []nautobotapi.Module
	rawModuleBays     []nautobotapi.ModuleBay
	rawCables         []nautobotapi.Cable
	rawInventoryItems []nautobotapi.InventoryItem
	rawStatuses       []nautobotapi.Status
	rawRoles          []nautobotapi.Role
}

// New creates a new instance of the Nautobot provider
func New() *Nautobot {
	return &Nautobot{
		Options: &NautobotOpts{
			URL:    "http://localhost:8081/api",
			Import: &NautobotImportOpts{},
			Export: &NautobotExportOpts{},
		},
	}
}

// Slug returns the slug for the CANI provider
func (p *Nautobot) Slug() string {
	return "nautobot"
}

// ClearRawData resets the raw data storage for a fresh import.
func (p *Nautobot) ClearRawData() {
	p.rawLocations = nil
	p.rawRacks = nil
	p.rawDevices = nil
	p.rawDeviceTypes = nil
	p.rawInterfaces = nil
	p.rawModules = nil
	p.rawModuleBays = nil
	p.rawCables = nil
	p.rawInventoryItems = nil
	p.rawStatuses = nil
	p.rawRoles = nil
}

// SetRawData stores fetched raw data from the import phase.
func (p *Nautobot) SetRawData(d imprt.RawData) {
	p.rawLocations = d.Locations
	p.rawRacks = d.Racks
	p.rawDevices = d.Devices
	p.rawDeviceTypes = d.DeviceTypes
	p.rawInterfaces = d.Interfaces
	p.rawModules = d.Modules
	p.rawModuleBays = d.ModuleBays
	p.rawCables = d.Cables
	p.rawInventoryItems = d.InventoryItems
	p.rawStatuses = d.Statuses
	p.rawRoles = d.Roles
}

// GetClient returns the Nautobot API client for the import subpackage.
func (p *Nautobot) GetClient() *nautobotapi.ClientWithResponses {
	return p.client.ClientWithResponses
}

// GetContext returns the context for the import subpackage.
func (p *Nautobot) GetContext() context.Context {
	return p.ctx
}
