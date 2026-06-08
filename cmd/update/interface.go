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
	"path/filepath"
	"strings"

	"github.com/Cray-HPE/cani/internal/util/resolve"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// newInterfaceCommand creates the "update interface" subcommand.
func newInterfaceCommand() *cobra.Command {
	cmd := &cobra.Command{
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
		Args: cobra.MaximumNArgs(1),
		RunE: updateInterface,
	}

	cmd.Flags().String("device", "", "Device name or UUID (required when not using positional UUID)")
	cmd.Flags().String("module", "", "Module name or UUID (targets only that module's own interfaces)")
	cmd.Flags().String("name", "", "Interface name or glob pattern (e.g. \"1/1/*\")")
	cmd.Flags().String("role", "", "Interface role (e.g. management, hsn, storage, access)")
	cmd.Flags().String("label", "", "Interface label")
	cmd.Flags().String("mac", "", "Interface MAC address (e.g. aa:bb:cc:dd:ee:ff)")
	cmd.Flags().BoolP("list", "L", false, "List interfaces for the specified device")

	return cmd
}

func updateInterface(cmd *cobra.Command, args []string) error {
	if err := datastores.SetDeviceStore(cmd, args); err != nil {
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

	role, _ := cmd.Flags().GetString("role")
	label, _ := cmd.Flags().GetString("label")
	mac, _ := cmd.Flags().GetString("mac")

	if !cmd.Flags().Changed("role") && !cmd.Flags().Changed("label") && !cmd.Flags().Changed("mac") {
		return fmt.Errorf("at least one of --role, --label, or --mac must be specified")
	}

	if role != "" {
		if warn := devicetypes.ValidateInterfaceRole(role); warn != "" {
			log.Printf("Warning: %s", warn)
		}
	}

	if cmd.Flags().Changed("mac") {
		normalized, nerr := devicetypes.NormalizeMAC(mac)
		if nerr != nil {
			return nerr
		}
		mac = normalized
	}

	// Resolve target interfaces
	targets, err := resolveInterfaces(cmd, args, inventory)
	if err != nil {
		return err
	}

	if len(targets) == 0 {
		return fmt.Errorf("no interfaces matched the specified criteria")
	}

	// Apply updates
	for _, t := range targets {
		if cmd.Flags().Changed("role") {
			t.instance.Role = role
			if t.spec != nil {
				t.spec.Role = role
			}
		}
		if cmd.Flags().Changed("label") {
			t.instance.Label = label
			if t.spec != nil {
				t.spec.Label = label
			}
		}
		if cmd.Flags().Changed("mac") {
			t.instance.MacAddress = mac
			if t.spec != nil {
				t.spec.MacAddress = mac
			}
		}
	}

	// Rebuild relationships so derived fields are updated.
	result := inventory.VerifyParentChildRelationships()
	if result.HasErrors() {
		return fmt.Errorf("relationship errors: %v", result.Errors)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	if len(targets) == 1 {
		log.Printf("Updated interface %s (%s)", targets[0].instance.Name, targets[0].instance.ID)
	} else {
		log.Printf("Updated %d interfaces", len(targets))
	}
	return nil
}

// interfaceTarget pairs a CaniInterface with its parent spec (if found).
type interfaceTarget struct {
	instance *devicetypes.CaniInterface
	spec     *devicetypes.InterfaceSpec
}

// resolveInterfaces finds the target interface(s) based on either:
//   - A positional UUID argument
//   - --device + --name flags (name supports glob patterns)
func resolveInterfaces(cmd *cobra.Command, args []string, inv *devicetypes.Inventory) ([]interfaceTarget, error) {
	// Case 1: Positional UUID argument
	if len(args) == 1 {
		id, err := uuid.Parse(args[0])
		if err != nil {
			return nil, fmt.Errorf("invalid interface UUID: %w", err)
		}
		iface, ok := inv.Interfaces[id]
		if !ok {
			return nil, fmt.Errorf("interface %s not found", id)
		}
		spec := findInterfaceSpec(inv, iface)
		return []interfaceTarget{{instance: iface, spec: spec}}, nil
	}

	// Case 2: --device/--module + --name
	deviceRef, _ := cmd.Flags().GetString("device")
	moduleRef, _ := cmd.Flags().GetString("module")
	namePattern, _ := cmd.Flags().GetString("name")

	if deviceRef == "" && moduleRef == "" {
		return nil, fmt.Errorf("either a positional UUID, --device, or --module flag is required")
	}
	if deviceRef != "" && moduleRef != "" {
		return nil, fmt.Errorf("--device and --module are mutually exclusive")
	}
	if namePattern == "" {
		return nil, fmt.Errorf("--name is required when using --device or --module")
	}

	// --module targets only the named module's own interfaces, which
	// disambiguates interface names shared with the parent device or
	// sibling modules (e.g. multiple "HSN 0" ports on one node).
	if moduleRef != "" {
		moduleID, err := resolve.Module(inv, moduleRef)
		if err != nil {
			return nil, fmt.Errorf("resolving --module: %w", err)
		}
		return collectModuleInterfaceTargets(inv, moduleID, namePattern)
	}

	deviceID, err := resolve.Device(inv, deviceRef)
	if err != nil {
		return nil, fmt.Errorf("resolving --device: %w", err)
	}
	return collectInterfaceTargets(inv, namePattern, func(iface *devicetypes.CaniInterface) bool {
		return belongsToDevice(inv, iface, deviceID)
	})
}

// collectModuleInterfaceTargets returns targets for the named module's own
// interfaces. Module interfaces are flattened into inv.Interfaces under the
// parent device's ID, so they cannot be distinguished by DeviceID; instead we
// walk the module's interface specs (which carry the canonical ID) and pair
// each with its top-level instance looked up by that ID.
func collectModuleInterfaceTargets(inv *devicetypes.Inventory, moduleID uuid.UUID, namePattern string) ([]interfaceTarget, error) {
	mod := inv.Modules[moduleID]
	if mod == nil {
		return nil, fmt.Errorf("module %s not found", moduleID)
	}
	var targets []interfaceTarget
	for i := range mod.Interfaces {
		spec := &mod.Interfaces[i]
		matched, merr := matchInterfaceName(namePattern, spec.Name)
		if merr != nil {
			return nil, fmt.Errorf("invalid --name pattern %q: %w", namePattern, merr)
		}
		if !matched && !strings.EqualFold(spec.Name, namePattern) {
			continue
		}
		instance := inv.Interfaces[spec.ID]
		if instance == nil {
			continue
		}
		targets = append(targets, interfaceTarget{instance: instance, spec: spec})
	}
	return targets, nil
}

// collectInterfaceTargets returns interface targets whose name matches the
// given pattern and that satisfy the supplied ownership predicate.
func collectInterfaceTargets(inv *devicetypes.Inventory, namePattern string, owns func(*devicetypes.CaniInterface) bool) ([]interfaceTarget, error) {
	var targets []interfaceTarget
	for _, iface := range inv.Interfaces {
		if iface == nil || !owns(iface) {
			continue
		}
		// Use matchInterfaceName so that '*' can match '/' characters
		// in interface names like "1/1/14".
		matched, merr := matchInterfaceName(namePattern, iface.Name)
		if merr != nil {
			return nil, fmt.Errorf("invalid --name pattern %q: %w", namePattern, merr)
		}
		if !matched && !strings.EqualFold(iface.Name, namePattern) {
			continue
		}
		spec := findInterfaceSpec(inv, iface)
		targets = append(targets, interfaceTarget{instance: iface, spec: spec})
	}
	return targets, nil
}

// belongsToDevice returns true if the interface belongs to the given device
// directly or via one of its child modules.
func belongsToDevice(inv *devicetypes.Inventory, iface *devicetypes.CaniInterface, deviceID uuid.UUID) bool {
	if iface.DeviceID == deviceID {
		return true
	}
	// Check if the interface's device is actually a module parented by this device
	for _, mod := range inv.Modules {
		if mod == nil {
			continue
		}
		if mod.ID == iface.DeviceID && mod.ParentDevice == deviceID {
			return true
		}
	}
	return false
}

// listDeviceInterfaces prints all interfaces for a device (and its modules).
func listDeviceInterfaces(cmd *cobra.Command, inv *devicetypes.Inventory) error {
	deviceRef, _ := cmd.Flags().GetString("device")
	if deviceRef == "" {
		return fmt.Errorf("--device is required with -L/--list")
	}

	deviceID, err := resolve.Device(inv, deviceRef)
	if err != nil {
		return fmt.Errorf("resolving --device: %w", err)
	}
	device := inv.Devices[deviceID]

	fmt.Printf("Interfaces for %s (%s):\n", device.Name, deviceID)
	fmt.Printf("  %-20s %-24s %-20s %s\n", "NAME", "TYPE", "ROLE", "SOURCE")
	fmt.Printf("  %-20s %-24s %-20s %s\n", "----", "----", "----", "------")

	// Device's own interfaces
	for _, iface := range device.Interfaces {
		mgmtOnly := iface.MgmtOnly != nil && *iface.MgmtOnly
		role := devicetypes.ResolveInterfaceRole(iface.Role, iface.Name, iface.Type, mgmtOnly)
		if role == "" {
			role = "-"
		}
		fmt.Printf("  %-20s %-24s %-20s %s\n", iface.Name, iface.Type, role, "device")
	}

	// Module interfaces
	for _, mod := range inv.Modules {
		if mod == nil || mod.ParentDevice != deviceID {
			continue
		}
		for _, iface := range mod.Interfaces {
			mgmtOnly := iface.MgmtOnly != nil && *iface.MgmtOnly
			role := devicetypes.ResolveInterfaceRole(iface.Role, iface.Name, iface.Type, mgmtOnly)
			if role == "" {
				role = "-"
			}
			fmt.Printf("  %-20s %-24s %-20s module:%s\n", iface.Name, iface.Type, role, mod.Name)
		}
	}

	return nil
}

// matchInterfaceName performs glob matching where '*' can match '/' characters.
// Standard filepath.Match treats '/' as a path separator that '*' cannot cross;
// this function works around that for interface names like "1/1/14".
func matchInterfaceName(pattern, name string) (bool, error) {
	p := strings.ReplaceAll(pattern, "/", "\x00")
	n := strings.ReplaceAll(name, "/", "\x00")
	return filepath.Match(p, n)
}

// findInterfaceSpec locates the InterfaceSpec on the parent device/module
// that corresponds to the given CaniInterface.
func findInterfaceSpec(inv *devicetypes.Inventory, iface *devicetypes.CaniInterface) *devicetypes.InterfaceSpec {
	// Check device interfaces
	if device, ok := inv.Devices[iface.DeviceID]; ok && device != nil {
		for i := range device.Interfaces {
			if device.Interfaces[i].Name == iface.Name {
				return &device.Interfaces[i]
			}
			if device.Interfaces[i].ID == iface.ID {
				return &device.Interfaces[i]
			}
		}
	}
	// Check module interfaces
	if mod, ok := inv.Modules[iface.DeviceID]; ok && mod != nil {
		for i := range mod.Interfaces {
			if mod.Interfaces[i].Name == iface.Name {
				return &mod.Interfaces[i]
			}
			if mod.Interfaces[i].ID == iface.ID {
				return &mod.Interfaces[i]
			}
		}
	}
	return nil
}
