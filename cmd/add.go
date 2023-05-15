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
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Cray-HPE/cani/cmd/blade"
	"github.com/Cray-HPE/cani/cmd/cabinet"
	"github.com/Cray-HPE/cani/cmd/chassis"
	"github.com/Cray-HPE/cani/cmd/hsn"
	"github.com/Cray-HPE/cani/cmd/inventory"
	"github.com/Cray-HPE/cani/cmd/node"
	"github.com/Cray-HPE/cani/cmd/pdu"
	sw "github.com/Cray-HPE/cani/cmd/switch"
)

var (
	vendor string
	name   string
	staged string
	models []string
	hwType string
	u      string
)

// addCmd represents the switch add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add assets to the inventory.",
	Long:  `Add assets to the inventory.`,
	Run: func(cmd *cobra.Command, args []string) {
		if simulation {
			blade.AddBladeCmd.SetArgs([]string{"-S"})
		}
		if len(args) == 0 {
			fmt.Println("Error: No asset type specified.")
			cmd.Help()
			os.Exit(1)
		}
	},
}

func init() {
	supportedHw := inventory.SupportedHardware()
	for _, hw := range supportedHw {
		models = append(models, hw.Model)
	}

	addCmd.AddCommand(blade.AddBladeCmd)
	addCmd.AddCommand(cabinet.AddCabinetCmd)
	addCmd.AddCommand(chassis.AddChassisCmd)
	addCmd.AddCommand(hsn.AddHsnCmd)
	addCmd.AddCommand(node.AddNodeCmd)
	addCmd.AddCommand(pdu.AddPduCmd)
	addCmd.AddCommand(sw.AddSwitchCmd)
	addCmd.PersistentFlags().StringVarP(&vendor, "vendor", "m", "HPE", "Vendor")
	addCmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Name")
	addCmd.PersistentFlags().StringVarP(&staged, "staged", "s", "Staged", "Hardware can be [staged, provisioned, decomissioned]")
	addCmd.PersistentFlags().StringVarP(&hwType, "type", "t", "", fmt.Sprintf("Hardware type.  Allowed values: [%+v]", strings.Join(models, "\", \"")))
	addCmd.PersistentFlags().StringVarP(&u, "uuid", "u", "", "Specific UUID to use")
	addCmd.PersistentFlags().BoolP("list-supported-types", "L", false, "List supported hardware types.")
}
