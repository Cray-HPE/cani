package transform

import (
	"fmt"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
)

// transformMetadata builds the role and status catalog from the system CSV's
// `role` and `status` sections. It returns nil when neither section is present.
func transformMetadata(data *import_.SystemCSV) (*devicetypes.InventoryMetadata, error) {
	roles, err := metadataEntriesFromRecords(data, data.Roles, "role")
	if err != nil {
		return nil, err
	}
	statuses, err := metadataEntriesFromRecords(data, data.Statuses, "status")
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 && len(statuses) == 0 {
		return nil, nil
	}
	return &devicetypes.InventoryMetadata{Roles: roles, Statuses: statuses}, nil
}

// metadataEntriesFromRecords converts catalog records (roles or statuses) into
// MetadataEntry items, applying defaults and parsing content types. The kind is
// used only for the missing-name error message.
func metadataEntriesFromRecords(data *import_.SystemCSV, records []import_.SystemRecord, kind string) ([]devicetypes.MetadataEntry, error) {
	var entries []devicetypes.MetadataEntry
	for _, rec := range records {
		rec = data.ApplyDefaults(rec)
		if rec.Name == "" {
			return nil, fmt.Errorf("%s record missing Name", kind)
		}
		entries = append(entries, devicetypes.MetadataEntry{
			Name:         rec.Name,
			Color:        rec.Color,
			Description:  rec.Description,
			ContentTypes: parseContentTypes(rec.ContentTypes),
		})
	}
	return entries, nil
}

// parseContentTypes splits a comma-separated content-type string and normalizes
// each entry to Nautobot's "<app>.<model>" form, dropping empties.
func parseContentTypes(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	for _, ct := range strings.Split(s, ",") {
		if norm := normalizeContentType(ct); norm != "" {
			out = append(out, norm)
		}
	}
	return out
}

// ipamBareContentTypes are unqualified content-type model names that belong to
// the Nautobot ipam app rather than dcim.
var ipamBareContentTypes = map[string]bool{
	"prefix":    true,
	"ipaddress": true,
	"vlan":      true,
	"vlangroup": true,
	"namespace": true,
	"vrf":       true,
}

// normalizeContentType converts a system CSV content type into Nautobot's
// "<app>.<model>" form. Values already containing a dot pass through unchanged;
// a bare model name is prefixed with its app label, defaulting to dcim and using
// ipam for IP-address-management models.
func normalizeContentType(ct string) string {
	ct = strings.TrimSpace(ct)
	if ct == "" || strings.Contains(ct, ".") {
		return ct
	}
	lower := strings.ToLower(ct)
	if ipamBareContentTypes[lower] {
		return "ipam." + lower
	}
	return "dcim." + lower
}
