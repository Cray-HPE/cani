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
package hsn

import (
	"fmt"
	"os"
	"strings"

	"github.com/Cray-HPE/cani/cmd/inventory"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	listSupportedTypes bool
	horizontalU        string
	models             []string
	hwType             string
	id                 string
	port               int
	cabinet            int
)

// AddHsnCmd represents the HSN add command
var AddHsnCmd = &cobra.Command{
	Use:   "hsn",
	Short: "Add high speed hardware to the inventory.",
	Long:  `Add high speed hardware to the inventory.`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		err := addHsn(args)
		if err != nil {
			log.Error().Err(err).Msg(err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	supportedHw := inventory.SupportedHardware()
	for _, hw := range supportedHw {
		models = append(models, hw.Model)
	}
	AddHsnCmd.Flags().BoolVarP(&listSupportedTypes, "list-supported-types", "l", false, "List supported hardware types.")
	AddHsnCmd.Flags().StringVarP(&hwType, "type", "t", "", fmt.Sprintf("Hardware type.  Allowed values: [%+v]", strings.Join(models, "\", \"")))
	AddHsnCmd.Flags().IntVarP(&cabinet, "cabinet", "C", 1000, "Cabinet ID.")
	AddHsnCmd.Flags().StringVarP(&id, "id", "i", "L", "ID")
	AddHsnCmd.Flags().IntVarP(&port, "port", "p", 0, "Switchport")
}

// addHsn adds a high speed network device to the inventory.
func addHsn(args []string) error {
	fmt.Println("add hsn called")
	return nil
}
