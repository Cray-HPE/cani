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
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

var (
	csvAllowedHeaders = map[string]string{
		"id":             "ID",
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

func (csm *CSM) ExportCsv(ctx context.Context, datastore inventory.Datastore, writer *csv.Writer, headers []string) error {
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
		ki := fmt.Sprintf("%v", keys[i])
		kj := fmt.Sprintf("%v", keys[j])
		return ki < kj
	})

	normalizedHeaders, err := toNormalizedHeaders(headers)
	if err != nil {
		return errors.Join(fmt.Errorf("invalid headers %v, allowed headers: %v", headers, csvAllowedHeaders), err)
	}

	// Write the first csv row (i.e. the headers)
	writer.Write(normalizedHeaders)

	row := make([]string, len(normalizedHeaders))
	for _, uuid := range keys {
		hw := inv.Hardware[uuid]
		rawCsmProps := hw.ProviderProperties["csm"]
		csmProps, ok := rawCsmProps.(map[string]interface{})
		if !ok {
			csmProps = make(map[string]interface{})
		}

		for i, header := range normalizedHeaders {
			switch header {
			case "ID":
				row[i] = fmt.Sprintf("%v", hw.ID)
			case "Name":
				row[i] = fmt.Sprintf("%v", hw.Name)
			case "Type":
				row[i] = fmt.Sprintf("%v", hw.Type)
			case "DeviceTypeSlug":
				row[i] = fmt.Sprintf("%v", hw.DeviceTypeSlug)
			case "Status":
				row[i] = fmt.Sprintf("%v", hw.Status)
			case "Vlan":
				row[i] = toString(csmProps["HMNVlan"])
			case "Role":
				row[i] = toString(csmProps["Role"])
			case "SubRole":
				row[i] = toString(csmProps["SubRole"])
			case "Alias":
				row[i] = toStringArray(csmProps["Alias"], 0)
			case "Nid":
				row[i] = toString(csmProps["Nid"])
			default:
				// This case should never be hit.
				// The call to normalize should return an error for unknown headers
				log.Error().Msgf("Unknown header %s", header)
				row[i] = ""
			}

		}
		writer.Write(row)
		if err != nil {
			return errors.Join(fmt.Errorf("failed to write csv data"), err)
		}
		writer.Flush()
	}
	return nil
}

func (csm *CSM) ImportCsv(ctx context.Context, datastore inventory.Datastore, reader *csv.Reader) (result provider.CsvImportResult, err error) {
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

		logMessage := fmt.Sprintf("Updated %v modifying the fields:", id)

		rawCsmProps := hw.ProviderProperties["csm"]
		csmProps, ok := rawCsmProps.(map[string]interface{})
		if !ok {
			// NodeCard's do not have csm props
			// todo possibly verify that the writable columns are empty in the csv input
			log.Debug().Msgf("Skipping %v of the type %v. It does not have writable properties", id, hw.Type)
			continue
		}

		// fields: Vlan,Role,SubRole,Alias,Nid
		hasChanges := false

		vlanStr, ok := row["Vlan"]
		if ok && vlanStr != "" {
			// todo should vlanStr == "" cause the "HMNVlan" field to be removed?
			vlan, err := strconv.ParseInt(vlanStr, 10, 64)
			if err != nil {
				return result, err
			}
			current := csmProps["HMNVlan"]
			if current != vlan {
				logMessage += " Vlan,"
				hasChanges = true
				csmProps["HMNVlan"] = vlan
			}
		}

		role, ok := row["Role"]
		if ok && role != "" {
			if role != csmProps["Role"] {
				logMessage += " Role,"
				hasChanges = true
				csmProps["Role"] = role
			}
		}

		subRole, ok := row["SubRole"]
		if ok {
			currentSubRole, ok := csmProps["SubRole"]
			if subRole == "" {
				if ok {
					if nil != currentSubRole && subRole != currentSubRole {
						logMessage += " SubRole,"
						hasChanges = true
						csmProps["SubRole"] = nil
					}
				}
			} else {
				if subRole != currentSubRole {
					logMessage += " SubRole,"
					hasChanges = true
					csmProps["SubRole"] = subRole
				}
			}
		}

		alias, ok := row["Alias"]
		if ok && alias != "" {
			rawAlias, ok := csmProps["Alias"]
			if !ok {
				logMessage += " Alias,"
				hasChanges = true
				var a [1]string
				a[0] = alias
				csmProps["Alias"] = a
			} else {
				v, ok := rawAlias.([]interface{})
				if !ok {
					return result, fmt.Errorf("expected the Alias field to be an array in the hardware %v", hw)
				}
				if len(v) > 0 {
					if v[0] != alias {
						logMessage += " Alias,"
						hasChanges = true
						v[0] = alias
					}
				} else {
					logMessage += " Alias,"
					hasChanges = true
					v = append(v, alias)
				}
				csmProps["Alias"] = v
			}
		}

		nidStr, ok := row["Nid"]
		if ok && nidStr != "" {
			nid, err := strconv.ParseInt(nidStr, 10, 64)
			if err != nil {
				return result, errors.Join(fmt.Errorf("failed to parse nid %v in row %d", nidStr, result.Total+1), err)
			}
			currentNidRaw := csmProps["Nid"]
			currentNid, ok := currentNidRaw.(float64)
			if !ok || float64(nid) != currentNid {
				logMessage += " nid,"
				hasChanges = true
				csmProps["Nid"] = nid
			}
		}

		if hasChanges {
			log.Debug().Msgf(logMessage)
			err = tempDatastore.Update(&hw)
			if err != nil {
				return result, errors.Join(fmt.Errorf("failed to write to the database the hardware %v", id), err)
			}
			result.Modified++
		}
	}

	if result.Modified > 0 {
		results, err := csm.ValidateInternal(ctx, tempDatastore, true)
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

func toStringArray(value interface{}, i int) string {
	if value == nil {
		return ""
	}
	v, ok := value.([]interface{})
	if !ok {
		return ""
	}
	if len(v) <= i {
		return ""
	}
	return fmt.Sprintf("%v", v[i])
}

func toString(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
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
