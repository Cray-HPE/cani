package node

import (
	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/spf13/cobra"

	"github.com/Cray-HPE/cani/cmd/session"
	"github.com/rs/zerolog/log"
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
	nodeMeta := map[string]interface{}{
		"role":    role,
		"subrole": subrole,
		"alias":   alias,
		"nid":     nid,
	}

	// Remove the node from the inventory using domain methods
	err = d.UpdateNode(cabinet, chassis, slot, bmc, node, nodeMeta)
	if err != nil {
		return err
	}
	log.Info().Msgf("Updated node %s", args[0])
	return nil
}
