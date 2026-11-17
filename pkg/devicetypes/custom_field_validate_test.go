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
	"strings"
	"testing"

	"github.com/google/uuid"
)

// deviceWithCustomFields adds a device holding the given custom fields to inv
// and returns it.
func deviceWithCustomFields(inv *Inventory, name string, fields map[string]any) *CaniDeviceType {
	device := &CaniDeviceType{ID: uuid.New(), Name: name, Type: TypeNode, ObjectMeta: ObjectMeta{CustomFields: fields}}
	inv.Devices[device.ID] = device
	return device
}

// assertOneCustomFieldReport fails unless result holds exactly one unresolved
// line and it contains every fragment.
func assertOneCustomFieldReport(t *testing.T, result *RelationshipResult, fragments ...string) {
	t.Helper()
	if len(result.Unresolved) != 1 {
		t.Fatalf("Unresolved = %q, want one line", result.Unresolved)
	}
	for _, fragment := range fragments {
		if !strings.Contains(result.Unresolved[0], fragment) {
			t.Errorf("Unresolved[0] = %q, want it to contain %q", result.Unresolved[0], fragment)
		}
	}
}

// TestValidateCustomFieldsReportsAnUndefinedKeyOncePerContentType verifies
// that a key no definition covers is reported as one line per content type,
// counting the objects that hold it and naming the first by label.
//
// Why it matters: a value an import or migration wrote under a key the
// catalog does not define is sent to Nautobot and dropped there without a
// word. The report is what tells the operator; one line per object would
// bury every command's output when a whole import is affected.
// Inputs: two devices and one rack holding "rack_pos" with no definition.
// Outputs: two unresolved lines, one for dcim.device counting 2 and naming
// the device that sorts first, one for dcim.rack counting 1.
// Data choice: device names sort opposite to insertion order so the example
// is shown to be chosen by label, not by map order.
func TestValidateCustomFieldsReportsAnUndefinedKeyOncePerContentType(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	deviceWithCustomFields(inv, "z-node", map[string]any{"rack_pos": 3.0})
	deviceWithCustomFields(inv, "a-node", map[string]any{"rack_pos": 4.0})
	rack := &CaniRackType{ID: uuid.New(), Name: "x3000", ObjectMeta: ObjectMeta{CustomFields: map[string]any{"rack_pos": 1.0}}}
	inv.Racks[rack.ID] = rack

	// Act.
	result := inv.validateCustomFields()

	// Assert.
	if len(result.Unresolved) != 2 {
		t.Fatalf("Unresolved = %q, want two lines", result.Unresolved)
	}
	deviceLine, rackLine := result.Unresolved[0], result.Unresolved[1]
	for _, want := range []string{`custom field "rack_pos" on dcim.device has no definition`, "2 dcim.device object(s)", `e.g. device "a-node"`, "cani alpha add metadata custom-field"} {
		if !strings.Contains(deviceLine, want) {
			t.Errorf("device line = %q, want it to contain %q", deviceLine, want)
		}
	}
	if !strings.Contains(rackLine, "1 dcim.rack object(s)") {
		t.Errorf("rack line = %q, want it to count one rack", rackLine)
	}
}

// TestValidateCustomFieldsReportsAKeyDefinedForOtherContentTypes verifies
// that a value held by an object whose content type the definition does not
// list is reported, naming the content types it does list.
//
// Why it matters: a custom field is defined per content type in Nautobot; a
// value on the wrong kind of object is rejected on export, and the fix is to
// extend the definition or move the value, which the line should make plain.
// Inputs: the catalog defines "row" for racks only; a device holds it.
// Outputs: one line saying "row" on dcim.device is defined for dcim.rack only.
// Data choice: the definition's own content type is in the message so the
// operator sees the fix without looking the field up.
func TestValidateCustomFieldsReportsAKeyDefinedForOtherContentTypes(t *testing.T) {
	// Arrange.
	inv := customFieldInventory(t)
	deviceWithCustomFields(inv, "n1", map[string]any{"row": "B"})

	// Act.
	result := inv.validateCustomFields()

	// Assert.
	assertOneCustomFieldReport(t, result, `custom field "row" on dcim.device is defined for dcim.rack only`, `device "n1"`)
}

// TestValidateCustomFieldsAcceptsWhatTheDefinitionsStore verifies that
// values of every kind, as the command line stores them and as JSON decodes
// them, pass without a report.
//
// Why it matters: a datastore is written as JSON and read back, so an integer
// typed on the command line comes back as a float64 and a multi-select as a
// list of any; the check must accept what a round trip produces or every
// load would report a correct inventory.
// Inputs: a device holding a value of every defined type, numbers as float64
// and the selection as []any, and a null in a field that is not required.
// Outputs: no unresolved lines.
// Data choice: the null proves an explicit clear of an optional field is not
// a problem, and 0 and false prove nothing is judged on truthiness.
func TestValidateCustomFieldsAcceptsWhatTheDefinitionsStore(t *testing.T) {
	// Arrange.
	inv := customFieldInventory(t)
	deviceWithCustomFields(inv, "n1", map[string]any{
		"site_code": "", "runbook": "https://wiki.example/leaf", "notes": nil,
		"spare_slots": 0.0, "decommissioned": false, "install_date": "2026-11-17",
		"bios": map[string]any{"boot": []any{"pxe"}}, "vsx_role": "primary", "fabrics": []any{"hsn", "mgmt"},
	})

	// Act.
	result := inv.validateCustomFields()

	// Assert.
	if len(result.Unresolved) != 0 {
		t.Errorf("Unresolved = %q, want none", result.Unresolved)
	}
}

// TestValidateCustomFieldsReportsValuesOfTheWrongKind verifies that a value
// of a kind the field's type does not store is reported, once per field,
// naming the kind it should be.
//
// Why it matters: a typed field holding the wrong kind of value is rejected by
// the system that owns the definition; the report tells the operator which
// field and what it expects, before an export fails on it.
// Inputs: one device with a fractional integer, a string boolean, a date
// that is not YYYY-MM-DD, a number in a text field, a select value outside
// the choices, a multi-select with one bad element, and a null in the
// required date field on a second device.
// Outputs: one line per field, each naming what the field stores.
// Data choice: 1.5 is a number but not a whole one, so the check is shown to
// be about integrality rather than numeric type.
func TestValidateCustomFieldsReportsValuesOfTheWrongKind(t *testing.T) {
	// Arrange.
	inv := customFieldInventory(t)
	deviceWithCustomFields(inv, "n1", map[string]any{
		"spare_slots": 1.5, "decommissioned": "yes", "install_date": "17/11/2026", "site_code": 7.0,
		"vsx_role": "tertiary", "fabrics": []any{"hsn", "wan"},
	})
	deviceWithCustomFields(inv, "n2", map[string]any{"install_date": nil})

	// Act.
	result := inv.validateCustomFields()

	// Assert.
	want := []string{
		`"decommissioned" on dcim.device holds a value that is not a boolean`,
		`"fabrics" on dcim.device holds "wan", which is not one of its choices (hsn, mgmt, storage)`,
		`"install_date" on dcim.device holds a value that is not a date (YYYY-MM-DD)`,
		`"install_date" on dcim.device is required but holds null`,
		`"site_code" on dcim.device holds a value that is not text`,
		`"spare_slots" on dcim.device holds a value that is not an integer`,
		`"vsx_role" on dcim.device holds "tertiary", which is not one of its choices (primary, secondary)`,
	}
	if len(result.Unresolved) != len(want) {
		t.Fatalf("Unresolved = %q, want %d lines", result.Unresolved, len(want))
	}
	for i, fragment := range want {
		if !strings.Contains(result.Unresolved[i], fragment) {
			t.Errorf("Unresolved[%d] = %q, want it to contain %q", i, result.Unresolved[i], fragment)
		}
	}
}

// TestValidateCustomFieldsCoversInterfaceSpecs verifies that a custom field
// on an interface a device's type declares is checked as dcim.interface and
// reported with the interface and its device.
//
// Why it matters: interface specs carry their own custom fields but are not
// records of the inventory, so a walk over the collections alone would miss
// them; a value a provider wrote there would go unreported.
// Inputs: a device whose one interface spec holds an undefined key.
// Outputs: one line for dcim.interface naming the interface and the device.
// Data choice: the device itself holds no custom fields, so the line can only
// have come from the spec.
func TestValidateCustomFieldsCoversInterfaceSpecs(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	device := deviceWithCustomFields(inv, "n1", nil)
	device.Interfaces = []InterfaceSpec{{Name: "mgmt0", CustomFields: map[string]any{"speed_class": "fast"}}}

	// Act.
	result := inv.validateCustomFields()

	// Assert.
	assertOneCustomFieldReport(t, result, `custom field "speed_class" on dcim.interface has no definition`, `interface "mgmt0" of device "n1"`)
}

// TestRebuildDerivedStateReportsCustomFieldProblems verifies that the check
// runs as part of rebuilding derived state, so every load and every mutation
// command surfaces it.
//
// Why it matters: the check is only useful if it runs where inventories are
// read; wiring it into RebuildDerivedState is what makes an import's or a
// migration's stray values visible without a dedicated command.
// Inputs: an inventory with one device holding an undefined key.
// Outputs: RebuildDerivedState returns the custom-field line among its
// unresolved conflicts and no errors.
// Data choice: an otherwise empty inventory, so the one line is the check's.
func TestRebuildDerivedStateReportsCustomFieldProblems(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	deviceWithCustomFields(inv, "n1", map[string]any{"rack_pos": 3.0})

	// Act.
	result := inv.RebuildDerivedState()

	// Assert.
	if result.HasErrors() {
		t.Fatalf("RebuildDerivedState() errors = %v, want none", result.Errors)
	}
	assertOneCustomFieldReport(t, result, `custom field "rack_pos" on dcim.device has no definition`)
}
