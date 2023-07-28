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
	"encoding/csv"
	"os"
	"strings"

	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	csvHeaders string
)

func init() {
	ExportCmd.PersistentFlags().StringVar(
		&csvHeaders, "headers", "Type,Vlan,Role,SubRole,Nid,Alias,Name,ID,Location", "Comma separated list of fields to get")
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

	headers := strings.Split(csvHeaders, ",")
	for i, header := range headers {
		headers[i] = strings.TrimSpace(header)
	}

	log.Debug().Msgf("headers: %v", headers)

	w := csv.NewWriter(os.Stdout)
	err = d.ExportCsv(cmd.Context(), w, headers)
	if err != nil {
		log.Error().Msgf("export failed: %s", err)
	}
	return nil
}
