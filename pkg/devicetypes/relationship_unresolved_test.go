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
	"bytes"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// captureLog sends the standard logger to a buffer for the rest of the test.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var out bytes.Buffer
	writer, flags := log.Writer(), log.Flags()
	log.SetOutput(&out)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(writer)
		log.SetFlags(flags)
	})
	return &out
}

// setDebug sets the package Debug switch for the rest of the test.
func setDebug(t *testing.T, on bool) {
	t.Helper()
	previous := Debug
	Debug = on
	t.Cleanup(func() { Debug = previous })
}

// TestVerifyRelationshipsLogsConflictsWithoutDebug verifies the relationship
// pass every change runs logs a duplicate prefix without --debug.
//
// Why it matters: the duplicate is a conflict the operator must resolve, and
// it was logged only with --debug, so in practice nobody saw it.
// Inputs: two Global 10.0.0.0/24 prefixes, with Debug off.
// Outputs: one log line naming the duplicate.
// Data choice: a duplicate prefix is the conflict legacy inventories hold most.
func TestVerifyRelationshipsLogsConflictsWithoutDebug(t *testing.T) {
	// Arrange.
	setDebug(t, false)
	out := captureLog(t)
	inv := NewInventory()
	for range 2 {
		id := uuid.New()
		inv.Prefixes[id] = &CaniPrefix{ID: id, Prefix: "10.0.0.0/24"}
	}

	// Act.
	inv.VerifyParentChildRelationships()

	// Assert.
	want := `Warning: prefix 10.0.0.0/24 exists twice in namespace "Global"`
	if got := strings.Count(out.String(), want); got != 1 {
		t.Errorf("duplicate logged %d times, want once; log:\n%s", got, out)
	}
}

// TestLogUnresolvedCapsUnlessDebug verifies the conflict list is sorted and
// capped without --debug, with a count of the rest, and complete with it.
//
// Why it matters: a legacy inventory can hold hundreds of duplicates, and
// every change would otherwise print all of them.
// Inputs: maxUnresolvedLogged+2 conflicts in reverse order, with Debug off
// and on.
// Outputs: off, the lowest maxUnresolvedLogged in order and a count of 2;
// on, every conflict and no count.
// Data choice: reverse input order shows the cap keeps the first conflicts
// in sorted order, not in input order.
func TestLogUnresolvedCapsUnlessDebug(t *testing.T) {
	total := maxUnresolvedLogged + 2
	result := &RelationshipResult{}
	for i := total - 1; i >= 0; i-- {
		result.Unresolved = append(result.Unresolved, fmt.Sprintf("conflict %02d", i))
	}
	last := fmt.Sprintf("conflict %02d", total-1)
	cases := []struct {
		name string
		want unresolvedLog
	}{
		{name: "debug off", want: unresolvedLog{debug: false, shown: maxUnresolvedLogged, last: false, count: true}},
		{name: "debug on", want: unresolvedLog{debug: true, shown: total, last: true, count: false}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			setDebug(t, tc.want.debug)
			out := captureLog(t)

			// Act.
			result.LogUnresolved()

			// Assert.
			assertUnresolvedLog(t, out.String(), last, tc.want)
		})
	}
}

// unresolvedLog is what one LogUnresolved run should have printed.
type unresolvedLog struct {
	debug       bool
	shown       int  // conflicts listed
	last, count bool // highest conflict listed; "more conflicts" line printed
}

func assertUnresolvedLog(t *testing.T, logged, last string, want unresolvedLog) {
	t.Helper()
	if got := strings.Count(logged, "Warning: conflict "); got != want.shown {
		t.Errorf("logged %d conflicts, want %d", got, want.shown)
	}
	if !strings.HasPrefix(logged, "Warning: conflict 00\n") {
		t.Errorf("first line is not the lowest conflict; log:\n%s", logged)
	}
	if got := strings.Contains(logged, last); got != want.last {
		t.Errorf("%s logged = %v, want %v", last, got, want.last)
	}
	if got := strings.Contains(logged, "2 more conflicts to resolve; enable debug output"); got != want.count {
		t.Errorf("count line logged = %v, want %v; log:\n%s", got, want.count, logged)
	}
}
