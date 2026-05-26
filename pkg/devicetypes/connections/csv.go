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
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// csvInterface holds the fields extracted from one row of a Nautobot
// interfaces CSV export.
type csvInterface struct {
	ID         string // interface UUID
	Name       string // port / interface name
	DeviceName string // device__name
	CablePeer  string // UUID of the peer interface on the other end of the cable
	CablePK    string // UUID of the cable object (optional, used for dedup)
	Type       string // interface type (e.g. 100gbase-x-qsfp28)
	Status     string // status__name
	Label      string // label
}

// requiredInterfaceColumns lists columns required for Nautobot interfaces CSV.
var requiredInterfaceColumns = []string{"name", "device__name", "id", "cable_peer"}

// requiredConnectionColumns lists columns required for human-friendly CSV.
var requiredConnectionColumns = []string{"a_device", "a_port", "b_device", "b_port"}

// ParseCSV reads a CSV and auto-detects the format from the header row.
// If the header contains "a_device" it uses the human-friendly single-row
// parser; if it contains "cable_peer" it uses the Nautobot interfaces parser.
func ParseCSV(r io.Reader) (*ConnectionMap, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading CSV: %w", err)
	}

	// Peek at header to detect format
	header, err := csv.NewReader(bytes.NewReader(data)).Read()
	if err != nil {
		return nil, fmt.Errorf("reading CSV header: %w", err)
	}

	idx := buildColumnIndex(header)
	if _, ok := idx["a_device"]; ok {
		return ParseConnectionsCSV(bytes.NewReader(data))
	}
	if _, ok := idx["cable_peer"]; ok {
		return ParseInterfacesCSV(bytes.NewReader(data))
	}

	return nil, fmt.Errorf("CSV format not recognized: header must contain either 'a_device' (human) or 'cable_peer' (Nautobot interfaces)")
}

// ParseConnectionsCSV reads a human-friendly CSV where each row is one
// cable connection. No UUIDs required.
//
// Required columns: a_device, a_port, b_device, b_port.
// Optional columns: type, label, color, length, length_unit, status.
//
// A row with a_device set to "_defaults" is treated as cable defaults
// that apply to all connections (like cable_defaults in YAML). The
// type, color, status, and length_unit columns from the defaults row
// are applied to every connection that does not set them explicitly.
func ParseConnectionsCSV(r io.Reader) (*ConnectionMap, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("reading CSV header: %w", err)
	}

	idx := buildColumnIndex(header)
	if err := validateColumns(idx, requiredConnectionColumns); err != nil {
		return nil, err
	}

	var defaults *CableDefaults
	var entries []ConnectionEntry
	lineNum := 1 // header is line 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading CSV row %d: %w", lineNum+1, err)
		}
		lineNum++

		aDevice := getField(record, idx, "a_device")

		// Sentinel row: "_defaults" sets cable defaults for all rows
		if aDevice == "_defaults" {
			defaults = buildCableDefaults(record, idx)
			continue
		}

		aPort := getField(record, idx, "a_port")
		bDevice := getField(record, idx, "b_device")
		bPort := getField(record, idx, "b_port")

		if aDevice == "" || aPort == "" || bDevice == "" || bPort == "" {
			continue // skip incomplete rows
		}

		entry := ConnectionEntry{
			A: Endpoint{Device: aDevice, Port: aPort},
			B: Endpoint{Device: bDevice, Port: bPort},
		}

		// Attach optional cable properties
		cable := buildCableProps(record, idx)
		if cable != nil {
			entry.Cable = cable
		}

		entries = append(entries, entry)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no connections found in CSV (0 complete rows)")
	}

	return &ConnectionMap{
		Version:       "v1",
		CableDefaults: defaults,
		Connections:   entries,
	}, nil
}

// buildCableDefaults extracts cable default properties from a _defaults row.
func buildCableDefaults(record []string, idx map[string]int) *CableDefaults {
	d := &CableDefaults{
		Type:       getField(record, idx, "type"),
		Status:     getField(record, idx, "status"),
		Color:      getField(record, idx, "color"),
		LengthUnit: getField(record, idx, "length_unit"),
	}
	if d.Type == "" && d.Status == "" && d.Color == "" && d.LengthUnit == "" {
		return nil
	}
	return d
}

// buildCableProps extracts optional cable properties from a CSV row.
// Returns nil if no cable properties are present.
func buildCableProps(record []string, idx map[string]int) *CableProps {
	cType := getField(record, idx, "type")
	label := getField(record, idx, "label")
	color := getField(record, idx, "color")
	lengthStr := getField(record, idx, "length")
	lengthUnit := getField(record, idx, "length_unit")
	status := getField(record, idx, "status")

	if cType == "" && label == "" && color == "" && lengthStr == "" && status == "" {
		return nil
	}

	props := &CableProps{
		Type:       cType,
		Label:      label,
		Color:      color,
		LengthUnit: lengthUnit,
		Status:     status,
	}

	if lengthStr != "" {
		var length float64
		if _, err := fmt.Sscanf(lengthStr, "%f", &length); err == nil {
			props.Length = &length
		}
	}

	return props
}

// ParseInterfacesCSV reads a Nautobot interfaces CSV export and
// reconstructs a ConnectionMap by pairing cabled interfaces.
//
// Minimum required columns: name, device__name, id, cable_peer.
// Optional columns: cable__pk, type, status__name, label.
//
// Each cable appears twice in the CSV (once per side). The function
// deduplicates by keeping only the pair where side-A's id is
// lexicographically less than side-B's id.
func ParseInterfacesCSV(r io.Reader) (*ConnectionMap, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1 // allow variable column counts

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("reading CSV header: %w", err)
	}

	colIndex := buildColumnIndex(header)
	if err := validateColumns(colIndex, requiredInterfaceColumns); err != nil {
		return nil, err
	}

	interfaces, err := readInterfaceRows(reader, colIndex)
	if err != nil {
		return nil, err
	}

	return pairInterfaces(interfaces)
}

// buildColumnIndex maps column names to their position indices.
func buildColumnIndex(header []string) map[string]int {
	idx := make(map[string]int, len(header))
	for i, col := range header {
		idx[strings.TrimSpace(col)] = i
	}
	return idx
}

// validateColumns checks that all specified columns are present in the index.
func validateColumns(idx map[string]int, required []string) error {
	var missing []string
	for _, col := range required {
		if _, ok := idx[col]; !ok {
			missing = append(missing, col)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("CSV missing required columns: %s", strings.Join(missing, ", "))
	}
	return nil
}

// readInterfaceRows parses all data rows into csvInterface structs,
// skipping rows that have no cable_peer (uncabled interfaces).
func readInterfaceRows(reader *csv.Reader, idx map[string]int) (map[string]csvInterface, error) {
	interfaces := make(map[string]csvInterface)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading CSV row: %w", err)
		}

		iface := csvInterface{
			ID:         getField(record, idx, "id"),
			Name:       getField(record, idx, "name"),
			DeviceName: getField(record, idx, "device__name"),
			CablePeer:  getField(record, idx, "cable_peer"),
			CablePK:    getField(record, idx, "cable__pk"),
			Type:       getField(record, idx, "type"),
			Status:     getField(record, idx, "status__name"),
			Label:      getField(record, idx, "label"),
		}

		// Skip uncabled interfaces
		if iface.CablePeer == "" || iface.CablePeer == "NULL" {
			continue
		}
		if iface.ID == "" {
			continue
		}

		interfaces[iface.ID] = iface
	}

	return interfaces, nil
}

// pairInterfaces matches cabled interfaces by cable_peer UUID and
// builds a ConnectionMap. Deduplicates by keeping the pair where
// side-A has the lexicographically smaller ID.
func pairInterfaces(interfaces map[string]csvInterface) (*ConnectionMap, error) {
	seen := make(map[string]bool)
	var entries []ConnectionEntry

	for _, iface := range interfaces {
		peer, ok := interfaces[iface.CablePeer]
		if !ok {
			// Peer not in this export — skip (partial export)
			continue
		}

		// Deduplicate: each cable produces two rows. Keep the pair where
		// side-A's ID is lexicographically smaller.
		cableKey := iface.ID
		if iface.ID > peer.ID {
			cableKey = peer.ID
		}
		if seen[cableKey] {
			continue
		}
		seen[cableKey] = true

		// Ensure consistent ordering: smaller ID is side A
		a, b := iface, peer
		if a.ID > b.ID {
			a, b = b, a
		}

		entry := ConnectionEntry{
			A: Endpoint{Device: a.DeviceName, Port: a.Name},
			B: Endpoint{Device: b.DeviceName, Port: b.Name},
		}

		// Attach any available cable metadata
		if a.Label != "" || a.Status != "" {
			entry.Cable = &CableProps{
				Label:  a.Label,
				Status: a.Status,
			}
		}

		entries = append(entries, entry)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no cable connections found in CSV (0 interfaces with cable_peer)")
	}

	return &ConnectionMap{
		Version:     "v1",
		Connections: entries,
	}, nil
}

// getField safely retrieves a column value from a CSV record.
// Returns empty string if the column doesn't exist or the record is
// too short. Trims whitespace and treats "NULL" as empty.
func getField(record []string, idx map[string]int, col string) string {
	i, ok := idx[col]
	if !ok || i >= len(record) {
		return ""
	}
	val := strings.TrimSpace(record[i])
	if val == "NULL" {
		return ""
	}
	return val
}
