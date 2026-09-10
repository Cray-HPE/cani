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
package transform

import (
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// BuildVLANLocationMap indexes each VLAN by its first valid location
// assignment. CANI currently models one location per VLAN.
func BuildVLANLocationMap(assignments []nautobotapi.VLANLocationAssignment) map[uuid.UUID]uuid.UUID {
	locations := make(map[uuid.UUID]uuid.UUID)
	for _, assignment := range assignments {
		vlanID := refID(assignment.Vlan)
		locationID := refID(assignment.Location)
		if vlanID == uuid.Nil || locationID == uuid.Nil {
			continue
		}
		if _, exists := locations[vlanID]; !exists {
			locations[vlanID] = locationID
		}
	}
	return locations
}

// BuildPrefixLocationMap indexes each prefix by its first valid location
// assignment. CANI currently models one location per prefix.
func BuildPrefixLocationMap(assignments []nautobotapi.PrefixLocationAssignment) map[uuid.UUID]uuid.UUID {
	locations := make(map[uuid.UUID]uuid.UUID)
	for _, assignment := range assignments {
		prefixID := refID(assignment.Prefix)
		locationID := refID(assignment.Location)
		if prefixID == uuid.Nil || locationID == uuid.Nil {
			continue
		}
		if _, exists := locations[prefixID]; !exists {
			locations[prefixID] = locationID
		}
	}
	return locations
}

func resolveAssignedLocation(
	objectID uuid.UUID,
	assignedLocations map[uuid.UUID]uuid.UUID,
	locationMap map[uuid.UUID]uuid.UUID,
) uuid.UUID {
	return locationMap[assignedLocations[objectID]]
}
