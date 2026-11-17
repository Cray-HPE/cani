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
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// inheritedFlagSite is the location the inherited-flag inventory holds.
const inheritedFlagSite = "audit-site"

// inheritedFlagField is a custom field the inherited-flag inventory defines.
const inheritedFlagField = "audit_note"

// itemContentTypes lists the content type of every kind of item `update`
// edits, an interface included.
var itemContentTypes = []string{
	devicetypes.ContentTypeDevice, devicetypes.ContentTypeRack, devicetypes.ContentTypeModule,
	devicetypes.ContentTypeLocation, devicetypes.ContentTypeCable, devicetypes.ContentTypePrefix,
	devicetypes.ContentTypeVRF, devicetypes.ContentTypeInterface,
}

// updateNounCases lists every `update` noun with an invocation, after the
// word "update", that succeeds on its own against inheritedFlagInventory.
var updateNounCases = []struct {
	name string
	argv []string
}{
	{"device", []string{"device", "dev-00", "--description", "base"}},
	{"rack", []string{"rack", "rack-01", "--description", "base"}},
	{"module", []string{"module", "mod-00", "--description", "base"}},
	{"location", []string{"location", inheritedFlagSite, "--description", "base"}},
	{"cable", []string{"cable", "cab-00", "--description", "base"}},
	{"interface", []string{"interface", "--device", "dev-00", "--name", "eth0", "--description", "base"}},
	{"prefix", []string{"prefix", "10.20.0.0/24", "--description", "base"}},
	{"vrf", []string{"vrf", "blue", "--description", "base"}},
	{"orphans", []string{"orphans"}},
}

// inheritedFlagInventory holds one of everything `update` can address: a
// location, a rack, a device with one interface, a spare module, a cable,
// a prefix and a VRF. It defines inheritedFlagField for the given content
// types.
func inheritedFlagInventory(contentTypes ...string) *devicetypes.Inventory {
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	var siteID uuid.UUID
	for id, location := range inventory.Locations {
		siteID, location.Name = id, inheritedFlagSite
	}
	deviceID := cmdtest.AddDevice(inventory, rackID, "dev-00", 1)
	inventory.Devices[deviceID].Interfaces = []devicetypes.InterfaceSpec{{ID: uuid.New(), Name: "eth0", Type: "1000base-t"}}

	moduleID, cableID, prefixID, vrfID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inventory.Modules[moduleID] = &devicetypes.CaniModuleType{ID: moduleID, Name: "mod-00", Location: siteID}
	inventory.Cables[cableID] = &devicetypes.CaniCableType{ID: cableID, Label: "cab-00"}
	inventory.Prefixes[prefixID] = &devicetypes.CaniPrefix{ID: prefixID, Prefix: "10.20.0.0/24"}
	inventory.VRFs[vrfID] = &devicetypes.CaniVRF{ID: vrfID, Name: "blue"}
	cmdtest.DefineCustomField(inventory, inheritedFlagField, "text", contentTypes...)
	return inventory
}

// runUpdateNoun runs `update <argv>` through the full parser against a fresh
// inventory that defines inheritedFlagField for every kind of item, and
// returns the resulting inventory's fingerprint.
func runUpdateNoun(t *testing.T, argv []string) (string, error) {
	t.Helper()
	harness := updateNounHarness(t, itemContentTypes...)
	err := harness.Execute(t, append([]string{"update"}, argv...)...)
	return cmdtest.Fingerprint(harness.Inventory()), err
}

// updateNounHarness returns a harness over a fresh inheritedFlagInventory.
func updateNounHarness(t *testing.T, contentTypes ...string) *cmdtest.Harness {
	t.Helper()
	return cmdtest.New(t, NewCommand(), newDeviceCommand(), inheritedFlagInventory(contentTypes...))
}

// TestUpdateInheritedFlagsTakeEffectOrFail verifies that --tag, --set,
// --custom-field and --clear-custom-field, which `update` hands to every
// subcommand, change what each noun stores or are refused by name; no noun
// parses one and drops it.
//
// Why it matters: `update module --tag`, `update location --tag` and
// `update cable --tag` exited 0 and stored nothing, and --set was ignored by
// the nouns that have no key table, so an edit an operator typed vanished.
// Inputs: every `update` noun run with a description alone, then with each
// of --tag audit, --set description=audited, --custom-field
// audit_note=checked and --clear-custom-field audit_note added.
// Outputs: for each noun and flag, either an inventory that differs from
// the description-only run or an error naming the flag.
// Data choice: the --set value differs from the description the baseline
// writes, so applying it is visible; the baseline is repeated to prove the
// fingerprint is stable; the custom field is defined for every kind of
// item, so a noun that refuses it has nowhere to send it; clearing a
// defined field an item does not hold stores an explicit null, which shows.
func TestUpdateInheritedFlagsTakeEffectOrFail(t *testing.T) {
	inherited := [][]string{
		{"--tag", "audit"}, {"--set", "description=audited"},
		{"--custom-field", inheritedFlagField + "=checked"}, {"--clear-custom-field", inheritedFlagField},
	}

	for _, tc := range updateNounCases {
		// Arrange.
		baseline, err := runUpdateNoun(t, tc.argv)
		if err != nil {
			t.Errorf("update %s: the baseline run failed, so the case proves nothing: %v", tc.name, err)
			continue
		}
		if again, _ := runUpdateNoun(t, tc.argv); again != baseline {
			t.Errorf("update %s: two baseline runs differ, so the fingerprint cannot isolate a flag", tc.name)
			continue
		}

		for _, flag := range inherited {
			// Act.
			got, err := runUpdateNoun(t, append(append([]string{}, tc.argv...), flag...))

			// Assert.
			switch {
			case err != nil && !strings.Contains(err.Error(), flag[0]):
				t.Errorf("update %s %s: failed without naming the flag: %v", tc.name, flag[0], err)
			case err == nil && got == baseline:
				t.Errorf("update %s %s: accepted and dropped; the flag must change the inventory or be refused", tc.name, flag[0])
			}
		}
	}
}

// updateNounContentTypes names, for each updateNounCases entry that stores
// custom fields, the content type its item is checked against.
var updateNounContentTypes = map[string]string{
	"device":   devicetypes.ContentTypeDevice,
	"rack":     devicetypes.ContentTypeRack,
	"module":   devicetypes.ContentTypeModule,
	"location": devicetypes.ContentTypeLocation,
	"cable":    devicetypes.ContentTypeCable,
	"prefix":   devicetypes.ContentTypePrefix,
	"vrf":      devicetypes.ContentTypeVRF,
}

// TestUpdateChecksCustomFieldAgainstTheNounsOwnKind verifies each noun
// accepts a custom field defined for the kind of item it edits and for no
// other.
//
// Why it matters: a definition lists the kinds of object it applies to, and
// Nautobot drops a value sent for any other kind; a noun that checked
// against the wrong kind would refuse a valid field or store one that never
// arrives.
// Inputs: every noun that stores custom fields, run with --custom-field
// against an inventory that defines the field for that noun's kind only,
// then for interfaces only.
// Outputs: one save in the first run, and in the second an error naming the
// flag and no save.
// Data choice: one content type per run, so a noun using another noun's
// kind fails the first run; interfaces are the kind no listed noun edits.
func TestUpdateChecksCustomFieldAgainstTheNounsOwnKind(t *testing.T) {
	for _, tc := range updateNounCases {
		contentType, stores := updateNounContentTypes[tc.name]
		if !stores {
			continue
		}
		// Arrange.
		argv := append(append([]string{"update"}, tc.argv...), "--custom-field", inheritedFlagField+"=checked")

		// Act. A harness installs its store when it is made, so each is run
		// before the next is made.
		own := updateNounHarness(t, contentType)
		ownErr := own.Execute(t, argv...)
		other := updateNounHarness(t, devicetypes.ContentTypeInterface)
		otherErr := other.Execute(t, argv...)

		// Assert.
		if ownErr != nil || own.Store.Saves != 1 {
			t.Errorf("update %s: a field defined for %s: error = %v, saves = %d, want it stored", tc.name, contentType, ownErr, own.Store.Saves)
		}
		if otherErr == nil || !strings.Contains(otherErr.Error(), "--custom-field") {
			t.Errorf("update %s: a field defined for interfaces only: error = %v, want a refusal naming the flag", tc.name, otherErr)
		}
		if other.Store.Saves != 0 {
			t.Errorf("update %s: a refused field was followed by %d save(s)", tc.name, other.Store.Saves)
		}
	}
}
