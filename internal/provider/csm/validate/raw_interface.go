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

package validate

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
		result := make([]map[string]interface{}, len(slice))
		for i, s := range slice {
			result[i], _ = GetMap(s)
		}
		return result, true
	}
	return make([]map[string]interface{}, 0), false
}

func GetSlice(raw interface{}, id ...string) ([]interface{}, bool) {
	idsUpToSlice := id[:pos(len(id)-1)]
	objectMap, found := GetMap(raw, idsUpToSlice...)

	if found {
		key := id[len(id)-1]
		if value, found := objectMap[key]; found {
			if slice, ok := value.([]interface{}); ok {
				return slice, true
			}
		}
	}
	return make([]interface{}, 0), false
}

func GetMap(raw interface{}, ids ...string) (map[string]interface{}, bool) {
	if len(ids) == 0 {
		if m, ok := raw.(map[string]interface{}); ok {
			return m, true
		}
	}

	object := raw
	for _, key := range ids {
		if m, ok := object.(map[string]interface{}); ok {
			if value, found := m[key]; found {
				object = value
			} else {
				return make(map[string]interface{}), false
			}
		}
	}
	return object.(map[string]interface{}), true
}
