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
package visual

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// detailModelWidth matches the model truncation used by the tree view.
const detailModelWidth = 30

// detailStartLabel renders "name Nu | model" for the first U of a device.
func detailStartLabel(dev *devicetypes.CaniDeviceType, uHeight int, cf colorFuncs) string {
	label := cf.white(fmt.Sprintf("%s %du", dev.Name, uHeight))
	if dev.Model == "" {
		return label
	}
	return label + cf.gray(" | "+Trunc(dev.Model, detailModelWidth))
}

// detailChildrenLabel summarises child devices (e.g. chassis nodes) as
// "child1, child2 | model", one group per model, sorted by name.
func detailChildrenLabel(dev *devicetypes.CaniDeviceType, inv *devicetypes.Inventory) string {
	children := ResolveDeviceChildren(dev.Children, inv)
	if len(children) == 0 {
		return ""
	}
	byModel := make(map[string][]string)
	for _, child := range children {
		byModel[child.Model] = append(byModel[child.Model], child.Name)
	}
	models := make([]string, 0, len(byModel))
	for model := range byModel {
		models = append(models, model)
	}
	sort.Strings(models)

	groups := make([]string, 0, len(models))
	for _, model := range models {
		names := byModel[model]
		sort.Strings(names)
		groups = append(groups, PipeSep(strings.Join(names, ", "), Trunc(model, detailModelWidth)))
	}
	return strings.Join(groups, "; ")
}
