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
package datastores

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func v1Alpha7BayInventory(t *testing.T) *devicetypes.Inventory {
	t.Helper()
	inventory, ids := newCRUDInventory(t)
	inventory.SchemaVersion = devicetypes.SchemaVersionV1Alpha7
	parent := inventory.Devices[ids.nodeID]
	parent.DeviceBays = []devicetypes.DeviceBaySpec{{Name: "Bay 1"}}
	parent.ModuleBays = []devicetypes.ModuleBaySpec{{Name: "Slot 1", Position: "1"}}
	if err := inventory.InstallInBay(inventory.Devices[ids.switchID], parent.ID, "Bay 1", nil); err != nil {
		t.Fatal(err)
	}
	name, label := "custom-{module}/0", "port {module}"
	module := &devicetypes.CaniModuleType{
		ID: uuid.New(), Name: "nic", ParentDevice: parent.ID, ModuleBayName: "Slot 1",
		Interfaces: []devicetypes.InterfaceSpec{{
			ID: uuid.New(), Name: "custom-1/0", Label: "port 1", NameTemplate: &name, LabelTemplate: &label,
		}},
	}
	inventory.Modules[module.ID] = module
	if err := inventory.RebuildDerivedState().Err(); err != nil {
		t.Fatal(err)
	}
	return inventory
}

func TestMigrateV1Alpha7PreservesInventory(t *testing.T) {
	inventory := v1Alpha7BayInventory(t)
	before := schemaMigrationJSON(t, inventory)

	migrateV1Alpha7(inventory)

	if inventory.SchemaVersion != devicetypes.SchemaVersionV1Alpha8 {
		t.Fatalf("version=%q, want %q", inventory.SchemaVersion, devicetypes.SchemaVersionV1Alpha8)
	}
	assertOnlySchemaVersionChanged(t, before, schemaMigrationJSON(t, inventory),
		devicetypes.SchemaVersionV1Alpha7, devicetypes.SchemaVersionV1Alpha8)
}

func TestJSONStoreMigratesV1Alpha7WithoutDataLoss(t *testing.T) {
	assertV1Alpha7StoreUpgrade(t, v1Alpha7BayInventory(t))
}

func TestJSONStoreMigratesV1Alpha7WithoutInferringNewFields(t *testing.T) {
	inventory := v1Alpha7BayInventory(t)
	for _, device := range inventory.Devices {
		device.ParentBay = ""
	}
	for _, module := range inventory.Modules {
		for i := range module.Interfaces {
			module.Interfaces[i].NameTemplate = nil
			module.Interfaces[i].LabelTemplate = nil
		}
	}
	assertV1Alpha7StoreUpgrade(t, inventory)
}

func TestJSONStoreRejectsV1Alpha7Writes(t *testing.T) {
	inventory := v1Alpha7BayInventory(t)
	path := filepath.Join(t.TempDir(), "inventory.json")
	store := &JSONStore{Path: path}

	if err := store.Save(inventory); !errors.Is(err, ErrUnsupportedSchemaVersion) {
		t.Fatalf("Save() error=%v, want ErrUnsupportedSchemaVersion for an unmigrated inventory", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("rejected write created a datastore: %v", err)
	}
}

func assertV1Alpha7StoreUpgrade(t *testing.T, inventory *devicetypes.Inventory) {
	t.Helper()
	original := schemaMigrationJSON(t, inventory)
	path := filepath.Join(t.TempDir(), "inventory.json")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	store := &JSONStore{Path: path}

	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}

	saved := readSchemaMigrationFile(t, path)
	assertOnlySchemaVersionChanged(t, original, saved,
		devicetypes.SchemaVersionV1Alpha7, devicetypes.CurrentSchemaVersion)
	if !bytes.Equal(schemaMigrationJSON(t, loaded), saved) {
		t.Fatal("loaded inventory differs from the persisted migration")
	}
	if !bytes.Equal(readSchemaMigrationFile(t, path+".canisave"), original) {
		t.Fatal("migration backup does not preserve the original v1alpha7 data")
	}

	if _, err := store.Load(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(readSchemaMigrationFile(t, path), saved) {
		t.Fatal("repeated load changed the current-generation datastore")
	}
	if !bytes.Equal(readSchemaMigrationFile(t, path+".canisave"), original) {
		t.Fatal("repeated load replaced the original migration backup")
	}
}

func schemaMigrationJSON(t *testing.T, inventory *devicetypes.Inventory) []byte {
	t.Helper()
	data, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func readSchemaMigrationFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func assertOnlySchemaVersionChanged(t *testing.T, before, after []byte, from, to string) {
	t.Helper()
	oldVersion := []byte(fmt.Sprintf(`"schemaVersion": %q`, from))
	newVersion := []byte(fmt.Sprintf(`"schemaVersion": %q`, to))
	want := bytes.Replace(before, oldVersion, newVersion, 1)
	if bytes.Equal(want, before) {
		t.Fatalf("original inventory does not contain schema version %q", from)
	}
	if !bytes.Equal(after, want) {
		t.Fatalf("migration changed inventory data beyond the schema version:\ngot: %s\nwant: %s", after, want)
	}
}
