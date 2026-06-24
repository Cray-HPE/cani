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
	"fmt"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// endpointKey uniquely identifies one physical termination: the device or
// module UUID together with its port name. A physical port hosts at most one
// cable, so two resolved connections that share an endpointKey conflict.
type endpointKey struct {
	device uuid.UUID
	port   string
}

// InterfaceConflict records a resolved connection that was dropped because one
// of its endpoints was already claimed by an earlier (winning) connection.
type InterfaceConflict struct {
	Dropped        ResolvedConnection // the connection that was discarded
	Winner         ResolvedConnection // the earlier connection holding the port
	ConflictDevice uuid.UUID          // device or module owning the contested port
	ConflictPort   string             // the contested port name
}

// FilterInterfaceConflicts removes resolved connections that would place a
// second cable on an interface already used by an earlier connection. A
// physical port hosts at most one cable, so when two connections share a
// (device, port) endpoint only the first in input order is kept; the rest are
// returned as conflicts. Input order is deterministic (it follows the
// connection-map file order), so the kept connection is predictable.
func FilterInterfaceConflicts(resolved []ResolvedConnection) (kept []ResolvedConnection, dropped []InterfaceConflict) {
	claimedBy := make(map[endpointKey]ResolvedConnection, len(resolved)*2)

	for _, conn := range resolved {
		aKey := endpointKey{device: conn.ADevice, port: conn.APort}
		bKey := endpointKey{device: conn.BDevice, port: conn.BPort}

		if winner, ok := claimedBy[aKey]; ok {
			dropped = append(dropped, InterfaceConflict{
				Dropped:        conn,
				Winner:         winner,
				ConflictDevice: conn.ADevice,
				ConflictPort:   conn.APort,
			})
			continue
		}
		if winner, ok := claimedBy[bKey]; ok {
			dropped = append(dropped, InterfaceConflict{
				Dropped:        conn,
				Winner:         winner,
				ConflictDevice: conn.BDevice,
				ConflictPort:   conn.BPort,
			})
			continue
		}

		claimedBy[aKey] = conn
		claimedBy[bKey] = conn
		kept = append(kept, conn)
	}

	return kept, dropped
}

// Describe renders a human-readable explanation of a dropped cable, resolving
// device and module UUIDs to names via the inventory.
func (c InterfaceConflict) Describe(inv *devicetypes.Inventory) string {
	return fmt.Sprintf(
		"dropping cable %s:%s <-> %s:%s — %s:%s is already cabled by %s:%s <-> %s:%s",
		connectableName(c.Dropped.ADevice, inv), c.Dropped.APort,
		connectableName(c.Dropped.BDevice, inv), c.Dropped.BPort,
		connectableName(c.ConflictDevice, inv), c.ConflictPort,
		connectableName(c.Winner.ADevice, inv), c.Winner.APort,
		connectableName(c.Winner.BDevice, inv), c.Winner.BPort,
	)
}

// connectableName resolves a device or module UUID to its name, falling back to
// the UUID string when the entity is absent or unnamed.
func connectableName(id uuid.UUID, inv *devicetypes.Inventory) string {
	if id == uuid.Nil {
		return "(none)"
	}
	if inv != nil {
		if dev, ok := inv.Devices[id]; ok && dev != nil && dev.Name != "" {
			return dev.Name
		}
		if mod, ok := inv.Modules[id]; ok && mod != nil && mod.Name != "" {
			return mod.Name
		}
	}
	return id.String()
}
