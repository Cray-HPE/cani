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
	"errors"
	"strconv"
	"strings"
	"testing"
)

// TestSetErrorNamesTheFlag verifies that a value a flag refuses is reported
// together with the flag it was given to, on both parse paths.
//
// Why it matters: a Value only sees its input, so without this wrapping an
// int flag fed "abc" surfaced as a bare "strconv.Atoi: parsing ..." line and
// the user had to guess which of a dozen flags was meant.
// Inputs: an int flag "position" parsed as "--position abc", "--position=abc"
// and "-p abc", plus a direct FlagSet.Set call.
// Outputs: every error names the value and "--position" and still wraps the
// strconv error so callers can inspect the cause.
// Data choice: "abc" is the shortest input strconv rejects.
func TestSetErrorNamesTheFlag(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"long space", []string{"--position", "abc"}},
		{"long equals", []string{"--position=abc"}},
		{"short space", []string{"-p", "abc"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			fs := &FlagSet{}
			fs.IntP("position", "p", 0, "")

			// Act.
			_, _, err := fs.parse(tc.args)

			// Assert.
			assertNamedValueError(t, err, "abc", "position")
		})
	}

	t.Run("direct Set", func(t *testing.T) {
		fs := &FlagSet{}
		fs.Int("position", 0, "")
		assertNamedValueError(t, fs.Set("position", "abc"), "abc", "position")
	})
}

// assertNamedValueError checks the wrapped error shape produced by set.
func assertNamedValueError(t *testing.T, err error, value, flag string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	want := `invalid argument "` + value + `" for --` + flag + `: `
	if !strings.HasPrefix(err.Error(), want) {
		t.Errorf("error = %q, want prefix %q", err, want)
	}
	var numErr *strconv.NumError
	if !errors.As(err, &numErr) {
		t.Errorf("error %q does not wrap the strconv cause", err)
	}
}

// TestSetErrorLeavesFlagUnchanged verifies a rejected value does not mark the
// flag as set.
//
// Why it matters: callers use Changed to decide whether to apply a partial
// update; a refused value must not look like a user-supplied one.
// Inputs: an int flag fed "abc" through FlagSet.Set.
// Outputs: Set returns an error and Changed reports false.
// Data choice: a single flag isolates the Changed bookkeeping from parsing.
func TestSetErrorLeavesFlagUnchanged(t *testing.T) {
	// Arrange.
	fs := &FlagSet{}
	fs.Int("position", 0, "")

	// Act.
	err := fs.Set("position", "abc")

	// Assert.
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if fs.Changed("position") {
		t.Error("Changed(position) = true, want false after a rejected value")
	}
}
