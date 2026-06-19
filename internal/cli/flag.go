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

// Flag describes a single command-line flag.  The exported fields mirror the
// pflag.Flag fields the codebase reads (Name, Shorthand, Value, etc.).
type Flag struct {
	// Name is the long flag name (without the leading "--").
	Name string
	// Shorthand is the optional single-character short name (without "-").
	Shorthand string
	// Usage is the help text for the flag.
	Usage string
	// Value holds the flag's typed value.
	Value Value
	// DefValue is the string form of the default, used in help output.
	DefValue string
	// Changed reports whether the flag was set on the command line.
	Changed bool
	// Hidden hides the flag from help output when true.
	Hidden bool
	// NoOptDefVal is the value assigned when the flag is supplied without an
	// explicit argument.  Bool flags use "true" and count flags use "+1";
	// flags that require an argument leave this empty.
	NoOptDefVal string
}
