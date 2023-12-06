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
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	csvAllowedHeaders = map[string]string{
		"id":             "ID",
		"location":       "Location",
		"name":           "Name",
		"type":           "Type",
		"devicetypeslug": "DeviceTypeSlug",
		"status":         "Status",
		"vlan":           "Vlan",
		"role":           "Role",
		"subrole":        "SubRole",
		"alias":          "Alias",
		"nid":            "Nid"}
)

func (csm *CSM) Export(cmd *cobra.Command, args []string, datastore inventory.Datastore) error {
	switch exportFormat {
	case "csv":
		return csm.exportCsv(cmd, args, datastore)
	case "sls-json":
		return csm.exportJson(cmd, args, datastore, ignoreValidation)
	default:
		return fmt.Errorf("the requested format, %s, is unsupported", exportFormat)
	}
}

func (csm *CSM) exportCsv(cmd *cobra.Command, args []string, datastore inventory.Datastore) error {
	if csvListOptions {
		err := csm.ListCsvOptions(cmd.Context())
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
		err := csm.ExportCsv(cmd.Context(), datastore, w, headers, types)
		if err != nil {
			return err
		}
	}
	return nil
}

func (csm *CSM) exportJson(cmd *cobra.Command, args []string, datastore inventory.Datastore, ignoreValidation bool) error {
	cmd.SilenceUsage = true

	f := os.Stdout
	writer := bufio.NewWriter(f)
	defer writer.Flush()
	err := csm.ExportJson2(cmd, args, datastore, writer, ignoreValidation)
	if err != nil {
		return err
	}
	writer.Flush() // explicitly calling Flush here makes sure that any following log messages come after the sls json

	if ignoreValidation {
		log.Warn().Msg("Validation was not run. The SLS json may not be valid. Remove the --ignore-validate option to validate it.")
	}

	return nil
}

func (csm *CSM) ListCsvOptions(ctx context.Context) error {
	metadata, err := csm.GetFieldMetadata()
	if err != nil {
		return err
	}

	minwidth := 0         // minimal cell width including any padding
	tabwidth := 8         // width of tab characters (equivalent number of spaces)
	padding := 1          // padding added to a cell before computing its width
	padchar := byte('\t') // ASCII char used for padding

	w := tabwriter.NewWriter(os.Stdout, minwidth, tabwidth, padding, padchar, tabwriter.AlignRight)
	defer w.Flush()

	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "Name", "Types", "Modifiable", "Description")
	for _, m := range metadata {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", m.Name, m.Types, strconv.FormatBool(m.IsModifiable), m.Description)
	}
	return nil
}

func (csm *CSM) ExportCsv(ctx context.Context, datastore inventory.Datastore, writer *csv.Writer, headers []string, types []string) error {
	// Get the entire inventory
	inv, err := datastore.List()
	if err != nil {
		return errors.Join(fmt.Errorf("failed to read the hardware from the database"), err)
	}

	keys := make([]uuid.UUID, 0, len(inv.Hardware))
	for key := range inv.Hardware {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		hwi := inv.Hardware[keys[i]]
		hwj := inv.Hardware[keys[j]]
		return inventory.CompareHardwareByTypeThenLocation(&hwi, &hwj)
	})

	normalizedHeaders, err := toNormalizedHeaders(headers)
	if err != nil {
		return errors.Join(fmt.Errorf("invalid headers %v, allowed headers: %v", headers, csvAllowedHeaders), err)
	}

	typeSet := make(map[string]struct{})
	for _, t := range types {
		typeSet[strings.ToLower(t)] = struct{}{}
	}
	allTypes := len(types) == 0

	// Write the first csv row (i.e. the headers)
	writer.Write(normalizedHeaders)

	for _, uuid := range keys {
		hw := inv.Hardware[uuid]
		if _, ok := typeSet[strings.ToLower(string(hw.Type))]; !allTypes && !ok {
			continue
		}
		row, err := csm.GetFields(&hw, normalizedHeaders)
		if err != nil {
			return errors.Join(fmt.Errorf("unexpected error getting fields, %v, from hardware %v", normalizedHeaders, hw.ID), err)
		}

		writer.Write(row)
		if err != nil {
			return errors.Join(fmt.Errorf("failed to write csv data"), err)
		}
		writer.Flush()
	}
	return nil
}

func (csm *CSM) ExportJson2(cmd *cobra.Command, args []string, datastore inventory.Datastore, writer io.Writer, skipValidation bool) error {
	exportedJson, err := csm.ExportJson(cmd, args, datastore, skipValidation)
	if err != nil {
		return err
	}
	writer.Write(exportedJson)
	writer.Write([]byte("\n"))

	return nil
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

func getNextRow(reader *csv.Reader) ([]string, error) {
	row, err := reader.Read()
	if err != nil {
		return nil, err
	}
	return row, nil
}

func getNextRowAsMap(reader *csv.Reader, headers []string) (map[string]string, error) {
	values := make(map[string]string)
	row, err := reader.Read()
	if err != nil {
		return values, err
	}
	columnCount := len(headers)
	for i, value := range row {
		if i < columnCount {
			values[headers[i]] = value
		} else {
			return values,
				fmt.Errorf("row had more columns than the header. Expected columns %d. Row lenth %d. Row %v",
					columnCount,
					len(row),
					row)
		}
	}
	return values, nil
}

// return the headers with there correct capitalization
func toNormalizedHeaders(headers []string) ([]string, error) {
	normalizedHeaders := make([]string, len(headers))
	var err error
	for i, header := range headers {
		h, found := csvAllowedHeaders[strings.ToLower(header)]
		if !found {
			err = errors.Join(err, fmt.Errorf("invalid header: %s", header))
		}
		normalizedHeaders[i] = h
	}
	return normalizedHeaders, err
}
