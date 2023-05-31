package node

import (
	"errors"
	"sort"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/Cray-HPE/cani/cmd/session"
)

// UpdateNodeCmd represents the node update command
var UpdateNodeCmd = &cobra.Command{
	Use:               "node",
	Short:             "Update nodes in the inventory.",
	Long:              `Update nodes in the inventory.`,
	PersistentPreRunE: session.DatastoreExists, // A session must be active to write to a datastore
	SilenceUsage:      true,                    // Errors are more important than the usage
	// Args:              validHardware,           // Hardware can only be valid if defined in the hardware library
	RunE: updateNode, // Update a node when this sub-command is called
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
	result, err := d.UpdateNode(cmd.Context(), cabinet, chassis, slot, bmc, node, nodeMeta)
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
