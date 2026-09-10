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
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/Cray-HPE/cani/internal/openapi/types"
)

func bindParamsToExplodedObject(paramName string, values url.Values, dest interface{}) (bool, error) {
	binder, value, valueType := indirect(dest)
	if binder != nil {
		if _, found := values[paramName]; !found {
			return false, nil
		}
		return true, BindStringToObject(values.Get(paramName), dest)
	}
	if valueType.Kind() != reflect.Struct {
		return false, fmt.Errorf("unmarshaling query arg '%s' into wrong type", paramName)
	}

	fieldsPresent := false
	for index := 0; index < valueType.NumField(); index++ {
		fieldType := valueType.Field(index)
		if !value.Field(index).CanSet() {
			continue
		}
		fieldName := fieldType.Name
		if tag := fieldType.Tag.Get("json"); tag != "" {
			if name := strings.Split(tag, ",")[0]; name != "" {
				fieldName = name
			}
		}
		fieldValues, found := values[fieldName]
		if !found {
			continue
		}
		if len(fieldValues) != 1 {
			return false, fmt.Errorf("field '%s' specified multiple times for param '%s'", fieldName, paramName)
		}
		if err := BindStringToObject(fieldValues[0], value.Field(index).Addr().Interface()); err != nil {
			return false, fmt.Errorf("could not bind query arg '%s' to request object: %s'", paramName, err)
		}
		fieldsPresent = true
	}
	return fieldsPresent, nil
}

func indirect(dest interface{}) (interface{}, reflect.Value, reflect.Type) {
	value := reflect.ValueOf(dest)
	if value.Type().NumMethod() > 0 && value.CanInterface() {
		if binder, ok := value.Interface().(Binder); ok {
			return binder, reflect.Value{}, nil
		}
	}
	value = reflect.Indirect(value)
	valueType := value.Type()
	if valueType.ConvertibleTo(reflect.TypeOf(time.Time{})) ||
		valueType.ConvertibleTo(reflect.TypeOf(types.Date{})) {
		return dest, reflect.Value{}, nil
	}
	return nil, value, valueType
}
