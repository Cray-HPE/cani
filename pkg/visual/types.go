/*
 *
 *  MIT License
 *
 *  (C) Copyright 2024-2026 Hewlett Packard Enterprise Development LP
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
package visual

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorGreen  = "\033[32m"
	ColorGray   = "\033[90m"
	ColorYellow = "\033[33m"
	ColorCyan   = "\033[36m"
	ColorBold   = "\033[1m"
)

// RackSlot represents a single U position in a rack
type RackSlot struct {
	Device      *devicetypes.CaniDeviceType
	IsStart     bool // True if this is the starting U of a multi-U device
	IsContinued bool // True if this is a continuation of a multi-U device
}

// RackView represents a rack visualization
type RackView struct {
	Rack                *devicetypes.CaniDeviceType
	Height              int               // Total rack height in U
	Slots               map[int]*RackSlot // U position -> slot info (1-indexed)
	UnpositionedDevices []*devicetypes.CaniDeviceType
}

// RenderOptions controls visual output
type RenderOptions struct {
	NoColor    bool
	RackFilter string                 // Filter to specific rack name
	ShowCables bool                   // Show cable connections
	Inventory  *devicetypes.Inventory // Full inventory for cable lookups
}

// NewRackView creates a new RackView for the given rack
func NewRackView(rack *devicetypes.CaniDeviceType, height int) *RackView {
	return &RackView{
		Rack:                rack,
		Height:              height,
		Slots:               make(map[int]*RackSlot),
		UnpositionedDevices: []*devicetypes.CaniDeviceType{},
	}
}

// AddDevice adds a device to the rack view at the specified position
func (rv *RackView) AddDevice(device *devicetypes.CaniDeviceType, position, uHeight int) {
	for u := position; u < position+uHeight && u <= rv.Height; u++ {
		rv.Slots[u] = &RackSlot{
			Device:      device,
			IsStart:     u == position,
			IsContinued: u != position,
		}
	}
}

// AddUnpositionedDevice adds a device that lacks rack position info
func (rv *RackView) AddUnpositionedDevice(device *devicetypes.CaniDeviceType) {
	rv.UnpositionedDevices = append(rv.UnpositionedDevices, device)
}

// GetSlot returns the slot at the given U position
func (rv *RackView) GetSlot(u int) *RackSlot {
	return rv.Slots[u]
}

// IsOccupied returns true if the U position is occupied
func (rv *RackView) IsOccupied(u int) bool {
	_, ok := rv.Slots[u]
	return ok
}

// OccupiedCount returns the number of occupied U positions
func (rv *RackView) OccupiedCount() int {
	return len(rv.Slots)
}

// EmptyCount returns the number of empty U positions
func (rv *RackView) EmptyCount() int {
	return rv.Height - len(rv.Slots)
}

// DeviceCount returns the number of unique devices in the rack
func (rv *RackView) DeviceCount() int {
	seen := make(map[uuid.UUID]bool)
	for _, slot := range rv.Slots {
		if slot != nil && slot.Device != nil {
			seen[slot.Device.ID] = true
		}
	}
	return len(seen)
}

// FieldMapping represents a single CSV field to target field mapping.
// Used for step-through display to show how raw CSV data maps to inventory objects.
type FieldMapping struct {
	SourceField string // CSV column name (e.g., "PartNumber", "Description")
	SourceValue string // Raw value from CSV
	TargetType  string // Target type (e.g., "CaniDeviceType", "CaniRackType")
	TargetField string // Target field name (e.g., "Name", "DeviceTypeSlug")
	TargetValue string // Transformed value (may differ from source)
	IsDerived   bool   // True if value is computed/derived (not a direct copy)
}

// CreatedItemInfo holds info about items created from a single CSV record.
type CreatedItemInfo struct {
	ID   string // Short UUID (first 8 chars)
	Name string // Generated name
}

// TransformStepInfo holds all info needed for a transform step display.
type TransformStepInfo struct {
	Quantity     int               // CSV Quantity value
	HwType       string            // Hardware type classification
	Mappings     []FieldMapping    // Field mappings (from first item as template)
	CreatedItems []CreatedItemInfo // All items created from this record
}
