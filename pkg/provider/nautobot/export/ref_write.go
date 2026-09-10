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
	target, err := writableTarget(field, "reference field")
	if err != nil {
		return err
	}
	if target.Kind() != reflect.Struct {
		return fmt.Errorf("reference field must point to a struct, got %s", target.Kind())
	}
	idField, err := writablePointerField(target, "Id")
	if err != nil {
		return fmt.Errorf("reference struct %s: %w", target.Type(), err)
	}
	idValue, err := newReferenceID(idField.Type().Elem(), id)
	if err != nil {
		return err
	}
	idField.Set(idValue)
	return nil
}

func newReferenceID(idType reflect.Type, id uuid.UUID) (reflect.Value, error) {
	value := reflect.New(idType)
	unmarshaler, ok := value.Interface().(json.Unmarshaler)
	if !ok {
		return reflect.Value{}, fmt.Errorf("reference Id type %s does not implement json.Unmarshaler", idType)
	}
	payload, err := json.Marshal(id)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("marshal reference UUID: %w", err)
	}
	if err := unmarshaler.UnmarshalJSON(payload); err != nil {
		return reflect.Value{}, fmt.Errorf("set reference UUID on %s: %w", idType, err)
	}
	return value, nil
}

// setRefSlice populates a request's repeated reference field (e.g. Tags,
// TaggedVlans) from a list of UUIDs. field must be a pointer to the field,
// which is typically `*[]struct{ Id *<...>_Id; ... }`; a nil slice pointer is
// allocated. A nil/empty ids list leaves the field untouched.
func setRefSlice(field any, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	target, err := writableTarget(field, "reference slice")
	if err != nil {
		return err
	}
	if target.Kind() != reflect.Slice {
		return fmt.Errorf("reference slice field must point to a slice, got %s", target.Kind())
	}
	values := reflect.MakeSlice(target.Type(), 0, len(ids))
	for _, id := range ids {
		item := reflect.New(target.Type().Elem())
		if err := setRefID(item.Interface(), id); err != nil {
			return fmt.Errorf("set reference slice item %s: %w", id, err)
		}
		values = reflect.Append(values, item.Elem())
	}
	target.Set(values)
	return nil
}

func writableTarget(field any, label string) (reflect.Value, error) {
	value := reflect.ValueOf(field)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return reflect.Value{}, fmt.Errorf("%s must be a non-nil pointer, got %T", label, field)
	}
	target := value.Elem()
	if target.Kind() != reflect.Ptr {
		return target, nil
	}
	if target.IsNil() {
		if !target.CanSet() {
			return reflect.Value{}, fmt.Errorf("%s pointer %T cannot be set", label, field)
		}
		target.Set(reflect.New(target.Type().Elem()))
	}
	return target.Elem(), nil
}

func writablePointerField(target reflect.Value, name string) (reflect.Value, error) {
	field := target.FieldByName(name)
	if !field.IsValid() || !field.CanSet() || field.Kind() != reflect.Ptr {
		return reflect.Value{}, fmt.Errorf("no settable pointer %s field", name)
	}
	return field, nil
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
	target, err := writableTarget(field, "device face")
	if err != nil || target.Kind() != reflect.Struct {
		return
	}
	value := reflect.New(target.Type())
	method := value.MethodByName("FromFaceEnum")
	if !method.IsValid() {
		return
	}
	method.Call([]reflect.Value{reflect.ValueOf(fe)})
	reflect.ValueOf(field).Elem().Set(value)
}
