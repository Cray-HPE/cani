/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
	"context"
	"fmt"
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
func compareDeviceFields(
	device *devicetypes.CaniDeviceType,
	remote *nautobotapi.Device,
	mapper *DeviceMapper,
) []FieldDiff {
	var diffs []FieldDiff

	// Compare device type
	if dt := resolveLocalDeviceType(device, mapper); dt != nil {
		if remoteID := refID(remote.DeviceType.Id); remoteID != dt.ID {
			remoteName := mapper.cache.FindNameByID("deviceType", remoteID)
			diffs = append(diffs, FieldDiff{
				Field:     "device_type",
				LocalVal:  dt.Name,
				RemoteVal: remoteName,
			})
		}
	}

	// Compare location
	if loc := resolveLocalLocation(device, mapper); loc != nil {
		if remoteID := refID(remote.Location.Id); remoteID != loc.ID {
			remoteName := mapper.cache.FindNameByID("location", remoteID)
			diffs = append(diffs, FieldDiff{
				Field:     "location",
				LocalVal:  loc.Name,
				RemoteVal: remoteName,
			})
		}
	}

	// Compare status
	if st := resolveLocalStatus(device, mapper); st != nil {
		if remoteID := refID(remote.Status.Id); remoteID != st.ID {
			remoteName := mapper.cache.FindNameByID("status", remoteID)
			diffs = append(diffs, FieldDiff{
				Field:     "status",
				LocalVal:  st.Name,
				RemoteVal: remoteName,
			})
		}
	}

	// Compare role
	if rl := resolveLocalRole(device, mapper); rl != nil {
		if remoteID := refID(remote.Role.Id); remoteID != rl.ID {
			remoteName := mapper.cache.FindNameByID("role", remoteID)
			diffs = append(diffs, FieldDiff{
				Field:     "role",
				LocalVal:  rl.Name,
				RemoteVal: remoteName,
			})
		}
	}

	// Compare rack
	diffs = append(diffs, compareRack(device, remote, mapper)...)

	// Compare position
	diffs = append(diffs, comparePosition(device, remote)...)

	// Compare face
	diffs = append(diffs, compareFace(device, remote)...)

	// Compare serial
	if device.Serial != "" {
		remoteSerial := ptrStr(remote.Serial)
		if device.Serial != remoteSerial {
			diffs = append(diffs, FieldDiff{
				Field:     "serial",
				LocalVal:  device.Serial,
				RemoteVal: remoteSerial,
			})
		}
	}

	// Compare asset tag
	if device.AssetTag != "" {
		remoteAsset := ptrStr(remote.AssetTag)
		if device.AssetTag != remoteAsset {
			diffs = append(diffs, FieldDiff{
				Field:     "asset_tag",
				LocalVal:  device.AssetTag,
				RemoteVal: remoteAsset,
			})
		}
	}

	return diffs
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

// fetchFullDeviceByID retrieves the full Device object from the Nautobot API by UUID.
// This avoids ambiguity when multiple devices share the same name.
func (e *Exporter) fetchFullDeviceByID(ctx context.Context, id uuid.UUID) (*nautobotapi.Device, error) {
	resp, err := e.Client.DcimDevicesRetrieveWithResponse(ctx, id, &nautobotapi.DcimDevicesRetrieveParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch device %s: %w", id, err)
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to fetch device %s: status %d", id, resp.StatusCode())
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("device %s not found", id)
	}
	return resp.JSON200, nil
}

// ptrStr dereferences a *string, returning "" for nil.
func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// orNone returns "(none)" when s is empty.
func orNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}

// printDeviceDiffs prints the field diffs for a device using colored output.
func printDeviceDiffs(deviceName string, diffs []FieldDiff) {
	if len(diffs) == 0 {
		return
	}
	clog.Changed("  %s: %d field(s) would change with --merge:", deviceName, len(diffs))
	for _, d := range diffs {
		clog.Diff(d.Field, d.RemoteVal, d.LocalVal)
	}
}

// fetchFullDevice retrieves the full Device object from the Nautobot API by name.
func (e *Exporter) fetchFullDevice(ctx context.Context, name string) (*nautobotapi.Device, error) {
	nameFilter := []string{name}
	resp, err := e.Client.DcimDevicesListWithResponse(ctx, &nautobotapi.DcimDevicesListParams{
		Name: &nameFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch full device %s: %w", name, err)
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to fetch full device %s: status %d", name, resp.StatusCode())
	}
	if resp.JSON200 == nil || resp.JSON200.Results == nil || len(resp.JSON200.Results) == 0 {
		return nil, fmt.Errorf("device %s not found", name)
	}
	d := resp.JSON200.Results[0]
	return &d, nil
}
