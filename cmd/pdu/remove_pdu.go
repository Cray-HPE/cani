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
	"os"

	"github.com/Cray-HPE/cani/cmd/inventory"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// RemovePduCmd represents the PDU remove command
var RemovePduCmd = &cobra.Command{
	Use:   "pdu",
	Short: "Remove PDUs from the inventory.",
	Long:  `Remove PDUs from the inventory.`,
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := removePdu(cmd, args)
		if err != nil {
			log.Error().Err(err).Msg(err.Error())
			os.Exit(1)
		}
		return err
	},
}

func removePdu(cmd *cobra.Command, args []string) error {
	_, err := inventory.Remove(cmd, args)
	if err != nil {
		return err
	}
	return nil
}
