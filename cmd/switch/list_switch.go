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
package sw

import (
	"encoding/json"
	"errors"
	"fmt"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// ListSwitchCmd represents the switch list command
var ListSwitchCmd = &cobra.Command{
	Use:   "switch",
	Short: "List switches in the inventory.",
	Long:  `List switches in the inventory.`,
	Args:  cobra.ArbitraryArgs,
	RunE:  listSwitch,
}

// listSwitch lists switches in the inventory
func listSwitch(cmd *cobra.Command, args []string) error {
	// Create a domain object to interact with the datastore
	d, err := domain.New(root.Conf.Session.DomainOptions)
	if err != nil {
		return err
	}

	// Get the entire inventory
	inv, err := d.List()
	if err != nil {
		return err
	}
	// Filter the inventory to only switches
	filtered := make(map[uuid.UUID]inventory.Hardware, 0)
	for key, hw := range inv.Hardware {
		if hw.Type == hardwaretypes.HighSpeedSwitch || hw.Type == hardwaretypes.ManagementSwitch {
			filtered[key] = hw
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
