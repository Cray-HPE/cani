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
package cmd

import (
	"os"

	"github.com/Cray-HPE/cani/cmd/cabinet"
	"github.com/Cray-HPE/cani/cmd/chassis"
	"github.com/Cray-HPE/cani/cmd/hsn"
	"github.com/Cray-HPE/cani/cmd/inventory"
	"github.com/Cray-HPE/cani/cmd/pdu"
	sw "github.com/Cray-HPE/cani/cmd/switch"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// removeCmd represents the switch remove command
var RemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove assets from the inventory.",
	Long:  `Remove assets from the inventory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := remove(cmd, args)
		if err != nil {
			log.Error().Err(err).Msg(err.Error())
			os.Exit(1)
		}
		return err
	},
}

func init() {
	RemoveCmd.AddCommand(cabinet.RemoveCabinetCmd)
	RemoveCmd.AddCommand(chassis.RemoveChassisCmd)
	RemoveCmd.AddCommand(hsn.RemoveHsnCmd)
	RemoveCmd.AddCommand(pdu.RemovePduCmd)
	RemoveCmd.AddCommand(sw.RemoveSwitchCmd)
}

func remove(cmd *cobra.Command, args []string) error {
	_, err := inventory.Remove(cmd, args)
	if err != nil {
		return err
	}
	return nil
}
