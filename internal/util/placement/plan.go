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
package placement

import (
	"fmt"
	"sort"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// PlacementEntry describes a single device placement in a rack.
type PlacementEntry struct {
	RackID      uuid.UUID
	RackName    string
	StartU      int
	Face        string
	Zone        string // "top", "middle", or "bottom"
	DeviceIndex int    // 0-based index in the batch
}

// Plan computes placement entries for qty devices across the given racks.
// Racks are sorted alphabetically by name for deterministic results.
// The function is read-only — it clones rack occupancy state internally
// so callers can commit results after review.
//
// zone overrides automatic zone selection when non-empty.
// hardwareType is used to auto-detect the zone when zone is empty.
func Plan(
	racks []*devicetypes.CaniRackType,
	height int,
	face string,
	isFullDepth bool,
	qty int,
	strategy Strategy,
	zone Zone,
	hardwareType string,
) ([]PlacementEntry, error) {
	if len(racks) == 0 {
		return nil, fmt.Errorf("no candidate racks available")
	}
	if height < 1 {
		return nil, fmt.Errorf("device height must be at least 1U")
	}
	if qty < 1 {
		return nil, fmt.Errorf("qty must be at least 1")
	}
	if face == "" {
		face = devicetypes.FaceFront
	}

	// Resolve the target zone.
	resolved := zone
	if resolved == "" {
		resolved = ZoneForHardwareType(hardwareType)
	}

	sorted := sortRacksByName(racks)
	clones := cloneRacks(sorted)

	switch strategy {
	case StrategyFill:
		return planFill(clones, height, face, isFullDepth, qty, resolved)
	case StrategySpread:
		return planSpread(clones, height, face, isFullDepth, qty, resolved)
	default:
		return nil, fmt.Errorf("unknown placement strategy %q", strategy)
	}
}

// planFill packs each rack before moving to the next.
func planFill(
	racks []rackClone,
	height int, face string, isFullDepth bool,
	qty int, zone Zone,
) ([]PlacementEntry, error) {
	var entries []PlacementEntry
	placed := 0

	for _, rc := range racks {
		bounds := ResolveZoneBounds(rc.rack)
		zr := bounds.RangeForZone(zone)

		for placed < qty {
			startU := FindSlotInZone(rc.rack, height, face, isFullDepth, zr)
			if startU == 0 {
				break // zone full in this rack, move to next
			}
			rc.rack.PlaceDevice(uuid.New(), startU, height, face, isFullDepth)
			entries = append(entries, PlacementEntry{
				RackID:      rc.id,
				RackName:    rc.name,
				StartU:      startU,
				Face:        face,
				Zone:        string(zone),
				DeviceIndex: placed,
			})
			placed++
		}
		if placed >= qty {
			break
		}
	}

	if placed < qty {
		return nil, fmt.Errorf(
			"insufficient rack capacity in %s zone: placed %d of %d devices (%dU each)",
			zone, placed, qty, height,
		)
	}
	return entries, nil
}

// planSpread distributes devices round-robin across racks.
func planSpread(
	racks []rackClone,
	height int, face string, isFullDepth bool,
	qty int, zone Zone,
) ([]PlacementEntry, error) {
	entries := make([]PlacementEntry, 0, qty)
	n := len(racks)

	// Pre-compute zone ranges per rack.
	zoneRanges := make([]URange, n)
	for i, rc := range racks {
		bounds := ResolveZoneBounds(rc.rack)
		zoneRanges[i] = bounds.RangeForZone(zone)
	}

	for i := range qty {
		idx := i % n
		rc := &racks[idx]
		startU := FindSlotInZone(rc.rack, height, face, isFullDepth, zoneRanges[idx])
		if startU == 0 {
			return nil, fmt.Errorf(
				"rack %q %s zone is full (need %dU)",
				rc.name, zone, height,
			)
		}
		rc.rack.PlaceDevice(uuid.New(), startU, height, face, isFullDepth)
		entries = append(entries, PlacementEntry{
			RackID:      rc.id,
			RackName:    rc.name,
			StartU:      startU,
			Face:        face,
			Zone:        string(zone),
			DeviceIndex: i,
		})
	}
	return entries, nil
}

// rackClone pairs identity fields with a mutable rack copy for planning.
type rackClone struct {
	id   uuid.UUID
	name string
	rack *devicetypes.CaniRackType
}

// sortRacksByName returns racks sorted alphabetically by name.
func sortRacksByName(racks []*devicetypes.CaniRackType) []*devicetypes.CaniRackType {
	sorted := make([]*devicetypes.CaniRackType, len(racks))
	copy(sorted, racks)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})
	return sorted
}

// cloneRacks creates shallow copies of racks with deep-copied OccupiedSlots
// so planning does not mutate the original inventory.
func cloneRacks(racks []*devicetypes.CaniRackType) []rackClone {
	clones := make([]rackClone, len(racks))
	for i, r := range racks {
		cp := *r
		cp.OccupiedSlots = make(map[int]map[string]uuid.UUID, len(r.OccupiedSlots))
		for u, faces := range r.OccupiedSlots {
			cp.OccupiedSlots[u] = make(map[string]uuid.UUID, len(faces))
			for face, id := range faces {
				cp.OccupiedSlots[u][face] = id
			}
		}
		clones[i] = rackClone{
			id:   r.ID,
			name: r.Name,
			rack: &cp,
		}
	}
	return clones
}
