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
package cabinet

import (
	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// ListCabinetCmd represents the cabinet list command
var ListCabinetCmd = &cobra.Command{
	Use:   "cabinet PROVIDER",
	Short: "List cabinets in the inventory.",
	Long:  `List cabinets in the inventory.`,
	Args:  cobra.ExactArgs(1),
	RunE:  listCabinet,
}

// listCabinet lists cabinets in the inventory
func listCabinet(cmd *cobra.Command, args []string) error {
	// Get the entire inventory
	inv, err := root.D.List()
	if err != nil {
		return err
	}

	// Filter the inventory to only cabinets
	filtered := make(map[uuid.UUID]inventory.Hardware, 0)
	for key, hw := range inv.Hardware {
		if hw.Type == hardwaretypes.Cabinet {
			filtered[key] = hw
		}
	}

	err = root.D.PrintHardware(cmd, args, filtered)
	if err != nil {
		return err
	}

	return nil
}
