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
package devicetypes

import (
	"reflect"
	"strings"
	"testing"
)

// customFieldInventory returns an inventory whose catalog defines one custom
// field of every type for devices, plus one field that applies to racks only.
func customFieldInventory(t *testing.T) *Inventory {
	t.Helper()
	inv := NewInventory()
	defs := []CustomFieldDefinition{
		{Key: "site_code", Type: "text"},
		{Key: "runbook", Type: "url"},
		{Key: "notes", Type: "markdown"},
		{Key: "spare_slots", Type: "integer"},
		{Key: "decommissioned", Type: "boolean"},
		{Key: "install_date", Type: "date", Required: true},
		{Key: "bios", Type: "json"},
		{Key: "vsx_role", Type: "select", Choices: []string{"primary", "secondary"}},
		{Key: "fabrics", Type: "multi-select", Choices: []string{"hsn", "mgmt", "storage"}},
	}
	for _, def := range defs {
		def.Label, def.ContentTypes = def.Key, []string{ContentTypeDevice}
		if err := inv.AddCustomField(def); err != nil {
			t.Fatalf("AddCustomField(%s): %v", def.Key, err)
		}
	}
	rackOnly := CustomFieldDefinition{Key: "row", Label: "row", Type: "text", ContentTypes: []string{ContentTypeRack}}
	if err := inv.AddCustomField(rackOnly); err != nil {
		t.Fatalf("AddCustomField(row): %v", err)
	}
	return inv
}

// TestSetCustomFieldStoresTheTypedValue verifies that text typed for a custom
// field is stored as the type its definition declares, for every type.
//
// Why it matters: a value written as text reaches a typed field as the wrong
// JSON type, so the system that owns the definition rejects it or stores
// something other than what was meant; false, 0 and "" are real values that
// must survive.
// Inputs: one value per field type set on a device's metadata, including
// false, 0, the empty string and a selection with spaces after the comma.
// Outputs: a bool, an int, a decoded JSON object, a list of two choices and
// the typed text for the rest.
// Data choice: the falsy value of each kind proves nothing is filtered on
// truthiness; the JSON value nests a list so decoding is visible.
func TestSetCustomFieldStoresTheTypedValue(t *testing.T) {
	// Arrange.
	inv := customFieldInventory(t)
	meta := &ObjectMeta{}
	typed := map[string]string{
		"site_code": "", "runbook": "https://wiki.example/leaf", "notes": "**check** PSU",
		"spare_slots": "0", "decommissioned": "false", "install_date": "2026-11-17",
		"bios": `{"boot":["pxe","disk"]}`, "vsx_role": "secondary", "fabrics": "hsn, mgmt",
	}

	// Act.
	for key, raw := range typed {
		if err := inv.SetCustomField(meta, ContentTypeDevice, key, raw); err != nil {
			t.Errorf("SetCustomField(%s, %q): unexpected error: %v", key, raw, err)
		}
	}

	// Assert.
	want := map[string]any{
		"site_code": "", "runbook": "https://wiki.example/leaf", "notes": "**check** PSU",
		"spare_slots": 0, "decommissioned": false, "install_date": "2026-11-17",
		"bios":     map[string]any{"boot": []any{"pxe", "disk"}},
		"vsx_role": "secondary", "fabrics": []string{"hsn", "mgmt"},
	}
	if !reflect.DeepEqual(meta.CustomFields, want) {
		t.Errorf("custom fields = %#v\nwant %#v", meta.CustomFields, want)
	}
}

// TestSetCustomFieldRejectsWhatTheDefinitionForbids verifies that a key with
// no definition for the object's kind, a value of the wrong type and a
// selection outside the choices are each refused, and that nothing is stored.
//
// Why it matters: an unchecked value was stored without complaint and was
// first rejected, or silently dropped, by the system it was exported to.
// Inputs: an undefined key, a rack-only key set on a device, and one
// malformed value for each typed field.
// Outputs: an error naming the key for every case, and metadata that still
// holds only the value it started with.
// Data choice: each bad value is one a person plausibly types: a number
// with a unit, a date in day-first order, a choice from another field.
func TestSetCustomFieldRejectsWhatTheDefinitionForbids(t *testing.T) {
	// Arrange.
	inv := customFieldInventory(t)
	meta := &ObjectMeta{CustomFields: map[string]any{"vsx_role": "primary"}}
	refused := map[string]string{
		"vsx_rol": "primary", "row": "A", "spare_slots": "4U", "decommissioned": "maybe",
		"install_date": "17/11/2026", "bios": "{boot:", "vsx_role": "hsn", "fabrics": "hsn,primary",
	}

	for key, raw := range refused {
		// Act.
		err := inv.SetCustomField(meta, ContentTypeDevice, key, raw)

		// Assert.
		if err == nil || !strings.Contains(err.Error(), `"`+key+`"`) {
			t.Errorf("SetCustomField(%s, %q) error = %v, want one naming the key", key, raw, err)
		}
	}
	if want := map[string]any{"vsx_role": "primary"}; !reflect.DeepEqual(meta.CustomFields, want) {
		t.Errorf("custom fields = %v, want them untouched: %v", meta.CustomFields, want)
	}
}

// TestClearCustomFieldNullsDefinedAndDeletesStrayKeys verifies the three
// outcomes of clearing: a defined field becomes an explicit null, a stray
// key with no definition is deleted, and a required field or an unknown key
// is refused.
//
// Why it matters: a system that already holds the object keeps the value of
// a key a write leaves out, so a defined field can only be cleared by
// sending null; a stray key has nothing on the other side and must simply
// go, or it could never be removed.
// Inputs: metadata holding a defined field, a required field and a stray
// key; each is cleared, then a key that is neither defined nor held.
// Outputs: vsx_role present with a nil value, legacy_note gone, the
// required install_date unchanged with an error, and an error for the
// unknown key.
// Data choice: the nil is asserted with the two-value lookup, because a
// deleted key also reads as nil.
func TestClearCustomFieldNullsDefinedAndDeletesStrayKeys(t *testing.T) {
	// Arrange.
	inv := customFieldInventory(t)
	meta := &ObjectMeta{CustomFields: map[string]any{
		"vsx_role": "primary", "install_date": "2026-11-17", "legacy_note": "from an old import",
	}}

	// Act.
	definedErr := inv.ClearCustomField(meta, ContentTypeDevice, "vsx_role")
	strayErr := inv.ClearCustomField(meta, ContentTypeDevice, "legacy_note")
	requiredErr := inv.ClearCustomField(meta, ContentTypeDevice, "install_date")
	unknownErr := inv.ClearCustomField(meta, ContentTypeDevice, "no_such_field")

	// Assert.
	if value, held := meta.CustomFields["vsx_role"]; definedErr != nil || !held || value != nil {
		t.Errorf("defined field: error = %v, held = %v, value = %v; want an explicit null", definedErr, held, value)
	}
	if _, held := meta.CustomFields["legacy_note"]; strayErr != nil || held {
		t.Errorf("stray key: error = %v, held = %v; want it deleted", strayErr, held)
	}
	if requiredErr == nil || meta.CustomFields["install_date"] != "2026-11-17" {
		t.Errorf("required field: error = %v, value = %v; want a refusal and the value kept", requiredErr, meta.CustomFields["install_date"])
	}
	if unknownErr == nil {
		t.Error("unknown key: expected an error, got nil")
	}
}
