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

package common

import (
	"reflect"
	"strings"
	"testing"
)

type structValue struct {
	Key0 string
	Key1 string
}

type values struct {
	Object      interface{}
	Level0Map   map[string]interface{}
	Level1Map   map[string]interface{}
	Level2Map   map[string]interface{}
	StructValue structValue
	MapValue    map[string]interface{}
}

const (
	level0       = "level0"
	level1       = "level1"
	level2       = "level2"
	key0         = "Key0"
	value0       = "Value0"
	key1         = "Key1"
	value1       = "Value1"
	structMapKey = "structMapKey"
	mapMapKey    = "mapMapKey"
)

func newValues() *values {
	v := &values{}
	v.Level0Map = make(map[string]interface{})
	v.Level1Map = make(map[string]interface{})
	v.Level2Map = make(map[string]interface{})
	v.Object = v.Level0Map

	v.Level0Map[level1] = v.Level1Map
	v.Level1Map[level2] = v.Level2Map

	v.StructValue = structValue{
		Key0: value0,
		Key1: value1,
	}
	v.MapValue = make(map[string]interface{})
	v.MapValue[key0] = value0
	v.MapValue[key1] = value1
	v.Level2Map[structMapKey] = v.StructValue
	v.Level2Map[mapMapKey] = v.MapValue

	return v
}

func TestGetMap(t *testing.T) {
	v := newValues()
	m, found := GetMap(v.Object)
	if !found {
		t.Fatalf("%s: Expected the top level to be a map, \n  Object: %v, \n  Map: %v", level0, v.Object, m)
	}

	if !reflect.DeepEqual(m, v.Level0Map) {
		t.Fatalf("%s: Expected the returned map to equal, \n  expected: %v, \n  actual: %v", level0, v.Level0Map, m)
	}

	m, found = GetMap(v.Object, level1)
	if !found {
		t.Fatalf("%s: Map not found, \n  Object: %v, \n  Map: %v", level1, v.Object, m)
	}

	if !reflect.DeepEqual(m, v.Level1Map) {
		t.Fatalf("%s: Expected the returned map to equal, \n  expected: %v, \n  actual: %v", level1, v.Level1Map, m)
	}

	m, found = GetMap(v.Object, level1, level2)
	if !found {
		t.Fatalf("%s: Map not found, \n  Object: %v, \n  Map: %v", level2, v.Object, m)
	}

	if !reflect.DeepEqual(m, v.Level2Map) {
		t.Fatalf("%s: Expected the returned map to equal, \n  expected: %v, \n  actual: %v", level2, v.Level2Map, m)
	}
}

func validateMapWithValues(t *testing.T, object interface{}, keys []string) {
	m, found := GetMap(object, keys...)
	if !found {
		t.Fatalf("%s: Map not found,", strings.Join(keys, ","))
	}

	v0, found := m[key0]
	if !found {
		t.Fatalf("Missing key %s in %v", key0, m)
	}
	if v0 != value0 {
		t.Fatalf("Wrong value. expected: %s, actual: %s, for key %s in %v", value0, v0, key0, m)

	}

	v1, found := m[key1]
	if !found {
		t.Fatalf("Missing key %s in %v", key1, m)
	}
	if v1 != value1 {
		t.Fatalf("Wrong value. expected: %s, actual: %s, for key %s in %v", value1, v1, key1, m)

	}
}

func TestGetMapWithValues(t *testing.T) {
	v := newValues()

	keysToStruct := []string{level1, level2, structMapKey}
	validateMapWithValues(t, v.Object, keysToStruct)

	keysToMap := []string{level1, level2, mapMapKey}
	validateMapWithValues(t, v.Object, keysToMap)
}
