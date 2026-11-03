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

// PrintModuleTable renders modules as a fixed-width table. NAME is the
// optional label an operator gave the module and is blank for a module
// identified only by its serial or asset tag, which the SERIAL column and
// the model carry; DEVICE and BAY show where an installed module sits.
func PrintModuleTable(modules []*devicetypes.CaniModuleType, inv *devicetypes.Inventory) {
	widths := []int{ColName, ColType, ColModel, ColSerial, ColStatus, ColDevice, ColBay}
	fmt.Println(moduleRow([]string{"NAME", "TYPE", "MODEL", "SERIAL", "STATUS", "DEVICE", "BAY"}, widths))
	rules := make([]string, len(widths))
	for i, w := range widths {
		rules[i] = strings.Repeat("-", w)
	}
	fmt.Println(moduleRow(rules, widths))
	for _, m := range modules {
		fmt.Println(moduleRow([]string{
			m.Name, string(m.Type), m.Model, m.Serial, m.Status, ResolveDeviceName(m.ParentDevice, inv), m.ModuleBayName,
		}, widths))
	}
	fmt.Printf("\nTotal: %d module(s)\n", len(modules))
}

// moduleRow joins the cells at their column widths.
func moduleRow(cells []string, widths []int) string {
	parts := make([]string, len(cells))
	for i, cell := range cells {
		parts[i] = Col(cell, widths[i])
	}
	return strings.Join(parts, "  ")
}
