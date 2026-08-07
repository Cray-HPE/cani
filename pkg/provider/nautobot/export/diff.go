/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024, 2026 Hewlett Packard Enterprise Development LP
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
package export

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// FieldDiff represents a single field difference between local and remote.
type FieldDiff struct {
	Field     string // e.g. "device_type", "location", "status"
	LocalVal  string // what the local inventory wants to set
	RemoteVal string // what Nautobot currently has
}

// compareDeviceFields compares the local device intent against the existing
// Nautobot device and returns a list of fields that would change.
// It resolves UUIDs to human-readable names via the mapper and cache.
// refFieldSpec drives the table-driven comparison of UUID-referenced fields.
type refFieldSpec struct {
	field    string                                                                  // diff field name
	cache    string                                                                  // cache category for reverse lookup
	resolve  func(*devicetypes.CaniDeviceType, *DeviceMapper) *CachedItem            // resolve local intent
	remoteID func(*nautobotapi.Device) *nautobotapi.BulkWritableCableRequestStatusId // extract remote ref
}

func compareDeviceFields(
	device *devicetypes.CaniDeviceType,
	remote *nautobotapi.Device,
	mapper *DeviceMapper,
) []FieldDiff {
	var diffs []FieldDiff

	// Table-driven comparison for UUID-referenced fields.
	refFields := []refFieldSpec{
		{"device_type", "deviceType", resolveLocalDeviceType, func(d *nautobotapi.Device) *nautobotapi.BulkWritableCableRequestStatusId { return d.DeviceType.Id }},
		{"location", "location", resolveLocalLocation, func(d *nautobotapi.Device) *nautobotapi.BulkWritableCableRequestStatusId { return d.Location.Id }},
		{"status", "status", resolveLocalStatus, func(d *nautobotapi.Device) *nautobotapi.BulkWritableCableRequestStatusId { return d.Status.Id }},
		{"role", "role", resolveLocalRole, func(d *nautobotapi.Device) *nautobotapi.BulkWritableCableRequestStatusId { return d.Role.Id }},
	}
	for _, spec := range refFields {
		if local := spec.resolve(device, mapper); local != nil {
			if remoteID := refID(spec.remoteID(remote)); remoteID != local.ID {
				diffs = append(diffs, FieldDiff{
					Field:     spec.field,
					LocalVal:  local.Name,
					RemoteVal: mapper.cache.FindNameByID(spec.cache, remoteID),
				})
			}
		}
	}

	diffs = append(diffs, compareRack(device, remote, mapper)...)
	diffs = append(diffs, comparePosition(device, remote)...)
	diffs = append(diffs, compareFace(device, remote)...)
	diffs = append(diffs, compareScalarFields(device, remote)...)
	diffs = append(diffs, compareCustomFields(device.FlattenProviderMetadata(), remote.CustomFields)...)

	return diffs
}

// compareCustomFields compares local custom field values against the remote
// object's custom_fields map, returning diffs for changed or new values.
func compareCustomFields(local map[string]any, remote *map[string]interface{}) []FieldDiff {
	if len(local) == 0 {
		return nil
	}
	remoteMap := map[string]interface{}{}
	if remote != nil {
		remoteMap = *remote
	}
	var diffs []FieldDiff
	for _, k := range sortedMapKeys(local) {
		localStr := fmt.Sprintf("%v", local[k])
		remoteVal, exists := remoteMap[k]
		remoteStr := ""
		if exists {
			remoteStr = fmt.Sprintf("%v", remoteVal)
		}
		if localStr != remoteStr {
			diffs = append(diffs, FieldDiff{
				Field:     "custom_field:" + k,
				LocalVal:  localStr,
				RemoteVal: orNone(remoteStr),
			})
		}
	}
	return diffs
}

// sortedMapKeys returns the keys of a map in sorted order.
func sortedMapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// resolveLocalDeviceType resolves the device type from the mapper, ignoring errors.
func resolveLocalDeviceType(device *devicetypes.CaniDeviceType, mapper *DeviceMapper) *CachedItem {
	item, err := mapper.resolveDeviceType(device)
	if err != nil {
		return nil
	}
	return item
}

// resolveLocalLocation resolves the location from the mapper, ignoring errors.
func resolveLocalLocation(device *devicetypes.CaniDeviceType, mapper *DeviceMapper) *CachedItem {
	item, err := mapper.resolveLocation(device)
	if err != nil {
		return nil
	}
	return item
}

// resolveLocalStatus resolves the status from the mapper, ignoring errors.
func resolveLocalStatus(device *devicetypes.CaniDeviceType, mapper *DeviceMapper) *CachedItem {
	item, err := mapper.resolveStatus(device)
	if err != nil {
		return nil
	}
	return item
}

// resolveLocalRole resolves the role from the mapper, ignoring errors.
func resolveLocalRole(device *devicetypes.CaniDeviceType, mapper *DeviceMapper) *CachedItem {
	item, err := mapper.resolveRole(device)
	if err != nil {
		return nil
	}
	return item
}

// compareScalarFields compares simple string fields (serial, asset_tag).
func compareScalarFields(device *devicetypes.CaniDeviceType, remote *nautobotapi.Device) []FieldDiff {
	type scalarSpec struct {
		field    string
		local    string
		remoteFn func() string
	}
	specs := []scalarSpec{
		{"serial", device.Serial, func() string { return ptrStr(remote.Serial) }},
		{"asset_tag", device.AssetTag, func() string { return ptrStr(remote.AssetTag) }},
	}
	var diffs []FieldDiff
	for _, s := range specs {
		if s.local == "" {
			continue
		}
		if rv := s.remoteFn(); s.local != rv {
			diffs = append(diffs, FieldDiff{Field: s.field, LocalVal: s.local, RemoteVal: rv})
		}
	}
	return diffs
}

// compareRack compares the rack assignment between local and remote.
// The comparison uses Nautobot rack UUIDs to avoid false positives caused
// by the rack cache not being populated in FindNameByID.
func compareRack(device *devicetypes.CaniDeviceType, remote *nautobotapi.Device, mapper *DeviceMapper) []FieldDiff {
	// Resolve local rack to its Nautobot UUID via a name-based API lookup.
	localRackName := resolveLocalRackName(device, mapper)
	var localRackID uuid.UUID
	if localRackName != "" {
		if cached, err := mapper.cache.GetRackByName(localRackName); err == nil && cached != nil {
			localRackID = cached.ID
		}
	}

	// Extract remote rack UUID from the Nautobot device response.
	var remoteRackID uuid.UUID
	if remote.Rack != nil && remote.Rack.Id != nil {
		remoteRackID = tenantRefID(remote.Rack.Id)
	}

	if localRackID == uuid.Nil && remoteRackID == uuid.Nil {
		return nil
	}
	if localRackID != remoteRackID {
		return []FieldDiff{{
			Field:     "rack",
			LocalVal:  orNone(localRackName),
			RemoteVal: orNone(remoteRackID.String()),
		}}
	}
	return nil
}

// comparePosition compares the rack position between local and remote.
func comparePosition(device *devicetypes.CaniDeviceType, remote *nautobotapi.Device) []FieldDiff {
	if device.RackPosition <= 0 {
		return nil
	}
	remotePos := 0
	if remote.Position != nil {
		remotePos = *remote.Position
	}
	if device.RackPosition != remotePos {
		return []FieldDiff{{
			Field:     "position",
			LocalVal:  strconv.Itoa(device.RackPosition),
			RemoteVal: strconv.Itoa(remotePos),
		}}
	}
	return nil
}

// compareFace compares the rack face between local and remote.
func compareFace(device *devicetypes.CaniDeviceType, remote *nautobotapi.Device) []FieldDiff {
	if device.Face == "" {
		return nil
	}
	remoteFace := ""
	if remote.Face != nil && remote.Face.Value != nil {
		remoteFace = string(*remote.Face.Value)
	}
	if device.Face != remoteFace {
		return []FieldDiff{{
			Field:     "face",
			LocalVal:  device.Face,
			RemoteVal: orNone(remoteFace),
		}}
	}
	return nil
}

// resolveLocalRackName returns the rack name from the local inventory.
func resolveLocalRackName(device *devicetypes.CaniDeviceType, mapper *DeviceMapper) string {
	if mapper.inventory == nil {
		return ""
	}
	rackID := device.GetRackID(mapper.inventory)
	if rackID == uuid.Nil {
		return ""
	}
	if rack, ok := mapper.inventory.Racks[rackID]; ok && rack != nil {
		return rack.Name
	}
	return ""
}

// refID extracts a UUID from a BulkWritableCableRequestStatusId union.
func refID(id *nautobotapi.BulkWritableCableRequestStatusId) uuid.UUID {
	if id == nil {
		return uuid.Nil
	}
	u, err := id.AsBulkWritableCableRequestStatusId0()
	if err != nil {
		return uuid.Nil
	}
	return uuid.UUID(u)
}

// tenantRefID extracts a UUID from a BulkWritableCableRequestStatusId used
// in BulkWritableCircuitRequestTenant references.
func tenantRefID(id *nautobotapi.BulkWritableCableRequestStatusId) uuid.UUID {
	return refID(id)
}
