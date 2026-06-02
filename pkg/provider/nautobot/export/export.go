/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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

// Package export contains Nautobot-specific export logic
package export

import (
	"fmt"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/nautobot/logcolor"
)

var clog = logcolor.New("[nautobot] ", false)

// ValidateInventory validates the inventory before export
func ValidateInventory(inv *devicetypes.Inventory) error {
	if inv == nil {
		return fmt.Errorf("inventory is nil")
	}

	if len(inv.Devices) == 0 {
		return fmt.Errorf("inventory is empty")
	}

	// Count valid devices (non-nil, non-system)
	validCount := 0
	for _, device := range inv.Devices {
		if device == nil || device.Name == "" {
			continue
		}
		if device.Type == "system" {
			continue
		}
		validCount++
	}

	if validCount == 0 {
		return fmt.Errorf("no exportable devices in inventory (only system objects found)")
	}

	return nil
}

// PrintSummary prints a summary of the export operation
func PrintSummary(result *LoadResult) {
	clog.Header("\n=== Export Summary ===")

	if len(result.LocationsCreated) > 0 {
		clog.Created("Created locations: %d", len(result.LocationsCreated))
	}
	if len(result.LocationsSkipped) > 0 {
		clog.Skipped("Skipped locations (already exist): %d", len(result.LocationsSkipped))
	}

	if len(result.RacksCreated) > 0 {
		clog.Created("Created racks: %d", len(result.RacksCreated))
	}
	if result.RacksSkipped > 0 {
		clog.Skipped("Skipped racks (already exist): %d", result.RacksSkipped)
	}

	if len(result.Created) > 0 {
		clog.Created("Created devices: %d", len(result.Created))
		for _, name := range result.Created {
			clog.SummaryCreated("%s", name)
		}
	}
	if len(result.Updated) > 0 {
		clog.Changed("Updated devices: %d", len(result.Updated))
		for _, name := range result.Updated {
			clog.SummaryChanged("%s", name)
		}
	}
	if len(result.Skipped) > 0 {
		clog.Skipped("Skipped devices (conflicts): %d", len(result.Skipped))
		for _, conflict := range result.Conflicts {
			clog.SummarySkipped("%s: %s", conflict.DeviceName, conflict.Reason)
		}
	}

	if result.IfacesCreated > 0 {
		clog.Created("Created interfaces: %d", result.IfacesCreated)
	}
	if result.IfacesSkipped > 0 {
		clog.Skipped("Skipped interfaces (already exist): %d", result.IfacesSkipped)
	}

	if result.ModulesCreated > 0 {
		clog.Created("Created modules: %d", result.ModulesCreated)
	}
	if result.ModulesSkipped > 0 {
		clog.Skipped("Skipped modules (already exist): %d", result.ModulesSkipped)
	}

	if result.FrusCreated > 0 {
		clog.Created("Created inventory items: %d", result.FrusCreated)
	}
	if result.FrusSkipped > 0 {
		clog.Skipped("Skipped inventory items (already exist): %d", result.FrusSkipped)
	}

	if result.CablesCreated > 0 {
		clog.Created("Created cables: %d", result.CablesCreated)
	}
	if result.CablesSkipped > 0 {
		clog.Skipped("Skipped cables (already exist): %d", result.CablesSkipped)
	}

	if result.VLANsCreated > 0 {
		clog.Created("Created VLANs: %d", result.VLANsCreated)
	}
	if result.VLANsSkipped > 0 {
		clog.Skipped("Skipped VLANs (already exist): %d", result.VLANsSkipped)
	}

	if result.PrefixesCreated > 0 {
		clog.Created("Created prefixes: %d", result.PrefixesCreated)
	}
	if result.PrefixesSkipped > 0 {
		clog.Skipped("Skipped prefixes (already exist): %d", result.PrefixesSkipped)
	}

	if result.IPAddressesCreated > 0 {
		clog.Created("Created IP addresses: %d", result.IPAddressesCreated)
	}
	if result.IPAddressesSkipped > 0 {
		clog.Skipped("Skipped IP addresses (already exist): %d", result.IPAddressesSkipped)
	}

	if len(result.Errors) > 0 {
		clog.Error("Errors: %d", len(result.Errors))
		for _, errMsg := range result.Errors {
			clog.SummaryError("%s", errMsg)
		}
	}
}
