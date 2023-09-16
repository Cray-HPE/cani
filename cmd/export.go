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
	validateSls       bool
)

func init() {
	ExportCmd.PersistentFlags().StringVar(
		&csvHeaders, "headers", "Type,Vlan,Role,SubRole,Status,Nid,Alias,Name,ID,Location", "Comma separated list of fields to get")
	ExportCmd.PersistentFlags().StringVarP(
		&csvComponentTypes, "type", "t", "Node,Cabinet", "Comma separated list of the types of components to output")
	ExportCmd.PersistentFlags().BoolVarP(&csvAllTypes, "all", "a", false, "List all components. This overrides the --type option")
	ExportCmd.PersistentFlags().BoolVarP(&csvListOptions, "list-fields", "L", false, "List details about the fields in the CSV")
	ExportCmd.PersistentFlags().StringVar(&exportFormat, "format", "csv", "Format option: csv or sls-json")
	ExportCmd.PersistentFlags().BoolVar(&validateSls, "validate", false, "Validate the SLS json. This only applies to the sls-json format.")
}

// ExportCmd represents the export command
var ExportCmd = &cobra.Command{
	Use:               "export",
	Short:             "Export assets from the inventory.",
	Long:              `Export assets from the inventory.`,
	PersistentPreRunE: DatastoreExists,
	RunE:              export,
}

// export is the main entry point for the update command.
func export(cmd *cobra.Command, args []string) error {
	// Create a domain object to interact with the datastore
	d, err := domain.New(Conf.Session.DomainOptions)
	if err != nil {
		return err
	}

	switch exportFormat {
	case "csv":
		return exportCsv(cmd, args, d)
	case "sls-json":
		return exportSlsJson(cmd, args, d, validateSls)
	default:
		return fmt.Errorf("the requested format, %s, is unsupported", exportFormat)
	}
}

func exportCsv(cmd *cobra.Command, args []string, d *domain.Domain) error {
	if csvListOptions {
		err := d.ListCsvOptions(cmd.Context(), Conf.Session.DomainOptions)
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

func exportSlsJson(cmd *cobra.Command, args []string, d *domain.Domain, validate bool) error {
	cmd.SilenceUsage = true

	if !validate {
		log.Warn().Msg("The SLS json is not being validated. Use the --validate option to validate it.")
	}

	f := os.Stdout
	writer := bufio.NewWriter(f)
	defer writer.Flush()
	err := d.ExportSls(cmd.Context(), writer, validate)
	if err != nil {
		return err
	}
	return nil
}
