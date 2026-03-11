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
package devicetypes

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
)

// PlanAssignment records a single parent assignment for an orphan.
type PlanAssignment struct {
	OrphanID     uuid.UUID `json:"orphan_id"`
	OrphanName   string    `json:"orphan_name"`
	OrphanKind   string    `json:"orphan_kind"` // "device" or "rack"
	ParentID     uuid.UUID `json:"parent_id"`
	ParentName   string    `json:"parent_name"`
	ParentKind   string    `json:"parent_kind"`             // "rack" or "location"
	RackPosition int       `json:"rack_position,omitempty"` // U-slot for device→rack placement
	Face         string    `json:"face,omitempty"`          // front, rear, or full
}

// ResolvePlan is a list of parent assignments produced by a dry-run.
// It can be saved to a file, edited, and applied later.
type ResolvePlan struct {
	Assignments []PlanAssignment `json:"assignments"`
}

// WritePlan serialises a plan to path as indented JSON.
func WritePlan(path string, plan *ResolvePlan) error {
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding plan: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing plan file: %w", err)
	}
	return nil
}

// ReadPlan loads a plan from a JSON file.
func ReadPlan(path string) (*ResolvePlan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading plan file: %w", err)
	}
	var plan ResolvePlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("parsing plan file: %w", err)
	}
	return &plan, nil
}

// ApplyPlan applies every assignment in the plan to the inventory.
// It sets Parent (for devices) or Location (for racks), then rebuilds
// relationships. Returns the list of changes applied.
func ApplyPlan(inv *Inventory, plan *ResolvePlan) ([]string, error) {
	var changes []string

	for _, a := range plan.Assignments {
		switch a.OrphanKind {
		case "device":
			dev, ok := inv.Devices[a.OrphanID]
			if !ok {
				return changes, fmt.Errorf("device %s (%s) not found in inventory", a.OrphanID, a.OrphanName)
			}
			dev.Parent = a.ParentID

			// Place device in rack at the recorded position (or next available).
			if rack, ok := inv.Racks[a.ParentID]; ok {
				PlaceDeviceInRack(dev, a.OrphanID, rack, a.RackPosition, a.Face)
			}

			changes = append(changes, fmt.Sprintf("device %q → parent %s (%s) U%d",
				a.OrphanName, a.ParentID, a.ParentName, dev.RackPosition))

		case "rack":
			rack, ok := inv.Racks[a.OrphanID]
			if !ok {
				return changes, fmt.Errorf("rack %s (%s) not found in inventory", a.OrphanID, a.OrphanName)
			}
			rack.Location = a.ParentID
			changes = append(changes, fmt.Sprintf("rack %q → location %s (%s)",
				a.OrphanName, a.ParentID, a.ParentName))

		default:
			return changes, fmt.Errorf("unknown orphan kind %q for %s", a.OrphanKind, a.OrphanID)
		}
	}

	result := inv.VerifyParentChildRelationships()
	if result.HasErrors() {
		return changes, fmt.Errorf("relationship errors after applying plan: %v", result.Errors)
	}

	return changes, nil
}

// PlaceDeviceInRack places a device into a rack at the given position.
// If startU is 0, it finds the next available slot. Sets the device's
// RackPosition and Face fields.
func PlaceDeviceInRack(dev *CaniDeviceType, devID uuid.UUID, rack *CaniRackType, startU int, face string) {
	height := dev.UHeight
	if height < 1 {
		height = 1
	}
	if face == "" {
		face = FaceFront
	}
	isFullDepth := dev.IsFullDepth

	// Use the requested position, or find the next available one.
	if startU <= 0 {
		startU = rack.FindNextAvailableSlot(height, face, isFullDepth)
	}
	if startU > 0 {
		rack.PlaceDevice(devID, startU, height, face, isFullDepth)
		dev.RackPosition = startU
		dev.Face = face
	}
}
