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
	"os"
	"sort"

	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	csvFile string
)

// ImportCmd represents the import command
var ImportCmd = &cobra.Command{
	Use:   "import [FILE]",
	Short: "Import assets into the inventory.",
	Long:  `Import assets into the inventory.`,
	RunE:  importAssets,
}

func createCsvReader(filename string) (*csv.Reader, error) {
	if filename == "-" {
		return csv.NewReader(os.Stdin), nil
	} else {
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		return csv.NewReader(f), err
	}
}

// import is the main entry point for the update command.
func importAssets(cmd *cobra.Command, args []string) (err error) {
	if len(args) == 1 {
		csvFile = args[0]
	} else if len(args) == 0 {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			csvFile = "-"
		} else {
			return errors.New("missing the csv input. This can be either a file or standard input")
		}
	} else {
		return errors.New("too many arguments")
	}

	r, err := createCsvReader(csvFile)
	if err != nil {
		log.Error().Msgf("Failed to open the file %s. %s", csvFile, err)
		return nil
	}

	result, err := D.ImportCsv(cmd, args, r)
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
	log.Info().Msgf("Success: Wrote %d records of a total %d records from the CSV data", result.Modified, result.Total)
	return nil
}
