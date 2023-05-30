/*
MIT License

(C) Copyright 2022 Hewlett Packard Enterprise Development LP

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/
package blade

import (
	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/cmd/session"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// AddBladeCmd represents the blade add command
var AddBladeCmd = &cobra.Command{
	Use:               "blade",
	Short:             "Add blades to the inventory.",
	Long:              `Add blades to the inventory.`,
	PersistentPreRunE: session.DatastoreExists, // A session must be active to write to a datastore
	SilenceUsage:      true,                    // Errors are more important than the usage
	Args:              validHardware,           // Hardware can only be valid if defined in the hardware library
	RunE:              addBlade,                // Add a blade when this sub-command is called
}

// addBlade adds a blade to the inventory
func addBlade(cmd *cobra.Command, args []string) error {
	// Create a domain object to interact with the datastore
	d, err := domain.New(root.Conf.Session.DomainOptions)
	if err != nil {
		return err
	}

	// Add the blade from the inventory using domain methods
	results, err := d.AddBlade(args[0], cabinet, chassis, slot)
	if err != nil {
		return err
	}
	log.Info().Msgf("Added blade %s", args[0])

	// Gather info about the parent node
	// inv, err := d.List()
	// if err != nil {
	// 	return err
	// }

	// A blade can have 1 or more nodes following this heirarchy:
	//
	// | hardwaretypes.Cabinet
	// |-- hardwaretypes.Chassis
	// |---- hardwaretypes.NodeBlade
	// |------ hardwaretypes.NodeCard
	// |-------- hardwaretypes.Node
	//
	// After adding a blade, we need to find the node(s) that were added to present the user
	// with the node(s) that may need additional metadata added

	// Use a map to track already added nodes.
	newNodes := []domain.AddHardwareResult{}

	for _, result := range results {
		// If the type is a Node
		if result.Hardware.Type == hardwaretypes.Node {
			log.Debug().Msg(result.Location.String())
			log.Debug().Msgf("This %s also contains a %s (added %s)",
				hardwaretypes.NodeBlade,
				hardwaretypes.Node,
				result.Hardware.ID.String())
			log.Debug().Msgf("This %s also contains a %s (%s) added at %s",
				hardwaretypes.NodeBlade,
				hardwaretypes.Node,
				result.Hardware.ID.String(),
				result.Location)
			// Add the node to the map
			newNodes = append(newNodes, result)
		}
	}

	if root.Conf.Session.DomainOptions.Provider == string(inventory.CSMProvider) {
		log.Info().Msgf("For provider '%s', additional metadata is needed for each %s in the %s:\n\n",
			root.Conf.Session.DomainOptions.Provider,
			hardwaretypes.Node,
			hardwaretypes.NodeBlade)
	}
	// for _, newNode := range newNodes {
	// 	cabinet := newNode.Location[0].Ordinal // bo.Cabinet()
	// 	chassis := newNode.Location[1].Ordinal // bo.Chassis()
	// 	slot := newNode.Location[2].Ordinal    // bo.Slot()
	// 	bmc := newNode.Location[3].Ordinal     // bo.BMC() // aka NodeCard
	// 	node := newNode.Location[4].Ordinal    // bo.Node()
	// 	log.Info().Msgf("cani update node --cabinet \"%d\" --chassis \"%d\" --slot \"%d\" --bmc \"%d\" --node \"%d\" --role \"FIXME\" --subrole \"FIXME\" --alias \"FIXME\" --nid \"FIXME\"\n",
	// 		cabinet,
	// 		chassis,
	// 		slot,
	// 		bmc,
	// 		node)
	// }

	return nil
}
