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
package nautobot

import (
	"context"
	"testing"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	imprt "github.com/Cray-HPE/cani/pkg/provider/nautobot/import"
)

func TestNew(t *testing.T) {
	p := New()

	if p == nil {
		t.Fatal("expected non-nil provider")
	}
	if p.Options == nil {
		t.Fatal("expected non-nil Options")
	}
	if p.Options.URL != "http://localhost:8081/api" {
		t.Errorf("URL = %q, want %q", p.Options.URL, "http://localhost:8081/api")
	}
	if p.Options.Import == nil {
		t.Error("expected non-nil Import opts")
	}
	if p.Options.Export == nil {
		t.Error("expected non-nil Export opts")
	}
}

func TestSlug(t *testing.T) {
	p := New()
	if got := p.Slug(); got != "nautobot" {
		t.Errorf("Slug() = %q, want %q", got, "nautobot")
	}
}

func TestClearRawData(t *testing.T) {
	p := New()
	p.rawLocations = []nautobotapi.Location{{Name: "loc"}}
	p.rawRacks = []nautobotapi.Rack{{Name: "rack"}}
	p.rawDevices = []nautobotapi.Device{{}}
	p.rawDeviceTypes = []nautobotapi.DeviceType{{}}
	p.rawInterfaces = []nautobotapi.Interface{{}}
	p.rawModules = []nautobotapi.Module{{}}
	p.rawModuleBays = []nautobotapi.ModuleBay{{}}
	p.rawCables = []nautobotapi.Cable{{}}
	p.rawInventoryItems = []nautobotapi.InventoryItem{{}}
	p.rawStatuses = []nautobotapi.Status{{}}
	p.rawRoles = []nautobotapi.Role{{}}

	p.ClearRawData()

	if p.rawLocations != nil {
		t.Error("rawLocations not cleared")
	}
	if p.rawRacks != nil {
		t.Error("rawRacks not cleared")
	}
	if p.rawDevices != nil {
		t.Error("rawDevices not cleared")
	}
	if p.rawDeviceTypes != nil {
		t.Error("rawDeviceTypes not cleared")
	}
	if p.rawInterfaces != nil {
		t.Error("rawInterfaces not cleared")
	}
	if p.rawModules != nil {
		t.Error("rawModules not cleared")
	}
	if p.rawModuleBays != nil {
		t.Error("rawModuleBays not cleared")
	}
	if p.rawCables != nil {
		t.Error("rawCables not cleared")
	}
	if p.rawInventoryItems != nil {
		t.Error("rawInventoryItems not cleared")
	}
	if p.rawStatuses != nil {
		t.Error("rawStatuses not cleared")
	}
	if p.rawRoles != nil {
		t.Error("rawRoles not cleared")
	}
}

func TestSetRawData(t *testing.T) {
	p := New()

	d := imprt.RawData{
		Locations:      []nautobotapi.Location{{Name: "site-a"}},
		Racks:          []nautobotapi.Rack{{Name: "rack-1"}},
		Devices:        []nautobotapi.Device{{}},
		DeviceTypes:    []nautobotapi.DeviceType{{Model: "DL380"}},
		Interfaces:     []nautobotapi.Interface{{Name: "eth0"}},
		Modules:        []nautobotapi.Module{{}},
		ModuleBays:     []nautobotapi.ModuleBay{{Name: "bay-0"}},
		Cables:         []nautobotapi.Cable{{}},
		InventoryItems: []nautobotapi.InventoryItem{{Name: "gpu-0"}},
		Statuses:       []nautobotapi.Status{{Name: "Active"}},
		Roles:          []nautobotapi.Role{{Name: "Compute"}},
	}

	p.SetRawData(d)

	if len(p.rawLocations) != 1 || p.rawLocations[0].Name != "site-a" {
		t.Error("rawLocations not set correctly")
	}
	if len(p.rawRacks) != 1 || p.rawRacks[0].Name != "rack-1" {
		t.Error("rawRacks not set correctly")
	}
	if len(p.rawDevices) != 1 {
		t.Error("rawDevices not set correctly")
	}
	if len(p.rawDeviceTypes) != 1 || p.rawDeviceTypes[0].Model != "DL380" {
		t.Error("rawDeviceTypes not set correctly")
	}
	if len(p.rawInterfaces) != 1 || p.rawInterfaces[0].Name != "eth0" {
		t.Error("rawInterfaces not set correctly")
	}
	if len(p.rawModules) != 1 {
		t.Error("rawModules not set correctly")
	}
	if len(p.rawModuleBays) != 1 || p.rawModuleBays[0].Name != "bay-0" {
		t.Error("rawModuleBays not set correctly")
	}
	if len(p.rawCables) != 1 {
		t.Error("rawCables not set correctly")
	}
	if len(p.rawInventoryItems) != 1 || p.rawInventoryItems[0].Name != "gpu-0" {
		t.Error("rawInventoryItems not set correctly")
	}
	if len(p.rawStatuses) != 1 || p.rawStatuses[0].Name != "Active" {
		t.Error("rawStatuses not set correctly")
	}
	if len(p.rawRoles) != 1 || p.rawRoles[0].Name != "Compute" {
		t.Error("rawRoles not set correctly")
	}
}

func TestGetContext(t *testing.T) {
	p := New()

	t.Run("nil context returns nil", func(t *testing.T) {
		if p.GetContext() != nil {
			t.Error("expected nil context")
		}
	})

	t.Run("returns stored context", func(t *testing.T) {
		ctx := context.Background()
		p.ctx = ctx
		if p.GetContext() != ctx {
			t.Error("GetContext() did not return stored context")
		}
	})
}
