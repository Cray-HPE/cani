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

// CSM provider metadata keys, centralized to avoid duplicated string literals.
const (
	hmnVlanKey = "hmnVlan"
	aliasesKey = "aliases"
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

		changed, rowErr := applyCSVRow(headers, row, inventory, total)
		if rowErr != nil {
			return modified, total, rowErr
		}
		if changed {
			modified++
		}
	}

	return modified, total, nil
}

// applyCSVRow parses one CSV row and applies its fields to the matching
// device or rack, returning whether the inventory was modified.
func applyCSVRow(
	headers, row []string,
	inventory *devicetypes.Inventory,
	rowNum int,
) (bool, error) {
	rowMap := rowToMap(headers, row)
	idStr, ok := rowMap["ID"]
	if !ok || idStr == "" {
		return false, fmt.Errorf("missing ID for row %d", rowNum+1)
	}

	id, parseErr := uuid.Parse(idStr)
	if parseErr != nil {
		return false, fmt.Errorf("failed to parse %q as a UUID: %w", idStr, parseErr)
	}

	dev, ok := inventory.Devices[id]
	if !ok {
		return applyCSVRowToRack(id, rowMap, inventory)
	}

	changed := setDeviceFields(dev, rowMap)
	if changed {
		log.Printf("Updated %s", id)
	}
	return changed, nil
}

// applyCSVRowToRack applies CSV fields to a rack (a cabinet added via the
// "add rack" flow). It errors when no device or rack matches the ID.
func applyCSVRowToRack(
	id uuid.UUID,
	rowMap map[string]string,
	inventory *devicetypes.Inventory,
) (bool, error) {
	rack, rok := inventory.Racks[id]
	if !rok {
		return false, fmt.Errorf("could not find device with UUID %s", id)
	}
	changed := setRackFields(rack, rowMap)
	if changed {
		log.Printf("Updated %s", id)
	}
	return changed, nil
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
	oldAliases := sub[aliasesKey]

	if value == "" {
		if oldAliases == nil || aliasesIsEmpty(oldAliases) {
			return false
		}
		dev.SetProviderMeta("csm", aliasesKey, []string{})
		return true
	}

	if aliasesMatchSingle(oldAliases, value) {
		return false
	}

	dev.SetProviderMeta("csm", aliasesKey, []string{value})
	return true
}

// aliasesIsEmpty reports whether an aliases metadata value is a
// zero-length string or interface slice.
func aliasesIsEmpty(aliases any) bool {
	switch a := aliases.(type) {
	case []string:
		return len(a) == 0
	case []interface{}:
		return len(a) == 0
	}
	return false
}

// aliasesMatchSingle reports whether aliases already holds exactly the
// single given value.
func aliasesMatchSingle(aliases any, value string) bool {
	switch a := aliases.(type) {
	case []string:
		return len(a) == 1 && a[0] == value
	case []interface{}:
		return len(a) == 1 && fmt.Sprintf("%v", a[0]) == value
	}
	return false
}

// setCSMMetaVlan sets the hmnVlan field, parsing as integer.
func setCSMMetaVlan(dev *devicetypes.CaniDeviceType, value string) bool {
	sub, _ := dev.GetProviderSubMap("csm")

	if value == "" {
		old := sub[hmnVlanKey]
		if old == nil {
			return false
		}
		dev.SetProviderMeta("csm", hmnVlanKey, nil)
		return true
	}

	vlan, err := strconv.Atoi(value)
	if err != nil {
		return false
	}

	oldVal := sub[hmnVlanKey]
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

	dev.SetProviderMeta("csm", hmnVlanKey, vlan)
	return true
}

// setRackFields updates the CSM provider metadata on a rack
// from CSV values. Returns true if any field was modified.
func setRackFields(rack *devicetypes.CaniRackType, values map[string]string) bool {
	changed := false

	if v, ok := values["Vlan"]; ok {
		if setRackMetaVlan(rack, v) {
			changed = true
		}
	}

	return changed
}

// setRackMetaVlan sets the hmnVlan field on a rack, parsing as integer.
func setRackMetaVlan(rack *devicetypes.CaniRackType, value string) bool {
	sub, _ := rack.GetProviderSubMap("csm")

	if value == "" {
		old := sub[hmnVlanKey]
		if old == nil {
			return false
		}
		rack.SetProviderMeta("csm", hmnVlanKey, nil)
		return true
	}

	vlan, err := strconv.Atoi(value)
	if err != nil {
		return false
	}

	oldVal := sub[hmnVlanKey]
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

	rack.SetProviderMeta("csm", hmnVlanKey, vlan)
	return true
}
