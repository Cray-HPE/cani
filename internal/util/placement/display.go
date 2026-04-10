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
package placement

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// PrintPlanWithHeight writes a placement table using an explicit device height.
func PrintPlanWithHeight(w io.Writer, entries []PlacementEntry, names []string, height int) {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "#\tRack\tStartU\tEndU\tFace\tZone\tName")
	fmt.Fprintln(tw, "-\t----\t------\t----\t----\t----\t----")
	for i, e := range entries {
		name := ""
		if i < len(names) {
			name = names[i]
		}
		endU := e.StartU + height - 1
		fmt.Fprintf(tw, "%d\t%s\t%d\t%d\t%s\t%s\t%s\n",
			i+1, e.RackName, e.StartU, endU, e.Face, e.Zone, name)
	}
	tw.Flush()
}
