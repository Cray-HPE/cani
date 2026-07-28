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
package store

import (
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

type recordingStore struct {
	saveCalls int
}

func (store *recordingStore) Load() (*devicetypes.Inventory, error) {
	return devicetypes.NewInventory(), nil
}

func (store *recordingStore) Save(_ *devicetypes.Inventory) error {
	store.saveCalls++
	return nil
}

func useRecordingStore(t *testing.T) (*recordingStore, validatingStore) {
	t.Helper()
	original := datastores.Datastore
	recorder := &recordingStore{}
	store := validatingStore{DeviceStore: recorder}
	datastores.Datastore = store
	t.Cleanup(func() { datastores.Datastore = original })
	return recorder, store
}

// TestValidatingStoreRejectsInvalidRelationships verifies validation failure
// prevents the configured datastore from receiving a save call.
//
// Why it matters: every CLI mutation uses this boundary, so invalid cable
// endpoints must be stopped before durable inventory is changed.
// Inputs: firewall-01 with port1 and a cable naming missing port3.
// Outputs: a contextual relationship error and zero datastore saves.
// Data choice: a valid device with an invalid port isolates cable resolution
// from datastore behavior and matches the reported regression.
func TestValidatingStoreRejectsInvalidRelationships(t *testing.T) {
	recorder, store := useRecordingStore(t)
	deviceID := uuid.New()
	inv := devicetypes.NewInventory()
	inv.Devices[deviceID] = &devicetypes.CaniDeviceType{
		ID: deviceID, Name: "firewall-01",
		Interfaces: []devicetypes.InterfaceSpec{{ID: uuid.New(), Name: "port1"}},
	}
	inv.Cables[uuid.New()] = &devicetypes.CaniCableType{
		Label:              "invalid-uplink",
		TerminationBDevice: deviceID,
		TerminationBPort:   "port3",
	}

	err := store.Save(inv)

	if err == nil {
		t.Fatal("expected invalid relationships to prevent save")
	}
	if !strings.Contains(err.Error(), `termination B port "port3" not found on device "firewall-01"`) {
		t.Errorf("validatingStore.Save() error = %q, want endpoint context", err)
	}
	if recorder.saveCalls != 0 {
		t.Errorf("datastore Save calls = %d, want 0", recorder.saveCalls)
	}
}

// TestValidatingStoreResolvesAndPersistsValidRelationships verifies a valid
// endpoint is rebuilt before exactly one datastore save.
//
// Why it matters: validation must retain the existing derivation behavior and
// must not block correctly connected inventory from being persisted.
// Inputs: switch-01 with port1 and a cable naming that device and port.
// Outputs: nil error, one save call, and the resolved interface UUID.
// Data choice: a zero termination UUID proves validatingStore runs rebuild before
// deciding whether the inventory is valid.
func TestValidatingStoreResolvesAndPersistsValidRelationships(t *testing.T) {
	recorder, store := useRecordingStore(t)
	deviceID := uuid.New()
	interfaceID := uuid.New()
	cableID := uuid.New()
	inv := devicetypes.NewInventory()
	inv.Devices[deviceID] = &devicetypes.CaniDeviceType{
		ID: deviceID, Name: "switch-01",
		Interfaces: []devicetypes.InterfaceSpec{{ID: interfaceID, Name: "port1"}},
	}
	inv.Cables[cableID] = &devicetypes.CaniCableType{
		ID:                 cableID,
		Label:              "uplink",
		TerminationADevice: deviceID,
		TerminationAPort:   "port1",
	}

	if err := store.Save(inv); err != nil {
		t.Fatalf("validatingStore.Save() unexpected error: %v", err)
	}
	if recorder.saveCalls != 1 {
		t.Errorf("datastore Save calls = %d, want 1", recorder.saveCalls)
	}
	if inv.Cables[cableID].TerminationA != interfaceID {
		t.Errorf("TerminationA = %s, want %s", inv.Cables[cableID].TerminationA, interfaceID)
	}
}

// TestSetupInstallsSingleValidatingStore verifies Setup decorates the selected
// backend exactly once on every invocation.
//
// Why it matters: command call sites keep using DeviceStore.Save, so Setup must
// establish the validation boundary without accumulating nested decorators.
// Inputs: a root command selecting the JSON backend and two Setup calls.
// Outputs: each call installs one validatingStore around a raw JSONStore.
// Data choice: repeated setup matches separate command executions in tests and
// proves backend selection resets the wrapper before decoration.
func TestSetupInstallsSingleValidatingStore(t *testing.T) {
	original := datastores.Datastore
	originalConfig := config.Cfg
	config.Cfg = &config.Config{Path: t.TempDir() + "/config.yaml", Datastore: "inventory.json"}
	t.Cleanup(func() {
		datastores.Datastore = original
		config.Cfg = originalConfig
	})

	root := &cli.Command{Use: "cani"}
	root.PersistentFlags().String("datastore", "json", "")
	command := &cli.Command{Use: "test"}
	root.AddCommand(command)

	for range 2 {
		if err := Setup(command); err != nil {
			t.Fatalf("Setup() unexpected error: %v", err)
		}
		wrapped, ok := datastores.Datastore.(validatingStore)
		if !ok {
			t.Fatalf("Datastore = %T, want validatingStore", datastores.Datastore)
		}
		if _, nested := wrapped.DeviceStore.(validatingStore); nested {
			t.Fatal("Setup() nested validatingStore decorators")
		}
		if _, ok := wrapped.DeviceStore.(*datastores.JSONStore); !ok {
			t.Errorf("wrapped datastore = %T, want *datastores.JSONStore", wrapped.DeviceStore)
		}
	}
}
