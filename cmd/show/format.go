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
package show

import (
	"fmt"
	"strings"

	"github.com/Cray-HPE/cani/pkg/visual"
	"github.com/spf13/cobra"
)

// validWithValues lists the accepted tokens for the --with flag.
var validWithValues = []string{"modules", "interfaces", "cables", "empty-us", "roles"}

// validateWithFlag checks that every --with token is recognised.
func validateWithFlag(cmd *cobra.Command) error {
	vals, _ := cmd.Flags().GetStringSlice("with")
	for _, v := range vals {
		if !contains(validWithValues, v) {
			return fmt.Errorf("invalid --with value %q; valid options: %s",
				v, strings.Join(validWithValues, ", "))
		}
	}
	return nil
}

// treeFilterFromCmd builds a TreeFilter from the --with flag.
func treeFilterFromCmd(cmd *cobra.Command) visual.TreeFilter {
	vals, _ := cmd.Flags().GetStringSlice("with")
	set := make(map[string]bool, len(vals))
	for _, v := range vals {
		set[v] = true
	}
	nc, _ := cmd.Flags().GetBool("no-color")
	return visual.TreeFilter{
		Modules:    set["modules"],
		Interfaces: set["interfaces"],
		Cables:     set["cables"],
		EmptyUs:    set["empty-us"],
		Roles:      set["roles"],
		NoColor:    nc,
	}
}
