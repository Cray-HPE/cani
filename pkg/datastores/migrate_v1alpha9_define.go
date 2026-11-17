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
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// Custom-field types the migration infers from the values it moved. The
// names are the ones a definition declares.
const (
	fieldTypeText    = "text"
	fieldTypeInteger = "integer"
	fieldTypeBoolean = "boolean"
	fieldTypeJSON    = "json"
)

// migratedFieldDescription is given to each definition the migration creates.
const migratedFieldDescription = "Defined by the migration to v1alpha10 for values moved out of provider metadata"

// movedField is one value the fold moved into an object's custom fields.
type movedField struct {
	ref metadataRef
	key string
}

// defineMigratedFields gives every custom field the fold created a
// definition, so that the values it moved are checked and exported like any
// other custom field. A key the catalog does not define is defined for the
// content types of the objects that hold it, with the type the values share:
// boolean, integer, text, or json when they are of several kinds or are maps
// or lists. A key the catalog defines has those content types added to its
// definition, and a moved text value is parsed as the definition's type, as
// it would be if typed; a value the definition does not accept is left as it
// is and reported. Only keys the fold moved are defined, and only where the
// value is still held: a custom field that had no definition before the
// migration stays as it was.
func (f *metadataFold) defineMigratedFields() {
	byKey := make(map[string][]movedField)
	for _, moved := range f.moved {
		if _, held := (*moved.ref.custom)[moved.key]; held {
			byKey[moved.key] = append(byKey[moved.key], moved)
		}
	}
	for _, key := range slices.Sorted(maps.Keys(byKey)) {
		if def, defined := f.definition(key); defined {
			f.extendDefinition(def, byKey[key])
		} else {
			f.addDefinition(key, byKey[key])
		}
	}
}

// definition returns the catalog's definition of key, when it has one, to
// be changed in place.
func (f *metadataFold) definition(key string) (*devicetypes.CustomFieldDefinition, bool) {
	defs := f.inv.ListCustomFields()
	for i := range defs {
		if defs[i].Key == key {
			return &defs[i], true
		}
	}
	return nil, false
}

// addDefinition defines key for the content types of the objects that hold
// it, with the type their values share.
func (f *metadataFold) addDefinition(key string, held []movedField) {
	values := make([]any, 0, len(held))
	for _, moved := range held {
		values = append(values, (*moved.ref.custom)[moved.key])
	}
	def := devicetypes.CustomFieldDefinition{
		Key:          key,
		Label:        key,
		Type:         inferFieldType(values),
		ContentTypes: contentTypesOf(held),
		Description:  migratedFieldDescription,
	}
	if err := f.inv.AddCustomField(def); err != nil {
		f.report = append(f.report, fmt.Sprintf("custom field %q could not be defined: %v", key, err))
		return
	}
	f.report = append(f.report, fmt.Sprintf("custom field %q is now defined for %s as a %s field, from the values moved into it",
		key, strings.Join(def.ContentTypes, ", "), def.Type))
}

// extendDefinition adds the content types that hold key to its definition
// and parses each moved text value as the definition's type.
func (f *metadataFold) extendDefinition(def *devicetypes.CustomFieldDefinition, held []movedField) {
	var added []string
	for _, moved := range held {
		if !slices.Contains(def.ContentTypes, moved.ref.contentType) {
			added = append(added, moved.ref.contentType)
		}
		f.parseMoved(*def, moved)
	}
	if len(added) == 0 {
		return
	}
	slices.Sort(added)
	added = slices.Compact(added)
	def.ContentTypes = append(def.ContentTypes, added...)
	f.report = append(f.report, fmt.Sprintf("custom field %q is now also defined for %s, which hold values moved into it",
		def.Key, strings.Join(added, ", ")))
}

// parseMoved stores a moved text value as the definition's type, as it
// would be if typed. A value the definition does not accept is left as it
// is and reported, and the check on load points it out from then on.
func (f *metadataFold) parseMoved(def devicetypes.CustomFieldDefinition, moved movedField) {
	custom := *moved.ref.custom
	text, isText := custom[moved.key].(string)
	if !isText {
		return
	}
	value, err := def.ParseValue(text)
	if err != nil {
		f.note(moved.ref.owner, "custom field %q holds a value its definition does not accept; it was left as it is", moved.key)
		return
	}
	custom[moved.key] = value
}

// contentTypesOf lists the content types of the objects holding a field,
// sorted, each once.
func contentTypesOf(held []movedField) []string {
	types := make([]string, 0, len(held))
	for _, moved := range held {
		types = append(types, moved.ref.contentType)
	}
	slices.Sort(types)
	return slices.Compact(types)
}

// inferFieldType names the type the values share. Values of several kinds,
// and maps and lists, are stored by a json field alone. Null says nothing
// about the type, and values that are all null are text.
func inferFieldType(values []any) string {
	inferred := ""
	for _, value := range values {
		kind := fieldTypeOf(value)
		switch {
		case kind == "":
			continue
		case inferred == "" || inferred == kind:
			inferred = kind
		default:
			return fieldTypeJSON
		}
	}
	if inferred == "" {
		return fieldTypeText
	}
	return inferred
}

// fieldTypeOf names the field type that stores value, or nothing for null.
// A number is an integer when it is whole, as JSON decodes one.
func fieldTypeOf(value any) string {
	switch number := value.(type) {
	case nil:
		return ""
	case bool:
		return fieldTypeBoolean
	case string:
		return fieldTypeText
	case int, int64:
		return fieldTypeInteger
	case float64:
		if number == math.Trunc(number) {
			return fieldTypeInteger
		}
		return fieldTypeJSON
	default:
		return fieldTypeJSON
	}
}
