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

// This is an external test package so it can import the cmd packages and
// register providers without contaminating the empty-registry assertions those
// packages make about themselves.
package cmdtest_test

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/cmd/add"
	"github.com/Cray-HPE/cani/cmd/remove"
	"github.com/Cray-HPE/cani/cmd/show"
	"github.com/Cray-HPE/cani/cmd/update"
	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
)

// rackSlug is a library rack used where a verb needs one.
const rackSlug = "hpe-42u-800mmx1200mm-g2-enterprise-shock-rack"

// subcommand finds a named subcommand on a parent built by a cmd package.
func subcommand(t *testing.T, parent *cli.Command, name string) *cli.Command {
	t.Helper()
	for _, candidate := range parent.Commands() {
		if candidate.Name() == name {
			return candidate
		}
	}
	t.Fatalf("subcommand %q not found on %q", name, parent.Name())
	return nil
}

// addDeviceFingerprint runs the same "add device" through a fresh harness and
// returns a UUID-free rendering of the resulting inventory.
func addDeviceFingerprint(t *testing.T) string {
	t.Helper()

	inventory, _ := cmdtest.InventoryWithRack("rack-01")
	parent := add.NewCommand()
	harness := cmdtest.New(t, parent, subcommand(t, parent, "device"), inventory)

	err := harness.Run(t, map[string][]string{
		"rack":     {"rack-01"},
		"position": {"10"},
		"face":     {"front"},
		"name":     {"cn-01"},
	}, cmdtest.DeviceSlug)
	if err != nil {
		t.Fatalf("add device: %v", err)
	}
	return cmdtest.Fingerprint(harness.Inventory())
}

// updateDeviceFingerprint runs the same "update device" and returns a UUID-free
// rendering of the resulting inventory.
func updateDeviceFingerprint(t *testing.T) string {
	t.Helper()

	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	cmdtest.AddDevice(inventory, rackID, "cn-01", 10)
	parent := update.NewCommand()
	harness := cmdtest.New(t, parent, subcommand(t, parent, "device"), inventory)

	if err := harness.Run(t, map[string][]string{"name": {"cn-99"}}, "cn-01"); err != nil {
		t.Fatalf("update device: %v", err)
	}
	return cmdtest.Fingerprint(harness.Inventory())
}

// removeDeviceFingerprint runs the same "remove device" and returns a UUID-free
// rendering of what survives.
func removeDeviceFingerprint(t *testing.T) string {
	t.Helper()

	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	cmdtest.AddDevice(inventory, rackID, "cn-01", 10)
	cmdtest.AddDevice(inventory, rackID, "cn-02", 12)
	parent := remove.NewCommand()
	harness := cmdtest.New(t, parent, subcommand(t, parent, "device"), inventory)

	if err := harness.Run(t, nil, "cn-01"); err != nil {
		t.Fatalf("remove device: %v", err)
	}
	return cmdtest.Fingerprint(harness.Inventory())
}

// showDeviceFingerprint runs the same "show device" and returns what it
// rendered.
//
// Show does not mutate, so fingerprinting the inventory afterwards would
// compare a constant and pass no matter what show printed. The rendered bytes
// are the only observable output it has, so those are what must be invariant.
//
// It prints with fmt.Println rather than cmd.OutOrStdout(), so harness.Out sees
// nothing and os.Stdout has to be captured instead.
func showDeviceFingerprint(t *testing.T) string {
	t.Helper()

	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	cmdtest.AddDevice(inventory, rackID, "cn-01", 10)
	parent := show.NewCommand()
	harness := cmdtest.New(t, parent, subcommand(t, parent, "device"), inventory)

	rendered := captureStdout(t, func() {
		if err := harness.Run(t, map[string][]string{"format": {"json"}}); err != nil {
			t.Fatalf("show device: %v", err)
		}
	})
	return cmdtest.MaskIDs(rendered)
}

// captureStdout redirects os.Stdout for the duration of run and returns what
// was written to it.
func captureStdout(t *testing.T, run func()) string {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	previous := os.Stdout
	os.Stdout = writer

	// Drain concurrently: a writer filling the pipe buffer would otherwise
	// block before run returns.
	drained := make(chan string, 1)
	go func() {
		var buffer bytes.Buffer
		_, _ = io.Copy(&buffer, reader)
		drained <- buffer.String()
	}()

	defer func() {
		os.Stdout = previous
		_ = writer.Close()
	}()
	run()

	os.Stdout = previous
	if err := writer.Close(); err != nil {
		t.Fatalf("closing capture pipe: %v", err)
	}
	return <-drained
}

// addRackFingerprint runs "add rack", the verb that dispatches the
// RackPostAddHook capability, and returns a UUID-free rendering of the result.
func addRackFingerprint(t *testing.T) string {
	t.Helper()

	parent := add.NewCommand()
	harness := cmdtest.New(t, parent, subcommand(t, parent, "rack"), nil)

	if err := harness.Run(t, map[string][]string{"name": {"rack-99"}}, rackSlug); err != nil {
		t.Fatalf("add rack: %v", err)
	}
	return cmdtest.Fingerprint(harness.Inventory())
}

// crudFingerprints renders every CRUD verb once, under whatever providers are
// currently registered.
func crudFingerprints(t *testing.T) map[string]string {
	t.Helper()
	return map[string]string{
		"add device":    addDeviceFingerprint(t),
		"add rack":      addRackFingerprint(t),
		"update device": updateDeviceFingerprint(t),
		"remove device": removeDeviceFingerprint(t),
		"show device":   showDeviceFingerprint(t),
	}
}

// TestCRUDResultIsInvariantAcrossRegisteredProviders verifies every CRUD verb
// produces the same result no matter which providers are registered.
//
// Why it matters: generic CRUD must remain independent of provider identity.
// Inputs: five CRUD scenarios, first without providers, then with inert and
// capable providers. Outputs: equal fingerprints and exact callback arguments.
// Data choice: a fresh test process keeps these registry states distinct even
// under repeated or shuffled runs, since registration has no deregister API.
func TestCRUDResultIsInvariantAcrossRegisteredProviders(t *testing.T) {
	if runInvarianceSubprocess(t) {
		return
	}
	cmdtest.RequireNoProviders(t)
	baseline := crudFingerprints(t)
	assertCRUDBaseline(t, baseline)
	states := []provider.Provider{
		inertProvider{slug: "invariance-alpha"},
		inertProvider{slug: "invariance-beta"},
		invarianceProvider,
	}
	for _, registeredProvider := range states {
		provider.Register(registeredProvider.Slug(), registeredProvider)
		invarianceProvider.calls = capabilityLog{}
		fingerprints := crudFingerprints(t)
		assertCRUDFingerprints(t, baseline, fingerprints, len(provider.GetProviders()))
		if registeredProvider == invarianceProvider {
			assertCRUDCapabilities(t, invarianceProvider.calls)
		}
	}
}

func assertCRUDBaseline(t *testing.T, baseline map[string]string) {
	t.Helper()
	for _, guard := range []struct{ verb, want string }{
		{"add device", `"name":"cn-01"`},
		{"add rack", `"name":"rack-99"`},
		{"update device", `"name":"cn-99"`},
		{"remove device", `"name":"cn-02"`},
		{"show device", `cn-01`},
		{"add device", `"interfaces"`},
	} {
		if !strings.Contains(baseline[guard.verb], guard.want) {
			t.Fatalf("%s baseline lacks %s:\n%s", guard.verb, guard.want, baseline[guard.verb])
		}
	}
	if strings.Contains(baseline["remove device"], `"name":"cn-01"`) {
		t.Fatalf("remove baseline still contains the removed device:\n%s", baseline["remove device"])
	}
}

func assertCRUDFingerprints(t *testing.T, baseline, fingerprints map[string]string, registered int) {
	t.Helper()
	for verb, want := range baseline {
		if got := fingerprints[verb]; got != want {
			t.Errorf("%s diverged with %d provider(s) registered:\n got:\n%s\nwant:\n%s",
				verb, registered, got, want)
		}
	}
}

func runInvarianceSubprocess(t *testing.T) bool {
	t.Helper()
	const childEnv = "CANI_TEST_INVARIANCE_CHILD"
	if os.Getenv(childEnv) == "1" {
		return false
	}
	command := exec.Command(os.Args[0], "-test.run=^"+t.Name()+"$", "-test.count=1")
	command.Env = append(os.Environ(), childEnv+"=1")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("isolated invariance test: %v\n%s", err, output)
	}
	return true
}
