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
	"encoding/json"
	"errors"
	"fmt"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// ListBladeCmd represents the blade list command
var ListBladeCmd = &cobra.Command{
	Use:   "blade",
	Short: "List blades in the inventory.",
	Long:  `List blades in the inventory.`,
	Args:  cobra.ArbitraryArgs,
	RunE:  listBlade,
}

// listBlade lists blades in the inventory
func listBlade(cmd *cobra.Command, args []string) error {
	// Instantiate a new logic object to interact with the datastore
	d, err := domain.New(root.Conf.Session.DomainOptions)
	if err != nil {
		return err
	}

	// Get the entire inventory
	inv, err := d.List()
	if err != nil {
		return err
	}

	// Filter the inventory to only blades
	filtered := make(map[uuid.UUID]inventory.Hardware, 0)

	// If no args are provided, list all blades
	if len(args) == 0 {
		for key, hw := range inv.Hardware {
			if hw.Type == hardwaretypes.HardwareTypeNodeBlade {
				filtered[key] = hw
			}
		}
	} else {
		// List each blade specified in the args
		for _, arg := range args {
			// Convert the argument to a UUID
			u, err := uuid.Parse(arg)
			if err != nil {
				return fmt.Errorf("Need a UUID: %s", err.Error())
			}
			// if blade does not exist, error
			if _, exists := inv.Hardware[u]; !exists {
				return fmt.Errorf("%s %s not found in inventory.", hardwaretypes.HardwareTypeNodeBlade, u)
			}
			// If the hardware is a blade
			if inv.Hardware[u].Type == hardwaretypes.HardwareTypeNodeBlade {
				// add it to the filtered inventory
				filtered[u] = inv.Hardware[u]
			} else {
				log.Debug().Msgf("%s is not a %s.  It is a %s", u, hardwaretypes.HardwareTypeNodeBlade, inv.Hardware[u].Type)
			}
		}
	}

	// Convert the filtered inventory into a formatted JSON string
	inventoryJSON, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return errors.New(fmt.Sprintf("Error marshaling inventory to JSON: %v", err))
	}

	// Print the inventory
	fmt.Println(string(inventoryJSON))
	return nil
}
