/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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
package add

import (
	"fmt"
	"log"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/util/validate"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// addAnyDevice adds device(s) using the resolved device type.
func addAnyDevice(cmd *cli.Command, args []string, device *devicetypes.CaniDeviceType, qty int) error {
	parentArg, _ := cmd.Flags().GetString("parent")
	statusArg, _ := cmd.Flags().GetString("status")
	serialArg, _ := cmd.Flags().GetString("serial")
	auto, _ := cmd.Flags().GetBool("auto")

	names, err := resolveNamesFromFlags(cmd, qty)
	if err != nil {
		return err
	}

	inventory, err := loadInventoryForAdd(cmd, args)
	if err != nil {
		return err
	}

	locationID := inventory.EnsureLocation()
	inventory.AssignRacksToLocation(locationID)

	// When --auto is set, try to stage existing imported devices
	// that match the requested slug instead of creating new ones.
	if auto && device.Slug != "" {
		handled, serr := attemptAutoStage(inventory, device, qty)
		if serr != nil {
			return serr
		}
		if handled {
			return nil
		}
	}

	statusArg, err = normalizeStatus(statusArg, inventory)
	if err != nil {
		return err
	}

	tags, _ := cmd.Flags().GetStringArray("tag")
	attrs := deviceAttrs{
		parentArg: parentArg,
		statusArg: statusArg,
		serialArg: serialArg,
		tags:      tags,
		provMeta:  collectProviderMetadata(cmd),
	}

	devicesToAdd := buildDevicesToAdd(device, names, attrs, qty)

	return saveAndLogDevices(inventory, devicesToAdd, qty)
}

// normalizeStatus validates a non-empty status against the inventory and
// returns the normalized value. An empty status is returned unchanged.
func normalizeStatus(statusArg string, inventory *devicetypes.Inventory) (string, error) {
	if statusArg == "" {
		return "", nil
	}
	return validate.StatusWithInventory(statusArg, inventory)
}

// deviceAttrs carries the per-device attributes applied to each new device.
type deviceAttrs struct {
	parentArg string
	statusArg string
	serialArg string
	tags      []string
	provMeta  map[string]string
}

// buildDevicesToAdd constructs the device(s) to add, expanding any child
// devices defined by device-bay defaults.
func buildDevicesToAdd(
	device *devicetypes.CaniDeviceType,
	names []string,
	attrs deviceAttrs,
	qty int,
) map[uuid.UUID]*devicetypes.CaniDeviceType {
	devicesToAdd := make(map[uuid.UUID]*devicetypes.CaniDeviceType)
	for i := range qty {
		var name string
		setName := names != nil
		if setName {
			name = names[i]
		}
		d := buildOneDevice(device, name, setName, attrs)
		devicesToAdd[d.ID] = d

		// Expand child devices from device-bay defaults.
		for cid, child := range devicetypes.ExpandChildren(d) {
			devicesToAdd[cid] = child
		}
	}
	return devicesToAdd
}

// buildOneDevice constructs a single device from the resolved type, applying
// the supplied parent, name, status, serial, tags, and provider metadata.
func buildOneDevice(
	device *devicetypes.CaniDeviceType,
	name string,
	setName bool,
	attrs deviceAttrs,
) *devicetypes.CaniDeviceType {
	d := *device
	d.ID = uuid.New()
	applyParentArg(&d, attrs.parentArg)
	if setName {
		d.Name = name
	}
	if attrs.statusArg != "" {
		d.Status = attrs.statusArg
	}
	if attrs.serialArg != "" {
		d.Serial = attrs.serialArg
	}
	applyTagsToDevice(&d, attrs.tags)
	applyProviderMetadataToDevice(&d, attrs.provMeta)
	return &d
}

// applyParentArg assigns a parent UUID to the device when parentArg is a
// non-empty, parseable, non-nil UUID string.
func applyParentArg(d *devicetypes.CaniDeviceType, parentArg string) {
	if parentArg == uuid.Nil.String() || parentArg == "" {
		return
	}
	if pid, perr := uuid.Parse(parentArg); perr == nil {
		d.Parent = pid
	}
}

// saveAndLogDevices adds the devices to the inventory, persists it, and logs
// the result.
func saveAndLogDevices(
	inventory *devicetypes.Inventory,
	devicesToAdd map[uuid.UUID]*devicetypes.CaniDeviceType,
	qty int,
) error {
	if err := inventory.AddDevices(devicesToAdd); err != nil {
		return fmt.Errorf("failed to add devices: %w", err)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	for _, d := range devicesToAdd {
		log.Printf("Added device %s (%s)", d.ID, d.Name)
	}
	log.Printf("%d device(s) added", qty)
	return nil
}
