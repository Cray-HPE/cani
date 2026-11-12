/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024, 2026 Hewlett Packard Enterprise Development LP
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
package imprt

import (
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// parentNamer resolves a parent ID to its name. An incoming device's parent
// carries the source system's ID, which differs from the local record's ID
// until the merge remaps it, so parents are compared by name.
type parentNamer struct {
	inv    *devicetypes.Inventory
	result *devicetypes.TransformResult
}

// name returns the name of the device or rack with the given ID, looking in
// the incoming result first, or "" when neither holds it.
func (p parentNamer) name(id uuid.UUID) string {
	if p.result != nil {
		if d := p.result.Devices[id]; d != nil {
			return d.Name
		}
		if r := p.result.Racks[id]; r != nil {
			return r.Name
		}
	}
	if p.inv == nil {
		return ""
	}
	if d := p.inv.Devices[id]; d != nil {
		return d.Name
	}
	if r := p.inv.Racks[id]; r != nil {
		return r.Name
	}
	return ""
}

// appendParentDiff records a parent change only when both parents resolve
// to names and the names differ; an unresolved parent is not reported.
func appendParentDiff(diffs []fieldDiff, names parentNamer, oldID, newID uuid.UUID) []fieldDiff {
	if newID == uuid.Nil || newID == oldID {
		return diffs
	}
	oldName, newName := names.name(oldID), names.name(newID)
	if oldName == "" || newName == "" || oldName == newName {
		return diffs
	}
	return append(diffs, fieldDiff{"parent", oldName, newName})
}

// appendCustomFieldDiffs records each incoming custom field whose value
// differs from the existing one, in key order.
func appendCustomFieldDiffs(diffs []fieldDiff, oldCF, newCF map[string]any) []fieldDiff {
	for _, k := range slices.Sorted(maps.Keys(newCF)) {
		oldVal, ok := oldCF[k]
		if ok && reflect.DeepEqual(oldVal, newCF[k]) {
			continue
		}
		diffs = append(diffs, fieldDiff{"customFields." + k, fmtCF(oldVal, ok), fmt.Sprint(newCF[k])})
	}
	return diffs
}

func fmtCF(v any, ok bool) string {
	if !ok {
		return "<unset>"
	}
	return fmt.Sprint(v)
}
