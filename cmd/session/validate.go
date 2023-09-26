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
package session

import (
	"fmt"
	"os"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/spf13/cobra"
)

// validProvider checks that the provider is valid and that at least one argument is provided
func validProvider(cmd *cobra.Command, args []string) (err error) {
	// the PersistentPreRunE from the root command does not work here
	// so it is explicitly called
	err = root.SetupDomain(cmd, args)
	if err != nil {
		os.Exit(1)
	}

	if len(args) < 1 {
		return fmt.Errorf("need a provider.  Choose from: %+v", taxonomy.SupportedProviders)
	}

	// This helper function checks if a provider is valid.
	isValidProvider := func(provider string) bool {
		for _, validArg := range cmd.ValidArgs {
			if provider == validArg {
				return true
			}
		}
		return false
	}

	// Check the session provider.
	if args[0] == "" {
		// Check that at least one argument is provided.
		if len(args) < 1 {
			return fmt.Errorf("Need a provider.  Choose from: %+v", taxonomy.SupportedProviders)
		}
	} else if !isValidProvider(args[0]) {
		return fmt.Errorf("%s is not a valid provider.  Valid providers: %+v", args[0], taxonomy.SupportedProviders)
	}

	// Check all argument providers.
	for _, arg := range args {
		if !isValidProvider(arg) {
			return fmt.Errorf("%s is not a valid provider.  Valid providers: %+v", arg, taxonomy.SupportedProviders)
		}
	}

	return nil
}
