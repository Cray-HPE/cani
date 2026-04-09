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
	"os"
	"strconv"

	"github.com/Cray-HPE/cani/internal/util/nameexpand"
	"github.com/Cray-HPE/cani/internal/util/placement"
	"github.com/Cray-HPE/cani/internal/util/resolve"
	"github.com/Cray-HPE/cani/internal/util/validate"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// newDeviceCommand creates the "add device" subcommand.
func newDeviceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "device <slug-or-part-number>",
		Short: "Add device(s) to the inventory.",
		Long:  "Add one or more devices to the inventory by slug or part number.",
		Args:  validSlugOrPartNumber(NounDevice),
		RunE:  addDevice,
	}

	cmd.Flags().String("rack", "", "Parent rack UUID, name, or strategy (%{FILL}, %{SPREAD})")
	cmd.Flags().Int("position", 0, "Rack U position")
	cmd.Flags().String("face", "", "Rack face (front, rear)")
	cmd.Flags().String("name", "", "Device name, expansion pattern, or template (%{RACK}, %{U}, %{SEQ}, %{ZONE})")
	cmd.Flags().String("location", "", "Location filter for rack selection (name or UUID)")
	cmd.Flags().String("zone", "", "Rack zone (top, middle, bottom) — overrides auto-detection from hardware type")
	cmd.Flags().Bool("dry-run", false, "Show placement plan without committing changes")

	return cmd
}

func addDevice(cmd *cobra.Command, args []string) error {
	qty, _ := cmd.Flags().GetInt("qty")
	if qty < 1 {
		qty = 1
	}

	result, err := lookupBySlugOrPart(NounDevice, args[0])
	if err != nil {
		return err
	}

	rackArg, _ := cmd.Flags().GetString("rack")
	nameArg, _ := cmd.Flags().GetString("name")

	strategy, isStrategy := placement.ParseStrategy(rackArg)
	if isStrategy {
		return addDeviceStrategy(cmd, result, qty, nameArg, strategy)
	}
	return addDeviceLiteral(cmd, result, qty, nameArg, rackArg)
}

// addDeviceStrategy handles multi-rack auto-placement with %{FILL}/%{SPREAD}.
func addDeviceStrategy(cmd *cobra.Command, result *lookupResult, qty int, nameArg string, strategy placement.Strategy) error {
	face, _ := cmd.Flags().GetString("face")
	locationArg, _ := cmd.Flags().GetString("location")
	zoneArg, _ := cmd.Flags().GetString("zone")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	statusArg, _ := cmd.Flags().GetString("status")
	serialArg, _ := cmd.Flags().GetString("serial")

	if face == "" {
		face = devicetypes.FaceFront
	}

	// Resolve the optional --zone flag.
	var zone placement.Zone
	if zoneArg != "" {
		z, err := placement.ParseZone(zoneArg)
		if err != nil {
			return err
		}
		zone = z
	}

	isTemplate := nameexpand.IsTemplate(nameArg)
	if !isTemplate && nameArg != "" {
		return fmt.Errorf("strategy placement requires template naming (%%{RACK}, %%{U}, etc.) or no --name flag")
	}

	if err := datastores.SetDeviceStore(cmd, nil); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}
	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	if statusArg != "" {
		normalized, verr := validate.StatusWithInventory(statusArg, inventory)
		if verr != nil {
			return verr
		}
		statusArg = normalized
	}

	var racks []*devicetypes.CaniRackType
	if locationArg != "" {
		locationID := resolveLocation(inventory, locationArg)
		inventory.AssignRacksToLocation(locationID)
		racks = inventory.RacksByLocation(locationID)
	} else {
		racks = inventory.AllRacks()
	}
	if len(racks) == 0 {
		return fmt.Errorf("no racks found at location %q", locationArg)
	}

	entries, err := placement.Plan(racks, result.Device.UHeight, face, result.Device.IsFullDepth, qty, strategy, zone, string(result.Device.Type))
	if err != nil {
		return err
	}

	names := resolveTemplateNames(nameArg, entries, face)

	if dryRun {
		placement.PrintPlanWithHeight(os.Stdout, entries, names, result.Device.UHeight)
		return nil
	}

	tags, _ := cmd.Flags().GetStringArray("tag")
	provMeta := collectProviderMetadata(cmd)

	devicesToAdd := make(map[uuid.UUID]*devicetypes.CaniDeviceType, qty)
	for i, e := range entries {
		device := *result.Device
		device.ID = uuid.New()
		device.Parent = e.RackID
		device.RackPosition = e.StartU
		device.Face = e.Face
		if i < len(names) {
			device.Name = names[i]
		}
		if statusArg != "" {
			device.Status = statusArg
		}
		if serialArg != "" {
			device.Serial = serialArg
		}
		applyTagsToDevice(&device, tags)
		applyProviderMetadataToDevice(&device, provMeta)
		devicesToAdd[device.ID] = &device

		// Expand child devices from device-bay defaults.
		for cid, child := range devicetypes.ExpandChildren(&device) {
			devicesToAdd[cid] = child
		}

		rack := inventory.Racks[e.RackID]
		if rack != nil {
			if !rack.PlaceDevice(device.ID, e.StartU, device.UHeight, e.Face, device.IsFullDepth) {
				return fmt.Errorf(
					"cannot place device at U%d–U%d (%s) in rack %s: position is already occupied or out of bounds",
					e.StartU, e.StartU+device.UHeight-1, e.Face, rack.Name,
				)
			}
		}
	}

	if err := inventory.AddDevices(devicesToAdd); err != nil {
		return fmt.Errorf("failed to add devices: %w", err)
	}
	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	for _, d := range devicesToAdd {
		log.Printf("Added device %s (%s) in rack %s at U%d", d.ID, d.Name, d.Parent, d.RackPosition)
	}
	log.Printf("%d device(s) added via %s strategy", qty, strategy)
	return nil
}

// addDeviceLiteral handles the original single-rack flow.
func addDeviceLiteral(cmd *cobra.Command, result *lookupResult, qty int, nameArg, rackArg string) error {
	parentArg, _ := cmd.Flags().GetString("parent")
	position, _ := cmd.Flags().GetInt("position")
	face, _ := cmd.Flags().GetString("face")
	statusArg, _ := cmd.Flags().GetString("status")
	serialArg, _ := cmd.Flags().GetString("serial")
	prefix, _ := cmd.Flags().GetString("prefix")
	start, _ := cmd.Flags().GetInt("start")
	padWidth, _ := cmd.Flags().GetInt("pad-width")

	names, err := nameexpand.ResolveNames(nameArg, prefix, start, padWidth, qty)
	if err != nil {
		return fmt.Errorf("name resolution failed: %w", err)
	}

	if err := datastores.SetDeviceStore(cmd, nil); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	locationID := inventory.EnsureLocation()
	inventory.AssignRacksToLocation(locationID)

	if statusArg != "" {
		normalized, verr := validate.StatusWithInventory(statusArg, inventory)
		if verr != nil {
			return verr
		}
		statusArg = normalized
	}

	tags, _ := cmd.Flags().GetStringArray("tag")
	provMeta := collectProviderMetadata(cmd)

	devicesToAdd := make(map[uuid.UUID]*devicetypes.CaniDeviceType)
	for i := range qty {
		device := *result.Device // shallow copy
		device.ID = uuid.New()

		if rackArg != "" {
			rackID, rerr := resolve.Rack(inventory, rackArg)
			if rerr != nil {
				return fmt.Errorf("resolving rack %q: %w", rackArg, rerr)
			}
			device.Parent = rackID
		} else if parentArg != "" {
			if pid, perr := uuid.Parse(parentArg); perr == nil {
				device.Parent = pid
			}
		}

		device.RackPosition = position
		device.Face = face

		if names != nil {
			device.Name = names[i]
		} else if nameexpand.IsTemplate(nameArg) {
			rackName := rackArg
			if rack := inventory.Racks[device.Parent]; rack != nil {
				rackName = rack.Name
			}
			vars := map[string]string{
				"RACK":   rackName,
				"PARENT": rackName,
				"U":      strconv.Itoa(device.RackPosition),
				"SEQ":    strconv.Itoa(i + 1),
				"FACE":   face,
			}
			expanded, eerr := nameexpand.ExpandTemplate(nameArg, vars)
			if eerr != nil {
				return fmt.Errorf("template expansion failed: %w", eerr)
			}
			device.Name = expanded
		}

		if statusArg != "" {
			device.Status = statusArg
		}
		if serialArg != "" {
			device.Serial = serialArg
		}
		applyTagsToDevice(&device, tags)
		applyProviderMetadataToDevice(&device, provMeta)

		// Place in rack OccupiedSlots so the rack view is accurate.
		if rack := inventory.Racks[device.Parent]; rack != nil && device.RackPosition > 0 {
			height := device.UHeight
			if height < 1 {
				height = 1
			}
			if !rack.PlaceDevice(device.ID, device.RackPosition, height, device.Face, device.IsFullDepth) {
				return fmt.Errorf(
					"cannot place device at U%d–U%d (%s) in rack %s: position is already occupied or out of bounds",
					device.RackPosition, device.RackPosition+height-1, device.Face, rack.Name,
				)
			}
		}

		devicesToAdd[device.ID] = &device

		// Expand child devices from device-bay defaults.
		for cid, child := range devicetypes.ExpandChildren(&device) {
			devicesToAdd[cid] = child
		}
	}

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

// resolveLocation finds or ensures a location in the inventory.
func resolveLocation(inventory *devicetypes.Inventory, locationArg string) uuid.UUID {
	if locationArg != "" {
		if loc := inventory.FindLocationByNameOrID(locationArg); loc != nil {
			return loc.ID
		}
	}
	return inventory.EnsureLocation()
}

// resolveTemplateNames expands template patterns for each placement entry.
func resolveTemplateNames(nameArg string, entries []placement.PlacementEntry, face string) []string {
	if nameArg == "" || !nameexpand.IsTemplate(nameArg) {
		return nil
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		vars := map[string]string{
			"RACK":   e.RackName,
			"PARENT": e.RackName,
			"U":      strconv.Itoa(e.StartU),
			"SEQ":    strconv.Itoa(i + 1),
			"FACE":   face,
			"ZONE":   e.Zone,
		}
		name, err := nameexpand.ExpandTemplate(nameArg, vars)
		if err != nil {
			log.Printf("warning: template expansion failed for entry %d: %v", i, err)
			continue
		}
		names[i] = name
	}
	return names
}
