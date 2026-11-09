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
	"strings"
	"testing"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// TestMapCustomFields_NamesUnmodelledAttributesWithoutValues verifies that a
// definition carrying attributes the catalog does not model is imported with
// one warning that names each of them and none of their values.
//
// Why it matters: the catalog drops these attributes, so a fresh destination
// gets a definition without its validation, filtering or placement; the
// mapping guide marks them Unsupported, which promises they are named.
// Inputs: a select field with validation minimum, maximum and regex, exact
// filter logic, a grouping, advanced UI, a scope filter, and choices weighted
// 1 and 2.
// Outputs: the definition is still imported; one warning names all eight
// attributes; no value appears in it.
// Data choice: each value is a distinct sentinel, so a warning that leaked
// any of them fails.
func TestMapCustomFields_NamesUnmodelledAttributesWithoutValues(t *testing.T) {
	fieldID := uuid.New()
	raw := rawCustomField(fieldID, "tier", "select", "dcim.device")
	minimum, maximum := int64(4242), int64(9191)
	advanced := true
	raw.ValidationMinimum, raw.ValidationMaximum = &minimum, &maximum
	raw.ValidationRegex, raw.Grouping = strPtr("^SENTINEL-RE$"), strPtr("SENTINEL-GROUP")
	raw.AdvancedUi = &advanced
	raw.FilterLogic = &struct {
		Label *nautobotapi.CustomFieldFilterLogicLabel `json:"label,omitempty"`
		Value *nautobotapi.CustomFieldFilterLogicValue `json:"value,omitempty"`
	}{Value: filterLogic("exact")}
	raw.ScopeFilter = map[string]interface{}{"SENTINEL-SCOPE": "x"}
	choices := []nautobotapi.CustomFieldChoice{rawChoice(fieldID, "low", 1), rawChoice(fieldID, "high", 2)}

	defs, warnings := MapCustomFields([]nautobotapi.CustomField{raw}, choices)

	if len(defs) != 1 {
		t.Fatalf("defs = %+v, want the definition imported", defs)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings = %v, want one", warnings)
	}
	for _, name := range []string{"validation_minimum", "validation_maximum", "validation_regex", "filter_logic", "grouping", "advanced_ui", "scope_filter", "choice weights"} {
		if !strings.Contains(warnings[0], name) {
			t.Errorf("warning %q does not name %s", warnings[0], name)
		}
	}
	for _, value := range []string{"4242", "9191", "SENTINEL", "exact"} {
		if strings.Contains(warnings[0], value) {
			t.Errorf("warning %q discloses the value %q", warnings[0], value)
		}
	}
}

// TestMapCustomFields_DefaultAttributesAreNotReported verifies that a
// definition whose unmodelled attributes all hold Nautobot's defaults
// produces no warning.
//
// Why it matters: a warning on every definition would bury the ones that
// lose something; the defaults are what a fresh destination gives anyway.
// Inputs: loose filter logic, advanced UI off, an empty grouping, regex and
// scope filter, and choices weighted 100 and 200.
// Outputs: no warnings.
// Data choice: every attribute is present and default rather than absent,
// which is how Nautobot answers.
func TestMapCustomFields_DefaultAttributesAreNotReported(t *testing.T) {
	fieldID := uuid.New()
	raw := rawCustomField(fieldID, "tier", "select", "dcim.device")
	advanced := false
	raw.AdvancedUi, raw.ValidationRegex, raw.Grouping = &advanced, strPtr(""), strPtr("")
	raw.FilterLogic = &struct {
		Label *nautobotapi.CustomFieldFilterLogicLabel `json:"label,omitempty"`
		Value *nautobotapi.CustomFieldFilterLogicValue `json:"value,omitempty"`
	}{Value: filterLogic("loose")}
	raw.ScopeFilter = map[string]interface{}{}
	choices := []nautobotapi.CustomFieldChoice{rawChoice(fieldID, "low", 100), rawChoice(fieldID, "high", 200)}

	_, warnings := MapCustomFields([]nautobotapi.CustomField{raw}, choices)

	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
}

func filterLogic(value string) *nautobotapi.CustomFieldFilterLogicValue {
	v := nautobotapi.CustomFieldFilterLogicValue(value)
	return &v
}
