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
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// LoadResult contains the results of a Load operation
type LoadResult struct {
	Created          []string       // Names of devices created
	Updated          []string       // Names of devices updated (merged)
	Skipped          []string       // Names of devices skipped (already exist, no --merge)
	Errors           []string       // Error messages
	Conflicts        []ConflictInfo // Detailed conflict information
	LocationsCreated []string       // Names of locations created
	LocationsSkipped []string       // Names of locations skipped (already exist)
	RacksCreated     []string       // Names of racks created
	IfacesCreated    int            // Number of interfaces created
	ModulesCreated   int            // Number of modules created
	FrusCreated      int            // Number of FRUs (inventory items) created
	CablesCreated    int            // Number of cables created
}

// ConflictInfo contains information about a device conflict
type ConflictInfo struct {
	DeviceName string
	ExistingID uuid.UUID
	LocalID    uuid.UUID
	Reason     string
	Diffs      []FieldDiff // fields that would change with --merge
}

// generateDeviceNames assigns a unique cani-prefixed name to every device in
// the inventory that has no name. The generated name uses the best available
// identifier: serial number, slug, model, or cani UUID.
func generateDeviceNames(inventory *devicetypes.Inventory) {
	for _, device := range inventory.Devices {
		if device == nil || device.Name != "" {
			continue
		}
		category := devicetypes.ClassifyForNautobot(device.HardwareType)
		if category != devicetypes.CategoryDevice {
			continue
		}

		var base string
		switch {
		case device.Serial != "":
			base = device.Serial
		case device.Slug != "":
			base = device.Slug
		case device.Model != "":
			base = strings.ToLower(strings.ReplaceAll(device.Model, " ", "-"))
		default:
			base = device.ID.String()[:8]
		}
		device.Name = "cani-" + base
	}
}

// disambiguateDeviceNames detects duplicate device names in the inventory
// and makes them unique by appending a suffix (serial, rack position, or index).
// Nautobot enforces name uniqueness per location+tenant, so duplicates must be
// resolved before export.
func disambiguateDeviceNames(inventory *devicetypes.Inventory) {
	// Collect device-category entries grouped by name.
	type entry struct {
		id     uuid.UUID
		device *devicetypes.CaniDeviceType
	}
	byName := make(map[string][]entry)
	for id, device := range inventory.Devices {
		if device == nil || device.Name == "" {
			continue
		}
		category := devicetypes.ClassifyForNautobot(device.HardwareType)
		if category != devicetypes.CategoryDevice {
			continue
		}
		byName[device.Name] = append(byName[device.Name], entry{id, device})
	}

	// Only process names that appear more than once.
	for _, entries := range byName {
		if len(entries) <= 1 {
			continue
		}
		for _, e := range entries {
			var suffix string
			switch {
			case e.device.Serial != "":
				suffix = " (" + e.device.Serial + ")"
			case e.device.RackPosition > 0:
				rackLabel := ""
				rackID := e.device.GetRackID(inventory)
				if rack, ok := inventory.Racks[rackID]; ok && rack != nil {
					rackLabel = rack.Name + " "
				}
				suffix = fmt.Sprintf(" (%sU%d)", rackLabel, e.device.RackPosition)
			default:
				suffix = fmt.Sprintf(" (%s)", e.id.String()[:8])
			}
			e.device.Name = e.device.Name + suffix
		}
	}
}

// setExternalID stores a remote UUID under the given provider key.
// It initialises the map when nil.
func setExternalID(m *map[string]uuid.UUID, provider string, id uuid.UUID) {
	if *m == nil {
		*m = make(map[string]uuid.UUID)
	}
	(*m)[provider] = id
}

// Load syncs the local inventory to Nautobot.
// The caller must initialise Client, Cache (with context) and Options before calling Load.
func (e *Exporter) Load(inventory *devicetypes.Inventory) error {
	ctx := context.Background()
	e.Cache.SetContext(ctx)

	result := &LoadResult{
		Created:          make([]string, 0),
		Updated:          make([]string, 0),
		Skipped:          make([]string, 0),
		Errors:           make([]string, 0),
		Conflicts:        make([]ConflictInfo, 0),
		LocationsCreated: make([]string, 0),
		LocationsSkipped: make([]string, 0),
		RacksCreated:     make([]string, 0),
	}

	// Create mapper with defaults from provider-global settings
	mapper := NewDeviceMapper(e.Cache, &MapperOpts{
		DefaultLocation: e.Options.DefaultLocation,
		DefaultRole:     e.Options.DefaultRole,
		DefaultStatus:   e.Options.DefaultStatus,
		Strict:          e.Options.Strict,
	})

	// Set inventory reference so mapper can resolve parent devices for rack positions
	mapper.SetInventory(inventory)

	// Phase 0: Create locations from inventory.Locations (top-down tree walk)
	createdLocationIDs, err := e.loadLocations(ctx, inventory, result)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("location phase error: %v", err))
	}
	_ = createdLocationIDs // available for rack/device location resolution

	// Track created device IDs for interface creation
	createdDeviceIDs := make(map[string]uuid.UUID) // deviceName -> Nautobot UUID

	// Track Nautobot IDs created in this session for same-name device detection
	createdThisSession := make(map[uuid.UUID]bool)

	// Track created rack IDs for device placement
	createdRackIDs := make(map[uuid.UUID]uuid.UUID) // cani rack ID -> Nautobot rack ID

	// Phase 1: Create racks from inventory.Racks (the primary rack storage)
	for rackID, rack := range inventory.Racks {
		if rack == nil {
			continue
		}

		// Resolve rack name: prefer Name, fall back to Model or Slug
		if rack.Name == "" {
			if rack.Model != "" {
				rack.Name = rack.Model
			} else if rack.Slug != "" {
				rack.Name = rack.Slug
			} else {
				continue
			}
		}

		// If we already have a Nautobot UUID from a previous export, use it
		if nid, ok := rack.ExternalIDs["nautobot"]; ok && nid != uuid.Nil {
			createdRackIDs[rackID] = nid
			clog.Skipped("Rack already exported: %s (nautobot:%s)", rack.Name, nid)
			result.Skipped = append(result.Skipped, rack.Name)
			continue
		}

		// Check if rack already exists in Nautobot by name
		existing, err := e.Cache.GetRackByName(rack.Name)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: rack lookup error: %v", rack.Name, err))
			continue
		}

		if existing != nil {
			clog.Skipped("Rack already exists: %s", rack.Name)
			result.Skipped = append(result.Skipped, rack.Name)
			createdRackIDs[rackID] = existing.ID
			setExternalID(&rack.ExternalIDs, "nautobot", existing.ID)
			continue
		}

		// Create rack in Nautobot
		nautobotRackID, err := e.createRackFromCaniRack(ctx, rack, inventory, mapper, result)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: rack create error: %v", rack.Name, err))
			continue
		}
		createdRackIDs[rackID] = nautobotRackID
		setExternalID(&rack.ExternalIDs, "nautobot", nautobotRackID)
	}

	// Phase 1b: Also check for racks in inventory.Devices (legacy/fallback)
	for id, device := range inventory.Devices {
		if device == nil || device.Name == "" {
			continue
		}

		category := devicetypes.ClassifyForNautobot(device.HardwareType)
		if category != devicetypes.CategoryRack {
			continue
		}

		// Check if rack already exists
		existing, err := e.Cache.GetRackByName(device.Name)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: rack lookup error: %v", device.Name, err))
			continue
		}

		if existing != nil {
			clog.Skipped("Rack already exists: %s", device.Name)
			result.Skipped = append(result.Skipped, device.Name)
			continue
		}

		// Create rack
		if err := e.createRack(ctx, device, mapper, result); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: rack create error: %v", device.Name, err))
		}
		_ = id // Prevent unused variable warning
	}

	// Pre-Phase 2: Ensure every device has a name, then disambiguate duplicates
	generateDeviceNames(inventory)
	disambiguateDeviceNames(inventory)

	// Pre-Phase 2b: Detect and resolve position swaps so that per-device
	// PATCH calls do not violate Nautobot's unique (rack, position, face)
	// constraint.
	if err := e.resolvePositionSwaps(ctx, inventory); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("position swap resolution: %v", err))
	}

	// Phase 2: Create devices
	for id, device := range inventory.Devices {
		if device == nil || device.Name == "" {
			continue
		}

		category := devicetypes.ClassifyForNautobot(device.HardwareType)
		if category != devicetypes.CategoryDevice {
			continue
		}

		// If this device already has a stored Nautobot UUID, use it directly.
		if nid, ok := device.ExternalIDs["nautobot"]; ok && nid != uuid.Nil {
			createdDeviceIDs[device.Name] = nid

			if e.Options.Merge {
				// Only update if there are actual field differences.
				var diffs []FieldDiff
				if fullDevice, err := e.fetchFullDeviceByID(ctx, nid); err == nil {
					diffs = compareDeviceFields(device, fullDevice, mapper)
				}
				if len(diffs) == 0 {
					result.Skipped = append(result.Skipped, device.Name)
					continue
				}
				if err := e.updateDevice(ctx, device, nid, mapper, result); err != nil {
					if errors.Is(err, ErrDeviceUnclassified) {
						clog.Skipped("  ~ %s: skipped (unclassified, no device type slug)", device.Name)
						result.Skipped = append(result.Skipped, device.Name)
					} else {
						result.Errors = append(result.Errors, fmt.Sprintf("%s: update error: %v", device.Name, err))
					}
				}
			} else {
				// Diff against the known Nautobot device
				var diffs []FieldDiff
				if fullDevice, err := e.fetchFullDeviceByID(ctx, nid); err == nil {
					diffs = compareDeviceFields(device, fullDevice, mapper)
				}
				if len(diffs) > 0 {
					result.Skipped = append(result.Skipped, device.Name)
					result.Conflicts = append(result.Conflicts, ConflictInfo{
						DeviceName: device.Name,
						ExistingID: nid,
						LocalID:    id,
						Reason:     "device already exists in Nautobot (use --merge to update)",
						Diffs:      diffs,
					})
				} else {
					result.Skipped = append(result.Skipped, device.Name)
				}
			}
			continue
		}

		// No stored Nautobot UUID — check by name for first-time export.
		existing, err := e.Cache.GetDeviceByName(device.Name)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: lookup error: %v", device.Name, err))
			continue
		}

		if existing != nil && !createdThisSession[existing.ID] {
			// Found an existing Nautobot device not created this session.
			createdDeviceIDs[device.Name] = existing.ID
			setExternalID(&device.ExternalIDs, "nautobot", existing.ID)

			if e.Options.Merge {
				// Only update if there are actual field differences.
				var diffs []FieldDiff
				if fullDevice, err := e.fetchFullDeviceByID(ctx, existing.ID); err == nil {
					diffs = compareDeviceFields(device, fullDevice, mapper)
				}
				if len(diffs) == 0 {
					result.Skipped = append(result.Skipped, device.Name)
					continue
				}
				if err := e.updateDevice(ctx, device, existing.ID, mapper, result); err != nil {
					if errors.Is(err, ErrDeviceUnclassified) {
						clog.Skipped("  ~ %s: skipped (unclassified, no device type slug)", device.Name)
						result.Skipped = append(result.Skipped, device.Name)
					} else {
						result.Errors = append(result.Errors, fmt.Sprintf("%s: update error: %v", device.Name, err))
					}
				}
			} else {
				var diffs []FieldDiff
				if fullDevice, err := e.fetchFullDeviceByID(ctx, existing.ID); err == nil {
					diffs = compareDeviceFields(device, fullDevice, mapper)
				}
				result.Skipped = append(result.Skipped, device.Name)
				result.Conflicts = append(result.Conflicts, ConflictInfo{
					DeviceName: device.Name,
					ExistingID: existing.ID,
					LocalID:    id,
					Reason:     "device already exists in Nautobot (use --merge to update)",
					Diffs:      diffs,
				})
			}
		} else {
			// Create new device
			nautobotID, err := e.createDeviceWithID(ctx, device, mapper, result)
			if err != nil {
				if errors.Is(err, ErrDeviceUnclassified) {
					clog.Skipped("  ~ %s: skipped (unclassified, no device type slug)", device.Name)
					result.Skipped = append(result.Skipped, device.Name)
				} else {
					result.Errors = append(result.Errors, fmt.Sprintf("%s: create error: %v", device.Name, err))
				}
			} else if nautobotID != uuid.Nil {
				createdDeviceIDs[device.Name] = nautobotID
				createdThisSession[nautobotID] = true
				setExternalID(&device.ExternalIDs, "nautobot", nautobotID)
			}
		}
	}

	// Phase 3: Create interfaces for devices (bulk batched)
	if err := e.loadInterfaces(ctx, inventory, createdDeviceIDs, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("interface phase error: %v", err))
	}

	// Phase 4: Create modules from inventory.Modules
	if err := e.loadModules(ctx, inventory, createdDeviceIDs, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("module phase error: %v", err))
	}

	// Phase 5: Create FRUs (inventory items) from inventory.Frus
	if err := e.loadFrus(ctx, inventory, createdDeviceIDs, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("fru phase error: %v", err))
	}

	// Phase 6: Create cables (connecting interfaces)
	if inventory.Cables != nil {
		for cableID, cable := range inventory.Cables {
			if cable == nil {
				continue
			}
			if err := e.createCaniCableType(ctx, cableID, cable, inventory, createdDeviceIDs, result); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("cable %s: create error: %v", cable.Label, err))
			}
		}
	}

	// Print field diffs for skipped devices before the summary
	e.printConflictDiffs(result)

	// Print summary
	e.printLoadSummary(result)

	// Return error if there were any errors
	if len(result.Errors) > 0 {
		return fmt.Errorf("encountered %d errors during sync", len(result.Errors))
	}

	return nil
}

// createDevice creates a new device in Nautobot
func (e *Exporter) createDevice(ctx context.Context, device *devicetypes.CaniDeviceType, mapper *DeviceMapper, result *LoadResult) error {
	req, err := mapper.MapToWritableDeviceRequest(device)
	if err != nil {
		return err
	}

	if e.Options.DryRun {
		clog.DryRun("Would create device: %s", device.Name)
		result.Created = append(result.Created, device.Name+" (dry-run)")
		return nil
	}

	resp, err := e.Client.DcimDevicesCreateWithResponse(ctx, &nautobotapi.DcimDevicesCreateParams{}, *req)
	if err != nil {
		return fmt.Errorf("API error: %w", err)
	}

	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		body := string(resp.Body)
		if resp.StatusCode() == http.StatusBadRequest && strings.Contains(body, "status") && strings.Contains(body, "Related object not found") {
			return fmt.Errorf("status '%s' does not support dcim.device content type in Nautobot — use a status like Active or Planned instead", device.Status)
		}
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), body)
	}

	clog.Created("Created device: %s", device.Name)
	result.Created = append(result.Created, device.Name)
	return nil
}

// updateDevice updates an existing device in Nautobot
func (e *Exporter) updateDevice(ctx context.Context, device *devicetypes.CaniDeviceType, existingID uuid.UUID, mapper *DeviceMapper, result *LoadResult) error {
	req, err := mapper.MapToPatchRequest(device, existingID)
	if err != nil {
		return err
	}

	if e.Options.DryRun {
		clog.DryRun("Would update device: %s (ID: %s)", device.Name, existingID)
		result.Updated = append(result.Updated, device.Name+" (dry-run)")
		return nil
	}

	resp, err := e.Client.DcimDevicesPartialUpdateWithResponse(ctx, existingID, &nautobotapi.DcimDevicesPartialUpdateParams{}, *req)
	if err != nil {
		return fmt.Errorf("API error: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), string(resp.Body))
	}

	clog.Changed("Updated device: %s", device.Name)
	result.Updated = append(result.Updated, device.Name)
	return nil
}

// createRack creates a new rack in Nautobot
func (e *Exporter) createRack(ctx context.Context, device *devicetypes.CaniDeviceType, mapper *DeviceMapper, result *LoadResult) error {
	req, err := mapper.MapToWritableRackRequest(device)
	if err != nil {
		return err
	}

	if e.Options.DryRun {
		clog.DryRun("Would create rack: %s", device.Name)
		result.RacksCreated = append(result.RacksCreated, device.Name+" (dry-run)")
		return nil
	}

	resp, err := e.Client.DcimRacksCreateWithResponse(ctx, &nautobotapi.DcimRacksCreateParams{}, *req)
	if err != nil {
		return fmt.Errorf("API error: %w", err)
	}

	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), string(resp.Body))
	}

	clog.Created("Created rack: %s", device.Name)
	result.RacksCreated = append(result.RacksCreated, device.Name)
	return nil
}

// createRackFromCaniRack creates a new rack in Nautobot from a CaniRackType
func (e *Exporter) createRackFromCaniRack(ctx context.Context, rack *devicetypes.CaniRackType, inventory *devicetypes.Inventory, mapper *DeviceMapper, result *LoadResult) (uuid.UUID, error) {
	// Resolve location - prefer rack-level UUID, fall back to provider default
	locationName := e.Options.DefaultLocation
	if rack.Location != uuid.Nil {
		if loc, ok := inventory.Locations[rack.Location]; ok && loc.Name != "" {
			locationName = loc.Name
		}
	}
	if locationName == "" {
		locationName = "Default"
	}

	location, err := e.Cache.GetLocation(locationName)
	if err != nil || location == nil {
		// Try to create location if allowed
		if e.Options.CreateLocations {
			location, err = e.Cache.CreateLocation(locationName)
			if err != nil {
				return uuid.Nil, fmt.Errorf("failed to create location: %w", err)
			}
		} else {
			return uuid.Nil, fmt.Errorf("location '%s' not found and create_locations is disabled", locationName)
		}
	}

	// Resolve status
	statusName := rack.Status
	if statusName == "" {
		statusName = e.Options.DefaultStatus
	}
	if statusName == "" {
		statusName = "Active"
	}
	status, err := e.Cache.GetStatus(statusName)
	if err != nil || status == nil {
		return uuid.Nil, fmt.Errorf("status '%s' not found", statusName)
	}

	// Build the request
	locationRef := makeStatusRef(location.ID)
	statusRef := makeStatusRef(status.ID)

	uHeight := rack.UHeight
	if uHeight == 0 {
		uHeight = 48 // default to 48U
	}

	req := nautobotapi.WritableRackRequest{
		Name:     rack.Name,
		Location: locationRef,
		Status:   statusRef,
		UHeight:  &uHeight,
	}

	// Map OuterWidth, OuterDepth, and OuterUnit if present.
	if rack.OuterWidth > 0 {
		ow := rack.OuterWidth
		req.OuterWidth = &ow
	}
	if rack.OuterDepth > 0 {
		od := rack.OuterDepth
		req.OuterDepth = &od
	}
	if rack.OuterWidth > 0 || rack.OuterDepth > 0 {
		unit := &nautobotapi.PatchedWritableRackRequestOuterUnit{}
		switch rack.OuterUnit {
		case "mm":
			_ = unit.FromOuterUnitEnum(nautobotapi.OuterUnitEnumMm)
		case "in":
			_ = unit.FromOuterUnitEnum(nautobotapi.OuterUnitEnumIn)
		default:
			_ = unit.FromOuterUnitEnum(nautobotapi.OuterUnitEnumMm) // default to mm
		}
		req.OuterUnit = unit
	}
	if rack.Comments != "" {
		req.Comments = &rack.Comments
	}

	// Map Width (rail-to-rail, enum: 10, 19, 21, 23)
	if rack.Width != "" {
		if w, err := strconv.Atoi(rack.Width); err == nil {
			we := nautobotapi.WidthEnum(w)
			req.Width = &we
		}
	}

	// Map RackType (enum: 2-post-frame, 4-post-cabinet, etc.)
	if rack.RackType != "" {
		rt := &nautobotapi.PatchedWritableRackRequestType{}
		if err := rt.FromRackTypeChoices(nautobotapi.RackTypeChoices(rack.RackType)); err == nil {
			req.Type = rt
		}
	}

	// Map scalar optional fields
	if rack.FacilityId != "" {
		req.FacilityId = &rack.FacilityId
	}
	if rack.Serial != "" {
		req.Serial = &rack.Serial
	}
	if rack.AssetTag != "" {
		req.AssetTag = &rack.AssetTag
	}
	if rack.DescUnits {
		req.DescUnits = &rack.DescUnits
	}

	if e.Options.DryRun {
		clog.DryRun("Would create rack: %s", rack.Name)
		result.RacksCreated = append(result.RacksCreated, rack.Name+" (dry-run)")
		return uuid.Nil, nil
	}

	resp, err := e.Client.DcimRacksCreateWithResponse(ctx, &nautobotapi.DcimRacksCreateParams{}, req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("API error: %w", err)
	}

	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		return uuid.Nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), string(resp.Body))
	}

	var nautobotID uuid.UUID
	if resp.JSON201 != nil && resp.JSON201.Id != nil {
		nautobotID = *resp.JSON201.Id
	}

	clog.Created("Created rack: %s (ID: %s)", rack.Name, nautobotID)
	result.RacksCreated = append(result.RacksCreated, rack.Name)
	return nautobotID, nil
}

// createDeviceWithID creates a new device in Nautobot and returns its Nautobot UUID
func (e *Exporter) createDeviceWithID(ctx context.Context, device *devicetypes.CaniDeviceType, mapper *DeviceMapper, result *LoadResult) (uuid.UUID, error) {
	req, err := mapper.MapToWritableDeviceRequest(device)
	if err != nil {
		return uuid.Nil, err
	}

	if e.Options.DryRun {
		clog.DryRun("Would create device: %s", device.Name)
		result.Created = append(result.Created, device.Name+" (dry-run)")
		return uuid.Nil, nil
	}

	resp, err := e.Client.DcimDevicesCreateWithResponse(ctx, &nautobotapi.DcimDevicesCreateParams{}, *req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("API error: %w", err)
	}

	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		body := string(resp.Body)
		if resp.StatusCode() == http.StatusBadRequest && strings.Contains(body, "status") && strings.Contains(body, "Related object not found") {
			return uuid.Nil, fmt.Errorf("status '%s' does not support dcim.device content type in Nautobot — use a status like Active or Planned instead", device.Status)
		}
		return uuid.Nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), body)
	}

	// Parse response to get the created device's ID
	var nautobotID uuid.UUID
	if resp.JSON201 != nil && resp.JSON201.Id != nil {
		nautobotID = *resp.JSON201.Id
	}

	clog.Created("Created device: %s (ID: %s)", device.Name, nautobotID)
	result.Created = append(result.Created, device.Name)
	return nautobotID, nil
}

// interfaceSpec describes an interface to create
type interfaceSpec struct {
	Name  string
	Type  string // 1000base-t, 10gbase-x-sfpp, etc.
	Speed int    // Speed in Kbps
}

// getDeviceInterfaceSpecs returns interface specifications based on device type.
// It first tries to use the device's instantiated interfaces (from device type library),
// falling back to hardcoded defaults if no interfaces are defined.
func getDeviceInterfaceSpecs(device *devicetypes.CaniDeviceType) []interfaceSpec {
	var specs []interfaceSpec

	// If device has instantiated interfaces from the device type library, use those
	if len(device.Interfaces) > 0 {
		for _, iface := range device.Interfaces {
			ifaceType := mapInterfaceType(string(iface.Type))
			speed := getSpeedForType(ifaceType)
			specs = append(specs, interfaceSpec{
				Name:  iface.Name,
				Type:  ifaceType,
				Speed: speed,
			})
		}
		return specs
	}

	// Fallback: Determine interface set based on hardware type and model
	switch devicetypes.Type(device.HardwareType) {
	case devicetypes.Blade, devicetypes.Node:
		// ProLiant servers typically have iLO + 4 Ethernet + optional IB
		specs = append(specs, interfaceSpec{Name: "iLO", Type: "1000base-t", Speed: 1000000})
		for i := 0; i < 4; i++ {
			specs = append(specs, interfaceSpec{Name: fmt.Sprintf("eth%d", i), Type: "1000base-t", Speed: 1000000})
		}
		// Check for InfiniBand adapters in metadata or model
		if containsInfiniband(device) {
			specs = append(specs, interfaceSpec{Name: "ib0", Type: "infiniband-hdr", Speed: 200000000})
			specs = append(specs, interfaceSpec{Name: "ib1", Type: "infiniband-hdr", Speed: 200000000})
		}

	case devicetypes.MgmtSwitch:
		// Management switches - use Aruba 2930F as template
		specs = append(specs, interfaceSpec{Name: "mgmt0", Type: "1000base-t", Speed: 1000000})
		for i := 1; i <= 48; i++ {
			specs = append(specs, interfaceSpec{Name: fmt.Sprintf("port%d", i), Type: "1000base-t", Speed: 1000000})
		}
		for i := 1; i <= 4; i++ {
			specs = append(specs, interfaceSpec{Name: fmt.Sprintf("sfp%d", i), Type: "10gbase-x-sfpp", Speed: 10000000})
		}

	case devicetypes.HSNSwitch:
		// InfiniBand NDR switches
		specs = append(specs, interfaceSpec{Name: "mgmt0", Type: "1000base-t", Speed: 1000000})
		for i := 1; i <= 64; i++ {
			specs = append(specs, interfaceSpec{Name: fmt.Sprintf("osfp%d", i), Type: "infiniband-ndr", Speed: 400000000})
		}

	case devicetypes.CabinetPDU:
		// PDUs just have management interface
		specs = append(specs, interfaceSpec{Name: "mgmt0", Type: "100base-tx", Speed: 100000})
	}

	return specs
}

// containsInfiniband checks if device has InfiniBand adapters
func containsInfiniband(device *devicetypes.CaniDeviceType) bool {
	// Check model name
	modelLower := strings.ToLower(device.Model)
	if strings.Contains(modelLower, "infiniband") || strings.Contains(modelLower, "ndr") ||
		strings.Contains(modelLower, "hdr") || strings.Contains(modelLower, "mcx") {
		return true
	}
	// Could also check children for NIC modules
	return false
}

// mapInterfaceType maps device type library interface types to Nautobot API types
func mapInterfaceType(ifaceType string) string {
	// Handle common mappings between devicetypes library and Nautobot API
	lower := strings.ToLower(ifaceType)
	switch {
	case strings.Contains(lower, "1000base-t"), strings.Contains(lower, "1gbase-t"):
		return "1000base-t"
	case strings.Contains(lower, "10gbase-x-sfpp"), strings.Contains(lower, "10gbase-x"):
		return "10gbase-x-sfpp"
	case strings.Contains(lower, "25gbase-x-sfp28"):
		return "25gbase-x-sfp28"
	case strings.Contains(lower, "40gbase-x-qsfpp"):
		return "40gbase-x-qsfpp"
	case strings.Contains(lower, "100gbase-x-qsfp28"):
		return "100gbase-x-qsfp28"
	case strings.Contains(lower, "200gbase-x-qsfp56"):
		return "200gbase-x-qsfp56"
	case strings.Contains(lower, "400gbase-x-osfp"), strings.Contains(lower, "400gbase-x-qsfpdd"):
		return "400gbase-x-osfp"
	case strings.Contains(lower, "infiniband-ndr"):
		return "infiniband-ndr"
	case strings.Contains(lower, "infiniband-hdr"):
		return "infiniband-hdr"
	case strings.Contains(lower, "100base-tx"):
		return "100base-tx"
	default:
		// Return as-is if no mapping needed
		if ifaceType != "" {
			return ifaceType
		}
		return "1000base-t" // Default fallback
	}
}

// getSpeedForType returns the speed in Kbps for a given interface type
func getSpeedForType(ifaceType string) int {
	switch ifaceType {
	case "100base-tx":
		return 100000
	case "1000base-t":
		return 1000000
	case "10gbase-x-sfpp":
		return 10000000
	case "25gbase-x-sfp28":
		return 25000000
	case "40gbase-x-qsfpp":
		return 40000000
	case "100gbase-x-qsfp28":
		return 100000000
	case "200gbase-x-qsfp56", "infiniband-hdr":
		return 200000000
	case "400gbase-x-osfp", "400gbase-x-qsfpdd", "infiniband-ndr":
		return 400000000
	default:
		return 1000000 // Default 1Gbps
	}
}

// createInterface creates a single interface on a device
func (e *Exporter) createInterface(ctx context.Context, deviceID uuid.UUID, iface interfaceSpec, result *LoadResult) error {
	if e.Options.DryRun {
		clog.DryRun("Would create interface: %s", iface.Name)
		result.IfacesCreated++
		return nil
	}

	// Build device reference with proper union type for ID
	var deviceIDUnion nautobotapi.BulkWritableCableRequestStatusId
	if err := deviceIDUnion.FromBulkWritableCableRequestStatusId0(deviceID); err != nil {
		return fmt.Errorf("failed to create device ID: %w", err)
	}
	deviceRef := &nautobotapi.BulkWritableCircuitRequestTenant{
		Id: &deviceIDUnion,
	}

	// Build status reference - get "Active" status
	statusItem, err := e.Cache.GetStatus("Active")
	if err != nil {
		return fmt.Errorf("failed to get Active status: %w", err)
	}
	var statusIDUnion nautobotapi.BulkWritableCableRequestStatusId
	if err := statusIDUnion.FromBulkWritableCableRequestStatusId0(statusItem.ID); err != nil {
		return fmt.Errorf("failed to create status ID: %w", err)
	}
	status := nautobotapi.BulkWritableCableRequestStatus{
		Id: &statusIDUnion,
	}

	// Build interface request
	ifaceType := nautobotapi.InterfaceTypeChoices(iface.Type)
	req := nautobotapi.WritableInterfaceRequest{
		Device: deviceRef,
		Name:   iface.Name,
		Type:   ifaceType,
		Status: status,
	}

	resp, err := e.Client.DcimInterfacesCreateWithResponse(ctx, &nautobotapi.DcimInterfacesCreateParams{}, req)
	if err != nil {
		return fmt.Errorf("API error: %w", err)
	}

	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), string(resp.Body))
	}

	// Cache the newly created interface for cable creation
	if resp.JSON201 != nil && resp.JSON201.Id != nil {
		cachedItem := &CachedItem{
			ID:      uuid.UUID(*resp.JSON201.Id),
			Name:    iface.Name,
			Display: iface.Name,
		}
		e.Cache.CacheInterface(deviceID, iface.Name, cachedItem)
	}

	result.IfacesCreated++
	return nil
}

// updateInterface updates an existing interface in Nautobot
func (e *Exporter) updateInterface(ctx context.Context, interfaceID uuid.UUID, deviceID uuid.UUID, iface interfaceSpec, result *LoadResult) error {
	if e.Options.DryRun {
		clog.DryRun("Would update interface: %s", iface.Name)
		return nil
	}

	// Build status reference - get "Active" status
	statusItem, err := e.Cache.GetStatus("Active")
	if err != nil {
		return fmt.Errorf("failed to get Active status: %w", err)
	}
	var statusIDUnion nautobotapi.BulkWritableCableRequestStatusId
	if err := statusIDUnion.FromBulkWritableCableRequestStatusId0(statusItem.ID); err != nil {
		return fmt.Errorf("failed to create status ID: %w", err)
	}
	status := nautobotapi.BulkWritableCableRequestStatus{
		Id: &statusIDUnion,
	}

	// Build device reference (required by Nautobot API)
	var deviceIDUnion nautobotapi.BulkWritableCableRequestStatusId
	if err := deviceIDUnion.FromBulkWritableCableRequestStatusId0(deviceID); err != nil {
		return fmt.Errorf("failed to create device ID: %w", err)
	}
	device := &nautobotapi.BulkWritableCircuitRequestTenant{
		Id: &deviceIDUnion,
	}

	// Build patch request - update type, status, and device
	ifaceType := nautobotapi.InterfaceTypeChoices(iface.Type)
	req := nautobotapi.PatchedWritableInterfaceRequest{
		Device: device,
		Type:   &ifaceType,
		Status: &status,
	}

	resp, err := e.Client.DcimInterfacesPartialUpdateWithResponse(ctx, interfaceID, &nautobotapi.DcimInterfacesPartialUpdateParams{}, req)
	if err != nil {
		return fmt.Errorf("API error: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), string(resp.Body))
	}

	return nil
}

// createCaniCableType creates a cable between two interfaces in Nautobot
func (e *Exporter) createCaniCableType(ctx context.Context, cableID uuid.UUID, cable *devicetypes.CaniCableType, inventory *devicetypes.Inventory, deviceIDMap map[string]uuid.UUID, result *LoadResult) error {
	// Skip cables that don't have both terminations
	if cable.TerminationADevice == uuid.Nil || cable.TerminationBDevice == uuid.Nil {
		clog.Detail("INFO: Cable %s has incomplete terminations (status=%s), skipping", cable.Label, cable.Status)
		return nil
	}

	// Look up device names from inventory
	deviceA := inventory.Devices[cable.TerminationADevice]
	deviceB := inventory.Devices[cable.TerminationBDevice]
	if deviceA == nil || deviceB == nil {
		return fmt.Errorf("cable endpoints reference unknown devices (A=%s, B=%s)", cable.TerminationADevice, cable.TerminationBDevice)
	}

	// Get Nautobot device IDs for both ends
	nautobotDeviceAID, okA := deviceIDMap[deviceA.Name]
	nautobotDeviceBID, okB := deviceIDMap[deviceB.Name]
	if !okA || !okB {
		// Try to look up devices by name from Nautobot cache
		if !okA {
			devA, err := e.Cache.GetDeviceByName(deviceA.Name)
			if err != nil || devA == nil {
				return fmt.Errorf("cannot find Nautobot device ID for %s", deviceA.Name)
			}
			nautobotDeviceAID = devA.ID
		}
		if !okB {
			devB, err := e.Cache.GetDeviceByName(deviceB.Name)
			if err != nil || devB == nil {
				return fmt.Errorf("cannot find Nautobot device ID for %s", deviceB.Name)
			}
			nautobotDeviceBID = devB.ID
		}
	}

	// Look up interface IDs for both terminations using fuzzy matching
	// This handles naming variations between cani inventory and Nautobot
	ifaceA, err := e.Cache.GetInterfaceByDeviceAndNameFuzzy(nautobotDeviceAID, cable.TerminationAPort)
	if err != nil || ifaceA == nil {
		return fmt.Errorf("cannot find interface %s on device %s: %v", cable.TerminationAPort, deviceA.Name, err)
	}

	ifaceB, err := e.Cache.GetInterfaceByDeviceAndNameFuzzy(nautobotDeviceBID, cable.TerminationBPort)
	if err != nil || ifaceB == nil {
		return fmt.Errorf("cannot find interface %s on device %s: %v", cable.TerminationBPort, deviceB.Name, err)
	}

	// Check if cable already exists between these interfaces
	existingCable, err := e.Cache.GetCableByTerminations(ifaceA.ID, ifaceB.ID)
	if err != nil {
		clog.Warn("Warning: failed to check for existing cable between %s:%s and %s:%s: %v",
			deviceA.Name, cable.TerminationAPort, deviceB.Name, cable.TerminationBPort, err)
		// Continue to try creating it anyway
	}
	if existingCable != nil {
		// Cable already exists between these exact interfaces
		clog.Skipped("Cable already exists: %s (%s:%s <-> %s:%s) - skipping",
			cable.Label, deviceA.Name, cable.TerminationAPort, deviceB.Name, cable.TerminationBPort)
		return nil
	}

	// Also check if either interface already has ANY cable attached
	// (to prevent "must make a unique set" errors)
	if ifaceA.CableID != uuid.Nil {
		clog.Skipped("Interface %s:%s already has a cable attached - skipping cable creation",
			deviceA.Name, cable.TerminationAPort)
		return nil
	}
	if ifaceB.CableID != uuid.Nil {
		clog.Skipped("Interface %s:%s already has a cable attached - skipping cable creation",
			deviceB.Name, cable.TerminationBPort)
		return nil
	}

	if e.Options.DryRun {
		clog.DryRun("Would create cable: %s (%s:%s <-> %s:%s)",
			cable.Label, deviceA.Name, cable.TerminationAPort, deviceB.Name, cable.TerminationBPort)
		result.CablesCreated++
		return nil
	}

	// Build cable status reference
	statusName := string(devicetypes.StatusConnected)
	if strings.EqualFold(cable.Status, "planned") {
		statusName = string(devicetypes.StatusPlanned)
	}
	statusItem, err := e.Cache.GetStatus(statusName)
	if err != nil {
		return fmt.Errorf("failed to get %s status: %w", statusName, err)
	}
	var statusIDUnion nautobotapi.BulkWritableCableRequestStatusId
	if err := statusIDUnion.FromBulkWritableCableRequestStatusId0(statusItem.ID); err != nil {
		return fmt.Errorf("failed to create status ID: %w", err)
	}
	status := nautobotapi.BulkWritableCableRequestStatus{
		Id: &statusIDUnion,
	}

	// Build cable request — prefer explicit types from fixture, default to dcim.interface
	terminationAType := "dcim.interface"
	if cable.TerminationAType != "" {
		terminationAType = cable.TerminationAType
	}
	terminationBType := "dcim.interface"
	if cable.TerminationBType != "" {
		terminationBType = cable.TerminationBType
	}

	// Determine cable type: prefer explicit CableType field, then derive from
	// CableCategory/ConnectorType, then fall back to slug heuristic.
	cableType := resolveCableType(cable)

	// Convert length to int if present
	var lengthInt *int
	if cable.Length != nil {
		l := int(*cable.Length)
		lengthInt = &l
	}

	req := nautobotapi.WritableCableRequest{
		TerminationAId:   (openapi_types.UUID)(ifaceA.ID),
		TerminationAType: terminationAType,
		TerminationBId:   (openapi_types.UUID)(ifaceB.ID),
		TerminationBType: terminationBType,
		Status:           status,
		Label:            &cable.Label,
		Type:             cableType,
		Length:           lengthInt,
	}

	// Map cable Color if present (RGB hex, e.g. "00ff00").
	// Accept both hex codes and common named colors.
	if cable.Color != "" {
		hex := colorNameToHex(cable.Color)
		req.Color = &hex
	}

	// Set length unit if provided
	if cable.LengthUnit != "" {
		var unit *nautobotapi.PatchedWritableCableRequestLengthUnit
		switch cable.LengthUnit {
		case "m":
			u := &nautobotapi.PatchedWritableCableRequestLengthUnit{}
			if err := u.FromLengthUnitEnum(nautobotapi.LengthUnitEnumM); err == nil {
				unit = u
			}
		case "cm":
			u := &nautobotapi.PatchedWritableCableRequestLengthUnit{}
			if err := u.FromLengthUnitEnum(nautobotapi.LengthUnitEnumCm); err == nil {
				unit = u
			}
		case "ft":
			u := &nautobotapi.PatchedWritableCableRequestLengthUnit{}
			if err := u.FromLengthUnitEnum(nautobotapi.LengthUnitEnumFt); err == nil {
				unit = u
			}
		case "in":
			u := &nautobotapi.PatchedWritableCableRequestLengthUnit{}
			if err := u.FromLengthUnitEnum(nautobotapi.LengthUnitEnumIn); err == nil {
				unit = u
			}
		}
		req.LengthUnit = unit
	}

	resp, err := e.Client.DcimCablesCreateWithResponse(ctx, &nautobotapi.DcimCablesCreateParams{}, req)
	if err != nil {
		return fmt.Errorf("API error: %w", err)
	}

	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), string(resp.Body))
	}

	clog.Created("Created cable: %s (%s:%s <-> %s:%s)",
		cable.Label, deviceA.Name, cable.TerminationAPort, deviceB.Name, cable.TerminationBPort)
	result.CablesCreated++
	return nil
}

// printConflictDiffs prints per-device field diffs for devices that already
// exist in Nautobot. This gives the user visibility into what --merge would
// change before the summary is printed.
func (e *Exporter) printConflictDiffs(result *LoadResult) {
	hasAnyDiffs := false
	for _, c := range result.Conflicts {
		if len(c.Diffs) > 0 {
			hasAnyDiffs = true
			break
		}
	}
	if !hasAnyDiffs {
		return
	}

	clog.Header("\n=== Devices with pending changes (use --merge to apply) ===")
	for _, c := range result.Conflicts {
		printDeviceDiffs(c.DeviceName, c.Diffs)
	}
}

// printLoadSummary prints a summary of the load operation
func (e *Exporter) printLoadSummary(result *LoadResult) {
	clog.Header("\n=== Nautobot Sync Summary ===")

	if len(result.LocationsCreated) > 0 {
		clog.Created("Locations created: %d", len(result.LocationsCreated))
		for _, name := range result.LocationsCreated {
			clog.SummaryCreated("%s", name)
		}
	}

	if len(result.LocationsSkipped) > 0 {
		clog.Ok("Locations skipped (already exist): %d", len(result.LocationsSkipped))
	}

	if len(result.RacksCreated) > 0 {
		clog.Created("Racks created: %d", len(result.RacksCreated))
		for _, name := range result.RacksCreated {
			clog.SummaryCreated("%s", name)
		}
	}

	if len(result.Created) > 0 {
		clog.Created("Devices created: %d", len(result.Created))
		for _, name := range result.Created {
			clog.SummaryCreated("%s", name)
		}
	}

	if len(result.Updated) > 0 {
		clog.Changed("Devices updated: %d", len(result.Updated))
		for _, name := range result.Updated {
			clog.SummaryChanged("%s", name)
		}
	}

	if result.IfacesCreated > 0 {
		clog.Created("Interfaces created: %d", result.IfacesCreated)
	}

	if result.ModulesCreated > 0 {
		clog.Created("Modules created: %d", result.ModulesCreated)
	}

	if result.FrusCreated > 0 {
		clog.Created("Inventory items (FRUs) created: %d", result.FrusCreated)
	}

	if result.CablesCreated > 0 {
		clog.Created("Cables created: %d", result.CablesCreated)
	}

	if len(result.Skipped) > 0 {
		clog.Skipped("Skipped (conflicts): %d devices", len(result.Skipped))
		for _, conflict := range result.Conflicts {
			if len(conflict.Diffs) > 0 {
				clog.SummaryChanged("%s: %d field(s) would change (use --merge to update)", conflict.DeviceName, len(conflict.Diffs))
			} else {
				clog.SummarySkipped("%s: up to date", conflict.DeviceName)
			}
		}
	}

	if len(result.Errors) > 0 {
		clog.Error("Errors: %d", len(result.Errors))
		for _, errMsg := range result.Errors {
			clog.SummaryError("%s", errMsg)
		}
	}

	total := len(result.LocationsCreated) + len(result.RacksCreated) + len(result.Created) + len(result.Updated) + len(result.Skipped) + len(result.Errors)
	clog.Info("\nTotal processed: %d objects (locations=%d, racks=%d, devices=%d, interfaces=%d, modules=%d, frus=%d, cables=%d)",
		total, len(result.LocationsCreated), len(result.RacksCreated),
		len(result.Created)+len(result.Updated), result.IfacesCreated,
		result.ModulesCreated, result.FrusCreated, result.CablesCreated)
}

// cableTypeMap maps CableCategory strings to their best-fit Nautobot CableTypeChoices.
var cableTypeMap = map[string]nautobotapi.CableTypeChoices{
	// Copper / Ethernet
	"cat3":  nautobotapi.CableTypeChoicesCat3,
	"cat5":  nautobotapi.CableTypeChoicesCat5,
	"cat5e": nautobotapi.CableTypeChoicesCat5e,
	"cat6":  nautobotapi.CableTypeChoicesCat6,
	"cat6a": nautobotapi.CableTypeChoicesCat6a,
	"cat7":  nautobotapi.CableTypeChoicesCat7,
	"cat7a": nautobotapi.CableTypeChoicesCat7a,
	"cat8":  nautobotapi.CableTypeChoicesCat8,
	// Coaxial
	"coaxial": nautobotapi.CableTypeChoicesCoaxial,
	// Direct-attach copper
	"dac-passive": nautobotapi.CableTypeChoicesDacPassive,
	"dac-active":  nautobotapi.CableTypeChoicesDacActive,
	// Active optical cable
	"aoc": nautobotapi.CableTypeChoicesAoc,
	// Fiber – multimode
	"mmf":     nautobotapi.CableTypeChoicesMmf,
	"mmf-om1": nautobotapi.CableTypeChoicesMmfOm1,
	"mmf-om2": nautobotapi.CableTypeChoicesMmfOm2,
	"mmf-om3": nautobotapi.CableTypeChoicesMmfOm3,
	"mmf-om4": nautobotapi.CableTypeChoicesMmfOm4,
	"mmf-om5": nautobotapi.CableTypeChoicesMmfOm5,
	// Fiber – singlemode
	"smf":     nautobotapi.CableTypeChoicesSmf,
	"smf-os1": nautobotapi.CableTypeChoicesSmfOs1,
	"smf-os2": nautobotapi.CableTypeChoicesSmfOs2,
	// MRJ21
	"mrj21-trunk": nautobotapi.CableTypeChoicesMrj21Trunk,
	// Power & other
	"power": nautobotapi.CableTypeChoicesPower,
	"other": nautobotapi.CableTypeChoicesOther,
}

// connectorToCableType maps common ConnectorType values to a reasonable
// CableTypeChoices default when CableType and CableCategory are empty.
var connectorToCableType = map[string]nautobotapi.CableTypeChoices{
	"rj45":   nautobotapi.CableTypeChoicesCat6,
	"rj-45":  nautobotapi.CableTypeChoicesCat6,
	"lc":     nautobotapi.CableTypeChoicesSmfOs2,
	"sc":     nautobotapi.CableTypeChoicesMmfOm4,
	"mpo":    nautobotapi.CableTypeChoicesMmfOm4,
	"mtp":    nautobotapi.CableTypeChoicesMmfOm4,
	"sfp":    nautobotapi.CableTypeChoicesDacPassive,
	"sfp+":   nautobotapi.CableTypeChoicesDacPassive,
	"sfp28":  nautobotapi.CableTypeChoicesDacPassive,
	"qsfp":   nautobotapi.CableTypeChoicesDacPassive,
	"qsfp+":  nautobotapi.CableTypeChoicesDacPassive,
	"qsfp28": nautobotapi.CableTypeChoicesDacPassive,
}

// resolveCableType determines the Nautobot CableTypeChoices for a cable.
// Priority: explicit CableType field → CableCategory lookup → ConnectorType
// heuristic → slug-based fallback.
func resolveCableType(cable *devicetypes.CaniCableType) *nautobotapi.PatchedWritableCableRequestType {
	// 1. Explicit CableType field (already a Nautobot enum string)
	if cable.CableType != "" {
		if choice, ok := cableTypeMap[cable.CableType]; ok {
			t := &nautobotapi.PatchedWritableCableRequestType{}
			if err := t.FromCableTypeChoices(choice); err == nil {
				return t
			}
		}
	}

	// 2. CableCategory lookup (e.g. "cat6a", "mmf-om4", "dac-passive")
	if cable.CableCategory != "" {
		if choice, ok := cableTypeMap[strings.ToLower(cable.CableCategory)]; ok {
			t := &nautobotapi.PatchedWritableCableRequestType{}
			if err := t.FromCableTypeChoices(choice); err == nil {
				return t
			}
		}
	}

	// 3. ConnectorType heuristic (e.g. "rj45" → cat6, "lc" → smf-os2)
	if cable.ConnectorType != "" {
		if choice, ok := connectorToCableType[strings.ToLower(cable.ConnectorType)]; ok {
			t := &nautobotapi.PatchedWritableCableRequestType{}
			if err := t.FromCableTypeChoices(choice); err == nil {
				return t
			}
		}
	}

	// 4. Legacy slug-based fallback
	slug := strings.ToLower(cable.Slug)
	switch {
	case strings.Contains(slug, "cat"):
		t := &nautobotapi.PatchedWritableCableRequestType{}
		if err := t.FromCableTypeChoices(nautobotapi.CableTypeChoicesCat5e); err == nil {
			return t
		}
	case strings.Contains(slug, "dac"):
		t := &nautobotapi.PatchedWritableCableRequestType{}
		if err := t.FromCableTypeChoices(nautobotapi.CableTypeChoicesDacPassive); err == nil {
			return t
		}
	case strings.Contains(slug, "aoc"):
		t := &nautobotapi.PatchedWritableCableRequestType{}
		if err := t.FromCableTypeChoices(nautobotapi.CableTypeChoicesAoc); err == nil {
			return t
		}
	case strings.Contains(slug, "fiber") || strings.Contains(slug, "mmf"):
		t := &nautobotapi.PatchedWritableCableRequestType{}
		if err := t.FromCableTypeChoices(nautobotapi.CableTypeChoicesMmfOm4); err == nil {
			return t
		}
	}

	return nil
}

// colorNameToHex converts common color names to 6-digit hex codes expected by
// Nautobot. If the value already looks like a hex code it is returned as-is.
func colorNameToHex(name string) string {
	namedColors := map[string]string{
		"black":  "000000",
		"white":  "ffffff",
		"red":    "ff0000",
		"green":  "00ff00",
		"blue":   "0000ff",
		"yellow": "ffff00",
		"orange": "ff8800",
		"purple": "800080",
		"grey":   "808080",
		"gray":   "808080",
		"brown":  "795548",
		"pink":   "e91e63",
		"teal":   "009688",
		"cyan":   "00bcd4",
	}

	lower := strings.ToLower(strings.TrimSpace(name))
	// Strip leading '#' if present
	lower = strings.TrimPrefix(lower, "#")

	if hex, ok := namedColors[lower]; ok {
		return hex
	}
	// Assume it's already a hex code
	return lower
}
