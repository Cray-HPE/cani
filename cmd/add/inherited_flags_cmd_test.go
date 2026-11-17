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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// inheritedFlagSite is the location the inherited-flag inventory holds.
const inheritedFlagSite = "audit-site"

// inheritedFlagField is a custom field the inherited-flag inventory defines
// for every kind of item `add` creates.
const inheritedFlagField = "audit_note"

// addNounCase is one `add` invocation, after the word "add", that succeeds
// on its own against the inventory runAddNoun builds.
type addNounCase struct {
	name string
	argv []string
}

// addNounCases lists every `add` noun, and the bare slug form, with the
// least each needs to succeed.
func addNounCases(t *testing.T) []addNounCase {
	t.Helper()
	emptyMap := filepath.Join(t.TempDir(), "connections.yml")
	if err := os.WriteFile(emptyMap, []byte("version: v1\nconnections: []\n"), 0o600); err != nil {
		t.Fatalf("write connection map: %v", err)
	}
	return []addNounCase{
		{"device", []string{"device", cmdtest.DeviceSlug, "--rack", "rack-01", "--position", "10", "--name", "dev-01"}},
		{"rack", []string{"rack", roleTestRackSlug, "--name", "rack-02"}},
		{"module", []string{"module", roleTestModuleSlug, "--spare", "--location", inheritedFlagSite, "--name", "mod-01"}},
		{"location", []string{"location", "dc", "--name", "loc-01"}},
		{"cable", []string{"cable", roleTestCableSlug, "--label", "cab-01"}},
		{"vlan", []string{"vlan", "100", "--name", "vlan-100"}},
		{"prefix", []string{"prefix", "10.20.0.0/24"}},
		{"ip", []string{"ip", "10.20.0.5/24"}},
		{"vrf", []string{"vrf", "blue"}},
		{"interface", []string{"interface", "lag9", "--device", "dev-00", "--type", "lag"}},
		{"metadata role", []string{"metadata", "role", "AuditRole"}},
		{"metadata status", []string{"metadata", "status", "Maintenance"}},
		{"metadata tag", []string{"metadata", "tag", "audit"}},
		{"metadata custom-field", []string{"metadata", "custom-field", "audit_cf", "--content-types", "dcim.device"}},
		{"connections", []string{"connections", emptyMap}},
		{"connections generate", []string{"connections", "generate", "star", "--hub", "dev-00", "--spokes", "node-{01..02}"}},
		{"bare device slug", []string{cmdtest.DeviceSlug, "--name", "dev-02"}},
		{"bare rack slug", []string{roleTestRackSlug, "--name", "rack-03"}},
		{"bare module slug", []string{roleTestModuleSlug, "--name", "mod-02"}},
		{"bare cable slug", []string{roleTestCableSlug, "--name", "cab-02"}},
	}
}

// itemContentTypes lists the content type of every kind of item `add`
// creates, an interface included.
var itemContentTypes = []string{
	devicetypes.ContentTypeDevice, devicetypes.ContentTypeRack, devicetypes.ContentTypeModule,
	devicetypes.ContentTypeLocation, devicetypes.ContentTypeCable, devicetypes.ContentTypeVLAN,
	devicetypes.ContentTypePrefix, devicetypes.ContentTypeIPAddress, devicetypes.ContentTypeVRF,
	devicetypes.ContentTypeInterface,
}

// runAddNoun runs `add <argv>` through the full parser against a fresh
// inventory holding one location, one rack, one device and a definition of
// inheritedFlagField for every kind of item, and returns the resulting
// inventory's fingerprint.
func runAddNoun(t *testing.T, argv []string) (string, error) {
	t.Helper()
	return runAddNounDefining(t, argv, itemContentTypes...)
}

// runAddNounDefining is runAddNoun with inheritedFlagField defined for the
// given content types only.
func runAddNounDefining(t *testing.T, argv []string, contentTypes ...string) (string, error) {
	t.Helper()
	harness := addNounHarness(t, contentTypes...)
	err := harness.Execute(t, append([]string{"add"}, argv...)...)
	return cmdtest.Fingerprint(harness.Inventory()), err
}

// addNounHarness returns a harness over a fresh inventory holding one
// location, one rack, one device and a definition of inheritedFlagField for
// the given content types.
func addNounHarness(t *testing.T, contentTypes ...string) *cmdtest.Harness {
	t.Helper()
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	for _, location := range inventory.Locations {
		location.Name = inheritedFlagSite
	}
	cmdtest.AddDevice(inventory, rackID, "dev-00", 1)
	cmdtest.DefineCustomField(inventory, inheritedFlagField, "text", contentTypes...)
	return cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)
}

// TestAddInheritedFlagsTakeEffectOrFail verifies that --tag, --status and
// --custom-field, which `add` hands to every subcommand, change what each
// noun stores or are refused by name; no noun parses one and drops it.
//
// Why it matters: `add module --tag`, `add location --status` and
// `add cable --tag` exited 0 and stored nothing, so a value an operator
// typed vanished without a word and only an export showed it was missing.
// Inputs: every `add` noun and the bare slug form, run with no extra flag,
// then with --tag audit, then with --status Planned, then with
// --custom-field audit_note=checked.
// Outputs: for each noun and flag, either an inventory that differs from
// the flagless run or an error naming the flag.
// Data choice: the flagless run is repeated to prove the fingerprint is
// stable, so a difference can only come from the flag; Planned is no
// noun's default status; the custom field is defined for every kind of
// item, an interface included, so a noun that refuses it does so because
// it has nowhere to send it and not for want of a definition.
func TestAddInheritedFlagsTakeEffectOrFail(t *testing.T) {
	inherited := [][]string{
		{"--tag", "audit"}, {"--status", "Planned"}, {"--custom-field", inheritedFlagField + "=checked"},
	}

	for _, tc := range addNounCases(t) {
		// Arrange.
		baseline, err := runAddNoun(t, tc.argv)
		if err != nil {
			t.Errorf("add %s: the flagless run failed, so the case proves nothing: %v", tc.name, err)
			continue
		}
		if again, _ := runAddNoun(t, tc.argv); again != baseline {
			t.Errorf("add %s: two flagless runs differ, so the fingerprint cannot isolate a flag", tc.name)
			continue
		}

		for _, flag := range inherited {
			// Act.
			got, err := runAddNoun(t, append(append([]string{}, tc.argv...), flag...))

			// Assert.
			switch {
			case err != nil && !strings.Contains(err.Error(), flag[0]):
				t.Errorf("add %s %s: failed without naming the flag: %v", tc.name, flag[0], err)
			case err == nil && got == baseline:
				t.Errorf("add %s %s: accepted and dropped; the flag must change the inventory or be refused", tc.name, flag[0])
			}
		}
	}
}
