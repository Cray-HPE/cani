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
	"strconv"
	"strings"

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

// interfaceUpdates holds the validated interface field values.
type interfaceUpdates struct {
	role         string
	label        string
	mac          string
	lag          string
	mode         string
	untaggedVLAN int
	taggedVLANs  []int
	vrf          string
}

// registeredInterfaceRoles returns the set of role names registered in the
// inventory metadata catalog for the dcim.interface content type. These are
// valid custom interface roles and must not trigger an unknown-role warning.
func registeredInterfaceRoles(inventory *devicetypes.Inventory) map[string]bool {
	registered := make(map[string]bool)
	for _, r := range inventory.ListMetadata("roles") {
		for _, ct := range r.ContentTypes {
			if ct == "dcim.interface" {
				registered[r.Name] = true
				break
			}
		}
	}
	return registered
}

// interfaceUpdateFlags lists every flag that mutates an interface field.
var interfaceUpdateFlags = []string{
	"role", "label", "mac", "lag", "mode", "untagged-vlan", "tagged-vlan", "vrf",
}

// anyInterfaceFlagChanged reports whether at least one mutating flag was set.
func anyInterfaceFlagChanged(cmd *cli.Command) bool {
	for _, f := range interfaceUpdateFlags {
		if cmd.Flags().Changed(f) {
			return true
		}
	}
	return false
}

// parseInterfaceUpdates reads and validates the interface update flags.
func parseInterfaceUpdates(cmd *cli.Command, inventory *devicetypes.Inventory) (interfaceUpdates, error) {
	u := interfaceUpdates{}
	u.role, _ = cmd.Flags().GetString("role")
	u.label, _ = cmd.Flags().GetString("label")
	u.mac, _ = cmd.Flags().GetString("mac")
	u.lag, _ = cmd.Flags().GetString("lag")
	u.mode, _ = cmd.Flags().GetString("mode")
	u.untaggedVLAN, _ = cmd.Flags().GetInt("untagged-vlan")
	u.vrf, _ = cmd.Flags().GetString("vrf")

	if !anyInterfaceFlagChanged(cmd) {
		return interfaceUpdates{}, fmt.Errorf("at least one interface field flag must be specified (e.g. --role, --mac, --lag, --mode, --untagged-vlan, --tagged-vlan, --vrf)")
	}

	if u.role != "" {
		registered := registeredInterfaceRoles(inventory)
		if warn := devicetypes.ValidateInterfaceRoleWithRegistered(u.role, registered); warn != "" {
			log.Printf("Warning: %s", warn)
		}
	}

	return validateInterfaceUpdates(cmd, u)
}

// validateInterfaceUpdates normalizes and range-checks the MAC, mode, and VLAN
// values, returning the updated struct or the first validation error.
func validateInterfaceUpdates(cmd *cli.Command, u interfaceUpdates) (interfaceUpdates, error) {
	if cmd.Flags().Changed("mac") {
		normalized, err := devicetypes.NormalizeMAC(u.mac)
		if err != nil {
			return interfaceUpdates{}, err
		}
		u.mac = normalized
	}
	if cmd.Flags().Changed("mode") {
		normalized, err := devicetypes.ValidateInterfaceMode(u.mode)
		if err != nil {
			return interfaceUpdates{}, err
		}
		u.mode = normalized
	}
	if u.untaggedVLAN != 0 {
		if err := devicetypes.ValidateVID(u.untaggedVLAN); err != nil {
			return interfaceUpdates{}, err
		}
	}
	taggedStrs, _ := cmd.Flags().GetStringSlice("tagged-vlan")
	tagged, err := parseVIDs(taggedStrs)
	if err != nil {
		return interfaceUpdates{}, err
	}
	u.taggedVLANs = tagged
	return u, nil
}

// parseVIDs converts a list of VLAN-ID strings into validated integers.
func parseVIDs(strs []string) ([]int, error) {
	if len(strs) == 0 {
		return nil, nil
	}
	vids := make([]int, 0, len(strs))
	for _, s := range strs {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		vid, cerr := strconv.Atoi(s)
		if cerr != nil {
			return nil, fmt.Errorf("invalid tagged VLAN %q: must be an integer", s)
		}
		if verr := devicetypes.ValidateVID(vid); verr != nil {
			return nil, verr
		}
		vids = append(vids, vid)
	}
	return vids, nil
}

// interfaceFieldChanges records which interface fields the user asked to update.
type interfaceFieldChanges struct {
	role, label, mac, lag, mode, untagged, tagged, vrf bool
}

// changedInterfaceFields snapshots which mutating flags were set on the command.
func changedInterfaceFields(cmd *cli.Command) interfaceFieldChanges {
	return interfaceFieldChanges{
		role:     cmd.Flags().Changed("role"),
		label:    cmd.Flags().Changed("label"),
		mac:      cmd.Flags().Changed("mac"),
		lag:      cmd.Flags().Changed("lag"),
		mode:     cmd.Flags().Changed("mode"),
		untagged: cmd.Flags().Changed("untagged-vlan"),
		tagged:   cmd.Flags().Changed("tagged-vlan"),
		vrf:      cmd.Flags().Changed("vrf"),
	}
}

// applyInterfaceUpdates applies the changed fields to each target interface
// (and its backing spec when present).
func applyInterfaceUpdates(cmd *cli.Command, targets []interfaceTarget, u interfaceUpdates) {
	changed := changedInterfaceFields(cmd)
	for _, t := range targets {
		applyInterfaceFields(t, u, changed)
	}
}

// applyInterfaceFields writes each requested field onto one target.
func applyInterfaceFields(t interfaceTarget, u interfaceUpdates, c interfaceFieldChanges) {
	if c.role {
		setInterfaceRole(t, u.role)
	}
	if c.label {
		setInterfaceLabel(t, u.label)
	}
	if c.mac {
		setInterfaceMAC(t, u.mac)
	}
	if c.lag {
		setInterfaceLag(t, u.lag)
	}
	if c.mode {
		setInterfaceMode(t, u.mode)
	}
	if c.untagged {
		setInterfaceUntaggedVLAN(t, u.untaggedVLAN)
	}
	if c.tagged {
		setInterfaceTaggedVLANs(t, u.taggedVLANs)
	}
	if c.vrf {
		setInterfaceVRF(t, u.vrf)
	}
}

// setInterfaceRole sets the role on the instance and its spec when present.
func setInterfaceRole(t interfaceTarget, role string) {
	t.instance.Role = role
	if t.spec != nil {
		t.spec.Role = role
	}
}

// setInterfaceLabel sets the label on the instance and its spec when present.
func setInterfaceLabel(t interfaceTarget, label string) {
	t.instance.Label = label
	if t.spec != nil {
		t.spec.Label = label
	}
}

// setInterfaceMAC sets the MAC on the instance and its spec when present.
func setInterfaceMAC(t interfaceTarget, mac string) {
	t.instance.MacAddress = mac
	if t.spec != nil {
		t.spec.MacAddress = mac
	}
}

// setInterfaceLag sets the parent LAG on the instance and its spec when present.
func setInterfaceLag(t interfaceTarget, lag string) {
	t.instance.Lag = lag
	if t.spec != nil {
		t.spec.Lag = lag
	}
}

// setInterfaceMode sets the switchport mode on the instance and its spec.
func setInterfaceMode(t interfaceTarget, mode string) {
	t.instance.Mode = mode
	if t.spec != nil {
		t.spec.Mode = mode
	}
}

// setInterfaceUntaggedVLAN sets the untagged VLAN on the instance and its spec.
func setInterfaceUntaggedVLAN(t interfaceTarget, vid int) {
	t.instance.UntaggedVLAN = vid
	if t.spec != nil {
		t.spec.UntaggedVLAN = vid
	}
}

// setInterfaceTaggedVLANs sets the tagged VLANs on the instance and its spec.
func setInterfaceTaggedVLANs(t interfaceTarget, vids []int) {
	t.instance.TaggedVLANs = vids
	if t.spec != nil {
		t.spec.TaggedVLANs = vids
	}
}

// setInterfaceVRF sets the VRF on the instance and its spec when present.
func setInterfaceVRF(t interfaceTarget, vrf string) {
	t.instance.VRF = vrf
	if t.spec != nil {
		t.spec.VRF = vrf
	}
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
