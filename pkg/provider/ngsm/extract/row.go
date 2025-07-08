/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
package extract

import (
	"encoding/csv"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/Cray-HPE/cani/internal/core"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// Row is a structured row from the CSV
type Row struct {
	ItemNumber         string  `csv:"Item#"`
	Quantity           int     `csv:"Qty"`
	ProductNumber      string  `csv:"Product #"`
	ProductDescription string  `csv:"Product Description"`
	ConfigName         string  `csv:"Config Name"`
	SolutionId         int     `csv:"Solution Id"`
	SupportFor         string  `csv:"Support For"`
	ClicStatus         string  `csv:"CLIC Status"`
	UnitPrice          float64 `csv:"Unit Price (USD) "`
	ExtendedListPrice  float64 `csv:"Extended List Price (USD)"`
	ExpertBomComment   string  `csv:"Expert BOM Comment"`
	Notes              string  `csv:"NOTES"`
	NetboxName         string
	Source             string
	Row                int
	NetboxDeviceTypeID int32
	DeviceTypeSlug     string
	HardwareType       devicetypes.Type
}

// Sanitize attempts to sanitize the row's content so valid data is returned
// this will do things like normalize part numbers, clean up blankspace, etc.
func (row *Row) Sanitize() (parsed *Row, err error) {
	// sanitize the row's content
	err = row.sanitize()
	if err != nil {
		return nil, err
	}

	return row, nil
}

// sanitize sanitizes an individual row in the bom
// At present, it only sanitizes the product number
// but other input validation could be added here
func (row *Row) sanitize() error {
	row.ProductNumber = row.sanitizeProductNumber()
	// log.Printf("Sanitized product number: %+v", row.ProductNumber)
	// get the device type IDs
	slug := core.Slugify(row.ProductDescription)
	if slug != "" {
		row.DeviceTypeSlug = slug
	}
	// log.Printf("Sanitized description to slug: %+v", row.DeviceTypeSlug)
	return nil
}

// sanitizeProductNumber converts all blankspace in a product number to a
// single '-', making it easier for parsing and comparison
func (row *Row) sanitizeProductNumber() (productNumber string) {
	// convert all whitespace to a single '-' character
	re := regexp.MustCompile(`\s+`)
	productNumber = re.ReplaceAllString(row.ProductNumber, "-")
	return productNumber
}

// SetNetboxName sets the netbox name for the row
func (row *Row) SetNetboxName(name string) {
	row.NetboxName = name
}

// SetSource sets the source of the row, which is the usually the CSV filename
func (row *Row) SetSource(path string) {
	row.Source = path
}

// SetNetboxDeviceTypeID sets the netbox device type ID for the row
func (row *Row) SetNetboxDeviceTypeID(id int32) {
	row.NetboxDeviceTypeID = id
}

// SetNetboxDeviceTypeSlug sets the netbox device type slug for the row
func (row *Row) SetNetboxDeviceTypeSlug(slug string) {
	row.DeviceTypeSlug = slug
}

// SetRow sets the row number for the row
func (row *Row) SetRow(n int) {
	row.Row = n
}

// IsRack returns true if the row is a rack based on some fuzzy regex matching
func (row *Row) IsRack() bool {
	// FIXME: do more fuzzy/regex checking
	// or choose from pre-existing rack part numbers
	// need multiple regexes because single there is no lookahead and lookbehind
	// assertions in the regexp package
	re1 := regexp.MustCompile(`\d+U`)
	re2 := regexp.MustCompile(`[Rr]ack`)
	re3 := regexp.MustCompile(`(?i)[Kk]it|[Ss]ervice`)
	// if a row has both 'rack', a U count, and does not contain 'kit' or 'service', it is likely a rack
	if re1.MatchString(row.ProductDescription) && re2.MatchString(row.ProductDescription) && !re3.MatchString(row.ProductDescription) {
		return true
	} else {
		return false
	}
}

// RackDimensions returns the width and depth of a rack in millimeters parsed
// via regex from the product description 'NNNmmxNNNmm'
func (row *Row) RackDimensions() (width int, depth int, unit string, err error) {
	re := regexp.MustCompile(`\d+mmx\d+mm`)
	dimensions := re.FindString(row.ProductDescription)
	if dimensions != "" {
		wh := strings.Split(dimensions, "x")
		w := strings.TrimSuffix(wh[0], "mm")
		d := strings.TrimSuffix(wh[1], "mm")
		width, err = strconv.Atoi(w)
		if err != nil {
			return 0, 0, "", err
		}
		depth, err = strconv.Atoi(d)
		if err != nil {
			return 0, 0, "", err
		}
	}
	return width, depth, "mm", nil
}

// RackHeightU returns the height of a rack in U parsed via regex from the product description
func (row *Row) RackHeightU() (u int, err error) {
	re := regexp.MustCompile(`\d+U`)
	height := re.FindString(row.ProductDescription)
	if height != "" {
		ustring := strings.TrimSuffix(height, "U")
		u, err = strconv.Atoi(ustring)
		if err != nil {
			return 0, err
		}
	}
	return u, nil
}

// IsDevice returns true if the row is a device based on the product number
// At present, this is the only method to determine if a row is a device since
// netbox requires the device type in order to add a device
// this could be expanded to include alternative methods to match other fields
// in the device type yaml format
func (row *Row) IsDevice() bool {
	slug := core.Slugify(row.ProductDescription)
	_, ok := devicetypes.All()[slug]
	return ok
}

// NewRowFromRow returns a new row from the existing row
// this is often used to create a duplicate row if the quantity is more than 1
// a few fields need to be set by the consumer of this function, like the netbox
// name, source, etc.
func (row *Row) NewRowFromRow() (newRow Row) {
	newRow = Row{}
	newRow.SetRow(row.Row)
	newRow.ItemNumber = row.ItemNumber
	newRow.Quantity = 0
	newRow.ProductNumber = row.ProductNumber
	newRow.ProductDescription = row.ProductDescription
	newRow.ConfigName = row.ConfigName
	newRow.SolutionId = row.SolutionId
	newRow.SupportFor = row.SupportFor
	newRow.ClicStatus = row.ClicStatus
	newRow.UnitPrice = row.UnitPrice
	newRow.ExtendedListPrice = row.ExtendedListPrice
	newRow.ExpertBomComment = row.ExpertBomComment
	newRow.Notes = row.Notes
	newRow.HardwareType = devicetypes.All()[row.DeviceTypeSlug].Type
	return newRow
}

func (row *Row) NewDeviceFromRow() (dev devicetypes.CaniDeviceType) {
	dev = devicetypes.CaniDeviceType{}
	dev.ID = uuid.New()
	dev.Name = row.NetboxName
	dev.Type = row.HardwareType
	dev.DeviceTypeSlug = core.Slugify(row.ProductDescription)
	dev.Vendor = core.Slugify("HPE")
	// dev.Architecture =
	dev.Model = row.ProductNumber
	dev.Status = "staged"
	dev.Properties = make(map[string]any)
	dev.Children = []uuid.UUID{}
	// lp := inventory.LocationPath{}
	// dev.LocationPath = lp
	// lo := int(0)
	// dev.LocationOrdinal = &lo

	// FIXME: make metadata a type or mapstructure
	// needed for the provider to know where the device came from so it can be added into the rack
	dev.ProviderMetadata = make(map[string]any)
	dev.ProviderMetadata["ngsm"] = map[string]any{
		"Source": row.Source,
		"Row":    row.Row,
	}

	return dev
}

// getRowsFromBom unmarshals the rows from the bom file into a slice of rows
// this retains the rows order, which helps in adding things to the Queue
//
//		TODO: gosv offers other unmarshal options, which may be helpful in
//		parsing the bom files, which are not completely standard and may have
//	  extra info at the top, instead of a header row
func getRowsFromBom(bom string) (rows []*Row, err error) {
	// open the bom file
	data, err := os.OpenFile(bom, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return rows, err
	}
	defer data.Close()

	// Create a new CSV reader
	reader := csv.NewReader(data)

	// You can adjust CSV parsing settings if needed
	reader.TrimLeadingSpace = true
	// reader.LazyQuotes = true
	// reader.FieldsPerRecord = -1 // Allow variable number of fields

	// Read all records at once
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// First row should be the header
	if len(records) == 0 {
		return []*Row{}, nil
	}

	headers := records[0]
	rows = make([]*Row, 0, len(records)-1)

	// Process each data row
	for i, record := range records[1:] {
		row := &Row{}

		// Map CSV fields to your struct properties
		// This depends on your Row struct definition
		for j, value := range record {
			if j >= len(headers) {
				continue
			}

			// Map each header to the corresponding field
			switch headers[j] {
			case "Item#":
				row.ItemNumber = value
			case "Qty":
				if value == "" {
					// log.Printf("Quantity is empty in BOM file '%s', setting to 1", bom)
					continue
				}
				// Convert the quantity to an integer
				quantity, err := strconv.Atoi(value)
				if err != nil {
					log.Fatalf("Error converting quantity '%s' to int: %v", value, err)
					return nil, err
				}
				row.Quantity = quantity
			case "Product #":
				row.ProductNumber = value
			case "Product Description":
				row.ProductDescription = value
			case "Config Name":
				row.ConfigName = value
			case "Solution Id":
				if value == "" {
					// log.Printf("Quantity is empty in BOM file '%s', setting to 1", bom)
					continue
				}
				// Convert the solution ID to an integer
				solutionId, err := strconv.Atoi(value)
				if err != nil {
					log.Fatalf("Error converting solution ID '%s' to int: %v", value, err)
					return nil, err
				}
				row.SolutionId = solutionId
			case "Support For":
				row.SupportFor = value
			case "CLIC Status":
				row.ClicStatus = value
			case "Unit Price (USD) ":
				if value == "" {
					// log.Printf("Quantity is empty in BOM file '%s', setting to 1", bom)
					continue
				}
				// Convert the unit price to a float64
				unitPrice, err := strconv.ParseFloat(value, 64)
				if err != nil {
					log.Fatalf("Error converting unit price '%s' to float64: %v", value, err)
					return nil, err
				}
				row.UnitPrice = unitPrice
			case "Extended List Price (USD)":
				if value == "" {
					// log.Printf("Quantity is empty in BOM file '%s', setting to 1", bom)
					continue
				}
				// Convert the extended list price to a float64
				extendedListPrice, err := strconv.ParseFloat(value, 64)
				if err != nil {
					log.Fatalf("Error converting extended list price '%s' to float64: %v", value, err)
					return nil, err
				}
				row.ExtendedListPrice = extendedListPrice
			case "Expert BOM Comment":
				row.ExpertBomComment = value
			case "NOTES":
				row.Notes = value
			default:
				// log.Printf("Unknown header '%s' in BOM file '%s'", headers[j], bom)
				// Skip unknown headers or handle them as needed
				continue

			}
		}

		// Set the row number
		row.SetRow(i + 2) // +1 for header, +1 for 0-based indexing
		rows = append(rows, row)
	}
	return rows, nil
}
