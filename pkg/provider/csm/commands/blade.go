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
package commands

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// wrapWithBladeHook wraps a command's RunE so that after the original
// add logic completes, any newly-added blade device gets assigned to
// the first available slot and its child nodes are staged.
func wrapWithBladeHook(cmd *cobra.Command) {
	orig := cmd.RunE
	if orig == nil {
		return
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := orig(cmd, args); err != nil {
			return err
		}
		return handleBladePostAdd(cmd)
	}
}

// handleBladePostAdd checks whether the most recently added device is
// a blade. If so, it auto-suggests a slot, assigns an xname, and
// stages the matching child nodes.
func handleBladePostAdd(cmd *cobra.Command) error {
	if datastores.Datastore == nil {
		return nil
	}

	inv, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("csm: failed to reload inventory: %w", err)
	}

	blade := findUnassignedBlade(inv)
	if blade == nil {
		return nil // no blade was added; nothing to do
	}

	auto, _ := cmd.Flags().GetBool("auto")
	if !auto {
		return nil
	}

	cab, chassis, slot, err := suggestBladeSlot(inv)
	if err != nil {
		return fmt.Errorf("csm: blade slot suggestion failed: %w", err)
	}

	log.Printf("Querying inventory to suggest cabinet, chassis, and blade for this NodeBlade")
	log.Printf("Suggested Cabinet number: %d", cab)
	log.Printf("Suggested Chassis number: %d", chassis)
	log.Printf("Suggested NodeBlade number: %d", slot)

	bladeXname := fmt.Sprintf("x%dc%ds%d", cab, chassis, slot)

	// If there is an existing placeholder blade at this slot (imported
	// from SLS with status "active"), update it in place and remove the
	// orphan device that addAnyDevice created.
	placeholder := findBladeAtXname(inv, bladeXname)
	if placeholder != nil {
		placeholder.Slug = blade.Slug
		placeholder.HardwareType = blade.HardwareType
		placeholder.Status = "staged"
		delete(inv.Devices, blade.ID)
		blade = placeholder
	} else {
		assignBladeXname(blade, bladeXname)
	}

	nodeSlug := resolveChildNodeSlug(blade.Slug)
	staged := stageChildNodes(inv, bladeXname, blade.Slug, nodeSlug)

	log.Println() // blank line between suggestion and confirmation
	log.Printf("NodeBlade was successfully staged to be added to the system")
	log.Printf("UUID: %s", blade.ID)
	log.Printf("Cabinet: %d", cab)
	log.Printf("Chassis: %d", chassis)
	log.Printf("Blade: %d", slot)
	_ = staged

	if err := datastores.Datastore.Save(inv); err != nil {
		return fmt.Errorf("csm: failed to save inventory: %w", err)
	}

	return nil
}

// findUnassignedBlade returns the first blade-type device that has no
// CSM xname assigned yet.
func findUnassignedBlade(inv *devicetypes.Inventory) *devicetypes.CaniDeviceType {
	for _, dev := range inv.Devices {
		if dev == nil {
			continue
		}
		if dev.GetType() != devicetypes.TypeBlade {
			continue
		}
		sub, ok := dev.GetProviderSubMap("csm")
		if ok {
			if xname, _ := sub["xname"].(string); xname != "" {
				continue
			}
		}
		if dev.Status == "staged" {
			return dev
		}
	}
	return nil
}

// suggestBladeSlot scans the inventory for Hill/Mountain chassis and
// finds the first empty blade slot. Returns cabinet, chassis, slot.
func suggestBladeSlot(inv *devicetypes.Inventory) (int, int, int, error) {
	type chassisInfo struct {
		cab     int
		chassis int
		xname   string
	}

	// Collect Hill/Mountain chassis (the only ones that accept compute blades).
	var allChassis []chassisInfo
	for _, dev := range inv.Devices {
		if dev == nil || dev.GetType() != devicetypes.TypeChassis {
			continue
		}
		sub, ok := dev.GetProviderSubMap("csm")
		if !ok {
			continue
		}
		xname, _ := sub["xname"].(string)
		if xname == "" {
			continue
		}
		class, _ := sub["class"].(string)
		if class != "Hill" && class != "Mountain" {
			continue
		}
		cab, ch, ok := parseChassisXname(xname)
		if !ok {
			continue
		}
		allChassis = append(allChassis, chassisInfo{
			cab: cab, chassis: ch, xname: xname,
		})
	}

	sort.Slice(allChassis, func(i, j int) bool {
		if allChassis[i].cab != allChassis[j].cab {
			return allChassis[i].cab < allChassis[j].cab
		}
		return allChassis[i].chassis < allChassis[j].chassis
	})

	// Collect blade xnames that are already occupied by user-added blades.
	// Imported placeholder blades (status "active") are available for
	// assignment; only blades explicitly added by the user (status
	// "staged") are considered occupied.
	occupiedSlots := make(map[string]bool)
	for _, dev := range inv.Devices {
		if dev == nil || dev.GetType() != devicetypes.TypeBlade {
			continue
		}
		sub, ok := dev.GetProviderSubMap("csm")
		if !ok {
			continue
		}
		xname, _ := sub["xname"].(string)
		if xname == "" {
			continue
		}
		if dev.Status == "staged" {
			occupiedSlots[xname] = true
		}
	}

	// Find the first unoccupied blade slot.
	for _, ch := range allChassis {
		for slot := 0; slot < 8; slot++ {
			bladeXname := fmt.Sprintf("%ss%d", ch.xname, slot)
			if !occupiedSlots[bladeXname] {
				return ch.cab, ch.chassis, slot, nil
			}
		}
	}

	return 0, 0, 0, fmt.Errorf("no empty blade slots found")
}

// parseChassisXname parses "x<cab>c<ch>" and returns cab, chassis, ok.
func parseChassisXname(xname string) (int, int, bool) {
	var cab, ch int
	n, _ := fmt.Sscanf(xname, "x%dc%d", &cab, &ch)
	return cab, ch, n == 2
}

// assignBladeXname sets the CSM xname on the blade device.
func assignBladeXname(blade *devicetypes.CaniDeviceType, xname string) {
	blade.SetProviderMeta("csm", "xname", xname)
}

// findBladeAtXname returns the blade device at the given xname, or nil.
func findBladeAtXname(inv *devicetypes.Inventory, xname string) *devicetypes.CaniDeviceType {
	for _, dev := range inv.Devices {
		if dev == nil || dev.GetType() != devicetypes.TypeBlade {
			continue
		}
		sub, ok := dev.GetProviderSubMap("csm")
		if !ok {
			continue
		}
		x, _ := sub["xname"].(string)
		if x == xname {
			return dev
		}
	}
	return nil
}

// resolveChildNodeSlug maps a blade slug to its child node slug.
func resolveChildNodeSlug(bladeSlug string) string {
	// Map known blade slugs to their node child slugs.
	nodeSlugMap := map[string]string{
		"hpe-crayex-ex235a-compute-blade": "hpe-crayex-ex235a-compute-node",
		"hpe-crayex-ex235n-compute-blade": "hpe-crayex-ex235n-compute-node",
		"hpe-crayex-ex254n-compute-blade": "hpe-crayex-ex254n-compute-node",
		"hpe-crayex-ex420-compute-blade":  "hpe-crayex-ex420-compute-node",
		"hpe-crayex-ex425-compute-blade":  "hpe-crayex-ex425-compute-node",
	}
	if slug, ok := nodeSlugMap[bladeSlug]; ok {
		return slug
	}
	// Fallback: replace "-compute-blade" with "-compute-node".
	return strings.Replace(bladeSlug, "-compute-blade", "-compute-node", 1)
}

// stageChildNodes finds existing phantom nodes under the given blade
// xname prefix that match the node positions defined by the blade type,
// marks them as "staged", and sets their slug to the child node slug.
// Positions without existing phantoms get new device records created.
// Returns the number of nodes staged.
func stageChildNodes(
	inv *devicetypes.Inventory,
	bladeXname string,
	bladeSlug string,
	nodeSlug string,
) int {
	nodePositions := bladeNodePositions(bladeSlug)

	// Track which positions were filled by existing phantoms.
	filled := make(map[string]bool, len(nodePositions))

	for _, dev := range inv.Devices {
		if dev == nil || dev.GetType() != devicetypes.TypeNode {
			continue
		}
		sub, ok := dev.GetProviderSubMap("csm")
		if !ok {
			continue
		}
		xname, _ := sub["xname"].(string)
		if xname == "" {
			continue
		}
		if !strings.HasPrefix(xname, bladeXname) {
			continue
		}
		suffix := xname[len(bladeXname):]
		if nodePositions[suffix] {
			dev.Status = "staged"
			dev.Slug = nodeSlug
			filled[suffix] = true
		}
	}

	// Create new nodes for positions without existing phantoms.
	for pos := range nodePositions {
		if filled[pos] {
			continue
		}
		newNode := createNewNode(inv, bladeXname+pos, nodeSlug)
		inv.Devices[newNode.ID] = newNode
		filled[pos] = true
	}

	return len(filled)
}

// createNewNode builds a new staged node device for a position that
// has no existing placeholder. The node gets an xname but no NID,
// alias, or role — those must be set via the update command.
func createNewNode(
	inv *devicetypes.Inventory,
	nodeXname string,
	nodeSlug string,
) *devicetypes.CaniDeviceType {
	id := uuid.New()

	// Derive class from the parent blade's chassis.
	nodeClass := resolveClassFromXname(inv, nodeXname)

	node := &devicetypes.CaniDeviceType{
		ID:     id,
		Name:   nodeXname,
		Slug:   nodeSlug,
		Type:   devicetypes.TypeNode,
		Status: "staged",
		ProviderMetadata: map[string]any{
			"csm": map[string]any{
				"xname": nodeXname,
				"class": nodeClass,
			},
		},
	}
	return node
}

// resolveClassFromXname looks up the chassis class for a given xname
// by scanning inventory devices for the matching chassis prefix.
func resolveClassFromXname(
	inv *devicetypes.Inventory,
	xname string,
) string {
	for _, dev := range inv.Devices {
		if dev == nil {
			continue
		}
		sub, ok := dev.GetProviderSubMap("csm")
		if !ok {
			continue
		}
		devXname, _ := sub["xname"].(string)
		if devXname == "" {
			continue
		}
		// Check if this device is an ancestor of the node.
		if strings.HasPrefix(xname, devXname) && xname != devXname {
			if cls, _ := sub["class"].(string); cls != "" {
				return cls
			}
		}
	}
	return "Hill"
}

// bladeNodePositions returns the set of bNnN suffixes (relative to
// the blade xname) that a blade type populates with compute nodes.
func bladeNodePositions(bladeSlug string) map[string]bool {
	switch bladeSlug {
	case "hpe-crayex-ex235a-compute-blade":
		// 2 node cards × 1 node each
		return map[string]bool{"b0n0": true, "b1n0": true}
	case "hpe-crayex-ex235n-compute-blade":
		// 1 node card × 2 nodes
		return map[string]bool{"b0n0": true, "b0n1": true}
	case "hpe-crayex-ex254n-compute-blade":
		// 2 node cards × 1 node each
		return map[string]bool{"b0n0": true, "b1n0": true}
	case "hpe-crayex-ex4252-compute-blade":
		// 1 node card × 4 nodes
		return map[string]bool{
			"b0n0": true, "b0n1": true,
			"b0n2": true, "b0n3": true,
		}
	default:
		// Generic: 2 node cards × 2 nodes
		return map[string]bool{
			"b0n0": true, "b0n1": true,
			"b1n0": true, "b1n1": true,
		}
	}
}
