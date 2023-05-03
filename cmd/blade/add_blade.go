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
	"fmt"
	"strings"

	"github.com/Cray-HPE/cani/cmd/inventory"
	"github.com/spf13/cobra"
)

// AddBladeCmd represents the blade add command
var AddBladeCmd = &cobra.Command{
	Use:   "blade",
	Short: "Add blades to the inventory.",
	Long:  `Add blades to the inventory.`,
	Run: func(cmd *cobra.Command, args []string) {
		addBlade(args)
	},
}

var (
	listSupportedTypes bool
	hmnVlanId          int
	cabinetId          int
	chassis            int
	models             []string
	hwType             string
	slot               int
	port               int
	role               string
	subRole            string
)

func init() {
	// Get the supported hardware types
	supportedHw := inventory.SupportedHardware()
	// Append the blade types to the models slice for use in the --help output
	for _, hw := range supportedHw {
		if hw.Type == "ComputeModule" {
			models = append(models, hw.Model)
		}
	}
	AddBladeCmd.Flags().BoolVarP(&listSupportedTypes, "list-supported-types", "l", false, "List supported hardware types.")
	AddBladeCmd.Flags().StringVarP(&hwType, "type", "t", "", fmt.Sprintf("Hardware type.  Allowed values: [%+v]", strings.Join(models, "\", \"")))
	AddBladeCmd.Flags().IntVarP(&cabinetId, "cabinet", "C", 1000, "Cabinet ID")
	AddBladeCmd.Flags().IntVarP(&chassis, "chassis", "c", 0, "Chassis ID")
	AddBladeCmd.Flags().IntVarP(&slot, "slot", "s", 0, "Slot ID")
	AddBladeCmd.Flags().IntVarP(&hmnVlanId, "hmn-vlan", "v", 0, "HMN VLAN ID")
	AddBladeCmd.Flags().IntVarP(&port, "port", "p", 0, "Switchport")
	AddBladeCmd.Flags().StringVarP(&role, "role", "R", "", "Role")
	AddBladeCmd.Flags().StringVarP(&subRole, "sub-role", "r", "", "Sub-role")
}

// addBlade adds a blade to the inventory
func addBlade(args []string) error {
	fmt.Println("add blade called")
	return nil
}
