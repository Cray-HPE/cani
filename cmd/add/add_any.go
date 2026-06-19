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
	"github.com/Cray-HPE/cani/internal/util/nameexpand"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// addAny looks up the argument across all hardware registries and delegates
// to the appropriate add logic based on the category it resolves to.
func addAny(cmd *cli.Command, args []string) error {
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

// resolveNamesFromFlags resolves device/rack/module/cable names from the
// standard naming flags (name, prefix, start, pad-width).
func resolveNamesFromFlags(cmd *cli.Command, qty int) ([]string, error) {
	nameArg, _ := cmd.Flags().GetString("name")
	prefix, _ := cmd.Flags().GetString("prefix")
	start, _ := cmd.Flags().GetInt("start")
	padWidth, _ := cmd.Flags().GetInt("pad-width")

	names, err := nameexpand.ResolveNames(nameArg, prefix, start, padWidth, qty)
	if err != nil {
		return nil, fmt.Errorf("name resolution failed: %w", err)
	}
	return names, nil
}

// loadInventoryForAdd sets the device store from the command and loads the
// current inventory.
func loadInventoryForAdd(cmd *cli.Command, args []string) (*devicetypes.Inventory, error) {
	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return nil, fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load inventory: %w", err)
	}
	return inventory, nil
}
