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
	"strings"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/util/nameexpand"
	"github.com/Cray-HPE/cani/internal/util/store"
	"github.com/Cray-HPE/cani/internal/util/validate"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// newCableCommand creates the "add cable" subcommand.
func newCableCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "cable <slug-or-part-number>",
		Short: "Add cable(s) to the inventory.",
		Long:  "Add one or more cables to the inventory by slug or part number.",
		Args:  validSlugOrPartNumber(NounCable),
		RunE:  addCable,
	}

	cmd.Flags().String("a-device", "", "Termination A device UUID or name")
	cmd.Flags().String("a-port", "", "Termination A port name")
	cmd.Flags().String("b-device", "", "Termination B device UUID or name")
	cmd.Flags().String("b-port", "", "Termination B port name")
	cmd.Flags().String("label", "", "Cable label")
	cmd.Flags().String("color", "", "Cable color")
	cmd.Flags().String("name", "", "Cable name or expansion pattern")

	return cmd
}

func addCable(cmd *cli.Command, args []string) error {
	qty, _ := cmd.Flags().GetInt("qty")
	if qty < 1 {
		qty = 1
	}

	result, err := lookupBySlugOrPart(NounCable, args[0])
	if err != nil {
		return err
	}

	aDeviceArg, _ := cmd.Flags().GetString("a-device")
	aPortArg, _ := cmd.Flags().GetString("a-port")
	bDeviceArg, _ := cmd.Flags().GetString("b-device")
	bPortArg, _ := cmd.Flags().GetString("b-port")
	label, _ := cmd.Flags().GetString("label")
	color, _ := cmd.Flags().GetString("color")
	statusArg, _ := cmd.Flags().GetString("status")
	nameArg, _ := cmd.Flags().GetString("name")
	prefix, _ := cmd.Flags().GetString("prefix")
	start, _ := cmd.Flags().GetInt("start")
	padWidth, _ := cmd.Flags().GetInt("pad-width")

	// Expand brace patterns on device/port arguments.
	aDevices := expandArg(aDeviceArg)
	aPorts := expandArg(aPortArg)
	bDevices := expandArg(bDeviceArg)
	bPorts := expandArg(bPortArg)

	// Determine effective cable count from expanded patterns.
	count, err := alignCableCount(qty, len(aDevices), len(aPorts), len(bDevices), len(bPorts))
	if err != nil {
		return err
	}

	// Broadcast single-element slices to match count.
	aDevices = broadcastSlice(aDevices, count)
	aPorts = broadcastSlice(aPorts, count)
	bDevices = broadcastSlice(bDevices, count)
	bPorts = broadcastSlice(bPorts, count)

	names, err := nameexpand.ResolveNames(nameArg, prefix, start, padWidth, count)
	if err != nil {
		return fmt.Errorf("name resolution failed: %w", err)
	}

	if err := store.Setup(cmd); err != nil {
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

	// Resolve device names/UUIDs to inventory UUIDs.
	aDeviceIDs, err := resolveDeviceRefs(aDevices, inventory)
	if err != nil {
		return fmt.Errorf("termination A: %w", err)
	}
	bDeviceIDs, err := resolveDeviceRefs(bDevices, inventory)
	if err != nil {
		return fmt.Errorf("termination B: %w", err)
	}

	for i := range count {
		cable := *result.Cable
		cable.ID = uuid.New()

		if names != nil {
			cable.Label = names[i]
		} else if label != "" {
			cable.Label = label
		}
		if color != "" {
			cable.Color = color
		}

		cable.TerminationADevice = aDeviceIDs[i]
		cable.TerminationAPort = aPorts[i]
		cable.TerminationBDevice = bDeviceIDs[i]
		cable.TerminationBPort = bPorts[i]

		if statusArg != "" {
			cable.Status = statusArg
		}

		if err := inventory.AddCable(&cable); err != nil {
			return fmt.Errorf("failed to add cable: %w", err)
		}
		log.Printf("Added cable %s (%s)", cable.ID, cable.Label)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("%d cable(s) added", count)
	return nil
}

// expandArg expands a CLI argument that may contain brace patterns.
// Returns nil for empty input, a single-element slice for plain strings,
// or the expanded list for patterns like "node-{01..04}".
func expandArg(s string) []string {
	if s == "" {
		return nil
	}
	if strings.Contains(s, "{") && strings.Contains(s, "}") {
		expanded, err := nameexpand.Expand(s)
		if err == nil && len(expanded) > 0 {
			return expanded
		}
	}
	return []string{s}
}

// alignCableCount determines the cable count from expanded pattern lengths.
// All non-zero, non-1 lengths must agree (zip semantics). A single-element
// list is broadcast to match. Returns the explicit qty if no patterns expand
// to more than 1 element, or validates qty matches the pattern count.
func alignCableCount(qty int, lengths ...int) (int, error) {
	maxLen := 1
	for _, l := range lengths {
		if l > 1 {
			if maxLen > 1 && l != maxLen {
				return 0, fmt.Errorf("pattern length mismatch: got %d and %d", maxLen, l)
			}
			maxLen = l
		}
	}
	if maxLen > 1 {
		if qty > 1 && qty != maxLen {
			return 0, fmt.Errorf("--qty %d conflicts with pattern expansion count %d", qty, maxLen)
		}
		return maxLen, nil
	}
	return qty, nil
}

// broadcastSlice returns s unchanged if its length matches n, replicates
// a single element n times, or returns a zero-value slice of length n
// when s is nil (no argument provided).
func broadcastSlice(s []string, n int) []string {
	if len(s) == n {
		return s
	}
	out := make([]string, n)
	if len(s) == 1 {
		for i := range out {
			out[i] = s[0]
		}
	}
	return out
}

// resolveDeviceRefs resolves a list of device/module name-or-UUID strings to
// inventory UUIDs. Empty strings map to uuid.Nil (no device specified).
func resolveDeviceRefs(refs []string, inv *devicetypes.Inventory) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, len(refs))
	for i, ref := range refs {
		if ref == "" {
			continue
		}
		id := inv.FindConnectableByNameOrID(ref)
		if id == uuid.Nil {
			return nil, fmt.Errorf("device not found: %s", ref)
		}
		ids[i] = id
	}
	return ids, nil
}
