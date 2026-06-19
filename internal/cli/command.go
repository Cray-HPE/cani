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
	"context"
	"io"
	"os"
	"strings"
)

// RunFunc is the signature for a command's run and pre-run hooks.
type RunFunc func(cmd *Command, args []string) error

// Command is a single node in the command tree.  Its exported fields mirror
// the subset of cobra.Command the codebase configures declaratively.
type Command struct {
	// Use is the one-line usage string; its first word is the command name.
	Use string
	// Short is the short description shown in the parent's command list.
	Short string
	// Long is the long description shown in the command's own help.
	Long string
	// Aliases are alternative names that resolve to this command.
	Aliases []string
	// Version, when set on the root command, enables the --version flag.
	Version string
	// Args validates positional arguments before the run hooks execute.
	Args PositionalArgs
	// ValidArgs is retained for source compatibility with callers that copy
	// the field (e.g. utils.CloneCommand); the framework does not act on it.
	ValidArgs []string

	// PersistentPreRunE runs before this command and every descendant.  Only
	// the nearest non-nil hook walking up from the target executes.
	PersistentPreRunE RunFunc
	// PreRunE runs after argument validation and before RunE.
	PreRunE RunFunc
	// RunE is the command's main action.
	RunE RunFunc

	parent   *Command
	commands []*Command
	ctx      context.Context

	lflags *FlagSet
	pflags *FlagSet

	flagGroupsRequiredTogether [][]string

	outWriter io.Writer
	errWriter io.Writer
	inReader  io.Reader
}

// AddCommand attaches one or more subcommands to c.
func (c *Command) AddCommand(cmds ...*Command) {
	for _, sub := range cmds {
		sub.parent = c
		c.commands = append(c.commands, sub)
	}
}

// Commands returns the direct subcommands of c.
func (c *Command) Commands() []*Command {
	return c.commands
}

// Parent returns c's parent command, or nil for the root.
func (c *Command) Parent() *Command {
	return c.parent
}

// Root returns the top-most command in the tree.
func (c *Command) Root() *Command {
	if c.parent != nil {
		return c.parent.Root()
	}
	return c
}

// Name returns the command's name: the first whitespace-delimited token of Use.
func (c *Command) Name() string {
	name := c.Use
	if i := strings.IndexAny(name, " \t"); i >= 0 {
		name = name[:i]
	}
	return name
}

// Flags returns the set used both to register local flags and, at execution
// time, to read every flag available to the command (local plus inherited
// persistent flags).
func (c *Command) Flags() *FlagSet {
	if c.lflags == nil {
		c.lflags = &FlagSet{}
	}
	return c.lflags
}

// PersistentFlags returns the set of flags this command shares with its
// descendants.
func (c *Command) PersistentFlags() *FlagSet {
	if c.pflags == nil {
		c.pflags = &FlagSet{}
	}
	return c.pflags
}

// MarkFlagsRequiredTogether records that the named flags must all be set
// together or not at all.
func (c *Command) MarkFlagsRequiredTogether(names ...string) {
	c.flagGroupsRequiredTogether = append(c.flagGroupsRequiredTogether, names)
}

// Context returns the command's context, inheriting from ancestors and
// defaulting to context.Background().
func (c *Command) Context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	if c.parent != nil {
		return c.parent.Context()
	}
	return context.Background()
}

// SetContext sets the context returned by Context.
func (c *Command) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// SetOut sets the writer used for standard output (and help).
func (c *Command) SetOut(w io.Writer) {
	c.outWriter = w
}

// SetErr sets the writer used for error output.
func (c *Command) SetErr(w io.Writer) {
	c.errWriter = w
}

// OutOrStdout returns the standard-output writer, inheriting from ancestors
// and defaulting to os.Stdout.
func (c *Command) OutOrStdout() io.Writer {
	return c.getOut(os.Stdout)
}

// OutOrStderr returns the standard-output writer, falling back to os.Stderr.
func (c *Command) OutOrStderr() io.Writer {
	return c.getOut(os.Stderr)
}

// ErrOrStderr returns the error writer, inheriting from ancestors and
// defaulting to os.Stderr.
func (c *Command) ErrOrStderr() io.Writer {
	return c.getErr(os.Stderr)
}

// InOrStdin returns the input reader, inheriting from ancestors and defaulting
// to os.Stdin.
func (c *Command) InOrStdin() io.Reader {
	if c.inReader != nil {
		return c.inReader
	}
	if c.parent != nil {
		return c.parent.InOrStdin()
	}
	return os.Stdin
}

// SetIn sets the reader used for input.
func (c *Command) SetIn(r io.Reader) {
	c.inReader = r
}

func (c *Command) getOut(def io.Writer) io.Writer {
	if c.outWriter != nil {
		return c.outWriter
	}
	if c.parent != nil {
		return c.parent.getOut(def)
	}
	return def
}

func (c *Command) getErr(def io.Writer) io.Writer {
	if c.errWriter != nil {
		return c.errWriter
	}
	if c.parent != nil {
		return c.parent.getErr(def)
	}
	return def
}

// findChild returns the subcommand matching name (by name or alias), or nil.
func (c *Command) findChild(name string) *Command {
	for _, sub := range c.commands {
		if sub.Name() == name {
			return sub
		}
		for _, alias := range sub.Aliases {
			if alias == name {
				return sub
			}
		}
	}
	return nil
}

// commandPath returns the space-joined names from the root down to c.
func (c *Command) commandPath() string {
	if c.parent == nil {
		return c.Name()
	}
	return c.parent.commandPath() + " " + c.Name()
}
