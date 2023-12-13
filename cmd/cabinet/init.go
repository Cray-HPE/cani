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
	"os"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	auto   bool
	accept bool
	format string
	sortBy string
)

func Init() {
	log.Trace().Msgf("%+v", "github.com/Cray-HPE/cani/cmd/cabinet.init")

	// Add subcommands to root commands
	root.AddCmd.AddCommand(AddCabinetCmd)
	root.ListCmd.AddCommand(ListCabinetCmd)
	root.RemoveCmd.AddCommand(RemoveCabinetCmd)

	// Common 'add cabinet' flags and then merge with provider-specified command
	AddCabinetCmd.Flags().BoolP("list-supported-types", "L", false, "List supported hardware types.")
	AddCabinetCmd.Flags().BoolVar(&auto, "auto", false, "Automatically recommend and assign required flags.")
	AddCabinetCmd.MarkFlagsMutuallyExclusive("auto")
	AddCabinetCmd.Flags().BoolVarP(&accept, "accept", "y", false, "Automatically accept recommended values.")

	// Common 'list cabinet' flags and then merge with provider-specified command
	ListCabinetCmd.Flags().StringVarP(&format, "format", "f", "pretty", "Format out output")
	ListCabinetCmd.Flags().StringVarP(&sortBy, "sort", "s", "location", "Sort by a specific key")

	// Register all provider commands during init()
	for _, p := range domain.GetProviders() {
		for _, c := range []*cobra.Command{AddCabinetCmd, ListCabinetCmd} {
			err := root.RegisterProviderCommand(p, c)
			if err != nil {
				log.Error().Msgf("Unable to get command '%s %s' from provider %s ", c.Parent().Name(), c.Name(), p.Slug())
				os.Exit(1)
			}
		}
	}
}
