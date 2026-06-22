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
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GenMarkdownTree writes a Markdown reference file for cmd and each of its
// subcommands into dir.  It is the standard-library replacement for
// cobra/doc.GenMarkdownTree and produces one file per command named after the
// command path (e.g. "cani alpha show" -> "cani_alpha_show.md").
func GenMarkdownTree(cmd *Command, dir string) error {
	for _, sub := range cmd.commands {
		if err := GenMarkdownTree(sub, dir); err != nil {
			return err
		}
	}
	filename := filepath.Join(dir, fileName(cmd))
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return writeMarkdown(cmd, f)
}

// fileName returns the Markdown file name for a command.
func fileName(cmd *Command) string {
	return strings.ReplaceAll(cmd.commandPath(), " ", "_") + ".md"
}

// writeMarkdown renders a single command's reference page.
func writeMarkdown(cmd *Command, w io.Writer) error {
	cmd.mergeInheritedFlags()
	var b strings.Builder

	fmt.Fprintf(&b, "## %s\n\n", cmd.commandPath())
	if cmd.Short != "" {
		fmt.Fprintf(&b, "%s\n\n", cmd.Short)
	}

	b.WriteString("### Synopsis\n\n")
	if cmd.Long != "" {
		fmt.Fprintf(&b, "%s\n\n", strings.TrimSpace(cmd.Long))
	} else if cmd.Short != "" {
		fmt.Fprintf(&b, "%s\n\n", cmd.Short)
	}
	fmt.Fprintf(&b, "```\n%s [flags]\n```\n\n", cmd.commandPath())

	if local := cmd.flagUsages(false); local != "" {
		fmt.Fprintf(&b, "### Options\n\n```\n%s```\n\n", local)
	}
	if inherited := cmd.flagUsages(true); inherited != "" {
		fmt.Fprintf(&b, "### Options inherited from parent commands\n\n```\n%s```\n\n", inherited)
	}

	writeSeeAlso(cmd, &b)

	_, err := io.WriteString(w, b.String())
	return err
}

// writeSeeAlso renders links to the parent command and direct subcommands.
func writeSeeAlso(cmd *Command, b *strings.Builder) {
	if cmd.parent == nil && len(cmd.commands) == 0 {
		return
	}
	b.WriteString("### SEE ALSO\n\n")
	if cmd.parent != nil {
		fmt.Fprintf(b, "* [%s](%s)\t - %s\n", cmd.parent.commandPath(), fileName(cmd.parent), cmd.parent.Short)
	}
	children := append([]*Command(nil), cmd.commands...)
	sort.Slice(children, func(i, j int) bool { return children[i].Name() < children[j].Name() })
	for _, child := range children {
		fmt.Fprintf(b, "* [%s](%s)\t - %s\n", child.commandPath(), fileName(child), child.Short)
	}
	b.WriteString("\n")
}
