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

package customfields

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// definedInventory returns an inventory that defines, for devices, a text, an
// integer, a multi-select, a JSON and a required date field, and one text
// field for racks only.
func definedInventory(t *testing.T) *devicetypes.Inventory {
	t.Helper()
	inv := devicetypes.NewInventory()
	defs := []devicetypes.CustomFieldDefinition{
		{Key: "site_code", Type: "text"},
		{Key: "spare_slots", Type: "integer"},
		{Key: "fabrics", Type: "multi-select", Choices: []string{"hsn", "mgmt", "storage"}},
		{Key: "bios", Type: "json"},
		{Key: "install_date", Type: "date", Required: true},
		{Key: "row", Type: "text", ContentTypes: []string{devicetypes.ContentTypeRack}},
	}
	for _, def := range defs {
		def.Label = def.Key
		if def.ContentTypes == nil {
			def.ContentTypes = []string{devicetypes.ContentTypeDevice}
		}
		if err := inv.AddCustomField(def); err != nil {
			t.Fatalf("AddCustomField(%s): %v", def.Key, err)
		}
	}
	return inv
}

// commandWith returns a command that has both flags, set as a command line
// would set them.
func commandWith(t *testing.T, set []string, cleared []string) *cli.Command {
	t.Helper()
	cmd := &cli.Command{Use: "device"}
	RegisterSet(cmd.Flags())
	RegisterClear(cmd.Flags())
	for flag, values := range map[string][]string{FlagSet: set, FlagClear: cleared} {
		for _, value := range values {
			if err := cmd.Flags().Set(flag, value); err != nil {
				t.Fatalf("set --%s %s: %v", flag, value, err)
			}
		}
	}
	return cmd
}

// TestFromFlags_StampsValuesInTheTypeTheDefinitionDeclares verifies text
// typed for --custom-field reaches a new object as the type each field's
// definition declares, with the comma of a selection left to the definition.
//
// Why it matters: the flag is the only way to give an item a custom field
// from the command line; a number or a list stored as text is rejected, or
// kept as the wrong thing, by the system the inventory is exported to.
// Inputs: --custom-field for a text, an integer and a multi-select field.
// Outputs: a string, an int and a list of two choices on the object.
// Data choice: the selection holds a comma, which a comma-splitting flag
// type would have cut into two malformed pairs.
func TestFromFlags_StampsValuesInTheTypeTheDefinitionDeclares(t *testing.T) {
	// Arrange.
	inv := definedInventory(t)
	cmd := commandWith(t, []string{"site_code=dc1", "spare_slots=4", "fabrics=hsn,mgmt"}, nil)
	var meta devicetypes.ObjectMeta

	// Act.
	edit, err := FromFlags(cmd, inv, devicetypes.ContentTypeDevice)
	edit.Stamp(&meta)

	// Assert.
	if err != nil {
		t.Fatalf("FromFlags: %v", err)
	}
	want := map[string]any{"site_code": "dc1", "spare_slots": 4, "fabrics": []string{"hsn", "mgmt"}}
	if !reflect.DeepEqual(meta.CustomFields, want) {
		t.Errorf("custom fields = %#v\nwant %#v", meta.CustomFields, want)
	}
}

// TestFromFlags_RefusalsNameTheFlag verifies every value the definitions do
// not allow is refused before anything is changed, with an error that names
// the flag and, for a key nobody has defined, says how to define it.
//
// Why it matters: an undefined key used to be stored and then dropped by
// Nautobot without a word; the refusal is only useful if it says which flag
// and what to do next, and it must not tell the operator to define a field
// that already exists for another kind of object.
// Inputs: an undefined key, a rack-only key on a device, a value of the
// wrong type, and one key both set and cleared.
// Outputs: an error holding each wanted fragment and none of the unwanted.
// Data choice: the undefined key is a plausible misspelling of a defined one.
func TestFromFlags_RefusalsNameTheFlag(t *testing.T) {
	const hint = "add metadata custom-field"
	tests := []struct {
		name    string
		set     []string
		clear   []string
		want    []string
		wantNot string
	}{
		{"key with no definition", []string{"site_cod=dc1"}, nil, []string{"--custom-field", `"site_cod"`, hint}, ""},
		{"key defined for another kind", []string{"row=A"}, nil, []string{"--custom-field", `"row"`, devicetypes.ContentTypeRack}, hint},
		{"value of the wrong type", []string{"spare_slots=4U"}, nil, []string{"--custom-field", `"spare_slots"`, "not an integer"}, hint},
		{"key set and cleared", []string{"site_code=dc1"}, []string{"site_code"}, []string{"--custom-field", "--clear-custom-field", `"site_code"`}, hint},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			cmd := commandWith(t, tc.set, tc.clear)

			// Act.
			_, err := FromFlags(cmd, definedInventory(t), devicetypes.ContentTypeDevice)

			// Assert.
			if err == nil {
				t.Fatal("FromFlags succeeded, want a refusal")
			}
			for _, want := range tc.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error = %q, want it to contain %q", err, want)
				}
			}
			if tc.wantNot != "" && strings.Contains(err.Error(), tc.wantNot) {
				t.Errorf("error = %q, want it not to contain %q", err, tc.wantNot)
			}
		})
	}
}

// TestEditStamp_GivesEachObjectItsOwnCopy verifies two objects stamped from
// one command line share neither a custom-field map nor a list or decoded
// JSON value in it, even when both began as copies of one hardware type.
//
// Why it matters: `add --qty` builds each item as a shallow copy of the
// library's type and stamps them from one parsed value; if they shared
// either, stamping or editing one item would change the others and the
// library's type itself.
// Inputs: two objects that share the map of a template holding one field,
// stamped with a multi-select and a JSON field; then the first object's
// list and nested map are changed in place.
// Outputs: the second object still holds what was typed, and the template
// holds only the field it started with.
// Data choice: the JSON nests a list in an object so a shallow copy of the
// outer map alone would still be caught.
func TestEditStamp_GivesEachObjectItsOwnCopy(t *testing.T) {
	// Arrange.
	cmd := commandWith(t, []string{"fabrics=hsn,mgmt", `bios={"boot":["pxe","disk"]}`}, nil)
	edit, err := FromFlags(cmd, definedInventory(t), devicetypes.ContentTypeDevice)
	if err != nil {
		t.Fatalf("FromFlags: %v", err)
	}
	template := devicetypes.ObjectMeta{CustomFields: map[string]any{"site_code": "dc1"}}
	first, second := template, template

	// Act.
	edit.Stamp(&first)
	edit.Stamp(&second)
	first.CustomFields["fabrics"].([]string)[0] = "storage"
	first.CustomFields["bios"].(map[string]any)["boot"].([]any)[0] = "usb"

	// Assert.
	want := map[string]any{
		"site_code": "dc1", "fabrics": []string{"hsn", "mgmt"}, "bios": map[string]any{"boot": []any{"pxe", "disk"}},
	}
	if !reflect.DeepEqual(second.CustomFields, want) {
		t.Errorf("second object's custom fields = %#v\nwant %#v", second.CustomFields, want)
	}
	if wantTemplate := map[string]any{"site_code": "dc1"}; !reflect.DeepEqual(template.CustomFields, wantTemplate) {
		t.Errorf("template's custom fields = %#v, want them untouched: %#v", template.CustomFields, wantTemplate)
	}
}

// TestEditApply_SetsThenClears verifies one command line can set a field,
// null a defined one and delete a stray one on an existing object.
//
// Why it matters: this is the whole of what `update` can do to custom
// fields; a defined field has to become an explicit null for the next export
// to clear it remotely, while a key with no definition can only be removed.
// Inputs: an object holding a defined field, a stray key and a selection;
// --custom-field for an integer, --clear-custom-field for the first two.
// Outputs: the integer set, the defined field present as nil, the stray key
// gone, the selection untouched.
// Data choice: the stray key is the kind of pair --metadata left behind.
func TestEditApply_SetsThenClears(t *testing.T) {
	// Arrange.
	inv := definedInventory(t)
	cmd := commandWith(t, []string{"spare_slots=2"}, []string{"site_code", "legacy_note"})
	meta := devicetypes.ObjectMeta{CustomFields: map[string]any{
		"site_code": "dc1", "legacy_note": "from --metadata", "fabrics": []string{"hsn"},
	}}

	// Act.
	err := Apply(cmd, inv, &meta, devicetypes.ContentTypeDevice)

	// Assert.
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	want := map[string]any{"site_code": nil, "spare_slots": 2, "fabrics": []string{"hsn"}}
	if !reflect.DeepEqual(meta.CustomFields, want) {
		t.Errorf("custom fields = %#v\nwant %#v", meta.CustomFields, want)
	}
}

// TestEditApply_LeavesObjectUntouchedWhenAClearIsRefused verifies a refused
// clear undoes the sets and clears made before it on the same command line.
//
// Why it matters: a command that fails must not have half-edited the object
// a caller may still hold.
// Inputs: a valid clear followed by a clear of a required field, then by a
// clear of a key that is neither defined nor held; each with and without a
// set in front.
// Outputs: an error naming --clear-custom-field and the key, and the object
// exactly as it was.
// Data choice: the valid edits come first, so they are the ones that would
// leak if the object were edited in place; the run without a set matters
// because setting a value already copies the object's fields.
func TestEditApply_LeavesObjectUntouchedWhenAClearIsRefused(t *testing.T) {
	for _, refused := range []string{"install_date", "no_such_field"} {
		for name, set := range map[string][]string{"after a set": {"spare_slots=2"}, "clears only": nil} {
			t.Run(refused+" "+name, func(t *testing.T) {
				// Arrange.
				cmd := commandWith(t, set, []string{"site_code", refused})
				meta := devicetypes.ObjectMeta{CustomFields: map[string]any{"site_code": "dc1", "install_date": "2026-11-17"}}

				// Act.
				err := Apply(cmd, definedInventory(t), &meta, devicetypes.ContentTypeDevice)

				// Assert.
				if err == nil || !strings.Contains(err.Error(), "--clear-custom-field") || !strings.Contains(err.Error(), `"`+refused+`"`) {
					t.Errorf("error = %v, want one naming --clear-custom-field and %q", err, refused)
				}
				want := map[string]any{"site_code": "dc1", "install_date": "2026-11-17"}
				if !reflect.DeepEqual(meta.CustomFields, want) {
					t.Errorf("custom fields = %#v, want them untouched: %#v", meta.CustomFields, want)
				}
			})
		}
	}
}

// TestFromFlags_CommandWithoutTheFlagsChangesNothing verifies a command that
// registered neither flag yields an edit that leaves an object as it is.
//
// Why it matters: `add` has no --clear-custom-field, and a noun's helper may
// be shared with a command that has neither flag; reading a flag that is not
// there must not be an error or allocate an empty map that then serializes.
// Inputs: a command with no flags and an object with no custom fields.
// Outputs: no error and a nil custom-field map after Stamp and after Apply.
// Data choice: nil rather than empty is asserted because the two serialize
// differently.
func TestFromFlags_CommandWithoutTheFlagsChangesNothing(t *testing.T) {
	// Arrange.
	cmd := &cli.Command{Use: "location"}
	var stamped, applied devicetypes.ObjectMeta

	// Act.
	edit, err := FromFlags(cmd, definedInventory(t), devicetypes.ContentTypeLocation)
	edit.Stamp(&stamped)
	applyErr := edit.Apply(&applied)

	// Assert.
	if err != nil || applyErr != nil {
		t.Fatalf("FromFlags error = %v, Apply error = %v, want neither", err, applyErr)
	}
	if stamped.CustomFields != nil || applied.CustomFields != nil {
		t.Errorf("custom fields = %v and %v, want both nil", stamped.CustomFields, applied.CustomFields)
	}
}
