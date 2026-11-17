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
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Content types of the inventory's object kinds. A custom-field definition
// lists the content types whose objects may carry a value for it.
const (
	ContentTypeDevice        = "dcim.device"
	ContentTypeRack          = "dcim.rack"
	ContentTypeLocation      = "dcim.location"
	ContentTypeModule        = "dcim.module"
	ContentTypeInterface     = "dcim.interface"
	ContentTypeInventoryItem = "dcim.inventoryitem"
	ContentTypeCable         = "dcim.cable"
	ContentTypeVLAN          = "ipam.vlan"
	ContentTypePrefix        = "ipam.prefix"
	ContentTypeIPAddress     = "ipam.ipaddress"
	ContentTypeVRF           = "ipam.vrf"
)

// Custom-field types whose values are not stored as the text that was typed.
const (
	customFieldTypeInteger     = "integer"
	customFieldTypeBoolean     = "boolean"
	customFieldTypeDate        = "date"
	customFieldTypeJSON        = "json"
	customFieldTypeSelect      = "select"
	customFieldTypeMultiSelect = "multi-select"
)

// CustomFieldFor returns the definition of the custom field key when the
// metadata catalog defines it for objects of contentType.
func (inv *Inventory) CustomFieldFor(contentType, key string) (CustomFieldDefinition, error) {
	for _, def := range inv.ListCustomFields() {
		if def.Key != key {
			continue
		}
		if !slices.Contains(def.ContentTypes, contentType) {
			return CustomFieldDefinition{}, fmt.Errorf("custom field %q is not defined for %s (it applies to %s)",
				key, contentType, strings.Join(def.ContentTypes, ", "))
		}
		return def, nil
	}
	return CustomFieldDefinition{}, fmt.Errorf("custom field %q is not defined in the inventory", key)
}

// SetCustomField parses raw, as it would be typed on a command line, into the
// type the field's definition declares and stores it on meta. A key with no
// definition for contentType, a value of the wrong type and a value outside
// a select field's choices are errors, and meta is left as it was.
func (inv *Inventory) SetCustomField(meta *ObjectMeta, contentType, key, raw string) error {
	def, err := inv.CustomFieldFor(contentType, key)
	if err != nil {
		return err
	}
	value, err := def.ParseValue(raw)
	if err != nil {
		return err
	}
	if meta.CustomFields == nil {
		meta.CustomFields = make(map[string]any)
	}
	meta.CustomFields[key] = value
	return nil
}

// ClearCustomField removes the value of the custom field key from meta. A
// field defined for contentType is set to an explicit null rather than
// deleted, because a system that holds the object keeps the value of a key a
// write leaves out; a required field cannot be cleared. A key that has no
// definition but is held by meta is deleted, so a stray value can be
// removed. A key that is neither defined nor held is an error.
func (inv *Inventory) ClearCustomField(meta *ObjectMeta, contentType, key string) error {
	def, err := inv.CustomFieldFor(contentType, key)
	if err != nil {
		if _, held := meta.CustomFields[key]; !held {
			return err
		}
		delete(meta.CustomFields, key)
		return nil
	}
	if def.Required {
		return fmt.Errorf("custom field %q is required and cannot be cleared", key)
	}
	if meta.CustomFields == nil {
		meta.CustomFields = make(map[string]any)
	}
	meta.CustomFields[key] = nil
	return nil
}

// ParseValue converts raw into the value the definition's type stores: an
// int for integer, a bool for boolean, the decoded value for json, a list of
// choices for multi-select, and the text itself for every other type. A date
// must read YYYY-MM-DD and a selection must be one of the declared choices.
func (d CustomFieldDefinition) ParseValue(raw string) (any, error) {
	switch d.Type {
	case customFieldTypeInteger:
		return d.parseInteger(raw)
	case customFieldTypeBoolean:
		return d.parseBoolean(raw)
	case customFieldTypeDate:
		return d.parseDate(raw)
	case customFieldTypeJSON:
		return d.parseJSON(raw)
	case customFieldTypeSelect:
		return raw, d.checkChoices(raw)
	case customFieldTypeMultiSelect:
		return d.parseChoices(raw)
	default:
		return raw, nil
	}
}

func (d CustomFieldDefinition) parseInteger(raw string) (any, error) {
	number, err := strconv.Atoi(raw)
	if err != nil {
		return nil, fmt.Errorf("custom field %q: %q is not an integer", d.Key, raw)
	}
	return number, nil
}

func (d CustomFieldDefinition) parseBoolean(raw string) (any, error) {
	truth, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, fmt.Errorf("custom field %q: %q is not a boolean (use true or false)", d.Key, raw)
	}
	return truth, nil
}

func (d CustomFieldDefinition) parseDate(raw string) (any, error) {
	if _, err := time.Parse(time.DateOnly, raw); err != nil {
		return nil, fmt.Errorf("custom field %q: %q is not a date (use YYYY-MM-DD)", d.Key, raw)
	}
	return raw, nil
}

func (d CustomFieldDefinition) parseJSON(raw string) (any, error) {
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return nil, fmt.Errorf("custom field %q: value is not valid JSON: %w", d.Key, err)
	}
	return value, nil
}

// parseChoices splits a comma-separated selection; empty text selects nothing.
func (d CustomFieldDefinition) parseChoices(raw string) (any, error) {
	selected := []string{}
	for _, choice := range strings.Split(raw, ",") {
		if choice = strings.TrimSpace(choice); choice != "" {
			selected = append(selected, choice)
		}
	}
	return selected, d.checkChoices(selected...)
}

// checkChoices reports the first value the definition does not declare.
func (d CustomFieldDefinition) checkChoices(values ...string) error {
	for _, value := range values {
		if !slices.Contains(d.Choices, value) {
			return fmt.Errorf("custom field %q: %q is not one of its choices (%s)", d.Key, value, strings.Join(d.Choices, ", "))
		}
	}
	return nil
}
