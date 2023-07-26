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
package cabinet

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
	library, err := hardwaretypes.NewEmbeddedLibrary()
	if err != nil {
		return err
	}

	// Get the list of hardware types that are cabinets
	deviceTypes := library.GetDeviceTypesByHardwareType(hardwaretypes.Cabinet)
	if cmd.Flags().Changed("list-supported-types") {
		for _, hw := range deviceTypes {
			cmd.Printf("- %s\n", hw.Slug)
		}
		os.Exit(0)
	}

	if len(args) == 0 {
		types := []string{}
		for _, hw := range deviceTypes {
			types = append(types, hw.Slug)
		}
		return fmt.Errorf("No hardware type provided: Choose from: %s", strings.Join(types, "\", \""))
	}

	// Check that each arg is a valid cabinet type
	for _, arg := range args {
		matchFound := false
		for _, device := range deviceTypes {
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

// validFlagCombos has additional flag logic to account for overiding required flags with the --auto flag
func validFlagCombos(cmd *cobra.Command, args []string) error {
	// ensure the session is up and the datastore exists
	err := root.DatastoreExists(cmd, args)
	if err != nil {
		return err
	}

	cabinetSet := cmd.Flags().Changed("cabinet")
	vlanIdSet := cmd.Flags().Changed("vlan-id")
	autoSet := cmd.Flags().Changed("auto")
	// if auto is set, the values are recommended and the required flags are bypassed
	if autoSet {
		return nil
	} else {
		if !cabinetSet && !vlanIdSet {
			return errors.New("required flag(s) \"cabinet\", \"vlan-id\" not set")
		}
		if cabinetSet && !vlanIdSet {
			return errors.New("required flag(s) \"vlan-id\" not set")
		}
		if !cabinetSet && vlanIdSet {
			return errors.New("required flag(s) \"cabinet\" not set")
		}
	}

	return nil
}
