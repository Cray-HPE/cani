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
	"strings"
)

// Help prints the command's long description (or short) followed by its usage.
func (c *Command) Help() error {
	c.mergeInheritedFlags()
	out := c.OutOrStdout()
	if c.Long != "" {
		fmt.Fprintln(out, strings.TrimSpace(c.Long))
	} else if c.Short != "" {
		fmt.Fprintln(out, strings.TrimSpace(c.Short))
	}
	fmt.Fprint(out, c.UsageString())
	return nil
}

// Print writes to the command's error stream, matching cobra's Print.
func (c *Command) Print(i ...interface{}) {
	fmt.Fprint(c.OutOrStderr(), i...)
}

// Println writes a line to the command's error stream, matching cobra.
func (c *Command) Println(i ...interface{}) {
	c.Print(fmt.Sprintln(i...))
}

// Printf writes a formatted string to the command's error stream, matching
// cobra's Printf.
func (c *Command) Printf(format string, i ...interface{}) {
	c.Print(fmt.Sprintf(format, i...))
}

// UsageString renders the cobra-style usage block: usage lines, aliases,
// available commands, and flag sections.
func (c *Command) UsageString() string {
	var b strings.Builder
	b.WriteString("\nUsage:\n")
	if c.RunE != nil || len(c.commands) == 0 {
		fmt.Fprintf(&b, "  %s [flags]\n", c.commandPath())
	}
	if len(c.commands) > 0 {
		fmt.Fprintf(&b, "  %s [command]\n", c.commandPath())
	}

	if len(c.Aliases) > 0 {
		fmt.Fprintf(&b, "\nAliases:\n  %s, %s\n", c.Name(), strings.Join(c.Aliases, ", "))
	}

	c.writeCommandList(&b)
	c.writeFlagSections(&b)

	if len(c.commands) > 0 {
		fmt.Fprintf(&b, "\nUse \"%s [command] --help\" for more information about a command.\n", c.commandPath())
	}
	return b.String()
}

// writeCommandList appends the "Available Commands" section.
func (c *Command) writeCommandList(b *strings.Builder) {
	if len(c.commands) == 0 {
		return
	}
	width := 0
	for _, sub := range c.commands {
		if n := len(sub.Name()); n > width {
			width = n
		}
	}
	b.WriteString("\nAvailable Commands:\n")
	for _, sub := range c.commands {
		fmt.Fprintf(b, "  %-*s  %s\n", width, sub.Name(), sub.Short)
	}
}

// writeFlagSections appends the local "Flags" and inherited "Global Flags".
func (c *Command) writeFlagSections(b *strings.Builder) {
	if local := c.flagUsages(false); local != "" {
		b.WriteString("\nFlags:\n")
		b.WriteString(local)
	}
	if global := c.flagUsages(true); global != "" {
		b.WriteString("\nGlobal Flags:\n")
		b.WriteString(global)
	}
}

// flagUsages formats either the inherited flags (global == true) or the
// command's own flags (global == false) as aligned help lines.
func (c *Command) flagUsages(global bool) string {
	var b strings.Builder
	for _, flag := range c.Flags().ordered {
		if flag.Hidden {
			continue
		}
		if c.isInherited(flag.Name) != global {
			continue
		}
		b.WriteString(formatFlagLine(flag))
	}
	return b.String()
}

// isInherited reports whether name comes from an ancestor's persistent flags.
func (c *Command) isInherited(name string) bool {
	for p := c.parent; p != nil; p = p.parent {
		if p.pflags != nil && p.pflags.Lookup(name) != nil {
			return true
		}
	}
	return false
}

// formatFlagLine renders one flag's help line, e.g.
// "  -o, --format string   Output format (default \"table\")".
func formatFlagLine(flag *Flag) string {
	var head string
	if flag.Shorthand != "" {
		head = fmt.Sprintf("  -%s, --%s", flag.Shorthand, flag.Name)
	} else {
		head = fmt.Sprintf("      --%s", flag.Name)
	}
	if t := flag.Value.Type(); t != "bool" && t != "count" {
		head += " " + t
	}
	line := fmt.Sprintf("%-30s %s", head, flag.Usage)
	if flag.DefValue != "" && flag.DefValue != "false" && flag.DefValue != "0" && flag.DefValue != "[]" {
		line += fmt.Sprintf(" (default %s)", quoteDefault(flag))
	}
	return strings.TrimRight(line, " ") + "\n"
}

// quoteDefault quotes a string default value, matching cobra's display.
func quoteDefault(flag *Flag) string {
	if flag.Value.Type() == "string" {
		return fmt.Sprintf("%q", flag.DefValue)
	}
	return flag.DefValue
}
