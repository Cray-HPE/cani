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
	"os"

	"github.com/Cray-HPE/cani/internal/cani/domain"
	hardware_type_library "github.com/Cray-HPE/cani/pkg/hardware-type-library"
	"github.com/spf13/cobra"
)

// AddBladeCmd represents the blade add command
var AddBladeCmd = &cobra.Command{
	Use:   "blade",
	Short: "Add blades to the inventory.",
	Long:  `Add blades to the inventory.`,
	// Hardware can only be valid if defined in the hardware library
	Args: validHardware,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := addBlade(cmd, args)
		if err != nil {
			return err
		}
		return nil
	},
}

var (
	cabinet int
	chassis int
	slot    int
)

func init() {
	AddBladeCmd.Flags().BoolP("list-supported-types", "L", false, "List supported hardware types.")

	AddBladeCmd.Flags().IntVar(&cabinet, "cabinet", 0, "Parent cabinet")
	// cobra.MarkFlagRequired(AddBladeCmd.Flags(), "cabinet")

	AddBladeCmd.Flags().IntVar(&chassis, "chassis", 0, "Parent chassis")
	// cobra.MarkFlagRequired(AddBladeCmd.Flags(), "chassis")

	AddBladeCmd.Flags().IntVar(&slot, "slot", 0, "Parent slot")
	// cobra.MarkFlagRequired(AddBladeCmd.Flags(), "slot")
}

// addBlade adds a blade to the inventory
func addBlade(cmd *cobra.Command, args []string) error {
	// _, err := inventory.Add(cmd, args)
	// if err != nil {
	// 	return err
	// }
	// return nil

	// Setup domain logic
	d, err := domain.New()
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("list-supported-types") {
		d.ListSupportedTypes(hardware_type_library.HardwareTypeNodeBlade)
		os.Exit(0)
	}

	// Gather user supplied input data
	for _, arg := range args {
		return d.AddBlade(arg, cabinet, chassis, slot)
	}
	return nil
}
