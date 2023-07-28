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
package csm

import (
	"testing"
)

func TestSetVlan(t *testing.T) {
	props := &CabinetMetadata{}

	modified, err := setVlan("", props)
	if err != nil {
		t.Errorf("set empty vlan unexpected error: %v", err)
	}
	if modified {
		t.Errorf("set empty vlan, modified should be false")
	}

	value := "1000"
	expected := 1000
	props.HMNVlan = new(int)
	*props.HMNVlan = expected
	modified, err = setVlan(value, props)
	if err != nil {
		t.Errorf("set same vlan unexpected error: %v", err)
	}
	if modified {
		t.Errorf("set same vlan, modified should be false")
	}
	if *props.HMNVlan != expected {
		t.Errorf("set same vlan, wrong value, expected: %v, actual: %v", expected, *props.HMNVlan)
	}

	value = "3333"
	expected = 3333
	modified, err = setVlan(value, props)
	if err != nil {
		t.Errorf("set new vlan unexpected error: %v", err)
	}
	if !modified {
		t.Errorf("set new vlan, modified should be true")
	}
	if *props.HMNVlan != expected {
		t.Errorf("set new vlan, wrong value, expected: %v, actual: %v", expected, *props.HMNVlan)
	}

	// test set vlan to empty
	modified, err = setVlan("", props)
	if err != nil {
		t.Errorf("set vlan to empty, unexpected error: %v", err)
	}
	if !modified {
		t.Errorf("set vlan to empty, modified should be true")
	}
	if props.HMNVlan != nil {
		t.Errorf("set vlan to empty, wrong value, expected: nil, actual: %v", *props.HMNVlan)
	}

	// test set vlan to empty again
	modified, err = setVlan("", props)
	if err != nil {
		t.Errorf("set vlan to empty again, unexpected error: %v", err)
	}
	if modified {
		t.Errorf("set vlan to empty again, modified should be false")
	}
	if props.HMNVlan != nil {
		t.Errorf("set vlan to empty again, wrong value, expected: nil, actual: %v", *props.HMNVlan)
	}
}

func TestSetRole(t *testing.T) {
	props := &NodeMetadata{}

	// test empty value
	modified := setRole("", props)
	if modified {
		t.Errorf("set empty role, modified should be false")
	}

	// test set to Compute
	expected := "Compute"
	modified = setRole(expected, props)
	if !modified {
		t.Errorf("set role Compute, modified should be true")
	}
	actual := *props.Role
	if expected != actual {
		t.Errorf("set role Compute, expected: %s, actual: %v", expected, actual)
	}

	// test set to same value
	modified = setRole(expected, props)
	if modified {
		t.Errorf("set role to same value, modified should be false")
	}
	actual = *props.Role
	if expected != actual {
		t.Errorf("set role to same value, expected: %s, actual: %v", expected, actual)
	}

	// test set to new value
	expected = "Application"
	modified = setRole(expected, props)
	if !modified {
		t.Errorf("set role Application, modified should be true")
	}
	actual = *props.Role
	if expected != actual {
		t.Errorf("set role Application, expected: %s, actual: %v", expected, actual)
	}

	// test set to empty
	modified = setRole("", props)
	if !modified {
		t.Errorf("set role to empty, modified should be true")
	}
	if props.Role != nil {
		t.Errorf("set role to empty, expected: nil, actual: %v", *props.Role)
	}
}

func TestSetSubRole(t *testing.T) {
	props := &NodeMetadata{}

	// test empty value
	modified := setSubRole("", props)
	if modified {
		t.Errorf("set empty subrole, modified should be false")
	}

	// test set to Compute
	expected := "Worker"
	modified = setSubRole(expected, props)
	if !modified {
		t.Errorf("set subrole Worker, modified should be true")
	}
	actual := props.SubRole
	if actual == nil || expected != *actual {
		t.Errorf("set subrole worker, expected: %s, actual: %v", expected, actual)
	}

	// test set to same value
	modified = setSubRole(expected, props)
	if modified {
		t.Errorf("set subrole to same value, modified should be false")
	}
	actual = props.SubRole
	if actual == nil || expected != *actual {
		t.Errorf("set subrole to same value, expected: %s, actual: %v", expected, actual)
	}

	// test set to new value
	expected = "Storage"
	modified = setSubRole(expected, props)
	if !modified {
		t.Errorf("set subrole Storage, modified should be true")
	}
	actual = props.SubRole
	if actual == nil || expected != *actual {
		t.Errorf("set subrole Storage, expected: %s, actual: %v", expected, actual)
	}

	// set back to empty
	modified = setSubRole("", props)
	if !modified {
		t.Errorf("set back to empty subrole, modified should be true")
	}
	actual = props.SubRole
	if actual != nil {
		t.Errorf("set subrole Storage, expected: nil, actual: %v", actual)
	}
}

func TestSetNid(t *testing.T) {
	props := &NodeMetadata{}

	// test empty value
	modified, err := setNid("", props)
	if err != nil {
		t.Errorf("set empty nid, unexpected error: %v", err)
	}
	if modified {
		t.Errorf("set empty nid, modified should be false")
	}

	// Test good value
	expected := 42
	modified, err = setNid("42", props)
	if err != nil {
		t.Errorf("set nid, unexpected error: %v", err)
	}
	if !modified {
		t.Errorf("set nid, modified should be true")
	}
	actual := props.Nid
	if actual == nil || *actual != expected {
		t.Errorf("set nid, expected: %v, actual: %v", expected, actual)
	}

	// Test same value
	modified, err = setNid("42", props)
	if err != nil {
		t.Errorf("set nid to same value, unexpected error: %v", err)
	}
	if modified {
		t.Errorf("set nid to same value, modified should be false")
	}
	actual = props.Nid
	if actual == nil || *actual != expected {
		t.Errorf("set nid to same value, expected: %v, actual: %v", expected, actual)
	}

	// Test bad value
	modified, err = setNid("42.42", props)
	if err == nil {
		t.Errorf("set nid to bad value, expected to get an error: %v", err)
	}
	if modified {
		t.Errorf("set nid to bad value, modified should be false")
	}
	actual = props.Nid
	if actual == nil || *actual != expected {
		t.Errorf("set nid to bad value, expected: %v, actual: %v", expected, actual)
	}

	// Test new value
	expected = 3
	modified, err = setNid("3", props)
	if err != nil {
		t.Errorf("set nid to new value, unexpected error: %v", err)
	}
	if !modified {
		t.Errorf("set nid to new value, modified should be true")
	}
	actual = props.Nid
	if actual == nil || *actual != expected {
		t.Errorf("set nid to new value, expected: %v, actual: %v", expected, actual)
	}

	// test empty value
	modified, err = setNid("", props)
	if err != nil {
		t.Errorf("set nid back to empty, unexpected error: %v", err)
	}
	if !modified {
		t.Errorf("set nid back to empty, modified should be true")
	}
	if props.Nid != nil {
		t.Errorf("set nid back to empty, expected: nil, actual: %v", *props.Nid)
	}
}

func TestSetAlias(t *testing.T) {
	props := &NodeMetadata{}

	// Test empty alias
	modified := setAlias("", props)
	if modified {
		t.Errorf("set empty alias, modified should be false")
	}

	// Test alias
	expected := "rabbit"
	modified = setAlias(expected, props)
	if !modified {
		t.Errorf("set alias, modified should be true")
	}
	actual := props.Alias
	if len(actual) != 1 {
		t.Errorf("set alias, expected length: %d, actual length: %d", 1, len(actual))
	} else if actual[0] != expected {
		t.Errorf("set alias, expected expected: %v, actual: %v, aliases: %v", expected, actual[0], actual)
	}

	// Test same alias
	modified = setAlias(expected, props)
	if modified {
		t.Errorf("set same alias, modified should be false")
	}
	actual = props.Alias
	if len(actual) != 1 {
		t.Errorf("set same alias, expected length: %d, actual length: %d", 1, len(actual))
	} else if actual[0] != expected {
		t.Errorf("set same alias, expected expected: %v, actual: %v, aliases: %v", expected, actual[0], actual)
	}

	// Test new alias
	expected = "squirrel"
	modified = setAlias(expected, props)
	if !modified {
		t.Errorf("set new alias, modified should be true")
	}
	actual = props.Alias
	if len(actual) != 1 {
		t.Errorf("set new alias, expected length: %d, actual length: %d", 1, len(actual))
	} else if actual[0] != expected {
		t.Errorf("set new alias, expected expected: %v, actual: %v, aliases: %v", expected, actual[0], actual)
	}

	// Test multiple aliases
	expectedAliases := make([]string, 0)
	expectedAliases = append(expectedAliases, "fish")
	expectedAliases = append(expectedAliases, "bird")
	props.Alias = expectedAliases
	expected = "pig"
	modified = setAlias(expected, props)
	if !modified {
		t.Errorf("set alias on multiple aliases, modified should be true")
	}
	actual = props.Alias
	if len(actual) != 2 {
		t.Errorf("set alias, expected length: %d, actual length: %d, aliases: %v", 1, len(actual), actual)
	} else if actual[0] != expected || actual[1] != "bird" {
		t.Errorf("set alias on multiple aliases, expected[0]: %v, actual[0]: %v, expected[1]: bird, actual[1]: %s, aliases: %v", expected, actual[0], actual[1], actual)
	}
}

func TestGetStringFromArray(t *testing.T) {
	var inter interface{}

	// test: nil
	value := getStringFromArray(inter, 0)
	if value != "" {
		t.Errorf("expected an empty string for nil, actual: %s", value)
	}

	// test: empty slice
	inter = make([]string, 0)
	value = getStringFromArray(inter, 0)
	if value != "" {
		t.Errorf("expected an empty string for an empty array, actual: %s", value)
	}

	expected0 := "fish"
	expected1 := "rabbit"
	arr := make([]interface{}, 0)
	arr = append(arr, expected0)
	arr = append(arr, expected1)

	// test: index 0
	value = getStringFromArray(arr, 0)
	if value != expected0 {
		t.Errorf("expected: %s, actual: %s, array: %v", expected0, value, arr)
	}

	// test: index 1
	value = getStringFromArray(arr, 1)
	if value != expected1 {
		t.Errorf("expected: %s, actual: %s, array: %v", expected1, value, arr)
	}

	// test: index past max
	index := 2
	value = getStringFromArray(arr, index)
	if value != "" {
		t.Errorf("expected an empty string for index %d, actual: %s, array: %v", index, value, arr)
	}

	index = -1
	value = getStringFromArray(arr, index)
	if value != "" {
		t.Errorf("expected an empty string for index %d, actual: %s, array: %v", index, value, arr)
	}
}

func TestToString(t *testing.T) {
	var inter interface{}
	value := toString(inter)
	if value != "" {
		t.Errorf("expected empty string, actual: %s", value)
	}

	expected := "cat"
	inter = expected
	value = toString(inter)
	if value != expected {
		t.Errorf("expected %s, actual: %s", expected, value)
	}
}
