/*
 *
 *  MIT License
 *
 *  (C) Copyright 2024-2026 Hewlett Packard Enterprise Development LP
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
package provider

import (
	"context"
	"testing"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// fakeProvider is a minimal Provider implementation used to exercise the
// registry without importing any real provider package (which would create an
// import cycle). It records only its slug; the other methods are no-ops.
type fakeProvider struct {
	slug string
}

func (f fakeProvider) Transform(ctx context.Context, existing devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	return nil, nil
}

func (f fakeProvider) NewProviderCmd(base *cli.Command) (*cli.Command, error) {
	return nil, nil
}

func (f fakeProvider) Slug() string { return f.slug }

// TestRegisterAndGetProvider verifies Register stores a provider under its name
// and GetProvider returns that instance, or nil for an unregistered name.
//
// Why it matters: the registry is the single lookup point the command layer
// uses to reach providers without hard-coding their packages; a registration
// that did not round-trip, or a missing-name lookup that did not return nil,
// would break provider dispatch or panic callers.
// Inputs: a fakeProvider registered under a unique test name, then GetProvider
// called with that name and with an unregistered name. Outputs: the same
// instance for the registered name and nil for the missing one. Data choice: a
// unique test-only name avoids colliding with any provider an init() might have
// registered, and cleanup deletes it so the global map is left unchanged.
func TestRegisterAndGetProvider(t *testing.T) {
	const name = "fake-registry-test"
	Register(name, fakeProvider{slug: name})
	t.Cleanup(func() { delete(providers, name) })

	got := GetProvider(name)
	if got == nil {
		t.Fatalf("GetProvider(%q) = nil, want registered provider", name)
	}
	if got.Slug() != name {
		t.Errorf("Slug() = %q, want %q", got.Slug(), name)
	}

	if GetProvider("no-such-provider-xyz") != nil {
		t.Error("GetProvider(unregistered) should return nil")
	}
}

// TestGetProviders verifies GetProviders returns a snapshot of the registry
// including a freshly registered provider.
//
// Why it matters: the command layer ranges over GetProviders to dispatch
// optional interfaces to every provider, so the returned map must contain all
// registered entries; a stale or filtered copy would silently skip providers.
// Inputs: a fakeProvider registered under a unique name, then a call to
// GetProviders. Outputs: a map containing that name mapped to the instance.
// Data choice: asserting only on the test-registered key keeps the test robust
// regardless of how many other providers are present in the shared map.
func TestGetProviders(t *testing.T) {
	const name = "fake-getproviders-test"
	Register(name, fakeProvider{slug: name})
	t.Cleanup(func() { delete(providers, name) })

	all := GetProviders()
	p, ok := all[name]
	if !ok {
		t.Fatalf("GetProviders() missing %q", name)
	}
	if p.Slug() != name {
		t.Errorf("GetProviders()[%q].Slug() = %q, want %q", name, p.Slug(), name)
	}
}

// TestRegisterPanicsOnDuplicate verifies Register panics when a name is
// registered a second time instead of silently overwriting the first provider.
//
// Why it matters: providers self-register by slug in init(); two providers
// claiming the same slug is a wiring bug, and a silent overwrite would drop one
// provider from dispatch in a way that is hard to diagnose. Failing loudly at
// startup surfaces the conflict immediately.
// Inputs: a fakeProvider registered once under a unique name, then a second
// Register call with the same name. Outputs: the second call panics. Data
// choice: a unique test-only name avoids colliding with other registrations and
// is removed in cleanup so the global map is left unchanged.
func TestRegisterPanicsOnDuplicate(t *testing.T) {
	const name = "fake-dup-test"
	Register(name, fakeProvider{slug: name})
	t.Cleanup(func() { delete(providers, name) })

	defer func() {
		if recover() == nil {
			t.Error("Register did not panic on duplicate registration")
		}
	}()
	Register(name, fakeProvider{slug: name})
}

// TestRegisterPanicsOnNil verifies Register panics when given a nil provider.
//
// Why it matters: a nil provider stored in the registry would panic the command
// layer later, far from the faulty init() that stored it; rejecting nil at
// registration time keeps the failure close to its cause.
// Inputs: a Register call with a non-empty name and a nil Provider. Outputs:
// the call panics and nothing is stored. Data choice: a test-only name plus the
// panic recovery proves nil is rejected before insertion.
func TestRegisterPanicsOnNil(t *testing.T) {
	t.Cleanup(func() { delete(providers, "fake-nil-test") })
	defer func() {
		if recover() == nil {
			t.Error("Register did not panic on nil provider")
		}
	}()
	Register("fake-nil-test", nil)
}
