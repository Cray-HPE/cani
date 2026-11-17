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
)

// definitionOf returns the catalog's definition of key, failing the test
// when there is none.
func definitionOf(t *testing.T, inventory *devicetypes.Inventory, key string) devicetypes.CustomFieldDefinition {
	t.Helper()
	for _, def := range inventory.ListCustomFields() {
		if def.Key == key {
			return def
		}
	}
	t.Fatalf("custom field %q has no definition; catalog = %+v", key, inventory.ListCustomFields())
	return devicetypes.CustomFieldDefinition{}
}

// definitionLines returns the report lines about definitions.
func definitionLines(report []string) []string {
	var lines []string
	for _, line := range report {
		if isDefinitionLine(line) {
			lines = append(lines, line)
		}
	}
	return lines
}

// TestMigrateV1Alpha9DefinesTheCustomFieldsItCreates verifies that every key
// the migration moves into custom fields gets a definition for the content
// types that hold it, typed by the values, so the migrated inventory passes
// the check that runs on load.
//
// Why it matters: a custom field without a definition is reported on every
// load and dropped by Nautobot on export; a migration that created such
// fields would turn every --metadata pair into a standing warning.
// Inputs: a device and a rack each holding four typed pairs: a string, a
// whole number, a bool and, on the device alone, a list; one key with a
// string on the device and a number on the rack.
// Outputs: one definition per key, text, integer, boolean and json by the
// values, json for the mixed key, listing both content types where both
// hold it; a report line per definition; RebuildDerivedState reports no
// custom-field conflict.
// Data choice: the number is a float64, as JSON decodes one, so inference is
// shown to work on a datastore's values and not only on Go ints.
func TestMigrateV1Alpha9DefinesTheCustomFieldsItCreates(t *testing.T) {
	// Arrange.
	inventory, ids := newCRUDInventory(t)
	node, rack := inventory.Devices[ids.nodeID], inventory.Racks[ids.rackID]
	node.ProviderMetadata = typed(map[string]any{"owner": "facilities", "slots": 4.0, "spare": true, "fabrics": []any{"hsn"}, "cost": "cc-1"})
	rack.ProviderMetadata = typed(map[string]any{"owner": "ops", "slots": 2.0, "spare": false, "cost": 4411.0})

	// Act.
	report := foldProviderMetadata(inventory)

	// Assert.
	want := map[string]struct {
		fieldType    string
		contentTypes []string
	}{
		"owner":   {"text", []string{devicetypes.ContentTypeDevice, devicetypes.ContentTypeRack}},
		"slots":   {"integer", []string{devicetypes.ContentTypeDevice, devicetypes.ContentTypeRack}},
		"spare":   {"boolean", []string{devicetypes.ContentTypeDevice, devicetypes.ContentTypeRack}},
		"fabrics": {"json", []string{devicetypes.ContentTypeDevice}},
		"cost":    {"json", []string{devicetypes.ContentTypeDevice, devicetypes.ContentTypeRack}},
	}
	for key, expected := range want {
		def := definitionOf(t, inventory, key)
		if def.Type != expected.fieldType || !reflect.DeepEqual(def.ContentTypes, expected.contentTypes) || def.Label != key {
			t.Errorf("definition of %q = %+v, want type %s for %v labelled %q", key, def, expected.fieldType, expected.contentTypes, key)
		}
		if def.Description != migratedFieldDescription {
			t.Errorf("definition of %q has description %q, want the migration's", key, def.Description)
		}
	}
	if lines := definitionLines(report); len(lines) != len(want) {
		t.Errorf("definition lines = %q, want one per key", lines)
	}
	result := inventory.RebuildDerivedState()
	if conflicts := slices.DeleteFunc(slices.Clone(result.Unresolved), func(line string) bool { return !strings.HasPrefix(line, "custom field") }); len(conflicts) != 0 {
		t.Errorf("custom-field conflicts after migration = %q, want none", conflicts)
	}
}

// TestMigrateV1Alpha9ExtendsAnExistingDefinition verifies that a moved key
// the catalog already defines keeps its definition, gains the content types
// that now hold it, and has its text values parsed as the definition's type.
//
// Why it matters: an operator who defined a field after typing values with
// --metadata expects the migration to honour that definition, not to fail
// on a duplicate or to leave text in an integer field for the check on load
// to report.
// Inputs: "slots" defined as an integer for racks; a device holds "5" and
// a rack holds "abc", both typed with --metadata.
// Outputs: the definition lists racks and devices; the device holds 5 as an
// int; the rack keeps "abc" and the report says it was left as it is,
// without printing the value.
// Data choice: one accepted and one rejected value in the same run show that
// a rejection does not stop the rest.
func TestMigrateV1Alpha9ExtendsAnExistingDefinition(t *testing.T) {
	// Arrange.
	inventory, ids := newCRUDInventory(t)
	if err := inventory.AddCustomField(devicetypes.CustomFieldDefinition{
		Key: "slots", Label: "Slots", Type: "integer", ContentTypes: []string{devicetypes.ContentTypeRack},
	}); err != nil {
		t.Fatalf("AddCustomField: %v", err)
	}
	node, rack := inventory.Devices[ids.nodeID], inventory.Racks[ids.rackID]
	node.ProviderMetadata = typed(map[string]any{"slots": "5"})
	rack.ProviderMetadata = typed(map[string]any{"slots": "abc"})

	// Act.
	report := foldProviderMetadata(inventory)

	// Assert.
	def := definitionOf(t, inventory, "slots")
	wantTypes := []string{devicetypes.ContentTypeRack, devicetypes.ContentTypeDevice}
	if def.Type != "integer" || def.Label != "Slots" || !reflect.DeepEqual(def.ContentTypes, wantTypes) {
		t.Errorf("definition = %+v, want the operator's integer field extended to %v", def, wantTypes)
	}
	if got := node.CustomFields["slots"]; got != 5 {
		t.Errorf("device slots = %v (%T), want the int 5", got, got)
	}
	if got := rack.CustomFields["slots"]; got != "abc" {
		t.Errorf("rack slots = %v, want the text left as it is", got)
	}
	wantNote := `rack "rack-a"`
	if !slices.ContainsFunc(report, func(line string) bool {
		return strings.Contains(line, wantNote) && strings.Contains(line, `custom field "slots" holds a value its definition does not accept`)
	}) {
		t.Errorf("report = %q, want a line saying the rack's value was left as it is", report)
	}
	if slices.ContainsFunc(report, func(line string) bool { return strings.Contains(line, "abc") }) {
		t.Errorf("report = %q, must not print the value", report)
	}
	if lines := definitionLines(report); len(lines) != 1 || !strings.Contains(lines[0], "also defined for dcim.device") {
		t.Errorf("definition lines = %q, want one saying slots is also defined for devices", lines)
	}
}

// TestMigrateV1Alpha9LeavesUndefinedCustomFieldsItDidNotMove verifies that a
// custom field the inventory already held without a definition gets none
// from the migration.
//
// Why it matters: releases through v0.8.4 imported custom-field values from
// Nautobot without their definitions. Guessing a type for them would clash
// with the definition the next import brings, and the HPCM step of this
// migration tells HPCM's own records by their having no definition.
// Inputs: a device holding "hpcm_uuid" as a custom field and no provider
// metadata to move.
// Outputs: no definition for "hpcm_uuid", no report line.
// Data choice: the key is one HPCM wrote, the case the rule exists for.
func TestMigrateV1Alpha9LeavesUndefinedCustomFieldsItDidNotMove(t *testing.T) {
	// Arrange.
	inventory, ids := newCRUDInventory(t)
	node := inventory.Devices[ids.nodeID]
	node.CustomFields = map[string]any{"hpcm_uuid": "2c1e3b6a"}

	// Act.
	report := foldProviderMetadata(inventory)

	// Assert.
	for _, def := range inventory.ListCustomFields() {
		if def.Key == "hpcm_uuid" {
			t.Errorf("hpcm_uuid was defined as %+v, want no definition", def)
		}
	}
	if len(report) != 0 {
		t.Errorf("report = %q, want nothing", report)
	}
}
