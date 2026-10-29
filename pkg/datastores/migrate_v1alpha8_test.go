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
	"os"
	"path/filepath"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// v1Alpha8ModuleInventory returns a v1alpha8 inventory that already carries
// the state v1alpha9 gives meaning to: a module-owned port and a spare module
// stored at a location with no parent device.
func v1Alpha8ModuleInventory(t *testing.T) *devicetypes.Inventory {
	t.Helper()
	inventory := v1Alpha7BayInventory(t)
	inventory.SchemaVersion = devicetypes.SchemaVersionV1Alpha8
	site := inventory.FindLocationByNameOrID("site-a")
	if site == nil {
		t.Fatal("fixture location site-a not found")
	}
	spare := &devicetypes.CaniModuleType{ID: uuid.New(), Name: "spare-nic", Location: site.ID}
	inventory.Modules[spare.ID] = spare
	if err := inventory.RebuildDerivedState().Err(); err != nil {
		t.Fatal(err)
	}
	return inventory
}

// installedModulePort returns the installed fixture module and the ID of the
// port its spec declares.
func installedModulePort(t *testing.T, inventory *devicetypes.Inventory) (*devicetypes.CaniModuleType, uuid.UUID) {
	t.Helper()
	for _, module := range inventory.Modules {
		if module.ParentDevice != uuid.Nil && len(module.Interfaces) == 1 {
			return module, module.Interfaces[0].ID
		}
	}
	t.Fatal("fixture has no installed module with one port")
	return nil, uuid.Nil
}

// withoutModuleOwnership deletes the moduleId key from every indexed
// interface in a marshaled inventory, reproducing a document written before
// v1alpha9 defined the field.
func withoutModuleOwnership(t *testing.T, data []byte) []byte {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	interfaces, _ := doc["interfaces"].(map[string]any)
	for _, iface := range interfaces {
		if fields, ok := iface.(map[string]any); ok {
			delete(fields, "moduleId")
		}
	}
	stripped, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(stripped, []byte(`"moduleId"`)) {
		t.Fatal("document still records module ownership")
	}
	return stripped
}

// TestMigrateV1Alpha8PreservesInventory verifies the v1alpha8 step changes
// nothing but the schema version.
//
// Why it matters: module ownership is rebuilt on load and spares use fields
// v1alpha8 already carries, so any other change would be an inferred value.
// Inputs: a v1alpha8 inventory with a module-owned port and a spare module.
// Outputs: the version is v1alpha9 and the marshaled document is otherwise
// byte-for-byte identical.
// Data choice: the spare and the module port are the state v1alpha9 adds
// meaning to, so a step that rewrote either would be visible.
func TestMigrateV1Alpha8PreservesInventory(t *testing.T) {
	inventory := v1Alpha8ModuleInventory(t)
	before := schemaMigrationJSON(t, inventory)

	migrateV1Alpha8(inventory)

	if inventory.SchemaVersion != devicetypes.SchemaVersionV1Alpha9 {
		t.Fatalf("version=%q, want %q", inventory.SchemaVersion, devicetypes.SchemaVersionV1Alpha9)
	}
	assertOnlySchemaVersionChanged(t, before, schemaMigrationJSON(t, inventory),
		devicetypes.SchemaVersionV1Alpha8, devicetypes.SchemaVersionV1Alpha9)
}

// TestJSONStoreMigratesV1Alpha8WithoutDataLoss verifies a v1alpha8 datastore
// is upgraded on load, backed up, and left stable on the next load.
//
// Why it matters: the version boundary must not cost an operator any data.
// Inputs: a v1alpha8 datastore with a spare and a bay-less installed module.
// Outputs: a v1alpha9 file differing only in its version, a byte-identical
// .canisave backup, and no further rewrite on a repeated load.
// Data choice: clearing the installed module's bay proves the step does not
// invent one from the module name, and the spare proves it is not given a
// parent device.
func TestJSONStoreMigratesV1Alpha8WithoutDataLoss(t *testing.T) {
	inventory := v1Alpha8ModuleInventory(t)
	module, _ := installedModulePort(t, inventory)
	module.ModuleBayName = ""

	assertStoreUpgrade(t, inventory, devicetypes.SchemaVersionV1Alpha8)
}

// TestJSONStoreRebuildsModuleOwnershipForV1Alpha8 verifies a v1alpha8
// datastore written without module ownership gains it from the module spec.
//
// Why it matters: a v1alpha8 writer had no moduleId field, so the upgrade must
// derive ownership rather than leave module ports looking device-owned.
// Inputs: a v1alpha8 datastore whose indexed interfaces carry no moduleId.
// Outputs: the saved v1alpha9 file records the declaring module on its port.
// Data choice: deleting the key from the marshaled document reproduces what an
// earlier cani release persisted; clearing the field would still serialize it.
func TestJSONStoreRebuildsModuleOwnershipForV1Alpha8(t *testing.T) {
	inventory := v1Alpha8ModuleInventory(t)
	module, portID := installedModulePort(t, inventory)
	path := filepath.Join(t.TempDir(), "inventory.json")
	if err := os.WriteFile(path, withoutModuleOwnership(t, schemaMigrationJSON(t, inventory)), 0600); err != nil {
		t.Fatal(err)
	}

	if _, err := (&JSONStore{Path: path}).Load(); err != nil {
		t.Fatal(err)
	}

	var saved devicetypes.Inventory
	if err := json.Unmarshal(readSchemaMigrationFile(t, path), &saved); err != nil {
		t.Fatal(err)
	}
	if saved.SchemaVersion != devicetypes.SchemaVersionV1Alpha9 {
		t.Fatalf("saved version=%q, want %q", saved.SchemaVersion, devicetypes.SchemaVersionV1Alpha9)
	}
	if port := saved.Interfaces[portID]; port == nil || port.ModuleID != module.ID {
		t.Fatalf("saved port %s = %+v, want ModuleID %s", portID, port, module.ID)
	}
}

// TestJSONStoreRejectsV1Alpha8Writes verifies an unmigrated v1alpha8
// inventory cannot be saved.
//
// Why it matters: only the current generation may be written, so a caller
// that skipped migration cannot stamp new data with the old version.
// Inputs: a v1alpha8 inventory and an empty path. Outputs: the
// unsupported-version sentinel and no file on disk.
// Data choice: the fixture is otherwise valid, so only the version can cause
// the rejection.
func TestJSONStoreRejectsV1Alpha8Writes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inventory.json")
	store := &JSONStore{Path: path}

	if err := store.Save(v1Alpha8ModuleInventory(t)); !errors.Is(err, ErrUnsupportedSchemaVersion) {
		t.Fatalf("Save() error=%v, want ErrUnsupportedSchemaVersion for an unmigrated inventory", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("rejected write created a datastore: %v", err)
	}
}
