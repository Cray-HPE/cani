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
package export

import (
	"reflect"

	openapi_types "github.com/Cray-HPE/cani/internal/openapi/types"
	"github.com/google/uuid"
)

// oaPtr returns a pointer to an openapi_types.UUID copy of id.
func oaPtr(id uuid.UUID) *openapi_types.UUID {
	u := openapi_types.UUID(id)
	return &u
}

// nbCF wraps a flat custom-fields map in the Nautobot 3.2 pointer-value shape.
func nbCF(cf map[string]interface{}) *map[string]*interface{} {
	return toNautobotCustomFields(cf)
}

// setNBValue sets the string ".Value" of a Nautobot label/value inline struct
// (e.g. Cable.Type, Cable.LengthUnit, Device.Face), allocating the pointer
// field as needed. field is a pointer to the (possibly anonymous) inline-struct
// field, which may itself be a pointer that gets allocated.
func setNBValue(field any, value string) {
	rv := reflect.ValueOf(field)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return
	}
	slot := rv.Elem()
	if slot.Kind() == reflect.Ptr {
		if slot.IsNil() {
			if !slot.CanSet() {
				return
			}
			slot.Set(reflect.New(slot.Type().Elem()))
		}
		slot = slot.Elem()
	}
	vf := slot.FieldByName("Value")
	if !vf.IsValid() || !vf.CanSet() || vf.Kind() != reflect.Ptr {
		return
	}
	ev := reflect.New(vf.Type().Elem())
	ev.Elem().SetString(value)
	vf.Set(ev)
}
