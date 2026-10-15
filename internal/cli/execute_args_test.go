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
	"bytes"
	"os"
	"slices"
	"strings"
	"testing"
)

// TestExecuteUsesSetArgs verifies that Execute parses the arguments given to
// SetArgs through the real parser, resolving subcommands and flags, and never
// looks at os.Args once they are set.
//
// Why it matters: tests that drive a command tree end to end need the parser
// in the loop (flag values, positional arity); mutating os.Args for that is a
// process-global side effect this seam removes.
// Inputs: a root -> sub tree with a persistent string flag; SetArgs on the
// root with a subcommand name, the flag and two positionals, while os.Args
// holds a value that would fail if it were parsed.
// Outputs: the subcommand's RunE sees the two positionals and the flag value.
// Data choice: a poisoned os.Args proves the override, not just the parse.
func TestExecuteUsesSetArgs(t *testing.T) {
	// Arrange.
	var gotArgs []string
	var gotName string
	root := &Command{Use: "root"}
	root.PersistentFlags().String("name", "", "")
	sub := &Command{
		Use: "sub",
		RunE: func(cmd *Command, args []string) error {
			gotArgs = args
			gotName, _ = cmd.Flags().GetString("name")
			return nil
		},
	}
	root.AddCommand(sub)
	saved := os.Args
	os.Args = []string{"root", "--no-such-flag"}
	t.Cleanup(func() { os.Args = saved })

	// Act.
	root.SetArgs([]string{"sub", "--name", "x", "a", "b"})
	err := root.Execute()

	// Assert.
	if err != nil {
		t.Fatalf("Execute: unexpected error: %v", err)
	}
	if !slices.Equal(gotArgs, []string{"a", "b"}) {
		t.Errorf("positionals = %v, want [a b]", gotArgs)
	}
	if gotName != "x" {
		t.Errorf("--name = %q, want %q", gotName, "x")
	}
}

// TestExecuteWithEmptySetArgsRunsRoot verifies that SetArgs with an empty
// slice means "no arguments" rather than "fall back to os.Args".
//
// Why it matters: a test for a root command's no-argument behaviour must not
// accidentally pick up the test binary's own flags from os.Args.
// Inputs: a root whose RunE records that it ran; SetArgs(nil-length slice)
// while os.Args carries a token the root does not accept.
// Outputs: RunE runs with no positionals and no error.
// Data choice: the poisoned os.Args again distinguishes the two meanings.
func TestExecuteWithEmptySetArgsRunsRoot(t *testing.T) {
	// Arrange.
	ran := false
	root := &Command{Use: "root", Args: NoArgs, RunE: func(cmd *Command, args []string) error {
		ran = len(args) == 0
		return nil
	}}
	saved := os.Args
	os.Args = []string{"root", "stray"}
	t.Cleanup(func() { os.Args = saved })

	// Act.
	root.SetArgs([]string{})
	err := root.Execute()

	// Assert.
	if err != nil {
		t.Fatalf("Execute: unexpected error: %v", err)
	}
	if !ran {
		t.Error("root RunE did not run with zero positionals")
	}
}

// TestExecuteReportsParseErrorsOnStderr verifies a parse failure under SetArgs
// is returned and printed with usage on the command's error writer.
//
// Why it matters: callers of the test seam assert on the error and on what the
// user would have seen; both must come from the same run.
// Inputs: a root with an int flag given "abc".
// Outputs: the returned error and the stderr text both name the flag.
// Data choice: an int flag exercises the Value.Set wrapping end to end.
func TestExecuteReportsParseErrorsOnStderr(t *testing.T) {
	// Arrange.
	var stderr bytes.Buffer
	root := &Command{Use: "root", RunE: func(*Command, []string) error { return nil }}
	root.Flags().Int("position", 0, "")
	root.SetErr(&stderr)

	// Act.
	root.SetArgs([]string{"--position", "abc"})
	err := root.Execute()

	// Assert.
	if err == nil || !strings.Contains(err.Error(), `invalid argument "abc" for --position`) {
		t.Fatalf("Execute error = %v, want the named value error", err)
	}
	if !strings.Contains(stderr.String(), `Error: invalid argument "abc" for --position`) {
		t.Errorf("stderr = %q, want the error line", stderr.String())
	}
}
