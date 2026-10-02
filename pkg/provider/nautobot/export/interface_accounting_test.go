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

import "testing"

// TestLossyExportCountSumsNonErrorLosses verifies lossyExportCount adds the
// counters that describe objects not exported as authored and ignores
// already-exists skips.
//
// Why it matters: the closing summary warning fires on this total; counting
// idempotent re-run skips would make every second export warn.
// Inputs: a LoadResult with one of each lossy counter and ten already-exists
// skips. Outputs: 3.
// Data choice: distinct non-zero values make a missed or doubled term visible.
func TestLossyExportCountSumsNonErrorLosses(t *testing.T) {
	result := &LoadResult{IfacesUnsupported: 1, IfacesUnresolvedRefs: 1,
		CablesConflicted: 1, IfacesSkipped: 10, CablesSkipped: 10}
	if got := lossyExportCount(result); got != 3 {
		t.Fatalf("lossyExportCount = %d, want 3", got)
	}
	if got := lossyExportCount(&LoadResult{IfacesSkipped: 3}); got != 0 {
		t.Fatalf("already-exists skips counted as lossy: %d", got)
	}
}
