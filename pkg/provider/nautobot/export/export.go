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
		if device.Type == "system" || device.HardwareType == "system" {
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

	if len(result.Created) > 0 {
		clog.Created("Created: %d devices", len(result.Created))
		for _, name := range result.Created {
			clog.SummaryCreated("%s", name)
		}
	}

	if len(result.Updated) > 0 {
		clog.Changed("Updated: %d devices", len(result.Updated))
		for _, name := range result.Updated {
			clog.SummaryChanged("%s", name)
		}
	}

	if len(result.Skipped) > 0 {
		clog.Skipped("Skipped (conflicts): %d devices", len(result.Skipped))
		for _, conflict := range result.Conflicts {
			clog.SummarySkipped("%s: %s", conflict.DeviceName, conflict.Reason)
		}
	}

	if len(result.Errors) > 0 {
		clog.Error("Errors: %d", len(result.Errors))
		for _, errMsg := range result.Errors {
			clog.SummaryError("%s", errMsg)
		}
	}

	total := len(result.Created) + len(result.Updated) + len(result.Skipped) + len(result.Errors)
	clog.Info("\nTotal processed: %d devices", total)
}
