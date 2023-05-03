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
package cabinet

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/Cray-HPE/cani/cmd/inventory"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// AddCabinetCmd represents the cabinet add command
var AddCabinetCmd = &cobra.Command{
	Use:   "cabinet",
	Short: "Add cabinets to the inventory.",
	Long:  `Add cabinets to the inventory.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := addCabinet(args)
		if err != nil {
			log.Error().Err(err).Msg(err.Error())
			os.Exit(1)
		}
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
	//go:embed scripts/*
	helperScripts              embed.FS
	addLiquidCooledCabinet     bool
	addLiquidCooledCabinetName = "add_liquid_cooled_cabinet.py"
	addLiqudCooledCabinetFlag  = "add-liquid-cooled-cabinet"
	backupSlsPostgres          bool
	backupSlsPostgresName      = "backup_sls_postgres.sh"
	backupSlsPostgresFlag      = "backup-sls-postgres"
	inspectSlsCabinets         bool
	inspectSlsCabinetsName     = "inspect_sls_cabinets.py"
	inspectSlsCabinetsFlag     = "inspect-sls-cabinets"
	updateNcnEtcHosts          bool
	updateNcnEtcHostsName      = "update_ncn_etc_hosts.py"
	updateNcnEtcHostsNameFlag  = "update-ncn-etc-hosts"
	updateNcnCabinetRoutes     bool
	updateNcnCabinetRoutesName = "update_ncn_cabinet_routes.py"
	updateNcnCabinetRoutesFlag = "update-ncn-cabinet-routes"
	verifyBmcCredentials       bool
	verifyBmcCredentialsName   = "verify_bmc_credentials.sh"
	verifyBmcCredentialsFlag   = "verify-bmc-credentials"
	runScriptUsage             = "Run the %s script"
)

func init() {
	supportedHw := inventory.SupportedHardware()
	for _, hw := range supportedHw {
		models = append(models, hw.Model)
	}
	AddCabinetCmd.Flags().BoolVarP(&listSupportedTypes, "list-supported-types", "l", false, "List supported hardware types.")
	AddCabinetCmd.Flags().StringVarP(&hwType, "type", "t", "", fmt.Sprintf("Hardware type.  Allowed values: [%+v]", strings.Join(models, "\", \"")))
	AddCabinetCmd.Flags().IntVarP(&cabinetId, "cabinet", "C", 1000, "Cabinet ID")
	AddCabinetCmd.Flags().IntVarP(&chassis, "chassis", "c", 0, "Chassis ID")
	AddCabinetCmd.Flags().IntVarP(&slot, "slot", "s", 0, "Slot ID")
	AddCabinetCmd.Flags().IntVarP(&hmnVlanId, "hmn-vlan", "v", 0, "HMN VLAN ID")
}

func addCabinet(args []string) error {
	fmt.Println("add cabinet called")
	return nil
}
