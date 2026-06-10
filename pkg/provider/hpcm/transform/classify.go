package transform

import (
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/hpcm/import"
)

// CaniCategory indicates which inventory collection a node maps to.
type CaniCategory string

const (
	CategoryLocation CaniCategory = "location"
	CategoryDevice   CaniCategory = "device"
	CategoryModule   CaniCategory = "module"
	CategorySkip     CaniCategory = "skip"
)

// aliasKeyProduct is the HPCM alias key holding the product/model name.
const aliasKeyProduct = "product"

// Classification holds the result of analysing an HPCM node. It tells the
// transform which CANI type to create and provides ranked lookup queries.
type Classification struct {
	Category       CaniCategory
	DeviceTypeHint devicetypes.Type
	LookupQueries  []string
	Criteria       map[string]string
}

// classifyNode determines the CANI category and type hint for a raw HPCM node.
//
// Rules (strongest signal first):
//  1. type field — admin→location, chassis/mgmt_switch/pdu→device
//  2. location hierarchy — for "compute", last non-nil field determines depth
//  3. product/part queries gathered from aliases, inventory, name
func classifyNode(node import_.Node) Classification {
	c := Classification{
		Criteria: make(map[string]string),
	}

	// Record raw type for diagnostics.
	c.Criteria["hpcm_type"] = node.Type

	// ── Rule 0: skip alias aggregator nodes ────────────────────────
	if isAliasNode(node) {
		c.Category = CategorySkip
		c.Criteria["rule"] = "alias-aggregator→skip"
		return c
	}

	// ── Rule 1: type field ──────────────────────────────────────────
	switch node.Type {
	case "admin":
		c.Category = CategoryLocation
		c.DeviceTypeHint = devicetypes.TypeNode // unused for locations
		c.Criteria["rule"] = "type=admin→location"
		c.LookupQueries = collectQueries(nil, node)
		return c
	case "chassis":
		c.Category = CategoryDevice
		c.DeviceTypeHint = devicetypes.TypeChassis
		c.Criteria["rule"] = "type=chassis→device"
		c.LookupQueries = collectQueries(nil, node)
		return c
	case "mgmt_switch":
		c.Category = CategoryDevice
		c.DeviceTypeHint = devicetypes.TypeMgmtSwitch
		c.Criteria["rule"] = "type=mgmt_switch→device"
		c.LookupQueries = collectQueries(nil, node)
		return c
	case "pdu":
		c.Category = CategoryDevice
		c.DeviceTypeHint = devicetypes.TypeCabinetPDU
		c.Criteria["rule"] = "type=pdu→device"
		c.LookupQueries = collectQueries(nil, node)
		return c
	}

	// ── Rule 2: location hierarchy (compute or unknown type) ────────
	c.Category, c.DeviceTypeHint = classifyByLocation(node.Location)
	c.Criteria["rule"] = "location→" + string(c.Category)
	c.LookupQueries = collectQueries(nil, node)
	return c
}

// classifyByLocation inspects the location fields in priority order
// (controller → node → tray → chassis → rack) and returns the
// category + type hint based on the deepest non-nil field.
func classifyByLocation(loc *import_.LocationSettings) (CaniCategory, devicetypes.Type) {
	if loc == nil {
		return CategoryDevice, devicetypes.TypeNode
	}

	// Walk deepest-first: the last non-nil field wins.
	if loc.Controller != nil {
		return CategoryModule, devicetypes.TypeModule
	}
	if loc.Node != nil {
		return CategoryModule, devicetypes.TypeModule
	}
	if loc.Tray != nil {
		return CategoryModule, devicetypes.TypeModule
	}
	if loc.Chassis != nil {
		return CategoryDevice, devicetypes.TypeNode
	}
	if loc.Rack != nil {
		return CategoryDevice, devicetypes.TypeNode
	}
	return CategoryDevice, devicetypes.TypeNode
}

// collectQueries gathers lookup strings from the node, ordered from most
// specific to least specific. frus may be nil when called before FRU
// construction (classification phase).
func collectQueries(frus []devicetypes.CaniFruType, node import_.Node) []string {
	seen := make(map[string]bool)
	var queries []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] || isSentinel(s) {
			return
		}
		seen[s] = true
		queries = append(queries, s)
	}

	// 1. FRU part numbers (most specific).
	for _, f := range frus {
		add(f.PartNumber)
	}

	// 2. Inventory SKU / part number fields.
	if node.Inventory != nil {
		add(node.Inventory["fru.system.SKU"])
		add(node.Inventory["sys.SKU Number"])
		add(node.Inventory["fru.PartNumber"])
		add(node.Inventory["fru.SKU"])
	}

	// 3. Product alias (highest-priority alias).
	if node.Aliases != nil {
		add(node.Aliases[aliasKeyProduct])
	}

	// 3b. Template ctrl_model alias (cm.config hardware model identifier).
	if node.Aliases != nil {
		add(node.Aliases["ctrl_model"])
	}

	// 4. Node name.
	add(node.Name)

	// 4b. Template name alias (cm.config secondary lookup signal).
	if node.Aliases != nil {
		add(node.Aliases["template_name"])
	}

	// 5. Inventory-derived model/product fields.
	if node.Inventory != nil {
		add(node.Inventory["fru.system.Model"])
		add(node.Inventory["sys.Product Name"])
		add(node.Inventory["board.Product Name"])
		add(node.Inventory["fru.Model"])
	}

	// 6. All remaining alias values (lowest priority, "product" already added).
	if node.Aliases != nil {
		for key, val := range node.Aliases {
			if key == aliasKeyProduct {
				continue
			}
			add(val)
		}
	}

	return queries
}

// UnmatchedNode records a node whose lookup queries returned no library hit.
type UnmatchedNode struct {
	Name     string
	Category CaniCategory
	Criteria map[string]string
	Queries  []string
}

// sentinels contains placeholder values commonly found in HPCM inventory
// fields that should never be used as lookup queries.
var sentinels = map[string]bool{
	"na":             true,
	"n/a":            true,
	"not specified":  true,
	"not available":  true,
	"default string": true,
	"unknown":        true,
	"unspecified":    true,
	"none":           true,
	"default":        true,
	"not provided":   true,
}

// isSentinel returns true if the value is a known placeholder or is too
// short / too generic to produce a reliable library match.
func isSentinel(s string) bool {
	if len(s) < 3 {
		return true
	}
	if sentinels[strings.ToLower(s)] {
		return true
	}
	// All-zero or all-same-digit strings (e.g. "000000000001", "01234567").
	if isNumericJunk(s) {
		return true
	}
	return false
}

// isAliasNode returns true for HPCM alias aggregator nodes that are
// virtual constructs (e.g. su-aliases, su-bmc-aliases) rather than
// physical hardware.
func isAliasNode(node import_.Node) bool {
	name := strings.ToLower(node.Name)
	if !strings.HasSuffix(name, "-aliases") {
		return false
	}
	// Alias aggregators have no inventory and no product alias.
	if len(node.Inventory) > 0 {
		return false
	}
	if node.Aliases != nil {
		if _, ok := node.Aliases[aliasKeyProduct]; ok {
			return false
		}
	}
	return true
}

// isNumericJunk returns true for strings that are purely hex digits and
// unlikely to be real part numbers — specifically, strings whose distinct
// digit set is trivially small (sequential or all-same).
func isNumericJunk(s string) bool {
	digits := 0
	for _, c := range s {
		switch {
		case c >= '0' && c <= '9', c >= 'a' && c <= 'f', c >= 'A' && c <= 'F':
			digits++
		default:
			return false // contains non-hex chars → not numeric junk
		}
	}
	if digits == 0 {
		return false
	}
	// Pure hex strings shorter than 6 chars are suspicious but could be legit
	// part numbers; reject only all-zero strings in that range.
	allZero := true
	for _, c := range s {
		if c != '0' {
			allZero = false
			break
		}
	}
	return allZero
}
