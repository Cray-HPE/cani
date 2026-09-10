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
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

type RequiredParameterError struct {
	ParamName string
}

func (e *RequiredParameterError) Error() string {
	return fmt.Sprintf("query parameter '%s' is required", e.ParamName)
}

type BindQueryParameterOptions struct {
	Type          string
	Format        string
	AllowReserved bool
}

func BindQueryParameter(style string, explode bool, required bool, paramName string,
	queryParams url.Values, dest interface{}) error {
	return BindQueryParameterWithOptions(style, explode, required, paramName, queryParams, dest, BindQueryParameterOptions{})
}

func BindQueryParameterWithOptions(style string, explode bool, required bool, paramName string,
	queryParams url.Values, dest interface{}, opts BindQueryParameterOptions) error {
	destination := reflect.Indirect(reflect.ValueOf(dest))
	value := destination
	var output interface{}
	extraIndirect := !required && value.Kind() == reflect.Pointer
	if !extraIndirect {
		output = dest
	} else {
		if value.IsNil() {
			output = reflect.New(value.Type().Elem()).Interface()
		} else {
			output = value.Interface()
		}
		value = reflect.Indirect(reflect.ValueOf(output))
	}

	kind := value.Type().Kind()
	switch style {
	case "form":
		var parts []string
		if explode {
			values, found := queryParams[paramName]
			var err error
			switch kind {
			case reflect.Slice:
				if !found {
					if required {
						return &RequiredParameterError{ParamName: paramName}
					}
					return nil
				}
				if opts.Format == "byte" && isByteSlice(value.Type()) {
					if len(values) != 1 {
						return fmt.Errorf("expected single base64 value for byte slice parameter '%s', got %d values", paramName, len(values))
					}
					decoded, decodeErr := base64Decode(values[0])
					if decodeErr != nil {
						return fmt.Errorf("error decoding base64 parameter '%s': %w", paramName, decodeErr)
					}
					value.SetBytes(decoded)
				} else {
					err = bindSplitPartsToDestinationArray(values, output)
				}
			case reflect.Struct:
				fieldsPresent, bindErr := bindParamsToExplodedObject(paramName, queryParams, output)
				err = bindErr
				if !fieldsPresent {
					return nil
				}
			default:
				if len(values) == 0 {
					if required {
						return &RequiredParameterError{ParamName: paramName}
					}
					return nil
				}
				if len(values) != 1 {
					return fmt.Errorf("multiple values for single value parameter '%s'", paramName)
				}
				if !found {
					if required {
						return &RequiredParameterError{ParamName: paramName}
					}
					return nil
				}
				err = BindStringToObject(values[0], output)
			}
			if err != nil {
				return err
			}
			if extraIndirect {
				destination.Set(reflect.ValueOf(output))
			}
			return nil
		}

		values, found := queryParams[paramName]
		if !found {
			if required {
				return &RequiredParameterError{ParamName: paramName}
			}
			return nil
		}
		if len(values) != 1 {
			return fmt.Errorf("parameter '%s' is not exploded, but is specified multiple times", paramName)
		}
		parts = strings.Split(values[0], ",")
		var err error
		switch kind {
		case reflect.Slice:
			if opts.Format == "byte" && isByteSlice(value.Type()) {
				decoded, decodeErr := base64Decode(strings.Join(parts, ","))
				if decodeErr != nil {
					return fmt.Errorf("error decoding base64 parameter '%s': %w", paramName, decodeErr)
				}
				value.SetBytes(decoded)
			} else {
				err = bindSplitPartsToDestinationArray(parts, output)
			}
		case reflect.Struct:
			err = bindSplitPartsToDestinationStruct(paramName, parts, explode, output)
		default:
			if len(parts) == 0 {
				if required {
					return &RequiredParameterError{ParamName: paramName}
				}
				return nil
			}
			if len(parts) != 1 {
				return fmt.Errorf("multiple values for single value parameter '%s'", paramName)
			}
			err = BindStringToObject(parts[0], output)
		}
		if err != nil {
			return err
		}
		if extraIndirect {
			destination.Set(reflect.ValueOf(output))
		}
		return nil
	case "deepObject":
		if !explode {
			return errors.New("deepObjects must be exploded")
		}
		return UnmarshalDeepObject(dest, paramName, queryParams)
	case "spaceDelimited", "pipeDelimited":
		return fmt.Errorf("query arguments of style '%s' aren't yet supported", style)
	default:
		return fmt.Errorf("style '%s' on parameter '%s' is invalid", style, paramName)
	}
}
