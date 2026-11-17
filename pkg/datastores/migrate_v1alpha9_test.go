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
package datastores

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// typed returns provider metadata as --metadata stored it: the pairs in the
// nautobot bucket.
func typed(pairs map[string]any) map[string]any {
	return map[string]any{"nautobot": pairs}
}

// assertReported fails unless the report has one line for each wanted
// fragment and no others. The lines that say which custom fields were
// defined are left out: the tests of defineMigratedFields cover them.
func assertReported(t *testing.T, report []string, wants ...string) {
	t.Helper()
	report = slices.DeleteFunc(slices.Clone(report), isDefinitionLine)
	if len(report) != len(wants) {
		t.Fatalf("report = %q, want %d line(s)", report, len(wants))
	}
	for _, want := range wants {
		if !slices.ContainsFunc(report, func(line string) bool { return strings.Contains(line, want) }) {
			t.Errorf("report = %q, want a line containing %q", report, want)
		}
	}
}

// isDefinitionLine reports whether a report line is about a custom-field
// definition the migration created or extended.
func isDefinitionLine(line string) bool {
	return strings.HasPrefix(line, `custom field "`) && strings.Contains(line, " defined for ")
}

// TestMigrateV1Alpha9MovesUnownedMetadataToCustomFields verifies the values
// no provider owns leave provider metadata for the custom fields.
//
// Why it matters: from v1alpha10 the exporter sends custom fields only, so a
// value left behind would silently stop reaching Nautobot.
// Inputs: a rack holding a pair typed with --metadata, a key outside any
// bucket, and a CSM bucket. Outputs: v1alpha10, both unowned values as custom
// fields, the CSM bucket alone in provider metadata.
// Data choice: one object carries all three kinds of holder, so the split
// between what moves and what stays is visible in a single comparison.
func TestMigrateV1Alpha9MovesUnownedMetadataToCustomFields(t *testing.T) {
	inventory, ids := newCRUDInventory(t)
	inventory.SchemaVersion = devicetypes.SchemaVersionV1Alpha9
	rack := inventory.Racks[ids.rackID]
	rack.ProviderMetadata = map[string]any{
		"nautobot":    map[string]any{"asset_owner": "facilities"},
		"cost_center": "cc-4411",
		"csm":         map[string]any{"xname": "x3000"},
	}

	migrateV1Alpha9(inventory)

	if inventory.SchemaVersion != devicetypes.SchemaVersionV1Alpha10 {
		t.Fatalf("version = %q, want %q", inventory.SchemaVersion, devicetypes.SchemaVersionV1Alpha10)
	}
	wantCustom := map[string]any{"asset_owner": "facilities", "cost_center": "cc-4411"}
	if !reflect.DeepEqual(rack.CustomFields, wantCustom) {
		t.Errorf("custom fields = %v, want %v", rack.CustomFields, wantCustom)
	}
	wantProvider := map[string]any{"csm": map[string]any{"xname": "x3000"}}
	if !reflect.DeepEqual(rack.ProviderMetadata, wantProvider) {
		t.Errorf("provider metadata = %v, want only the CSM bucket %v", rack.ProviderMetadata, wantProvider)
	}
}

// TestMigrateV1Alpha9FoldsEveryObjectKind verifies no kind of object is
// skipped.
//
// Why it matters: a kind the migration forgot would keep a nautobot bucket
// that nothing reads or can edit once --metadata is gone.
// Inputs: one object of every kind that carries metadata, and an interface
// spec on a device and on a module, each with one typed pair. Outputs: every
// one holds the pair as a custom field and no provider metadata; nothing is
// reported.
// Data choice: the kinds are listed by hand rather than discovered, so a new
// collection has to be added here and to the migration alike.
func TestMigrateV1Alpha9FoldsEveryObjectKind(t *testing.T) {
	inventory, ids := newCRUDInventory(t)
	module := &devicetypes.CaniModuleType{ID: uuid.New(), Name: "nic-1",
		Interfaces: []devicetypes.InterfaceSpec{{ID: uuid.New(), Name: "p0"}}}
	fru := &devicetypes.CaniFruType{ID: uuid.New(), Name: "dimm-0"}
	vrf := &devicetypes.CaniVRF{ID: uuid.New(), Name: "hmn"}
	inventory.Modules[module.ID], inventory.Frus[fru.ID], inventory.VRFs[vrf.ID] = module, fru, vrf
	location, rack, node := inventory.Locations[ids.locationID], inventory.Racks[ids.rackID], inventory.Devices[ids.nodeID]
	cable, vlan := inventory.Cables[ids.cableID], inventory.VLANs[ids.vlanID]
	prefix, address := inventory.Prefixes[ids.prefixID], inventory.IPAddresses[ids.ipID]
	type holder struct{ custom, provider *map[string]any }
	kinds := map[string]holder{
		"location":         {&location.CustomFields, &location.ProviderMetadata},
		"rack":             {&rack.CustomFields, &rack.ProviderMetadata},
		"device":           {&node.CustomFields, &node.ProviderMetadata},
		"device interface": {&node.Interfaces[0].CustomFields, &node.Interfaces[0].ProviderMetadata},
		"module":           {&module.CustomFields, &module.ProviderMetadata},
		"module interface": {&module.Interfaces[0].CustomFields, &module.Interfaces[0].ProviderMetadata},
		"cable":            {&cable.CustomFields, &cable.ProviderMetadata},
		"inventory item":   {&fru.CustomFields, &fru.ProviderMetadata},
		"VLAN":             {&vlan.CustomFields, &vlan.ProviderMetadata},
		"prefix":           {&prefix.CustomFields, &prefix.ProviderMetadata},
		"IP address":       {&address.CustomFields, &address.ProviderMetadata},
		"VRF":              {&vrf.CustomFields, &vrf.ProviderMetadata},
	}
	for _, kind := range kinds {
		*kind.provider = typed(map[string]any{"support_tier": "gold"})
	}

	report := foldProviderMetadata(inventory)

	assertReported(t, report)
	for name, kind := range kinds {
		if got := (*kind.custom)["support_tier"]; got != "gold" {
			t.Errorf("%s: custom field support_tier = %v, want gold", name, got)
		}
		if *kind.provider != nil {
			t.Errorf("%s: provider metadata = %v, want none left", name, *kind.provider)
		}
	}
}

// TestMigrateV1Alpha9ReplacesCustomFieldAndReportsIt verifies a typed value
// overrides a different custom field of the same key and says so without
// printing either value.
//
// Why it matters: provider metadata won that collision on every export, so
// keeping the custom field would change what Nautobot receives; and a custom
// field can hold a credential, which must not reach a log.
// Inputs: a VLAN whose custom field and typed pair disagree on one key, and
// a prefix where they agree. Outputs: the VLAN holds the typed value and is
// reported by name, ID and key; the prefix is not reported.
// Data choice: the two values are distinctive strings, so finding either in
// the report proves a leak.
func TestMigrateV1Alpha9ReplacesCustomFieldAndReportsIt(t *testing.T) {
	inventory, ids := newCRUDInventory(t)
	vlan, prefix := inventory.VLANs[ids.vlanID], inventory.Prefixes[ids.prefixID]
	vlan.CustomFields = map[string]any{"owner": "netops-secret", "tier": "gold"}
	vlan.ProviderMetadata = typed(map[string]any{"owner": "secops-secret"})
	prefix.CustomFields = map[string]any{"owner": "netops"}
	prefix.ProviderMetadata = typed(map[string]any{"owner": "netops"})

	report := foldProviderMetadata(inventory)

	wantCustom := map[string]any{"owner": "secops-secret", "tier": "gold"}
	if !reflect.DeepEqual(vlan.CustomFields, wantCustom) {
		t.Errorf("VLAN custom fields = %v, want %v", vlan.CustomFields, wantCustom)
	}
	assertReported(t, report, `VLAN "HMN" (`+ids.vlanID.String()+`): custom field "owner" held another value`)
	if line := strings.Join(report, "\n"); strings.Contains(line, "secret") {
		t.Errorf("report prints a value: %s", line)
	}
}

// TestMigrateV1Alpha9KeepsTheCopyExportSent verifies that of two copies of
// one key, the one the exporter sent becomes the custom field.
//
// Why it matters: the migration must never make live a value that an export
// was not already sending; a typed pair shadowed by another holder was dead.
// Inputs: a cable whose typed pair shares its key with a key outside the
// buckets, or with another provider's bucket. Outputs: the custom field the
// exporter's name-ordered projection produced, the other provider's bucket
// untouched, and a line for each unowned copy dropped.
// Data choice: keys and provider names are picked on both sides of
// "nautobot" in sort order, which is what decided the winner.
func TestMigrateV1Alpha9KeepsTheCopyExportSent(t *testing.T) {
	tests := []struct {
		name         string
		provider     map[string]any
		wantCustom   map[string]any
		wantProvider map[string]any
		wantReport   []string
	}{
		{
			name:       "key sorted before the bucket loses to it",
			provider:   map[string]any{"asset": "outside", "nautobot": map[string]any{"asset": "typed"}},
			wantCustom: map[string]any{"asset": "typed"},
			wantReport: []string{`provider metadata "asset" was dropped`},
		},
		{
			name:       "key sorted after the bucket wins over it",
			provider:   map[string]any{"zone": "outside", "nautobot": map[string]any{"zone": "typed"}},
			wantCustom: map[string]any{"zone": "outside"},
			wantReport: []string{`provider metadata "zone" was dropped`},
		},
		{
			name: "provider sorted after the bucket shadows the typed pair",
			provider: map[string]any{
				"nautobot": map[string]any{"bmc_type": "typed"},
				"redfish":  map[string]any{"bmc_type": "iLO"},
			},
			wantProvider: map[string]any{"redfish": map[string]any{"bmc_type": "iLO"}},
			wantReport:   []string{`provider metadata "bmc_type" was dropped`},
		},
		{
			name: "provider sorted before the bucket is overridden by the typed pair",
			provider: map[string]any{
				"csm":      map[string]any{"class": "River"},
				"nautobot": map[string]any{"class": "typed"},
			},
			wantCustom:   map[string]any{"class": "typed"},
			wantProvider: map[string]any{"csm": map[string]any{"class": "River"}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			inventory, ids := newCRUDInventory(t)
			cable := inventory.Cables[ids.cableID]
			cable.ProviderMetadata = tc.provider

			report := foldProviderMetadata(inventory)

			if !reflect.DeepEqual(cable.CustomFields, tc.wantCustom) {
				t.Errorf("custom fields = %v, want %v", cable.CustomFields, tc.wantCustom)
			}
			if !reflect.DeepEqual(cable.ProviderMetadata, tc.wantProvider) {
				t.Errorf("provider metadata = %v, want %v", cable.ProviderMetadata, tc.wantProvider)
			}
			assertReported(t, report, tc.wantReport...)
		})
	}
}

// TestMigrateV1Alpha9MovesRackNumberToHPCMBucket verifies the one key a
// provider wrote outside its bucket goes home instead of becoming a custom
// field.
//
// Why it matters: the HPCM provider owns a rack's number; as a custom field
// it would be a second copy the next HPCM import cannot update.
// Inputs: a rack with rack_number outside any bucket, as HPCM wrote it.
// Outputs: the number under the hpcm bucket, no custom field, no report.
// Data choice: the number is a float64, which is what a JSON datastore
// yields for the integer HPCM stored.
func TestMigrateV1Alpha9MovesRackNumberToHPCMBucket(t *testing.T) {
	inventory, ids := newCRUDInventory(t)
	rack := inventory.Racks[ids.rackID]
	rack.ProviderMetadata = map[string]any{"rack_number": float64(7)}

	report := foldProviderMetadata(inventory)

	want := map[string]any{"hpcm": map[string]any{"rack_number": float64(7)}}
	if !reflect.DeepEqual(rack.ProviderMetadata, want) {
		t.Errorf("provider metadata = %v, want %v", rack.ProviderMetadata, want)
	}
	if rack.CustomFields != nil {
		t.Errorf("custom fields = %v, want none", rack.CustomFields)
	}
	assertReported(t, report)
}

// TestMigrateV1Alpha9DeviceRole verifies a device's role field ends up
// holding the role an export sent, whichever holder it was read from.
//
// Why it matters: the exporter stops reading provider metadata for a role,
// so a device whose role lived there would be exported with the default role
// unless the field is filled; and a role the exporter passed over must not
// replace the one Nautobot already has.
// Inputs: a device with a role in the field, typed with --metadata, outside
// the buckets, or in another provider's bucket, alone and in combination.
// Outputs: the role field, the provider metadata left, and the report.
// Data choice: csm sorts before nautobot and redfish after it, the two
// positions that decided which bucket the exporter read first.
func TestMigrateV1Alpha9DeviceRole(t *testing.T) {
	tests := []struct {
		name         string
		role         string
		provider     map[string]any
		wantRole     string
		wantProvider map[string]any
		wantReport   []string
	}{
		{
			name:     "typed with --metadata",
			provider: typed(map[string]any{"role": "Compute"}),
			wantRole: "Compute",
		},
		{
			name:     "outside the buckets",
			provider: map[string]any{"role": "Spine"},
			wantRole: "Spine",
		},
		{
			name:         "read from another provider, which keeps its copy",
			provider:     map[string]any{"csm": map[string]any{"role": "Management"}},
			wantRole:     "Management",
			wantProvider: map[string]any{"csm": map[string]any{"role": "Management"}},
		},
		{
			name:       "role field overrode a different typed role",
			role:       "Storage",
			provider:   typed(map[string]any{"role": "Compute"}),
			wantRole:   "Storage",
			wantReport: []string{`provider metadata "role" was dropped: it was not the role export sent`},
		},
		{
			name:     "role field agrees with the typed role",
			role:     "Compute",
			provider: typed(map[string]any{"role": "Compute"}),
			wantRole: "Compute",
		},
		{
			name: "earlier provider shadowed the typed role",
			provider: map[string]any{
				"csm":      map[string]any{"role": "Management"},
				"nautobot": map[string]any{"role": "Compute"},
			},
			wantRole:     "Management",
			wantProvider: map[string]any{"csm": map[string]any{"role": "Management"}},
			wantReport:   []string{`provider metadata "role" was dropped: another copy of that key took precedence`},
		},
		{
			name: "typed role was read ahead of a later provider",
			provider: map[string]any{
				"nautobot": map[string]any{"role": "Compute"},
				"redfish":  map[string]any{"role": "BMC"},
			},
			wantRole:     "Compute",
			wantProvider: map[string]any{"redfish": map[string]any{"role": "BMC"}},
		},
		{
			name:       "role outside the buckets was read ahead of the typed role",
			provider:   map[string]any{"role": "Spine", "nautobot": map[string]any{"role": "Leaf"}},
			wantRole:   "Spine",
			wantReport: []string{`provider metadata "role" was dropped: another copy of that key took precedence`},
		},
		{
			name:       "typed value is not text",
			provider:   typed(map[string]any{"role": float64(7)}),
			wantReport: []string{`provider metadata "role" was dropped: it was not the role export sent`},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			inventory, ids := newCRUDInventory(t)
			node := inventory.Devices[ids.nodeID]
			node.Role, node.ProviderMetadata = tc.role, tc.provider

			report := foldProviderMetadata(inventory)

			if node.Role != tc.wantRole {
				t.Errorf("role = %q, want %q", node.Role, tc.wantRole)
			}
			if !reflect.DeepEqual(node.ProviderMetadata, tc.wantProvider) {
				t.Errorf("provider metadata = %v, want %v", node.ProviderMetadata, tc.wantProvider)
			}
			if node.CustomFields != nil {
				t.Errorf("custom fields = %v, want none: a device's role was never a custom field", node.CustomFields)
			}
			assertReported(t, report, tc.wantReport...)
		})
	}
}

// TestMigrateV1Alpha9RackKeptAsDevice verifies a cabinet stored as a device
// keeps the keys the exporter sent for a rack and loses the geometry it
// reserved but never sent.
//
// Why it matters: a CSM cabinet is a device mounted in its own rack; giving
// it the typed rack position would make it occupy slots in that rack, and
// its role was a custom field of the rack, not a device role.
// Inputs: a cabinet device with a typed role, height, rack position and one
// ordinary pair. Outputs: role and the pair as custom fields, the role,
// height and position fields untouched, and a line for each geometry key.
// Data choice: the values are text, which is how --metadata stored numbers.
func TestMigrateV1Alpha9RackKeptAsDevice(t *testing.T) {
	inventory, ids := newCRUDInventory(t)
	cabinet := &devicetypes.CaniDeviceType{ID: uuid.New(), Name: "x3000", Type: devicetypes.TypeCabinet, Parent: ids.rackID}
	cabinet.ProviderMetadata = typed(map[string]any{
		"role": "Cabinet", "u_height": "48", "rack_position": "1", "asset": "A-17",
	})
	inventory.Devices[cabinet.ID] = cabinet

	report := foldProviderMetadata(inventory)

	wantCustom := map[string]any{"role": "Cabinet", "asset": "A-17"}
	if !reflect.DeepEqual(cabinet.CustomFields, wantCustom) {
		t.Errorf("custom fields = %v, want %v", cabinet.CustomFields, wantCustom)
	}
	if cabinet.Role != "" || cabinet.UHeight != 0 || cabinet.RackPosition != 0 {
		t.Errorf("role=%q height=%d position=%d, want all three fields untouched",
			cabinet.Role, cabinet.UHeight, cabinet.RackPosition)
	}
	if cabinet.ProviderMetadata != nil {
		t.Errorf("provider metadata = %v, want none left", cabinet.ProviderMetadata)
	}
	assertReported(t, report,
		`provider metadata "u_height" was dropped: export never sent it`,
		`provider metadata "rack_position" was dropped: export never sent it`)
}
