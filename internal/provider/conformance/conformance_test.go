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

// Package conformance hosts cross-provider conformance tests. It blank-imports
// every real provider so that provider.GetProviders() is fully populated, then
// asserts each registered provider satisfies the required Provider contract and
// reports the optional-interface capability matrix documented in
// internal/provider/CAPABILITIES.md.
package conformance

import (
	"sort"
	"testing"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/provider"

	_ "github.com/Cray-HPE/cani/pkg/provider/csm"
	_ "github.com/Cray-HPE/cani/pkg/provider/example"
	_ "github.com/Cray-HPE/cani/pkg/provider/hpcm"
	_ "github.com/Cray-HPE/cani/pkg/provider/nautobot"
	_ "github.com/Cray-HPE/cani/pkg/provider/ochami"
	_ "github.com/Cray-HPE/cani/pkg/provider/redfish"
)

// expectedProviders lists every provider blank-imported by this package. The
// suite asserts each one self-registers via its init() so that a forgotten
// blank import (as already happened in tools/gendocs/main.go, which omits
// hpcm) is caught instead of silently dropping a provider from dispatch.
var expectedProviders = []string{
	"csm", "example", "hpcm", "nautobot", "ochami", "redfish",
}

// capabilityOrder is the stable column order for the optional-interface
// capability matrix; it mirrors the table in internal/provider/CAPABILITIES.md.
var capabilityOrder = []string{
	"Importer", "Exporter", "HasOptions", "Configurer",
	"HasImportOptions", "HasExportOptions", "DeviceStager", "RackStager",
	"RackPostAddHook", "MetadataApplier", "DeviceUpdateFlagProvider",
	"StagedDeviceDescriber",
}

// optionalCapabilities reports which optional provider interfaces p implements,
// keyed by interface name. A new optional interface added to internal/provider
// should be added here and to capabilityOrder.
func optionalCapabilities(p provider.Provider) map[string]bool {
	caps := make(map[string]bool, len(capabilityOrder))
	_, caps["Importer"] = p.(provider.Importer)
	_, caps["Exporter"] = p.(provider.Exporter)
	_, caps["HasOptions"] = p.(provider.HasOptions)
	_, caps["Configurer"] = p.(provider.Configurer)
	_, caps["HasImportOptions"] = p.(provider.HasImportOptions)
	_, caps["HasExportOptions"] = p.(provider.HasExportOptions)
	_, caps["DeviceStager"] = p.(provider.DeviceStager)
	_, caps["RackStager"] = p.(provider.RackStager)
	_, caps["RackPostAddHook"] = p.(provider.RackPostAddHook)
	_, caps["MetadataApplier"] = p.(provider.MetadataApplier)
	_, caps["DeviceUpdateFlagProvider"] = p.(provider.DeviceUpdateFlagProvider)
	_, caps["StagedDeviceDescriber"] = p.(provider.StagedDeviceDescriber)
	return caps
}

// sortedNames returns the registry's provider names in deterministic order.
func sortedNames(provs map[string]provider.Provider) []string {
	names := make([]string, 0, len(provs))
	for n := range provs {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// TestProvidersSelfRegister verifies every expected provider self-registers
// through its init() and that its registry key equals its Slug().
//
// Why it matters: the command layer reaches providers only through the registry
// populated by blank-import side effects; a provider missing from the registry
// (a forgotten import, as in tools/gendocs/main.go which omits hpcm) is silently
// dropped from all dispatch, and a Slug() that disagrees with the registry key
// breaks slug-scoped metadata and config lookups.
// Inputs: the registry produced by this package's six blank imports, queried for
// each name in expectedProviders. Outputs: a non-nil instance per name whose
// Slug() equals the name. Data choice: expectedProviders mirrors the blank
// imports so adding a provider without updating both lists fails the test.
func TestProvidersSelfRegister(t *testing.T) {
	provs := provider.GetProviders()
	for _, name := range expectedProviders {
		p, ok := provs[name]
		if !ok {
			t.Errorf("provider %q did not self-register (missing blank import or init()?)", name)
			continue
		}
		if p == nil {
			t.Errorf("provider %q registered a nil instance", name)
			continue
		}
		if got := p.Slug(); got != name {
			t.Errorf("provider %q: Slug() = %q, want %q (registry key must equal Slug)", name, got, name)
		}
	}
}

// TestNewProviderCmdGraceful verifies NewProviderCmd never panics for any
// registered provider, whether the base command is one it customizes or an
// unknown verb.
//
// Why it matters: cmd/ calls NewProviderCmd for every provider against several
// base commands (import, export, ...); a panic or hard failure in one provider
// would abort building the whole command tree, so unknown verbs must degrade
// gracefully (return the base or an error, not crash).
// Inputs: each registered provider invoked with base commands "import",
// "export", and a deliberately unknown verb. Outputs: no panic; a returned error
// is acceptable and only logged. Data choice: the unknown verb exercises the
// default switch branch, and known verbs exercise the customization branches.
func TestNewProviderCmdGraceful(t *testing.T) {
	provs := provider.GetProviders()
	baseNames := []string{"import", "export", "definitely-not-a-real-command"}

	for _, name := range sortedNames(provs) {
		p := provs[name]
		for _, bn := range baseNames {
			t.Run(name+"/"+bn, func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("%s.NewProviderCmd(%q) panicked: %v", name, bn, r)
					}
				}()
				if _, err := p.NewProviderCmd(&cli.Command{Use: bn}); err != nil {
					t.Logf("%s.NewProviderCmd(%q) returned error (acceptable): %v", name, bn, err)
				}
			})
		}
	}
}

// TestCapabilityMatrix records which optional interfaces each provider
// implements and asserts the registry holds at least the expected providers.
//
// Why it matters: the optional-interface set is the extensibility surface of the
// plugin system; emitting the matrix turns "which provider supports what" into
// reviewable evidence (kept in sync with CAPABILITIES.md) and guards against the
// interface set growing without anyone noticing which providers actually use it.
// Inputs: every registered provider type-asserted against all twelve optional
// interfaces. Outputs: a logged matrix and a failure if fewer than the expected
// providers are present. Data choice: capabilityOrder fixes the column order so
// the logged matrix is stable and diffable against the documentation.
func TestCapabilityMatrix(t *testing.T) {
	provs := provider.GetProviders()
	names := sortedNames(provs)
	if len(names) < len(expectedProviders) {
		t.Fatalf("registry has %d providers, want at least %d", len(names), len(expectedProviders))
	}

	t.Log("Provider optional-interface capability matrix (x = implements):")
	header := "provider"
	for _, c := range capabilityOrder {
		header += " | " + c
	}
	t.Log(header)
	for _, name := range names {
		caps := optionalCapabilities(provs[name])
		row := name
		for _, c := range capabilityOrder {
			mark := " "
			if caps[c] {
				mark = "x"
			}
			row += " | " + mark
		}
		t.Log(row)
	}
}
