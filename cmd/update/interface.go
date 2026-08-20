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
package update

import (
	"fmt"
	"log"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/util/store"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// newInterfaceCommand creates the "update interface" subcommand.
func newInterfaceCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "interface [uuid]",
		Short: "Update interface properties.",
		Long: `Update one or more interfaces on a device or module.

Examples:
  # List interfaces on a device
  cani update interface --device switch-01 -L

  # Set role by device name and interface name
  cani update interface --device switch-01 --name osfp1 --role hsn

  # Set role on multiple interfaces matching a glob pattern
  cani update interface --device switch-01 --name "1/1/*" --role UplinkInterface

  # Set role by interface UUID
  cani update interface 3fa85f64-5717-4562-b3fc-2c963f66afa6 --role management

  # Set label on an interface
  cani update interface --device server-01 --name eth0 --label "BMC Network"

  # Set MAC address on an interface
  cani update interface --device server-01 --name iLO --mac aa:bb:cc:dd:ee:ff

  # Target a specific module's interface (disambiguates names shared with the
  # parent device or sibling modules, e.g. multiple "HSN 0" ports on one node)
  cani update interface --module CX7-server-01 --name "HSN 0" --mac aa:bb:cc:dd:ee:ff`,
		Args: cli.MaximumNArgs(1),
		RunE: updateInterface,
	}

	cmd.Flags().String("device", "", "Device name or UUID (required when not using positional UUID)")
	cmd.Flags().String("module", "", "Module name or UUID (targets only that module's own interfaces)")
	cmd.Flags().String("name", "", "Interface name or glob pattern (e.g. \"1/1/*\")")
	cmd.Flags().String("role", "", "Interface role (e.g. management, hsn, storage, access)")
	cmd.Flags().String("label", "", "Interface label")
	cmd.Flags().String("mac", "", "Interface MAC address (e.g. aa:bb:cc:dd:ee:ff)")
	cmd.Flags().String("lag", "", "Parent LAG interface name (adds this interface as a LAG member)")
	cmd.Flags().String("mode", "", "Switchport mode: access, tagged, or tagged-all")
	cmd.Flags().Int("untagged-vlan", 0, "Untagged (native) VLAN ID")
	cmd.Flags().StringSlice("tagged-vlan", nil, "Tagged VLAN ID (comma-separated or repeatable)")
	cmd.Flags().String("vrf", "", "VRF name to assign to the interface")
	cmd.Flags().String(flagDescription, "", "Interface description")
	cmd.Flags().BoolP("list", "L", false, "List interfaces for the specified device")

	return cmd
}

func updateInterface(cmd *cli.Command, args []string) error {
	if err := store.Setup(cmd); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	// Handle --list / -L mode
	listMode, _ := cmd.Flags().GetBool("list")
	if listMode {
		return listDeviceInterfaces(cmd, inventory)
	}

	updates, err := parseInterfaceUpdates(cmd, inventory)
	if err != nil {
		return err
	}

	// Resolve target interfaces
	targets, err := resolveInterfaces(cmd, args, inventory)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return fmt.Errorf("no interfaces matched the specified criteria")
	}

	applyInterfaceUpdates(cmd, targets, updates)

	if err := finalizeInterfaceUpdate(inventory); err != nil {
		return err
	}

	logInterfaceUpdate(targets)
	return nil
}

// finalizeInterfaceUpdate rebuilds derived relationships and persists the inventory.
func finalizeInterfaceUpdate(inventory *devicetypes.Inventory) error {
	// Rebuild relationships so derived fields are updated.
	result := inventory.VerifyParentChildRelationships()
	if result.HasErrors() {
		return fmt.Errorf("relationship errors: %v", result.Errors)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}
	return nil
}

// logInterfaceUpdate logs the result of an interface update.
func logInterfaceUpdate(targets []interfaceTarget) {
	if len(targets) == 1 {
		log.Printf("Updated interface %s (%s)", targets[0].instance.Name, targets[0].instance.ID)
	} else {
		log.Printf("Updated %d interfaces", len(targets))
	}
}

// interfaceTarget pairs a CaniInterface with its parent spec (if found).
type interfaceTarget struct {
	instance *devicetypes.CaniInterface
	spec     *devicetypes.InterfaceSpec
}
