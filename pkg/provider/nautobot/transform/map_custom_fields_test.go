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
package transform

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// rawCustomField builds a Nautobot custom-field definition of the given
// type and content types.
func rawCustomField(id uuid.UUID, key, typ string, contentTypes ...string) nautobotapi.CustomField {
	required := true
	weight := 150
	cf := nautobotapi.CustomField{
		Id:           oaPtr(id),
		Key:          strPtr(key),
		Label:        strings.ToUpper(key[:1]) + key[1:],
		Description:  strPtr("imported"),
		ContentTypes: contentTypes,
		Required:     &required,
		Weight:       &weight,
		Default:      "n/a",
	}
	setNBValue(&cf.Type, typ)
	return cf
}

// rawChoice builds a choice of the given field with a weight.
func rawChoice(fieldID uuid.UUID, value string, weight int) nautobotapi.CustomFieldChoice {
	c := nautobotapi.CustomFieldChoice{Value: value, Weight: intPtr(weight)}
	setNBRef(&c.CustomField, fieldID)
	return c
}

// TestMapCustomFields_CarriesDefinitionAndOrderedChoices verifies that a
// select field arrives in the catalog with every attribute the export needs
// to recreate it and with its choices in Nautobot's weight order.
//
// Why it matters: a fresh destination has no definitions; EnsureCustomFields
// can only create what the catalog holds, so each attribute the create
// request sends must survive the import.
// Inputs: one select field on dcim.device and dcim.rack with three choices
// whose weights are out of list order.
// Outputs: one definition with key, label, type, both content types sorted,
// description, required, weight, default and choices [low, mid, high].
// Data choice: choice weights 300/100/200 given as high/low/mid prove
// ordering comes from weight, not input order, and are the weights the
// export writes back, so no attribute is reported lost.
func TestMapCustomFields_CarriesDefinitionAndOrderedChoices(t *testing.T) {
	fieldID := uuid.New()
	raws := []nautobotapi.CustomField{rawCustomField(fieldID, "tier", "select", "dcim.rack", "dcim.device")}
	choices := []nautobotapi.CustomFieldChoice{
		rawChoice(fieldID, "high", 300), rawChoice(fieldID, "low", 100), rawChoice(fieldID, "mid", 200),
		rawChoice(uuid.New(), "other-field", 5),
	}

	defs, warnings := MapCustomFields(raws, choices)

	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	want := []devicetypes.CustomFieldDefinition{{
		Key: "tier", Label: "Tier", Type: "select", ContentTypes: []string{"dcim.device", "dcim.rack"},
		Description: "imported", Required: true, Default: "n/a", Choices: []string{"low", "mid", "high"}, Weight: 150,
	}}
	if !reflect.DeepEqual(defs, want) {
		t.Errorf("defs = %+v, want %+v", defs, want)
	}
}

// TestMapCustomFields_KeepsOnlyContentTypesCANIExports verifies that content
// types outside CANI's vocabulary are dropped with a warning, and that a
// definition left with none is not imported.
//
// Why it matters: the catalog rejects unknown content types, and an import
// that handed one through would abort the whole merge; a definition for
// virtual machines only has no CANI object to carry it.
// Inputs: a field on dcim.device and virtualization.virtualmachine, and a
// field on virtualization.virtualmachine alone.
// Outputs: one definition, on dcim.device only, and two warnings naming
// each field.
// Data choice: the same foreign content type in both roles separates "trim"
// from "skip".
func TestMapCustomFields_KeepsOnlyContentTypesCANIExports(t *testing.T) {
	raws := []nautobotapi.CustomField{
		rawCustomField(uuid.New(), "owner", "text", "dcim.device", "virtualization.virtualmachine"),
		rawCustomField(uuid.New(), "vm_only", "text", "virtualization.virtualmachine"),
	}

	defs, warnings := MapCustomFields(raws, nil)

	if len(defs) != 1 || defs[0].Key != "owner" || !reflect.DeepEqual(defs[0].ContentTypes, []string{"dcim.device"}) {
		t.Fatalf("defs = %+v, want only owner on dcim.device", defs)
	}
	if len(warnings) != 2 || !strings.Contains(warnings[0], `"owner"`) || !strings.Contains(warnings[1], `"vm_only"`) {
		t.Errorf("warnings = %v, want one for owner's trimmed type and one for vm_only being skipped", warnings)
	}
}

// TestTransformRaw_PutsCustomFieldDefinitionsInMetadata verifies the
// pipeline hands the mapped definitions to the result's metadata catalog,
// where the inventory merge picks them up.
//
// Why it matters: mergeTransformMetadata reads TransformResult.Metadata; a
// mapper that is never wired into the result changes nothing on disk.
// Inputs: raw data holding one custom field and nothing else.
// Outputs: result.Metadata.CustomFields with that one key; a raw set without
// definitions leaves Metadata nil.
// Data choice: the empty case guards the save format of inventories that
// never carried a catalog.
func TestTransformRaw_PutsCustomFieldDefinitionsInMetadata(t *testing.T) {
	raw := &RawData{CustomFields: []nautobotapi.CustomField{rawCustomField(uuid.New(), "tier", "text", "dcim.device")}}

	result, err := Transform(devicetypes.Inventory{}, raw)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}
	if result.Metadata == nil || len(result.Metadata.CustomFields) != 1 || result.Metadata.CustomFields[0].Key != "tier" {
		t.Errorf("Metadata = %+v, want one custom field tier", result.Metadata)
	}

	empty, err := Transform(devicetypes.Inventory{}, &RawData{})
	if err != nil {
		t.Fatalf("Transform(empty) error = %v", err)
	}
	if empty.Metadata != nil {
		t.Errorf("Metadata = %+v for no definitions, want nil", empty.Metadata)
	}
}
