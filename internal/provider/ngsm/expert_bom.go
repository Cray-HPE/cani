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
package ngsm

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
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
	netboxName         string
	source             string
	row                int
	rack               string // when a device, the rack it will go in
	netboxDeviceTypeID int32
}

// sanitize sanitizes an individual row in the bom
// At present, it only sanitizes the product number
// but other input validation could be added here
func (row *Row) sanitize() error {
	row.ProductNumber = row.sanitizeProductNumber()
	log.Trace().Msgf("Sanitized product number: %+v", row.ProductNumber)
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
	row.netboxName = name
}

// SetSource sets the source of the row, which is the usually the CSV filename
func (row *Row) SetSource(path string) {
	row.source = path
}

// SetNetboxDeviceTypeID sets the netbox device type ID for the row
func (row *Row) SetNetboxDeviceTypeID(id int32) {
	row.netboxDeviceTypeID = id
}

// SetRow sets the row number for the row
func (row *Row) SetRow(n int) {
	row.row = n
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

// RackDimensions returns the width and depth of a rack in millimeters parsed
// via regex from the product description 'NNNmmxNNNmm'
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
func (row *Row) IsDevice(deviceTypeIds map[string]int32) bool {
	_, ok := deviceTypeIds[row.ProductNumber]
	if ok {
		row.SetNetboxDeviceTypeID(deviceTypeIds[row.ProductNumber])
		return true
	}

	return false
}

// NewRowFromRow returns a new row from the existing row
// this is often used to create a duplicate row if the quantity is more than 1
// a few fields need to be set by the consumer of this function, like the netbox
// namne, source, etc.
func (row *Row) NewRowFromRow() (newRow Row) {
	newRow = Row{}
	newRow.SetRow(row.row)
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
	return newRow
}
