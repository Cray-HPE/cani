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
package node

import (
	"os"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	cabinet               int
	chassis               int
	blade                 int
	nodecard              int
	node                  int
	nodeUuid              string
	format                string
	sortBy                string
	ProviderAddNodeCmd    = &cobra.Command{}
	ProviderListNodeCmd   = &cobra.Command{}
	ProviderUpdateNodeCmd = &cobra.Command{}
)

func init() {
	var err error
	// Add variants to root commands
	root.AddCmd.AddCommand(AddNodeCmd)
	root.ListCmd.AddCommand(ListNodeCmd)
	root.RemoveCmd.AddCommand(RemoveNodeCmd)
	root.UpdateCmd.AddCommand(UpdateNodeCmd)

	// Add a flag to show supported types
	AddNodeCmd.Flags().BoolP("list-supported-types", "L", false, "List supported hardware types.")

	// Merge CANI's command with the provider-specified command
	// this allows for CANI's operations to remain consistent, while adding provider config on top
	err = root.MergeProviderCommand(AddNodeCmd, ProviderAddNodeCmd)
	if err != nil {
		log.Error().Msgf("%+v", err)
		os.Exit(1)
	}

	// Blades have several parents, so we need to add flags for each
	UpdateNodeCmd.Flags().IntVar(&cabinet, "cabinet", 1001, "Parent cabinet")
	UpdateNodeCmd.Flags().IntVar(&chassis, "chassis", 7, "Parent chassis")
	UpdateNodeCmd.Flags().IntVar(&blade, "blade", 1, "Parent blade")
	UpdateNodeCmd.Flags().IntVar(&nodecard, "nodecard", 1, "Parent node card")
	UpdateNodeCmd.Flags().IntVar(&node, "node", 1, "Node to update")
	UpdateNodeCmd.Flags().StringVar(&nodeUuid, "uuid", "", "UUID of the node to update")

	UpdateNodeCmd.MarkFlagsRequiredTogether("cabinet", "chassis", "blade", "nodecard", "node")
	UpdateNodeCmd.MarkFlagsMutuallyExclusive("uuid")
	ListNodeCmd.Flags().StringVarP(&format, "format", "f", "pretty", "Format output")
	ListNodeCmd.Flags().StringVarP(&sortBy, "sort", "s", "location", "Sort by a specific key")

	// Merge CANI's command with the provider-specified command
	// this allows for CANI's operations to remain consistent, while adding provider config on top
	err = root.MergeProviderCommand(UpdateNodeCmd, ProviderUpdateNodeCmd)
	if err != nil {
		log.Error().Msgf("%+v", err)
		os.Exit(1)
	}
	err = root.MergeProviderCommand(ListNodeCmd, ProviderListNodeCmd)
	if err != nil {
		log.Error().Msgf("%+v", err)
		os.Exit(1)
	}
}
