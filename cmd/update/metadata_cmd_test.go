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
package update

import (
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
)

// TestUpdateDeviceRejectsMetadataAtParseTime verifies that update's
// --metadata fails before the command runs, with the same errors add gives,
// for both a malformed pair and a flag written without its value.
//
// Why it matters: add and update expose the same flag; until they shared the
// flag type, update re-parsed the pairs after loading the device and reported
// the mistake as an "invalid --set value", so the two verbs disagreed about
// the same typo and the device had already been resolved for nothing.
// Inputs: an inventory holding "cn-01"; "--metadata bogus", then
// "--metadata --status Active".
// Outputs: an error naming --metadata for each, zero Save calls.
// Data choice: the second argv is the one reported against add, replayed
// against update to prove the parser guard is not add-specific.
func TestUpdateDeviceRejectsMetadataAtParseTime(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"malformed pair", []string{"--metadata", "bogus"}, `invalid argument "bogus" for --metadata: expected key=value`},
		{"swallowed flag", []string{"--metadata", "--status", "Active"}, "flag needs an argument: --metadata"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { checkMetadataRejected(t, tc.args, tc.want) })
	}
}

// checkMetadataRejected runs update device cn-01 with the given flags and
// checks the error and that nothing was saved.
func checkMetadataRejected(t *testing.T, flags []string, want string) {
	t.Helper()
	// Arrange.
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	cmdtest.AddDevice(inventory, rackID, "cn-01", 10)
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	// Act.
	err := harness.Execute(t, append([]string{"update", "device", "cn-01"}, flags...)...)

	// Assert.
	if err == nil {
		t.Fatal("expected a parse error, got nil")
	}
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want %q", err, want)
	}
	if harness.Store.Saves != 0 {
		t.Errorf("Saves = %d, want 0", harness.Store.Saves)
	}
}

// TestUpdateDeviceAppliesMetadataPairs verifies well-formed --metadata pairs
// reach the device's provider metadata through the shared flag type.
//
// Why it matters: switching the flag type must not change what providers
// receive; with no provider registered the pairs are parsed but nothing is
// written, so this proves the parse-and-save path and leaves the dispatch to
// the cmdtest capability tests.
// Inputs: "update device cn-01 --metadata owner=ops --metadata note=a=b".
// Outputs: a nil error, one Save, and the two pairs readable from the flag.
// Data choice: "note=a=b" guards the first-"=" split end to end.
func TestUpdateDeviceAppliesMetadataPairs(t *testing.T) {
	// Arrange.
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	cmdtest.AddDevice(inventory, rackID, "cn-01", 10)
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)

	// Act.
	err := harness.Execute(t, "update", "device", "cn-01", "--metadata", "owner=ops", "--metadata", "note=a=b")

	// Assert.
	if err != nil {
		t.Fatalf("update device: unexpected error: %v", err)
	}
	if harness.Store.Saves != 1 {
		t.Errorf("Saves = %d, want 1", harness.Store.Saves)
	}
	got, _ := harness.Target.Flags().GetStringToString("metadata")
	if got["owner"] != "ops" || got["note"] != "a=b" {
		t.Errorf("metadata = %v, want owner=ops note=a=b", got)
	}
}
