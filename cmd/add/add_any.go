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
	"regexp"
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/internal/util/nameexpand"
	"github.com/Cray-HPE/cani/internal/util/validate"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// addAny looks up the argument across all hardware registries and delegates
// to the appropriate add logic based on the category it resolves to.
func addAny(cmd *cobra.Command, args []string) error {
	key := args[0]

	result, err := devicetypes.LookupAny(key)
	if err != nil {
		return err
	}

	qty, _ := cmd.Flags().GetInt("qty")
	if qty < 1 {
		qty = 1
	}

	log.Printf("Resolved %q as %s", key, result.Category)

	switch result.Category {
	case devicetypes.CategoryRack:
		return addAnyRack(cmd, args, result.Rack, qty)
	case devicetypes.CategoryDevice:
		return addAnyDevice(cmd, args, result.Device, qty)
	case devicetypes.CategoryModule:
		return addAnyModule(cmd, args, result.Module, qty)
	case devicetypes.CategoryCable:
		return addAnyCable(cmd, args, result.Cable, qty)
	default:
		return fmt.Errorf("unsupported category %q for %q", result.Category, key)
	}
}

// addAnyRack adds rack(s) using the resolved rack type.
func addAnyRack(cmd *cobra.Command, args []string, rack *devicetypes.CaniRackType, qty int) error {
	statusArg, _ := cmd.Flags().GetString("status")
	serialArg, _ := cmd.Flags().GetString("serial")

	nameArg, _ := cmd.Flags().GetString("name")
	prefix, _ := cmd.Flags().GetString("prefix")
	start, _ := cmd.Flags().GetInt("start")
	padWidth, _ := cmd.Flags().GetInt("pad-width")

	names, err := nameexpand.ResolveNames(nameArg, prefix, start, padWidth, qty)
	if err != nil {
		return fmt.Errorf("name resolution failed: %w", err)
	}

	locationArg, _ := cmd.Flags().GetString("location")

	if err := datastores.SetDeviceStore(cmd, args); err != nil {
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

	locationID := resolveLocation(inventory, locationArg)

	tags, _ := cmd.Flags().GetStringArray("tag")
	provMeta := collectProviderMetadata(cmd)

	for i := range qty {
		r := *rack
		r.ID = uuid.New()
		r.Location = locationID
		if names != nil {
			r.Name = names[i]
		} else if r.Name == "" && r.Model != "" {
			r.Name = r.Model
		}
		if statusArg != "" {
			r.Status = statusArg
		}
		if serialArg != "" {
			r.Serial = serialArg
		}
		applyTagsToRack(&r, tags)
		applyProviderMetadataToRack(&r, provMeta)

		// Let registered providers apply post-add logic.
		if err := runRackPostAddHooks(&r, inventory); err != nil {
			return fmt.Errorf("provider hook failed: %w", err)
		}

		if err := inventory.AddRack(&r); err != nil {
			return fmt.Errorf("failed to add rack: %w", err)
		}
		inventory.AssignRacksToLocation(locationID)

		log.Printf("Added rack %s (%s)", r.ID, r.Name)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("%d rack(s) added", qty)
	return nil
}

// addAnyDevice adds device(s) using the resolved device type.
func addAnyDevice(cmd *cobra.Command, args []string, device *devicetypes.CaniDeviceType, qty int) error {
	parentArg, _ := cmd.Flags().GetString("parent")
	statusArg, _ := cmd.Flags().GetString("status")
	serialArg, _ := cmd.Flags().GetString("serial")
	nameArg, _ := cmd.Flags().GetString("name")
	prefix, _ := cmd.Flags().GetString("prefix")
	start, _ := cmd.Flags().GetInt("start")
	padWidth, _ := cmd.Flags().GetInt("pad-width")
	auto, _ := cmd.Flags().GetBool("auto")

	names, err := nameexpand.ResolveNames(nameArg, prefix, start, padWidth, qty)
	if err != nil {
		return fmt.Errorf("name resolution failed: %w", err)
	}

	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	locationID := inventory.EnsureLocation()
	inventory.AssignRacksToLocation(locationID)

	// When --auto is set, try to stage existing imported devices
	// that match the requested slug instead of creating new ones.
	// Delegate to the first registered provider that implements DeviceStager.
	if auto && device.Slug != "" {
		// Snapshot which devices are already staged before staging.
		alreadyStaged := make(map[uuid.UUID]bool)
		for id, d := range inventory.Devices {
			if strings.EqualFold(d.Status, string(devicetypes.StatusStaged)) {
				alreadyStaged[id] = true
			}
		}

		staged := 0
		for _, p := range provider.GetProviders() {
			// First try staging under a staged rack (new cabinet).
			if rs, ok := p.(provider.RackStager); ok {
				for range qty {
					if rs.StageNewInRack(inventory, device.Slug) {
						staged++
					}
				}
			}
			if staged > 0 {
				break
			}
			// Fall back to re-staging existing devices.
			if stager, ok := p.(provider.DeviceStager); ok {
				for range qty {
					if stager.StageExisting(inventory, device.Slug) {
						staged++
					}
				}
			}
			if staged > 0 {
				break
			}
		}
		if staged > 0 {
			if err := datastores.Datastore.Save(inventory); err != nil {
				return fmt.Errorf("failed to save inventory: %w", err)
			}
			logStagedDevices(inventory, alreadyStaged)
			log.Printf("%d device(s) added", staged)
			return nil
		}
	}

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
		d := *device
		d.ID = uuid.New()

		if parentArg != uuid.Nil.String() && parentArg != "" {
			if pid, perr := uuid.Parse(parentArg); perr == nil {
				d.Parent = pid
			}
		}

		if names != nil {
			d.Name = names[i]
		}
		if statusArg != "" {
			d.Status = statusArg
		}
		if serialArg != "" {
			d.Serial = serialArg
		}
		applyTagsToDevice(&d, tags)
		applyProviderMetadataToDevice(&d, provMeta)

		devicesToAdd[d.ID] = &d

		// Expand child devices from device-bay defaults.
		for cid, child := range devicetypes.ExpandChildren(&d) {
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

// addAnyModule adds module(s) using the resolved module type.
func addAnyModule(cmd *cobra.Command, args []string, mod *devicetypes.CaniModuleType, qty int) error {
	parentArg, _ := cmd.Flags().GetString("parent")
	statusArg, _ := cmd.Flags().GetString("status")
	serialArg, _ := cmd.Flags().GetString("serial")
	nameArg, _ := cmd.Flags().GetString("name")
	prefix, _ := cmd.Flags().GetString("prefix")
	start, _ := cmd.Flags().GetInt("start")
	padWidth, _ := cmd.Flags().GetInt("pad-width")

	names, err := nameexpand.ResolveNames(nameArg, prefix, start, padWidth, qty)
	if err != nil {
		return fmt.Errorf("name resolution failed: %w", err)
	}

	if err := datastores.SetDeviceStore(cmd, args); err != nil {
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

	for i := range qty {
		m := *mod
		m.ID = uuid.New()

		if parentArg != uuid.Nil.String() && parentArg != "" {
			if did, derr := uuid.Parse(parentArg); derr == nil {
				m.ParentDevice = did
			}
		}

		if names != nil {
			m.Name = names[i]
		}
		if statusArg != "" {
			m.Status = statusArg
		}
		if serialArg != "" {
			m.Serial = serialArg
		}

		if err := inventory.AddModule(&m); err != nil {
			return fmt.Errorf("failed to add module: %w", err)
		}
		log.Printf("Added module %s (%s)", m.ID, m.Name)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("%d module(s) added", qty)
	return nil
}

// addAnyCable adds cable(s) using the resolved cable type.
func addAnyCable(cmd *cobra.Command, args []string, cable *devicetypes.CaniCableType, qty int) error {
	statusArg, _ := cmd.Flags().GetString("status")
	nameArg, _ := cmd.Flags().GetString("name")
	prefix, _ := cmd.Flags().GetString("prefix")
	start, _ := cmd.Flags().GetInt("start")
	padWidth, _ := cmd.Flags().GetInt("pad-width")

	names, err := nameexpand.ResolveNames(nameArg, prefix, start, padWidth, qty)
	if err != nil {
		return fmt.Errorf("name resolution failed: %w", err)
	}

	if err := datastores.SetDeviceStore(cmd, args); err != nil {
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

	for i := range qty {
		c := *cable
		c.ID = uuid.New()

		if names != nil {
			c.Label = names[i]
		}
		if statusArg != "" {
			c.Status = statusArg
		}

		if err := inventory.AddCable(&c); err != nil {
			return fmt.Errorf("failed to add cable: %w", err)
		}
		log.Printf("Added cable %s (%s)", c.ID, c.Label)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("%d cable(s) added", qty)
	return nil
}

// logStagedDevices logs detailed staging info for devices that were
// newly staged (not in alreadyStaged). It prints the display type,
// UUID, and xname components (Cabinet, Chassis, Blade) for each.
func logStagedDevices(inv *devicetypes.Inventory, alreadyStaged map[uuid.UUID]bool) {
	for id, dev := range inv.Devices {
		if alreadyStaged[id] {
			continue
		}
		if !strings.EqualFold(dev.Status, string(devicetypes.StatusStaged)) {
			continue
		}
		log.Printf("Added device %s (%s)", id, dev.Name)
		typeName := displayTypeName(dev.GetType())
		log.Printf("%s was successfully staged to be added to the system", typeName)
		log.Printf("UUID: %s", id)
		sub, ok := dev.GetProviderSubMap("csm")
		if ok {
			if xn, _ := sub["xname"].(string); xn != "" {
				cab, chas, slot := parseXnameParts(xn)
				log.Printf("Cabinet: %d", cab)
				log.Printf("Chassis: %d", chas)
				log.Printf("Blade: %d", slot)
			}
		}
	}
}

// displayTypeName returns a human-friendly name for a device type.
func displayTypeName(t devicetypes.Type) string {
	switch devicetypes.Type(strings.ToLower(string(t))) {
	case devicetypes.TypeCabinet:
		return "Cabinet"
	case devicetypes.TypeChassis, devicetypes.TypeRack:
		return "Chassis"
	case devicetypes.TypeBlade:
		return "Blade"
	case devicetypes.TypeNode:
		return "Node"
	case devicetypes.TypeNodeCard:
		return "NodeBlade"
	default:
		s := string(t)
		if len(s) == 0 {
			return s
		}
		return strings.ToUpper(s[:1]) + s[1:]
	}
}

// xnameRe matches xname strings like x8000, x8000c0, x8000c0s0, etc.
var xnameRe = regexp.MustCompile(`^x(\d+)(?:c(\d+))?(?:s(\d+))?`)

// parseXnameParts extracts cabinet, chassis, and slot numbers from an
// xname string. Returns (0, 0, 0) for unrecognised formats.
func parseXnameParts(xn string) (cabinet, chassis, slot int) {
	m := xnameRe.FindStringSubmatch(xn)
	if m == nil {
		return 0, 0, 0
	}
	cabinet, _ = strconv.Atoi(m[1])
	if m[2] != "" {
		chassis, _ = strconv.Atoi(m[2])
	}
	if m[3] != "" {
		slot, _ = strconv.Atoi(m[3])
	}
	return cabinet, chassis, slot
}
