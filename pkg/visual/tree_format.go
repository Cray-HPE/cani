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

// TreeFilter controls which child types are included in tree output.
type TreeFilter struct {
	Modules    bool
	Interfaces bool
	Cables     bool
	EmptyUs    bool
	Roles      bool
	NoColor    bool
}

// Tree icons per node type.
const (
	IconRack              = "■"
	IconDevice            = "●"
	IconModule            = "◆"
	IconCable             = "═"
	IconCableDisconnected = "≄"
	IconInterface         = "○"
	IconFru               = "□"
	IconLocation          = "◇"
)

// TreeIcon returns "icon (type) name" with the icon colored, type dimmed, and name bold white.
func TreeIcon(icon, color, typeTag, name string, noColor bool) string {
	if noColor {
		return icon + " (" + typeTag + ") " + name
	}
	return color + icon + ColorReset + " " +
		ColorGray + "(" + typeTag + ")" + ColorReset + " " +
		ColorBold + ColorWhite + name + ColorReset
}

// TreeIconColored is like TreeIcon but colors the name with nameColor instead of white.
func TreeIconColored(icon, iconColor, typeTag, name, nameColor string, noColor bool) string {
	if noColor {
		return icon + " (" + typeTag + ") " + name
	}
	return iconColor + icon + ColorReset + " " +
		ColorGray + "(" + typeTag + ")" + ColorReset + " " +
		nameColor + name + ColorReset
}

// StatusAnsi returns the ANSI color code for a status string.
func StatusAnsi(status string) string {
	switch StatusColor(status) {
	case "green":
		return ColorGreen
	case "red":
		return ColorRed
	case "yellow":
		return ColorYellow
	default:
		return ColorCyan
	}
}

// ColorInGray wraps text in a color then switches back to gray.
// When noColor is true, returns the text unchanged.
func ColorInGray(s, color string, noColor bool) string {
	if noColor {
		return s
	}
	return color + s + ColorGray
}

// PipeSep joins non-empty strings with " | ".
func PipeSep(parts ...string) string {
	var out []string
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return JoinNonEmpty(out, " | ")
}

// RenderTreeOutput renders tree nodes to stdout.
func RenderTreeOutput(nodes []TreeNode, noColor bool) {
	opts := TreeOptions{NoColor: noColor}
	RenderTreeToStdout(nodes, opts)
}
