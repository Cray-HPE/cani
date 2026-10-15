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

	"github.com/google/uuid"
)

// ---------- LoadResult / ConflictInfo structs ----------

// TestLoadResultZeroValue verifies a freshly constructed LoadResult has zero
// counters (CablesCreated, IfacesCreated).
//
// Why it matters: summary tallies start empty and are incremented during load;
// a non-zero zero-value would corrupt every export summary.
// Inputs: an empty &LoadResult{}. Outputs: zero-valued counter fields.
// Data choice: checking two representative counters confirms the struct's zero
// value without enumerating every field.
func TestLoadResultZeroValue(t *testing.T) {
	result := &LoadResult{}
	if result.CablesCreated != 0 {
		t.Error("expected zero value for CablesCreated")
	}
	if result.IfacesCreated != 0 {
		t.Error("expected zero value for IfacesCreated")
	}
}

// TestConflictInfoDiffs verifies a ConflictInfo carries its FieldDiff slice
// intact (length and Field value).
//
// Why it matters: ConflictInfo is what the exporter reports for devices that
// differ from Nautobot; its diffs must round-trip so operators see what would
// change under --merge.
// Inputs: a ConflictInfo literal with one "status" FieldDiff. Outputs: a slice
// of length 1 whose Field is "status".
// Data choice: a single status diff is the minimal case proving the slice field
// is populated and readable.
func TestConflictInfoDiffs(t *testing.T) {
	ci := ConflictInfo{
		DeviceName: "test-device",
		ExistingID: uuid.New(),
		LocalID:    uuid.New(),
		Reason:     "already exists",
		Diffs: []FieldDiff{
			{Field: "status", LocalVal: "Active", RemoteVal: "Planned"},
		},
	}
	if len(ci.Diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(ci.Diffs))
	}
	if ci.Diffs[0].Field != "status" {
		t.Errorf("field = %q, want 'status'", ci.Diffs[0].Field)
	}
}

// ---------- printLoadSummary ----------

// TestPrintLoadSummaryAllBranches verifies printLoadSummary runs without
// panicking when every result category is populated.
//
// Why it matters: the end-of-export summary must render every created/skipped
// tally (locations, racks, devices, interfaces, IPAM, ...) so operators get a
// complete report.
// Inputs: a LoadResult with all slices/counters non-empty. Outputs: console
// output only (no return value); the test asserts no panic.
// Data choice: populating every field forces each conditional print branch to
// execute at least once.
func TestPrintLoadSummaryAllBranches(t *testing.T) {
	e := &Exporter{Options: &ExporterOpts{}}

	// Test with all fields populated — exercises every branch
	result := &LoadResult{
		Created:            []string{"dev-1", "dev-2"},
		Updated:            []string{"dev-3"},
		Skipped:            []string{"dev-4"},
		Errors:             []string{"something failed"},
		Conflicts:          []ConflictInfo{{DeviceName: "dev-4", Diffs: []FieldDiff{{Field: "status", LocalVal: "Active", RemoteVal: "Planned"}}}},
		LocationsCreated:   []string{"DC1"},
		LocationsSkipped:   []string{"DC2"},
		RacksCreated:       []string{"Rack-A"},
		RacksSkipped:       1,
		IfacesCreated:      5,
		IfacesSkipped:      2,
		IfacesUnsupported:  1,
		ModulesCreated:     3,
		ModulesSkipped:     1,
		FrusCreated:        4,
		FrusSkipped:        2,
		CablesCreated:      10,
		CablesSkipped:      3,
		CablesConflicted:   2,
		VLANsCreated:       2,
		VLANsSkipped:       1,
		PrefixesCreated:    6,
		PrefixesSkipped:    2,
		IPAddressesCreated: 8,
		IPAddressesSkipped: 4,
	}

	// Should not panic — just prints to clog
	e.printLoadSummary(result)
}

// TestPrintLoadSummaryEmpty verifies printLoadSummary runs without panicking
// when the result is empty (most branches skipped).
//
// Why it matters: an export that changed nothing must still print a clean
// summary rather than crash on empty slices/zero counters.
// Inputs: an empty &LoadResult{}. Outputs: console output only; the test
// asserts no panic.
// Data choice: the zero-value result exercises the all-branches-skipped path,
// complementing the all-populated test.
func TestPrintLoadSummaryEmpty(t *testing.T) {
	e := &Exporter{Options: &ExporterOpts{}}
	result := &LoadResult{}

	// Empty result — most branches are skipped
	e.printLoadSummary(result)
}

// ---------- printConflictDiffs ----------

// TestPrintConflictDiffsNoDiffs verifies printConflictDiffs returns early
// (prints nothing) when no conflict carries any FieldDiff.
//
// Why it matters: the "pending changes" banner should only appear when there is
// actually something to merge, keeping export output uncluttered.
// Inputs: a LoadResult with one conflict whose Diffs is nil. Outputs: console
// output only; the test asserts no panic on the early-return path.
// Data choice: a conflict with nil Diffs is the exact condition that triggers
// the early return.
func TestPrintConflictDiffsNoDiffs(t *testing.T) {
	e := &Exporter{Options: &ExporterOpts{}}
	result := &LoadResult{
		Conflicts: []ConflictInfo{
			{DeviceName: "dev-1", Diffs: nil},
		},
	}
	// Should return early since no conflict has diffs
	e.printConflictDiffs(result)
}

// TestPrintConflictDiffsWithDiffs verifies printConflictDiffs prints the diff
// banner and per-device diffs when a conflict has FieldDiffs.
//
// Why it matters: operators rely on this output to see exactly which fields
// would change under --merge before mutating Nautobot.
// Inputs: a LoadResult with one conflict carrying two FieldDiffs. Outputs:
// console output only; the test asserts no panic.
// Data choice: two diffs (status, role) exercise the per-diff loop more than
// once.
func TestPrintConflictDiffsWithDiffs(t *testing.T) {
	e := &Exporter{Options: &ExporterOpts{}}
	result := &LoadResult{
		Conflicts: []ConflictInfo{
			{
				DeviceName: "dev-conflict",
				Diffs: []FieldDiff{
					{Field: "status", LocalVal: "Active", RemoteVal: "Planned"},
					{Field: "role", LocalVal: "Server", RemoteVal: "Switch"},
				},
			},
		},
	}
	// Should print the diffs without panicking
	e.printConflictDiffs(result)
}

// TestPrintConflictDiffsEmptyConflicts verifies printConflictDiffs handles an
// empty Conflicts slice without panicking (nothing to print).
//
// Why it matters: a clean export with zero conflicts must not crash the summary
// stage on an empty slice.
// Inputs: a LoadResult with Conflicts set to an empty slice. Outputs: console
// output only; the test asserts no panic.
// Data choice: an explicitly empty (non-nil) slice exercises the no-iteration
// path of the conflict loop.
func TestPrintConflictDiffsEmptyConflicts(t *testing.T) {
	e := &Exporter{Options: &ExporterOpts{}}
	result := &LoadResult{
		Conflicts: []ConflictInfo{},
	}
	e.printConflictDiffs(result)
}

// ---------- printDeviceDiffs ----------

// TestPrintDeviceDiffsNoDiffs verifies printDeviceDiffs returns early without
// printing when the diff slice is nil or empty.
//
// Why it matters: when a local device already matches its Nautobot counterpart
// the export must stay silent, so the sync summary never lists phantom changes.
// Inputs: a device name plus nil and an empty []FieldDiff. Outputs: none; the
// function returns before touching the colored logger.
// Data choice: both nil and an empty slice are exercised to cover the two
// "no diffs" shapes callers produce. NOTE: a smoke test — it only confirms no
// panic and asserts nothing.
func TestPrintDeviceDiffsNoDiffs(t *testing.T) {
	// Should not panic or output when diffs is empty
	printDeviceDiffs("test-device", nil)
	printDeviceDiffs("test-device", []FieldDiff{})
}

// TestPrintDeviceDiffsWithDiffs verifies printDeviceDiffs runs without panicking
// when given a non-empty list of field diffs.
//
// Why it matters: diff output is how a dry-run export tells operators which
// fields would change on a Nautobot device before --merge is applied.
// Inputs: a device name and three FieldDiffs (status, role, location).
// Outputs: colored log lines only — there is no return value to assert.
// Data choice: three commonly drifted fields exercise the print loop body.
// NOTE: a smoke test — it asserts nothing beyond not panicking.
func TestPrintDeviceDiffsWithDiffs(t *testing.T) {
	diffs := []FieldDiff{
		{Field: "status", LocalVal: "Active", RemoteVal: "Planned"},
		{Field: "role", LocalVal: "Server", RemoteVal: "Switch"},
		{Field: "location", LocalVal: "DC1", RemoteVal: "DC2"},
	}
	// Should not panic — just prints to clog
	printDeviceDiffs("diff-device", diffs)
}

// ---------- printLoadSummary conflict without diffs ----------

// TestPrintLoadSummaryConflictNoDiffs verifies printLoadSummary handles a skipped
// device whose conflict carries no field diffs (the "up to date" branch).
//
// Why it matters: the end-of-sync summary must distinguish a conflict with real
// pending changes from one that is already current.
// Inputs: a LoadResult with Skipped=["dev-5"] and a ConflictInfo with nil Diffs.
// Outputs: colored summary log lines only.
// Data choice: a conflict with nil Diffs targets the empty-diffs "up to date"
// path. NOTE: a smoke test — it asserts nothing beyond not panicking.
func TestPrintLoadSummaryConflictNoDiffs(t *testing.T) {
	e := &Exporter{}
	result := &LoadResult{
		Skipped: []string{"dev-5"},
		Conflicts: []ConflictInfo{
			{DeviceName: "dev-5", Diffs: nil},
		},
	}
	e.printLoadSummary(result)
}
