package transform

import (
	"fmt"
	"log"
	"strings"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/hpcm/commands"
	import_ "github.com/Cray-HPE/cani/pkg/provider/hpcm/import"
	"github.com/Cray-HPE/cani/pkg/visual"
	"github.com/google/uuid"
)

// providerGetter returns the Hpcm singleton with raw nodes.
// Set by the parent package to break import cycles.
var providerGetter func() interface {
	GetNodes() []import_.Node
}

// SetProviderGetter allows the parent package to provide singleton access.
func SetProviderGetter(getter func() interface {
	GetNodes() []import_.Node
}) {
	providerGetter = getter
}

// Transform converts raw HPCM nodes (stored by the import step) into CANI
// inventory types. Returns a TransformResult containing devices, racks, and
// FRUs.
func Transform(existing devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	p := providerGetter()
	nodes := p.GetNodes()
	if len(nodes) == 0 {
		log.Println("No raw nodes to transform")
		return &devicetypes.TransformResult{}, nil
	}
	return transformNodes(nodes, &existing)
}

// transformNodes converts raw nodes into devices, modules, locations, racks, and FRUs.
// Two-pass approach: chassis devices are created first so module parents exist.
func transformNodes(nodes []import_.Node, existing *devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	result := &devicetypes.TransformResult{
		Locations: make(map[uuid.UUID]*devicetypes.CaniLocationType),
		Devices:   make(map[uuid.UUID]*devicetypes.CaniDeviceType),
		Modules:   make(map[uuid.UUID]*devicetypes.CaniModuleType),
		Racks:     make(map[uuid.UUID]*devicetypes.CaniRackType),
		Frus:      make(map[uuid.UUID]*devicetypes.CaniFruType),
	}

	stepMode := config.Cfg != nil && config.Cfg.StepMode
	noColor := config.Cfg != nil && config.Cfg.NoColor
	opts := visual.ETLOptions{NoColor: noColor}
	tally := visual.StepTally{}
	racksByNumber := make(map[int32]uuid.UUID)

	// Classify every node up front.
	classifications := make([]Classification, len(nodes))
	for i, node := range nodes {
		classifications[i] = classifyNode(node)
	}

	// chassisByLoc maps "rack-chassis" → device UUID for module parenting.
	chassisByLoc := make(map[string]uuid.UUID)
	// chassisByXname maps chassis xname (e.g. "x9000c1") → device UUID
	// for geoloc-based module parenting.
	chassisByXname := make(map[string]uuid.UUID)

	// ── Pass 1: chassis devices (so their UUIDs exist for modules). ──
	for i, node := range nodes {
		cl := classifications[i]
		if cl.Category != CategoryDevice || cl.DeviceTypeHint != devicetypes.TypeChassis {
			continue
		}
		dev, frus := buildDeviceFromNode(node, cl, existing)
		_, mq, ms := enrichDeviceFromLibrary(&dev, frus, cl)
		result.Devices[dev.ID] = &dev
		addFrus(result, frus)

		rack := buildRack(node, racksByNumber, existing)
		if rack != nil {
			result.Racks[rack.ID] = rack
			tally.Racks++
		}
		assignRack(&dev, node, racksByNumber, result)

		// Register chassis for module parenting.
		if node.Location != nil && node.Location.Rack != nil && node.Location.Chassis != nil {
			key := chassisKey(*node.Location.Rack, *node.Location.Chassis)
			chassisByLoc[key] = dev.ID
		}
		// Register by xname for geoloc-based resolution.
		chassisByXname[node.Name] = dev.ID

		tally.Devices++
		tally.Cables += len(frus)
		if stepMode {
			info := buildNodeStepInfo(i+1, len(nodes), node, &dev, frus, dev.Slug, mq, ms)
			if err := visual.PromptNodeTransformStep(info, tally, opts); err != nil {
				return nil, fmt.Errorf("step interrupted: %w", err)
			}
		}
	}

	// ── Pass 2: everything else. ─────────────────────────────────────
	var unmatched []UnmatchedNode

	for i, node := range nodes {
		cl := classifications[i]

		// Skip chassis (already handled) and alias aggregators.
		if cl.Category == CategoryDevice && cl.DeviceTypeHint == devicetypes.TypeChassis {
			continue
		}
		if cl.Category == CategorySkip {
			continue
		}

		switch cl.Category {
		case CategoryLocation:
			loc := buildLocationFromNode(node)
			result.Locations[loc.ID] = &loc

		case CategoryDevice:
			dev, frus := buildDeviceFromNode(node, cl, existing)
			slug, mq, ms := enrichDeviceFromLibrary(&dev, frus, cl)
			if slug == "" {
				unmatched = append(unmatched, newUnmatched(node, cl))
			}
			result.Devices[dev.ID] = &dev
			addFrus(result, frus)

			rack := buildRack(node, racksByNumber, existing)
			if rack != nil {
				result.Racks[rack.ID] = rack
				tally.Racks++
			}
			assignRack(&dev, node, racksByNumber, result)

			tally.Devices++
			tally.Cables += len(frus)
			if stepMode {
				info := buildNodeStepInfo(i+1, len(nodes), node, &dev, frus, slug, mq, ms)
				if err := visual.PromptNodeTransformStep(info, tally, opts); err != nil {
					return nil, fmt.Errorf("step interrupted: %w", err)
				}
			}

		case CategoryModule:
			mod, frus := buildModuleFromNode(node, cl, chassisByLoc, chassisByXname)
			slug, _, _ := enrichModuleFromLibrary(&mod, frus, cl)
			if slug == "" {
				unmatched = append(unmatched, newUnmatched(node, cl))
			}
			result.Modules[mod.ID] = &mod

			// FRUs belong to the module they were discovered on.
			for j := range frus {
				frus[j].Device = mod.ID
				frus[j].Parent = mod.ID
				result.Frus[frus[j].ID] = &frus[j]
			}
			tally.Cables += len(frus)
		}
	}

	logUnmatched(unmatched)

	log.Printf("Transformed %d nodes → %d devices, %d modules, %d locations, %d racks, %d FRUs",
		len(nodes), len(result.Devices), len(result.Modules),
		len(result.Locations), len(result.Racks), len(result.Frus))
	return result, nil
}

// chassisKey produces a dedup key for locating a chassis by rack+chassis number.
func chassisKey(rack, chassis int32) string {
	return fmt.Sprintf("%d-%d", rack, chassis)
}

// addFrus inserts FRU entries into the result.
func addFrus(result *devicetypes.TransformResult, frus []devicetypes.CaniFruType) {
	for j := range frus {
		result.Frus[frus[j].ID] = &frus[j]
	}
}

// assignRack sets the Rack and Parent FKs on a device and re-stores the pointer.
// Parent is set to the rack UUID so downstream consumers (e.g. the Nautobot mapper)
// that resolve rack placement via device.Parent also work correctly.
func assignRack(dev *devicetypes.CaniDeviceType, node import_.Node, racksByNumber map[int32]uuid.UUID, result *devicetypes.TransformResult) {
	if node.Location == nil || node.Location.Rack == nil {
		return
	}
	if rackID, ok := racksByNumber[*node.Location.Rack]; ok {
		dev.Rack = rackID
		dev.Parent = rackID
		result.Devices[dev.ID] = dev
	}
}

// newUnmatched creates an unmatched record for logging.
func newUnmatched(node import_.Node, cl Classification) UnmatchedNode {
	return UnmatchedNode{
		Name:     node.Name,
		Category: cl.Category,
		Criteria: cl.Criteria,
		Queries:  cl.LookupQueries,
	}
}

// logUnmatched prints a summary table of nodes with no library match.
func logUnmatched(nodes []UnmatchedNode) {
	if len(nodes) == 0 {
		return
	}
	log.Printf("WARNING: %d node(s) had no matching library entry:", len(nodes))
	for _, n := range nodes {
		log.Printf("  %-30s  category=%-8s  hpcm_type=%-12s  queries=%v",
			n.Name, n.Category, n.Criteria["hpcm_type"], n.Queries)
	}
}

// allowedChildrenForType returns the allowed children list based on type.
func allowedChildrenForType(hint devicetypes.Type) []string {
	switch hint {
	case devicetypes.TypeChassis:
		return []string{"blade", "node", "cdu", "power-supply"}
	case devicetypes.TypeMgmtSwitch:
		return []string{"nic", "power-supply"}
	default:
		return []string{"cpu", "dimm", "disk", "gpu", "nic", "power-supply"}
	}
}

// ── Builders ────────────────────────────────────────────────────────

// buildDeviceFromNode creates a CaniDeviceType from a classified node.
// All HPCM-specific metadata is nested under the "hpcm" provider key.
func buildDeviceFromNode(node import_.Node, cl Classification, existing *devicetypes.Inventory) (devicetypes.CaniDeviceType, []devicetypes.CaniFruType) {
	hpcmMeta := make(map[string]any)
	if node.UUID != "" {
		hpcmMeta["hpcm_uuid"] = node.UUID
	}
	if len(node.Aliases) > 0 {
		hpcmMeta["aliases"] = node.Aliases
	}
	if node.Location != nil {
		hpcmMeta["location"] = locationToMap(node.Location)
	}

	id := resolveExistingDeviceID(node.Name, node.UUID, existing)

	dev := devicetypes.CaniDeviceType{
		ID:           id,
		Name:         node.Name,
		Type:         cl.DeviceTypeHint,
		HardwareType: node.Type,
		ProviderMetadata: map[string]any{
			"hpcm": hpcmMeta,
		},
		Parent:          uuid.Nil,
		AllowedChildren: allowedChildrenForType(cl.DeviceTypeHint),
	}

	// Set import source from the --node-json-file flag value.
	dev.SetImportSource("hpcm", commands.NodeJsonFile)

	frus := buildFrusFromInventory(node, dev.ID)
	return dev, frus
}

// buildModuleFromNode creates a CaniModuleType from a classified node.
// chassisByXname provides fallback parent lookup via geoloc xnames.
func buildModuleFromNode(node import_.Node, cl Classification, chassisByLoc, chassisByXname map[string]uuid.UUID) (devicetypes.CaniModuleType, []devicetypes.CaniFruType) {
	mod := devicetypes.CaniModuleType{
		ID:           uuid.New(),
		Name:         node.Name,
		HardwareType: node.Type,
		Status:       "active",
		CustomFields: make(map[string]any),
	}

	if node.UUID != "" {
		mod.CustomFields["hpcm_uuid"] = node.UUID
	}
	if len(node.Aliases) > 0 {
		mod.CustomFields["aliases"] = node.Aliases
	}
	if node.Location != nil {
		mod.CustomFields["location"] = locationToMap(node.Location)

		// Set parent device to the chassis, and module-bay from tray/node.
		if node.Location.Rack != nil && node.Location.Chassis != nil {
			key := chassisKey(*node.Location.Rack, *node.Location.Chassis)
			if parentID, ok := chassisByLoc[key]; ok {
				mod.ParentDevice = parentID
			}
		}
		mod.ModuleBayName = moduleBayName(node.Location)
	}

	// Fallback: resolve parent from geoloc xname when location-based
	// lookup did not find a chassis.
	if mod.ParentDevice == uuid.Nil {
		geo := nodeGeolocXname(node.Inventory, node.Aliases)
		mod.ParentDevice = resolveGeolocParent(geo, chassisByLoc, chassisByXname)
	}

	// Store geoloc xname in module metadata for traceability.
	if geo := nodeGeolocXname(node.Inventory, node.Aliases); geo != "" {
		mod.CustomFields["geoloc"] = geo
	}

	frus := buildFrusFromInventory(node, uuid.Nil)
	return mod, frus
}

// moduleBayName derives a bay/slot name from the location fields.
func moduleBayName(loc *import_.LocationSettings) string {
	if loc == nil {
		return ""
	}
	parts := []string{}
	if loc.Tray != nil {
		parts = append(parts, fmt.Sprintf("tray-%d", *loc.Tray))
	}
	if loc.Node != nil {
		parts = append(parts, fmt.Sprintf("node-%d", *loc.Node))
	}
	if loc.Controller != nil {
		parts = append(parts, fmt.Sprintf("ctrl-%d", *loc.Controller))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "-")
}

// buildLocationFromNode creates a CaniLocationType from an admin node.
func buildLocationFromNode(node import_.Node) devicetypes.CaniLocationType {
	return devicetypes.CaniLocationType{
		ID:           uuid.New(),
		Name:         node.Name,
		LocationType: "site",
		Status:       "active",
		CustomFields: map[string]any{"hpcm_uuid": node.UUID},
	}
}

// buildFrusFromInventory groups inventory keys into CaniFruType entries.
// parentID is the UUID of the device that owns the inventory data.
func buildFrusFromInventory(node import_.Node, parentID uuid.UUID) []devicetypes.CaniFruType {
	if node.Inventory == nil {
		return nil
	}
	groups := groupInventory(node.Inventory)
	frus := make([]devicetypes.CaniFruType, 0, len(groups))
	for id, entries := range groups {
		fru := buildCaniFru(node.Name, id, entries)
		fru.Device = parentID
		fru.Parent = parentID
		frus = append(frus, fru)
	}
	return frus
}

// locationToMap converts a LocationSettings to a provider metadata map.
func locationToMap(loc *import_.LocationSettings) map[string]any {
	m := make(map[string]any)
	if loc.Rack != nil {
		m["rack"] = *loc.Rack
	}
	if loc.Chassis != nil {
		m["chassis"] = *loc.Chassis
	}
	if loc.Tray != nil {
		m["tray"] = *loc.Tray
	}
	if loc.Node != nil {
		m["node"] = *loc.Node
	}
	if loc.Controller != nil {
		m["controller"] = *loc.Controller
	}
	return m
}

// buildNodeStepInfo constructs a NodeStepInfo showing raw HPCM fields and
// their mappings to CANI fields. Used for step-through display.
func buildNodeStepInfo(nodeNum, total int, node import_.Node, dev *devicetypes.CaniDeviceType, frus []devicetypes.CaniFruType, libSlug, matchQuery string, matchScore int) visual.NodeStepInfo {
	rawRack := ""
	if node.Location != nil && node.Location.Rack != nil {
		rawRack = fmt.Sprintf("%d", *node.Location.Rack)
	}

	mappings := []visual.FieldMapping{
		{
			SourceField: "name",
			SourceValue: node.Name,
			TargetType:  "CaniDeviceType",
			TargetField: "Name",
			TargetValue: dev.Name,
		},
		{
			SourceField: "type",
			SourceValue: node.Type,
			TargetType:  "CaniDeviceType",
			TargetField: "Type",
			TargetValue: string(dev.Type),
			IsDerived:   true,
		},
		{
			SourceField: "uuid",
			SourceValue: node.UUID,
			TargetType:  "CaniDeviceType",
			TargetField: "ProviderMetadata[hpcm][hpcm_uuid]",
			TargetValue: node.UUID,
		},
	}

	if rawRack != "" {
		mappings = append(mappings, visual.FieldMapping{
			SourceField: "location.rack",
			SourceValue: rawRack,
			TargetType:  "CaniRackType",
			TargetField: "Name",
			TargetValue: fmt.Sprintf("rack-%s", rawRack),
			IsDerived:   true,
		})
	}

	if libSlug != "" {
		mappings = append(mappings, visual.FieldMapping{
			SourceField: "(library)",
			SourceValue: libSlug,
			TargetType:  "CaniDeviceType",
			TargetField: "Slug",
			TargetValue: dev.Slug,
			IsDerived:   true,
		})
		if dev.Manufacturer != "" {
			mappings = append(mappings, visual.FieldMapping{
				SourceField: "(library)",
				SourceValue: libSlug,
				TargetType:  "CaniDeviceType",
				TargetField: "Manufacturer",
				TargetValue: dev.Manufacturer,
				IsDerived:   true,
			})
		}
		if dev.Model != "" {
			mappings = append(mappings, visual.FieldMapping{
				SourceField: "(library)",
				SourceValue: libSlug,
				TargetType:  "CaniDeviceType",
				TargetField: "Model",
				TargetValue: dev.Model,
				IsDerived:   true,
			})
		}
	}

	// Collect unique FRU group IDs for display.
	fruNames := make([]string, 0, len(frus))
	for _, f := range frus {
		fruNames = append(fruNames, f.Name)
	}

	return visual.NodeStepInfo{
		NodeNum:         nodeNum,
		Total:           total,
		RawName:         node.Name,
		RawType:         node.Type,
		RawUUID:         node.UUID,
		RawRack:         rawRack,
		FruCount:        len(frus),
		FruNames:        fruNames,
		Mappings:        mappings,
		LibMatch:        libSlug,
		MatchQuery:      matchQuery,
		MatchScore:      matchScore,
		LibModel:        dev.Model,
		LibManufacturer: dev.Manufacturer,
	}
}

// enrichDeviceFromLibrary attempts to match the device against the device
// type library. Tries ALL queries and keeps the highest-scored match.
// Returns the matched slug, or "" if no match was found.
func enrichDeviceFromLibrary(dev *devicetypes.CaniDeviceType, frus []devicetypes.CaniFruType, cl Classification) (slug, matchQuery string, matchScore int) {
	// Rebuild queries now that FRUs exist (may contain part numbers).
	queries := collectQueries(frus, import_.Node{
		Name:      dev.Name,
		Aliases:   aliasesFromMeta(dev.ProviderMetadata),
		Inventory: nil, // inventory queries already present from classification
	})
	// Prepend classification queries (SKU etc.) ahead of FRU-derived ones.
	queries = mergeQueries(cl.LookupQueries, queries)

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
			break // exact match, no need to check further
		}
	}
	if bestDT.Slug != "" {
		applyDeviceDefaults(dev, bestDT)
		return bestDT.Slug, bestQuery, bestScore
	}
	return "", "", 0
}

// enrichModuleFromLibrary attempts to match the module against the module
// type library. Tries ALL queries and keeps the highest-scored match.
// Returns the matched slug, or "" if no match was found.
func enrichModuleFromLibrary(mod *devicetypes.CaniModuleType, frus []devicetypes.CaniFruType, cl Classification) (slug, matchQuery string, matchScore int) {
	queries := cl.LookupQueries
	// Also try FRU part numbers.
	for _, f := range frus {
		pn := strings.TrimSpace(f.PartNumber)
		if pn != "" {
			queries = append(queries, pn)
		}
	}

	// Try module library first — keep best score.
	var bestMT devicetypes.CaniModuleType
	var bestQuery string
	bestScore := 0
	for _, q := range queries {
		mt, score := devicetypes.LookupModuleScored(q)
		if score > bestScore {
			bestScore = score
			bestMT = mt
			bestQuery = q
		}
		if bestScore >= 100 {
			break
		}
	}
	if bestMT.Slug != "" {
		applyModuleDefaults(mod, bestMT)
		return bestMT.Slug, bestQuery, bestScore
	}

	// Fall back to device library (some modules are listed as device-types).
	var bestDT devicetypes.CaniDeviceType
	bestQuery = ""
	bestScore = 0
	for _, q := range queries {
		dt, score := devicetypes.LookupScored(q)
		if score > bestScore {
			bestScore = score
			bestDT = dt
			bestQuery = q
		}
		if bestScore >= 100 {
			break
		}
	}
	if bestDT.Slug != "" {
		mod.Slug = bestDT.Slug
		if mod.Manufacturer == "" {
			mod.Manufacturer = bestDT.Manufacturer
		}
		if mod.Model == "" {
			mod.Model = bestDT.Model
		}
		return bestDT.Slug, bestQuery, bestScore
	}
	return "", "", 0
}

// mergeQueries returns a combined list with a's entries first, deduped.
func mergeQueries(a, b []string) []string {
	seen := make(map[string]bool, len(a)+len(b))
	out := make([]string, 0, len(a)+len(b))
	for _, s := range a {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	for _, s := range b {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// aliasesFromMeta extracts the aliases map from provider metadata.
func aliasesFromMeta(meta map[string]any) map[string]string {
	if meta == nil {
		return nil
	}
	v, ok := meta["aliases"]
	if !ok {
		return nil
	}
	m, ok := v.(map[string]string)
	if ok {
		return m
	}
	return nil
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
	if dev.HardwareType == "" || dev.HardwareType == string(dev.Type) {
		if lib.HardwareType != "" {
			dev.HardwareType = lib.HardwareType
		}
	}
}

// applyModuleDefaults copies non-empty library fields into the module.
func applyModuleDefaults(mod *devicetypes.CaniModuleType, lib devicetypes.CaniModuleType) {
	if mod.Slug == "" {
		mod.Slug = lib.Slug
	}
	if mod.Manufacturer == "" {
		mod.Manufacturer = lib.Manufacturer
	}
	if mod.Model == "" {
		mod.Model = lib.Model
	}
	if mod.Description == "" {
		mod.Description = lib.Description
	}
	if mod.HardwareType == "" {
		mod.HardwareType = lib.HardwareType
	}
}

// buildRack creates a CaniRackType from the node's location.rack field.
// Uses racksByNumber for deduplication; returns nil if already created or
// if the node has no rack location.
func buildRack(node import_.Node, racksByNumber map[int32]uuid.UUID, existing *devicetypes.Inventory) *devicetypes.CaniRackType {
	if node.Location == nil || node.Location.Rack == nil {
		return nil
	}
	rackNum := *node.Location.Rack
	if _, exists := racksByNumber[rackNum]; exists {
		return nil
	}

	name := fmt.Sprintf("rack-%d", rackNum)
	id := resolveExistingRackID(name, existing)

	rack := &devicetypes.CaniRackType{
		ID:               id,
		Name:             name,
		HardwareType:     string(devicetypes.TypeRack),
		Status:           "active",
		UHeight:          42,
		ProviderMetadata: map[string]any{"rack_number": rackNum},
	}
	racksByNumber[rackNum] = rack.ID
	return rack
}

// fruPrefixes lists the inventory key prefixes that produce FRU entries.
var fruPrefixes = []string{
	"disk.", "nic.", "cpu.", "dimm.",
	"gpu.", "fru.", "fw.", "bios.",
	"sys.", "board.", "chassis.", "cdu.",
}

// kvEntry holds a key-value pair from the inventory map.
type kvEntry struct {
	key   string
	value string
}

// groupInventory groups inventory keys by component identifier.
// Three-part keys (e.g. "disk.disk0.model") are grouped by "disk.disk0".
// Two-part keys (e.g. "gpu.vendor") are grouped by "gpu".
func groupInventory(inv map[string]string) map[string][]kvEntry {
	groups := make(map[string][]kvEntry)
	for k, v := range inv {
		if !hasKnownPrefix(k) {
			continue
		}
		groupID, field := splitInventoryKey(k)
		if groupID == "" || field == "" {
			continue
		}
		groups[groupID] = append(groups[groupID], kvEntry{key: field, value: v})
	}
	return groups
}

// hasKnownPrefix checks if a key starts with any recognized prefix.
func hasKnownPrefix(key string) bool {
	for _, p := range fruPrefixes {
		if strings.HasPrefix(key, p) {
			return true
		}
	}
	return false
}

// splitInventoryKey splits an inventory key into group ID and field name.
// "disk.disk0.model" -> ("disk.disk0", "model")
// "gpu.vendor"       -> ("gpu", "vendor")
func splitInventoryKey(key string) (string, string) {
	parts := strings.SplitN(key, ".", 3)
	switch len(parts) {
	case 3:
		return parts[0] + "." + parts[1], parts[2]
	case 2:
		return parts[0], parts[1]
	default:
		return "", ""
	}
}

// buildCaniFru constructs a CaniFruType from grouped inventory entries.
func buildCaniFru(parentName, groupID string, entries []kvEntry) devicetypes.CaniFruType {
	hwType := strings.SplitN(groupID, ".", 2)[0]
	fru := devicetypes.CaniFruType{
		ID:           uuid.New(),
		Name:         fmt.Sprintf("%s-%s", parentName, groupID),
		HardwareType: hwType,
		Status:       "active",
		Discovered:   true,
		CustomFields: make(map[string]any),
	}

	for _, e := range entries {
		switch normalizeField(e.key) {
		case "serial":
			fru.Serial = e.value
		case "part_number":
			fru.PartNumber = e.value
		case "model":
			fru.Model = e.value
		case "manufacturer":
			fru.Manufacturer = e.value
		case "name":
			fru.Label = e.value
		default:
			fru.CustomFields[e.key] = e.value
		}
	}

	return fru
}

// normalizeField maps common inventory field name variants to canonical names.
func normalizeField(field string) string {
	switch field {
	case "serialNumber", "Serial Number", "serial_number", "SerialNumber":
		return "serial"
	case "partNumber", "Part Number", "part_number", "PartNumber":
		return "part_number"
	case "model", "model_name", "Model":
		return "model"
	case "vendor", "Vendor", "manufacturer", "Manufacturer":
		return "manufacturer"
	case "name", "info_name", "Name":
		return "name"
	default:
		return field
	}
}
