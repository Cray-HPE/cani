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
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

// metaOwner is one object whose metadata a validator inspects: how a report
// names it, the content type its custom fields are defined for, and its two
// metadata maps.
type metaOwner struct {
	label       string
	contentType string
	custom      map[string]any
	provider    map[string]any
}

// eachObjectMeta calls visit for every object that carries metadata: each
// record of each collection, and each interface spec a device or module
// holds. The interface index is left out: it is rebuilt from the specs, so
// its metadata is theirs.
func (inv *Inventory) eachObjectMeta(visit func(metaOwner)) {
	eachMeta(inv.Locations, "location", ContentTypeLocation, visit, func(o *CaniLocationType) (string, *ObjectMeta) { return o.Name, &o.ObjectMeta })
	eachMeta(inv.Racks, "rack", ContentTypeRack, visit, func(o *CaniRackType) (string, *ObjectMeta) { return o.Name, &o.ObjectMeta })
	eachMeta(inv.Devices, "device", ContentTypeDevice, visit, func(o *CaniDeviceType) (string, *ObjectMeta) {
		visitInterfaceSpecs(visit, "device", o.Name, o.Interfaces)
		return o.Name, &o.ObjectMeta
	})
	eachMeta(inv.Modules, "module", ContentTypeModule, visit, func(o *CaniModuleType) (string, *ObjectMeta) {
		visitInterfaceSpecs(visit, "module", o.Label(), o.Interfaces)
		return o.Label(), &o.ObjectMeta
	})
	eachMeta(inv.Cables, "cable", ContentTypeCable, visit, func(o *CaniCableType) (string, *ObjectMeta) { return o.Label, &o.ObjectMeta })
	eachMeta(inv.Frus, "inventory item", ContentTypeInventoryItem, visit, func(o *CaniFruType) (string, *ObjectMeta) { return o.Name, &o.ObjectMeta })
	eachMeta(inv.VLANs, "VLAN", ContentTypeVLAN, visit, func(o *CaniVLAN) (string, *ObjectMeta) { return o.Name, &o.ObjectMeta })
	eachMeta(inv.Prefixes, "prefix", ContentTypePrefix, visit, func(o *CaniPrefix) (string, *ObjectMeta) { return o.Prefix, &o.ObjectMeta })
	eachMeta(inv.IPAddresses, "IP address", ContentTypeIPAddress, visit, func(o *CaniIPAddress) (string, *ObjectMeta) { return o.Address, &o.ObjectMeta })
	eachMeta(inv.VRFs, "VRF", ContentTypeVRF, visit, func(o *CaniVRF) (string, *ObjectMeta) { return o.Name, &o.ObjectMeta })
}

// eachMeta calls visit once per record a collection holds. describe returns
// the record's name and metadata.
func eachMeta[T any](records map[uuid.UUID]*T, kind, contentType string, visit func(metaOwner), describe func(*T) (string, *ObjectMeta)) {
	for id, record := range records {
		if record == nil {
			continue
		}
		name, meta := describe(record)
		visit(metaOwner{
			label:       fmt.Sprintf("%s %q (%s)", kind, name, id),
			contentType: contentType,
			custom:      meta.CustomFields,
			provider:    meta.ProviderMetadata,
		})
	}
}

// visitInterfaceSpecs calls visit for each interface spec an object owns.
func visitInterfaceSpecs(visit func(metaOwner), kind, owner string, specs []InterfaceSpec) {
	for i := range specs {
		visit(metaOwner{
			label:       fmt.Sprintf("interface %q of %s %q", specs[i].Name, kind, owner),
			contentType: ContentTypeInterface,
			custom:      specs[i].CustomFields,
			provider:    specs[i].ProviderMetadata,
		})
	}
}

// customFieldIssue counts the objects that share one problem with one custom
// field, and keeps the first of them by label as the example to report.
type customFieldIssue struct {
	contentType string
	count       int
	example     string
}

// customFieldIssues collects the problems found with stored custom-field
// values, one entry per problem description.
type customFieldIssues struct {
	definitions map[string]CustomFieldDefinition
	found       map[string]*customFieldIssue
}

// validateCustomFields reports every stored custom-field value that does not
// fit the metadata catalog: a key the catalog does not define, one defined
// for other content types only, a value of the wrong type or outside the
// field's choices, and a required field holding null. The command line checks
// a value as it is typed; values an import, a migration or a hand edit wrote
// are checked here. Each problem is reported once per field and content type,
// with how many objects have it and one example, so a widespread problem does
// not bury command output.
func (inv *Inventory) validateCustomFields() *RelationshipResult {
	issues := &customFieldIssues{definitions: make(map[string]CustomFieldDefinition), found: make(map[string]*customFieldIssue)}
	for _, def := range inv.ListCustomFields() {
		issues.definitions[def.Key] = def
	}
	inv.eachObjectMeta(issues.collect)
	result := &RelationshipResult{}
	for _, problem := range slices.Sorted(maps.Keys(issues.found)) {
		issue := issues.found[problem]
		result.Unresolved = append(result.Unresolved, fmt.Sprintf(
			"%s; %d %s object(s) hold it, e.g. %s", problem, issue.count, issue.contentType, issue.example))
	}
	return result
}

// collect records each problem with the custom fields one object holds.
func (c *customFieldIssues) collect(owner metaOwner) {
	for _, key := range sortedKeys(owner.custom) {
		problem, found := customFieldProblem(c.definitions, owner.contentType, key, owner.custom[key])
		if !found {
			continue
		}
		issue := c.found[problem]
		if issue == nil {
			issue = &customFieldIssue{contentType: owner.contentType}
			c.found[problem] = issue
		}
		issue.count++
		if issue.example == "" || owner.label < issue.example {
			issue.example = owner.label
		}
	}
}

// customFieldProblem describes why value, stored for the custom field key of
// an object of contentType, does not fit the catalog; found is false when it
// fits. The description names the field and content type and nothing about
// the object, so one description collects every object with that problem.
func customFieldProblem(definitions map[string]CustomFieldDefinition, contentType, key string, value any) (string, bool) {
	prefix := fmt.Sprintf("custom field %q on %s", key, contentType)
	def, defined := definitions[key]
	switch {
	case !defined:
		return prefix + ` has no definition; define it with "cani alpha add metadata custom-field" or remove the value`, true
	case !slices.Contains(def.ContentTypes, contentType):
		return fmt.Sprintf("%s is defined for %s only", prefix, strings.Join(def.ContentTypes, ", ")), true
	case value == nil && def.Required:
		return prefix + " is required but holds null", true
	case value == nil:
		return "", false
	}
	if problem := def.valueProblem(value); problem != "" {
		return prefix + " " + problem, true
	}
	return "", false
}

// valueProblem describes why a stored value is not of the kind the
// definition's type stores; it is empty when the value fits. A json field
// stores any value.
func (d CustomFieldDefinition) valueProblem(value any) string {
	switch d.Type {
	case customFieldTypeJSON:
		return ""
	case customFieldTypeSelect, customFieldTypeMultiSelect:
		return d.choiceProblem(value)
	}
	if !storesValue(d.Type, value) {
		return "holds a value that is not " + describeFieldType(d.Type)
	}
	return ""
}

// storesValue reports whether value is of the kind a field of fieldType
// stores: a whole number for integer, a bool for boolean, a YYYY-MM-DD string
// for date, and a string for every text-like type.
func storesValue(fieldType string, value any) bool {
	switch fieldType {
	case customFieldTypeInteger:
		return isIntegral(value)
	case customFieldTypeBoolean:
		_, ok := value.(bool)
		return ok
	case customFieldTypeDate:
		text, ok := value.(string)
		return ok && isDate(text)
	default:
		_, ok := value.(string)
		return ok
	}
}

// describeFieldType names the kind of value a field type stores.
func describeFieldType(fieldType string) string {
	switch fieldType {
	case customFieldTypeInteger:
		return "an integer"
	case customFieldTypeBoolean:
		return "a boolean"
	case customFieldTypeDate:
		return "a date (YYYY-MM-DD)"
	default:
		return "text"
	}
}

// choiceProblem describes a stored selection the definition does not
// declare: a select field stores one choice, a multi-select a list of them.
func (d CustomFieldDefinition) choiceProblem(value any) string {
	choices := strings.Join(d.Choices, ", ")
	selected, shaped := selectedChoices(d.Type, value)
	if !shaped && d.Type == customFieldTypeSelect {
		return fmt.Sprintf("holds a value that is not one of its choices (%s)", choices)
	}
	if !shaped {
		return fmt.Sprintf("holds a value that is not a list of its choices (%s)", choices)
	}
	for _, choice := range selected {
		if !slices.Contains(d.Choices, choice) {
			return fmt.Sprintf("holds %q, which is not one of its choices (%s)", choice, choices)
		}
	}
	return ""
}

// selectedChoices returns the choices a stored select or multi-select value
// names, and whether the value has the shape its type stores. A list is
// []string as the command line stores it or []any as JSON decodes it.
func selectedChoices(fieldType string, value any) ([]string, bool) {
	if fieldType == customFieldTypeSelect {
		choice, ok := value.(string)
		return []string{choice}, ok
	}
	switch list := value.(type) {
	case []string:
		return list, true
	case []any:
		choices := make([]string, 0, len(list))
		for _, item := range list {
			choice, ok := item.(string)
			if !ok {
				return nil, false
			}
			choices = append(choices, choice)
		}
		return choices, true
	}
	return nil, false
}

// isIntegral reports whether value is a whole number: the int the command
// line stores, or the float64 a JSON number decodes to when it has no
// fraction.
func isIntegral(value any) bool {
	switch number := value.(type) {
	case int, int64:
		return true
	case float64:
		return number == math.Trunc(number)
	}
	return false
}

// isDate reports whether text reads YYYY-MM-DD.
func isDate(text string) bool {
	_, err := time.Parse(time.DateOnly, text)
	return err == nil
}
