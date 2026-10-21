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
package export

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// remoteDeviceBay is the part of a Nautobot device bay the exporter needs.
type remoteDeviceBay struct {
	id        uuid.UUID
	parent    uuid.UUID
	name      string
	installed uuid.UUID
}

// DeviceBayCounts tallies the device-bay phase of an export.
type DeviceBayCounts struct {
	DeviceBaysCleared   int // obsolete assignments removed from managed devices
	DeviceBaysInstalled int // child devices installed into parent device bays
	DeviceBaysSkipped   int // child devices already installed in their bay
}

// loadDeviceBays installs each child device into its parent's device bay.
// Nautobot creates bays from the parent device type's bay templates; when the
// type has none, the named bay is created on the parent device. Devices that
// were not created or found in Nautobot are skipped because their own phase
// already reported the failure.
func (e *Exporter) loadDeviceBays(ctx context.Context, inventory *devicetypes.Inventory, deviceIDs map[string]uuid.UUID, result *LoadResult) {
	for _, child := range inventory.Devices {
		if child == nil || child.ParentBay == "" {
			continue
		}
		parent := inventory.Devices[child.Parent]
		if parent == nil {
			continue
		}
		parentID, childID := deviceIDs[parent.Name], deviceIDs[child.Name]
		if parentID == uuid.Nil || childID == uuid.Nil {
			continue
		}
		if err := e.installDeviceInBay(ctx, parentID, child.ParentBay, childID, result); err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("%s: install in bay %q of %s: %v", child.Name, child.ParentBay, parent.Name, err))
			continue
		}
		clog.Detail("[device-bay] %s -> %s %q", child.Name, parent.Name, child.ParentBay)
	}
}

// installDeviceInBay sets the bay's installed device, skipping bays that
// already hold it and refusing bays that hold a different device.
func (e *Exporter) installDeviceInBay(ctx context.Context, parentID uuid.UUID, bayName string, childID uuid.UUID, result *LoadResult) error {
	if e.Options.DryRun {
		clog.DryRun("Would install device %s in bay %q of %s", childID, bayName, parentID)
		result.DeviceBaysInstalled++
		return nil
	}
	bay, err := e.getOrCreateDeviceBay(ctx, parentID, bayName)
	if err != nil {
		return err
	}
	switch bay.installed {
	case childID:
		result.DeviceBaysSkipped++
		return nil
	case uuid.Nil:
	default:
		return fmt.Errorf("bay already holds device %s", bay.installed)
	}
	var patch nautobotapi.PatchedDeviceBayRequest
	if err := setRefID(&patch.InstalledDevice, childID); err != nil {
		return fmt.Errorf("set installed device reference: %w", err)
	}
	resp, err := e.Client.DcimDeviceBaysPartialUpdateWithResponse(ctx, bay.id,
		&nautobotapi.DcimDeviceBaysPartialUpdateParams{}, patch)
	if err != nil {
		return fmt.Errorf("device bay update: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("device bay update: status %d: %s", resp.StatusCode(), string(resp.Body))
	}
	result.DeviceBaysInstalled++
	return nil
}

// getOrCreateDeviceBay returns the named bay on the parent device, creating
// it when the parent's device type defines no matching bay template.
func (e *Exporter) getOrCreateDeviceBay(ctx context.Context, parentID uuid.UUID, bayName string) (remoteDeviceBay, error) {
	names, devices := []string{bayName}, []string{parentID.String()}
	bays, err := e.listDeviceBays(ctx,
		nautobotapi.DcimDeviceBaysListParams{Name: &names, Device: &devices})
	if err != nil {
		return remoteDeviceBay{}, fmt.Errorf("device bay lookup: %w", err)
	}
	if len(bays) > 0 {
		return bays[0], nil
	}
	return e.createDeviceBay(ctx, parentID, bayName)
}

// createDeviceBay creates an empty named bay on the parent device.
func (e *Exporter) createDeviceBay(ctx context.Context, parentID uuid.UUID, bayName string) (remoteDeviceBay, error) {
	request := nautobotapi.DeviceBayRequest{Name: bayName}
	if err := setRefID(&request.Device, parentID); err != nil {
		return remoteDeviceBay{}, fmt.Errorf("set device bay parent reference: %w", err)
	}
	resp, err := e.Client.DcimDeviceBaysCreateWithResponse(ctx, &nautobotapi.DcimDeviceBaysCreateParams{}, request)
	if err != nil {
		return remoteDeviceBay{}, fmt.Errorf("device bay create: %w", err)
	}
	if resp.StatusCode() != http.StatusCreated || resp.JSON201 == nil {
		return remoteDeviceBay{}, fmt.Errorf("device bay create: status %d: %s", resp.StatusCode(), string(resp.Body))
	}
	if err := validateDeviceBay(*resp.JSON201); err != nil {
		return remoteDeviceBay{}, fmt.Errorf("device bay create: %w", err)
	}
	return toRemoteDeviceBay(*resp.JSON201), nil
}

func validateDeviceBay(bay nautobotapi.DeviceBay) error {
	if toUUID(bay.Id) == uuid.Nil || refID(bay.Device.Id) == uuid.Nil || bay.Name == "" {
		return fmt.Errorf("invalid bay %q", bay.Name)
	}
	return nil
}

// toRemoteDeviceBay extracts the bay and installed-device IDs.
func toRemoteDeviceBay(bay nautobotapi.DeviceBay) remoteDeviceBay {
	out := remoteDeviceBay{id: toUUID(bay.Id), parent: refID(bay.Device.Id), name: bay.Name}
	if bay.InstalledDevice != nil && bay.InstalledDevice.Id != nil {
		out.installed = refID(bay.InstalledDevice.Id)
	}
	return out
}

// printDeviceBaySummary reports device-bay install/skip counts.
func printDeviceBaySummary(result *LoadResult) {
	if result.DeviceBaysCleared > 0 {
		clog.Changed("Cleared obsolete device bay assignments: %d", result.DeviceBaysCleared)
	}
	if result.DeviceBaysInstalled > 0 {
		clog.Created("Installed devices in bays: %d", result.DeviceBaysInstalled)
	}
	if result.DeviceBaysSkipped > 0 {
		clog.Skipped("Skipped device bays (already installed): %d", result.DeviceBaysSkipped)
	}
}
