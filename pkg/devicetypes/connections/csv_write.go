/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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
package connections

import (
	"encoding/csv"
	"fmt"
	"io"
)

// csvHeader is the column order for human-friendly connection CSV files.
var csvHeader = []string{
	"a_device", "a_port", "b_device", "b_port",
	"type", "label", "color", "length", "length_unit", "status",
}

// WriteCSV writes a ConnectionMap as a human-friendly CSV to w.
// The format is round-trippable with ParseConnectionsCSV.
func WriteCSV(w io.Writer, cm ConnectionMap) error {
	writer := csv.NewWriter(w)

	if err := writer.Write(csvHeader); err != nil {
		return fmt.Errorf("writing CSV header: %w", err)
	}

	if cm.CableDefaults != nil {
		row := []string{
			"_defaults", "", "", "",
			cm.CableDefaults.Type,
			"",
			cm.CableDefaults.Color,
			"",
			cm.CableDefaults.LengthUnit,
			cm.CableDefaults.Status,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("writing CSV defaults row: %w", err)
		}
	}

	for _, entry := range cm.Connections {
		row := []string{
			entry.A.Device,
			entry.A.Port,
			entry.B.Device,
			entry.B.Port,
			"", "", "", "", "", "",
		}

		if entry.Cable != nil {
			row[4] = entry.Cable.Type
			row[5] = entry.Cable.Label
			row[6] = entry.Cable.Color
			if entry.Cable.Length != nil {
				row[7] = formatLength(*entry.Cable.Length)
			}
			row[8] = entry.Cable.LengthUnit
			row[9] = entry.Cable.Status
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("writing CSV row: %w", err)
		}
	}

	writer.Flush()
	return writer.Error()
}

// formatLength formats a float64 length, using integer notation when possible.
func formatLength(v float64) string {
	if v == float64(int64(v)) {
		return fmt.Sprintf("%d", int64(v))
	}
	return fmt.Sprintf("%g", v)
}
