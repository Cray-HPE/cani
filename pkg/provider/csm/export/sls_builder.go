package export

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/csm/import"
)

// caniMetadataSchemaVersion is the current CANI SLS schema version.
const caniMetadataSchemaVersion = "v1alpha1"

// buildExpectedHardware iterates over the CANI inventory devices and
// produces SLS hardware entries with CANI metadata injected into
// ExtraProperties. It merges existing SLS ExtraProperties so that
// provider-specific fields (Role, NID, Aliases, Networks, etc.) are
// preserved alongside the new CANI fields.
func buildExpectedHardware(
	inventory devicetypes.Inventory,
	current map[string]import_.SlsHardware,
) map[string]import_.SlsHardware {
	expected := make(map[string]import_.SlsHardware, len(current))

	// Copy all current hardware as the baseline.
	for xname, hw := range current {
		expected[xname] = hw
	}

	// Overlay CANI metadata onto hardware entries that match devices
	// in the CANI inventory.
	for _, dev := range inventory.Devices {
		xname := extractXname(dev)
		if xname == "" {
			continue
		}

		hw, exists := expected[xname]
		if !exists {
			// Device exists in CANI but not in SLS.  Build a new
			// SLS entry for staged hardware so the export can push them.
			switch {
			case strings.EqualFold(dev.Status, "staged") && dev.GetType() == devicetypes.TypeNode:
				hw = buildNewNodeEntry(dev, xname)
			case strings.EqualFold(dev.Status, "staged") && dev.GetType() == devicetypes.TypeCabinet:
				hw = buildNewCabinetEntry(dev, xname)
				// Also generate chassis and chassis-BMC children.
				for _, child := range buildCabinetChildren(dev, xname) {
					child.ExtraProperties = injectCaniMetadata(
						child.ExtraProperties, dev.ID.String(), "provisioned",
					)
					expected[child.Xname] = child
				}
			default:
				continue
			}
		}

		hw.ExtraProperties = injectCaniMetadata(
			hw.ExtraProperties, dev.ID.String(), caniStatus(dev),
		)
		expected[xname] = hw
	}

	// Second pass: inject CANI metadata into SLS-only entries whose
	// parent device is in the inventory.  This covers ChassisBMC and
	// similar entries that the transform skips but SLS still carries.
	for xname, hw := range expected {
		if hasCaniMetadata(hw) {
			continue // already handled above
		}
		parentDev := findParentDevice(xname, inventory)
		if parentDev == nil {
			continue
		}
		hw.ExtraProperties = injectCaniMetadata(
			hw.ExtraProperties, parentDev.ID.String(), caniStatus(parentDev),
		)
		expected[xname] = hw
	}

	return expected
}

// findParentDevice walks up the xname hierarchy looking for a device
// whose xname matches a prefix of the given xname.  For example, for
// ChassisBMC "x9000c1b0" it strips suffixes until it finds chassis
// device "x9000c1".
func findParentDevice(
	xname string,
	inventory devicetypes.Inventory,
) *devicetypes.CaniDeviceType {
	// Build a lookup map once per call — in practice this function
	// is called few times (only for SLS-only entries), so the cost
	// is acceptable.  For hot paths a cache could be added later.
	for _, dev := range inventory.Devices {
		devXname := extractXname(dev)
		if devXname != "" && strings.HasPrefix(xname, devXname) && xname != devXname {
			return dev
		}
	}
	return nil
}

// extractXname retrieves the xname stored in ProviderMetadata["csm"]["xname"].
func extractXname(dev *devicetypes.CaniDeviceType) string {
	if dev == nil {
		return ""
	}
	sub, ok := dev.GetProviderSubMap("csm")
	if !ok {
		return ""
	}
	xname, _ := sub["xname"].(string)
	return xname
}

// caniStatus derives the CANI status string from a device.
// Nodes are "empty" phantom placeholders unless explicitly staged by
// the user.  Staged nodes become "provisioned" when exported (the SLS
// state reflects the committed result, not the internal session state).
// Non-node hardware imported from SLS is always "provisioned".
func caniStatus(dev *devicetypes.CaniDeviceType) string {
	if dev == nil {
		return "empty"
	}
	if dev.GetType() == devicetypes.TypeNode {
		if strings.EqualFold(dev.Status, "staged") {
			return "provisioned"
		}
		return "empty"
	}
	return "provisioned"
}

// injectCaniMetadata returns a new ExtraProperties map that contains
// all entries from existing plus the four CANI metadata keys.
// A fresh map is always created so the caller never mutates a shared reference.
func injectCaniMetadata(
	existing map[string]any, caniID string, status string,
) map[string]any {
	result := make(map[string]any, len(existing)+4)
	for k, v := range existing {
		result[k] = v
	}
	result["@cani.id"] = caniID
	result["@cani.lastModified"] = time.Now().UTC().String()
	result["@cani.slsSchemaVersion"] = caniMetadataSchemaVersion
	result["@cani.status"] = status
	return result
}

// buildNewNodeEntry creates an SLS hardware entry for a staged node
// that does not yet exist in SLS.  The entry is populated from the
// CANI device's ProviderMetadata (xname, class, nid, role, aliases).
func buildNewNodeEntry(
	dev *devicetypes.CaniDeviceType,
	xname string,
) import_.SlsHardware {
	sub, _ := dev.GetProviderSubMap("csm")
	class, _ := sub["class"].(string)

	ep := make(map[string]any)

	// Copy node-specific metadata from the CANI device.
	if nid, ok := getIntMeta(sub, "nid"); ok && nid != 0 {
		ep["NID"] = nid
	}
	if role, _ := sub["role"].(string); role != "" {
		ep["Role"] = role
	} else if dev.Role != "" {
		ep["Role"] = dev.Role
	}
	if aliases := getStringSliceMeta(sub, "aliases"); len(aliases) > 0 {
		ep["Aliases"] = aliases
	}
	if subRole, _ := sub["subRole"].(string); subRole != "" {
		ep["SubRole"] = subRole
	}

	return import_.SlsHardware{
		Xname:           xname,
		Parent:          deriveParentXname(xname),
		Type:            "comptype_node",
		TypeString:      "Node",
		Class:           class,
		ExtraProperties: ep,
	}
}

// buildNewCabinetEntry creates an SLS hardware entry for a staged
// cabinet that does not yet exist in SLS.
func buildNewCabinetEntry(
	dev *devicetypes.CaniDeviceType,
	xname string,
) import_.SlsHardware {
	sub, _ := dev.GetProviderSubMap("csm")
	class, _ := sub["class"].(string)

	ep := make(map[string]any)

	return import_.SlsHardware{
		Xname:           xname,
		Parent:          "s0",
		Type:            "comptype_cabinet",
		TypeString:      "Cabinet",
		Class:           class,
		ExtraProperties: ep,
	}
}

// buildCabinetChildren generates chassis and chassis-BMC SLS entries
// for a new cabinet based on its device-type definition.
func buildCabinetChildren(
	dev *devicetypes.CaniDeviceType,
	cabinetXname string,
) []import_.SlsHardware {
	sub, _ := dev.GetProviderSubMap("csm")
	class, _ := sub["class"].(string)

	ordinals := chassisOrdinalsForSlug(dev.Slug)

	var children []import_.SlsHardware
	for _, ord := range ordinals {
		chassis := fmt.Sprintf("%sc%d", cabinetXname, ord)
		children = append(children, import_.SlsHardware{
			Xname:      chassis,
			Parent:     cabinetXname,
			Type:       "comptype_chassis",
			TypeString: "Chassis",
			Class:      class,
		})

		bmc := fmt.Sprintf("%sb0", chassis)
		children = append(children, import_.SlsHardware{
			Xname:      bmc,
			Parent:     chassis,
			Type:       "comptype_chassis_bmc",
			TypeString: "ChassisBMC",
			Class:      class,
		})
	}
	return children
}

// chassisOrdinalsForSlug returns the chassis bay ordinals defined in
// the rack-type YAML for the given slug.
func chassisOrdinalsForSlug(slug string) []int {
	rt, ok := devicetypes.GetRackTypeBySlug(slug)
	if !ok {
		return nil
	}
	var ordinals []int
	for _, bay := range rt.DeviceBays {
		if strings.Contains(strings.ToLower(bay.Name), "chassis") {
			ordinals = append(ordinals, bayOrdinal(bay))
		}
	}
	return ordinals
}

// bayOrdinal extracts the provider-specific ordinal from a
// DeviceBaySpec's Extra map.  Returns 0 when not present.
func bayOrdinal(bay devicetypes.DeviceBaySpec) int {
	switch n := bay.Extra["ordinal"].(type) {
	case int:
		return n
	case float64:
		return int(n)
	}
	return 0
}

// deriveParentXname removes the last xname component (e.g. "n2" from
// "x9000c1s0b0n2") to produce the parent xname ("x9000c1s0b0").
func deriveParentXname(xname string) string {
	// Walk backwards to find the last alphabetic prefix of the last
	// component (e.g. 'n' in "b0n2").
	for i := len(xname) - 1; i >= 0; i-- {
		c := xname[i]
		if c >= 'a' && c <= 'z' {
			return xname[:i]
		}
	}
	return xname
}

// getIntMeta extracts an integer from a metadata map, handling both
// int and float64 (from JSON round-trip).
func getIntMeta(m map[string]any, key string) (int, bool) {
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case int:
		return n, true
	case float64:
		return int(n), true
	}
	return 0, false
}

// getStringSliceMeta extracts a string slice from a metadata map,
// handling both []string and []any (from JSON round-trip).
func getStringSliceMeta(m map[string]any, key string) []string {
	v, ok := m[key]
	if !ok {
		return nil
	}
	switch s := v.(type) {
	case []string:
		return s
	case []any:
		out := make([]string, 0, len(s))
		for _, a := range s {
			if str, ok := a.(string); ok {
				out = append(out, str)
			}
		}
		return out
	}
	return nil
}

// marshalHardware serializes an SLS hardware entry to JSON for PUT.
func marshalHardware(hw import_.SlsHardware) ([]byte, error) {
	data, err := json.Marshal(hw)
	if err != nil {
		return nil, fmt.Errorf("marshaling SLS hardware %s: %w", hw.Xname, err)
	}
	return data, nil
}
