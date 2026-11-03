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

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// TestRenderTreeConnectors verifies RenderTree emits the exact connector layout
// for nested children and sibling leaves.
//
// Why it matters: tree connector glyphs are part of the human-readable output
// contract for visual inventory views.
// Inputs: one root with two children, where the second child has two children of
// its own. Outputs: a complete no-color tree string.
// Data choice: first/last siblings and grandchildren exercise both ├── and └──
// branches plus child indentation.
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

	var output bytes.Buffer
	options := TreeOptions{NoColor: true, Writer: &output}

	RenderTree(&output, roots, options)

	want := strings.Join([]string{
		"root info",
		"├── child-a detail-a",
		"└── child-b detail-b",
		"    ├── grandchild-1",
		"    └── grandchild-2",
		"",
	}, "\n")
	assertExactOutput(t, output.String(), want)
}

// TestRenderTreeVerticalContinuation verifies descendants under a non-last
// sibling retain the vertical continuation glyph.
//
// Why it matters: without the │ continuation, larger inventory trees become
// visually ambiguous about which parent owns a child.
// Inputs: one root with a non-last child that owns a leaf and a following
// sibling. Outputs: a complete no-color tree string.
// Data choice: the mid node is intentionally followed by a sibling so the child
// leaf must render under a continued vertical branch.
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

	var output bytes.Buffer
	options := TreeOptions{NoColor: true, Writer: &output}

	RenderTree(&output, roots, options)

	want := strings.Join([]string{
		"top",
		"├── mid",
		"│   └── leaf",
		"└── sibling",
		"",
	}, "\n")
	assertExactOutput(t, output.String(), want)
}

// TestRenderTreeMultipleRoots verifies RenderTree prints each root and its child
// independently.
//
// Why it matters: full inventory views may have multiple root sections, and a
// connector from one root must not bleed into the next.
// Inputs: two roots with one child each. Outputs: both roots and their child
// connectors in order.
// Data choice: single-child roots prove the lone child is treated as a last child
// under each independent root.
func TestRenderTreeMultipleRoots(t *testing.T) {
	roots := []TreeNode{
		{Label: "root-a", Children: []TreeNode{{Label: "a-child"}}},
		{Label: "root-b", Children: []TreeNode{{Label: "b-child"}}},
	}

	var output bytes.Buffer
	options := TreeOptions{NoColor: true, Writer: &output}

	RenderTree(&output, roots, options)

	want := strings.Join([]string{
		"root-a",
		"└── a-child",
		"root-b",
		"└── b-child",
		"",
	}, "\n")
	assertExactOutput(t, output.String(), want)
}

// TestRenderTreeEmptyRoots verifies nil roots render no output.
//
// Why it matters: callers should be able to pass an empty tree without producing
// stray blank lines in command output.
// Inputs: nil roots and a buffer writer. Outputs: an empty string.
// Data choice: nil roots cover the empty input path directly without requiring
// unrelated inventory setup.
func TestRenderTreeEmptyRoots(t *testing.T) {
	var output bytes.Buffer
	options := TreeOptions{NoColor: true, Writer: &output}

	RenderTree(&output, nil, options)

	assertExactOutput(t, output.String(), "")
}

// TestRenderTreeNoColorDisablesAnsiCodes verifies no-color tree rendering emits
// plain text only.
//
// Why it matters: tests and users that request no color need stable output
// without terminal escape sequences.
// Inputs: one root with detail and one child, rendered with NoColor true. Outputs:
// a complete plain-text tree string without ANSI color codes.
// Data choice: detail text is included because detail is the part normally
// wrapped in gray when color is enabled.
func TestRenderTreeNoColorDisablesAnsiCodes(t *testing.T) {
	roots := []TreeNode{
		{Label: "r", Detail: "d", Children: []TreeNode{{Label: "c"}}},
	}

	var output bytes.Buffer
	options := TreeOptions{NoColor: true, Writer: &output}

	RenderTree(&output, roots, options)

	want := "r d\n└── c\n"
	assertExactOutput(t, output.String(), want)
	if strings.Contains(output.String(), ColorGray) || strings.Contains(output.String(), ColorReset) {
		t.Fatalf("NoColor=true emitted ANSI codes:\n%s", output.String())
	}
}

// TestRenderTreeColorEmitsAnsiCodes verifies colored tree rendering wraps detail
// and connector text with gray ANSI codes.
//
// Why it matters: colorized tree output is a user-visible format and should keep
// its dimmed detail/connectors behavior.
// Inputs: one root with detail and one child, rendered with NoColor false.
// Outputs: a complete tree string containing gray and reset escape sequences.
// Data choice: a root detail plus child connector covers both call sites that use
// the gray formatter.
func TestRenderTreeColorEmitsAnsiCodes(t *testing.T) {
	roots := []TreeNode{
		{Label: "r", Detail: "d", Children: []TreeNode{{Label: "c"}}},
	}

	var output bytes.Buffer
	options := TreeOptions{NoColor: false, Writer: &output}

	RenderTree(&output, roots, options)

	want := "r " + ColorGray + "d" + ColorReset + "\n" +
		ColorGray + "└── " + ColorReset + "c\n"
	assertExactOutput(t, output.String(), want)
}

// TestBuildModuleTreeGroupsByPlacement verifies the module tree places an
// installed module under its parent device with its bay, a spare under the
// location that stores it tagged "spare", and a module with neither
// placement under "(unassigned)".
//
// Why it matters: this is how an operator finds a card; Nautobot displays a
// module as "installed in <device>" or "at location <location>", and a
// spare listed under "(unassigned)" with an empty bay tag read as an error.
// Inputs: device cn-01 with module nic-a in bay PCIe1, an unnamed spare with
// serial S at location lab, and a module with no placement.
// Outputs: three roots labelled cn-01 (device), lab (location) and
// "(unassigned)"; the spare's line shows "spare" and "CX4 (Serial: S)".
func TestBuildModuleTreeGroupsByPlacement(t *testing.T) {
	inv := devicetypes.NewInventory()
	dev, loc := uuid.New(), uuid.New()
	inv.Devices[dev] = &devicetypes.CaniDeviceType{ID: dev, Name: "cn-01"}
	inv.Locations[loc] = &devicetypes.CaniLocationType{ID: loc, Name: "lab"}
	modules := []*devicetypes.CaniModuleType{
		{ID: uuid.New(), Name: "nic-a", ParentDevice: dev, ModuleBayName: "PCIe1"},
		{ID: uuid.New(), Model: "CX4", Serial: "S", Location: loc},
		{ID: uuid.New(), Name: "lost"},
	}

	nodes := BuildModuleTree(modules, inv, TreeFilter{NoColor: true})
	if len(nodes) != 3 {
		t.Fatalf("roots = %d, want 3", len(nodes))
	}
	want := []struct{ label, detail, child string }{
		{"cn-01", "device", "bay:PCIe1 | nic-a"},
		{"lab", "location", "spare | CX4 (Serial: S)"},
		{"(unassigned)", "device", "bay: | lost"},
	}
	for i, w := range want {
		if nodes[i].Label != w.label || nodes[i].Detail != w.detail || len(nodes[i].Children) != 1 || !strings.Contains(nodes[i].Children[0].Label, w.child) {
			t.Errorf("root %d = {%q %q %v}, want {%q %q child containing %q}", i, nodes[i].Label, nodes[i].Detail, nodes[i].Children, w.label, w.detail, w.child)
		}
	}
}
