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

// TestParseLongAndShortForms verifies the parser accepts every flag syntax the
// codebase relies on: "--name value", "--name=value", "-s value", "-svalue",
// "-s=value", and combined bool clusters.
//
// Why it matters: these forms are the public CLI contract inherited from
// pflag; any regression would silently change how users pass options to cani.
// Inputs: a flag set with a string flag (shorthand o) and two bool flags
// (a, b), parsed against several argument vectors. Outputs: the stored flag
// values and the leftover positionals. Data choice: each vector isolates one
// syntax so a failure pinpoints the exact form that broke.
func TestParseLongAndShortForms(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"long space", []string{"--out", "x"}, "x"},
		{"long equals", []string{"--out=y"}, "y"},
		{"short space", []string{"-o", "z"}, "z"},
		{"short attached", []string{"-ow"}, "w"},
		{"short equals", []string{"-o=v"}, "v"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := &FlagSet{}
			out := fs.StringP("out", "o", "", "")
			fs.BoolP("a", "a", false, "")
			fs.BoolP("b", "b", false, "")
			if _, _, err := fs.parse(tc.args); err != nil {
				t.Fatalf("parse(%v) error = %v", tc.args, err)
			}
			if *out != tc.want {
				t.Errorf("out = %q, want %q", *out, tc.want)
			}
		})
	}
}

// TestParseCombinedBoolCluster verifies that "-ab" sets both bool flags and
// that an unknown shorthand in a cluster is reported.
//
// Why it matters: cani uses bundled short bools (e.g. add's -a -y); bundling
// must set every flag in the cluster. Inputs: a set with bool flags a and b
// parsed from "-ab", then a cluster with an undefined letter. Outputs: both
// bools true; an error for the undefined letter. Data choice: two adjacent
// bools is the minimal cluster that proves iteration past the first letter.
func TestParseCombinedBoolCluster(t *testing.T) {
	fs := &FlagSet{}
	a := fs.BoolP("a", "a", false, "")
	b := fs.BoolP("b", "b", false, "")
	if _, _, err := fs.parse([]string{"-ab"}); err != nil {
		t.Fatalf("parse error = %v", err)
	}
	if !*a || !*b {
		t.Fatalf("a=%v b=%v, want both true", *a, *b)
	}
	if _, _, err := fs.parse([]string{"-az"}); err == nil {
		t.Error("expected error for unknown shorthand z")
	}
}

// TestParseCountFlag verifies a count flag increments once per occurrence for
// both repeated ("-V -V") and bundled ("-VV") forms and accepts an explicit
// numeric value.
//
// Why it matters: show's -V/-VV verbosity depends on count semantics; getting
// this wrong changes how much detail users see. Inputs: a count flag parsed
// from three vectors. Outputs: counts 2, 2, and 5 respectively. Data choice:
// the bundled and separated forms both yield 2 to prove they are equivalent,
// and "=5" proves explicit assignment overrides incrementing.
func TestParseCountFlag(t *testing.T) {
	cases := []struct {
		args []string
		want int
	}{
		{[]string{"-V", "-V"}, 2},
		{[]string{"-VV"}, 2},
		{[]string{"--verbose=5"}, 5},
	}
	for _, tc := range cases {
		fs := &FlagSet{}
		v := fs.CountP("verbose", "V", "")
		if _, _, err := fs.parse(tc.args); err != nil {
			t.Fatalf("parse(%v) error = %v", tc.args, err)
		}
		if *v != tc.want {
			t.Errorf("parse(%v) count = %d, want %d", tc.args, *v, tc.want)
		}
	}
}

// TestParseSliceVsArray verifies StringSlice comma-splits and StringArray does
// not, and that both replace their default on the first command-line use then
// append afterwards.
//
// Why it matters: cani uses StringSlice for comma lists (--types-dirs) and
// StringArray for repeatable raw values (--tag, --metadata); confusing the two
// would corrupt user input containing commas. Inputs: a slice flag and array
// flag each given two occurrences. Outputs: the slice splits and accumulates
// to four items; the array keeps two verbatim. Data choice: a value containing
// a comma ("x,y") distinguishes splitting from literal storage.
func TestParseSliceVsArray(t *testing.T) {
	fs := &FlagSet{}
	slice := fs.StringSlice("slice", []string{"default"}, "")
	array := fs.StringArray("array", []string{"default"}, "")
	args := []string{"--slice", "x,y", "--slice", "z", "--array", "x,y", "--array", "z"}
	if _, _, err := fs.parse(args); err != nil {
		t.Fatalf("parse error = %v", err)
	}
	if got := *slice; len(got) != 3 || got[0] != "x" || got[1] != "y" || got[2] != "z" {
		t.Errorf("slice = %v, want [x y z]", got)
	}
	if got := *array; len(got) != 2 || got[0] != "x,y" || got[1] != "z" {
		t.Errorf("array = %v, want [x,y z]", got)
	}
}

// TestChangedSemantics verifies a flag reports Changed only after being set on
// the command line, preserving the set-vs-default distinction.
//
// Why it matters: cani's update commands apply only flags the user explicitly
// set (cmd.Flags().Changed); a false positive would overwrite fields with zero
// values. Inputs: a string flag read before and after parsing "--name v".
// Outputs: Changed is false initially and true afterward, with the default
// retained until set. Data choice: a non-empty default ("def") proves the
// value is the default, not a coincidental zero value, before the flag is set.
func TestChangedSemantics(t *testing.T) {
	fs := &FlagSet{}
	name := fs.String("name", "def", "")
	if fs.Changed("name") {
		t.Error("Changed should be false before parsing")
	}
	if *name != "def" {
		t.Errorf("default = %q, want def", *name)
	}
	if _, _, err := fs.parse([]string{"--name", "v"}); err != nil {
		t.Fatalf("parse error = %v", err)
	}
	if !fs.Changed("name") {
		t.Error("Changed should be true after parsing")
	}
}

// TestParseErrors verifies the parser rejects unknown flags and flags missing
// a required argument.
//
// Why it matters: clear errors for bad input are part of the CLI contract;
// silently ignoring an unknown flag could mask user mistakes. Inputs: an empty
// set given "--nope", and a value flag given no argument. Outputs: a non-nil
// error in both cases. Data choice: "--need" with no following token is the
// minimal case that exercises the end-of-args branch.
func TestParseErrors(t *testing.T) {
	fs := &FlagSet{}
	fs.String("need", "", "")
	if _, _, err := fs.parse([]string{"--nope"}); err == nil {
		t.Error("expected unknown-flag error")
	}
	if _, _, err := fs.parse([]string{"--need"}); err == nil {
		t.Error("expected needs-argument error")
	}
}

// TestParseDoubleDashTerminator verifies that "--" stops flag parsing and
// treats the rest of the arguments as positionals.
//
// Why it matters: the "--" terminator lets users pass values that look like
// flags; dropping it would misparse such arguments. Inputs: arguments with a
// flag, "--", then a flag-like positional. Outputs: the flag is set and the
// flag-like token is returned as a positional. Data choice: a positional that
// begins with "--" proves it is not interpreted as a flag after the terminator.
func TestParseDoubleDashTerminator(t *testing.T) {
	fs := &FlagSet{}
	flag := fs.Bool("flag", false, "")
	pos, _, err := fs.parse([]string{"--flag", "--", "--not-a-flag", "x"})
	if err != nil {
		t.Fatalf("parse error = %v", err)
	}
	if !*flag {
		t.Error("--flag should be set")
	}
	if len(pos) != 2 || pos[0] != "--not-a-flag" || pos[1] != "x" {
		t.Errorf("positionals = %v, want [--not-a-flag x]", pos)
	}
}
