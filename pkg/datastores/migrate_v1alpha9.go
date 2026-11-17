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
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// Provider-metadata names the v1alpha9 generation gave a meaning to.
const (
	// legacyNautobotBucket is where --metadata stored its pairs.
	legacyNautobotBucket = "nautobot"
	// hpcmBucket and hpcmRackNumberKey: the HPCM provider wrote a rack's
	// number outside any bucket.
	hpcmBucket        = "hpcm"
	hpcmRackNumberKey = "rack_number"
)

// migrateV1Alpha9 upgrades a v1alpha9 inventory to v1alpha10 in place.
//
// Through v1alpha9 one value could be held twice: in an object's custom
// fields, and in its provider metadata, which the Nautobot exporter also sent
// as custom fields and read a device's role and location from. v1alpha10
// gives each kind of value one home. A modeled attribute is a field, an
// operator-defined attribute is a custom field, and provider metadata is a
// provider's own bucket that no other provider reads.
//
// The values that have to move are the ones no provider owns: the nautobot
// bucket, which held what was typed with --metadata, and every key outside a
// bucket. Each moves only when it is the value an export was sending, so the
// migration never makes a value live that was not:
//
//   - On a device, the role and location the exporter read from provider
//     metadata fill the device's fields (see migrateDeviceRole and
//     migrateDeviceLocation).
//   - Every other such key becomes a custom field and replaces a custom
//     field of the same key, as it did on export. The migration defines
//     the custom fields it creates this way (see defineMigratedFields).
//   - A copy the exporter passed over for another is dropped.
//   - A rack number outside a bucket goes to the HPCM bucket, which wrote it.
//
// Other providers' buckets are left alone. The custom fields the HPCM import
// wrote are its own records too, and move to its bucket (see
// moveHPCMCustomFields).
//
// From this generation a rack-mounted device may record a location of its
// own, so a location stored on a parented device, which earlier generations
// re-derived on every load, is dropped unless it is its rack's.
//
// Each value the migration drops, replaces or cannot place is logged by
// object and key, never by value.
func migrateV1Alpha9(inv *devicetypes.Inventory) {
	for _, line := range foldProviderMetadata(inv) {
		log.Printf("Migration to %s: %s", devicetypes.SchemaVersionV1Alpha10, line)
	}
	inv.SchemaVersion = devicetypes.SchemaVersionV1Alpha10
}

// metadataFold carries the inventory being migrated, the custom fields it
// has moved values into, and what it reports.
type metadataFold struct {
	inv    *devicetypes.Inventory
	moved  []movedField
	report []string
}

// foldProviderMetadata applies the v1alpha10 ownership rules to every object
// and returns the lines to report, sorted, each once.
func foldProviderMetadata(inv *devicetypes.Inventory) []string {
	fold := &metadataFold{inv: inv}
	eachRecord(inv.Devices, fold.device)
	for _, ref := range otherMetadata(inv) {
		fold.store(ref, fold.take(ref.owner, ref.provider))
	}
	// HPCM's records move out of the custom fields before any are defined:
	// a field with no definition is how its step tells them apart.
	fold.moveHPCMCustomFields()
	fold.defineMigratedFields()
	slices.Sort(fold.report)
	return slices.Compact(fold.report)
}

// note records one line about an object.
func (f *metadataFold) note(owner, format string, args ...any) {
	f.report = append(f.report, owner+": "+fmt.Sprintf(format, args...))
}

// take removes from provider every value no provider owns and returns the
// ones the exporter sent as custom fields. A copy it passed over is dropped
// and reported, a rack number outside a bucket moves to the HPCM bucket, and
// a map left empty becomes nil.
func (f *metadataFold) take(owner string, provider *map[string]any) map[string]any {
	sent := legacyFlatten(*provider)
	taken := make(map[string]any)
	for _, unowned := range removeUnowned(provider) {
		switch {
		case unowned.loose && unowned.key == hpcmRackNumberKey:
			setBucketKey(provider, hpcmBucket, hpcmRackNumberKey, unowned.value)
		case reflect.DeepEqual(unowned.value, sent[unowned.key]):
			taken[unowned.key] = unowned.value
		default:
			f.note(owner, "provider metadata %q was dropped: another copy of that key took precedence", unowned.key)
		}
	}
	if len(*provider) == 0 {
		*provider = nil
	}
	return taken
}

// legacyFlatten repeats the v1alpha9 exporter's projection of provider
// metadata onto custom fields. Names were visited in order, a bucket
// contributing its keys and any other value its own name, so the value
// under the later name was the one sent.
func legacyFlatten(provider map[string]any) map[string]any {
	flat := make(map[string]any)
	for _, name := range slices.Sorted(maps.Keys(provider)) {
		if bucket, held := provider[name].(map[string]any); held {
			maps.Copy(flat, bucket)
			continue
		}
		flat[name] = provider[name]
	}
	return flat
}

// unownedValue is one provider-metadata value no provider owns: a key
// outside the buckets (loose), or one in the nautobot bucket.
type unownedValue struct {
	key   string
	value any
	loose bool
}

// removeUnowned deletes every key outside a bucket, and the nautobot bucket,
// and returns what they held in a stable order.
func removeUnowned(provider *map[string]any) []unownedValue {
	var removed []unownedValue
	for _, name := range slices.Sorted(maps.Keys(*provider)) {
		bucket, isBucket := (*provider)[name].(map[string]any)
		switch {
		case !isBucket:
			removed = append(removed, unownedValue{key: name, value: (*provider)[name], loose: true})
		case name == legacyNautobotBucket:
			for _, key := range slices.Sorted(maps.Keys(bucket)) {
				removed = append(removed, unownedValue{key: key, value: bucket[key]})
			}
		default:
			continue
		}
		delete(*provider, name)
	}
	return removed
}

// store writes the taken values into the object's custom fields and records
// each so that defineMigratedFields can define it. Provider metadata took
// precedence over a custom field of the same key on export, so it does here;
// a custom field it replaces with another value is reported.
func (f *metadataFold) store(ref metadataRef, taken map[string]any) {
	for _, key := range slices.Sorted(maps.Keys(taken)) {
		if old, held := (*ref.custom)[key]; held && !reflect.DeepEqual(old, taken[key]) {
			f.note(ref.owner, "custom field %q held another value; the provider metadata value took precedence and replaced it", key)
		}
		if *ref.custom == nil {
			*ref.custom = make(map[string]any)
		}
		(*ref.custom)[key] = taken[key]
		f.moved = append(f.moved, movedField{ref: ref, key: key})
	}
}

// setBucketKey sets one key in a provider's bucket, creating what is missing.
func setBucketKey(provider *map[string]any, bucket, key string, value any) {
	if *provider == nil {
		*provider = make(map[string]any)
	}
	held, _ := (*provider)[bucket].(map[string]any)
	if held == nil {
		held = make(map[string]any)
		(*provider)[bucket] = held
	}
	held[key] = value
}

// metadataRef names one object's custom fields and provider metadata, and
// the content type its custom fields are defined for.
type metadataRef struct {
	owner            string
	contentType      string
	custom, provider *map[string]any
}

// otherMetadata lists every object but the devices, whose reserved keys
// need their own handling, along with every interface spec. The interface
// index is left out: it is rebuilt from the specs when the migration is done.
func otherMetadata(inv *devicetypes.Inventory) []metadataRef {
	var refs []metadataRef
	add := func(owner, contentType string, custom, provider *map[string]any) {
		refs = append(refs, metadataRef{owner: owner, contentType: contentType, custom: custom, provider: provider})
	}
	eachRecord(inv.Locations, func(id uuid.UUID, o *devicetypes.CaniLocationType) {
		add(ownerLabel("location", o.Name, id), devicetypes.ContentTypeLocation, &o.CustomFields, &o.ProviderMetadata)
	})
	eachRecord(inv.Racks, func(id uuid.UUID, o *devicetypes.CaniRackType) {
		add(ownerLabel("rack", o.Name, id), devicetypes.ContentTypeRack, &o.CustomFields, &o.ProviderMetadata)
	})
	eachRecord(inv.Modules, func(id uuid.UUID, o *devicetypes.CaniModuleType) {
		owner := ownerLabel("module", o.Label(), id)
		add(owner, devicetypes.ContentTypeModule, &o.CustomFields, &o.ProviderMetadata)
		addInterfaceSpecs(add, o.Interfaces, owner)
	})
	eachRecord(inv.Cables, func(id uuid.UUID, o *devicetypes.CaniCableType) {
		add(ownerLabel("cable", o.Label, id), devicetypes.ContentTypeCable, &o.CustomFields, &o.ProviderMetadata)
	})
	eachRecord(inv.Frus, func(id uuid.UUID, o *devicetypes.CaniFruType) {
		add(ownerLabel("inventory item", o.Name, id), devicetypes.ContentTypeInventoryItem, &o.CustomFields, &o.ProviderMetadata)
	})
	eachRecord(inv.VLANs, func(id uuid.UUID, o *devicetypes.CaniVLAN) {
		add(ownerLabel("VLAN", o.Name, id), devicetypes.ContentTypeVLAN, &o.CustomFields, &o.ProviderMetadata)
	})
	eachRecord(inv.Prefixes, func(id uuid.UUID, o *devicetypes.CaniPrefix) {
		add(ownerLabel("prefix", o.Prefix, id), devicetypes.ContentTypePrefix, &o.CustomFields, &o.ProviderMetadata)
	})
	eachRecord(inv.IPAddresses, func(id uuid.UUID, o *devicetypes.CaniIPAddress) {
		add(ownerLabel("IP address", o.Address, id), devicetypes.ContentTypeIPAddress, &o.CustomFields, &o.ProviderMetadata)
	})
	eachRecord(inv.VRFs, func(id uuid.UUID, o *devicetypes.CaniVRF) {
		add(ownerLabel("VRF", o.Name, id), devicetypes.ContentTypeVRF, &o.CustomFields, &o.ProviderMetadata)
	})
	eachRecord(inv.Devices, func(id uuid.UUID, o *devicetypes.CaniDeviceType) {
		addInterfaceSpecs(add, o.Interfaces, ownerLabel("device", o.Name, id))
	})
	return refs
}

// ownerLabel names an object in a report line. Names repeat, so the ID the
// inventory keys the object by is always given.
func ownerLabel(kind, name string, id uuid.UUID) string {
	return fmt.Sprintf("%s %q (%s)", kind, name, id)
}

// eachRecord calls visit for every record a collection actually holds.
func eachRecord[T any](records map[uuid.UUID]*T, visit func(uuid.UUID, *T)) {
	for id, record := range records {
		if record != nil {
			visit(id, record)
		}
	}
}

// addInterfaceSpecs lists the interface specs a device or module owns.
func addInterfaceSpecs(add func(string, string, *map[string]any, *map[string]any), specs []devicetypes.InterfaceSpec, owner string) {
	for i := range specs {
		add(fmt.Sprintf("interface %q of %s", specs[i].Name, owner), devicetypes.ContentTypeInterface, &specs[i].CustomFields, &specs[i].ProviderMetadata)
	}
}
