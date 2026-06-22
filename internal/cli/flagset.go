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

// FlagSet is an ordered collection of flags addressable by long name and by
// shorthand.  It is the standard-library replacement for pflag.FlagSet.
type FlagSet struct {
	formal    map[string]*Flag
	shorthand map[string]*Flag
	ordered   []*Flag
}

// VarP registers value under name with an optional shorthand and returns the
// created flag so callers can tweak fields such as NoOptDefVal.
func (f *FlagSet) VarP(value Value, name, shorthand, usage string) *Flag {
	flag := &Flag{
		Name:      name,
		Shorthand: shorthand,
		Usage:     usage,
		Value:     value,
		DefValue:  value.String(),
	}
	f.addFlag(flag)
	return flag
}

// addFlag inserts flag into the set's name, shorthand, and ordered indexes.
func (f *FlagSet) addFlag(flag *Flag) {
	if f.formal == nil {
		f.formal = map[string]*Flag{}
		f.shorthand = map[string]*Flag{}
	}
	if _, exists := f.formal[flag.Name]; exists {
		return
	}
	f.formal[flag.Name] = flag
	f.ordered = append(f.ordered, flag)
	if flag.Shorthand != "" {
		f.shorthand[flag.Shorthand] = flag
	}
}

// Lookup returns the flag with the given long name, or nil.
func (f *FlagSet) Lookup(name string) *Flag {
	if f.formal == nil {
		return nil
	}
	return f.formal[name]
}

// shorthandLookup returns the flag with the given single-character shorthand.
func (f *FlagSet) shorthandLookup(name string) *Flag {
	if f.shorthand == nil {
		return nil
	}
	return f.shorthand[name]
}

// Changed reports whether the named flag was set on the command line.
func (f *FlagSet) Changed(name string) bool {
	flag := f.Lookup(name)
	return flag != nil && flag.Changed
}

// MarkHidden hides the named flag from help output.
func (f *FlagSet) MarkHidden(name string) error {
	flag := f.Lookup(name)
	if flag == nil {
		return fmt.Errorf("flag %q does not exist", name)
	}
	flag.Hidden = true
	return nil
}

// AddFlagSet copies the flags from src into f, skipping any whose long name is
// already present.  The flag pointers are shared so that setting a value
// through one set is observable through the other (used for persistent-flag
// inheritance).
func (f *FlagSet) AddFlagSet(src *FlagSet) {
	if src == nil {
		return
	}
	for _, flag := range src.ordered {
		if f.Lookup(flag.Name) == nil {
			f.addFlag(flag)
		}
	}
}

// Set assigns value to the named flag and marks it changed.
func (f *FlagSet) Set(name, value string) error {
	flag := f.Lookup(name)
	if flag == nil {
		return fmt.Errorf("no such flag -%v", name)
	}
	return f.set(flag, value)
}

// set assigns value to flag and marks it changed.
func (f *FlagSet) set(flag *Flag, value string) error {
	if err := flag.Value.Set(value); err != nil {
		return err
	}
	flag.Changed = true
	return nil
}
