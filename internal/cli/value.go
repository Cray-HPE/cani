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

// Package cli is a small, standard-library-only command-line framework that
// replaces the third-party cobra/pflag/viper dependencies.  It provides just
// the subset of features the cani codebase relies on: a command tree with
// inherited (persistent) flags, POSIX-style flag parsing with long/short
// names, slice/array/count flag types, the "changed" partial-update semantics,
// positional-argument validators, and cobra-style help output.
package cli

import (
	"strconv"
	"strings"
)

// Value is the dynamic value stored in a flag.  It mirrors the small subset of
// the pflag.Value interface that the codebase depends on.
type Value interface {
	// String returns the current value formatted for display.
	String() string
	// Set parses s and stores it as the current value.
	Set(s string) error
	// Type returns the value's type name (e.g. "string", "bool").
	Type() string
}

// --- string ---------------------------------------------------------------

type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error { *s = stringValue(val); return nil }
func (s *stringValue) Type() string         { return "string" }
func (s *stringValue) String() string       { return string(*s) }

// --- bool ------------------------------------------------------------------

type boolValue bool

func newBoolValue(val bool, p *bool) *boolValue {
	*p = val
	return (*boolValue)(p)
}

func (b *boolValue) Set(val string) error {
	v, err := strconv.ParseBool(val)
	if err != nil {
		return err
	}
	*b = boolValue(v)
	return nil
}

func (b *boolValue) Type() string   { return "bool" }
func (b *boolValue) String() string { return strconv.FormatBool(bool(*b)) }

// --- int -------------------------------------------------------------------

type intValue int

func newIntValue(val int, p *int) *intValue {
	*p = val
	return (*intValue)(p)
}

func (i *intValue) Set(val string) error {
	v, err := strconv.Atoi(val)
	if err != nil {
		return err
	}
	*i = intValue(v)
	return nil
}

func (i *intValue) Type() string   { return "int" }
func (i *intValue) String() string { return strconv.Itoa(int(*i)) }

// --- count -----------------------------------------------------------------

// countValue increments on every occurrence of the flag (e.g. -V -V or -VV),
// matching pflag's Count behaviour.  An explicit numeric value sets it
// directly (e.g. --verbose=3).
type countValue int

func newCountValue(val int, p *int) *countValue {
	*p = val
	return (*countValue)(p)
}

func (c *countValue) Set(val string) error {
	// "+1" is the NoOptDefVal supplied when the flag appears without an
	// explicit value, meaning "increment".  Any other input sets the count
	// directly (e.g. --verbose=3).
	if val == "+1" {
		*c++
		return nil
	}
	v, err := strconv.Atoi(val)
	if err != nil {
		return err
	}
	*c = countValue(v)
	return nil
}

func (c *countValue) Type() string   { return "count" }
func (c *countValue) String() string { return strconv.Itoa(int(*c)) }

// --- string slice (comma-split) -------------------------------------------

// stringSliceValue splits each Set call on commas.  The default is replaced on
// the first command-line occurrence, then appended on subsequent ones.
type stringSliceValue struct {
	value   *[]string
	changed bool
}

func newStringSliceValue(val []string, p *[]string) *stringSliceValue {
	*p = val
	return &stringSliceValue{value: p}
}

func (s *stringSliceValue) Set(val string) error {
	parts := strings.Split(val, ",")
	if !s.changed {
		*s.value = parts
	} else {
		*s.value = append(*s.value, parts...)
	}
	s.changed = true
	return nil
}

func (s *stringSliceValue) Type() string   { return "stringSlice" }
func (s *stringSliceValue) String() string { return "[" + strings.Join(*s.value, ",") + "]" }

// --- string array (repeatable, no split) ----------------------------------

// stringArrayValue appends each Set call verbatim (no comma splitting).  The
// default is replaced on the first command-line occurrence.
type stringArrayValue struct {
	value   *[]string
	changed bool
}

func newStringArrayValue(val []string, p *[]string) *stringArrayValue {
	*p = val
	return &stringArrayValue{value: p}
}

func (s *stringArrayValue) Set(val string) error {
	if !s.changed {
		*s.value = []string{val}
	} else {
		*s.value = append(*s.value, val)
	}
	s.changed = true
	return nil
}

func (s *stringArrayValue) Type() string   { return "stringArray" }
func (s *stringArrayValue) String() string { return "[" + strings.Join(*s.value, ",") + "]" }
