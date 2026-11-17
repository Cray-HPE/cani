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

// Package customfields reads the --custom-field and --clear-custom-field
// flags, which `add` and `update` hand to every subcommand, and applies them
// to an object's custom fields.
package customfields

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

const (
	// FlagSet names the repeatable key=value flag that sets a custom field.
	FlagSet = "custom-field"
	// FlagClear names the repeatable flag that clears one.
	FlagClear = "clear-custom-field"
)

// RegisterSet adds --custom-field to flags.
func RegisterSet(flags *cli.FlagSet) {
	flags.StringToString(FlagSet, nil,
		"Custom field to set, as key=value (repeatable). The key must be defined for the item's type; see 'add metadata custom-field'")
}

// RegisterClear adds --clear-custom-field to flags.
func RegisterClear(flags *cli.FlagSet) {
	flags.StringArray(FlagClear, nil, "Custom field to clear, by key (repeatable)")
}

// Edit is the change one command line asks for to the custom fields of
// objects of one content type: the values it sets, parsed into the type each
// field's definition declares, and the keys it clears.
type Edit struct {
	inv         *devicetypes.Inventory
	contentType string
	set         map[string]any
	clear       []string
}

// FromFlags reads the edit from the flags cmd has. A key with no definition
// for contentType, a value its definition does not accept, and a key that is
// both set and cleared are errors that name the flag.
func FromFlags(cmd *cli.Command, inv *devicetypes.Inventory, contentType string) (Edit, error) {
	typed, _ := cmd.Flags().GetStringToString(FlagSet)
	var parsed devicetypes.ObjectMeta
	for _, key := range slices.Sorted(maps.Keys(typed)) {
		if err := inv.SetCustomField(&parsed, contentType, key, typed[key]); err != nil {
			return Edit{}, setError(err)
		}
	}
	// A command without --clear-custom-field clears nothing.
	cleared, _ := cmd.Flags().GetStringArray(FlagClear)
	for _, key := range cleared {
		if _, set := parsed.CustomFields[key]; set {
			return Edit{}, fmt.Errorf("--%s and --%s both name %q", FlagSet, FlagClear, key)
		}
	}
	return Edit{inv: inv, contentType: contentType, set: parsed.CustomFields, clear: cleared}, nil
}

// setError names the flag in a refusal, and says how a field with no
// definition gets one.
func setError(err error) error {
	if errors.Is(err, devicetypes.ErrCustomFieldNotDefined) {
		return fmt.Errorf("--%s: %w; define it with 'add metadata custom-field' or import its definition", FlagSet, err)
	}
	return fmt.Errorf("--%s: %w", FlagSet, err)
}

// Stamp writes the values the edit sets onto a new object. The object gets a
// custom-field map and values of its own, so items built from one command
// line, or copied from one hardware type, share none of them.
func (e Edit) Stamp(meta *devicetypes.ObjectMeta) {
	if len(e.set) == 0 {
		return
	}
	fields := maps.Clone(meta.CustomFields)
	if fields == nil {
		fields = make(map[string]any, len(e.set))
	}
	for key, value := range e.set {
		fields[key] = cloneValue(value)
	}
	meta.CustomFields = fields
}

// Apply makes the edit to an existing object: it sets, then clears. A
// required field, and a key the object neither holds nor has a definition
// for, cannot be cleared; meta is then left as it was.
func (e Edit) Apply(meta *devicetypes.ObjectMeta) error {
	edited := devicetypes.ObjectMeta{CustomFields: maps.Clone(meta.CustomFields)}
	e.Stamp(&edited)
	for _, key := range e.clear {
		if err := e.inv.ClearCustomField(&edited, e.contentType, key); err != nil {
			return fmt.Errorf("--%s: %w", FlagClear, err)
		}
	}
	meta.CustomFields = edited.CustomFields
	return nil
}

// Apply reads the flags of cmd and makes the edit they ask for to meta, an
// object of contentType.
func Apply(cmd *cli.Command, inv *devicetypes.Inventory, meta *devicetypes.ObjectMeta, contentType string) error {
	edit, err := FromFlags(cmd, inv, contentType)
	if err != nil {
		return err
	}
	return edit.Apply(meta)
}

// cloneValue copies a parsed value: a list of choices, or whatever a JSON
// field decoded to.
func cloneValue(value any) any {
	switch typed := value.(type) {
	case []string:
		return slices.Clone(typed)
	case []any:
		items := make([]any, len(typed))
		for i, item := range typed {
			items[i] = cloneValue(item)
		}
		return items
	case map[string]any:
		fields := make(map[string]any, len(typed))
		for key, item := range typed {
			fields[key] = cloneValue(item)
		}
		return fields
	default:
		return value
	}
}
