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
	"fmt"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// MapperOpts contains options for mapping CaniDeviceType to Nautobot
type MapperOpts struct {
	DefaultLocation string
	DefaultRole     string
	DefaultStatus   string
	Strict          bool // when false, devices without a slug are skipped rather than errored
}

// DeviceMapper converts CaniDeviceType to Nautobot API requests
type DeviceMapper struct {
	cache     *LookupCache
	defaults  *MapperOpts
	inventory *devicetypes.Inventory
}

// NewDeviceMapper creates a new device mapper
func NewDeviceMapper(cache *LookupCache, defaults *MapperOpts) *DeviceMapper {
	return &DeviceMapper{
		cache:    cache,
		defaults: defaults,
	}
}

// SetInventory sets the inventory reference for resolving parent devices
func (m *DeviceMapper) SetInventory(inv *devicetypes.Inventory) {
	m.inventory = inv
}

// errMsgDeviceNil is returned by mapper methods when given a nil device.
const errMsgDeviceNil = "device is nil"

// MapToNautobotDevice converts a CaniDeviceType to a BulkWritableDeviceRequest
func (m *DeviceMapper) MapToNautobotDevice(device *devicetypes.CaniDeviceType) (*nautobotapi.BulkWritableDeviceRequest, error) {
	if device == nil {
		return nil, fmt.Errorf(errMsgDeviceNil)
	}

	// Lookup required references
	deviceType, err := m.resolveDeviceType(device)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve device type for %s: %w", device.Name, err)
	}

	location, err := m.resolveLocation(device)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve location for %s: %w", device.Name, err)
	}

	status, err := m.resolveStatus(device)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve status for %s: %w", device.Name, err)
	}

	role, err := m.resolveRole(device)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve role for %s: %w", device.Name, err)
	}

	// Build the request
	req := &nautobotapi.BulkWritableDeviceRequest{
		Id:         device.ID,
		Name:       &device.Name,
		DeviceType: makeStatusRef(deviceType.ID),
		Location:   makeStatusRef(location.ID),
		Status:     makeStatusRef(status.ID),
		Role:       makeStatusRef(role.ID),
	}

	// Map optional fields - use flattened ProviderMetadata for custom fields
	if flat := device.FlattenProviderMetadata(); len(flat) > 0 {
		customFields := make(map[string]interface{}, len(flat))
		for k, v := range flat {
			customFields[k] = v
		}
		req.CustomFields = &customFields
	}

	// Map serial number if available
	if device.Serial != "" {
		serial := device.Serial
		req.Serial = &serial
	}

	// Map asset tag if available
	if device.AssetTag != "" {
		assetTag := device.AssetTag
		req.AssetTag = &assetTag
	}

	// Map comments if available
	if device.Comments != "" {
		comments := device.Comments
		req.Comments = &comments
	}

	return req, nil
}

// MapToWritableDeviceRequest converts a CaniDeviceType to a WritableDeviceRequest for single creates
func (m *DeviceMapper) MapToWritableDeviceRequest(device *devicetypes.CaniDeviceType) (*nautobotapi.WritableDeviceRequest, error) {
	if device == nil {
		return nil, fmt.Errorf(errMsgDeviceNil)
	}

	// Lookup required references
	deviceType, err := m.resolveDeviceType(device)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve device type for %s: %w", device.Name, err)
	}

	location, err := m.resolveLocation(device)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve location for %s: %w", device.Name, err)
	}

	status, err := m.resolveStatus(device)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve status for %s: %w", device.Name, err)
	}

	role, err := m.resolveRole(device)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve role for %s: %w", device.Name, err)
	}

	// Build the request
	req := &nautobotapi.WritableDeviceRequest{
		Name:       &device.Name,
		DeviceType: makeStatusRef(deviceType.ID),
		Location:   makeStatusRef(location.ID),
		Status:     makeStatusRef(status.ID),
		Role:       makeStatusRef(role.ID),
	}

	// Map optional fields - use flattened ProviderMetadata for custom fields
	if flat := device.FlattenProviderMetadata(); len(flat) > 0 {
		customFields := make(map[string]interface{}, len(flat))
		for k, v := range flat {
			customFields[k] = v
		}
		req.CustomFields = &customFields
	}

	// Map serial number if available
	if device.Serial != "" {
		serial := device.Serial
		req.Serial = &serial
	}

	// Map asset tag if available
	if device.AssetTag != "" {
		assetTag := device.AssetTag
		req.AssetTag = &assetTag
	}

	// Map comments if available
	if device.Comments != "" {
		req.Comments = &device.Comments
	}

	// Map rack and position if device has a parent rack
	rackID := device.GetRackID(m.inventory)
	if rackID != uuid.Nil && m.inventory != nil {
		// First check if the parent is a rack in the Racks collection
		if parentRack, ok := m.inventory.Racks[rackID]; ok && parentRack != nil {
			// Look up the rack in Nautobot by name
			rack, err := m.cache.GetRackByName(parentRack.Name)
			if err == nil && rack != nil {
				// Build rack reference using the same pattern as other references
				rackIDUnion := nautobotapi.BulkWritableCableRequestStatusId{}
				rackIDUnion.FromBulkWritableCableRequestStatusId0(rack.ID)
				req.Rack = &nautobotapi.BulkWritableCircuitRequestTenant{
					Id: &rackIDUnion,
				}

				// Set position from RackPosition field
				if device.RackPosition > 0 {
					pos := device.RackPosition
					req.Position = &pos
				}

				// Use actual Face value, defaulting to "front" if unset
				req.Face = resolveFace(device.Face)
			}
		} else if parentDevice := m.inventory.Devices[rackID]; parentDevice != nil && parentDevice.Type == devicetypes.Rack {
			// Fallback: check if parent is a rack-type device in Devices collection (legacy)
			rack, err := m.cache.GetRackByName(parentDevice.Name)
			if err == nil && rack != nil {
				rackIDUnion := nautobotapi.BulkWritableCableRequestStatusId{}
				rackIDUnion.FromBulkWritableCableRequestStatusId0(rack.ID)
				req.Rack = &nautobotapi.BulkWritableCircuitRequestTenant{
					Id: &rackIDUnion,
				}

				if device.RackPosition > 0 {
					pos := device.RackPosition
					req.Position = &pos
				}

				req.Face = resolveFace(device.Face)
			}
		}
	}

	return req, nil
}

// MapToPatchRequest converts a CaniDeviceType to a PatchedWritableDeviceRequest for updates
func (m *DeviceMapper) MapToPatchRequest(device *devicetypes.CaniDeviceType, existingID uuid.UUID) (*nautobotapi.PatchedWritableDeviceRequest, error) {
	if device == nil {
		return nil, fmt.Errorf(errMsgDeviceNil)
	}

	req := &nautobotapi.PatchedWritableDeviceRequest{
		Name: &device.Name,
	}

	// Only set device type if specified
	if device.Slug != "" {
		deviceType, err := m.resolveDeviceType(device)
		if err == nil {
			ref := makeStatusRef(deviceType.ID)
			req.DeviceType = &ref
		}
	}

	// Resolve and set location
	if location, err := m.resolveLocation(device); err == nil {
		ref := makeStatusRef(location.ID)
		req.Location = &ref
	}

	// Resolve and set status
	if status, err := m.resolveStatus(device); err == nil {
		ref := makeStatusRef(status.ID)
		req.Status = &ref
	}

	// Resolve and set role
	if role, err := m.resolveRole(device); err == nil {
		ref := makeStatusRef(role.ID)
		req.Role = &ref
	}

	// Map optional fields - use flattened ProviderMetadata for custom fields
	if flat := device.FlattenProviderMetadata(); len(flat) > 0 {
		customFields := make(map[string]interface{}, len(flat))
		for k, v := range flat {
			customFields[k] = v
		}
		req.CustomFields = &customFields
	}

	// Map serial number if available
	if device.Serial != "" {
		serial := device.Serial
		req.Serial = &serial
	}

	// Map asset tag if available
	if device.AssetTag != "" {
		assetTag := device.AssetTag
		req.AssetTag = &assetTag
	}

	// Map comments if available
	if device.Comments != "" {
		req.Comments = &device.Comments
	}

	// Map rack and position if device has a parent rack
	if device.Parent != uuid.Nil && m.inventory != nil {
		// First check if the parent is a rack in the Racks collection
		if parentRack, ok := m.inventory.Racks[device.Parent]; ok && parentRack != nil {
			// Look up the rack in Nautobot by name
			rack, err := m.cache.GetRackByName(parentRack.Name)
			if err == nil && rack != nil {
				rackIDUnion := nautobotapi.BulkWritableCableRequestStatusId{}
				rackIDUnion.FromBulkWritableCableRequestStatusId0(rack.ID)
				rackRef := &nautobotapi.BulkWritableCircuitRequestTenant{
					Id: &rackIDUnion,
				}
				req.Rack = rackRef

				if device.RackPosition > 0 {
					pos := device.RackPosition
					req.Position = &pos
				}

				req.Face = resolveFace(device.Face)
			}
		} else if parentDevice := m.inventory.Devices[device.Parent]; parentDevice != nil && parentDevice.Type == devicetypes.Rack {
			// Fallback: check if parent is a rack-type device in Devices collection (legacy)
			rack, err := m.cache.GetRackByName(parentDevice.Name)
			if err == nil && rack != nil {
				rackIDUnion := nautobotapi.BulkWritableCableRequestStatusId{}
				rackIDUnion.FromBulkWritableCableRequestStatusId0(rack.ID)
				rackRef := &nautobotapi.BulkWritableCircuitRequestTenant{
					Id: &rackIDUnion,
				}
				req.Rack = rackRef

				if device.RackPosition > 0 {
					pos := device.RackPosition
					req.Position = &pos
				}

				req.Face = resolveFace(device.Face)
			}
		}
	}

	return req, nil
}

// ErrDeviceUnclassified is a sentinel indicating a device has no slug.
// Callers can check for this to skip rather than fail.
var ErrDeviceUnclassified = fmt.Errorf("device is unclassified (no slug or model)")

// resolveDeviceType gets the device type ID, erroring if not found.
// When mapper opts are non-strict, returns ErrDeviceUnclassified instead of
// a generic error so callers can choose to skip.
// Falls back to CreateDeviceTypeFromCaniDevice when the local library lookup
// fails but the inventory record has enough data to create the type.
func (m *DeviceMapper) resolveDeviceType(device *devicetypes.CaniDeviceType) (*CachedItem, error) {
	slug := device.Slug
	if slug == "" {
		slug = device.Model
	}
	if slug == "" {
		if !m.defaults.Strict {
			return nil, ErrDeviceUnclassified
		}
		return nil, fmt.Errorf("device type slug is required")
	}

	item, err := m.cache.GetDeviceType(slug)
	if err == nil {
		return item, nil
	}

	// Fallback: create the device type from inventory data when the local
	// YAML library does not contain the slug.
	if m.cache.createDeviceTypes {
		return m.cache.CreateDeviceTypeFromCaniDevice(device)
	}
	return nil, err
}

// resolveLocation gets the location ID, using default if device doesn't specify one.
// Priority: device ProviderMetadata["location"] → parent rack's location → DefaultLocation → "Default".
// If create_locations is enabled and no location name is available, uses "Default" as the location name.
func (m *DeviceMapper) resolveLocation(device *devicetypes.CaniDeviceType) (*CachedItem, error) {
	name := ""
	if device.ProviderMetadata != nil {
		if loc, _ := device.GetProviderMeta("location"); loc != nil {
			if locStr, ok := loc.(string); ok {
				name = locStr
			}
		}
	}

	// Inherit location from parent rack so the device stays in the rack's
	// location hierarchy (Nautobot rejects devices whose location does not
	// contain their rack).
	if name == "" {
		name = m.locationFromParentRack(device)
	}

	if name == "" {
		name = m.defaults.DefaultLocation
	}
	if name == "" {
		if m.cache.createLocations {
			name = "Default"
			clog.Detail("[mapper] No location specified, using '%s' for auto-creation", name)
		} else {
			return nil, fmt.Errorf("location is required (use --default-location)")
		}
	}
	return m.cache.GetLocation(name)
}

// locationFromParentRack returns the location name of the device's parent rack,
// or "" if the device has no parent rack or the rack has no location set.
// If the rack's location type doesn't support devices, it walks the location
// tree to find the deepest descendant that does.
func (m *DeviceMapper) locationFromParentRack(device *devicetypes.CaniDeviceType) string {
	if m.inventory == nil {
		return ""
	}
	rackID := device.GetRackID(m.inventory)
	if rackID == uuid.Nil {
		return ""
	}
	if rack, ok := m.inventory.Racks[rackID]; ok && rack != nil && rack.Location != uuid.Nil {
		resolved := resolveContentLocation(rack.Location, "device", m.inventory)
		if resolved != "" {
			return resolved
		}
	}
	return ""
}

// resolveStatus gets the status ID, using default if device doesn't specify one
// If create_statuses is enabled and no status name is available, uses "Active" as the status name
func (m *DeviceMapper) resolveStatus(device *devicetypes.CaniDeviceType) (*CachedItem, error) {
	name := device.Status
	if name == "" {
		name = m.defaults.DefaultStatus
	}
	if name == "" {
		// If create_statuses is enabled, use "Active" as the auto-created status name
		if m.cache.createStatuses {
			name = "Active"
			clog.Detail("[mapper] No status specified, using '%s' for auto-creation", name)
		} else {
			return nil, fmt.Errorf("status is required (use --default-status)")
		}
	}
	return m.cache.GetStatus(name)
}

// resolveRole gets the role ID, using default if device doesn't specify one.
// Checks the explicit Role field first, then falls back to ProviderMetadata["role"],
// then to the provider default, then to "Generic" if auto-creation is enabled.
func (m *DeviceMapper) resolveRole(device *devicetypes.CaniDeviceType) (*CachedItem, error) {
	// Prefer the explicit Role field
	name := device.Role

	// Fall back to ProviderMetadata["role"] for backwards compatibility
	if name == "" && device.ProviderMetadata != nil {
		if role, _ := device.GetProviderMeta("role"); role != nil {
			if roleStr, ok := role.(string); ok {
				name = roleStr
			}
		}
	}
	if name == "" {
		name = m.defaults.DefaultRole
	}
	if name == "" {
		// If create_roles is enabled, use "Generic" as the auto-created role name
		if m.cache.createRoles {
			name = "Generic"
			clog.Detail("[mapper] No role specified, using '%s' for auto-creation", name)
		} else {
			return nil, fmt.Errorf("role is required (use --default-role)")
		}
	}
	role, err := m.cache.GetRole(name)
	if err != nil {
		return nil, err
	}
	// clog.Detail("[mapper] Resolved role '%s' to ID: %s", name, role.ID)
	return role, nil
}

// makeStatusRef creates a BulkWritableCableRequestStatus reference from a UUID
func makeStatusRef(id uuid.UUID) nautobotapi.BulkWritableCableRequestStatus {
	statusID := nautobotapi.BulkWritableCableRequestStatusId{}
	statusID.FromBulkWritableCableRequestStatusId0(id)
	return nautobotapi.BulkWritableCableRequestStatus{
		Id: &statusID,
	}
}

// MapToWritableRackRequest converts a CaniDeviceType (rack) to a WritableRackRequest
func (m *DeviceMapper) MapToWritableRackRequest(device *devicetypes.CaniDeviceType) (*nautobotapi.WritableRackRequest, error) {
	if device == nil {
		return nil, fmt.Errorf(errMsgDeviceNil)
	}

	// Resolve location
	location, err := m.resolveLocation(device)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve location for rack %s: %w", device.Name, err)
	}

	// Resolve status
	status, err := m.resolveStatus(device)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve status for rack %s: %w", device.Name, err)
	}

	// Build the request
	req := &nautobotapi.WritableRackRequest{
		Name:     device.Name,
		Location: makeStatusRef(location.ID),
		Status:   makeStatusRef(status.ID),
	}

	// Set rack height (default to 48U if not specified)
	uHeight := 48
	if device.ProviderMetadata != nil {
		if h, _ := device.GetProviderMeta("u_height"); h != nil {
			if hInt, ok := h.(int); ok && hInt > 0 {
				uHeight = hInt
			}
		}
	}
	req.UHeight = &uHeight

	// Map optional fields
	if flat := device.FlattenProviderMetadata(); len(flat) > 0 {
		customFields := make(map[string]interface{}, len(flat))
		for k, v := range flat {
			if k != "u_height" && k != "rack_position" {
				customFields[k] = v
			}
		}
		if len(customFields) > 0 {
			req.CustomFields = &customFields
		}
	}

	return req, nil
}

// resolveFace converts a CANI face string to a Nautobot RackFace pointer.
// Defaults to "front" when the face string is empty, since Nautobot requires
// a face value whenever a rack position is defined.
// Recognized values are "front" and "rear"; anything else defaults to front.
func resolveFace(face string) *nautobotapi.RackFace {
	rf := &nautobotapi.RackFace{}

	switch face {
	case "rear":
		_ = rf.FromFaceEnum(nautobotapi.FaceEnumRear)
	default:
		_ = rf.FromFaceEnum(nautobotapi.FaceEnumFront)
	}

	return rf
}
