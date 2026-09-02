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
	"context"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/cmd/add"
	"github.com/Cray-HPE/cani/cmd/update"
	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// inertProvider satisfies the required Provider contract and nothing else. It
// implements no optional capability interface, so a correct CRUD path has no
// way to consult it.
type inertProvider struct{ slug string }

func (p inertProvider) Transform(_ context.Context, _ devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	return &devicetypes.TransformResult{}, nil
}

func (p inertProvider) NewProviderCmd(_ *cli.Command) (*cli.Command, error) { return nil, nil }

func (p inertProvider) Slug() string { return p.slug }

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

// TestCRUDResultIsInvariantAcrossRegisteredProviders verifies add and update
// produce the same inventory no matter which providers are registered.
//
// Why it matters: the portable model's whole premise is that CRUD generalises
// over every provider, translating only at the Import/Transform/Export edges.
// Asserting merely that the verbs run with an empty registry would test a
// configuration that never occurs in production, where main.go blank-imports
// every provider. This compares outcomes across registry states instead, so a
// verb that started consulting the registry mid-path would diverge and fail.
// Inputs: the same add and the same update executed three times, with zero, one
// and two providers registered.
// Outputs: identical fingerprints across all three registry states.
// Data choice: registration only ever grows because the registry has no
// deregister, so the states are visited in increasing order within one test.
// The providers are inert and implement no optional interface, which isolates
// "a provider exists" from "a provider was asked to do something".
func TestCRUDResultIsInvariantAcrossRegisteredProviders(t *testing.T) {
	addBaseline := addDeviceFingerprint(t)
	updateBaseline := updateDeviceFingerprint(t)

	// Guard against a degenerate fingerprint: if it ever rendered nothing, or
	// dropped the very field under test, every comparison below would pass for
	// the wrong reason.
	if !strings.Contains(addBaseline, "device name=cn-01") {
		t.Fatalf("add baseline does not describe the added device:\n%s", addBaseline)
	}
	if !strings.Contains(updateBaseline, "device name=cn-99") {
		t.Fatalf("update baseline does not describe the renamed device:\n%s", updateBaseline)
	}

	for _, slug := range []string{"invariance-alpha", "invariance-beta"} {
		provider.Register(slug, inertProvider{slug: slug})

		registered := len(provider.GetProviders())
		if got := addDeviceFingerprint(t); got != addBaseline {
			t.Errorf("add device diverged with %d provider(s) registered:\n got:\n%s\nwant:\n%s",
				registered, got, addBaseline)
		}
		if got := updateDeviceFingerprint(t); got != updateBaseline {
			t.Errorf("update device diverged with %d provider(s) registered:\n got:\n%s\nwant:\n%s",
				registered, got, updateBaseline)
		}
	}
}
