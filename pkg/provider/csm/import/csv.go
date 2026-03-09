package import_

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// ImportCSV reads a CSV from stdin and updates devices in the inventory.
// It returns the number of modified and total records.
func ImportCSV(inventory *devicetypes.Inventory) (modified, total int, err error) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return 0, 0, fmt.Errorf("no CSV input: pipe data to stdin or provide a file")
	}

	reader := csv.NewReader(os.Stdin)
	return importCSVFromReader(reader, inventory)
}

// IsStdinPiped returns true when stdin has piped data.
func IsStdinPiped() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// importCSVFromReader reads CSV data and updates matching devices.
func importCSVFromReader(reader *csv.Reader, inventory *devicetypes.Inventory) (modified, total int, err error) {
	headers, err := reader.Read()
	if err == io.EOF {
		return 0, 0, fmt.Errorf("the CSV file is empty")
	}
	if err != nil {
		return 0, 0, fmt.Errorf("reading CSV headers: %w", err)
	}

	if !hasIDColumn(headers) {
		return 0, 0, fmt.Errorf("ID column is missing")
	}

	for {
		row, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return modified, total, fmt.Errorf("reading CSV row: %w", readErr)
		}

		total++

		rowMap := rowToMap(headers, row)
		idStr, ok := rowMap["ID"]
		if !ok || idStr == "" {
			return modified, total, fmt.Errorf("missing ID for row %d", total+1)
		}

		id, parseErr := uuid.Parse(idStr)
		if parseErr != nil {
			return modified, total, fmt.Errorf("failed to parse %q as a UUID: %w", idStr, parseErr)
		}

		dev, ok := inventory.Devices[id]
		if !ok {
			return modified, total, fmt.Errorf("could not find device with UUID %s", id)
		}

		changed := setDeviceFields(dev, rowMap)
		if changed {
			log.Printf("Updated %s", id)
			modified++
		}
	}

	return modified, total, nil
}

// hasIDColumn checks whether "ID" is present in the headers.
func hasIDColumn(headers []string) bool {
	for _, h := range headers {
		if strings.EqualFold(strings.TrimSpace(h), "id") {
			return true
		}
	}
	return false
}

// rowToMap converts a CSV row to a map using normalized header names.
func rowToMap(headers, row []string) map[string]string {
	m := make(map[string]string, len(headers))
	for i, h := range headers {
		key := normalizeHeader(h)
		if i < len(row) {
			m[key] = row[i]
		}
	}
	return m
}

// normalizeHeader maps a header to its canonical form.
func normalizeHeader(h string) string {
	switch strings.ToLower(strings.TrimSpace(h)) {
	case "id":
		return "ID"
	case "type":
		return "Type"
	case "name":
		return "Name"
	case "status":
		return "Status"
	case "vlan":
		return "Vlan"
	case "role":
		return "Role"
	case "subrole":
		return "SubRole"
	case "alias":
		return "Alias"
	case "nid":
		return "Nid"
	default:
		return h
	}
}

// setDeviceFields updates the CSM provider metadata on a device
// from CSV values. Returns true if any field was modified.
func setDeviceFields(dev *devicetypes.CaniDeviceType, values map[string]string) bool {
	changed := false

	// Node-specific fields: Role, SubRole, Nid, Alias
	if v, ok := values["Role"]; ok {
		if setCSMMetaString(dev, "role", v) {
			changed = true
		}
	}
	if v, ok := values["SubRole"]; ok {
		if setCSMMetaString(dev, "subRole", v) {
			changed = true
		}
	}
	if v, ok := values["Nid"]; ok {
		if setCSMMetaNid(dev, v) {
			changed = true
		}
	}
	if v, ok := values["Alias"]; ok {
		if setCSMMetaAlias(dev, v) {
			changed = true
		}
	}

	// Cabinet-specific: Vlan (HMN VLAN)
	if v, ok := values["Vlan"]; ok {
		if setCSMMetaVlan(dev, v) {
			changed = true
		}
	}

	return changed
}

// setCSMMetaString sets a string field in csm metadata,
// returning true if the value changed.
func setCSMMetaString(dev *devicetypes.CaniDeviceType, key, value string) bool {
	sub, _ := dev.GetProviderSubMap("csm")
	old, _ := sub[key].(string)
	if old == value {
		return false
	}
	dev.SetProviderMeta("csm", key, value)
	return true
}

// setCSMMetaNid sets the nid field, parsing the string as an integer.
func setCSMMetaNid(dev *devicetypes.CaniDeviceType, value string) bool {
	sub, _ := dev.GetProviderSubMap("csm")

	if value == "" {
		old := sub["nid"]
		if old == nil {
			return false
		}
		dev.SetProviderMeta("csm", "nid", nil)
		return true
	}

	nid, err := strconv.Atoi(value)
	if err != nil {
		return false
	}

	// Compare with current value
	oldVal := sub["nid"]
	switch v := oldVal.(type) {
	case float64:
		if int(v) == nid {
			return false
		}
	case int:
		if v == nid {
			return false
		}
	}

	dev.SetProviderMeta("csm", "nid", nid)
	return true
}

// setCSMMetaAlias sets the aliases field from a single alias string.
func setCSMMetaAlias(dev *devicetypes.CaniDeviceType, value string) bool {
	sub, _ := dev.GetProviderSubMap("csm")
	oldAliases := sub["aliases"]

	if value == "" {
		if oldAliases == nil {
			return false
		}
		// Check if old is empty slice
		switch a := oldAliases.(type) {
		case []string:
			if len(a) == 0 {
				return false
			}
		case []interface{}:
			if len(a) == 0 {
				return false
			}
		}
		dev.SetProviderMeta("csm", "aliases", []string{})
		return true
	}

	newAliases := []string{value}

	// Compare with current first alias
	switch a := oldAliases.(type) {
	case []string:
		if len(a) == 1 && a[0] == value {
			return false
		}
	case []interface{}:
		if len(a) == 1 && fmt.Sprintf("%v", a[0]) == value {
			return false
		}
	}

	dev.SetProviderMeta("csm", "aliases", newAliases)
	return true
}

// setCSMMetaVlan sets the hmnVlan field, parsing as integer.
func setCSMMetaVlan(dev *devicetypes.CaniDeviceType, value string) bool {
	sub, _ := dev.GetProviderSubMap("csm")

	if value == "" {
		old := sub["hmnVlan"]
		if old == nil {
			return false
		}
		dev.SetProviderMeta("csm", "hmnVlan", nil)
		return true
	}

	vlan, err := strconv.Atoi(value)
	if err != nil {
		return false
	}

	oldVal := sub["hmnVlan"]
	switch v := oldVal.(type) {
	case float64:
		if int(v) == vlan {
			return false
		}
	case int:
		if v == vlan {
			return false
		}
	}

	dev.SetProviderMeta("csm", "hmnVlan", vlan)
	return true
}
