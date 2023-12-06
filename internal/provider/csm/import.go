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
package csm

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func (csm *CSM) Import(cmd *cobra.Command, args []string, datastore inventory.Datastore) error {
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

	result, err := csm.ImportCsv(cmd, args, datastore, r)
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

func (csm *CSM) ImportCsv(cmd *cobra.Command, args []string, datastore inventory.Datastore, reader *csv.Reader) (result provider.CsvImportResult, err error) {
	tempDatastore, err := datastore.Clone()
	if err != nil {
		return result, err
	}

	headers, err := getNextRow(reader)
	if err == io.EOF {
		return result, fmt.Errorf("the CSV file is empty")
	}
	if err != nil {
		return result, err
	}

	foundIDHeader := false
	for _, header := range headers {
		if header == "ID" {
			foundIDHeader = true
		}
	}
	if !foundIDHeader {
		return result, fmt.Errorf("ID column is missing")
	}

	for {
		row, err := getNextRowAsMap(reader, headers)
		if err == io.EOF {
			break
		}
		if err != nil {
			return result, err
		}

		result.Total++

		idStr, ok := row["ID"]
		if !ok {
			return result, fmt.Errorf("missing ID for row %d", result.Total+1)
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			return result, errors.Join(fmt.Errorf("failed to parse %v as a UUID", idStr), err)
		}

		hw, err := tempDatastore.Get(id)
		if err != nil {
			return result, errors.Join(fmt.Errorf("could not find hardware with the UUID %v. This call can only be used to update existing hardware", id), err)
		}

		setResult, err := csm.SetFields(&hw, row)
		if err != nil {
			return result,
				errors.Join(fmt.Errorf("unexpected error setting fields, %v, from hardware %v", row, hw.ID), err)
		}

		if len(setResult.ModifiedFields) > 0 {
			log.Debug().Msgf("Updated %v modifying the fields: %v", id, setResult.ModifiedFields)
			err = tempDatastore.Update(&hw)
			if err != nil {
				return result, errors.Join(fmt.Errorf("failed to write to the database the hardware %v", id), err)
			}
			result.Modified++
		}
	}

	if result.Modified > 0 {
		results, err := csm.ValidateInternal(cmd, args, tempDatastore, false)
		if err != nil {
			result.ValidationResults = results
			return result, err
		}

		if err := datastore.Merge(tempDatastore); err != nil {
			return result, errors.Join(fmt.Errorf("failed to merge temporary datastore with actual datastore"), err)
		}
		if err := datastore.Flush(); err != nil {
			return result, errors.Join(fmt.Errorf("failed to write datastore to disk"), err)
		}
	}
	return result, nil
}
