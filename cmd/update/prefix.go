/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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

// newPrefixCommand creates the "update prefix" subcommand.
func newPrefixCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "prefix <uuid | [namespace/]cidr>",
		Short: "Update a prefix in the inventory.",
		Long: `Update a prefix's VRF memberships or description.

The prefix is matched by UUID, by CIDR within --namespace (default Global), or
by a qualified "namespace/CIDR". VRFs are matched by UUID or by name within the
prefix's namespace.

A datastore migrated from an older schema can leave a legacy VRF name on a
prefix that did not resolve to one VRF. Adding a VRF with --vrf replaces it;
--remove-vrf with that name clears it.

Examples:
  cani alpha update prefix 10.0.0.0/24 --vrf blue
  cani alpha update prefix tenant-a/10.0.0.0/24 --remove-vrf red
  cani alpha update prefix 10.0.0.0/24 --description "Management"`,
		Args: cli.ExactArgs(1),
		RunE: updatePrefix,
	}

	cmd.Flags().String(flagNamespace, devicetypes.DefaultNamespace, "IPAM namespace to search when the CIDR is unqualified")
	cmd.Flags().String(flagDescription, "", "New description")
	cmd.Flags().StringArray(flagVRF, nil, "Add the prefix to a VRF (name or UUID, repeatable)")
	cmd.Flags().StringArray(flagRemoveVRF, nil, "Remove the prefix from a VRF, or clear its legacy VRF name (repeatable)")

	return cmd
}

func updatePrefix(cmd *cli.Command, args []string) error {
	if err := store.Setup(cmd); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}
	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	namespace, _ := cmd.Flags().GetString(flagNamespace)
	prefix, err := inventory.ResolvePrefixReference(namespace, args[0])
	if err != nil {
		return err
	}
	changed, err := applyPrefixUpdates(cmd, inventory, prefix)
	if err != nil {
		return err
	}
	if !changed {
		return fmt.Errorf("no changes specified")
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}
	log.Printf("Updated prefix %s in namespace %s (%s)", prefix.Prefix, prefix.EffectiveNamespace(), prefix.ID)
	return nil
}

// applyPrefixUpdates applies the description and VRF flags and reports
// whether anything changed.
func applyPrefixUpdates(cmd *cli.Command, inventory *devicetypes.Inventory, prefix *devicetypes.CaniPrefix) (bool, error) {
	changed := cmd.Flags().Changed(flagDescription)
	if changed {
		prefix.Description, _ = cmd.Flags().GetString(flagDescription)
	}
	added, err := addPrefixVRFs(cmd, inventory, prefix)
	if err != nil {
		return false, err
	}
	removed, err := removePrefixVRFs(cmd, inventory, prefix)
	if err != nil {
		return false, err
	}
	return changed || added || removed, nil
}

// addPrefixVRFs adds each --vrf membership; the model clears a legacy VRF name
// that an explicit membership supersedes.
func addPrefixVRFs(cmd *cli.Command, inventory *devicetypes.Inventory, prefix *devicetypes.CaniPrefix) (bool, error) {
	refs, _ := cmd.Flags().GetStringArray(flagVRF)
	changed := false
	for _, ref := range refs {
		vrf, err := inventory.ResolveVRFReference(prefix.EffectiveNamespace(), ref)
		if err != nil {
			return false, fmt.Errorf("resolving --%s %q: %w", flagVRF, ref, err)
		}
		added, err := inventory.AddPrefixVRF(prefix, vrf.ID)
		if err != nil {
			return false, err
		}
		changed = changed || added
	}
	return changed, nil
}

// removePrefixVRFs drops each --remove-vrf membership through the model, which
// also clears a legacy VRF name given by name and drops a UUID whose VRF is gone.
func removePrefixVRFs(cmd *cli.Command, inventory *devicetypes.Inventory, prefix *devicetypes.CaniPrefix) (bool, error) {
	refs, _ := cmd.Flags().GetStringArray(flagRemoveVRF)
	changed := false
	for _, ref := range refs {
		removed, err := inventory.RemovePrefixVRFReference(prefix, ref)
		if err != nil {
			return false, fmt.Errorf("resolving --%s %q: %w", flagRemoveVRF, ref, err)
		}
		changed = changed || removed
	}
	return changed, nil
}
