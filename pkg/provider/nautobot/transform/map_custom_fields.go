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
	"sort"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// MapCustomFields converts Nautobot custom-field definitions into catalog
// entries so a later export can recreate them before it sends values. Only
// the content types CANI exports are kept, because the catalog validates
// against that list; a definition that applies to none of them is left out,
// since no CANI object could carry its value, and both cases are reported.
func MapCustomFields(raws []nautobotapi.CustomField, choices []nautobotapi.CustomFieldChoice) ([]devicetypes.CustomFieldDefinition, []string) {
	known := make(map[string]bool, len(devicetypes.ValidContentTypes))
	for _, ct := range devicetypes.ValidContentTypes {
		known[ct] = true
	}
	choicesByField := groupChoicesByField(choices)

	var defs []devicetypes.CustomFieldDefinition
	var warnings []string
	for i := range raws {
		fieldChoices := choicesByField[directUUID(raws[i].Id)]
		def, dropped := mapCustomField(&raws[i], fieldChoices.values, known)
		if len(def.ContentTypes) == 0 {
			warnings = append(warnings, fmt.Sprintf("custom field %q: applies only to %v, which CANI does not export; definition not imported", def.Key, dropped))
			continue
		}
		if len(dropped) > 0 {
			warnings = append(warnings, fmt.Sprintf("custom field %q: content type(s) %v are not exported by CANI and were left out of its definition", def.Key, dropped))
		}
		warnings = append(warnings, unsupportedAttributeWarnings(&raws[i], fieldChoices.weights)...)
		defs = append(defs, def)
	}
	sort.Slice(defs, func(i, j int) bool { return defs[i].Key < defs[j].Key })
	return defs, warnings
}

// mapCustomField converts one definition, returning it and the content types
// it lost to the CANI vocabulary.
func mapCustomField(raw *nautobotapi.CustomField, choices []string, known map[string]bool) (devicetypes.CustomFieldDefinition, []string) {
	def := devicetypes.CustomFieldDefinition{
		Key:         strVal(raw.Key),
		Label:       raw.Label,
		Description: strVal(raw.Description),
		Default:     raw.Default,
		Choices:     choices,
	}
	if raw.Type.Value != nil {
		def.Type = string(*raw.Type.Value)
	}
	if raw.Required != nil {
		def.Required = *raw.Required
	}
	if raw.Weight != nil {
		def.Weight = *raw.Weight
	}
	var dropped []string
	for _, ct := range raw.ContentTypes {
		if known[ct] {
			def.ContentTypes = append(def.ContentTypes, ct)
		} else {
			dropped = append(dropped, ct)
		}
	}
	sort.Strings(def.ContentTypes)
	return def, dropped
}

// choiceList is one field's choice values in Nautobot's weight order, with
// the weights Nautobot holds for them.
type choiceList struct {
	values  []string
	weights []int
}

// groupChoicesByField indexes choices by the custom field they belong to, in
// Nautobot's weight order so the export recreates them in sequence.
func groupChoicesByField(choices []nautobotapi.CustomFieldChoice) map[uuid.UUID]choiceList {
	sorted := append([]nautobotapi.CustomFieldChoice(nil), choices...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return intVal(sorted[i].Weight) < intVal(sorted[j].Weight)
	})
	grouped := make(map[uuid.UUID]choiceList)
	for i := range sorted {
		fieldID := refID(sorted[i].CustomField)
		list := grouped[fieldID]
		list.values = append(list.values, sorted[i].Value)
		list.weights = append(list.weights, intVal(sorted[i].Weight))
		grouped[fieldID] = list
	}
	return grouped
}
