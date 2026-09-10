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
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func sortedKeys(strMap map[string]string) []string {
	keys := make([]string, 0, len(strMap))
	for key := range strMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func styleMap(style string, explode bool, paramName string, paramLocation ParamLocation, allowReserved bool, value interface{}) (string, error) {
	if style == "deepObject" {
		if !explode {
			return "", errors.New("deepObjects must be exploded")
		}
		return MarshalDeepObject(value, paramName)
	}
	mapValue := reflect.ValueOf(value)
	fieldDict := make(map[string]string)
	for _, fieldName := range mapValue.MapKeys() {
		str, err := primitiveToString(mapValue.MapIndex(fieldName).Interface())
		if err != nil {
			return "", fmt.Errorf("error formatting '%s': %w", paramName, err)
		}
		fieldDict[fieldName.String()] = str
	}
	return processFieldDict(style, explode, paramName, paramLocation, allowReserved, fieldDict)
}

func processFieldDict(style string, explode bool, paramName string, paramLocation ParamLocation, allowReserved bool, fieldDict map[string]string) (string, error) {
	var parts []string
	if style != "deepObject" {
		if explode {
			for _, key := range sortedKeys(fieldDict) {
				parts = append(parts, key+"="+escapeParameterString(fieldDict[key], paramLocation, allowReserved))
			}
		} else {
			for _, key := range sortedKeys(fieldDict) {
				parts = append(parts, key, escapeParameterString(fieldDict[key], paramLocation, allowReserved))
			}
		}
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
			separator = prefix
		} else {
			separator = ","
		}
	case "matrix":
		if explode {
			separator, prefix = ";", ";"
		} else {
			separator, prefix = ",", fmt.Sprintf(";%s=", escapedName)
		}
	case "form":
		if explode {
			separator = "&"
		} else {
			prefix, separator = fmt.Sprintf("%s=", escapedName), ","
		}
	case "deepObject":
		if !explode {
			return "", fmt.Errorf("deepObject parameters must be exploded")
		}
		for _, key := range sortedKeys(fieldDict) {
			parts = append(parts, fmt.Sprintf("%s[%s]=%s", escapedName, key, fieldDict[key]))
		}
		separator = "&"
	default:
		return "", fmt.Errorf("unsupported style '%s'", style)
	}
	return prefix + strings.Join(parts, separator), nil
}

func stylePrimitive(style string, explode bool, paramName string, paramLocation ParamLocation, allowReserved bool, value interface{}) (string, error) {
	strValue, err := primitiveToString(value)
	if err != nil {
		return "", err
	}
	var prefix string
	escapedName := escapeParameterName(paramName, paramLocation)
	switch style {
	case "simple":
	case "label":
		prefix = "."
	case "matrix":
		prefix = fmt.Sprintf(";%s=", escapedName)
	case "form":
		prefix = fmt.Sprintf("%s=", escapedName)
	default:
		return "", fmt.Errorf("unsupported style '%s'", style)
	}
	return prefix + escapeParameterString(strValue, paramLocation, allowReserved), nil
}

func primitiveToString(value interface{}) (string, error) {
	if result, ok := marshalKnownTypes(value); ok {
		return result, nil
	}
	reflected := reflect.Indirect(reflect.ValueOf(value))
	var output string
	switch reflected.Type().Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		output = strconv.FormatInt(reflected.Int(), 10)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		output = strconv.FormatUint(reflected.Uint(), 10)
	case reflect.Float64:
		output = strconv.FormatFloat(reflected.Float(), 'f', -1, 64)
	case reflect.Float32:
		output = strconv.FormatFloat(reflected.Float(), 'f', -1, 32)
	case reflect.Bool:
		output = strconv.FormatBool(reflected.Bool())
	case reflect.String:
		output = reflected.String()
	case reflect.Struct:
		if id, ok := value.(uuid.UUID); ok {
			return id.String(), nil
		}
		if marshaler, ok := value.(json.Marshaler); ok {
			buffer, err := marshaler.MarshalJSON()
			if err != nil {
				return "", fmt.Errorf("failed to marshal input to JSON: %w", err)
			}
			decoder := json.NewDecoder(bytes.NewReader(buffer))
			decoder.UseNumber()
			var decoded interface{}
			if err := decoder.Decode(&decoded); err != nil {
				return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
			}
			output, err = primitiveToString(decoded)
			if err != nil {
				return "", fmt.Errorf("error convert JSON structure: %w", err)
			}
			break
		}
		fallthrough
	default:
		stringer, ok := value.(fmt.Stringer)
		if !ok {
			return "", fmt.Errorf("unsupported type %s", reflect.TypeOf(value).String())
		}
		output = stringer.String()
	}
	return output, nil
}

func escapeParameterName(name string, paramLocation ParamLocation) string {
	return escapeParameterString(name, paramLocation, false)
}

func escapeParameterString(value string, paramLocation ParamLocation, allowReserved bool) string {
	switch paramLocation {
	case ParamLocationQuery:
		if allowReserved {
			return escapeQueryAllowReserved(value)
		}
		return url.QueryEscape(value)
	case ParamLocationPath:
		return url.PathEscape(value)
	default:
		return value
	}
}

func escapeQueryAllowReserved(value string) string {
	const reserved = `:/?#[]@!$&'()*+,;=`
	var buffer strings.Builder
	for _, char := range []byte(value) {
		if isUnreserved(char) || strings.IndexByte(reserved, char) >= 0 {
			buffer.WriteByte(char)
		} else {
			fmt.Fprintf(&buffer, "%%%02X", char)
		}
	}
	return buffer.String()
}

func isUnreserved(char byte) bool {
	return char >= 'A' && char <= 'Z' || char >= 'a' && char <= 'z' ||
		char >= '0' && char <= '9' || char == '-' || char == '.' || char == '_' || char == '~'
}
