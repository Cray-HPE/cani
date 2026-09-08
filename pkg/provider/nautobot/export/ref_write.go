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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// Nautobot 3.2 replaced the single shared reference request type with a
// distinct inline struct per foreign-key field, each shaped
// `struct{ Id *<Type>_<Field>_Id; ObjectType *string; Url *string }` where the
// Id is a generated oneOf(UUID|int) union. A single constructor can no longer
// build every reference, so setRefID populates any such field generically.

// setRefID sets the Id union of a Nautobot request reference field to a UUID.
// field must be a pointer to the reference field, which may be a value struct
// or a (possibly nil) pointer to struct; nil pointers are allocated.
func setRefID(field any, id uuid.UUID) error {
	rv := reflect.ValueOf(field)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("reference field must be a non-nil pointer, got %T", field)
	}
	target := rv.Elem()
	if target.Kind() == reflect.Ptr {
		if target.IsNil() {
			if !target.CanSet() {
				return fmt.Errorf("reference pointer %T cannot be set", field)
			}
			target.Set(reflect.New(target.Type().Elem()))
		}
		target = target.Elem()
	}
	if target.Kind() != reflect.Struct {
		return fmt.Errorf("reference field must point to a struct, got %s", target.Kind())
	}
	idField := target.FieldByName("Id")
	if !idField.IsValid() || !idField.CanSet() || idField.Kind() != reflect.Ptr {
		return fmt.Errorf("reference struct %s has no settable pointer Id field", target.Type())
	}
	// idField is *<...>_Id; allocate one and load the UUID via its
	// json.Unmarshaler (member 0 is the UUID variant).
	nv := reflect.New(idField.Type().Elem())
	u, ok := nv.Interface().(json.Unmarshaler)
	if !ok {
		return fmt.Errorf("reference Id type %s does not implement json.Unmarshaler", idField.Type().Elem())
	}
	b, err := json.Marshal(id)
	if err != nil {
		return fmt.Errorf("marshal reference UUID: %w", err)
	}
	if err := u.UnmarshalJSON(b); err != nil {
		return fmt.Errorf("set reference UUID on %s: %w", target.Type(), err)
	}
	idField.Set(nv)
	return nil
}

// setRefSlice populates a request's repeated reference field (e.g. Tags,
// TaggedVlans) from a list of UUIDs. field must be a pointer to the field,
// which is typically `*[]struct{ Id *<...>_Id; ... }`; a nil slice pointer is
// allocated. A nil/empty ids list leaves the field untouched.
func setRefSlice(field any, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	rv := reflect.ValueOf(field)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("reference slice must be a non-nil pointer, got %T", field)
	}
	target := rv.Elem()
	if target.Kind() == reflect.Ptr {
		if target.IsNil() {
			if !target.CanSet() {
				return fmt.Errorf("reference slice pointer %T cannot be set", field)
			}
			target.Set(reflect.New(target.Type().Elem()))
		}
		target = target.Elem()
	}
	if target.Kind() != reflect.Slice {
		return fmt.Errorf("reference slice field must point to a slice, got %s", target.Kind())
	}
	elemType := target.Type().Elem()
	out := reflect.MakeSlice(target.Type(), 0, len(ids))
	for _, id := range ids {
		ev := reflect.New(elemType)
		if err := setRefID(ev.Interface(), id); err != nil {
			return fmt.Errorf("set reference slice item %s: %w", id, err)
		}
		out = reflect.Append(out, ev.Elem())
	}
	target.Set(out)
	return nil
}

// setDeviceFace sets a device request's Face union field ("front"/"rear",
// defaulting to front) generically across the Writable/Bulk/Patched variants,
// each of which has a distinct *_Face union type implementing FromFaceEnum.
// field must be a pointer to the Face field (a *<...>_Face).
func setDeviceFace(field any, face string) {
	fe := nautobotapi.FaceEnumFront
	if strings.EqualFold(face, "rear") {
		fe = nautobotapi.FaceEnumRear
	}
	rv := reflect.ValueOf(field)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return
	}
	slot := rv.Elem() // the *_Face pointer field
	if slot.Kind() != reflect.Ptr || !slot.CanSet() {
		return
	}
	nv := reflect.New(slot.Type().Elem())
	m := nv.MethodByName("FromFaceEnum")
	if !m.IsValid() {
		return
	}
	m.Call([]reflect.Value{reflect.ValueOf(fe)})
	slot.Set(nv)
}
