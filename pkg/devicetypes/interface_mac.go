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
	"fmt"
	"net"
	"strings"

	"github.com/google/uuid"
)

// NormalizeMAC validates a MAC address and returns it in canonical
// lowercase, colon-separated form (e.g. "aa:bb:cc:dd:ee:ff"). Any format
// accepted by net.ParseMAC (colon, hyphen, or dotted notation) is allowed
// on input. An empty (or whitespace-only) value returns an empty string
// with no error so callers may treat "unset" as valid.
func NormalizeMAC(mac string) (string, error) {
	trimmed := strings.TrimSpace(mac)
	if trimmed == "" {
		return "", nil
	}
	hw, err := net.ParseMAC(trimmed)
	if err != nil {
		return "", fmt.Errorf("invalid MAC address %q: %w", mac, err)
	}
	return hw.String(), nil
}

// SetInterfaceMAC validates and normalizes mac, then assigns it to the named
// interface of the device or module referenced by name or UUID. It is a
// convenience wrapper around SetInterfaceMACByID for callers that work with
// names (e.g. CSV import).
func (inv *Inventory) SetInterfaceMAC(ownerRef, ifaceName, mac string) error {
	ownerID := inv.FindConnectableByNameOrID(ownerRef)
	if ownerID == uuid.Nil {
		return fmt.Errorf("device or module %q not found", ownerRef)
	}
	return inv.SetInterfaceMACByID(ownerID, ifaceName, mac)
}

// SetInterfaceMACByID validates and normalizes mac, then assigns it to the
// named interface owned by the given device or module ID. The MAC is written
// to the persistent InterfaceSpec (the source of truth) and mirrored onto the
// rebuilt CaniInterface when one is already indexed.
func (inv *Inventory) SetInterfaceMACByID(ownerID uuid.UUID, ifaceName, mac string) error {
	normalized, err := NormalizeMAC(mac)
	if err != nil {
		return err
	}

	spec := inv.findInterfaceSpecOnOwner(ownerID, ifaceName)
	if spec == nil {
		return fmt.Errorf("interface %q not found on %s", ifaceName, ownerID)
	}
	spec.MacAddress = normalized

	if spec.ID != uuid.Nil {
		if inst, ok := inv.Interfaces[spec.ID]; ok && inst != nil {
			inst.MacAddress = normalized
		}
	}
	return nil
}

// findInterfaceSpecOnOwner returns the InterfaceSpec with the given name on
// the device or module identified by ownerID, or nil if not found.
func (inv *Inventory) findInterfaceSpecOnOwner(ownerID uuid.UUID, ifaceName string) *InterfaceSpec {
	if dev, ok := inv.Devices[ownerID]; ok && dev != nil {
		for i := range dev.Interfaces {
			if dev.Interfaces[i].Name == ifaceName {
				return &dev.Interfaces[i]
			}
		}
	}
	if mod, ok := inv.Modules[ownerID]; ok && mod != nil {
		for i := range mod.Interfaces {
			if mod.Interfaces[i].Name == ifaceName {
				return &mod.Interfaces[i]
			}
		}
	}
	return nil
}
