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
	"maps"
	"reflect"
	"slices"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// Provider-metadata keys the v1alpha9 exporter reserved on a device: it
// mapped them to the device's own Nautobot fields and never sent them as
// custom fields.
const (
	legacyRoleKey         = "role"
	legacyLocationKey     = "location"
	legacyUHeightKey      = "u_height"
	legacyRackPositionKey = "rack_position"
)

// device migrates one device. The keys the exporter reserved for the
// device's kind are handled first, and the rest become custom fields. A
// rack kept in the device collection reserved its geometry and every other
// device its role; both reserved a location.
//
// The location is placed after the unowned values are taken: a name that
// cannot be placed is put back in the nautobot bucket, which take empties.
func (f *metadataFold) device(id uuid.UUID, device *devicetypes.CaniDeviceType) {
	owner := ownerLabel("device", device.Name, id)
	provider := &device.ProviderMetadata
	f.dropDerivedLocation(device)
	location := f.reserved(owner, provider, legacyLocationKey)
	if devicetypes.ClassifyForNautobot(string(device.Type)) == devicetypes.CategoryRack {
		f.dropUnsent(owner, provider, legacyUHeightKey, legacyRackPositionKey)
	} else {
		f.migrateDeviceRole(owner, device, f.reserved(owner, provider, legacyRoleKey))
	}
	taken := f.take(owner, provider)
	f.migrateDeviceLocation(owner, device, location)
	f.store(metadataRef{owner: owner, contentType: devicetypes.ContentTypeDevice, custom: &device.CustomFields}, taken)
}

// dropDerivedLocation clears the location a parented device stored unless it
// is its rack's. Earlier generations re-derived it on every load, so any
// other stored value meant nothing, and from v1alpha10 it would be read as a
// location the device records for itself.
func (f *metadataFold) dropDerivedLocation(device *devicetypes.CaniDeviceType) {
	if device.Parent == uuid.Nil {
		return
	}
	if rack := f.inv.Racks[device.Parent]; rack == nil || rack.Location != device.Location {
		device.Location = uuid.Nil
	}
}

// reservedValue is what the v1alpha9 exporter read for a reserved key.
type reservedValue struct {
	value any
	// unowned is set when the value was read from outside the buckets or
	// from the nautobot bucket, the copies the migration removes.
	unowned bool
}

// reserved removes a reserved key from outside the buckets and from the
// nautobot bucket, and returns the value the exporter read for it. A removed
// copy that is not that value was never exported, and is reported.
func (f *metadataFold) reserved(owner string, provider *map[string]any, key string) reservedValue {
	read := legacyLookup(*provider, key)
	for _, removed := range removeReserved(*provider, key) {
		if !reflect.DeepEqual(removed, read.value) {
			f.note(owner, "provider metadata %q was dropped: another copy of that key took precedence", key)
		}
	}
	return read
}

// legacyLookup repeats the v1alpha9 exporter's search for a reserved key:
// the value outside a bucket first, then the first bucket, in name order,
// that holds the key, whichever provider's it is.
func legacyLookup(provider map[string]any, key string) reservedValue {
	if value, held := provider[key]; held && !isBucket(value) {
		return reservedValue{value: value, unowned: true}
	}
	for _, name := range slices.Sorted(maps.Keys(provider)) {
		bucket, _ := provider[name].(map[string]any)
		if value, held := bucket[key]; held {
			return reservedValue{value: value, unowned: name == legacyNautobotBucket}
		}
	}
	return reservedValue{}
}

// removeReserved deletes key from outside the buckets and from the nautobot
// bucket, and returns the copies it deleted.
func removeReserved(provider map[string]any, key string) []any {
	var removed []any
	if value, held := provider[key]; held && !isBucket(value) {
		removed = append(removed, value)
		delete(provider, key)
	}
	typed, _ := provider[legacyNautobotBucket].(map[string]any)
	if value, held := typed[key]; held {
		removed = append(removed, value)
		delete(typed, key)
	}
	return removed
}

// isBucket reports whether a provider-metadata value is a provider's bucket.
func isBucket(value any) bool {
	_, bucket := value.(map[string]any)
	return bucket
}

// dropUnsent removes the geometry keys of a rack kept as a device. The
// exporter reserved them, so they were never custom fields, yet it never
// sent them either: it read no rack position, and a height only from a Go
// int, which neither a number in a datastore nor text typed with --metadata
// is. The rack's own record holds its height.
func (f *metadataFold) dropUnsent(owner string, provider *map[string]any, keys ...string) {
	for _, key := range keys {
		if len(removeReserved(*provider, key)) > 0 {
			f.note(owner, "provider metadata %q was dropped: export never sent it", key)
		}
	}
}

// migrateDeviceRole fills an empty role field with the role the exporter
// read from provider metadata, so the field alone says what an export sends.
// A role read from another provider's bucket also stays in that bucket. A
// role field that is already set is kept, as it was on export, and the
// unowned copy it overrode is reported.
func (f *metadataFold) migrateDeviceRole(owner string, device *devicetypes.CaniDeviceType, read reservedValue) {
	role, isText := read.value.(string)
	if device.Role == "" {
		device.Role = role
	}
	if read.unowned && (!isText || (role != "" && role != device.Role)) {
		f.note(owner, "provider metadata %q was dropped: it was not the role export sent", legacyRoleKey)
	}
}

// migrateDeviceLocation places a device in the location the exporter read
// from provider metadata, ahead of the inventory's own placement. Only text
// was read, as a location's name. A name that is not that of exactly one
// location in the inventory cannot become a placement: when it was unowned
// it is kept in the nautobot bucket, which nothing exports from this
// generation on, and it is reported either way.
func (f *metadataFold) migrateDeviceLocation(owner string, device *devicetypes.CaniDeviceType, read reservedValue) {
	name, isText := read.value.(string)
	if name == "" {
		if read.unowned && !isText {
			f.note(owner, "provider metadata %q was dropped: it is not a location name, so export never read it", legacyLocationKey)
		}
		return
	}
	target, problem := f.locationNamed(name)
	if problem == "" {
		f.placeDevice(owner, device, target)
		return
	}
	if read.unowned {
		setBucketKey(&device.ProviderMetadata, legacyNautobotBucket, legacyLocationKey, name)
	}
	f.note(owner, "provider metadata %q %s; export no longer places the device by it", legacyLocationKey, problem)
}

// locationNamed returns the one location with the name, or why there is
// none.
func (f *metadataFold) locationNamed(name string) (uuid.UUID, string) {
	var named []uuid.UUID
	for id, location := range f.inv.Locations {
		if location != nil && location.Name == name {
			named = append(named, id)
		}
	}
	switch len(named) {
	case 1:
		return named[0], ""
	case 0:
		return uuid.Nil, "names no location in the inventory"
	default:
		return uuid.Nil, "names more than one location in the inventory"
	}
}

// placeDevice records a migrated location under the rules a device's
// location follows from this generation on.
//
// A device with no parent takes it, in place of a location it already
// records: export sent the provider-metadata one. A racked device takes it
// only when it holds the rack's, and a device installed in another device
// takes that device's; Nautobot refused the first and overwrote the second,
// so neither was a placement that held.
func (f *metadataFold) placeDevice(owner string, device *devicetypes.CaniDeviceType, target uuid.UUID) {
	rack := f.inv.Racks[device.Parent]
	switch {
	case device.Parent != uuid.Nil && rack == nil:
		f.note(owner, "provider metadata %q was not applied: a device installed in another device takes that device's location", legacyLocationKey)
		return
	case rack != nil && !f.inv.LocationHolds(target, rack.Location):
		f.note(owner, "provider metadata %q was not applied: it does not hold the location of the device's rack", legacyLocationKey)
		return
	case rack == nil && device.Location != uuid.Nil && device.Location != target:
		f.note(owner, "provider metadata %q replaced the location the device recorded, as it did on export", legacyLocationKey)
	}
	device.Location = target
}
