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

import "testing"

// TestParseRefusesFlagTokenAsValue verifies that a flag written without its
// value does not swallow the flag that follows it.
//
// Why it matters: "--metadata --status Active" used to store "--status" as
// the metadata value and leave "Active" as a stray positional, so the status
// was silently lost; the same happened to every value flag.
// Inputs: a string flag "name" (shorthand n) and a bool flag "status"
// (shorthand s); "name" is given the next token in long, equals and short
// forms, and the built-in help token.
// Outputs: each parse fails with the "flag needs an argument" message naming
// the flag that lacked a value, and the flag stays unchanged.
// Data choice: both long and short spellings of the swallowed flag are
// covered because they take different code paths in the parser.
func TestParseRefusesFlagTokenAsValue(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"long then long", []string{"--name", "--status"}, "flag needs an argument: --name"},
		{"long then long equals", []string{"--name", "--status=true"}, "flag needs an argument: --name"},
		{"long then short", []string{"--name", "-s"}, "flag needs an argument: --name"},
		{"long then help", []string{"--name", "--help"}, "flag needs an argument: --name"},
		{"short then long", []string{"-n", "--status"}, "flag needs an argument: -n"},
		{"short then short help", []string{"-n", "-h"}, "flag needs an argument: -n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			fs := &FlagSet{}
			fs.StringP("name", "n", "", "")
			fs.BoolP("status", "s", false, "")

			// Act.
			_, _, err := fs.parse(tc.args)

			// Assert.
			if err == nil || err.Error() != tc.want {
				t.Fatalf("parse(%v) error = %v, want %q", tc.args, err, tc.want)
			}
			if fs.Changed("name") {
				t.Error("Changed(name) = true, want false")
			}
		})
	}
}

// TestParseAcceptsDashValuesThatAreNotFlags verifies the guard only reacts to
// tokens that name a flag in the set.
//
// Why it matters: values legitimately start with "-" (negative numbers, names
// chosen by the user); refusing them would make such values unspellable.
// Inputs: a string flag "name" and an int flag "offset" given "-1", "-x",
// "--not-a-flag" and the bare "--" token as values.
// Outputs: each value is stored verbatim.
// Data choice: "-x" is one character, like a shorthand, but is not
// registered, which pins the lookup to registered flags only.
func TestParseAcceptsDashValuesThatAreNotFlags(t *testing.T) {
	cases := []struct {
		name  string
		args  []string
		check func(name string, offset int) bool
	}{
		{"negative int", []string{"--offset", "-1"}, func(_ string, o int) bool { return o == -1 }},
		{"unregistered short", []string{"--name", "-x"}, func(n string, _ int) bool { return n == "-x" }},
		{"unregistered long", []string{"--name", "--not-a-flag"}, func(n string, _ int) bool { return n == "--not-a-flag" }},
		{"bare double dash", []string{"--name", "--"}, func(n string, _ int) bool { return n == "--" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			fs := &FlagSet{}
			name := fs.String("name", "", "")
			offset := fs.Int("offset", 0, "")

			// Act.
			_, _, err := fs.parse(tc.args)

			// Assert.
			if err != nil {
				t.Fatalf("parse(%v) error = %v", tc.args, err)
			}
			if !tc.check(*name, *offset) {
				t.Errorf("parse(%v) stored name=%q offset=%d", tc.args, *name, *offset)
			}
		})
	}
}
