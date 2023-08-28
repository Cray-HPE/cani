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
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/spf13/cobra"
)

// validHardware checks that the hardware type is valid by comparing it against the list of hardware types
func validHardware(cmd *cobra.Command, args []string) error {
	library, err := hardwaretypes.NewEmbeddedLibrary(root.Conf.Session.DomainOptions.CustomHardwareTypesDir)
	if err != nil {
		return err
	}

	// Get the list of hardware types that are switches
	mgmtSwitchTypes := library.GetDeviceTypesByHardwareType(hardwaretypes.ManagementSwitch)
	hsnSwitchTypes := library.GetDeviceTypesByHardwareType(hardwaretypes.HighSpeedSwitch)

	if cmd.Flags().Changed("list-supported-types") {
		for _, hw := range mgmtSwitchTypes {
			cmd.Printf("- %s\n", hw.Slug)
		}
		for _, hw := range hsnSwitchTypes {
			cmd.Printf("- %s\n", hw.Slug)
		}
		os.Exit(0)
	}

	if len(args) == 0 {
		types := []string{}
		for _, hw := range mgmtSwitchTypes {
			types = append(types, hw.Slug)
		}
		for _, hw := range hsnSwitchTypes {
			types = append(types, hw.Slug)
		}
		return fmt.Errorf("No hardware type provided: Choose from: %s", strings.Join(types, "\", \""))
	}

	// Check that each arg is a valid switch type
	for _, arg := range args {
		matchFound := false
		for _, device := range mgmtSwitchTypes {
			if arg == device.Slug {
				matchFound = true
				break
			}
		}
		for _, device := range hsnSwitchTypes {
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
