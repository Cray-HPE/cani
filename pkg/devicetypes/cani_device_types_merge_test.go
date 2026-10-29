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
package devicetypes

import (
	"slices"
	"testing"
)

// TestMergePropertiesCarriesDeclaredModuleBays verifies the bays an incoming
// record declares replace the existing ones and report a change, while an
// incoming record that declares none leaves the existing bays untouched.
//
// Why it matters: a device re-imported from Nautobot carries the bays Nautobot
// declares; without this the existing record keeps the library's bay names and
// the merge clears every module bay Nautobot spells differently.
// Inputs: a receiver declaring PCIe1, an incoming record declaring "Slot B",
// then an incoming record declaring nothing. Outputs: true and ["Slot B"], then
// false and ["Slot B"].
// Data choice: "Slot B" is a spelling no library template uses, so the result
// can only come from the incoming record.
func TestMergePropertiesCarriesDeclaredModuleBays(t *testing.T) {
	d := &CaniDeviceType{Name: "srv", ModuleBays: []ModuleBaySpec{{Name: "PCIe1"}}}
	declared := []ModuleBaySpec{{Name: "Slot B", Position: "2"}}

	if !d.MergeProperties(&CaniDeviceType{Name: "srv", ModuleBays: declared}) || !slices.Equal(d.ModuleBays, declared) {
		t.Fatalf("bays after merge = %+v, want %+v and a reported change", d.ModuleBays, declared)
	}
	if d.MergeProperties(&CaniDeviceType{Name: "srv"}) || !slices.Equal(d.ModuleBays, declared) {
		t.Fatalf("bays after empty merge = %+v, want %+v and no change", d.ModuleBays, declared)
	}
}
