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
package validate

import (
	"fmt"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// Status validates that s is an allowed user-selectable Nautobot status
// (case-insensitive) and returns the canonical Title-case form.
func Status(s string) (string, error) {
	ns, err := devicetypes.ValidateUserStatus(s)
	if err != nil {
		return "", err
	}
	return string(ns), nil
}

// StatusWithInventory validates s against both the builtin Nautobot
// statuses and any custom statuses registered in the inventory metadata
// catalog (via "add metadata status"). Returns the canonical name or an
// error listing all valid options.
func StatusWithInventory(s string, inv *devicetypes.Inventory) (string, error) {
	// Try builtin statuses first.
	if canonical, err := Status(s); err == nil {
		return canonical, nil
	}

	// Check custom statuses from the inventory metadata catalog.
	if inv != nil && inv.Metadata != nil {
		key := strings.ToLower(s)
		for _, entry := range inv.Metadata.Statuses {
			if strings.ToLower(entry.Name) == key {
				return entry.Name, nil
			}
		}
	}

	// Build combined error message.
	custom := customStatusNames(inv)
	builtin := devicetypes.UserStatusNames()
	if custom != "" {
		return "", fmt.Errorf("invalid status %q: must be one of [%s] or a custom status [%s]", s, builtin, custom)
	}
	return "", fmt.Errorf("invalid status %q: must be one of [%s]", s, builtin)
}

// customStatusNames returns a comma-separated list of custom status
// names from the inventory metadata catalog, or "" if none exist.
func customStatusNames(inv *devicetypes.Inventory) string {
	if inv == nil || inv.Metadata == nil || len(inv.Metadata.Statuses) == 0 {
		return ""
	}
	names := make([]string, len(inv.Metadata.Statuses))
	for i, e := range inv.Metadata.Statuses {
		names[i] = e.Name
	}
	return strings.Join(names, ", ")
}
