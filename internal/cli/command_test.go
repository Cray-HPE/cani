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
	"strings"
	"testing"
)

// silence redirects a command's output and error streams to buffers so tests
// do not print usage/help noise.
func silence(c *Command) {
	c.SetOut(&bytes.Buffer{})
	c.SetErr(&bytes.Buffer{})
}

// TestFindResolvesNestedSubcommands verifies the tree walker descends through
// multiple subcommand levels and returns the leftover flag/positional tokens.
//
// Why it matters: cani nests commands several levels deep (e.g. alpha show
// rack); the walker must reach the leaf so the right RunE executes. Inputs: a
// root→mid→leaf tree resolved against "mid leaf pos". Outputs: the leaf command
// and the positional "pos". Data choice: a trailing positional confirms only
// subcommand names are consumed, not arguments.
func TestFindResolvesNestedSubcommands(t *testing.T) {
	root := &Command{Use: "root"}
	mid := &Command{Use: "mid"}
	leaf := &Command{Use: "leaf"}
	root.AddCommand(mid)
	mid.AddCommand(leaf)

	got, rest := root.find([]string{"mid", "leaf", "pos"})
	if got != leaf {
		t.Fatalf("find resolved %q, want leaf", got.Name())
	}
	if len(rest) != 1 || rest[0] != "pos" {
		t.Errorf("rest = %v, want [pos]", rest)
	}
}

// TestFindSkipsInterspersedFlags verifies the walker still finds a subcommand
// when a value-taking persistent flag precedes the subcommand name.
//
// Why it matters: users may write "import --phase et csm"; the walker must not
// mistake the flag value "et" for a subcommand. Inputs: an import command with
// a persistent string flag "phase" and a child "csm", resolved against
// "--phase et csm". Outputs: the csm command with "--phase et" remaining. Data
// choice: placing the flag before the subcommand is the exact ordering that
// breaks naive splitting.
func TestFindSkipsInterspersedFlags(t *testing.T) {
	root := &Command{Use: "root"}
	imp := &Command{Use: "import"}
	imp.PersistentFlags().String("phase", "etl", "")
	csm := &Command{Use: "csm"}
	imp.AddCommand(csm)
	root.AddCommand(imp)

	got, rest := root.find([]string{"import", "--phase", "et", "csm"})
	if got != csm {
		t.Fatalf("find resolved %q, want csm", got.Name())
	}
	if strings.Join(rest, " ") != "--phase et" {
		t.Errorf("rest = %v, want [--phase et]", rest)
	}
}

// TestPersistentFlagInheritance verifies a child command can read a flag
// defined as persistent on an ancestor after merging.
//
// Why it matters: cani leaf commands read root/parent persistent flags (e.g.
// datastore-path, format) via cmd.Flags().Get*; inheritance must surface those
// flags on the child. Inputs: a parent with persistent string flag "format"
// and a child run with "--format json". Outputs: the child reads "json". Data
// choice: setting the value on the child invocation proves the shared flag
// pointer is updated, not just present.
func TestPersistentFlagInheritance(t *testing.T) {
	root := &Command{Use: "root"}
	parent := &Command{Use: "show"}
	parent.PersistentFlags().StringP("format", "o", "table", "")
	var got string
	child := &Command{
		Use: "rack",
		RunE: func(cmd *Command, args []string) error {
			got, _ = cmd.Flags().GetString("format")
			return nil
		},
	}
	parent.AddCommand(child)
	root.AddCommand(parent)
	silence(root)

	target, rest := root.find([]string{"show", "rack", "--format", "json"})
	if err := target.run(rest); err != nil {
		t.Fatalf("run error = %v", err)
	}
	if got != "json" {
		t.Errorf("inherited format = %q, want json", got)
	}
}

// TestRunPipelineOrder verifies the execution order: nearest PersistentPreRunE,
// then Args validation, then PreRunE, then RunE.
//
// Why it matters: cani's root PersistentPreRunE loads config/providers before
// any command logic; running hooks out of order would execute commands against
// uninitialised state. Inputs: a root with PersistentPreRunE and a child with
// Args, PreRunE, and RunE that append labels to a slice. Outputs: the labels in
// the order persistent, args, prerun, run. Data choice: a shared order slice
// records actual invocation sequence rather than asserting each in isolation.
func TestRunPipelineOrder(t *testing.T) {
	var order []string
	root := &Command{
		Use:               "root",
		PersistentPreRunE: func(c *Command, a []string) error { order = append(order, "persistent"); return nil },
	}
	child := &Command{
		Use:     "child",
		Args:    func(c *Command, a []string) error { order = append(order, "args"); return nil },
		PreRunE: func(c *Command, a []string) error { order = append(order, "prerun"); return nil },
		RunE:    func(c *Command, a []string) error { order = append(order, "run"); return nil },
	}
	root.AddCommand(child)
	silence(root)

	target, rest := root.find([]string{"child"})
	if err := target.run(rest); err != nil {
		t.Fatalf("run error = %v", err)
	}
	want := "persistent args prerun run"
	if strings.Join(order, " ") != want {
		t.Errorf("order = %v, want %q", order, want)
	}
}

// TestArgsValidatorRejects verifies a failing Args validator stops execution
// before RunE.
//
// Why it matters: commands such as remove require exactly one argument; the
// validator must block RunE when the count is wrong to avoid acting on bad
// input. Inputs: a child with ExactArgs(1) and a RunE that records execution,
// run with zero arguments. Outputs: a non-nil error and RunE never invoked.
// Data choice: zero args against ExactArgs(1) is the minimal violating case.
func TestArgsValidatorRejects(t *testing.T) {
	ran := false
	root := &Command{Use: "root"}
	child := &Command{
		Use:  "rm",
		Args: ExactArgs(1),
		RunE: func(c *Command, a []string) error { ran = true; return nil },
	}
	root.AddCommand(child)
	silence(root)

	target, rest := root.find([]string{"rm"})
	if err := target.run(rest); err == nil {
		t.Error("expected error from ExactArgs(1) with no args")
	}
	if ran {
		t.Error("RunE should not run when Args validation fails")
	}
}

// TestRequiredTogetherValidation verifies MarkFlagsRequiredTogether errors when
// only some flags in a group are set and passes when all or none are set.
//
// Why it matters: the CSM provider requires username, password, and host
// together; a partial set must fail fast rather than attempt a broken auth.
// Inputs: a command with a three-flag group, run with one flag set and again
// with all three set. Outputs: an error for the partial case, success for the
// complete case. Data choice: setting exactly one of three is the clearest
// violation of "all or nothing".
func TestRequiredTogetherValidation(t *testing.T) {
	newCmd := func() *Command {
		root := &Command{Use: "root"}
		c := &Command{Use: "auth", RunE: func(c *Command, a []string) error { return nil }}
		c.Flags().String("user", "", "")
		c.Flags().String("pass", "", "")
		c.Flags().String("host", "", "")
		c.MarkFlagsRequiredTogether("user", "pass", "host")
		root.AddCommand(c)
		silence(root)
		return root
	}

	root := newCmd()
	target, rest := root.find([]string{"auth", "--user", "bob"})
	if err := target.run(rest); err == nil {
		t.Error("expected error when only one of the group is set")
	}

	root = newCmd()
	target, rest = root.find([]string{"auth", "--user", "bob", "--pass", "x", "--host", "h"})
	if err := target.run(rest); err != nil {
		t.Errorf("unexpected error when all set: %v", err)
	}
}

// TestAliasResolution verifies a subcommand is reachable through one of its
// aliases.
//
// Why it matters: cani exposes plural aliases (e.g. vlans→vlan); alias lookup
// must resolve to the same command. Inputs: a child named "vlan" with alias
// "vlans", resolved by the alias. Outputs: the vlan command. Data choice: a
// single alias is enough to prove alias matching is consulted alongside names.
func TestAliasResolution(t *testing.T) {
	root := &Command{Use: "root"}
	vlan := &Command{Use: "vlan", Aliases: []string{"vlans"}}
	root.AddCommand(vlan)

	got := root.findChild("vlans")
	if got != vlan {
		t.Errorf("findChild(vlans) = %v, want vlan", got)
	}
}

// TestVersionFlag verifies the root command prints its version when --version
// is supplied and Version is set.
//
// Why it matters: cani sets root.Version and users rely on "cani --version";
// the auto-registered flag must short-circuit before running any command.
// Inputs: a root with Version "1.2.3" run with "--version". Outputs: stdout
// contains "1.2.3". Data choice: a recognisable semantic version string makes
// the assertion unambiguous.
func TestVersionFlag(t *testing.T) {
	var out bytes.Buffer
	root := &Command{Use: "cani", Version: "1.2.3", RunE: func(c *Command, a []string) error { return nil }}
	root.SetOut(&out)
	root.SetErr(&bytes.Buffer{})

	target, rest := root.find([]string{"--version"})
	if err := target.run(rest); err != nil {
		t.Fatalf("run error = %v", err)
	}
	if !strings.Contains(out.String(), "1.2.3") {
		t.Errorf("version output = %q, want it to contain 1.2.3", out.String())
	}
}

// TestHelpListsSubcommandsAndFlags verifies usage output includes available
// commands and local flags.
//
// Why it matters: help is the primary discovery mechanism for the CLI; missing
// commands or flags in usage would degrade usability. Inputs: a parent with a
// child and a local string flag. Outputs: usage text containing both the child
// name and the flag name. Data choice: asserting on substrings keeps the test
// robust to exact spacing while still proving the sections render.
func TestHelpListsSubcommandsAndFlags(t *testing.T) {
	root := &Command{Use: "root", Short: "root cmd"}
	child := &Command{Use: "child", Short: "a child"}
	root.AddCommand(child)
	root.Flags().String("thing", "", "a thing flag")

	usage := root.UsageString()
	if !strings.Contains(usage, "child") {
		t.Errorf("usage missing subcommand: %q", usage)
	}
	if !strings.Contains(usage, "--thing") {
		t.Errorf("usage missing flag: %q", usage)
	}
}
