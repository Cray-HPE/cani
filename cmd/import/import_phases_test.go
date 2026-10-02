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
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

// captureLog redirects the standard logger into a buffer for one test.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	return &buf
}

// TestPrintTransformWarningsListsEachRecord verifies the transform warnings
// a provider returns are printed with their count, and that nothing is
// printed when there are none.
//
// Why it matters: a provider that leaves source records out of its result
// has no other channel to the operator; without this block the import ends
// with a plain success message.
// Inputs: two warning lines, then an empty list. Outputs: a header naming
// two records followed by both lines; no output for the empty list.
// Data choice: two distinct lines show order and completeness, the empty
// case guards against a header with nothing under it.
func TestPrintTransformWarningsListsEachRecord(t *testing.T) {
	buf := captureLog(t)
	warnings := []string{"prefix 10.0.0.0/24: namespace 1 not found", "VRF blue: namespace 1 not found"}

	printTransformWarnings(warnings)

	out := buf.String()
	for _, want := range append([]string{"=== 2 source record(s) not imported ==="}, warnings...) {
		if !strings.Contains(out, want) {
			t.Errorf("output %q missing %q", out, want)
		}
	}

	buf.Reset()
	printTransformWarnings(nil)
	if buf.Len() != 0 {
		t.Errorf("output for no warnings = %q, want none", buf.String())
	}
}

// TestNotImportedSuffixQualifiesOnlyPartialImports verifies the completion
// message gains a count only when records were left out.
//
// Why it matters: the shellspec suites match the plain success line, and a
// partial import must not read as a clean one.
// Inputs: counts 0 and 3. Outputs: an empty suffix, then one naming three
// records and pointing at the warnings.
// Data choice: zero is the common path; three shows the count is substituted.
func TestNotImportedSuffixQualifiesOnlyPartialImports(t *testing.T) {
	if got := notImportedSuffix(0); got != "" {
		t.Errorf("notImportedSuffix(0) = %q, want empty", got)
	}
	want := " (3 source record(s) not imported, see warnings above)"
	if got := notImportedSuffix(3); got != want {
		t.Errorf("notImportedSuffix(3) = %q, want %q", got, want)
	}
}
