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
package blade

import (
	"errors"
	"fmt"
	"os"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// validHardware checks that the hardware type is valid by comparing it against the list of hardware types
func validHardware(cmd *cobra.Command, args []string) (err error) {
	log.Debug().Msgf("Validating hardware %+v", root.D)
	if cmd.Flags().Changed("list-supported-types") {
		cmd.SetOut(os.Stdout)
		for _, hw := range root.BladeTypes {
			// print additional provider defaults
			if root.Verbose {
				cmd.Printf("%s\n", hw.Slug)
			} else {
				cmd.Printf("%s\n", hw.Slug)
			}
		}
		os.Exit(0)
	}

	if len(args) == 0 {
		bladeTypes := []string{}
		for _, hw := range root.BladeTypes {
			bladeTypes = append(bladeTypes, hw.Slug)
		}
		return fmt.Errorf("No hardware type provided: Choose from: %s", bladeTypes)
	}

	// Check that each arg is a valid blade type
	for _, arg := range args {
		matchFound := false
		for _, device := range root.BladeTypes {
			if arg == device.Slug {
				matchFound = true
				break
			}
		}
		if !matchFound {
			return errors.New("Invalid hardware type: " + arg)
		}
	}

	cabinetSet := cmd.Flags().Changed("cabinet")
	chassisSet := cmd.Flags().Changed("chassis")
	bladeSet := cmd.Flags().Changed("blade")
	autoSet := cmd.Flags().Changed("auto")
	// if auto is set, the values are recommended and the required flags are bypassed
	if autoSet {
		return nil
	} else {
		// No flags set
		if !cabinetSet && !chassisSet && !bladeSet {
			return errors.New("required flag(s) \"blade\", \"cabinet\", \"chassis\" not set")
		}
		// permutations with cabinet set
		if cabinetSet && !chassisSet && !bladeSet {
			return errors.New("required flag(s) \"blade\", \"chassis\" not set")
		}
		if cabinetSet && chassisSet && !bladeSet {
			return errors.New("required flag(s) \"blade\", not set")
		}
		// permutations with chassis set
		if !cabinetSet && chassisSet && !bladeSet {
			return errors.New("required flag(s) \"blade\", \"cabinet\" not set")
		}
		if !cabinetSet && chassisSet && bladeSet {
			return errors.New("required flag(s) \"cabinet\" not set")
		}
		// permutations with blade set
		if !cabinetSet && !chassisSet && bladeSet {
			return errors.New("required flag(s) \"cabinet\", \"chassis\" not set")
		}
		if cabinetSet && !chassisSet && bladeSet {
			return errors.New("required flag(s) \"chassis\" not set")
		}
		if !cabinetSet && chassisSet && bladeSet {
			return errors.New("required flag(s) \"cabinet\" not set")
		}
	}

	return nil
}
