package import_

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// CsvRecord holds a parsed row from the CSV file.
// Contains only raw CSV column data; no inferred or computed fields.
type CsvRecord struct {
	PartNumber  string
	Description string
	Quantity    int
	ConfigGroup string // optional, e.g., "0100", "0200"
	// Cable fields (optional)
	SourceDevice string // device name for cable termination A
	SourcePort   string // interface name on source device
	DestDevice   string // device name for cable termination B
	DestPort     string // interface name on destination device
	CableType    string // e.g., "cat6", "dac-passive"
	CableLength  string // e.g., "2m", "10ft"
}

// columnIndex maps normalized header names to column indices.
type columnIndex struct {
	partNumber   int
	description  int
	quantity     int
	configGroup  int // -1 if not present
	sourceDevice int // -1 if not present
	sourcePort   int // -1 if not present
	destDevice   int // -1 if not present
	destPort     int // -1 if not present
	cableType    int // -1 if not present
	cableLength  int // -1 if not present
}

// ParseCSV reads and parses a CSV file into CsvRecord slices.
// Skips malformed rows with warnings.
func ParseCSV(filePath string) ([]CsvRecord, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("CSV must have a header row and at least one data row")
	}

	idx, err := parseHeader(rows[0])
	if err != nil {
		return nil, err
	}

	var records []CsvRecord
	for lineNum, row := range rows[1:] {
		record, err := parseRow(row, idx, lineNum+2) // +2 for 1-based and header offset
		if err != nil {
			log.Printf("WARN: line %d: %v, skipping", lineNum+2, err)
			continue
		}
		records = append(records, record)
	}

	return records, nil
}

// parseHeader maps header columns to indices, normalizing variations.
func parseHeader(header []string) (columnIndex, error) {
	idx := columnIndex{
		partNumber:   -1,
		description:  -1,
		quantity:     -1,
		configGroup:  -1,
		sourceDevice: -1,
		sourcePort:   -1,
		destDevice:   -1,
		destPort:     -1,
		cableType:    -1,
		cableLength:  -1,
	}

	for i, col := range header {
		normalized := normalizeHeader(col)
		switch normalized {
		case "partnumber":
			idx.partNumber = i
		case "description":
			idx.description = i
		case "quantity":
			idx.quantity = i
		case "configgroup":
			idx.configGroup = i
		case "sourcedevice":
			idx.sourceDevice = i
		case "sourceport":
			idx.sourcePort = i
		case "destdevice":
			idx.destDevice = i
		case "destport":
			idx.destPort = i
		case "cabletype":
			idx.cableType = i
		case "cablelength":
			idx.cableLength = i
		}
	}

	if idx.partNumber < 0 {
		return idx, fmt.Errorf("missing required column: PartNumber (or ProductNumber, HPEProductNumber)")
	}
	if idx.description < 0 {
		return idx, fmt.Errorf("missing required column: Description (or Name, ProductDescription)")
	}
	if idx.quantity < 0 {
		return idx, fmt.Errorf("missing required column: Quantity (or Qty)")
	}

	return idx, nil
}

// normalizeHeader converts header variations to canonical names.
func normalizeHeader(col string) string {
	lower := strings.ToLower(strings.TrimSpace(col))
	lower = strings.ReplaceAll(lower, "_", "")
	lower = strings.ReplaceAll(lower, "-", "")
	lower = strings.ReplaceAll(lower, " ", "")

	switch lower {
	case "partnumber", "productnumber", "hpeproductnumber", "part", "product":
		return "partnumber"
	case "description", "name", "productdescription", "productname", "desc":
		return "description"
	case "quantity", "qty", "count", "orderquantity":
		return "quantity"
	case "configgroup", "configitemnumber", "config", "group":
		return "configgroup"
	case "sourcedevice", "srcdevice", "devicea":
		return "sourcedevice"
	case "sourceport", "srcport", "porta":
		return "sourceport"
	case "destdevice", "destinationdevice", "deviceb":
		return "destdevice"
	case "destport", "destinationport", "portb":
		return "destport"
	case "cabletype", "type":
		return "cabletype"
	case "cablelength", "length":
		return "cablelength"
	default:
		return lower
	}
}

// parseRow extracts a CsvRecord from a row.
func parseRow(row []string, idx columnIndex, lineNum int) (CsvRecord, error) {
	if len(row) <= idx.partNumber || len(row) <= idx.description || len(row) <= idx.quantity {
		return CsvRecord{}, fmt.Errorf("row has insufficient columns")
	}

	partNumber := strings.TrimSpace(row[idx.partNumber])
	description := strings.TrimSpace(row[idx.description])
	qtyStr := strings.TrimSpace(row[idx.quantity])

	if partNumber == "" {
		return CsvRecord{}, fmt.Errorf("empty PartNumber")
	}
	if description == "" {
		return CsvRecord{}, fmt.Errorf("empty Description")
	}

	qty, err := strconv.Atoi(qtyStr)
	if err != nil {
		return CsvRecord{}, fmt.Errorf("invalid Quantity %q: %w", qtyStr, err)
	}
	if qty < 1 {
		return CsvRecord{}, fmt.Errorf("Quantity must be >= 1, got %d", qty)
	}

	var configGroup string
	if idx.configGroup >= 0 && len(row) > idx.configGroup {
		configGroup = strings.TrimSpace(row[idx.configGroup])
	}

	// Extract optional cable fields
	sourceDevice := getColumnValue(row, idx.sourceDevice)
	sourcePort := getColumnValue(row, idx.sourcePort)
	destDevice := getColumnValue(row, idx.destDevice)
	destPort := getColumnValue(row, idx.destPort)
	cableType := getColumnValue(row, idx.cableType)
	cableLength := getColumnValue(row, idx.cableLength)

	return CsvRecord{
		PartNumber:   partNumber,
		Description:  description,
		Quantity:     qty,
		ConfigGroup:  configGroup,
		SourceDevice: sourceDevice,
		SourcePort:   sourcePort,
		DestDevice:   destDevice,
		DestPort:     destPort,
		CableType:    cableType,
		CableLength:  cableLength,
	}, nil
}

// getColumnValue safely extracts a value from a row at the given index.
func getColumnValue(row []string, idx int) string {
	if idx >= 0 && len(row) > idx {
		return strings.TrimSpace(row[idx])
	}
	return ""
}

// IsCableRecord returns true if the record defines a cable connection.
// Exported so the provider can separate device records from cable records during import.
func IsCableRecord(rec CsvRecord) bool {
	return rec.SourceDevice != "" && rec.DestDevice != ""
}

// isConfigParent returns true if configGroup ends in "00" (e.g., "0100", "0200").
func isConfigParent(configGroup string) bool {
	if len(configGroup) < 2 {
		return false
	}
	return strings.HasSuffix(configGroup, "00")
}

// getConfigParentGroup returns the parent group for a config group.
// E.g., "0201" -> "0200", "0315" -> "0300".
func getConfigParentGroup(configGroup string) string {
	if len(configGroup) < 2 {
		return ""
	}
	prefix := configGroup[:len(configGroup)-2]
	return prefix + "00"
}

// getConfigGroupPrefix returns the two-digit prefix of a config group.
// E.g., "0200" -> "02", "0315" -> "03".
func getConfigGroupPrefix(configGroup string) string {
	if len(configGroup) < 2 {
		return ""
	}
	return configGroup[:2]
}
