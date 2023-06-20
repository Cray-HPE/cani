/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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
/*
MIT License

(C) Copyright 2023 Hewlett Packard Enterprise Development LP

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/

package common

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/fatih/structs"
)

func pos(length int) int {
	if length < 0 {
		return 0
	}
	return length
}

// Returns a list of maps
func GetSliceOfMaps(raw interface{}, id ...string) ([]map[string]interface{}, bool) {
	slice, found := GetSlice(raw, id...)
	if found {
		if rawSlice, ok := slice.([]interface{}); ok {
			result := make([]map[string]interface{}, len(rawSlice))
			for i, s := range rawSlice {
				result[i], _ = GetMap(s)
			}
			return result, true
		}
	}
	return make([]map[string]interface{}, 0), false
}

// The returned interface{} will be a slice or an array
func GetSlice(raw interface{}, id ...string) (interface{}, bool) {
	if raw == nil {
		return make([]interface{}, 0), false
	}
	if len(id) == 0 {
		kind := reflect.TypeOf(raw).Kind()
		if kind == reflect.Slice || kind == reflect.Array {
			return raw, true
		} else {
			return make([]interface{}, 0), false
		}
	}
	idsUpToSlice := id[:pos(len(id)-1)]
	objectMap, found := GetMap(raw, idsUpToSlice...)

	if found {
		key := id[len(id)-1]
		if value, found := objectMap[key]; found {
			kind := reflect.TypeOf(value).Kind()
			if kind == reflect.Slice || kind == reflect.Array {
				return value, true
			}
		}
	}
	return make([]interface{}, 0), false
}

func GetSliceOfStrings(raw interface{}, id ...string) ([]string, bool) {
	slice, found := GetSlice(raw, id...)
	if found {
		if rawSlice, ok := slice.([]interface{}); ok {
			stringSlice := make([]string, 0)
			for _, val := range rawSlice {
				stringSlice = append(stringSlice, fmt.Sprintf("%v", val))
			}
			return stringSlice, true
		} else if stringSlice, ok := slice.([]string); ok {
			return stringSlice, true
		}
	}
	return make([]string, 0), false
}

func ToInt(raw interface{}) (int64, bool) {
	if raw == nil {
		return -1, false
	}
	str := fmt.Sprintf("%v", raw)
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return -2, false
	}
	return i, true
}

func GetString(raw interface{}, ids ...string) (string, bool) {
	if raw == nil || len(ids) == 0 {
		return "", false
	}

	idsUpToString := ids[:pos(len(ids)-1)]
	object, found := GetMap(raw, idsUpToString...)
	if !found {
		return "", false
	}

	key := ids[len(ids)-1]
	val, ok := object[key]
	if !ok {
		return "", false
	}
	return fmt.Sprintf("%v", val), true
}

func GetMap(raw interface{}, ids ...string) (map[string]interface{}, bool) {
	if len(ids) == 0 {
		return toMap(raw)
	}

	object := raw
	for _, key := range ids {
		if m, ok := toMap(object); ok {
			if value, found := m[key]; found {
				object = value
			} else {
				return make(map[string]interface{}), false
			}
		}
	}
	return toMap(object)
}

func toMap(raw interface{}) (map[string]interface{}, bool) {
	if raw != nil {
		if m, ok := raw.(map[string]interface{}); ok {
			return m, true
		} else if structs.IsStruct(raw) {
			return structs.Map(raw), true
		}
	}

	return make(map[string]interface{}), false
}
