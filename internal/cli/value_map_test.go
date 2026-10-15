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

import (
	"maps"
	"strings"
	"testing"
)

// TestStringToStringCollectsPairs verifies the repeatable key=value flag
// keeps every pair, splits on the first "=" only, and lets a later pair
// overwrite an earlier key.
//
// Why it matters: providers receive these pairs verbatim as the generic CLI
// input for their own metadata sections; a value that itself contains "="
// must survive, and "last wins" is the behaviour callers already relied on.
// Inputs: "--metadata owner=ops --metadata note=a=b --metadata owner=net".
// Outputs: GetStringToString returns {owner: net, note: a=b}; String renders
// the pairs sorted by key; Type reports stringToString for help output.
// Data choice: "note=a=b" is the shortest value that distinguishes a
// first-"=" split from a last-"=" split.
func TestStringToStringCollectsPairs(t *testing.T) {
	// Arrange.
	fs := &FlagSet{}
	fs.StringToString("metadata", nil, "")

	// Act.
	_, _, err := fs.parse([]string{"--metadata", "owner=ops", "--metadata", "note=a=b", "--metadata", "owner=net"})

	// Assert.
	if err != nil {
		t.Fatalf("parse: unexpected error: %v", err)
	}
	got, err := fs.GetStringToString("metadata")
	if err != nil {
		t.Fatalf("GetStringToString: unexpected error: %v", err)
	}
	want := map[string]string{"owner": "net", "note": "a=b"}
	if !maps.Equal(got, want) {
		t.Errorf("pairs = %v, want %v", got, want)
	}
	flag := fs.Lookup("metadata")
	if s := flag.Value.String(); s != "[note=a=b,owner=net]" {
		t.Errorf("String() = %q, want %q", s, "[note=a=b,owner=net]")
	}
	if typ := flag.Value.Type(); typ != "stringToString" {
		t.Errorf("Type() = %q, want %q", typ, "stringToString")
	}
}

// TestStringToStringDefaultAndUnset verifies the default map is returned until
// the flag is set, and that an unset flag with no default reads as nil.
//
// Why it matters: callers test len(pairs) to decide whether anything was
// given; a non-nil empty default would make "nothing given" indistinguishable
// from "given nothing", and the first command-line pair must replace the
// default rather than extend it.
// Inputs: one flag with a default pair and one without; the former is then
// set once.
// Outputs: nil for the unset flag; the default before parsing and only the
// parsed pair after it for the other.
// Data choice: a single default pair is enough to show replacement.
func TestStringToStringDefaultAndUnset(t *testing.T) {
	// Arrange.
	fs := &FlagSet{}
	fs.StringToString("metadata", nil, "")
	fs.StringToString("labels", map[string]string{"tier": "gold"}, "")

	// Act.
	unset, _ := fs.GetStringToString("metadata")
	before, _ := fs.GetStringToString("labels")
	_, _, err := fs.parse([]string{"--labels", "env=prod"})
	after, _ := fs.GetStringToString("labels")

	// Assert.
	if unset != nil {
		t.Errorf("unset flag = %v, want nil", unset)
	}
	if !maps.Equal(before, map[string]string{"tier": "gold"}) {
		t.Errorf("default = %v, want tier=gold", before)
	}
	if err != nil {
		t.Fatalf("parse: unexpected error: %v", err)
	}
	if !maps.Equal(after, map[string]string{"env": "prod"}) {
		t.Errorf("after parse = %v, want the default replaced by env=prod", after)
	}
}

// TestStringToStringRejectsMalformedPairs verifies a pair without a key=value
// shape fails at parse time, naming the flag and the offending value.
//
// Why it matters: before this type existed each command re-checked the shape
// after parsing and one of them dropped every pair silently on a bad one.
// Inputs: a value without "=", an empty key, a bare "=", and a key that is
// itself a flag.
// Outputs: an error per case that names "--metadata" and the value; nothing
// is recorded; the key-shaped-like-a-flag case says so.
// Data choice: "--status=Active" is the reported mistake in its "=" form,
// which the parser's own flag-token guard does not see.
func TestStringToStringRejectsMalformedPairs(t *testing.T) {
	cases := []struct {
		pair string
		want string
	}{
		{"bogus", "expected key=value"},
		{"=orphan", "expected key=value"},
		{"=", "expected key=value"},
		{"--status=Active", `key "--status" looks like a flag`},
	}
	for _, tc := range cases {
		t.Run(tc.pair, func(t *testing.T) { checkRejectedPair(t, tc.pair, tc.want) })
	}
}

// checkRejectedPair parses one malformed --metadata pair and checks the error
// shape and that nothing was recorded.
func checkRejectedPair(t *testing.T, pair, want string) {
	t.Helper()
	// Arrange.
	fs := &FlagSet{}
	fs.StringToString("metadata", nil, "")

	// Act.
	_, _, err := fs.parse([]string{"--metadata=" + pair})

	// Assert.
	if err == nil {
		t.Fatalf("parse(%q): expected an error, got nil", pair)
	}
	prefix := `invalid argument "` + pair + `" for --metadata: `
	if !strings.HasPrefix(err.Error(), prefix) || !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want prefix %q and %q", err, prefix, want)
	}
	if got, _ := fs.GetStringToString("metadata"); len(got) != 0 {
		t.Errorf("rejected value was recorded: %v", got)
	}
}

// TestGetStringToStringRejectsOtherTypes verifies the getter refuses a flag of
// another type instead of returning an empty map.
//
// Why it matters: GetStringArray on a key=value flag (or the reverse) is a
// programming error that must surface, not read as "no pairs given".
// Inputs: a string array flag and an undefined name.
// Outputs: an error naming the actual type, and the undefined-flag error.
// Data choice: stringArray is the type --metadata used to have.
func TestGetStringToStringRejectsOtherTypes(t *testing.T) {
	// Arrange.
	fs := &FlagSet{}
	fs.StringArray("tag", nil, "")

	// Act.
	_, typeErr := fs.GetStringToString("tag")
	_, missingErr := fs.GetStringToString("nope")

	// Assert.
	if typeErr == nil || !strings.Contains(typeErr.Error(), "of type stringArray") {
		t.Errorf("type error = %v, want one naming stringArray", typeErr)
	}
	if missingErr == nil || !strings.Contains(missingErr.Error(), "not defined") {
		t.Errorf("missing error = %v, want the undefined-flag error", missingErr)
	}
}
