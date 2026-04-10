package transform

import (
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/redfish/import"
)

// enrichDeviceFromLibrary attempts to match the device against the device
// type library using queries derived from the ServiceRoot. Tries all queries
// and keeps the highest-scored match. Returns the matched slug, winning
// query, and score.
func enrichDeviceFromLibrary(dev *devicetypes.CaniDeviceType, root import_.ServiceRoot) (slug, matchQuery string, matchScore int) {
	queries := buildLookupQueries(root)

	var bestDT devicetypes.CaniDeviceType
	var bestQuery string
	bestScore := 0

	for _, q := range queries {
		dt, score := devicetypes.LookupScored(q)
		if score > bestScore {
			bestScore = score
			bestDT = dt
			bestQuery = q
		}
		if bestScore >= 100 {
			break // exact match
		}
	}

	if bestDT.Slug != "" {
		applyDeviceDefaults(dev, bestDT)
		return bestDT.Slug, bestQuery, bestScore
	}
	return "", "", 0
}

// buildLookupQueries creates an ordered list of lookup strings from a
// ServiceRoot, from most specific to least specific.
func buildLookupQueries(root import_.ServiceRoot) []string {
	seen := make(map[string]bool)
	var queries []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			return
		}
		seen[s] = true
		queries = append(queries, s)
	}

	// 1. Product field — most specific (e.g. "ProLiant DL325 Gen11").
	add(root.Product)

	// 2. HPE product tag (e.g. "HPE iLO 6") — may hit library entries.
	add(root.ProductTag())

	// 3. Vendor + Product combined (e.g. "HPE ProLiant DL325 Gen11").
	if root.Vendor != "" && root.Product != "" {
		add(root.Vendor + " " + root.Product)
	}

	// 4. System family (e.g. "ProLiant") — very broad, low priority.
	add(root.SystemFamily())

	// 5. OEM product name if different from Product.
	if root.Oem.Hpe != nil {
		add(root.Oem.Hpe.Moniker.PRODNAM)
		add(root.Oem.Hpe.Moniker.PRODGEN)
	}

	return queries
}

// applyDeviceDefaults copies non-empty library fields into the device.
func applyDeviceDefaults(dev *devicetypes.CaniDeviceType, lib devicetypes.CaniDeviceType) {
	if dev.Slug == "" {
		dev.Slug = lib.Slug
	}
	if dev.PartNumber == "" {
		dev.PartNumber = lib.PartNumber
	}
	if dev.Manufacturer == "" {
		dev.Manufacturer = lib.Manufacturer
	}
	if dev.Model == "" {
		dev.Model = lib.Model
	}
	if dev.Description == "" {
		dev.Description = lib.Description
	}
	if dev.UHeight == 0 {
		dev.UHeight = lib.UHeight
	}
	if dev.Type == "" || dev.Type == "server" {
		if lib.Type != "" {
			dev.Type = lib.Type
		}
	}
}
