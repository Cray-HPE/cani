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

// refs.go is hand-written: oapi-codegen only rewrites nautobot_api.go.
package nautobot

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/google/uuid"
)

// RefUUIDChecked extracts a UUID from any generated per-field reference union
// type (e.g. Device_Status_Id, Cable_Status_Id). Nautobot 3.2 emits a distinct
// oneOf(UUID|int) union per reference field; every such type implements
// json.Marshaler with the UUID as member 0, so marshaling then parsing keeps
// the read path type-agnostic. A nil or integer-valued union yields uuid.Nil
// without error; malformed union output is returned as an error.
func RefUUIDChecked(m json.Marshaler) (uuid.UUID, error) {
	if isNilMarshaler(m) {
		return uuid.Nil, nil
	}
	b, err := m.MarshalJSON()
	if err != nil {
		return uuid.Nil, fmt.Errorf("marshal reference ID %T: %w", m, err)
	}
	if len(b) == 0 {
		return uuid.Nil, fmt.Errorf("marshal reference ID %T: empty JSON", m)
	}
	var value any
	if err := json.Unmarshal(b, &value); err != nil {
		return uuid.Nil, fmt.Errorf("decode reference ID %T: %w", m, err)
	}
	return parseRefValue(m, value)
}

// isNilMarshaler also catches a typed-nil pointer (e.g. a nil *Device_Status_Id)
// whose value-receiver MarshalJSON would panic on dereference.
func isNilMarshaler(m json.Marshaler) bool {
	if m == nil {
		return true
	}
	rv := reflect.ValueOf(m)
	return rv.Kind() == reflect.Pointer && rv.IsNil()
}

// parseRefValue reads the decoded union member: a UUID string, or nil/number
// for the absent and integer-id variants.
func parseRefValue(m json.Marshaler, value any) (uuid.UUID, error) {
	switch v := value.(type) {
	case nil, float64:
		return uuid.Nil, nil
	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			return uuid.Nil, fmt.Errorf("decode reference ID %T: %w", m, err)
		}
		return id, nil
	default:
		return uuid.Nil, fmt.Errorf("decode reference ID %T: unexpected JSON type %T", m, value)
	}
}

// SetRefUUID sets any generated per-field reference union type to a UUID by
// unmarshaling the JSON-encoded UUID into the union's member 0. The concrete
// per-field union type is named once at the call site's declaration.
func SetRefUUID(u json.Unmarshaler, id uuid.UUID) error {
	b, err := json.Marshal(id)
	if err != nil {
		return err
	}
	return u.UnmarshalJSON(b)
}
