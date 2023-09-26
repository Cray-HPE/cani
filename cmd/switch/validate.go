/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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
package sw

import (
	"errors"
	"fmt"
	"os"
	"strings"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/spf13/cobra"
)

// validHardware checks that the hardware type is valid by comparing it against the list of hardware types
func validHardware(cmd *cobra.Command, args []string) (err error) {
	// the PersistentPreRunE from the root command does not work here
	// so it is explicitly called
	err = root.SetupDomain(cmd, args)
	if err != nil {
		os.Exit(1)
	}

	if cmd.Flags().Changed("list-supported-types") {
		cmd.SetOut(os.Stdout)
		for _, hw := range root.MgmtSwitchTypes {
			cmd.Printf("%s\n", hw.Slug)
		}
		for _, hw := range root.HsnSwitchTypes {
			cmd.Printf("%s\n", hw.Slug)
		}
		os.Exit(0)
	}

	if len(args) == 0 {
		types := []string{}
		for _, hw := range root.MgmtSwitchTypes {
			types = append(types, hw.Slug)
		}
		for _, hw := range root.HsnSwitchTypes {
			types = append(types, hw.Slug)
		}
		return fmt.Errorf("No hardware type provided: Choose from: %s", strings.Join(types, "\", \""))
	}

	// Check that each arg is a valid switch type
	for _, arg := range args {
		matchFound := false
		for _, device := range root.MgmtSwitchTypes {
			if arg == device.Slug {
				matchFound = true
				break
			}
		}
		for _, device := range root.HsnSwitchTypes {
			if arg == device.Slug {
				matchFound = true
				break
			}
		}
		if !matchFound {
			return errors.New("Invalid hardware type: " + arg)
		}
	}

	return nil
}
