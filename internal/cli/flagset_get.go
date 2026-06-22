/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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

package cli

import "fmt"

// This file provides the typed flag readers that mirror pflag's Get* methods.
// Each returns the typed value plus an error when the flag is missing or has a
// different type, matching the signatures the codebase already calls.

// GetString returns the value of a string flag.
func (f *FlagSet) GetString(name string) (string, error) {
	flag := f.Lookup(name)
	if flag == nil {
		return "", fmt.Errorf("flag accessed but not defined: %s", name)
	}
	if v, ok := flag.Value.(*stringValue); ok {
		return string(*v), nil
	}
	return flag.Value.String(), nil
}

// GetBool returns the value of a bool flag.
func (f *FlagSet) GetBool(name string) (bool, error) {
	flag := f.Lookup(name)
	if flag == nil {
		return false, fmt.Errorf("flag accessed but not defined: %s", name)
	}
	v, ok := flag.Value.(*boolValue)
	if !ok {
		return false, fmt.Errorf("trying to get bool value of flag of type %s", flag.Value.Type())
	}
	return bool(*v), nil
}

// GetInt returns the value of an int flag.
func (f *FlagSet) GetInt(name string) (int, error) {
	flag := f.Lookup(name)
	if flag == nil {
		return 0, fmt.Errorf("flag accessed but not defined: %s", name)
	}
	v, ok := flag.Value.(*intValue)
	if !ok {
		return 0, fmt.Errorf("trying to get int value of flag of type %s", flag.Value.Type())
	}
	return int(*v), nil
}

// GetCount returns the value of a count flag.
func (f *FlagSet) GetCount(name string) (int, error) {
	flag := f.Lookup(name)
	if flag == nil {
		return 0, fmt.Errorf("flag accessed but not defined: %s", name)
	}
	v, ok := flag.Value.(*countValue)
	if !ok {
		return 0, fmt.Errorf("trying to get count value of flag of type %s", flag.Value.Type())
	}
	return int(*v), nil
}

// GetStringSlice returns the value of a comma-split string slice flag.
func (f *FlagSet) GetStringSlice(name string) ([]string, error) {
	flag := f.Lookup(name)
	if flag == nil {
		return nil, fmt.Errorf("flag accessed but not defined: %s", name)
	}
	v, ok := flag.Value.(*stringSliceValue)
	if !ok {
		return nil, fmt.Errorf("trying to get stringSlice value of flag of type %s", flag.Value.Type())
	}
	return *v.value, nil
}

// GetStringArray returns the value of a repeatable string array flag.
func (f *FlagSet) GetStringArray(name string) ([]string, error) {
	flag := f.Lookup(name)
	if flag == nil {
		return nil, fmt.Errorf("flag accessed but not defined: %s", name)
	}
	v, ok := flag.Value.(*stringArrayValue)
	if !ok {
		return nil, fmt.Errorf("trying to get stringArray value of flag of type %s", flag.Value.Type())
	}
	return *v.value, nil
}
