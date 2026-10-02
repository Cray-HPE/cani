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
package export

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// TestNoteExistingModuleInterfaceSeparatesDuplicateFromRerun verifies that an
// existing module port is counted as a duplicate-name skip when the parent
// device spec declares the same name, and as an already-existing interface
// otherwise.
//
// Why it matters: the DL380 Gen11 + CX7 "HSN 0" collision silently dropped
// the module copy that carried the MAC; the summary must distinguish that
// from the idempotent re-run case.
// Inputs: a parent declaring "HSN 0" and a module reporting "HSN 0" and then
// "Port 1" as existing. Outputs: IfacesSkippedDuplicate == 1, IfacesSkipped == 1.
// Data choice: one name per branch of the helper.
func TestNoteExistingModuleInterfaceSeparatesDuplicateFromRerun(t *testing.T) {
	parent := &devicetypes.CaniDeviceType{Name: "repro-dl380",
		Interfaces: []devicetypes.InterfaceSpec{{Name: "HSN 0", Type: devicetypes.InterfacesElemTypeA400GbaseXQsfpdd}}}
	module := &devicetypes.CaniModuleType{Name: "cx7-a"}
	result := &LoadResult{}

	noteExistingModuleInterface(result, module, parent, "HSN 0")
	noteExistingModuleInterface(result, module, parent, "Port 1")

	if result.IfacesSkippedDuplicate != 1 || result.IfacesSkipped != 1 {
		t.Fatalf("duplicate=%d skipped=%d, want 1 and 1", result.IfacesSkippedDuplicate, result.IfacesSkipped)
	}
	if deviceDeclaresInterface(nil, "HSN 0") {
		t.Fatal("nil device must not declare any interface")
	}
}

// TestLossyExportCountSumsNonErrorLosses verifies lossyExportCount adds the
// counters that describe objects not exported as authored and ignores
// already-exists skips.
//
// Why it matters: the closing summary warning fires on this total; counting
// idempotent re-run skips would make every second export warn.
// Inputs: a LoadResult with one of each lossy counter and ten already-exists
// skips. Outputs: 4.
// Data choice: distinct non-zero values make a missed or doubled term visible.
func TestLossyExportCountSumsNonErrorLosses(t *testing.T) {
	result := &LoadResult{IfacesSkippedDuplicate: 1, IfacesUnsupported: 1,
		IfacesUnresolvedRefs: 1, CablesConflicted: 1, IfacesSkipped: 10, CablesSkipped: 10}
	if got := lossyExportCount(result); got != 4 {
		t.Fatalf("lossyExportCount = %d, want 4", got)
	}
	if got := lossyExportCount(&LoadResult{IfacesSkipped: 3}); got != 0 {
		t.Fatalf("already-exists skips counted as lossy: %d", got)
	}
}
