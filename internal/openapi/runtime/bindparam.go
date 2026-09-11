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
	"encoding"
	"fmt"
	"net/url"
	"reflect"
)

// BindStyledParameter binds a parameter as described in the Path Parameters
// section here to a Go object:
// https://swagger.io/docs/specification/serialization/
// It is a backward compatible function to clients generated with codegen
// up to version v1.5.5. v1.5.6+ calls the function below.
// Deprecated: BindStyledParameter is deprecated.
func BindStyledParameter(style string, explode bool, paramName string,
	value string, dest interface{}) error {
	return BindStyledParameterWithOptions(style, paramName, value, dest, BindStyledParameterOptions{
		ParamLocation: ParamLocationUndefined,
		Explode:       explode,
		Required:      true,
	})
}

// BindStyledParameterWithLocation binds a parameter as described in the Path Parameters
// section here to a Go object:
// https://swagger.io/docs/specification/serialization/
// This is a compatibility function which is used by oapi-codegen v2.0.0 and earlier.
// Deprecated: BindStyledParameterWithLocation is deprecated.
func BindStyledParameterWithLocation(style string, explode bool, paramName string,
	paramLocation ParamLocation, value string, dest interface{}) error {
	return BindStyledParameterWithOptions(style, paramName, value, dest, BindStyledParameterOptions{
		ParamLocation: paramLocation,
		Explode:       explode,
		Required:      true, // This emulates behavior before the required parameter was optional.
	})
}

// BindStyledParameterOptions defines optional arguments for BindStyledParameterWithOptions
type BindStyledParameterOptions struct {
	// ParamLocation tells us where the parameter is located in the request.
	ParamLocation ParamLocation
	// Whether the parameter should use exploded structure
	Explode bool
	// Whether the parameter is required in the query
	Required bool
	// Type is the OpenAPI type of the parameter (e.g. "string", "integer").
	Type string
	// Format is the OpenAPI format of the parameter (e.g. "uuid", "date-time").
	Format string
	// ValueIsUnescaped indicates the value was already unescaped by the caller
	// (e.g. std-http's r.PathValue), so no further unescaping is needed.
	ValueIsUnescaped bool
}

// BindStyledParameterWithOptions binds a parameter as described in the Path Parameters
// section here to a Go object:
// https://swagger.io/docs/specification/serialization/
func BindStyledParameterWithOptions(style string, paramName string, value string, dest any, opts BindStyledParameterOptions) error {
	if opts.Required {
		if value == "" {
			return fmt.Errorf("parameter '%s' is empty, can't bind its value", paramName)
		}
	}

	// Based on the location of the parameter, we need to unescape it properly,
	// unless the caller already provided an unescaped value.
	var err error
	if !opts.ValueIsUnescaped {
		switch opts.ParamLocation {
		case ParamLocationQuery, ParamLocationUndefined:
			// We unescape undefined parameter locations here for older generated code,
			// since prior to this refactoring, they always query unescaped.
			value, err = url.QueryUnescape(value)
			if err != nil {
				return fmt.Errorf("error unescaping query parameter '%s': %w", paramName, err)
			}
		case ParamLocationPath:
			value, err = url.PathUnescape(value)
			if err != nil {
				return fmt.Errorf("error unescaping path parameter '%s': %w", paramName, err)
			}
		default:
			// Headers and cookies aren't escaped.
		}
	}

	// If the destination implements encoding.TextUnmarshaler we use it for binding
	if tu, ok := dest.(encoding.TextUnmarshaler); ok {
		if err := tu.UnmarshalText([]byte(value)); err != nil {
			return fmt.Errorf("error unmarshaling '%s' text as %T: %w", value, dest, err)
		}

		return nil
	}

	// Everything comes in by pointer, dereference it
	v := reflect.Indirect(reflect.ValueOf(dest))

	// This is the basic type of the destination object.
	t := v.Type()

	if t.Kind() == reflect.Struct {
		// We've got a destination object, we'll create a JSON representation
		// of the input value, and let the json library deal with the unmarshaling
		parts, err := splitStyledParameter(style, opts.Explode, true, paramName, value)
		if err != nil {
			return err
		}

		return bindSplitPartsToDestinationStruct(paramName, parts, opts.Explode, dest)
	}

	if t.Kind() == reflect.Slice {
		if opts.Format == "byte" && isByteSlice(t) {
			parts, err := splitStyledParameter(style, opts.Explode, false, paramName, value)
			if err != nil {
				return fmt.Errorf("error splitting input '%s' into parts: %w", value, err)
			}
			if len(parts) != 1 {
				return fmt.Errorf("expected single base64 value for byte slice parameter '%s', got %d parts", paramName, len(parts))
			}
			decoded, err := base64Decode(parts[0])
			if err != nil {
				return fmt.Errorf("error decoding base64 parameter '%s': %w", paramName, err)
			}
			v.SetBytes(decoded)
			return nil
		}

		// Chop up the parameter into parts based on its style
		parts, err := splitStyledParameter(style, opts.Explode, false, paramName, value)
		if err != nil {
			return fmt.Errorf("error splitting input '%s' into parts: %w", value, err)
		}

		return bindSplitPartsToDestinationArray(parts, dest)
	}

	// Try to bind the remaining types as a base type.
	return BindStringToObject(value, dest)
}
