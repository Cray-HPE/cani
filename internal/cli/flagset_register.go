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

// This file provides the typed flag-registration helpers that mirror the
// subset of the pflag API the codebase uses.  Every "P" variant takes a
// shorthand; the non-P variant delegates with an empty shorthand.  Every
// "Var" variant binds an existing pointer instead of allocating one.

// --- string ---------------------------------------------------------------

// StringVarP binds an existing *string to a flag with a shorthand.
func (f *FlagSet) StringVarP(p *string, name, shorthand, value, usage string) {
	f.VarP(newStringValue(value, p), name, shorthand, usage)
}

// StringVar binds an existing *string to a flag.
func (f *FlagSet) StringVar(p *string, name, value, usage string) {
	f.StringVarP(p, name, "", value, usage)
}

// StringP defines a string flag with a shorthand and returns its pointer.
func (f *FlagSet) StringP(name, shorthand, value, usage string) *string {
	p := new(string)
	f.StringVarP(p, name, shorthand, value, usage)
	return p
}

// String defines a string flag and returns its pointer.
func (f *FlagSet) String(name, value, usage string) *string {
	return f.StringP(name, "", value, usage)
}

// --- bool ------------------------------------------------------------------

// BoolVarP binds an existing *bool to a flag with a shorthand.
func (f *FlagSet) BoolVarP(p *bool, name, shorthand string, value bool, usage string) {
	flag := f.VarP(newBoolValue(value, p), name, shorthand, usage)
	flag.NoOptDefVal = "true"
}

// BoolVar binds an existing *bool to a flag.
func (f *FlagSet) BoolVar(p *bool, name string, value bool, usage string) {
	f.BoolVarP(p, name, "", value, usage)
}

// BoolP defines a bool flag with a shorthand and returns its pointer.
func (f *FlagSet) BoolP(name, shorthand string, value bool, usage string) *bool {
	p := new(bool)
	f.BoolVarP(p, name, shorthand, value, usage)
	return p
}

// Bool defines a bool flag and returns its pointer.
func (f *FlagSet) Bool(name string, value bool, usage string) *bool {
	return f.BoolP(name, "", value, usage)
}

// --- int -------------------------------------------------------------------

// IntVarP binds an existing *int to a flag with a shorthand.
func (f *FlagSet) IntVarP(p *int, name, shorthand string, value int, usage string) {
	f.VarP(newIntValue(value, p), name, shorthand, usage)
}

// IntVar binds an existing *int to a flag.
func (f *FlagSet) IntVar(p *int, name string, value int, usage string) {
	f.IntVarP(p, name, "", value, usage)
}

// IntP defines an int flag with a shorthand and returns its pointer.
func (f *FlagSet) IntP(name, shorthand string, value int, usage string) *int {
	p := new(int)
	f.IntVarP(p, name, shorthand, value, usage)
	return p
}

// Int defines an int flag and returns its pointer.
func (f *FlagSet) Int(name string, value int, usage string) *int {
	return f.IntP(name, "", value, usage)
}

// --- count -----------------------------------------------------------------

// CountP defines a count flag with a shorthand and returns its pointer.
func (f *FlagSet) CountP(name, shorthand, usage string) *int {
	p := new(int)
	flag := f.VarP(newCountValue(0, p), name, shorthand, usage)
	flag.NoOptDefVal = "+1"
	return p
}

// Count defines a count flag and returns its pointer.
func (f *FlagSet) Count(name, usage string) *int {
	return f.CountP(name, "", usage)
}

// --- string slice (comma-split) -------------------------------------------

// StringSlice defines a comma-split string slice flag and returns its pointer.
func (f *FlagSet) StringSlice(name string, value []string, usage string) *[]string {
	p := new([]string)
	f.VarP(newStringSliceValue(value, p), name, "", usage)
	return p
}

// --- string array (repeatable, no split) ----------------------------------

// StringArray defines a repeatable string array flag and returns its pointer.
func (f *FlagSet) StringArray(name string, value []string, usage string) *[]string {
	p := new([]string)
	f.VarP(newStringArrayValue(value, p), name, "", usage)
	return p
}
