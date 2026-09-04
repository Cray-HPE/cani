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
func setRefID(field any, id uuid.UUID) {
	rv := reflect.ValueOf(field)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return
	}
	target := rv.Elem()
	if target.Kind() == reflect.Ptr {
		if target.IsNil() {
			if !target.CanSet() {
				return
			}
			target.Set(reflect.New(target.Type().Elem()))
		}
		target = target.Elem()
	}
	if target.Kind() != reflect.Struct {
		return
	}
	idField := target.FieldByName("Id")
	if !idField.IsValid() || !idField.CanSet() || idField.Kind() != reflect.Ptr {
		return
	}
	// idField is *<...>_Id; allocate one and load the UUID via its
	// json.Unmarshaler (member 0 is the UUID variant).
	nv := reflect.New(idField.Type().Elem())
	u, ok := nv.Interface().(json.Unmarshaler)
	if !ok {
		return
	}
	b, err := json.Marshal(id)
	if err != nil {
		return
	}
	if err := u.UnmarshalJSON(b); err != nil {
		return
	}
	idField.Set(nv)
}

// newRef returns a new reference value of type T with its Id union set to the
// given UUID. T is the concrete (possibly anonymous) reference struct type of a
// request field, inferred at the call site.
func newRef[T any](id uuid.UUID) T {
	var ref T
	setRefID(&ref, id)
	return ref
}

// newRefPtr returns a pointer to a new reference value of type T with its Id
// union set to the given UUID.
func newRefPtr[T any](id uuid.UUID) *T {
	ref := newRef[T](id)
	return &ref
}

// setRefSlice populates a request's repeated reference field (e.g. Tags,
// TaggedVlans) from a list of UUIDs. field must be a pointer to the field,
// which is typically `*[]struct{ Id *<...>_Id; ... }`; a nil slice pointer is
// allocated. A nil/empty ids list leaves the field untouched.
func setRefSlice(field any, ids []uuid.UUID) {
	if len(ids) == 0 {
		return
	}
	rv := reflect.ValueOf(field)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return
	}
	target := rv.Elem()
	if target.Kind() == reflect.Ptr {
		if target.IsNil() {
			if !target.CanSet() {
				return
			}
			target.Set(reflect.New(target.Type().Elem()))
		}
		target = target.Elem()
	}
	if target.Kind() != reflect.Slice {
		return
	}
	elemType := target.Type().Elem()
	out := reflect.MakeSlice(target.Type(), 0, len(ids))
	for _, id := range ids {
		ev := reflect.New(elemType)
		setRefID(ev.Interface(), id)
		out = reflect.Append(out, ev.Elem())
	}
	target.Set(out)
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
