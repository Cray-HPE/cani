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
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	csvFile string
)

func init() {
	ImportCmd.PersistentFlags().StringVarP(&csvFile, "file", "f", "", "Path to the data file")
}

// ImportCmd represents the import command
var ImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import assets into the inventory.",
	Long:  `Import assets into the inventory.`,
	RunE:  importAssets,
}

// import is the main entry point for the update command.
func importAssets(cmd *cobra.Command, args []string) error {
	// Create a domain object to interact with the datastore
	d, err := domain.New(Conf.Session.DomainOptions)
	if err != nil {
		log.Error().Msgf("Import CSV failed internal error. %s", err)
		return nil
	}

	f, err := os.Open(csvFile)
	if err != nil {
		log.Error().Msgf("Failed to open the file %s. %s", csvFile, err)
		return nil
	}

	r := csv.NewReader(f)
	result, err := d.ImportCsv(cmd.Context(), r)
	if errors.Is(err, provider.ErrDataValidationFailure) {
		log.Error().Msgf("The changes are invalid.")
		for id, failedValidation := range result.ValidationResults {
			log.Error().Msgf("  %s: %s", id, failedValidation.Hardware.LocationPath.String())
			sort.Strings(failedValidation.Errors)
			for _, validationError := range failedValidation.Errors {
				log.Error().Msgf("    - %s", validationError)
			}
		}
		return nil
	} else if err != nil {
		log.Error().Msgf("import failed. %s", err)
		return nil
	}
	fmt.Printf("Success: Wrote %d records of a total %d records in the CSV data\n", result.Modified, result.Total)
	return nil
}
