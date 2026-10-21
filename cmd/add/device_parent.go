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
package add

import (
	"fmt"
	"log"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/util/placement"
	"github.com/Cray-HPE/cani/internal/util/resolve"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// deviceParent is the resolved --parent target for "add device".
type deviceParent struct {
	id       uuid.UUID
	isDevice bool
}

func validDeviceArgs(cmd *cli.Command, args []string) error {
	if err := validSlugOrPartNumber(NounDevice)(cmd, args); err != nil {
		return err
	}
	rackArg, _ := cmd.Flags().GetString("rack")
	if _, isStrategy := placement.ParseStrategy(rackArg); !isStrategy {
		return nil
	}
	parentArg, _ := cmd.Flags().GetString("parent")
	if parentArg != "" && parentArg != uuid.Nil.String() {
		return fmt.Errorf("--parent cannot be combined with a --rack strategy")
	}
	bayArg, _ := cmd.Flags().GetString("bay")
	if bayArg != "" {
		return fmt.Errorf("--bay cannot be combined with a --rack strategy; use --parent without --rack for device bays")
	}
	return nil
}

// resolveDeviceParent resolves --parent (UUID or name of a rack or device).
// An empty or nil-UUID argument means no parent.
func resolveDeviceParent(inv *devicetypes.Inventory, parentArg string) (deviceParent, error) {
	if parentArg == "" || parentArg == uuid.Nil.String() {
		return deviceParent{}, nil
	}
	if id, err := resolve.Rack(inv, parentArg); err == nil {
		return deviceParent{id: id}, nil
	}
	if id, err := resolve.Device(inv, parentArg); err == nil {
		return deviceParent{id: id, isDevice: true}, nil
	}
	return deviceParent{}, fmt.Errorf("--parent %q not found as rack or device", parentArg)
}

// checkBayArgs rejects --bay unless the parent is a device.
func checkBayArgs(parent deviceParent, bayArg string) error {
	if bayArg != "" && !parent.isDevice {
		return fmt.Errorf("--bay requires --parent to name a device with device bays")
	}
	return nil
}

// literalDeviceParent resolves the parent for the literal (non-strategy) add
// path from --rack, or --parent when --rack is unset, and returns --bay.
func literalDeviceParent(cmd *cli.Command, inv *devicetypes.Inventory, rackArg, parentArg string) (deviceParent, string, error) {
	bayArg, _ := cmd.Flags().GetString("bay")
	var parent deviceParent
	if rackArg != "" {
		rackID, err := resolve.Rack(inv, rackArg)
		if err != nil {
			return parent, "", fmt.Errorf("resolving rack %q: %w", rackArg, err)
		}
		parent.id = rackID
	} else {
		var err error
		if parent, err = resolveDeviceParent(inv, parentArg); err != nil {
			return parent, "", err
		}
	}
	return parent, bayArg, checkBayArgs(parent, bayArg)
}

// logAddedDevices reports each added device, noting its bay when installed
// in a parent device.
func logAddedDevices(devices map[uuid.UUID]*devicetypes.CaniDeviceType, qty int) {
	for _, d := range devices {
		if d.ParentBay != "" {
			log.Printf("Added device %s (%s) in bay %q", d.ID, d.Name, d.ParentBay)
			continue
		}
		log.Printf("Added device %s (%s)", d.ID, d.Name)
	}
	log.Printf("%d device(s) added", qty)
}

// applyDeviceParent sets device's parent, installing it into a device bay when
// the parent is a device with bays. pending holds devices built earlier in
// this command.
func applyDeviceParent(inv *devicetypes.Inventory, device *devicetypes.CaniDeviceType, parent deviceParent, bayArg string, pending map[uuid.UUID]*devicetypes.CaniDeviceType) error {
	if parent.id == uuid.Nil {
		return nil
	}
	if !parent.isDevice || (bayArg == "" && !inv.Devices[parent.id].HasDeviceBays()) {
		device.Parent = parent.id
		return nil
	}
	return inv.InstallInBay(device, parent.id, bayArg, pending)
}
