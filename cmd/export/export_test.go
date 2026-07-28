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
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
 *  FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
 *  IN THE SOFTWARE.
 *
 */
package export

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

type fakeExportProvider struct {
	exportCalls int
	exportErr   error
	mutate      func(*devicetypes.Inventory)
}

func (provider *fakeExportProvider) Transform(context.Context, devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	return nil, nil
}

func (provider *fakeExportProvider) NewProviderCmd(*cli.Command) (*cli.Command, error) {
	return &cli.Command{}, nil
}

func (provider *fakeExportProvider) Slug() string {
	return "fake-export"
}

func (provider *fakeExportProvider) Export(_ context.Context, _ *cli.Command, _ []string, inventory *devicetypes.Inventory) error {
	provider.exportCalls++
	if provider.mutate != nil {
		provider.mutate(inventory)
	}
	return provider.exportErr
}

func prepareExportTest(t *testing.T, inventory *devicetypes.Inventory) (*cli.Command, string) {
	t.Helper()
	directory := t.TempDir()
	inventoryPath := filepath.Join(directory, "inventory.json")
	jsonStore := &datastores.JSONStore{Path: inventoryPath}
	if err := jsonStore.Save(inventory); err != nil {
		t.Fatalf("writing export test inventory: %v", err)
	}

	originalConfig := config.Cfg
	originalDatastore := datastores.Datastore
	config.Cfg = &config.Config{
		Path:      filepath.Join(directory, "config.yaml"),
		Datastore: inventoryPath,
	}
	t.Cleanup(func() {
		config.Cfg = originalConfig
		datastores.Datastore = originalDatastore
	})

	root := &cli.Command{Use: "cani"}
	root.PersistentFlags().String("datastore", "json", "")
	command := &cli.Command{Use: "fake-export"}
	root.AddCommand(command)
	return command, inventoryPath
}

func connectedExportInventory() (*devicetypes.Inventory, uuid.UUID) {
	deviceID := uuid.New()
	interfaceID := uuid.New()
	cableID := uuid.New()
	inventory := devicetypes.NewInventory()
	inventory.Devices[deviceID] = &devicetypes.CaniDeviceType{
		ID: deviceID, Name: "switch-01",
		Interfaces: []devicetypes.InterfaceSpec{{ID: interfaceID, Name: "port1"}},
	}
	inventory.Cables[cableID] = &devicetypes.CaniCableType{
		ID:                 cableID,
		Label:              "uplink",
		TerminationADevice: deviceID,
		TerminationAPort:   "port1",
		TerminationA:       interfaceID,
	}
	return inventory, deviceID
}

// TestRunExportRejectsInvalidCableBeforeProvider verifies relationship
// preflight aborts export without invoking the provider or rewriting inventory.
//
// Why it matters: provider phases may mutate remote objects before reaching
// cable creation, so endpoint errors must be detected at the command boundary.
// Inputs: firewall-01 with port1 and a cable naming nonexistent port3.
// Outputs: a contextual error, zero fake Export calls, and unchanged file bytes.
// Data choice: a valid device plus missing port reproduces the reported late
// Nautobot failure without involving provider-specific behavior.
func TestRunExportRejectsInvalidCableBeforeProvider(t *testing.T) {
	deviceID := uuid.New()
	inventory := devicetypes.NewInventory()
	inventory.Devices[deviceID] = &devicetypes.CaniDeviceType{
		ID: deviceID, Name: "firewall-01",
		Interfaces: []devicetypes.InterfaceSpec{{ID: uuid.New(), Name: "port1"}},
	}
	inventory.Cables[uuid.New()] = &devicetypes.CaniCableType{
		Label:              "invalid-uplink",
		TerminationBDevice: deviceID,
		TerminationBPort:   "port3",
	}
	command, inventoryPath := prepareExportTest(t, inventory)
	before, err := os.ReadFile(inventoryPath)
	if err != nil {
		t.Fatalf("reading inventory before export: %v", err)
	}
	provider := &fakeExportProvider{}

	err = runExport(command, nil, provider)

	if err == nil {
		t.Fatal("expected invalid cable preflight to fail export")
	}
	if !strings.Contains(err.Error(), `termination B port "port3" not found on device "firewall-01"`) {
		t.Errorf("runExport() error = %q, want endpoint context", err)
	}
	if provider.exportCalls != 0 {
		t.Errorf("provider Export calls = %d, want 0", provider.exportCalls)
	}
	after, err := os.ReadFile(inventoryPath)
	if err != nil {
		t.Fatalf("reading inventory after export: %v", err)
	}
	if !bytes.Equal(after, before) {
		t.Error("invalid export rewrote the inventory datastore")
	}
}

// TestRunExportInvokesProviderForValidInventory verifies relationship preflight
// allows a valid connected inventory through the normal export and save path.
//
// Why it matters: fail-fast validation must not block providers from exporting
// inventory whose endpoint tuples are coherent.
// Inputs: a device with port1 and a cable resolved to that interface.
// Outputs: runExport returns nil and invokes the fake exporter exactly once.
// Data choice: one complete side plus one absent side exercises the selected
// endpoint completeness policy at the real command boundary.
func TestRunExportInvokesProviderForValidInventory(t *testing.T) {
	inventory, _ := connectedExportInventory()
	command, _ := prepareExportTest(t, inventory)
	provider := &fakeExportProvider{}

	if err := runExport(command, nil, provider); err != nil {
		t.Fatalf("runExport() unexpected error: %v", err)
	}
	if provider.exportCalls != 1 {
		t.Errorf("provider Export calls = %d, want 1", provider.exportCalls)
	}
}

// TestRunExportPersistsExternalIDAfterProviderError verifies a provider error
// retains inventory mutations made before the failure.
//
// Why it matters: exporters may reconcile objects before a later phase fails;
// persisting their external IDs prevents duplicate remote objects on retry.
// Inputs: valid inventory and a fake exporter that stamps an external ID then
// returns an error. Outputs: export fails but the stamped ID remains on disk.
// Data choice: a random provider UUID proves the best-effort save preserved the
// mutation rather than merely rewriting identical inventory.
func TestRunExportPersistsExternalIDAfterProviderError(t *testing.T) {
	inventory, deviceID := connectedExportInventory()
	command, inventoryPath := prepareExportTest(t, inventory)
	externalID := uuid.New()
	provider := &fakeExportProvider{
		exportErr: errors.New("provider failed after reconciliation"),
		mutate: func(inventory *devicetypes.Inventory) {
			inventory.Devices[deviceID].ExternalIDs = map[string]uuid.UUID{"fake-export": externalID}
		},
	}

	if err := runExport(command, nil, provider); err == nil {
		t.Fatal("expected provider error to be returned")
	}
	reloaded, err := (&datastores.JSONStore{Path: inventoryPath}).Load()
	if err != nil {
		t.Fatalf("reloading inventory after provider error: %v", err)
	}
	if got := reloaded.Devices[deviceID].ExternalIDs["fake-export"]; got != externalID {
		t.Errorf("persisted external ID = %s, want %s", got, externalID)
	}
}
