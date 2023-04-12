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
	"embed"
	"fmt"
	"io/fs"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// addCabinetCmd represents the cabinet add command
var addCabinetCmd = &cobra.Command{
	Use:   "cabinet",
	Short: "Add cabinets to the inventory.",
	Long:  `Add cabinets to the inventory.`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// ensure the scripts are written to disk
		// TODO: running direct from memory takes a bit more effort in code
		writeHelperScriptsToDisk()

		// if addLiquidCooledCabinet {
		// 	_, stderr, err := shell("echo", []string{"test from shell command wrapper"})
		// 	if err != nil {
		// 		fmt.Println(stderr)
		// 		panic(err)
		// 	}
		// }
		// CreateNewContainer("hello-world")
		// ListContainers()
		addCabinet(args)
	},
}

var (
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
	addCmd.AddCommand(addCabinetCmd)

	// flags for running different helper scripts from the legacy procedures
	addCabinetCmd.Flags().BoolVarP(&addLiquidCooledCabinet, addLiqudCooledCabinetFlag, "a", false, fmt.Sprintf(runScriptUsage, addLiquidCooledCabinetName))
	addCabinetCmd.Flags().BoolVarP(&backupSlsPostgres, backupSlsPostgresFlag, "b", false, fmt.Sprintf(runScriptUsage, backupSlsPostgresName))
	addCabinetCmd.Flags().BoolVarP(&inspectSlsCabinets, inspectSlsCabinetsFlag, "i", false, fmt.Sprintf(runScriptUsage, inspectSlsCabinetsName))
	addCabinetCmd.Flags().BoolVarP(&updateNcnEtcHosts, updateNcnEtcHostsNameFlag, "u", false, fmt.Sprintf(runScriptUsage, updateNcnEtcHostsName))
	addCabinetCmd.Flags().BoolVarP(&updateNcnCabinetRoutes, updateNcnCabinetRoutesFlag, "U", false, fmt.Sprintf(runScriptUsage, updateNcnCabinetRoutesName))
	addCabinetCmd.Flags().BoolVarP(&verifyBmcCredentials, verifyBmcCredentialsFlag, "V", false, fmt.Sprintf(runScriptUsage, verifyBmcCredentialsName))
	// run each script independently
	addCabinetCmd.MarkFlagsMutuallyExclusive(addLiqudCooledCabinetFlag, backupSlsPostgresFlag, inspectSlsCabinetsFlag, updateNcnEtcHostsNameFlag, updateNcnCabinetRoutesFlag, verifyBmcCredentialsFlag)
}

func addCabinet(args []string) {
	fmt.Println("add cabinet called")
	for _, arg := range args {
		// code to add cabinet
		// ...

		if debug {
			log.Debug().Msgf("Added cabinet %s", arg)
		}
	}
}

// writeHelperScriptsToDisk writes the helper scripts for adding a cabinet to disk
func writeHelperScriptsToDisk() {
	// loop through all files in the helperScripts embed.FS
	files, err := fs.ReadDir(helperScripts, "scripts")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	for _, file := range files {
		src := fmt.Sprintf("scripts/%s", file.Name())
		content, err := helperScripts.ReadFile(src)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		fmt.Printf("Unpacking %s...\n", file.Name())
		dest := fmt.Sprintf("/tmp/%s", file.Name())
		err = os.WriteFile(dest, content, 0755)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
	}
}
