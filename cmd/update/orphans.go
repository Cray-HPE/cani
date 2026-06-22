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
	"time"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func newOrphansCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "orphans",
		Short: "Interactively assign parents to orphaned items.",
		Long: `Walk all orphaned racks and devices and interactively prompt
for a parent assignment. Candidates are ranked by name similarity,
hardware-type affinity, and provider metadata.

In --dry-run mode the assignments are written to a JSON plan file
that can be reviewed, edited, and applied later with --apply-plan.`,
		Args: cli.NoArgs,
		RunE: updateOrphans,
	}

	cmd.Flags().Bool("dry-run", false, "Preview changes and save to a plan file without modifying inventory")
	cmd.Flags().String("apply-plan", "", "Apply a previously saved plan file instead of prompting")

	return cmd
}

func updateOrphans(cmd *cli.Command, args []string) error {
	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	applyPath, _ := cmd.Flags().GetString("apply-plan")

	// --apply-plan: load plan and apply without prompting.
	if applyPath != "" {
		return applyOrphanPlanFromFile(cmd, inventory, applyPath)
	}

	orphanRacks := inventory.OrphanRacks()
	orphanDevices := inventory.OrphanDevices()

	total := len(orphanRacks) + len(orphanDevices)
	if total == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No orphans found.")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Found %d orphan rack(s), %d orphan device(s)\n\n",
		len(orphanRacks), len(orphanDevices))

	opts := devicetypes.ClassifyOptions{
		Writer: cmd.OutOrStdout(),
		Reader: cmd.InOrStdin(),
	}

	var plan devicetypes.ResolvePlan

	// Resolve orphan racks first (devices depend on racks).
	for _, orphan := range orphanRacks {
		parentID, parentName, parentKind, err := promptOrphanParent(inventory, orphan, opts)
		if err != nil {
			return err
		}
		if parentID == uuid.Nil {
			continue
		}
		rack := inventory.Racks[orphan.ID]
		if rack != nil {
			rack.Location = parentID
		}
		plan.Assignments = append(plan.Assignments, devicetypes.PlanAssignment{
			OrphanID:   orphan.ID,
			OrphanName: orphan.Name,
			OrphanKind: orphan.Kind,
			ParentID:   parentID,
			ParentName: parentName,
			ParentKind: parentKind,
		})
	}

	// Resolve orphan devices. Mutate inventory in-place so rack
	// occupancy counts stay accurate for subsequent prompts.
	for _, orphan := range orphanDevices {
		parentID, parentName, parentKind, err := promptOrphanParent(inventory, orphan, opts)
		if err != nil {
			return err
		}
		if parentID == uuid.Nil {
			continue
		}
		device := inventory.Devices[orphan.ID]
		if device != nil {
			device.Parent = parentID
		}

		// Auto-place the device in the rack at the next available U-slot.
		var rackPos int
		var face string
		if device != nil && parentKind == "rack" {
			if rack, ok := inventory.Racks[parentID]; ok {
				devicetypes.PlaceDeviceInRack(device, orphan.ID, rack, 0, "")
				rackPos = device.RackPosition
				face = device.Face
			}
		}

		plan.Assignments = append(plan.Assignments, devicetypes.PlanAssignment{
			OrphanID:     orphan.ID,
			OrphanName:   orphan.Name,
			OrphanKind:   orphan.Kind,
			ParentID:     parentID,
			ParentName:   parentName,
			ParentKind:   parentKind,
			RackPosition: rackPos,
			Face:         face,
		})
	}

	if len(plan.Assignments) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "\nNo changes made.")
		return nil
	}

	// Rebuild all relationships after assignments.
	result := inventory.VerifyParentChildRelationships()
	if result.HasErrors() {
		return fmt.Errorf("relationship errors after reparenting: %v", result.Errors)
	}

	printOrphanSummary(cmd, &plan)

	if dryRun {
		planPath := orphanPlanPath()
		if err := devicetypes.WritePlan(planPath, &plan); err != nil {
			return fmt.Errorf("writing plan file: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "\n(dry-run: no changes saved)\n")
		fmt.Fprintf(cmd.OutOrStdout(), "Plan saved to %s\n", planPath)
		fmt.Fprintf(cmd.OutOrStdout(), "Apply with: cani alpha update orphans --apply-plan %s\n", planPath)
		return nil
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("Resolved %d orphan(s)", len(plan.Assignments))
	return nil
}

// applyOrphanPlanFromFile loads a plan file and applies it to the inventory.
func applyOrphanPlanFromFile(cmd *cli.Command, inv *devicetypes.Inventory, path string) error {
	plan, err := devicetypes.ReadPlan(path)
	if err != nil {
		return err
	}

	changes, err := devicetypes.ApplyPlan(inv, plan)
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Applied %d assignment(s) from %s\n", len(changes), path)
	for _, c := range changes {
		fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", c)
	}

	if err := datastores.Datastore.Save(inv); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("Applied plan with %d assignment(s)", len(changes))
	return nil
}

// promptOrphanParent shows suggestions and prompts the user for a parent choice.
// Returns the selected parent's UUID, name, and kind.
func promptOrphanParent(inv *devicetypes.Inventory, orphan devicetypes.OrphanItem, opts devicetypes.ClassifyOptions) (uuid.UUID, string, string, error) {
	suggestions := devicetypes.SuggestParents(inv, orphan, 8)
	id, err := devicetypes.PromptForParent(inv, orphan, suggestions, opts)
	if err != nil {
		return uuid.Nil, "", "", err
	}
	if id == uuid.Nil {
		return uuid.Nil, "", "", nil
	}
	name, kind := lookupParentInfo(inv, id)
	return id, name, kind, nil
}

// lookupParentInfo returns the name and kind of a parent by ID.
func lookupParentInfo(inv *devicetypes.Inventory, id uuid.UUID) (string, string) {
	if rack, ok := inv.Racks[id]; ok && rack != nil {
		return rack.Name, "rack"
	}
	if loc, ok := inv.Locations[id]; ok && loc != nil {
		return loc.Name, "location"
	}
	return id.String(), "unknown"
}

// printOrphanSummary writes the assignment summary to stdout.
func printOrphanSummary(cmd *cli.Command, plan *devicetypes.ResolvePlan) {
	fmt.Fprintf(cmd.OutOrStdout(), "\n── Summary ──\n")
	for _, a := range plan.Assignments {
		pos := ""
		if a.RackPosition > 0 {
			pos = fmt.Sprintf(" @ U%d %s", a.RackPosition, a.Face)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  %s %q → %s %q%s\n",
			a.OrphanKind, a.OrphanName, a.ParentKind, a.ParentName, pos)
	}
}

// orphanPlanPath returns a timestamped plan file path next to the inventory.
func orphanPlanPath() string {
	dir := filepath.Dir(config.Cfg.Path)
	ts := time.Now().Format("2006-01-02_150405")
	return filepath.Join(dir, fmt.Sprintf("resolve-plan_%s.json", ts))
}
