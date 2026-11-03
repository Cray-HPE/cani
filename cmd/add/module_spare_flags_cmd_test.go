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
package add

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
)

// runCapturingStdout runs the harness target with the flags and returns what
// it printed to os.Stdout, where the add module plan printers write.
func runCapturingStdout(t *testing.T, harness *cmdtest.Harness, flags map[string][]string, args ...string) (string, error) {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	previous := os.Stdout
	os.Stdout = writer
	runErr := harness.Run(t, flags, args...)
	os.Stdout = previous
	_ = writer.Close()
	var output bytes.Buffer
	_, _ = io.Copy(&output, reader)
	return output.String(), runErr
}

// spareHarness returns a harness over an inventory with one location and the
// location's name.
func spareHarness(t *testing.T) (*cmdtest.Harness, string) {
	t.Helper()
	inventory, _ := cmdtest.InventoryWithRack("rack-01")
	var siteName string
	for _, loc := range inventory.Locations {
		siteName = loc.Name
	}
	return cmdtest.New(t, NewCommand(), newModuleCommand(), inventory), siteName
}

// TestAddModuleSpareDryRunSavesNothing verifies add module --spare --dry-run
// prints where the spare would be stored and neither adds nor saves it.
//
// Why it matters: the single-device add path ignored --dry-run and saved,
// so an operator previewing a spare wrote it; the flag is documented as a
// preview on the same command and honored on every other placement path.
// Inputs: --spare --location <site> --serial SN-1 --dry-run for a library
// NIC. Outputs: no error, one "Would store module" line naming the type,
// serial and site on stdout, zero modules, zero saves.
// Data choice: the serial proves the preview describes the unnamed module
// as its label would read once added.
func TestAddModuleSpareDryRunSavesNothing(t *testing.T) {
	harness, siteName := spareHarness(t)

	out, err := runCapturingStdout(t, harness, map[string][]string{"spare": {"true"}, "location": {siteName}, "serial": {"SN-1"}, "dry-run": {"true"}}, "hpe-2p-25gbe-pcie-nic")

	if err != nil {
		t.Fatalf("dry run: unexpected error: %v", err)
	}
	if !strings.Contains(out, "Would store module") || !strings.Contains(out, "(Serial: SN-1)") || !strings.Contains(out, "at location "+siteName) {
		t.Errorf("stdout = %q, want the planned spare with its serial and location", out)
	}
	if n := len(harness.Inventory().Modules); n != 0 || harness.Store.Saves != 0 {
		t.Errorf("modules = %d, saves = %d; want nothing added or saved", n, harness.Store.Saves)
	}
}

// TestAddModuleSpareRejectsDeviceAndBay verifies --spare is refused together
// with --device, including a placement strategy, and with --bay, before
// anything is placed or saved.
//
// Why it matters: a strategy in --device bypassed the spare path and
// installed the modules the operator asked to store, a literal --device was
// silently dropped, and --bay was persisted on a module that has no bay.
// Inputs: --spare --location <site> with --device %{FILL}, then with
// --device rack-01-node, then --spare with --bay PCIe1 (the device flag
// cleared). Outputs: an error naming both flags each time, zero modules,
// zero saves.
// Data choice: the strategy form is the one that reached a path never
// reading --spare.
func TestAddModuleSpareRejectsDeviceAndBay(t *testing.T) {
	harness, siteName := spareHarness(t)
	cases := []struct {
		name  string
		flags map[string][]string
	}{
		{"strategy device", map[string][]string{"spare": {"true"}, "location": {siteName}, "device": {"%{FILL}"}}},
		{"literal device", map[string][]string{"device": {"rack-01-node"}}},
		{"bay", map[string][]string{"device": {""}, "bay": {"PCIe1"}}},
	}
	for _, tc := range cases {
		err := harness.Run(t, tc.flags, "hpe-2p-25gbe-pcie-nic")
		if err == nil || !strings.Contains(err.Error(), "neither --device nor --bay") {
			t.Errorf("%s: error = %v, want the --spare conflict", tc.name, err)
		}
	}
	if n := len(harness.Inventory().Modules); n != 0 || harness.Store.Saves != 0 {
		t.Errorf("modules = %d, saves = %d; want nothing added or saved", n, harness.Store.Saves)
	}
}
