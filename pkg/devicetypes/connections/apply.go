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
)

// ApplyConnections creates cables in the inventory from a list of
// resolved connections. Returns the number of cables created and any
// per-cable errors encountered.
func ApplyConnections(resolved []ResolvedConnection, inv *devicetypes.Inventory) (int, []error) {
	var errs []error
	created := 0

	for _, conn := range resolved {
		cable := devicetypes.NewCable(conn.Cable.Type, conn.Cable.Label)
		cable.TerminationADevice = conn.ADevice
		cable.TerminationAPort = conn.APort
		cable.TerminationBDevice = conn.BDevice
		cable.TerminationBPort = conn.BPort

		if conn.Cable.Color != "" {
			cable.Color = conn.Cable.Color
		}
		if conn.Cable.Length != nil {
			cable.Length = conn.Cable.Length
		}
		if conn.Cable.LengthUnit != "" {
			cable.LengthUnit = conn.Cable.LengthUnit
		}
		if conn.Cable.Status != "" {
			cable.Status = conn.Cable.Status
		}

		if err := inv.AddCable(cable); err != nil {
			errs = append(errs, fmt.Errorf("cable %s->%s: %w", conn.APort, conn.BPort, err))
			continue
		}
		created++

		if conn.AMac != "" {
			if err := inv.SetInterfaceMACByID(conn.ADevice, conn.APort, conn.AMac); err != nil {
				errs = append(errs, fmt.Errorf("set mac on %s: %w", conn.APort, err))
			}
		}
		if conn.BMac != "" {
			if err := inv.SetInterfaceMACByID(conn.BDevice, conn.BPort, conn.BMac); err != nil {
				errs = append(errs, fmt.Errorf("set mac on %s: %w", conn.BPort, err))
			}
		}
	}

	return created, errs
}
