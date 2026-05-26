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
	"bytes"
	"strings"
	"testing"
)

func TestRenderTreeConnectors(t *testing.T) {
	roots := []TreeNode{
		{
			Label:  "root",
			Detail: "info",
			Children: []TreeNode{
				{Label: "child-a", Detail: "detail-a"},
				{
					Label:  "child-b",
					Detail: "detail-b",
					Children: []TreeNode{
						{Label: "grandchild-1"},
						{Label: "grandchild-2"},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	opts := TreeOptions{NoColor: true, Writer: &buf}
	RenderTree(&buf, roots, opts)
	got := buf.String()

	// Root should print without connectors
	if !strings.Contains(got, "root info\n") {
		t.Errorf("root line missing or malformed:\n%s", got)
	}

	// First child uses ├──
	if !strings.Contains(got, "├── child-a detail-a\n") {
		t.Errorf("expected ├── connector for child-a:\n%s", got)
	}

	// Last child uses └──
	if !strings.Contains(got, "└── child-b detail-b\n") {
		t.Errorf("expected └── connector for child-b:\n%s", got)
	}

	// Grandchildren under child-b should have continuation lines
	if !strings.Contains(got, "    ├── grandchild-1\n") {
		t.Errorf("expected continuation + ├── for grandchild-1:\n%s", got)
	}
	if !strings.Contains(got, "    └── grandchild-2\n") {
		t.Errorf("expected continuation + └── for grandchild-2:\n%s", got)
	}
}

func TestRenderTreeVerticalContinuation(t *testing.T) {
	roots := []TreeNode{
		{
			Label: "top",
			Children: []TreeNode{
				{
					Label: "mid",
					Children: []TreeNode{
						{Label: "leaf"},
					},
				},
				{Label: "sibling"},
			},
		},
	}

	var buf bytes.Buffer
	opts := TreeOptions{NoColor: true, Writer: &buf}
	RenderTree(&buf, roots, opts)
	got := buf.String()

	// mid is not last child, so its children should have │ continuation
	if !strings.Contains(got, "│   └── leaf\n") {
		t.Errorf("expected │ continuation for leaf under non-last mid:\n%s", got)
	}
}

func TestRenderTreeMultipleRoots(t *testing.T) {
	roots := []TreeNode{
		{Label: "root-a", Children: []TreeNode{{Label: "a-child"}}},
		{Label: "root-b", Children: []TreeNode{{Label: "b-child"}}},
	}

	var buf bytes.Buffer
	opts := TreeOptions{NoColor: true, Writer: &buf}
	RenderTree(&buf, roots, opts)
	got := buf.String()

	if !strings.Contains(got, "root-a\n") {
		t.Errorf("expected root-a line:\n%s", got)
	}
	if !strings.Contains(got, "root-b\n") {
		t.Errorf("expected root-b line:\n%s", got)
	}
	if !strings.Contains(got, "└── a-child\n") {
		t.Errorf("expected └── for a-child:\n%s", got)
	}
	if !strings.Contains(got, "└── b-child\n") {
		t.Errorf("expected └── for b-child:\n%s", got)
	}
}

func TestRenderTreeEmptyRoots(t *testing.T) {
	var buf bytes.Buffer
	opts := TreeOptions{NoColor: true, Writer: &buf}
	RenderTree(&buf, nil, opts)

	if buf.Len() != 0 {
		t.Errorf("expected empty output for nil roots, got: %s", buf.String())
	}
}

func TestRenderTreeNoColorDisablesAnsiCodes(t *testing.T) {
	roots := []TreeNode{
		{Label: "r", Detail: "d", Children: []TreeNode{{Label: "c"}}},
	}

	var buf bytes.Buffer
	opts := TreeOptions{NoColor: true, Writer: &buf}
	RenderTree(&buf, roots, opts)
	got := buf.String()

	if strings.Contains(got, ColorGray) || strings.Contains(got, ColorReset) {
		t.Errorf("NoColor=true but ANSI codes found in output:\n%s", got)
	}
}

func TestRenderTreeColorEmitsAnsiCodes(t *testing.T) {
	roots := []TreeNode{
		{Label: "r", Detail: "d", Children: []TreeNode{{Label: "c"}}},
	}

	var buf bytes.Buffer
	opts := TreeOptions{NoColor: false, Writer: &buf}
	RenderTree(&buf, roots, opts)
	got := buf.String()

	if !strings.Contains(got, ColorGray) {
		t.Errorf("expected ANSI gray codes in colored output:\n%s", got)
	}
}
