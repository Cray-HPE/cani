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
	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	locationID := inventory.EnsureLocation()

	for range qty {
		r := *rack
		r.ID = uuid.New()
		r.Location = locationID
		if r.Name == "" && r.Model != "" {
			r.Name = r.Model
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

	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	locationID := inventory.EnsureLocation()
	inventory.AssignRacksToLocation(locationID)

	devicesToAdd := make(map[uuid.UUID]*devicetypes.CaniDeviceType)
	for range qty {
		d := *device
		d.ID = uuid.New()

		if parentArg != uuid.Nil.String() && parentArg != "" {
			if pid, perr := uuid.Parse(parentArg); perr == nil {
				d.Parent = pid
			}
		}

		devicesToAdd[d.ID] = &d
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

	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	for range qty {
		m := *mod
		m.ID = uuid.New()

		if parentArg != uuid.Nil.String() && parentArg != "" {
			if did, derr := uuid.Parse(parentArg); derr == nil {
				m.ParentDevice = did
			}
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
	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	for range qty {
		c := *cable
		c.ID = uuid.New()

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
