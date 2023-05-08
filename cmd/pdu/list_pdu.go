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
package pdu

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Cray-HPE/cani/cmd/inventory"
	"github.com/spf13/cobra"
)

// ListPduCmd represents the PDU add command
var ListPduCmd = &cobra.Command{
	Use:   "pdu",
	Short: "List PDUs in the inventory.",
	Long:  `List PDUs in the inventory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := listPdu(cmd, args)
		if err != nil {
			return err
		}
		return nil
	},
}

func listPdu(cmd *cobra.Command, args []string) error {
	inv, err := inventory.List(cmd, args)
	if err != nil {
		return err
	}

	filtered := inventory.Inventory{}
	for key, hw := range inv {
		if hw.Type == "PDU" {
			filtered[key] = hw
		}
	}
	// Convert the filtered inventory into a formatted JSON string
	inventoryJSON, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return errors.New(fmt.Sprintf("Error marshaling inventory to JSON: %v", err))
	}

	fmt.Println(string(inventoryJSON))
	return nil
}
