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
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// addAnyModule adds module(s) using the resolved module type.
func addAnyModule(cmd *cli.Command, args []string, mod *devicetypes.CaniModuleType, qty int) error {
	parentArg, _ := cmd.Flags().GetString("parent")
	statusArg, _ := cmd.Flags().GetString("status")
	serialArg, _ := cmd.Flags().GetString("serial")

	names, err := resolveNamesFromFlags(cmd, qty)
	if err != nil {
		return err
	}

	inventory, err := loadInventoryForAdd(cmd, args)
	if err != nil {
		return err
	}

	statusArg, err = normalizeStatus(statusArg, inventory)
	if err != nil {
		return err
	}

	if err := addModules(inventory, mod, names, parentArg, statusArg, serialArg, qty); err != nil {
		return err
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("%d module(s) added", qty)
	return nil
}

// addModules builds and adds qty modules to the inventory.
func addModules(inventory *devicetypes.Inventory, mod *devicetypes.CaniModuleType, names []string, parentArg, statusArg, serialArg string, qty int) error {
	for i := range qty {
		m := *mod
		m.ID = uuid.New()
		setModuleParent(&m, parentArg)
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
	return nil
}

// setModuleParent assigns a parent device UUID to the module when parentArg is
// a non-empty, parseable, non-nil UUID string.
func setModuleParent(m *devicetypes.CaniModuleType, parentArg string) {
	if parentArg == uuid.Nil.String() || parentArg == "" {
		return
	}
	if did, derr := uuid.Parse(parentArg); derr == nil {
		m.ParentDevice = did
	}
}

// addAnyCable adds cable(s) using the resolved cable type.
func addAnyCable(cmd *cli.Command, args []string, cable *devicetypes.CaniCableType, qty int) error {
	statusArg, _ := cmd.Flags().GetString("status")

	names, err := resolveNamesFromFlags(cmd, qty)
	if err != nil {
		return err
	}

	inventory, err := loadInventoryForAdd(cmd, args)
	if err != nil {
		return err
	}

	statusArg, err = normalizeStatus(statusArg, inventory)
	if err != nil {
		return err
	}

	if err := addCables(inventory, cable, names, statusArg, qty); err != nil {
		return err
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("%d cable(s) added", qty)
	return nil
}

// addCables builds and adds qty cables to the inventory.
func addCables(inventory *devicetypes.Inventory, cable *devicetypes.CaniCableType, names []string, statusArg string, qty int) error {
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
	return nil
}
