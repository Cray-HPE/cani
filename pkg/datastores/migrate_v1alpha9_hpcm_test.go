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
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// hpcmLegacyInventory is a v1alpha9 inventory holding what the HPCM import
// wrote: a rack-mounted server and a blade, each with an inventory item, and
// the admin location. Every HPCM value but the server's own is a custom
// field, as that import stored it.
type hpcmLegacyInventory struct {
	inv                              *devicetypes.Inventory
	server, blade, admin, disk, dimm uuid.UUID
}

func newHPCMLegacyInventory() hpcmLegacyInventory {
	f := hpcmLegacyInventory{
		inv:    devicetypes.NewInventory(),
		server: uuid.New(), blade: uuid.New(), admin: uuid.New(), disk: uuid.New(), dimm: uuid.New(),
	}
	f.inv.SchemaVersion = devicetypes.SchemaVersionV1Alpha9
	f.inv.Devices[f.server] = &devicetypes.CaniDeviceType{ID: f.server, Name: "node001", ObjectMeta: devicetypes.ObjectMeta{
		ProviderMetadata: map[string]any{"hpcm": map[string]any{"hpcm_uuid": "11111111"}},
	}}
	f.inv.Modules[f.blade] = &devicetypes.CaniModuleType{ID: f.blade, Name: "x9000c1s0b0n0", ObjectMeta: devicetypes.ObjectMeta{
		CustomFields: map[string]any{
			"aliases":   map[string]any{"host": "nid000001"},
			"geoloc":    "x9000c1s0b0n0",
			"hpcm_uuid": "bbbbbbbb",
			"location":  map[string]any{"rack": float64(9000), "chassis": float64(1)},
		},
	}}
	f.inv.Locations[f.admin] = &devicetypes.CaniLocationType{ID: f.admin, Name: "admin", LocationType: "site", ObjectMeta: devicetypes.ObjectMeta{
		CustomFields: map[string]any{"hpcm_uuid": "aaaaaaaa"},
	}}
	f.inv.Frus[f.disk] = &devicetypes.CaniFruType{ID: f.disk, Name: "node001-disk.disk0", Type: "disk", Device: f.server, Parent: f.server,
		ObjectMeta: devicetypes.ObjectMeta{CustomFields: map[string]any{"size": "960GB", "wwn": "5002538f"}}}
	f.inv.Frus[f.dimm] = &devicetypes.CaniFruType{ID: f.dimm, Name: "x9000c1s0b0n0-dimm.dimm0", Type: "dimm", Device: f.blade, Parent: f.blade,
		ObjectMeta: devicetypes.ObjectMeta{CustomFields: map[string]any{"size": "64GB"}}}
	return f
}

// hpcmBucketKeys returns the sorted keys of an object's HPCM bucket.
func hpcmBucketKeys(meta devicetypes.ObjectMeta) []string {
	bucket, _ := meta.ProviderMetadata["hpcm"].(map[string]any)
	return slices.Sorted(maps.Keys(bucket))
}

// TestMigrateV1Alpha9MovesHPCMCustomFieldsToItsBucket verifies that what the
// HPCM import stored as custom fields on a module, a location and the
// inventory items it built moves to each object's HPCM bucket, and that one
// line reports how many moved.
//
// Why it matters: the import now writes these values to the HPCM bucket and
// a later import refreshes only that copy; a custom field left behind would
// go stale and keep being exported as an attribute nobody defined.
// Inputs: a blade with the four fields the import wrote on a module, the
// admin location with its UUID, and an inventory item on a server and on
// the blade.
// Outputs: no custom fields on any of them, the same keys in each HPCM
// bucket, and one report line counting eight.
// Data choice: the item on the blade is told to be HPCM's only by the
// bucket the blade gains in the same pass, so the order of the passes shows.
func TestMigrateV1Alpha9MovesHPCMCustomFieldsToItsBucket(t *testing.T) {
	// Arrange.
	f := newHPCMLegacyInventory()

	// Act.
	report := foldProviderMetadata(f.inv)

	// Assert.
	got := map[string]devicetypes.ObjectMeta{
		"blade": f.inv.Modules[f.blade].ObjectMeta, "admin": f.inv.Locations[f.admin].ObjectMeta,
		"disk": f.inv.Frus[f.disk].ObjectMeta, "dimm": f.inv.Frus[f.dimm].ObjectMeta,
	}
	want := map[string][]string{
		"blade": {"aliases", "geoloc", "hpcm_uuid", "location"}, "admin": {"hpcm_uuid"},
		"disk": {"size", "wwn"}, "dimm": {"size"},
	}
	for object, meta := range got {
		if meta.CustomFields != nil {
			t.Errorf("%s custom fields = %v, want none", object, meta.CustomFields)
		}
		if keys := hpcmBucketKeys(meta); !reflect.DeepEqual(keys, want[object]) {
			t.Errorf("%s HPCM bucket holds %v, want %v", object, keys, want[object])
		}
	}
	if location := got["blade"].ProviderMetadata["hpcm"].(map[string]any)["location"]; !reflect.DeepEqual(location, map[string]any{"rack": float64(9000), "chassis": float64(1)}) {
		t.Errorf("blade location = %v, want the value moved as it was", location)
	}
	assertReported(t, report, "8 custom field(s) the HPCM import wrote on modules, locations and inventory items moved to providerMetadata.hpcm")
}

// TestMigrateV1Alpha9LeavesCustomFieldsItCannotTellAreHPCMs verifies that a
// custom field stays where it is unless it can be told to be HPCM's.
//
// Why it matters: releases through v0.8.4 imported custom-field values from
// Nautobot without their definitions, so "no definition" does not make a
// custom field HPCM's. One moved by mistake would stop being exported.
// Inputs: one object per case, each holding a custom field that resembles
// an HPCM one but fails a single test.
// Outputs: every custom field still in place, no HPCM bucket gained, and
// nothing reported.
// Data choice: each case fails exactly one test: the shape of the value,
// the key, a definition of the key, a bucket that already holds the key,
// the item's name, the item's type, and the item's owner.
func TestMigrateV1Alpha9LeavesCustomFieldsItCannotTellAreHPCMs(t *testing.T) {
	hpcmServer := func(inv *devicetypes.Inventory) uuid.UUID {
		id := uuid.New()
		inv.Devices[id] = &devicetypes.CaniDeviceType{ID: id, Name: "node001", ObjectMeta: devicetypes.ObjectMeta{
			ProviderMetadata: map[string]any{"hpcm": map[string]any{"hpcm_uuid": "11111111"}},
		}}
		return id
	}
	item := func(inv *devicetypes.Inventory, name, kind string, owner uuid.UUID) *devicetypes.ObjectMeta {
		id := uuid.New()
		inv.Frus[id] = &devicetypes.CaniFruType{ID: id, Name: name, Type: devicetypes.Type(kind), Device: owner, Parent: owner,
			ObjectMeta: devicetypes.ObjectMeta{CustomFields: map[string]any{"size": "960GB"}}}
		return &inv.Frus[id].ObjectMeta
	}
	module := func(inv *devicetypes.Inventory, custom map[string]any) *devicetypes.ObjectMeta {
		id := uuid.New()
		inv.Modules[id] = &devicetypes.CaniModuleType{ID: id, Name: "nic-1", ObjectMeta: devicetypes.ObjectMeta{CustomFields: custom}}
		return &inv.Modules[id].ObjectMeta
	}
	tests := []struct {
		name  string
		build func(*devicetypes.Inventory) *devicetypes.ObjectMeta
	}{
		{"module location that is text, as Nautobot holds one", func(inv *devicetypes.Inventory) *devicetypes.ObjectMeta {
			return module(inv, map[string]any{"location": "aisle-4"})
		}},
		{"module geoloc that is a number", func(inv *devicetypes.Inventory) *devicetypes.ObjectMeta {
			return module(inv, map[string]any{"geoloc": float64(7)})
		}},
		{"module key the import never wrote", func(inv *devicetypes.Inventory) *devicetypes.ObjectMeta {
			return module(inv, map[string]any{"warranty": "2030-01-01"})
		}},
		{"module key the inventory defines", func(inv *devicetypes.Inventory) *devicetypes.ObjectMeta {
			if err := inv.AddCustomField(devicetypes.CustomFieldDefinition{Key: "geoloc", Type: "text", ContentTypes: []string{devicetypes.ContentTypeModule}}); err != nil {
				t.Fatal(err)
			}
			return module(inv, map[string]any{"geoloc": "x9000c1s0b0n0"})
		}},
		{"location key the import never wrote", func(inv *devicetypes.Inventory) *devicetypes.ObjectMeta {
			id := uuid.New()
			inv.Locations[id] = &devicetypes.CaniLocationType{ID: id, Name: "dc1", LocationType: "site",
				ObjectMeta: devicetypes.ObjectMeta{CustomFields: map[string]any{"ntp_server": "10.0.0.1"}}}
			return &inv.Locations[id].ObjectMeta
		}},
		{"item not named after the server it is on", func(inv *devicetypes.Inventory) *devicetypes.ObjectMeta {
			return item(inv, "disk.disk0", "disk", hpcmServer(inv))
		}},
		{"item whose type is not the one its name gives", func(inv *devicetypes.Inventory) *devicetypes.ObjectMeta {
			return item(inv, "node001-disk.disk0", "nic", hpcmServer(inv))
		}},
		{"item on a device HPCM did not import", func(inv *devicetypes.Inventory) *devicetypes.ObjectMeta {
			id := uuid.New()
			inv.Devices[id] = &devicetypes.CaniDeviceType{ID: id, Name: "node001"}
			return item(inv, "node001-disk.disk0", "disk", id)
		}},
		{"item key the inventory defines", func(inv *devicetypes.Inventory) *devicetypes.ObjectMeta {
			if err := inv.AddCustomField(devicetypes.CustomFieldDefinition{Key: "size", Type: "text", ContentTypes: []string{devicetypes.ContentTypeInventoryItem}}); err != nil {
				t.Fatal(err)
			}
			return item(inv, "node001-disk.disk0", "disk", hpcmServer(inv))
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			inv := devicetypes.NewInventory()
			inv.SchemaVersion = devicetypes.SchemaVersionV1Alpha9
			meta := tc.build(inv)
			before := maps.Clone(meta.CustomFields)

			// Act.
			report := foldProviderMetadata(inv)

			// Assert.
			if !reflect.DeepEqual(meta.CustomFields, before) {
				t.Errorf("custom fields = %v, want them left as %v", meta.CustomFields, before)
			}
			if _, gained := meta.ProviderMetadata["hpcm"]; gained {
				t.Errorf("provider metadata = %v, want no HPCM bucket", meta.ProviderMetadata)
			}
			assertReported(t, report)
		})
	}
}

// TestMigrateV1Alpha9KeepsAnHPCMValueTheBucketAlreadyHolds verifies a custom
// field is left alone when the HPCM bucket already holds its key.
//
// Why it matters: the bucket is where the import writes now; overwriting it
// with an older custom field would undo a value the import just refreshed.
// Inputs: a module whose HPCM bucket and custom fields both hold hpcm_uuid,
// with different values. Outputs: both unchanged, nothing reported.
// Data choice: the two values differ, so a move in either direction shows.
func TestMigrateV1Alpha9KeepsAnHPCMValueTheBucketAlreadyHolds(t *testing.T) {
	// Arrange.
	inv := devicetypes.NewInventory()
	inv.SchemaVersion = devicetypes.SchemaVersionV1Alpha9
	id := uuid.New()
	inv.Modules[id] = &devicetypes.CaniModuleType{ID: id, Name: "x9000c1s0b0n0", ObjectMeta: devicetypes.ObjectMeta{
		CustomFields:     map[string]any{"hpcm_uuid": "older"},
		ProviderMetadata: map[string]any{"hpcm": map[string]any{"hpcm_uuid": "newer"}},
	}}

	// Act.
	report := foldProviderMetadata(inv)

	// Assert.
	meta := inv.Modules[id].ObjectMeta
	if meta.CustomFields["hpcm_uuid"] != "older" || meta.ProviderMetadata["hpcm"].(map[string]any)["hpcm_uuid"] != "newer" {
		t.Errorf("custom fields = %v, provider metadata = %v; want both as they were", meta.CustomFields, meta.ProviderMetadata)
	}
	assertReported(t, report)
}

// TestMigrateV1Alpha9MovesHPCMFieldsBeforeDefiningMigratedOnes verifies that
// HPCM's records leave the custom fields before the migration defines the
// fields it created, so a value HPCM's step claims is never given a
// definition that would keep it from moving.
//
// Why it matters: the HPCM step tells its records by their having no
// definition. Run after the definitions are made, it would leave a moved
// value that is HPCM's in the custom fields, defined and exported, where
// the import refreshes only the bucket.
// Inputs: a module holding hpcm_uuid as a custom field, as the import wrote
// it, and in its nautobot bucket geoloc, which HPCM's step claims on a
// module, and support_tier, which it does not.
// Outputs: hpcm_uuid and geoloc in the HPCM bucket with no definition;
// support_tier as a defined custom field.
// Data choice: geoloc is both moved by the fold and claimed by HPCM, so the
// order of the two steps alone decides where it ends up.
func TestMigrateV1Alpha9MovesHPCMFieldsBeforeDefiningMigratedOnes(t *testing.T) {
	// Arrange.
	inv := devicetypes.NewInventory()
	inv.SchemaVersion = devicetypes.SchemaVersionV1Alpha9
	id := uuid.New()
	inv.Modules[id] = &devicetypes.CaniModuleType{ID: id, Name: "x9000c1s0b0n0", ObjectMeta: devicetypes.ObjectMeta{
		CustomFields:     map[string]any{"hpcm_uuid": "bbbbbbbb"},
		ProviderMetadata: map[string]any{"nautobot": map[string]any{"geoloc": "x9000c1s0b0n0", "support_tier": "gold"}},
	}}

	// Act.
	foldProviderMetadata(inv)

	// Assert.
	meta := inv.Modules[id].ObjectMeta
	wantCustom := map[string]any{"support_tier": "gold"}
	wantBucket := map[string]any{"hpcm_uuid": "bbbbbbbb", "geoloc": "x9000c1s0b0n0"}
	if !reflect.DeepEqual(meta.CustomFields, wantCustom) || !reflect.DeepEqual(meta.ProviderMetadata["hpcm"], wantBucket) {
		t.Errorf("custom fields = %v, provider metadata = %v; want %v and an HPCM bucket %v", meta.CustomFields, meta.ProviderMetadata, wantCustom, wantBucket)
	}
	defined := make(map[string]bool)
	for _, def := range inv.ListCustomFields() {
		defined[def.Key] = true
	}
	if defined["hpcm_uuid"] || defined["geoloc"] || !defined["support_tier"] {
		t.Errorf("defined = %v, want support_tier defined and HPCM's keys not", defined)
	}
}

// TestJSONStoreMovesHPCMCustomFieldsOfAV1Alpha9Datastore verifies a load of
// a datastore the v1alpha9 HPCM import wrote leaves every HPCM value in the
// HPCM bucket and none as a custom field, and says how many moved.
//
// Why it matters: the rules above are written from reading the old import;
// this file is what that import actually produced, so it holds them to the
// stored shapes: numbers as float64, the rack number outside a bucket, and
// an inventory item whose parent is a module.
// Inputs: testdata/fixtures/cani/legacy/canitestdb_v1alpha9_hpcm.json, the
// datastore the build of main wrote from testdata/fixtures/hpcm/
// blade_system.json.
// Outputs: v1alpha10; on all ten objects no custom fields and provider
// metadata that is the HPCM bucket alone; the blade's four keys; a log line
// counting nine; and no rewrite on a second load.
// Data choice: the file is kept as written rather than built in the test,
// because the point is that the test does not share the migration's reading
// of the old format.
func TestJSONStoreMovesHPCMCustomFieldsOfAV1Alpha9Datastore(t *testing.T) {
	// Arrange.
	original, err := os.ReadFile("../../testdata/fixtures/cani/legacy/canitestdb_v1alpha9_hpcm.json")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "inventory.json")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	var logged bytes.Buffer
	writer := log.Writer()
	log.SetOutput(&logged)
	t.Cleanup(func() { log.SetOutput(writer) })
	store := &JSONStore{Path: path}

	// Act.
	loaded, err := store.Load()

	// Assert.
	if err != nil {
		t.Fatal(err)
	}
	if loaded.SchemaVersion != devicetypes.CurrentSchemaVersion {
		t.Errorf("version = %q, want %q", loaded.SchemaVersion, devicetypes.CurrentSchemaVersion)
	}
	metas := map[string]devicetypes.ObjectMeta{}
	for _, o := range loaded.Locations {
		metas["location "+o.Name] = o.ObjectMeta
	}
	for _, o := range loaded.Racks {
		metas["rack "+o.Name] = o.ObjectMeta
	}
	for _, o := range loaded.Devices {
		metas["device "+o.Name] = o.ObjectMeta
	}
	for _, o := range loaded.Modules {
		metas["module "+o.Name] = o.ObjectMeta
	}
	for _, o := range loaded.Frus {
		metas["inventory item "+o.Name] = o.ObjectMeta
	}
	if len(metas) != 10 {
		t.Fatalf("objects = %v, want ten", slices.Sorted(maps.Keys(metas)))
	}
	for object, meta := range metas {
		if len(meta.CustomFields) != 0 {
			t.Errorf("%s custom fields = %v, want none", object, meta.CustomFields)
		}
		if names := slices.Sorted(maps.Keys(meta.ProviderMetadata)); !reflect.DeepEqual(names, []string{"hpcm"}) {
			t.Errorf("%s provider metadata is held under %v, want the HPCM bucket alone", object, names)
		}
	}
	if keys, want := hpcmBucketKeys(metas["module x9000c1s0b0n0"]), []string{"aliases", "geoloc", "hpcm_uuid", "location"}; !reflect.DeepEqual(keys, want) {
		t.Errorf("blade HPCM bucket holds %v, want %v", keys, want)
	}
	if keys, want := hpcmBucketKeys(metas["inventory item x9000c1s0b0n0-dimm.dimm0"]), []string{"size"}; !reflect.DeepEqual(keys, want) {
		t.Errorf("blade DIMM HPCM bucket holds %v, want %v", keys, want)
	}
	if wantLine := "Migration to v1alpha10: 9 custom field(s) the HPCM import wrote"; !strings.Contains(logged.String(), wantLine) {
		t.Errorf("log = %q, want a line containing %q", logged.String(), wantLine)
	}
	saved := readSchemaMigrationFile(t, path)
	if _, err := store.Load(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(readSchemaMigrationFile(t, path), saved) {
		t.Error("repeated load changed the migrated datastore")
	}
}
