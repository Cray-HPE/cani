// Copyright 2019 DeepMap, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated from github.com/oapi-codegen/runtime; DO NOT EDIT.
package runtime

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Cray-HPE/cani/internal/openapi/types"
)

// Parameter escaping works differently based on where a header is found

type ParamLocation int

const (
	ParamLocationUndefined ParamLocation = iota
	ParamLocationQuery
	ParamLocationPath
	ParamLocationHeader
	ParamLocationCookie
)

// StyleParam is used by older generated code, and must remain compatible
// with that code. It is not to be used in new templates. Please see the
// function below, which can specialize its output based on the location of
// the parameter.
func StyleParam(style string, explode bool, paramName string, value interface{}) (string, error) {
	return StyleParamWithLocation(style, explode, paramName, ParamLocationUndefined, value)
}

// StyleParamWithLocation serializes a Go value into an OpenAPI-styled parameter
// string, performing escaping based on parameter location.
func StyleParamWithLocation(style string, explode bool, paramName string, paramLocation ParamLocation, value interface{}) (string, error) {
	return StyleParamWithOptions(style, explode, paramName, value, StyleParamOptions{
		ParamLocation: paramLocation,
	})
}

// StyleParamOptions defines optional arguments for StyleParamWithOptions.
type StyleParamOptions struct {
	// ParamLocation controls URL escaping behavior.
	ParamLocation ParamLocation
	// Type is the OpenAPI type of the parameter (e.g. "string", "array").
	Type string
	// Format is the OpenAPI format of the parameter (e.g. "byte", "date-time").
	Format string
	// Required indicates whether the parameter is required.
	Required bool
	// AllowReserved preserves RFC 3986 reserved characters in query values.
	AllowReserved bool
}

// StyleParamWithOptions serializes a Go value with additional OpenAPI options.
func StyleParamWithOptions(style string, explode bool, paramName string, value interface{}, opts StyleParamOptions) (string, error) {
	t := reflect.TypeOf(value)
	v := reflect.ValueOf(value)

	// Things may be passed in by pointer, we need to dereference, so return
	// error on nil.
	if t.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "", fmt.Errorf("value is a nil pointer")
		}
		v = reflect.Indirect(v)
		t = v.Type()
	}

	// If the value implements encoding.TextMarshaler we use it for marshaling
	// https://github.com/deepmap/oapi-codegen/issues/504
	if tu, ok := value.(encoding.TextMarshaler); ok {
		t := reflect.Indirect(reflect.ValueOf(value)).Type()
		convertableToTime := t.ConvertibleTo(reflect.TypeOf(time.Time{}))
		convertableToDate := t.ConvertibleTo(reflect.TypeOf(types.Date{}))

		// Since both time.Time and types.Date implement encoding.TextMarshaler
		// we should avoid calling theirs MarshalText()
		if !convertableToTime && !convertableToDate {
			b, err := tu.MarshalText()
			if err != nil {
				return "", fmt.Errorf("error marshaling '%s' as text: %w", value, err)
			}

			return stylePrimitive(style, explode, paramName, opts.ParamLocation, opts.AllowReserved, string(b))
		}
	}

	switch t.Kind() {
	case reflect.Slice:
		if opts.Format == "byte" && isByteSlice(t) {
			encoded := base64.StdEncoding.EncodeToString(v.Bytes())
			return stylePrimitive(style, explode, paramName, opts.ParamLocation, opts.AllowReserved, encoded)
		}
		n := v.Len()
		sliceVal := make([]interface{}, n)
		for i := 0; i < n; i++ {
			sliceVal[i] = v.Index(i).Interface()
		}
		return styleSlice(style, explode, paramName, opts.ParamLocation, opts.AllowReserved, sliceVal)
	case reflect.Struct:
		return styleStruct(style, explode, paramName, opts.ParamLocation, opts.AllowReserved, value)
	case reflect.Map:
		return styleMap(style, explode, paramName, opts.ParamLocation, opts.AllowReserved, value)
	default:
		return stylePrimitive(style, explode, paramName, opts.ParamLocation, opts.AllowReserved, value)
	}
}

func styleSlice(style string, explode bool, paramName string, paramLocation ParamLocation, allowReserved bool, values []interface{}) (string, error) {
	if style == "deepObject" {
		if !explode {
			return "", errors.New("deepObjects must be exploded")
		}
		return MarshalDeepObject(values, paramName)
	}

	var prefix string
	var separator string
	escapedName := escapeParameterName(paramName, paramLocation)

	switch style {
	case "simple":
		separator = ","
	case "label":
		prefix = "."
		if explode {
			separator = "."
		} else {
			separator = ","
		}
	case "matrix":
		prefix = fmt.Sprintf(";%s=", escapedName)
		if explode {
			separator = prefix
		} else {
			separator = ","
		}
	case "form":
		prefix = fmt.Sprintf("%s=", escapedName)
		if explode {
			separator = "&" + prefix
		} else {
			separator = ","
		}
	case "spaceDelimited":
		prefix = fmt.Sprintf("%s=", escapedName)
		if explode {
			separator = "&" + prefix
		} else {
			separator = " "
		}
	case "pipeDelimited":
		prefix = fmt.Sprintf("%s=", escapedName)
		if explode {
			separator = "&" + prefix
		} else {
			separator = "|"
		}
	default:
		return "", fmt.Errorf("unsupported style '%s'", style)
	}

	// We're going to assume here that the array is one of simple types.
	var err error
	var part string
	parts := make([]string, len(values))
	for i, v := range values {
		part, err = primitiveToString(v)
		part = escapeParameterString(part, paramLocation, allowReserved)
		parts[i] = part
		if err != nil {
			return "", fmt.Errorf("error formatting '%s': %w", paramName, err)
		}
	}
	return prefix + strings.Join(parts, separator), nil
}

// These are special cases. The value may be a date, time, or uuid,
// in which case, marshal it into the correct format.
func marshalKnownTypes(value interface{}) (string, bool) {
	v := reflect.Indirect(reflect.ValueOf(value))
	t := v.Type()

	if t.ConvertibleTo(reflect.TypeOf(time.Time{})) {
		tt := v.Convert(reflect.TypeOf(time.Time{}))
		timeVal := tt.Interface().(time.Time)
		return timeVal.Format(time.RFC3339Nano), true
	}

	if t.ConvertibleTo(reflect.TypeOf(types.Date{})) {
		d := v.Convert(reflect.TypeOf(types.Date{}))
		dateVal := d.Interface().(types.Date)
		return dateVal.Format(types.DateFormat), true
	}

	if t.ConvertibleTo(reflect.TypeOf(types.UUID{})) {
		u := v.Convert(reflect.TypeOf(types.UUID{}))
		uuidVal := u.Interface().(types.UUID)
		return uuidVal.String(), true
	}

	return "", false
}

func styleStruct(style string, explode bool, paramName string, paramLocation ParamLocation, allowReserved bool, value interface{}) (string, error) {
	if timeVal, ok := marshalKnownTypes(value); ok {
		styledVal, err := stylePrimitive(style, explode, paramName, paramLocation, allowReserved, timeVal)
		if err != nil {
			return "", fmt.Errorf("failed to style time: %w", err)
		}
		return styledVal, nil
	}

	if style == "deepObject" {
		if !explode {
			return "", errors.New("deepObjects must be exploded")
		}
		return MarshalDeepObject(value, paramName)
	}

	// If input has Marshaler, such as object has Additional Property or AnyOf,
	// We use this Marshaler and convert into interface{} before styling.
	if m, ok := value.(json.Marshaler); ok {
		buf, err := m.MarshalJSON()
		if err != nil {
			return "", fmt.Errorf("failed to marshal input to JSON: %w", err)
		}
		e := json.NewDecoder(bytes.NewReader(buf))
		e.UseNumber()
		var i2 interface{}
		err = e.Decode(&i2)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		s, err := StyleParamWithOptions(style, explode, paramName, i2, StyleParamOptions{
			ParamLocation: paramLocation,
			AllowReserved: allowReserved,
		})
		if err != nil {
			return "", fmt.Errorf("error style JSON structure: %w", err)
		}
		return s, nil
	}

	// Otherwise, we need to build a dictionary of the struct's fields. Each
	// field may only be a primitive value.
	v := reflect.ValueOf(value)
	t := reflect.TypeOf(value)
	fieldDict := make(map[string]string)

	for i := 0; i < t.NumField(); i++ {
		fieldT := t.Field(i)
		// Find the json annotation on the field, and use the json specified
		// name if available, otherwise, just the field name.
		tag := fieldT.Tag.Get("json")
		fieldName := fieldT.Name
		if tag != "" {
			tagParts := strings.Split(tag, ",")
			name := tagParts[0]
			if name != "" {
				fieldName = name
			}
		}
		f := v.Field(i)

		// Unset optional fields will be nil pointers, skip over those.
		if f.Type().Kind() == reflect.Ptr && f.IsNil() {
			continue
		}
		str, err := primitiveToString(f.Interface())
		if err != nil {
			return "", fmt.Errorf("error formatting '%s': %w", paramName, err)
		}
		fieldDict[fieldName] = str
	}

	return processFieldDict(style, explode, paramName, paramLocation, allowReserved, fieldDict)
}
