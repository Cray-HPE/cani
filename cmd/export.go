/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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
package cmd

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	csvHeaders        string
	csvComponentTypes string
	csvAllTypes       bool
	csvListOptions    bool
	exportFormat      string
	ignoreValidation  bool
)

func init() {
	ExportCmd.PersistentFlags().StringVar(
		&csvHeaders, "headers", "Type,Vlan,Role,SubRole,Status,Nid,Alias,Name,ID,Location", "Comma separated list of fields to get")
	ExportCmd.PersistentFlags().StringVarP(
		&csvComponentTypes, "type", "t", "Node,Cabinet", "Comma separated list of the types of components to output")
	ExportCmd.PersistentFlags().BoolVarP(&csvAllTypes, "all", "a", false, "List all components. This overrides the --type option")
	ExportCmd.PersistentFlags().BoolVarP(&csvListOptions, "list-fields", "L", false, "List details about the fields in the CSV")
	ExportCmd.PersistentFlags().StringVar(&exportFormat, "format", "csv", "Format option: [csv, sls-json]")
	ExportCmd.PersistentFlags().BoolVar(&ignoreValidation, "ignore-validation", false, "Skip validating the sls data. This only applies to the sls-json format.")
}

// ExportCmd represents the export command
var ExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export assets from the inventory.",
	Long:  `Export assets from the inventory.`,
	RunE:  export,
}

// export is the main entry point for the update command.
func export(cmd *cobra.Command, args []string) (err error) {
	switch exportFormat {
	case "csv":
		return exportCsv(cmd, args, D)
	case "sls-json":
		return exportJson(cmd, args, D, ignoreValidation)
	default:
		return fmt.Errorf("the requested format, %s, is unsupported", exportFormat)
	}
}

func exportCsv(cmd *cobra.Command, args []string, d *domain.Domain) error {
	if csvListOptions {
		err := d.ListCsvOptions(cmd.Context())
		if err != nil {
			return err
		}
	} else {
		headers := strings.Split(csvHeaders, ",")
		for i, header := range headers {
			headers[i] = strings.TrimSpace(header)
		}
		log.Debug().Msgf("headers: %v", headers)

		var types []string
		if csvAllTypes {
			// empty list means all types
			log.Debug().Msgf("types: all")
		} else {
			types = strings.Split(csvComponentTypes, ",")
			for i, t := range types {
				types[i] = strings.TrimSpace(t)
			}
			log.Debug().Msgf("types: %v", types)
		}

		w := csv.NewWriter(os.Stdout)
		err := d.ExportCsv(cmd.Context(), w, headers, types)
		if err != nil {
			return err
		}
	}
	return nil
}

func exportJson(cmd *cobra.Command, args []string, d *domain.Domain, ignoreValidation bool) error {
	cmd.SilenceUsage = true

	f := os.Stdout
	writer := bufio.NewWriter(f)
	defer writer.Flush()
	err := d.ExportJson(cmd, args, writer, ignoreValidation)
	if err != nil {
		return err
	}
	writer.Flush() // explicitly calling Flush here makes sure that any following log messages come after the sls json

	if ignoreValidation {
		log.Warn().Msg("Validation was not run. The SLS json may not be valid. Remove the --ignore-validate option to validate it.")
	}

	return nil
}
