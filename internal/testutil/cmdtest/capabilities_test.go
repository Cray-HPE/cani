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

package cmdtest_test

import (
	"context"
	"maps"
	"testing"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

const capabilityFlag = "test-provider-value"

type inertProvider struct{ slug string }

func (provider inertProvider) Transform(_ context.Context, _ devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	return &devicetypes.TransformResult{}, nil
}

func (provider inertProvider) NewProviderCmd(_ *cli.Command) (*cli.Command, error) {
	return nil, nil
}

func (provider inertProvider) Slug() string { return provider.slug }

type rackHookCall struct {
	rack      *devicetypes.CaniRackType
	inventory *devicetypes.Inventory
}

type deviceUpdateCall struct {
	command *cli.Command
	device  *devicetypes.CaniDeviceType
}

type metadataCall struct {
	target *map[string]any
	values map[string]string
}

type stageCall struct {
	inventory *devicetypes.Inventory
	slug      string
}

type capabilityLog struct {
	rackHooks    []rackHookCall
	registration []*cli.Command
	updates      []deviceUpdateCall
	metadata     []metadataCall
	descriptions []*devicetypes.CaniDeviceType
	newStages    []stageCall
	existing     []stageCall
}

type capableProvider struct {
	inertProvider
	calls         capabilityLog
	applyMetadata func(*map[string]any, map[string]string)
	stageNew      func(*devicetypes.Inventory, string) bool
	stageExisting func(*devicetypes.Inventory, string) bool
}

var invarianceProvider = &capableProvider{inertProvider: inertProvider{slug: "invariance-capable"}}

func (provider *capableProvider) OnRackAdded(rack *devicetypes.CaniRackType, inventory *devicetypes.Inventory) error {
	provider.calls.rackHooks = append(provider.calls.rackHooks, rackHookCall{rack, inventory})
	return nil
}

func (provider *capableProvider) ApplyMetadata(target *map[string]any, values map[string]string) {
	provider.calls.metadata = append(provider.calls.metadata, metadataCall{target, maps.Clone(values)})
	if provider.applyMetadata != nil {
		provider.applyMetadata(target, values)
	}
}

func (provider *capableProvider) DescribeStagedDevice(device *devicetypes.CaniDeviceType) []string {
	provider.calls.descriptions = append(provider.calls.descriptions, device)
	return nil
}

func (provider *capableProvider) RegisterDeviceUpdateFlags(command *cli.Command) {
	provider.calls.registration = append(provider.calls.registration, command)
	command.Flags().String(capabilityFlag, "", "Test provider value")
}

func (provider *capableProvider) ApplyDeviceUpdateFlags(command *cli.Command, device *devicetypes.CaniDeviceType) error {
	provider.calls.updates = append(provider.calls.updates, deviceUpdateCall{command, device})
	return nil
}

func (provider *capableProvider) StageNewInRack(inventory *devicetypes.Inventory, slug string) bool {
	provider.calls.newStages = append(provider.calls.newStages, stageCall{inventory, slug})
	return provider.stageNew != nil && provider.stageNew(inventory, slug)
}

func (provider *capableProvider) StageExisting(inventory *devicetypes.Inventory, slug string) bool {
	provider.calls.existing = append(provider.calls.existing, stageCall{inventory, slug})
	return provider.stageExisting != nil && provider.stageExisting(inventory, slug)
}

func useCapabilities(t *testing.T) *capableProvider {
	t.Helper()
	registered := provider.GetProvider(invarianceProvider.Slug())
	if registered == nil {
		provider.Register(invarianceProvider.Slug(), invarianceProvider)
	} else if registered != invarianceProvider {
		t.Fatal("capability provider registration was replaced")
	}
	previous := *invarianceProvider
	*invarianceProvider = capableProvider{inertProvider: previous.inertProvider}
	t.Cleanup(func() { *invarianceProvider = previous })
	return invarianceProvider
}

func assertCallCount(t *testing.T, method string, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("%s calls = %d, want %d", method, got, want)
	}
}

func assertCRUDCapabilities(t *testing.T, calls capabilityLog) {
	t.Helper()
	for method, count := range map[string]int{
		"OnRackAdded":               len(calls.rackHooks),
		"RegisterDeviceUpdateFlags": len(calls.registration),
		"ApplyDeviceUpdateFlags":    len(calls.updates),
	} {
		assertCallCount(t, method, count, 1)
	}
	rackCall := calls.rackHooks[0]
	if rackCall.rack.Name != "rack-99" || rackCall.inventory.Racks[rackCall.rack.ID] != rackCall.rack {
		t.Error("OnRackAdded did not receive the newly added rack and its inventory")
	}
	if calls.registration[0].Name() != "device" {
		t.Error("RegisterDeviceUpdateFlags did not receive the device command")
	}
	updateCall := calls.updates[0]
	if updateCall.command != calls.registration[0] || updateCall.device.Name != "cn-99" {
		t.Error("ApplyDeviceUpdateFlags did not receive the registered command and renamed device")
	}
}
