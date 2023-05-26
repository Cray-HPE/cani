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
package cabinet

import (
	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/cmd/session"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// AddCabinetCmd represents the cabinet add command
var AddCabinetCmd = &cobra.Command{
	Use:               "cabinet",
	Short:             "Add cabinets to the inventory.",
	Long:              `Add cabinets to the inventory.`,
	PersistentPreRunE: session.DatastoreExists, // A session must be active to write to a datastore
	Args:              validHardware,           // Hardware can only be valid if defined in the hardware library
	RunE:              addCabinet,              // Add a cabinet when this sub-command is called
}

// addCabinet adds a cabinet to the inventory
func addCabinet(cmd *cobra.Command, args []string) error {
	// Create a domain object to interact with the datastore
	d, err := domain.New(root.Conf.Session.DomainOptions)
	if err != nil {
		return err
	}

	// Add the blade from the inventory using domain methods
	results, err := d.AddCabinet(args[0], cabinet)
	if err != nil {
		return err
	}
	log.Info().Msgf("Added cabinet %s", args[0])

	// Use a map to track already added nodes.
	newNodes := []domain.AddHardwareResult{}

	for _, result := range results {
		// If the type is a Node
		if result.Hardware.Type == hardwaretypes.HardwareTypeCabinet {
			log.Debug().Msg(result.Location.String())
			log.Debug().Msgf("This %s also contains a %s (%s) added at %s",
				hardwaretypes.HardwareTypeNodeBlade,
				hardwaretypes.HardwareTypeNode,
				result.Hardware.ID.String(),
				result.Location)
			// Add the node to the map
			newNodes = append(newNodes, result)
		}
	}

	return nil
}
