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
package transform

import (
	"maps"
	"reflect"
	"slices"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/hpcm/import"
)

// hpcmSystem is a small system as HPCM reports it: an admin node, a chassis
// with one blade whose location reaches a node slot, and a rack-mounted
// server, with the aliases and inventory HPCM attaches to them.
func hpcmSystem() []import_.Node {
	return []import_.Node{
		{Name: "admin", UUID: "aaaaaaaa-0000-4000-8000-000000000001", Type: "admin"},
		{
			Name: "x9000c1", UUID: "cccccccc-0000-4000-8000-000000000001", Type: "chassis",
			Location: &import_.LocationSettings{Rack: import_.Int32Ptr(9000), Chassis: import_.Int32Ptr(1)},
		},
		{
			Name: "x9000c1s0b0n0", UUID: "bbbbbbbb-0000-4000-8000-000000000001", Type: "compute",
			Aliases: map[string]string{"cm-geo-name": "x9000c1s0b0n0", "host": "nid000001"},
			Location: &import_.LocationSettings{
				Rack: import_.Int32Ptr(9000), Chassis: import_.Int32Ptr(1), Tray: import_.Int32Ptr(0), Node: import_.Int32Ptr(0),
			},
			Inventory: map[string]string{
				"geoloc": "x9000c1s0b0n0", "dimm.dimm0.serial": "DIMM-S1", "dimm.dimm0.size": "64GB",
			},
		},
		{
			Name: "node001", UUID: "11111111-1111-4111-8111-111111111111", Type: "compute",
			Aliases:   map[string]string{"host": "login1"},
			Location:  &import_.LocationSettings{Rack: import_.Int32Ptr(1)},
			Inventory: map[string]string{"disk.disk0.model": "MZ7L3960", "disk.disk0.size": "960GB"},
		},
	}
}

// hpcmBucketOf returns the HPCM bucket of an object and fails the test when
// the object has custom fields or provider metadata outside that bucket.
func hpcmBucketOf(t *testing.T, object string, meta devicetypes.ObjectMeta) map[string]any {
	t.Helper()
	if len(meta.CustomFields) != 0 {
		t.Errorf("%s: custom fields = %v, want none", object, meta.CustomFields)
	}
	if names := slices.Sorted(maps.Keys(meta.ProviderMetadata)); len(names) > 1 || (len(names) == 1 && names[0] != hpcmBucket) {
		t.Errorf("%s: provider metadata is held under %v, want the hpcm bucket only", object, names)
	}
	bucket, _ := meta.ProviderMetadata[hpcmBucket].(map[string]any)
	return bucket
}

// TestTransformNodes_KeepsProviderDataInTheHPCMBucket verifies that every
// object the transform builds has no custom fields and holds what HPCM
// records about it in providerMetadata.hpcm and nowhere else.
//
// Why it matters: a custom field is an attribute an operator defines. The
// import used to write a module's HPCM UUID, aliases, location and geoloc, a
// location's UUID and every inventory field it had no place for as custom
// fields, and a rack's number outside any bucket, so another provider's
// export sent them as custom fields nobody had defined.
// Inputs: an admin node, a chassis, a blade in it and a rack-mounted
// server, with aliases, locations and inventory.
// Outputs: no custom fields and nothing outside the hpcm bucket on any
// object; the bucket of each kind holds the keys HPCM records for it; and
// a modeled inventory field is still a field.
// Data choice: the blade carries all four module keys, and each inventory
// group pairs a field the inventory models (serial, model) with one it
// does not (size), so both destinations are seen.
func TestTransformNodes_KeepsProviderDataInTheHPCMBucket(t *testing.T) {
	// Act.
	result, err := transformNodes(hpcmSystem(), devicetypes.NewInventory())

	// Assert.
	if err != nil {
		t.Fatalf("transformNodes: %v", err)
	}
	buckets := map[string]map[string]any{}
	for _, o := range result.Locations {
		buckets["location "+o.Name] = hpcmBucketOf(t, "location "+o.Name, o.ObjectMeta)
	}
	for _, o := range result.Racks {
		buckets["rack "+o.Name] = hpcmBucketOf(t, "rack "+o.Name, o.ObjectMeta)
	}
	for _, o := range result.Devices {
		buckets["device "+o.Name] = hpcmBucketOf(t, "device "+o.Name, o.ObjectMeta)
	}
	fruFields := map[string]string{}
	for _, o := range result.Frus {
		buckets["fru "+o.Name] = hpcmBucketOf(t, "fru "+o.Name, o.ObjectMeta)
		fruFields[o.Name] = o.Serial + o.Model
	}
	var blade *devicetypes.CaniModuleType
	for _, o := range result.Modules {
		buckets["module "+o.Name] = hpcmBucketOf(t, "module "+o.Name, o.ObjectMeta)
		blade = o
	}

	wantKeys := map[string][]string{
		"location admin":               {"hpcm_uuid"},
		"rack rack-1":                  {"rack_number"},
		"rack rack-9000":               {"rack_number"},
		"device x9000c1":               {"hpcm_uuid", "import_source", "location"},
		"device node001":               {"aliases", "hpcm_uuid", "import_source", "location"},
		"module x9000c1s0b0n0":         {"aliases", "geoloc", "hpcm_uuid", "location"},
		"fru x9000c1s0b0n0-dimm.dimm0": {"size"},
		"fru node001-disk.disk0":       {"size"},
	}
	for object, want := range wantKeys {
		if got := slices.Sorted(maps.Keys(buckets[object])); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: hpcm bucket holds %v, want %v", object, got, want)
		}
	}
	if len(buckets) != len(wantKeys) {
		t.Errorf("objects built = %v, want exactly %d", slices.Sorted(maps.Keys(buckets)), len(wantKeys))
	}
	if got := buckets["rack rack-9000"]["rack_number"]; got != int32(9000) {
		t.Errorf("rack-9000 rack_number = %v (%T), want 9000", got, got)
	}
	if got := buckets["module x9000c1s0b0n0"]["geoloc"]; got != "x9000c1s0b0n0" {
		t.Errorf("blade geoloc = %v, want x9000c1s0b0n0", got)
	}
	if fruFields["x9000c1s0b0n0-dimm.dimm0"] != "DIMM-S1" || fruFields["node001-disk.disk0"] != "MZ7L3960" {
		t.Errorf("modeled inventory fields = %v, want the DIMM's serial and the disk's model as fields", fruFields)
	}
	if blade == nil || result.Devices[blade.ParentDevice] == nil || result.Devices[blade.ParentDevice].Name != "x9000c1" {
		t.Errorf("blade = %+v, want it parented to chassis x9000c1", blade)
	}
}
