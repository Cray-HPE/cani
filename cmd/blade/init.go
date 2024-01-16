/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
	"os"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	auto      bool
	accept    bool
	cabinet   int
	chassis   int
	blade     int
	recursion bool
	format    string
	sortBy    string
)

func Init() {
	log.Trace().Msgf("%+v", "github.com/Cray-HPE/cani/cmd/blade.init")
	// Add blade variants to root commands
	root.AddCmd.AddCommand(AddBladeCmd)
	root.ListCmd.AddCommand(ListBladeCmd)
	root.RemoveCmd.AddCommand(RemoveBladeCmd)

	// Add a flag to show supported types
	AddBladeCmd.Flags().BoolP("list-supported-types", "L", false, "List supported hardware types.")

	AddBladeCmd.Flags().BoolP("geoloc", "g", false, "Require geolocation (LocationPaths, Ordinals, Parents, Children)")
	// Blades have several parents, so we need to add flags for each
	AddBladeCmd.Flags().IntVar(&cabinet, "cabinet", 1001, "Parent cabinet")
	AddBladeCmd.Flags().IntVar(&chassis, "chassis", 7, "Parent chassis")
	AddBladeCmd.Flags().IntVar(&blade, "blade", 1, "Blade")

	AddBladeCmd.Flags().BoolVar(&auto, "auto", false, "Automatically recommend values for parent hardware")
	AddBladeCmd.MarkFlagsRequiredTogether("list-supported-types")
	AddBladeCmd.MarkFlagsMutuallyExclusive("auto")
	AddBladeCmd.Flags().BoolVarP(&accept, "accept", "y", false, "Automatically accept recommended values.")

	RemoveBladeCmd.Flags().BoolVarP(&recursion, "recursive", "R", false, "Recursively delete child hardware")
	ListBladeCmd.Flags().StringVarP(&format, "format", "f", "pretty", "Format output")
	ListBladeCmd.Flags().StringVarP(&sortBy, "sort", "s", "location", "Sort by a specific key")

	// Register all provider commands during init()
	for _, p := range domain.GetProviders() {
		for _, c := range []*cobra.Command{AddBladeCmd, ListBladeCmd} {
			err := root.RegisterProviderCommand(p, c)
			if err != nil {
				log.Error().Msgf("Unable to get command '%s %s' from provider %s ", c.Parent().Name(), c.Name(), p.Slug())
				os.Exit(1)
			}
		}
	}
}
