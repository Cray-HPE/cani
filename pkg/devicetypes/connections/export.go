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
package connections

import (
	"sort"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// InventoryToConnectionMap converts inventory cables into a ConnectionMap.
// Output is sorted by cable UUID for deterministic results.
func InventoryToConnectionMap(inv *devicetypes.Inventory) ConnectionMap {
	cm := ConnectionMap{Version: "v1"}

	// Sort cable IDs for deterministic output.
	cableIDs := make([]uuid.UUID, 0, len(inv.Cables))
	for id := range inv.Cables {
		cableIDs = append(cableIDs, id)
	}
	sort.Slice(cableIDs, func(i, j int) bool {
		return cableIDs[i].String() < cableIDs[j].String()
	})

	for _, id := range cableIDs {
		cable := inv.Cables[id]
		if cable == nil {
			continue
		}

		entry := ConnectionEntry{
			A: Endpoint{
				Device: resolveDeviceName(cable.TerminationADevice, inv),
				Port:   cable.TerminationAPort,
			},
			B: Endpoint{
				Device: resolveDeviceName(cable.TerminationBDevice, inv),
				Port:   cable.TerminationBPort,
			},
		}

		// Only emit cable props when they differ from empty defaults.
		if cable.Slug != "" || cable.Label != "" || cable.Color != "" || cable.Length != nil {
			entry.Cable = &CableProps{
				Type:       cable.Slug,
				Label:      cable.Label,
				Color:      cable.Color,
				Length:     cable.Length,
				LengthUnit: cable.LengthUnit,
			}
		}

		cm.Connections = append(cm.Connections, entry)
	}

	return cm
}

// resolveDeviceName returns the device name for a UUID, falling back to
// the UUID string if the device is not in the inventory.
func resolveDeviceName(id uuid.UUID, inv *devicetypes.Inventory) string {
	if id == uuid.Nil {
		return ""
	}
	if dev, ok := inv.Devices[id]; ok && dev.Name != "" {
		return dev.Name
	}
	return id.String()
}
