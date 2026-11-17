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
	"reflect"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// v1Alpha9ModuleInventory returns the v1alpha8 fixture stamped v1alpha9. It
// holds provider metadata in a CSM bucket only, so nothing in it is unowned.
func v1Alpha9ModuleInventory(t *testing.T) *devicetypes.Inventory {
	t.Helper()
	inventory := v1Alpha8ModuleInventory(t)
	inventory.SchemaVersion = devicetypes.SchemaVersionV1Alpha9
	return inventory
}

// TestJSONStoreMigratesV1Alpha9WithoutDataLoss verifies a v1alpha9 datastore
// with nothing to move is upgraded on load, backed up, and left stable on
// the next load.
//
// Why it matters: most datastores hold provider metadata only in a
// provider's own bucket; for them the version boundary must change nothing.
// Inputs: a v1alpha9 datastore whose only provider metadata is a CSM bucket,
// with racked devices, an installed module and a spare. Outputs: a file
// differing only in its version, a byte-identical .canisave backup, and no
// further rewrite on a repeated load.
// Data choice: the racked devices store their rack's location, which is the
// one stored location on a parented device the migration must keep.
func TestJSONStoreMigratesV1Alpha9WithoutDataLoss(t *testing.T) {
	assertStoreUpgrade(t, v1Alpha9ModuleInventory(t), devicetypes.SchemaVersionV1Alpha9)
}

// TestJSONStoreMovesV1Alpha9MetadataOnLoad verifies a load applies the
// ownership rules to a v1alpha9 file, writes the result once, keeps the
// original, and says what it dropped.
//
// Why it matters: the migration is only reached through Load, and its
// report is the operator's one notice that a typed value was not carried.
// Inputs: a v1alpha9 file whose node has a role field, a different typed
// role and a typed pair. Outputs: the loaded and the saved inventory hold
// the pair as a custom field and no nautobot bucket; the backup is the
// original file; a second load rewrites nothing; the log names the device
// and the key but not the dropped role.
// Data choice: the conflicting role exercises a reported drop, and its
// value is distinctive so a leak into the log would be found.
func TestJSONStoreMovesV1Alpha9MetadataOnLoad(t *testing.T) {
	inventory, ids := newCRUDInventory(t)
	inventory.SchemaVersion = devicetypes.SchemaVersionV1Alpha9
	node := inventory.Devices[ids.nodeID]
	node.Role = "Storage"
	node.ProviderMetadata["nautobot"] = map[string]any{"role": "typed-role-value", "asset_owner": "hpc-ops"}
	original := schemaMigrationJSON(t, inventory)
	path := filepath.Join(t.TempDir(), "inventory.json")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	var logged bytes.Buffer
	writer := log.Writer()
	log.SetOutput(&logged)
	t.Cleanup(func() { log.SetOutput(writer) })
	store := &JSONStore{Path: path}

	loaded, err := store.Load()

	if err != nil {
		t.Fatal(err)
	}
	saved := readSchemaMigrationFile(t, path)
	var persisted devicetypes.Inventory
	if err := json.Unmarshal(saved, &persisted); err != nil {
		t.Fatal(err)
	}
	if persisted.SchemaVersion != devicetypes.CurrentSchemaVersion {
		t.Errorf("saved version = %q, want %q", persisted.SchemaVersion, devicetypes.CurrentSchemaVersion)
	}
	for source, device := range map[string]*devicetypes.CaniDeviceType{
		"loaded": loaded.Devices[ids.nodeID], "saved": persisted.Devices[ids.nodeID],
	} {
		if device.Role != "Storage" {
			t.Errorf("%s role = %q, want the role field kept", source, device.Role)
		}
		if wantCustom := map[string]any{"asset_owner": "hpc-ops"}; !reflect.DeepEqual(device.CustomFields, wantCustom) {
			t.Errorf("%s custom fields = %v, want %v", source, device.CustomFields, wantCustom)
		}
		if _, held := device.ProviderMetadata["nautobot"]; held || device.ProviderMetadata["csm"] == nil {
			t.Errorf("%s provider metadata = %v, want the CSM bucket alone", source, device.ProviderMetadata)
		}
	}
	if !bytes.Equal(readSchemaMigrationFile(t, path+".canisave"), original) {
		t.Error("migration backup does not preserve the original v1alpha9 data")
	}
	wantLine := `Migration to v1alpha10: device "node-a" (` + ids.nodeID.String() + `): provider metadata "role" was dropped`
	if !strings.Contains(logged.String(), wantLine) {
		t.Errorf("log = %q, want a line containing %q", logged.String(), wantLine)
	}
	if strings.Contains(logged.String(), "typed-role-value") {
		t.Errorf("log prints the dropped value: %q", logged.String())
	}
	if _, err := store.Load(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(readSchemaMigrationFile(t, path), saved) {
		t.Error("repeated load changed the migrated datastore")
	}
}

// TestJSONStoreRejectsV1Alpha9Writes verifies an unmigrated v1alpha9
// inventory cannot be saved.
//
// Why it matters: only the current generation may be written, so a caller
// that skipped migration cannot stamp unowned provider metadata with a
// version that promises there is none.
// Inputs: a v1alpha9 inventory and an empty path. Outputs: the
// unsupported-version sentinel and no file on disk.
// Data choice: the fixture is otherwise valid, so only the version can cause
// the rejection.
func TestJSONStoreRejectsV1Alpha9Writes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inventory.json")
	store := &JSONStore{Path: path}

	if err := store.Save(v1Alpha9ModuleInventory(t)); !errors.Is(err, ErrUnsupportedSchemaVersion) {
		t.Fatalf("Save() error=%v, want ErrUnsupportedSchemaVersion for an unmigrated inventory", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("rejected write created a datastore: %v", err)
	}
}
