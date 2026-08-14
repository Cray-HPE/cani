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
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

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
	description  string
}

// interfaceField binds a mutating flag to the setter that writes its value onto
// a target interface instance and its backing spec (when present).
type interfaceField struct {
	flag  string
	apply func(t interfaceTarget, u interfaceUpdates)
}

// interfaceFields is the single source of truth for the interface-mutating
// flags and how each is applied. Adding a new updatable field is one entry.
var interfaceFields = []interfaceField{
	{"role", func(t interfaceTarget, u interfaceUpdates) {
		t.instance.Role = u.role
		if t.spec != nil {
			t.spec.Role = u.role
		}
	}},
	{"label", func(t interfaceTarget, u interfaceUpdates) {
		t.instance.Label = u.label
		if t.spec != nil {
			t.spec.Label = u.label
		}
	}},
	{"mac", func(t interfaceTarget, u interfaceUpdates) {
		t.instance.MacAddress = u.mac
		if t.spec != nil {
			t.spec.MacAddress = u.mac
		}
	}},
	{"lag", func(t interfaceTarget, u interfaceUpdates) {
		t.instance.Lag = u.lag
		if t.spec != nil {
			t.spec.Lag = u.lag
		}
	}},
	{"mode", func(t interfaceTarget, u interfaceUpdates) {
		t.instance.Mode = u.mode
		if t.spec != nil {
			t.spec.Mode = u.mode
		}
	}},
	{"untagged-vlan", func(t interfaceTarget, u interfaceUpdates) {
		t.instance.UntaggedVLAN = u.untaggedVLAN
		if t.spec != nil {
			t.spec.UntaggedVLAN = u.untaggedVLAN
		}
	}},
	{"tagged-vlan", func(t interfaceTarget, u interfaceUpdates) {
		t.instance.TaggedVLANs = u.taggedVLANs
		if t.spec != nil {
			t.spec.TaggedVLANs = u.taggedVLANs
		}
	}},
	{"vrf", func(t interfaceTarget, u interfaceUpdates) {
		t.instance.VRF = u.vrf
		if t.spec != nil {
			t.spec.VRF = u.vrf
		}
	}},
	{flagDescription, func(t interfaceTarget, u interfaceUpdates) {
		t.instance.Description = u.description
		if t.spec != nil {
			t.spec.Description = u.description
		}
	}},
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

// changedInterfaceFlags returns the set of mutating flags the user set.
func changedInterfaceFlags(cmd *cli.Command) map[string]bool {
	changed := make(map[string]bool)
	for _, f := range interfaceFields {
		if cmd.Flags().Changed(f.flag) {
			changed[f.flag] = true
		}
	}
	return changed
}

// anyInterfaceFlagChanged reports whether at least one mutating flag was set.
func anyInterfaceFlagChanged(cmd *cli.Command) bool {
	return len(changedInterfaceFlags(cmd)) > 0
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
	u.description, _ = cmd.Flags().GetString(flagDescription)

	if !anyInterfaceFlagChanged(cmd) {
		return interfaceUpdates{}, fmt.Errorf("at least one interface field flag must be specified (e.g. --role, --mac, --lag, --mode, --untagged-vlan, --tagged-vlan, --vrf, --description)")
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

// applyInterfaceUpdates applies the changed fields to each target interface
// (and its backing spec when present).
func applyInterfaceUpdates(cmd *cli.Command, targets []interfaceTarget, u interfaceUpdates) {
	changed := changedInterfaceFlags(cmd)
	for _, t := range targets {
		applyInterfaceFields(t, u, changed)
	}
}

// applyInterfaceFields writes every changed field onto one target.
func applyInterfaceFields(t interfaceTarget, u interfaceUpdates, changed map[string]bool) {
	for _, f := range interfaceFields {
		if changed[f.flag] {
			f.apply(t, u)
		}
	}
}
