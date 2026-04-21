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
	"fmt"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// RackFormat identifies a rack-specific output format.
type RackFormat string

const (
	RackFormatClassic RackFormat = "classic"
	RackFormatMinimap RackFormat = "minimap"
	RackFormatDetail  RackFormat = "detail"
	RackFormatRouting RackFormat = "routing"
)

// ValidRackFormats returns the rack-specific format names.
func ValidRackFormats() []string {
	return []string{
		string(RackFormatClassic),
		string(RackFormatMinimap),
		string(RackFormatDetail),
		string(RackFormatRouting),
	}
}

// RenderRack dispatches to the appropriate rack renderer for the given format.
func RenderRack(inv *devicetypes.Inventory, format RackFormat, opts CompactRenderOptions) error {
	switch format {
	case RackFormatClassic:
		return RenderAllRacks(inv, RenderOptions{
			NoColor:    opts.NoColor,
			RackFilter: opts.RackFilter,
			ShowCables: true,
			Inventory:  inv,
		})
	case RackFormatMinimap:
		return RenderMinimapRacks(inv, opts)
	case RackFormatDetail:
		opts.Detail = true
		return RenderMinimapDetailAll(inv, opts)
	case RackFormatRouting:

		if opts.Interactive {
			return RunInteractiveRouting(inv, opts)
		}
		return RenderRoutingView(inv, opts)
	default:
		return fmt.Errorf("invalid rack format %q; valid options: %s",
			format, strings.Join(ValidRackFormats(), ", "))
	}
}
