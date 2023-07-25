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
	"errors"
	"sort"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// UpdateNodeCmd represents the node update command
var UpdateNodeCmd = &cobra.Command{
	Use:               "node",
	Short:             "Update nodes in the inventory.",
	Long:              `Update nodes in the inventory.`,
	PersistentPreRunE: root.DatastoreExists, // A session must be active to write to a datastore
	SilenceUsage:      true,                 // Errors are more important than the usage
	RunE:              updateNode,           // Update a node when this sub-command is called
}

// updateNode updates a node to the inventory
func updateNode(cmd *cobra.Command, args []string) error {
	// Create a domain object to interact with the datastore
	d, err := domain.New(root.Conf.Session.DomainOptions)
	if err != nil {
		return err
	}

	// Push all the CLI flags that were provided into a generic map
	// TODO Need to figure out how to specify to unset something
	// Right now the build metadata function in the CSM provider will
	// unset options if nil is passed in.
	nodeMeta := map[string]interface{}{}
	if cmd.Flags().Changed("role") {
		nodeMeta["role"] = role
	}
	if cmd.Flags().Changed("subrole") {
		nodeMeta["subrole"] = subrole
	}
	if cmd.Flags().Changed("alias") {
		nodeMeta["alias"] = alias
	}
	if cmd.Flags().Changed("alias") {
		nodeMeta["nid"] = nid
	}

	// Remove the node from the inventory using domain methods
	result, err := d.UpdateNode(cmd.Context(), cabinet, chassis, blade, nodecard, node, nodeMeta)
	if errors.Is(err, provider.ErrDataValidationFailure) {
		// TODO the following should probably suggest commands to fix the issue?
		log.Error().Msgf("Inventory data validation errors encountered")
		for id, failedValidation := range result.ProviderValidationErrors {
			log.Error().Msgf("  %s: %s", id, failedValidation.Hardware.LocationPath.String())
			sort.Strings(failedValidation.Errors)
			for _, validationError := range failedValidation.Errors {
				log.Error().Msgf("    - %s", validationError)
			}
		}

		return err
	} else if err != nil {
		return err
	}

	// TODO need a better identify, perhaps its UUID, or its location path?
	// log.Info().Msgf("Updated node %s", args[0])
	log.Info().Msgf("Updated node")
	return nil
}
