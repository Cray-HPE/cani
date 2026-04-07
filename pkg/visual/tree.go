/*
 *
 *  MIT License
 *
 *  (C) Copyright 2024-2026 Hewlett Packard Enterprise Development LP
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
package visual

import (
	"fmt"
	"io"
	"os"
)

// TreeNode represents a single node in a display tree.
type TreeNode struct {
	Label    string // Primary name (rendered in normal text)
	Detail   string // Additional metadata (rendered in gray/dimmed)
	Children []TreeNode
}

// TreeOptions controls tree rendering output.
type TreeOptions struct {
	NoColor bool
	Writer  io.Writer
}

// RenderTree writes a tree of nodes to w using box-drawing characters.
func RenderTree(w io.Writer, roots []TreeNode, opts TreeOptions) {
	gray := makeGrayFunc(opts)
	for _, root := range roots {
		detail := ""
		if root.Detail != "" {
			detail = " " + gray(root.Detail)
		}
		fmt.Fprintf(w, "%s%s\n", root.Label, detail)
		for i, child := range root.Children {
			isLast := i == len(root.Children)-1
			renderNode(w, child, "", isLast, gray)
		}
	}
}

// RenderTreeToStdout is a convenience wrapper that writes to stdout.
func RenderTreeToStdout(roots []TreeNode, opts TreeOptions) {
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}
	RenderTree(opts.Writer, roots, opts)
}

// makeGrayFunc returns a function that wraps text in gray ANSI codes.
func makeGrayFunc(opts TreeOptions) func(string) string {
	if opts.NoColor {
		return func(s string) string { return s }
	}
	return func(s string) string { return ColorGray + s + ColorReset }
}

// renderNode recursively renders a single tree node with proper prefixes.
func renderNode(w io.Writer, node TreeNode, prefix string, isLast bool, gray func(string) string) {
	connector := "├── "
	continuation := "│   "
	if isLast {
		connector = "└── "
		continuation = "    "
	}

	detail := ""
	if node.Detail != "" {
		detail = " " + gray(node.Detail)
	}

	fmt.Fprintf(w, "%s%s%s\n", gray(prefix+connector), node.Label, detail)

	childPrefix := prefix + continuation
	for i, child := range node.Children {
		childIsLast := i == len(node.Children)-1
		renderNode(w, child, childPrefix, childIsLast, gray)
	}
}
