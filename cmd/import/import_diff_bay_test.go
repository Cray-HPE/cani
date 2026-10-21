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
package imprt

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// TestComputeDeviceDiffsBay verifies a bay change is reported and an incoming
// device without a bay reports nothing.
//
// Why it matters: the import preview must show when a re-import will move a
// device between chassis bays. Inputs: an existing device in "Bay 1" versus
// incoming "Bay 2" and an incoming device with no bay. Outputs: one "bay"
// diff, then none.
func TestComputeDeviceDiffsBay(t *testing.T) {
	existing := &devicetypes.CaniDeviceType{ParentBay: "Bay 1"}
	diffs := computeDeviceDiffs(existing, &devicetypes.CaniDeviceType{ParentBay: "Bay 2"})
	if len(diffs) != 1 || diffs[0].field != "bay" || diffs[0].newVal != "Bay 2" {
		t.Errorf("diffs = %+v, want one bay diff to Bay 2", diffs)
	}
	if diffs := computeDeviceDiffs(existing, &devicetypes.CaniDeviceType{}); len(diffs) != 0 {
		t.Errorf("diffs = %+v, want none for an incoming device without a bay", diffs)
	}
}
