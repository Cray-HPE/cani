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
	"fmt"
	"strings"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
)

// exportChoiceWeightStep is the weight interval the export recreates choices
// with, in order: 100, 200, … (ensureCustomFieldChoices in the export
// package).
const exportChoiceWeightStep = 100

// defaultFilterLogic is the filter logic Nautobot gives a definition that
// does not set one.
const defaultFilterLogic = "loose"

// unsupportedAttributeWarnings names the attributes of a Nautobot definition
// the catalog does not model and that hold something other than Nautobot's
// default, so a fresh destination would get the definition without them. It
// names the attributes, never their values.
func unsupportedAttributeWarnings(raw *nautobotapi.CustomField, choiceWeights []int) []string {
	attributes := []struct {
		name string
		set  bool
	}{
		{"validation_minimum", raw.ValidationMinimum != nil},
		{"validation_maximum", raw.ValidationMaximum != nil},
		{"validation_regex", strVal(raw.ValidationRegex) != ""},
		{"filter_logic", filterLogicSet(raw)},
		{"grouping", strVal(raw.Grouping) != ""},
		{"advanced_ui", raw.AdvancedUi != nil && *raw.AdvancedUi},
		{"scope_filter", scopeFilterSet(raw.ScopeFilter)},
		{"choice weights", !choiceWeightsReproduced(choiceWeights)},
	}
	var set []string
	for _, attribute := range attributes {
		if attribute.set {
			set = append(set, attribute.name)
		}
	}
	if len(set) == 0 {
		return nil
	}
	return []string{fmt.Sprintf("custom field %q: %s not carried by CANI; a fresh destination gets the definition without them (choices keep their order)",
		strVal(raw.Key), strings.Join(set, ", "))}
}

func filterLogicSet(raw *nautobotapi.CustomField) bool {
	if raw.FilterLogic == nil || raw.FilterLogic.Value == nil {
		return false
	}
	return string(*raw.FilterLogic.Value) != defaultFilterLogic
}

// scopeFilterSet reports whether a scope filter narrows the field; Nautobot
// returns an empty object for none.
func scopeFilterSet(filter interface{}) bool {
	switch f := filter.(type) {
	case nil:
		return false
	case map[string]interface{}:
		return len(f) > 0
	default:
		return true
	}
}

// choiceWeightsReproduced reports whether the export, which recreates the
// choices in order with weights 100, 200, …, writes the weights Nautobot
// holds.
func choiceWeightsReproduced(weights []int) bool {
	for i, weight := range weights {
		if weight != (i+1)*exportChoiceWeightStep {
			return false
		}
	}
	return true
}
