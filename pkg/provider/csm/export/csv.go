package export

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// csvAllowedHeaders maps lowercase header names to their canonical form.
var csvAllowedHeaders = map[string]string{
	"id":      "ID",
	"name":    "Name",
	"type":    "Type",
	"status":  "Status",
	"vlan":    "Vlan",
	"role":    "Role",
	"subrole": "SubRole",
	"alias":   "Alias",
	"nid":     "Nid",
}

// ExportCSV writes the inventory as CSV to w, filtered by headers and types.
func ExportCSV(w io.Writer, inv devicetypes.Inventory, headers []string, types []string) error {
	normalizedHeaders, err := normalizeHeaders(headers)
	if err != nil {
		return err
	}

	typeSet := buildTypeSet(types)

	// Collect and sort device UUIDs for deterministic output.
	keys := sortedDeviceKeys(inv)

	cw := csv.NewWriter(w)
	cw.Write(normalizedHeaders)

	for _, id := range keys {
		dev := inv.Devices[id]
		if !matchesType(dev, typeSet) {
			continue
		}
		row := getFields(dev, normalizedHeaders)
		cw.Write(row)
	}
	cw.Flush()
	return cw.Error()
}

// normalizeHeaders converts user-supplied header names to their canonical form.
func normalizeHeaders(headers []string) ([]string, error) {
	out := make([]string, len(headers))
	var errs []string
	for i, h := range headers {
		canon, ok := csvAllowedHeaders[strings.ToLower(strings.TrimSpace(h))]
		if !ok {
			errs = append(errs, fmt.Sprintf("invalid header: %s", h))
			continue
		}
		out[i] = canon
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return out, nil
}

// typeAliases maps user-friendly type names to their internal Type constant.
// This enables backwards-compatible filtering with names like "nodeblade".
var typeAliases = map[string]string{
	"nodeblade":  string(devicetypes.TypeNodeCard),
	"nodecard":   string(devicetypes.TypeNodeCard),
	"node":       string(devicetypes.TypeNode),
	"cabinet":    string(devicetypes.TypeCabinet),
	"chassis":    string(devicetypes.TypeChassis),
	"blade":      string(devicetypes.TypeBlade),
	"switch":     string(devicetypes.TypeSwitch),
	"mgmtswitch": string(devicetypes.TypeMgmtSwitch),
	"hsnswitch":  string(devicetypes.TypeHsnSwitch),
	"cabinetpdu": string(devicetypes.TypeCabinetPDU),
	"cdu":        string(devicetypes.TypeCDU),
}

// buildTypeSet creates a lowercase set of type names for filtering.
// An empty set means "all types".
func buildTypeSet(types []string) map[string]struct{} {
	set := make(map[string]struct{}, len(types))
	for _, t := range types {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		lower := strings.ToLower(t)
		if resolved, ok := typeAliases[lower]; ok {
			set[resolved] = struct{}{}
		} else {
			set[lower] = struct{}{}
		}
	}
	return set
}

// matchesType returns true if the device matches the type filter.
func matchesType(dev *devicetypes.CaniDeviceType, typeSet map[string]struct{}) bool {
	if len(typeSet) == 0 {
		return true
	}
	_, ok := typeSet[string(dev.Type)]
	return ok
}

// sortedDeviceKeys returns device UUIDs sorted by Type then ID for
// deterministic CSV output.
func sortedDeviceKeys(inv devicetypes.Inventory) []uuid.UUID {
	keys := make([]uuid.UUID, 0, len(inv.Devices))
	for k := range inv.Devices {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		di := inv.Devices[keys[i]]
		dj := inv.Devices[keys[j]]
		if di.Type != dj.Type {
			return di.Type < dj.Type
		}
		return keys[i].String() < keys[j].String()
	})
	return keys
}

// getFields extracts field values from a device for the given headers.
func getFields(dev *devicetypes.CaniDeviceType, headers []string) []string {
	values := make([]string, len(headers))
	for i, h := range headers {
		values[i] = getField(dev, h)
	}
	return values
}

// getField returns a single field value from a device.
func getField(dev *devicetypes.CaniDeviceType, header string) string {
	switch header {
	case "ID":
		return dev.ID.String()
	case "Name":
		return dev.Name
	case "Type":
		return canonicalTypeName(dev.Type)
	case "Status":
		return dev.Status
	case "Vlan":
		return getCSMMetaString(dev, "hmnVlan")
	case "Role":
		return getCSMMetaString(dev, "role")
	case "SubRole":
		return getCSMMetaString(dev, "subRole")
	case "Nid":
		return getCSMMetaString(dev, "nid")
	case "Alias":
		return getCSMMetaFirstAlias(dev)
	default:
		return ""
	}
}

// canonicalTypeName returns a display-friendly type name
// with the first letter uppercase and the rest as-is.
func canonicalTypeName(t devicetypes.Type) string {
	s := string(t)
	if s == "" {
		return ""
	}
	// Map known types to their display names.
	switch t {
	case devicetypes.TypeCabinet:
		return "Cabinet"
	case devicetypes.TypeChassis:
		return "Chassis"
	case devicetypes.TypeBlade:
		return "Blade"
	case devicetypes.TypeNode:
		return "Node"
	case devicetypes.TypeNodeCard:
		return "NodeBlade"
	case devicetypes.TypeSwitch:
		return "Switch"
	case devicetypes.TypeMgmtSwitch:
		return "MgmtSwitch"
	case devicetypes.TypeHsnSwitch:
		return "HsnSwitch"
	case devicetypes.TypeCabinetPDU:
		return "CabinetPDU"
	case devicetypes.TypeCDU:
		return "CDU"
	default:
		// Capitalize first letter for unknown types.
		if len(s) == 0 {
			return s
		}
		return strings.ToUpper(s[:1]) + s[1:]
	}
}

// getCSMMetaString returns a string value from the csm provider metadata.
func getCSMMetaString(dev *devicetypes.CaniDeviceType, key string) string {
	sub, ok := dev.GetProviderSubMap("csm")
	if !ok {
		return ""
	}
	v, ok := sub[key]
	if !ok || v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// getCSMMetaFirstAlias returns the first alias from csm metadata.
func getCSMMetaFirstAlias(dev *devicetypes.CaniDeviceType) string {
	sub, ok := dev.GetProviderSubMap("csm")
	if !ok {
		return ""
	}
	v, ok := sub["aliases"]
	if !ok {
		return ""
	}
	switch a := v.(type) {
	case []string:
		if len(a) > 0 {
			return a[0]
		}
	case []interface{}:
		if len(a) > 0 {
			return fmt.Sprintf("%v", a[0])
		}
	case string:
		return a
	}
	return ""
}
