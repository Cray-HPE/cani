package import_

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// SystemRecord holds a parsed row from a system CSV file.
// The Section field determines which other fields are relevant.
type SystemRecord struct {
	Section      string // role, rack, device, module, connection
	PartNumber   string
	Name         string
	Qty          int
	Rack         string // parent rack name (devices)
	Position     int    // U position in rack (devices)
	Face         string // front, rear (devices)
	Role         string // role name (devices)
	Status       string
	Serial       string
	Device       string // parent device name (modules)
	Bay          string // module bay (modules)
	ADevice      string // connection endpoint A device
	APort        string // connection endpoint A port
	BDevice      string // connection endpoint B device
	BPort        string // connection endpoint B port
	Color        string // cable color (connections)
	Length       string // cable length value (connections)
	LengthUnit   string // cable length unit (connections)
	Location     string // location name (racks)
	ContentTypes string // comma-separated content types (roles)
	MacAddress   string // MAC address (interfaces)
}

// SystemCSV holds parsed system CSV data grouped by section.
type SystemCSV struct {
	Defaults        SystemRecord            // global _defaults row
	SectionDefaults map[string]SystemRecord // per-section defaults (e.g. device_defaults)
	Roles           []SystemRecord
	Locations       []SystemRecord
	Racks           []SystemRecord
	Devices         []SystemRecord
	Modules         []SystemRecord
	Interfaces      []SystemRecord
	Connections     []SystemRecord
}

// systemColumnIndex maps header positions for system CSV columns.
type systemColumnIndex struct {
	section      int
	partNumber   int
	name         int
	qty          int
	rack         int
	position     int
	face         int
	role         int
	status       int
	serial       int
	device       int
	bay          int
	aDevice      int
	aPort        int
	bDevice      int
	bPort        int
	color        int
	length       int
	lengthUnit   int
	location     int
	contentTypes int
	macAddress   int
}

// IsSystemCSV returns true if the header row contains a "Section" column,
// indicating this is a system CSV rather than a traditional BOM CSV.
func IsSystemCSV(header []string) bool {
	for _, col := range header {
		if normalizeSystemHeader(col) == "section" {
			return true
		}
	}
	return false
}

// ParseSystemCSV reads a system CSV file and returns grouped records.
func ParseSystemCSV(filePath string) (*SystemCSV, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open system CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	reader.Comment = '#'

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read system CSV: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("system CSV must have a header row and at least one data row")
	}

	idx, err := parseSystemHeader(rows[0])
	if err != nil {
		return nil, err
	}

	result := &SystemCSV{
		SectionDefaults: make(map[string]SystemRecord),
	}

	for lineNum, row := range rows[1:] {
		rec, err := parseSystemRow(row, idx, lineNum+2)
		if err != nil {
			log.Printf("WARN: system CSV line %d: %v, skipping", lineNum+2, err)
			continue
		}

		section := strings.ToLower(rec.Section)

		// Handle defaults rows
		if section == "_defaults" {
			result.Defaults = rec
			continue
		}
		if strings.HasSuffix(section, "_defaults") {
			key := strings.TrimSuffix(section, "_defaults")
			result.SectionDefaults[key] = rec
			continue
		}

		switch section {
		case "role":
			result.Roles = append(result.Roles, rec)
		case "location":
			result.Locations = append(result.Locations, rec)
		case "rack":
			result.Racks = append(result.Racks, rec)
		case "device":
			result.Devices = append(result.Devices, rec)
		case "module":
			result.Modules = append(result.Modules, rec)
		case "interface":
			result.Interfaces = append(result.Interfaces, rec)
		case "connection":
			result.Connections = append(result.Connections, rec)
		default:
			log.Printf("WARN: system CSV line %d: unknown section %q, skipping", lineNum+2, rec.Section)
		}
	}

	return result, nil
}

// parseSystemHeader maps header columns to a systemColumnIndex.
func parseSystemHeader(header []string) (systemColumnIndex, error) {
	idx := systemColumnIndex{
		section:      -1,
		partNumber:   -1,
		name:         -1,
		qty:          -1,
		rack:         -1,
		position:     -1,
		face:         -1,
		role:         -1,
		status:       -1,
		serial:       -1,
		device:       -1,
		bay:          -1,
		aDevice:      -1,
		aPort:        -1,
		bDevice:      -1,
		bPort:        -1,
		color:        -1,
		length:       -1,
		lengthUnit:   -1,
		location:     -1,
		contentTypes: -1,
		macAddress:   -1,
	}

	for i, col := range header {
		switch normalizeSystemHeader(col) {
		case "section":
			idx.section = i
		case "partnumber":
			idx.partNumber = i
		case "name":
			idx.name = i
		case "qty":
			idx.qty = i
		case "rack":
			idx.rack = i
		case "position":
			idx.position = i
		case "face":
			idx.face = i
		case "role":
			idx.role = i
		case "status":
			idx.status = i
		case "serial":
			idx.serial = i
		case "device":
			idx.device = i
		case "bay":
			idx.bay = i
		case "adevice":
			idx.aDevice = i
		case "aport":
			idx.aPort = i
		case "bdevice":
			idx.bDevice = i
		case "bport":
			idx.bPort = i
		case "color":
			idx.color = i
		case "length":
			idx.length = i
		case "lengthunit":
			idx.lengthUnit = i
		case "location":
			idx.location = i
		case "contenttypes":
			idx.contentTypes = i
		case "macaddress":
			idx.macAddress = i
		}
	}

	if idx.section < 0 {
		return idx, fmt.Errorf("missing required column: Section")
	}

	return idx, nil
}

// normalizeSystemHeader converts column header variations to canonical names.
func normalizeSystemHeader(col string) string {
	lower := strings.ToLower(strings.TrimSpace(col))
	lower = strings.ReplaceAll(lower, "_", "")
	lower = strings.ReplaceAll(lower, "-", "")
	lower = strings.ReplaceAll(lower, " ", "")

	switch lower {
	case "section", "type", "recordtype":
		return "section"
	case "partnumber", "productnumber", "slug", "partno":
		return "partnumber"
	case "name", "devicename", "rackname":
		return "name"
	case "qty", "quantity", "count":
		return "qty"
	case "rack", "rackid", "parentrack":
		return "rack"
	case "position", "uposition", "u", "rackunit":
		return "position"
	case "face", "rackface", "side":
		return "face"
	case "role", "devicerole":
		return "role"
	case "status":
		return "status"
	case "serial", "serialnumber", "sn":
		return "serial"
	case "device", "parentdevice":
		return "device"
	case "bay", "modulebay", "slot":
		return "bay"
	case "adevice", "devicea", "sourcedevice":
		return "adevice"
	case "aport", "porta", "sourceport":
		return "aport"
	case "bdevice", "deviceb", "destdevice":
		return "bdevice"
	case "bport", "portb", "destport":
		return "bport"
	case "color", "cablecolor":
		return "color"
	case "length", "cablelength":
		return "length"
	case "lengthunit", "unit":
		return "lengthunit"
	case "location", "loc", "site":
		return "location"
	case "contenttypes", "contenttype":
		return "contenttypes"
	case "macaddress", "mac", "macaddr", "hwaddr", "hardwareaddress":
		return "macaddress"
	default:
		return lower
	}
}

// parseSystemRow extracts a SystemRecord from a CSV row.
func parseSystemRow(row []string, idx systemColumnIndex, lineNum int) (SystemRecord, error) {
	section := getColumnValue(row, idx.section)
	if section == "" {
		return SystemRecord{}, fmt.Errorf("empty Section")
	}

	qtyStr := getColumnValue(row, idx.qty)
	qty := 1
	if qtyStr != "" {
		var err error
		qty, err = strconv.Atoi(qtyStr)
		if err != nil {
			return SystemRecord{}, fmt.Errorf("invalid Qty %q: %w", qtyStr, err)
		}
		if qty < 1 {
			return SystemRecord{}, fmt.Errorf("Qty must be >= 1, got %d", qty)
		}
	}

	posStr := getColumnValue(row, idx.position)
	pos := 0
	if posStr != "" {
		var err error
		pos, err = strconv.Atoi(posStr)
		if err != nil {
			return SystemRecord{}, fmt.Errorf("invalid Position %q: %w", posStr, err)
		}
	}

	return SystemRecord{
		Section:      section,
		PartNumber:   getColumnValue(row, idx.partNumber),
		Name:         getColumnValue(row, idx.name),
		Qty:          qty,
		Rack:         getColumnValue(row, idx.rack),
		Position:     pos,
		Face:         getColumnValue(row, idx.face),
		Role:         getColumnValue(row, idx.role),
		Status:       getColumnValue(row, idx.status),
		Serial:       getColumnValue(row, idx.serial),
		Device:       getColumnValue(row, idx.device),
		Bay:          getColumnValue(row, idx.bay),
		ADevice:      getColumnValue(row, idx.aDevice),
		APort:        getColumnValue(row, idx.aPort),
		BDevice:      getColumnValue(row, idx.bDevice),
		BPort:        getColumnValue(row, idx.bPort),
		Color:        getColumnValue(row, idx.color),
		Length:       getColumnValue(row, idx.length),
		LengthUnit:   getColumnValue(row, idx.lengthUnit),
		Location:     getColumnValue(row, idx.location),
		ContentTypes: getColumnValue(row, idx.contentTypes),
		MacAddress:   getColumnValue(row, idx.macAddress),
	}, nil
}

// ApplyDefaults merges defaults into a record. Fields already set on rec
// take precedence. Global defaults are applied first, then section-specific.
func (s *SystemCSV) ApplyDefaults(rec SystemRecord) SystemRecord {
	section := strings.ToLower(rec.Section)

	// Apply global defaults first
	rec = mergeSystemDefaults(rec, s.Defaults)

	// Apply section-specific defaults (override global)
	if sd, ok := s.SectionDefaults[section]; ok {
		rec = mergeSystemDefaults(rec, sd)
	}

	return rec
}

// mergeSystemDefaults fills empty fields in rec from defaults.
func mergeSystemDefaults(rec, defaults SystemRecord) SystemRecord {
	if rec.Status == "" && defaults.Status != "" {
		rec.Status = defaults.Status
	}
	if rec.Role == "" && defaults.Role != "" {
		rec.Role = defaults.Role
	}
	if rec.Face == "" && defaults.Face != "" {
		rec.Face = defaults.Face
	}
	if rec.Location == "" && defaults.Location != "" {
		rec.Location = defaults.Location
	}
	if rec.Color == "" && defaults.Color != "" {
		rec.Color = defaults.Color
	}
	if rec.LengthUnit == "" && defaults.LengthUnit != "" {
		rec.LengthUnit = defaults.LengthUnit
	}
	return rec
}
