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
	"fmt"
	"os"
	"strings"
)

// Execute resolves the target command from os.Args and runs it.  It is called
// on the root command.
func (c *Command) Execute() error {
	root := c.Root()
	target, rest := root.find(os.Args[1:])
	return target.run(rest)
}

// find walks the command tree, consuming subcommand-name tokens (which may be
// interspersed with flags) until no further subcommand matches.  It returns the
// resolved command and the remaining flag/positional tokens.
func (c *Command) find(args []string) (*Command, []string) {
	cur := c
	rest := args
	for {
		cur.mergeInheritedFlags()
		sub, idx := cur.scanForSubcommand(rest)
		if sub == nil {
			return cur, rest
		}
		rest = append(rest[:idx:idx], rest[idx+1:]...)
		cur = sub
	}
}

// scanForSubcommand returns the first subcommand named by a non-flag token in
// args, along with that token's index, skipping over flags and their values.
func (c *Command) scanForSubcommand(args []string) (*Command, int) {
	i := 0
	for i < len(args) {
		tok := args[i]
		if tok == "--" {
			return nil, -1
		}
		if strings.HasPrefix(tok, "-") && len(tok) > 1 {
			i += c.flagAdvance(args, i)
			continue
		}
		if sub := c.findChild(tok); sub != nil {
			return sub, i
		}
		return nil, -1
	}
	return nil, -1
}

// flagAdvance reports how many tokens the flag at args[i] occupies (1 or 2)
// during subcommand scanning.  It mirrors cobra's stripFlags heuristic: a long
// or single-character flag that requires a value consumes the following token.
func (c *Command) flagAdvance(args []string, i int) int {
	s := args[i]
	hasNext := i+1 < len(args)
	if strings.HasPrefix(s, "--") {
		if strings.Contains(s, "=") {
			return 1
		}
		flag := c.Flags().Lookup(s[2:])
		if flag != nil && flag.NoOptDefVal != "" {
			return 1
		}
		if hasNext {
			return 2
		}
		return 1
	}
	if len(s) == 2 {
		flag := c.Flags().shorthandLookup(s[1:])
		if flag != nil && flag.NoOptDefVal != "" {
			return 1
		}
		if hasNext {
			return 2
		}
	}
	return 1
}

// mergeInheritedFlags folds this command's own and all ancestors' persistent
// flags into the command's flag set so they are parseable and readable here.
func (c *Command) mergeInheritedFlags() {
	fs := c.Flags()
	for p := c; p != nil; p = p.parent {
		if p.pflags != nil {
			fs.AddFlagSet(p.pflags)
		}
	}
}

// run parses args for the resolved command and executes the hook pipeline:
// help/version short-circuits, persistent pre-run, argument validation,
// pre-run, and the main run function.
func (c *Command) run(args []string) error {
	c.mergeInheritedFlags()
	if c.parent == nil && c.Version != "" && c.Flags().Lookup("version") == nil {
		c.Flags().Bool("version", false, "version for "+c.Name())
	}

	positionals, helpRequested, err := c.Flags().parse(args)
	if err != nil {
		return c.fail(err)
	}
	if helpRequested {
		return c.Help()
	}
	if c.parent == nil && c.Version != "" && c.Flags().Changed("version") {
		fmt.Fprintf(c.OutOrStdout(), "%s version %s\n", c.Name(), c.Version)
		return nil
	}
	if err := c.runPipeline(positionals); err != nil {
		return c.fail(err)
	}
	return nil
}

// runPipeline executes the validation and run hooks once flags are parsed.
// Errors are returned unadorned; run wraps them with fail so the message and
// usage are printed once.
func (c *Command) runPipeline(positionals []string) error {
	if err := c.validateFlagGroups(); err != nil {
		return err
	}
	if pre := c.nearestPersistentPreRunE(); pre != nil {
		if err := pre(c, positionals); err != nil {
			return err
		}
	}
	if c.Args != nil {
		if err := c.Args(c, positionals); err != nil {
			return err
		}
	}
	if c.PreRunE != nil {
		if err := c.PreRunE(c, positionals); err != nil {
			return err
		}
	}
	if c.RunE != nil {
		return c.RunE(c, positionals)
	}
	return c.Help()
}

// nearestPersistentPreRunE returns the closest PersistentPreRunE hook walking
// up from c to the root, or nil when none is defined.
func (c *Command) nearestPersistentPreRunE() RunFunc {
	for p := c; p != nil; p = p.parent {
		if p.PersistentPreRunE != nil {
			return p.PersistentPreRunE
		}
	}
	return nil
}

// validateFlagGroups enforces every MarkFlagsRequiredTogether constraint: in
// each group either all flags are set or none are.
func (c *Command) validateFlagGroups() error {
	for _, group := range c.flagGroupsRequiredTogether {
		set := 0
		for _, name := range group {
			if c.Flags().Changed(name) {
				set++
			}
		}
		if set != 0 && set != len(group) {
			return fmt.Errorf("if any flags in the group [%s] are set they must all be set; missing %s",
				strings.Join(group, " "), missingFlags(c, group))
		}
	}
	return nil
}

// missingFlags lists the flags in group that were not set.
func missingFlags(c *Command, group []string) string {
	var missing []string
	for _, name := range group {
		if !c.Flags().Changed(name) {
			missing = append(missing, name)
		}
	}
	return "[" + strings.Join(missing, " ") + "]"
}

// fail prints the error and the command usage, then returns the error so the
// root caller can set a non-zero exit code (matching cobra's default of
// printing usage on error).
func (c *Command) fail(err error) error {
	fmt.Fprintln(c.ErrOrStderr(), "Error:", err)
	fmt.Fprint(c.ErrOrStderr(), c.UsageString())
	return err
}
