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
package batch

import (
	"reflect"
	"testing"
)

// TestTokenize verifies shell-style tokenization: single and double quotes
// group spaces, and an unquoted '#' begins a comment to end of line.
//
// Why it matters: the batch runner parses each line of a cani script with this
// tokenizer, so quoted names like "MAN-3502u48" and trailing "# ..." comments in
// the maple inventory script must be handled exactly as a shell would.
// Inputs: representative lines. Outputs: the expected token slices.
// Data choice: cases mirror real maple lines (quoted names, inline comments,
// tabs, blank/comment-only lines) to cover the constructs the format relies on.
func TestTokenize(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"simple", "a b c", []string{"a", "b", "c"}},
		{"double quotes group spaces", `add --name "MAN 48"`, []string{"add", "--name", "MAN 48"}},
		{"single quotes group spaces", `--name 'gpu-%{BAY}'`, []string{"--name", "gpu-%{BAY}"}},
		{"inline comment dropped", "add x # trailing note", []string{"add", "x"}},
		{"comment only", "# whole line", nil},
		{"blank", "   ", nil},
		{"tabs split", "a\tb", []string{"a", "b"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tokenize(tc.in); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("tokenize(%q) = %#v, want %#v", tc.in, got, tc.want)
			}
		})
	}
}

// TestCommandTokens verifies a batch line is reduced to command arguments, with
// non-cani / blank / comment lines reported as skippable.
//
// Why it matters: the runner skips setup-script noise (rm, make, comments) and
// strips the leading program name so the remaining tokens dispatch straight into
// the cani command tree.
// Inputs: representative lines. Outputs: the argument slice and an ok flag.
// Data choice: includes both "cani" and "bin/cani" prefixes, a bare program with
// no args, and shell helpers that must be skipped.
func TestCommandTokens(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
		ok   bool
	}{
		{"bin/cani prefix stripped", "bin/cani alpha add device x", []string{"alpha", "add", "device", "x"}, true},
		{"plain cani prefix stripped", "cani show", []string{"show"}, true},
		{"inline comment stripped", "bin/cani add x # note", []string{"add", "x"}, true},
		{"non-cani rm skipped", "rm -f file", nil, false},
		{"comment skipped", "# note", nil, false},
		{"blank skipped", "", nil, false},
		{"cani without args skipped", "cani", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := commandTokens(tc.in)
			if ok != tc.ok {
				t.Fatalf("commandTokens(%q) ok = %v, want %v", tc.in, ok, tc.ok)
			}
			if ok && !reflect.DeepEqual(got, tc.want) {
				t.Errorf("commandTokens(%q) = %#v, want %#v", tc.in, got, tc.want)
			}
		})
	}
}

// TestIsCaniProgram verifies the program-name check accepts cani with any path
// prefix and rejects other commands.
//
// Why it matters: only cani invocations should be dispatched; other shell lines
// in a reused setup script must be ignored.
// Inputs: candidate first tokens. Outputs: a bool.
// Data choice: covers bare, relative, and absolute cani paths plus common
// non-cani helpers.
func TestIsCaniProgram(t *testing.T) {
	cases := map[string]bool{
		"cani":                true,
		"bin/cani":            true,
		"./cani":              true,
		"/usr/local/bin/cani": true,
		"rm":                  false,
		"make":                false,
		"canister":            false,
	}
	for tok, want := range cases {
		if got := isCaniProgram(tok); got != want {
			t.Errorf("isCaniProgram(%q) = %v, want %v", tok, got, want)
		}
	}
}
