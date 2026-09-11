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
package devicetypes

import (
	"maps"
	"reflect"
	"slices"

	"github.com/google/uuid"
)

func restoreInventoryDynamicFields(source, clone *Inventory) {
	restoreObjectMap(source.Locations, clone.Locations, func(item *CaniLocationType) *ObjectMeta { return &item.ObjectMeta })
	restoreObjectMap(source.Racks, clone.Racks, func(item *CaniRackType) *ObjectMeta { return &item.ObjectMeta })
	restoreObjectMap(source.Devices, clone.Devices, func(item *CaniDeviceType) *ObjectMeta { return &item.ObjectMeta })
	restoreObjectMap(source.Modules, clone.Modules, func(item *CaniModuleType) *ObjectMeta { return &item.ObjectMeta })
	restoreObjectMap(source.Cables, clone.Cables, func(item *CaniCableType) *ObjectMeta { return &item.ObjectMeta })
	restoreObjectMap(source.Frus, clone.Frus, func(item *CaniFruType) *ObjectMeta { return &item.ObjectMeta })
	restoreObjectMap(source.Interfaces, clone.Interfaces, func(item *CaniInterface) *ObjectMeta { return &item.ObjectMeta })
	restoreObjectMap(source.VLANs, clone.VLANs, func(item *CaniVLAN) *ObjectMeta { return &item.ObjectMeta })
	restoreObjectMap(source.Prefixes, clone.Prefixes, func(item *CaniPrefix) *ObjectMeta { return &item.ObjectMeta })
	restoreObjectMap(source.IPAddresses, clone.IPAddresses, func(item *CaniIPAddress) *ObjectMeta { return &item.ObjectMeta })
	restoreObjectMap(source.VRFs, clone.VRFs, func(item *CaniVRF) *ObjectMeta { return &item.ObjectMeta })
	restoreHardwareDynamicFields(source, clone)
	restoreMetadataDefaults(source.Metadata, clone.Metadata)
}

func restoreTransformDynamicFields(source, clone *TransformResult) {
	restoreObjectMap(source.Locations, clone.Locations, func(item *CaniLocationType) *ObjectMeta { return &item.ObjectMeta })
	restoreObjectMap(source.Racks, clone.Racks, func(item *CaniRackType) *ObjectMeta { return &item.ObjectMeta })
	restoreObjectMap(source.Devices, clone.Devices, func(item *CaniDeviceType) *ObjectMeta { return &item.ObjectMeta })
	restoreObjectMap(source.Modules, clone.Modules, func(item *CaniModuleType) *ObjectMeta { return &item.ObjectMeta })
	restoreObjectMap(source.Cables, clone.Cables, func(item *CaniCableType) *ObjectMeta { return &item.ObjectMeta })
	restoreObjectMap(source.Frus, clone.Frus, func(item *CaniFruType) *ObjectMeta { return &item.ObjectMeta })
	restoreObjectMap(source.VLANs, clone.VLANs, func(item *CaniVLAN) *ObjectMeta { return &item.ObjectMeta })
	restoreObjectMap(source.Prefixes, clone.Prefixes, func(item *CaniPrefix) *ObjectMeta { return &item.ObjectMeta })
	restoreObjectMap(source.IPAddresses, clone.IPAddresses, func(item *CaniIPAddress) *ObjectMeta { return &item.ObjectMeta })
	restoreObjectMap(source.VRFs, clone.VRFs, func(item *CaniVRF) *ObjectMeta { return &item.ObjectMeta })
	restoreTransformHardwareDynamicFields(source, clone)
	restoreMetadataDefaults(source.Metadata, clone.Metadata)
}

func restoreObjectMap[T any](
	source, clone map[uuid.UUID]*T,
	metadata func(*T) *ObjectMeta,
) {
	for id, sourceItem := range source {
		cloneItem := clone[id]
		if sourceItem == nil || cloneItem == nil {
			continue
		}
		*metadata(cloneItem) = cloneObjectMeta(*metadata(sourceItem))
	}
}

func restoreHardwareDynamicFields(source, clone *Inventory) {
	for id, rack := range source.Racks {
		if rack != nil && clone.Racks[id] != nil {
			clone.Racks[id].ProviderDefaults = cloneAnyMap(rack.ProviderDefaults)
			clone.Racks[id].Source = rack.Source
			restoreDeviceBays(rack.DeviceBays, clone.Racks[id].DeviceBays)
		}
	}
	restoreDeviceDynamicFields(source.Devices, clone.Devices)
	restoreModuleDynamicFields(source.Modules, clone.Modules)
	restoreSources(source.Cables, clone.Cables, func(item *CaniCableType) *string { return &item.Source })
	restoreSources(source.Frus, clone.Frus, func(item *CaniFruType) *string { return &item.Source })
}

func restoreTransformHardwareDynamicFields(source, clone *TransformResult) {
	for id, rack := range source.Racks {
		if rack != nil && clone.Racks[id] != nil {
			clone.Racks[id].ProviderDefaults = cloneAnyMap(rack.ProviderDefaults)
			clone.Racks[id].Source = rack.Source
			restoreDeviceBays(rack.DeviceBays, clone.Racks[id].DeviceBays)
		}
	}
	restoreDeviceDynamicFields(source.Devices, clone.Devices)
	restoreModuleDynamicFields(source.Modules, clone.Modules)
	restoreSources(source.Cables, clone.Cables, func(item *CaniCableType) *string { return &item.Source })
	restoreSources(source.Frus, clone.Frus, func(item *CaniFruType) *string { return &item.Source })
}

func restoreDeviceDynamicFields(source, clone map[uuid.UUID]*CaniDeviceType) {
	for id, device := range source {
		if device == nil || clone[id] == nil {
			continue
		}
		clone[id].Source = device.Source
		restoreInterfaceMetadata(device.Interfaces, clone[id].Interfaces)
		restoreDeviceBays(device.DeviceBays, clone[id].DeviceBays)
	}
}

func restoreModuleDynamicFields(source, clone map[uuid.UUID]*CaniModuleType) {
	for id, module := range source {
		if module == nil || clone[id] == nil {
			continue
		}
		clone[id].Source = module.Source
		restoreInterfaceMetadata(module.Interfaces, clone[id].Interfaces)
	}
}

func restoreInterfaceMetadata(source, clone []InterfaceSpec) {
	for index := range source {
		if index < len(clone) {
			clone[index].ProviderMetadata = cloneAnyMap(source[index].ProviderMetadata)
		}
	}
}

func restoreDeviceBays(source, clone []DeviceBaySpec) {
	for index := range source {
		if index >= len(clone) {
			continue
		}
		clone[index].Extra = cloneAnyMap(source[index].Extra)
		clone[index].Allowed = cloneDeviceBaySlugRef(source[index].Allowed)
		clone[index].Default = cloneDeviceBaySlugRef(source[index].Default)
	}
}

func cloneDeviceBaySlugRef(source *DeviceBaySlugRef) *DeviceBaySlugRef {
	if source == nil {
		return nil
	}
	return &DeviceBaySlugRef{Slug: cloneAny(source.Slug), Types: cloneAny(source.Types)}
}

func restoreSources[T any](source, clone map[uuid.UUID]*T, field func(*T) *string) {
	for id, sourceItem := range source {
		if sourceItem != nil && clone[id] != nil {
			*field(clone[id]) = *field(sourceItem)
		}
	}
}

func restoreMetadataDefaults(source, clone *InventoryMetadata) {
	if source == nil || clone == nil {
		return
	}
	for index := range source.CustomFields {
		if index < len(clone.CustomFields) {
			clone.CustomFields[index].Default = cloneAny(source.CustomFields[index].Default)
		}
	}
}

func cloneObjectMeta(source ObjectMeta) ObjectMeta {
	clone := source
	clone.Tags = slices.Clone(source.Tags)
	clone.CustomFields = cloneAnyMap(source.CustomFields)
	clone.ExternalIDs = maps.Clone(source.ExternalIDs)
	clone.ProviderMetadata = cloneAnyMap(source.ProviderMetadata)
	return clone
}

func cloneAnyMap(source map[string]any) map[string]any {
	if source == nil {
		return nil
	}
	clone := make(map[string]any, len(source))
	for key, value := range source {
		clone[key] = cloneAny(value)
	}
	return clone
}

func cloneAny(value any) any {
	if value == nil {
		return nil
	}
	return cloneDynamicValue(reflect.ValueOf(value)).Interface()
}

func cloneDynamicValue(value reflect.Value) reflect.Value {
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		clone := cloneDynamicValue(value.Elem())
		wrapped := reflect.New(value.Type()).Elem()
		wrapped.Set(clone)
		return wrapped
	case reflect.Map:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		clone := reflect.MakeMapWithSize(value.Type(), value.Len())
		iterator := value.MapRange()
		for iterator.Next() {
			clone.SetMapIndex(iterator.Key(), cloneDynamicValue(iterator.Value()))
		}
		return clone
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		clone := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for index := 0; index < value.Len(); index++ {
			clone.Index(index).Set(cloneDynamicValue(value.Index(index)))
		}
		return clone
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		clone := reflect.New(value.Type().Elem())
		clone.Elem().Set(cloneDynamicValue(value.Elem()))
		return clone
	default:
		return value
	}
}
