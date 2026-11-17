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
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// The custom fields the v1alpha9 HPCM import wrote on a module and on a
// location, and whether each held a map rather than text.
var (
	hpcmModuleFields   = map[string]bool{"aliases": true, "geoloc": false, "hpcm_uuid": false, "location": true}
	hpcmLocationFields = map[string]bool{"hpcm_uuid": false}
)

// moveHPCMCustomFields moves what the v1alpha9 HPCM import stored as custom
// fields into the HPCM bucket, where the import writes it from v1alpha10.
//
// The import wrote a module's UUID, aliases, location and geoloc, a
// location's UUID, and every field of an inventory item it had no field
// for, as custom fields. They are HPCM's own records of an object, not
// attributes an operator defined, and as custom fields every export sent
// them. A later import refreshes only the bucket, so a copy left behind
// would go stale and still be exported.
//
// Releases through v0.8.4 also imported custom-field values from Nautobot
// without their definitions, so a custom field with no definition is not
// HPCM's for that reason alone. One moves only when it can be told to be:
// on a module or location by its key and the shape of its value, and on an
// inventory item by the item being one the import built. A key the
// inventory defines is an attribute in its own right and stays, and so does
// a key the bucket already holds.
func (f *metadataFold) moveHPCMCustomFields() {
	defined := make(map[string]bool)
	for _, definition := range f.inv.ListCustomFields() {
		defined[definition.Key] = true
	}
	named := func(fields map[string]bool) func(string, any) bool {
		return func(key string, value any) bool { return !defined[key] && isHPCMField(fields, key, value) }
	}

	moved := 0
	eachRecord(f.inv.Modules, func(_ uuid.UUID, module *devicetypes.CaniModuleType) {
		moved += moveToHPCMBucket(&module.ObjectMeta, named(hpcmModuleFields))
	})
	eachRecord(f.inv.Locations, func(_ uuid.UUID, location *devicetypes.CaniLocationType) {
		moved += moveToHPCMBucket(&location.ObjectMeta, named(hpcmLocationFields))
	})
	// After the modules: an item found on a module is told by the module's
	// bucket.
	eachRecord(f.inv.Frus, func(_ uuid.UUID, item *devicetypes.CaniFruType) {
		if f.builtByHPCM(item) {
			moved += moveToHPCMBucket(&item.ObjectMeta, func(key string, _ any) bool { return !defined[key] })
		}
	})
	if moved > 0 {
		f.report = append(f.report, fmt.Sprintf(
			"%d custom field(s) the HPCM import wrote on modules, locations and inventory items moved to providerMetadata.%s",
			moved, hpcmBucket))
	}
}

// isHPCMField reports whether a custom field is one of the named HPCM
// fields and holds the kind of value the import gave it.
func isHPCMField(fields map[string]bool, key string, value any) bool {
	holdsMap, named := fields[key]
	if !named {
		return false
	}
	if _, isMap := value.(map[string]any); isMap {
		return holdsMap
	}
	_, isText := value.(string)
	return isText && !holdsMap
}

// moveToHPCMBucket moves every custom field owned says is HPCM's into the
// object's HPCM bucket and returns how many moved. A custom-field map left
// empty becomes nil.
func moveToHPCMBucket(meta *devicetypes.ObjectMeta, owned func(key string, value any) bool) int {
	bucket, _ := meta.ProviderMetadata[hpcmBucket].(map[string]any)
	moved := 0
	for key, value := range meta.CustomFields {
		if _, held := bucket[key]; held || !owned(key, value) {
			continue
		}
		setBucketKey(&meta.ProviderMetadata, hpcmBucket, key, value)
		delete(meta.CustomFields, key)
		moved++
	}
	if len(meta.CustomFields) == 0 {
		meta.CustomFields = nil
	}
	return moved
}

// builtByHPCM reports whether the HPCM import built an inventory item. It
// named each one after the node it was found on and the inventory group it
// came from, as in "node001-disk.disk0", and typed it by that group,
// "disk". The node is a device or module that has an HPCM bucket.
func (f *metadataFold) builtByHPCM(item *devicetypes.CaniFruType) bool {
	for _, id := range []uuid.UUID{item.Parent, item.Device} {
		node, held := f.hpcmNodeName(id)
		group, named := strings.CutPrefix(item.Name, node+"-")
		kind, _, _ := strings.Cut(group, ".")
		if held && named && kind != "" && kind == string(item.Type) {
			return true
		}
	}
	return false
}

// hpcmNodeName returns the name of the device or module with this ID, and
// whether it has an HPCM bucket.
func (f *metadataFold) hpcmNodeName(id uuid.UUID) (string, bool) {
	if device := f.inv.Devices[id]; device != nil {
		_, held := device.ProviderMetadata[hpcmBucket].(map[string]any)
		return device.Name, held
	}
	if module := f.inv.Modules[id]; module != nil {
		_, held := module.ProviderMetadata[hpcmBucket].(map[string]any)
		return module.Name, held
	}
	return "", false
}
