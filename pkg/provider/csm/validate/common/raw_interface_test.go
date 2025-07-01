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
	"testing"
)

type values struct {
	Object    interface{}
	Level0Map map[string]interface{}
	Level1Map map[string]interface{}
	Level2Map map[string]interface{}
}

func newValues() *values {
	v := &values{}
	v.Level0Map = make(map[string]interface{})
	v.Level1Map = make(map[string]interface{})
	v.Level2Map = make(map[string]interface{})

	v.Level0Map["level1"] = v.Level1Map
	v.Level1Map["level2"] = make(map[string]interface{})

	v.Object = v.Level0Map
	return v
}

func TestGetMap(t *testing.T) {
	v := newValues()
	m, found := GetMap(v.Object)
	if !found {
		t.Fatalf("Level0: Expected the top level to be a map, \n  Object: %v, \n  Map: %v", v.Object, m)
	}

	if !reflect.DeepEqual(m, v.Level0Map) {
		t.Fatalf("Level0: Expected the returned map to equal, \n  expected: %v, \n  actual: %v", v.Level0Map, m)
	}

	m, found = GetMap(v.Object, "level1")
	if !found {
		t.Fatalf("Level1: Expected the top level to be a map, \n  Object: %v, \n  Map: %v", v.Object, m)
	}

	if !reflect.DeepEqual(m, v.Level1Map) {
		t.Fatalf("Level1: Expected the returned map to equal, \n  expected: %v, \n  actual: %v", v.Level1Map, m)
	}

	m, found = GetMap(v.Object, "level1", "level2")
	if !found {
		t.Fatalf("Level1: Expected the top level to be a map, \n  Object: %v, \n  Map: %v", v.Object, m)
	}

	if !reflect.DeepEqual(m, v.Level2Map) {
		t.Fatalf("Level1: Expected the returned map to equal, \n  expected: %v, \n  actual: %v", v.Level2Map, m)
	}
}
