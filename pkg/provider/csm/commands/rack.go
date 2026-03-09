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
package commands

import (
	"fmt"
	"log"
	"strings"

	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// CSM provider defaults for HMN VLAN auto-assignment.
const (
	defaultStartingOrdinal = 9000
	defaultStartingHmnVlan = 3000
	defaultEndingHmnVlan   = 3999
)

// wrapWithCabinetHook wraps a command's RunE so that after the original
// logic completes, any racks missing a cabinet device get one created.
func wrapWithCabinetHook(cmd *cobra.Command) {
	orig := cmd.RunE
	if orig == nil {
		return
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := orig(cmd, args); err != nil {
			return err
		}
		return ensureCabinetDevices(cmd)
	}
}

// ensureCabinetDevices reloads the inventory and creates a cabinet device
// for every rack that does not already have one. When --auto is set it
// assigns a cabinet number (xname) and HMN VLAN from the provider
// defaults defined in the rack-type YAML.
func ensureCabinetDevices(cmd *cobra.Command) error {
	if datastores.Datastore == nil {
		return nil
	}

	inv, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("csm: failed to reload inventory: %w", err)
	}

	autoAssign, _ := cmd.Flags().GetBool("auto")
	changed := false

	for _, rack := range inv.Racks {
		if hasCabinetDevice(inv, rack.ID) {
			continue
		}

		defaults := rackCSMDefaults(rack.Slug)

		cabDev := &devicetypes.CaniDeviceType{
			ID:     uuid.New(),
			Name:   rack.Name,
			Slug:   rack.Slug,
			Type:   devicetypes.TypeCabinet,
			Status: "staged",
			Parent: rack.ID,
		}

		if autoAssign {
			cabNum := suggestCabinetNumber(inv, defaults)
			vlan := suggestCabinetVlan(inv, defaults)

			log.Printf("Querying inventory to suggest Cabinet")
			log.Printf("Suggested cabinet number: %d", cabNum)
			log.Printf("Suggested VLAN ID: %d", vlan)

			maxVlan := defaults.EndingHmnVlan
			if maxVlan == 0 {
				maxVlan = defaultEndingHmnVlan
			}
			if vlan > maxVlan {
				return fmt.Errorf("VLAN exceeds the provider's maximum range (%d).  Please choose a valid VLAN", maxVlan)
			}

			xname := fmt.Sprintf("x%d", cabNum)
			cabDev.ProviderMetadata = map[string]any{
				"csm": map[string]any{
					"xname":   xname,
					"class":   defaults.Class,
					"hmnVlan": vlan,
				},
			}

			log.Printf("Cabinet was successfully staged to be added to the system")
			log.Println() // blank separator
			log.Printf("Cabinet Number: %d", cabNum)
		}

		if err := inv.AddDevices(map[uuid.UUID]*devicetypes.CaniDeviceType{
			cabDev.ID: cabDev,
		}); err != nil {
			return fmt.Errorf("csm: failed to add cabinet device: %w", err)
		}

		// Create chassis devices for device bays defined in the rack type.
		if autoAssign {
			if err := ensureChassisDevices(inv, cabDev, rack, defaults); err != nil {
				return fmt.Errorf("csm: failed to add chassis devices: %w", err)
			}
		}

		changed = true
	}

	if !changed {
		return nil
	}
	return datastores.Datastore.Save(inv)
}

// rackCSMDefaults returns the CSM provider defaults for a rack slug.
// Falls back to sensible defaults when the rack-type YAML has none.
func rackCSMDefaults(slug string) RackProviderDefaultsCSM {
	rt, ok := devicetypes.GetRackTypeBySlug(slug)
	if ok && len(rt.ProviderDefaults) > 0 {
		if d := DecodeRackCSMDefaults(rt.ProviderDefaults); d != nil {
			return *d
		}
	}
	return RackProviderDefaultsCSM{
		Class:           "Hill",
		Ordinal:         defaultStartingOrdinal,
		StartingHmnVlan: defaultStartingHmnVlan,
		EndingHmnVlan:   defaultEndingHmnVlan,
	}
}

// suggestCabinetNumber determines the next available cabinet ordinal
// by scanning existing cabinet xnames and finding the first gap at or
// above the starting ordinal from the provider defaults.
func suggestCabinetNumber(inv *devicetypes.Inventory, defaults RackProviderDefaultsCSM) int {
	used := usedCabinetOrdinals(inv)
	start := defaults.Ordinal
	if start == 0 {
		start = defaultStartingOrdinal
	}
	return nextAvailableInt(start, used)
}

// suggestCabinetVlan determines the next available HMN VLAN by scanning
// existing cabinet devices and finding the first unused value at or above
// the starting VLAN.  The caller is responsible for checking whether the
// returned value exceeds the provider's configured maximum.
func suggestCabinetVlan(inv *devicetypes.Inventory, defaults RackProviderDefaultsCSM) int {
	used := usedHMNVlans(inv)
	start := defaults.StartingHmnVlan
	if start == 0 {
		start = defaultStartingHmnVlan
	}
	return nextAvailableInt(start, used)
}

// usedCabinetOrdinals collects all cabinet ordinals already present in
// the inventory by parsing xnames like "x9000".
func usedCabinetOrdinals(inv *devicetypes.Inventory) map[int]struct{} {
	used := make(map[int]struct{})
	for _, dev := range inv.Devices {
		if dev == nil || dev.GetType() != devicetypes.TypeCabinet {
			continue
		}
		sub, ok := dev.GetProviderSubMap("csm")
		if !ok {
			continue
		}
		xname, _ := sub["xname"].(string)
		if xname == "" {
			continue
		}
		var n int
		if _, err := fmt.Sscanf(xname, "x%d", &n); err == nil {
			used[n] = struct{}{}
		}
	}
	return used
}

// hasCabinetDevice reports whether a cabinet device already exists
// for the given rack ID.
func hasCabinetDevice(inv *devicetypes.Inventory, rackID uuid.UUID) bool {
	for _, dev := range inv.Devices {
		if dev.Parent == rackID && dev.Type == devicetypes.TypeCabinet {
			return true
		}
	}
	return false
}

// usedHMNVlans collects all hmnVlan values already assigned to devices.
func usedHMNVlans(inv *devicetypes.Inventory) map[int]struct{} {
	used := make(map[int]struct{})
	for _, dev := range inv.Devices {
		if dev.ProviderMetadata == nil {
			continue
		}
		csm, ok := dev.ProviderMetadata["csm"]
		if !ok {
			continue
		}
		sub, ok := csm.(map[string]any)
		if !ok {
			continue
		}
		v, ok := sub["hmnVlan"]
		if !ok {
			continue
		}
		switch n := v.(type) {
		case int:
			if n != 0 {
				used[n] = struct{}{}
			}
		case float64:
			if n != 0 {
				used[int(n)] = struct{}{}
			}
		}
	}
	return used
}

// nextAvailableInt returns the first unused integer starting from start.
// It scans upward with no upper bound.
func nextAvailableInt(start int, used map[int]struct{}) int {
	for n := start; ; n++ {
		if _, taken := used[n]; !taken {
			return n
		}
	}
}

// ensureChassisDevices creates chassis child devices for a newly added
// cabinet based on chassis-type device bays defined in the rack type.
func ensureChassisDevices(
	inv *devicetypes.Inventory,
	cabDev *devicetypes.CaniDeviceType,
	rack *devicetypes.CaniRackType,
	defaults RackProviderDefaultsCSM,
) error {
	cabSub, ok := cabDev.GetProviderSubMap("csm")
	if !ok {
		return nil
	}
	cabXname, _ := cabSub["xname"].(string)
	if cabXname == "" {
		return nil
	}

	devs := make(map[uuid.UUID]*devicetypes.CaniDeviceType)
	for _, bay := range rack.DeviceBays {
		if !strings.HasPrefix(bay.Name, "Chassis") {
			continue
		}
		chassisXname := fmt.Sprintf("%sc%d", cabXname, BayOrdinal(bay))
		chDev := &devicetypes.CaniDeviceType{
			ID:     uuid.New(),
			Name:   chassisXname,
			Type:   devicetypes.TypeChassis,
			Status: "staged",
			Parent: cabDev.ID,
			ProviderMetadata: map[string]any{
				"csm": map[string]any{
					"xname": chassisXname,
					"class": defaults.Class,
				},
			},
		}
		devs[chDev.ID] = chDev
	}
	if len(devs) == 0 {
		return nil
	}
	return inv.AddDevices(devs)
}
