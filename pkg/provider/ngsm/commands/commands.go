/*
 *
 *  MIT License
 *
 *  (C) Copyright 2024 Hewlett Packard Enterprise Development LP
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
package commands

import (
	"log"

	"github.com/spf13/cobra"
)

// NewProviderCmd creates commands for the NGSM provider
// This function is called by the cani command to create provider-specific commands
// based on the parent command (caniCmd).
func NewProviderCmd(caniCmd *cobra.Command) (providerCmd *cobra.Command, err error) {
	switch caniCmd.Name() {
	case "init": // session init
		providerCmd, _ = newInitCmd(caniCmd)
	case "add":
		providerCmd, _ = newAddCmd(caniCmd)
	case "show":
		providerCmd, _ = newShowCmd(caniCmd)
	default:
		// this provider doesn’t customize that verb
		return nil, nil
	}

	return providerCmd, nil
}

// newInitCmd creates the "init" command for the NGSM provider.
// This command is used to run NGSM-specific initialization tasks before the cani init command.
// It returns a cobra.Command that can be added to the cani command hierarchy.
// The command accepts a flag for specifying a BoM file, which can be provided multiple times.
func newInitCmd(caniCmd *cobra.Command) (providerCmd *cobra.Command, err error) {
	providerCmd = &cobra.Command{
		RunE: initializeNGSM,
	}
	providerCmd.Flags().StringArrayP("bom", "b", []string{}, "Path to a BoM file (can specify multiple times)")

	return providerCmd, nil
}

// initializeNGSM is the function that runs during "session init" for the NGSM provider.
// but runs prior to the task of the cani providerd init command.
func initializeNGSM(cmd *cobra.Command, args []string) error {
	log.Println("Running NGSM tasks prior to cani init...")
	return nil
}

// newAddCmd creates the "add" command for the NGSM provider.
func newAddCmd(caniCmd *cobra.Command) (providerCmd *cobra.Command, err error) {
	providerCmd = &cobra.Command{}
	providerCmd.Flags().StringP("superadd", "s", "", "Super add")

	return providerCmd, nil
}

func newShowCmd(caniCmd *cobra.Command) (providerCmd *cobra.Command, err error) {
	providerCmd = &cobra.Command{
		RunE: showNGSM,
	}
	providerCmd.Flags().StringP("superadd", "s", "", "Super add")
	return providerCmd, nil
}

func showNGSM(cmd *cobra.Command, args []string) error {
	// add additional output formats

	return nil
}
