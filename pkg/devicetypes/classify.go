package devicetypes

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// ErrNilDevice is returned when a nil device pointer is passed.
var ErrNilDevice = errors.New("device is nil")

// ErrSlugNotFound returns an error indicating the given slug is not in the library.
func ErrSlugNotFound(slug string) error {
	return fmt.Errorf("device type slug %q not found in library", slug)
}

// UnclassifiedDevice holds the summary fields of a device that lacks a resolved CaniType.
type UnclassifiedDevice struct {
	ID               uuid.UUID
	Name             string
	HardwareType     string
	Model            string
	Manufacturer     string
	DeviceType       string // e.g. node, blade, chassis, nodecard
	Status           string
	Role             string
	ChildrenCount    int
	ProviderMetadata map[string]any // e.g. csm xname, class, aliases
}

// SuggestTypes queries the device-type library with every non-empty field of the
// device and returns the top unique candidates ordered by descending score.
// At most maxResults entries are returned. When text-based matching yields
// fewer than maxResults candidates, a hardware-type fallback fills in
// additional suggestions from device types whose HardwareType matches.
func SuggestTypes(device UnclassifiedDevice, maxResults int) []MatchResult {
	seen := make(map[string]int) // slug -> best score

	queries := collectQueries(device)
	for _, q := range queries {
		// Exact / high-confidence single-result lookup.
		dt, score := LookupScored(q)
		if score > 0 && dt.Slug != "" {
			if prev, ok := seen[dt.Slug]; !ok || score > prev {
				seen[dt.Slug] = score
			}
		}
		// Also gather all fuzzy matches (multiple candidates per query).
		for _, mr := range FuzzyMatchAll(q, maxResults) {
			if prev, ok := seen[mr.Slug]; !ok || mr.Score > prev {
				seen[mr.Slug] = mr.Score
			}
		}
	}

	// Hardware-type fallback: if we have fewer than maxResults, append
	// device types with a matching HardwareType at low confidence.
	if len(seen) < maxResults && device.HardwareType != "" {
		hwFallback := hardwareTypeFallback(device.HardwareType, maxResults-len(seen))
		for _, slug := range hwFallback {
			if _, ok := seen[slug]; !ok {
				seen[slug] = 15 // low-confidence fallback
			}
		}
	}

	results := make([]MatchResult, 0, len(seen))
	for slug, score := range seen {
		results = append(results, MatchResult{Slug: slug, Score: score})
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Slug < results[j].Slug
	})

	if maxResults > 0 && len(results) > maxResults {
		results = results[:maxResults]
	}
	return results
}

// collectQueries builds a deduplicated list of non-empty query strings from
// the device's identity fields. It also decomposes compound names (e.g.
// "dl360gen11") into sub-tokens ("dl360", "gen11") so each sub-token can
// independently match against the device-type library.
func collectQueries(device UnclassifiedDevice) []string {
	candidates := []string{
		device.Name,
		device.Model,
		device.HardwareType,
		device.Manufacturer,
		device.DeviceType,
		device.Role,
	}
	seen := make(map[string]bool)
	out := make([]string, 0, len(candidates)*2)
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			return
		}
		seen[s] = true
		out = append(out, s)
	}

	for _, c := range candidates {
		add(c)
		// Decompose compound names into sub-tokens.
		subs := tokenizeCamelNum(c)
		if len(subs) >= 2 {
			for _, sub := range subs {
				if len(sub) >= minFuzzyLen {
					add(sub)
				}
			}
		}
	}
	return out
}

// normalizeHardwareType converts HPCM-style hardware types (underscore
// separated) into the canonical form used in the device-type registry
// (hyphen separated). For example "mgmt_switch" → "mgmt-switch".
func normalizeHardwareType(ht string) string {
	return strings.ReplaceAll(ht, "_", "-")
}

// hardwareTypeFallback returns up to max slugs whose HardwareType matches
// the given hardware type. It tries the normalized form first and falls
// back to common aliases.
func hardwareTypeFallback(hwType string, max int) []string {
	if max <= 0 {
		return nil
	}
	norm := normalizeHardwareType(hwType)
	types := []Type{Type(norm)}
	// Add related type aliases so "compute" also matches "blade" and "node".
	for _, alias := range relatedHardwareTypes(norm) {
		types = append(types, alias)
	}

	matches := ListCaniDeviceTypes(types...)
	slugs := make([]string, 0, len(matches))
	for slug := range matches {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs) // deterministic order
	if len(slugs) > max {
		slugs = slugs[:max]
	}
	return slugs
}

// relatedHardwareTypes returns additional Type values that are conceptually
// related to the given normalized hardware type. This broadens the fallback
// search so that e.g. a "compute" node also considers "blade" and "node"
// device types.
func relatedHardwareTypes(norm string) []Type {
	switch norm {
	case "compute":
		return []Type{TypeBlade, TypeNode}
	case "blade":
		return []Type{TypeNode}
	case "node":
		return []Type{TypeBlade}
	case "mgmt-switch":
		return []Type{TypeSwitch}
	case "switch":
		return []Type{TypeMgmtSwitch}
	default:
		return nil
	}
}

// GetAllSlugs returns a sorted list of all registered device-type slugs.
func GetAllSlugs() []string {
	slugs := make([]string, 0, len(allDeviceTypes))
	for slug := range allDeviceTypes {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)
	return slugs
}

// FindUnclassifiedDevices scans the inventory for devices that have no Slug and
// no Model (i.e. they cannot be matched to a CaniType). Returns a list of
// summaries suitable for interactive classification.
func FindUnclassifiedDevices(inv *Inventory) []UnclassifiedDevice {
	var result []UnclassifiedDevice
	for _, device := range inv.Devices {
		if device == nil {
			continue
		}
		if device.Slug != "" || device.Model != "" {
			continue
		}
		result = append(result, UnclassifiedDevice{
			ID:               device.ID,
			Name:             device.Name,
			HardwareType:     device.HardwareType,
			Model:            device.Model,
			Manufacturer:     device.Manufacturer,
			DeviceType:       string(device.Type),
			Status:           device.Status,
			Role:             device.Role,
			ChildrenCount:    len(device.Children),
			ProviderMetadata: device.ProviderMetadata,
		})
	}
	// Sort by name for deterministic output.
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// ApplyDeviceType looks up the slug in the library and copies classification
// fields from the library template onto the device. Only fields that are empty
// on the device are overwritten. Returns an error if the slug is not found.
func ApplyDeviceType(device *CaniDeviceType, slug string) error {
	if device == nil {
		return ErrNilDevice
	}
	tmpl, ok := GetBySlug(slug)
	if !ok {
		return ErrSlugNotFound(slug)
	}

	device.Slug = tmpl.Slug
	if device.Model == "" {
		device.Model = tmpl.Model
	}
	if device.Manufacturer == "" {
		device.Manufacturer = tmpl.Manufacturer
	}
	if device.PartNumber == "" {
		device.PartNumber = tmpl.PartNumber
	}
	if device.Type == "" {
		device.Type = tmpl.Type
	}
	if device.HardwareType == "" {
		device.HardwareType = tmpl.HardwareType
	}
	if device.SubdeviceRole == "" {
		device.SubdeviceRole = tmpl.SubdeviceRole
	}
	if device.UHeight == 0 {
		device.UHeight = tmpl.UHeight
	}
	if device.Description == "" {
		device.Description = tmpl.Description
	}
	if len(device.Interfaces) == 0 {
		device.Interfaces = tmpl.Interfaces
	}
	if len(device.ModuleBays) == 0 {
		device.ModuleBays = tmpl.ModuleBays
	}
	if len(device.DeviceBays) == 0 {
		device.DeviceBays = tmpl.DeviceBays
	}
	if len(device.ConsolePorts) == 0 {
		device.ConsolePorts = tmpl.ConsolePorts
	}
	if len(device.PowerPorts) == 0 {
		device.PowerPorts = tmpl.PowerPorts
	}
	if len(device.Identifications) == 0 {
		device.Identifications = tmpl.Identifications
	}
	if len(device.AllowedChildren) == 0 {
		device.AllowedChildren = tmpl.AllowedChildren
	}
	return nil
}
