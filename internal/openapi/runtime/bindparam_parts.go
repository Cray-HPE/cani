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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// This is a complex set of operations, but each given parameter style can be
// packed together in multiple ways, using different styles of separators, and
// different packing strategies based on the explode flag. This function takes
// as input any parameter format, and unpacks it to a simple list of strings
// or key-values which we can then treat generically.
// Why, oh why, great Swagger gods, did you have to make this so complicated?
func splitStyledParameter(style string, explode bool, object bool, paramName string, value string) ([]string, error) {
	switch style {
	case "simple":
		return strings.Split(value, ","), nil
	case "label":
		if explode {
			parts := strings.Split(value, ".")
			if parts[0] != "" {
				return nil, fmt.Errorf("invalid format for label parameter '%s', should start with '.'", paramName)
			}
			return parts[1:], nil
		}
		if value[0] != '.' {
			return nil, fmt.Errorf("invalid format for label parameter '%s', should start with '.'", paramName)
		}
		return strings.Split(value[1:], ","), nil
	case "matrix":
		if explode {
			parts := strings.Split(value, ";")
			if parts[0] != "" {
				return nil, fmt.Errorf("invalid format for matrix parameter '%s', should start with ';'", paramName)
			}
			parts = parts[1:]
			if !object {
				prefix := paramName + "="
				for index := range parts {
					parts[index] = strings.TrimPrefix(parts[index], prefix)
				}
			}
			return parts, nil
		}
		prefix := ";" + paramName + "="
		if !strings.HasPrefix(value, prefix) {
			return nil, fmt.Errorf("expected parameter '%s' to start with %s", paramName, prefix)
		}
		return strings.Split(strings.TrimPrefix(value, prefix), ","), nil
	case "form":
		if explode {
			parts := strings.Split(value, "&")
			if !object {
				prefix := paramName + "="
				for index := range parts {
					parts[index] = strings.TrimPrefix(parts[index], prefix)
				}
			}
			return parts, nil
		}
		parts := strings.Split(value, ",")
		prefix := paramName + "="
		for index := range parts {
			parts[index] = strings.TrimPrefix(parts[index], prefix)
		}
		return parts, nil
	}
	return nil, fmt.Errorf("unhandled parameter style: %s", style)
}

func bindSplitPartsToDestinationArray(parts []string, dest interface{}) error {
	value := reflect.Indirect(reflect.ValueOf(dest))
	array := reflect.MakeSlice(value.Type(), len(parts), len(parts))
	for index, part := range parts {
		if err := BindStringToObject(part, array.Index(index).Addr().Interface()); err != nil {
			return fmt.Errorf("error setting array element: %w", err)
		}
	}
	value.Set(array)
	return nil
}

func bindSplitPartsToDestinationStruct(paramName string, parts []string, explode bool, dest interface{}) error {
	var fields []string
	if explode {
		fields = make([]string, len(parts))
		for index, property := range parts {
			propertyParts := strings.Split(property, "=")
			if len(propertyParts) != 2 {
				return fmt.Errorf("parameter '%s' has invalid exploded format", paramName)
			}
			fields[index] = "\"" + propertyParts[0] + "\":\"" + propertyParts[1] + "\""
		}
	} else {
		if len(parts)%2 != 0 {
			return fmt.Errorf("parameter '%s' has invalid format, property/values need to be pairs", paramName)
		}
		fields = make([]string, len(parts)/2)
		for index := 0; index < len(parts); index += 2 {
			fields[index/2] = "\"" + parts[index] + "\":\"" + parts[index+1] + "\""
		}
	}
	if err := json.Unmarshal([]byte("{"+strings.Join(fields, ",")+"}"), dest); err != nil {
		return fmt.Errorf("error binding parameter %s fields: %w", paramName, err)
	}
	return nil
}
