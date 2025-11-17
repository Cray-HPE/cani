package transform

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/Cray-HPE/cani/pkg/visual"
	"github.com/google/uuid"
)

// Cable type constants for common cable types.
const (
	cableTypeCat5e      = "cat5e"
	cableTypeCat6       = "cat6"
	cableTypeCat6a      = "cat6a"
	cableTypeDacPassive = "dac-passive"
	cableTypeAoc        = "aoc"
	cableTypeMmfOm4     = "mmf-om4"
	cableTypeSmf        = "smf"
	cableTypePower      = "power"
	cableTypeOther      = "other"
)

// Cable type inference patterns (case-insensitive).
var (
	cat5Pattern  = regexp.MustCompile(`(?i)cat5e?`)
	cat6Pattern  = regexp.MustCompile(`(?i)cat6a?`)
	cat6aPattern = regexp.MustCompile(`(?i)cat6a`)
	dacPattern   = regexp.MustCompile(`(?i)dac|direct.?attach`)
	aocPattern   = regexp.MustCompile(`(?i)aoc|active.?optical`)
	mmfPattern   = regexp.MustCompile(`(?i)om[34]|mmf|\bMM\b`)
	smfPattern   = regexp.MustCompile(`(?i)smf|single.?mode|\bSM\b`)
	powerPattern = regexp.MustCompile(`(?i)power.?cord|jumper`)
)

// resolveCableTypeSlug resolves the cable type slug using the cascade:
// 1. Lookup by part number in cable type library
// 2. Lookup by generated slug in cable type library
// 3. Infer from description patterns
func resolveCableTypeSlug(partNumber, description string) string {
	// 1. Try part number lookup first
	if partNumber != "" {
		if ct, ok := devicetypes.GetCableTypeByPartNumber(partNumber); ok {
			return ct.Slug
		}
	}

	// 2. Try slug lookup (generate slug from description)
	slug := slugify(description)
	if ct, ok := devicetypes.GetCableTypeBySlug(slug); ok {
		return ct.Slug
	}

	// 3. Fall back to description pattern inference
	return inferCableTypeSlug(description)
}

// inferCableTypeSlug derives cable type slug from description patterns.
// Uses patterns from .forge.md cable handling specification.
func inferCableTypeSlug(description string) string {
	switch {
	case cat6aPattern.MatchString(description):
		return cableTypeCat6a
	case cat6Pattern.MatchString(description):
		return cableTypeCat6
	case cat5Pattern.MatchString(description):
		return cableTypeCat5e
	case dacPattern.MatchString(description):
		return cableTypeDacPassive
	case aocPattern.MatchString(description):
		return cableTypeAoc
	case mmfPattern.MatchString(description):
		return cableTypeMmfOm4
	case smfPattern.MatchString(description):
		return cableTypeSmf
	case powerPattern.MatchString(description):
		return cableTypePower
	default:
		return cableTypeOther
	}
}

// transformCables creates cable connections from pre-sorted cable records.
// Handles two types of cable records:
// 1. Explicit cable connections (SourceDevice/DestDevice columns present)
// 2. Cable products (detected by hardware type inference from description)
// For cable products, creates N individual CaniCableType objects per quantity.
// Auto-connects cables to devices based on config group relationships.
// Accepts step mode state from the caller for unified progress tracking.
func transformCables(
	inventory *devicetypes.Inventory,
	records []import_.CsvRecord,
	stepMode bool,
	opts visual.ETLOptions,
	tally *visual.StepTally,
	recordNum *int,
	totalRecords int,
) (map[uuid.UUID]*devicetypes.CaniCableType, error) {
	cables := make(map[uuid.UUID]*devicetypes.CaniCableType)

	if len(records) == 0 {
		return cables, nil
	}

	// Ensure inventory.Cables is initialized
	if inventory.Cables == nil {
		inventory.Cables = make(map[uuid.UUID]*devicetypes.CaniCableType)
	}

	// Separate explicit cable connection records and cable product records
	var explicitRecords []import_.CsvRecord
	var productRecords []import_.CsvRecord

	for _, rec := range records {
		if import_.IsCableRecord(rec) {
			explicitRecords = append(explicitRecords, rec)
		} else {
			productRecords = append(productRecords, rec)
		}
	}

	// Process explicit cable connection records (have SourceDevice/DestDevice)
	for i, rec := range explicitRecords {
		*recordNum++
		cable, err := createCableFromExplicitRecord(inventory, rec)
		if err != nil {
			return nil, fmt.Errorf("explicit cable record %d: %w", i+1, err)
		}

		cables[cable.ID] = cable
		inventory.Cables[cable.ID] = cable

		if err := linkInterfacesToCable(inventory, cable); err != nil {
			return nil, fmt.Errorf("explicit cable record %d: failed to link interfaces: %w", i+1, err)
		}

		// Show step output if step mode is enabled
		if stepMode {
			tally.Cables++
			stepInfo := buildCableStepInfo(rec, cable)
			if err := visual.PromptTransformStep(*recordNum, totalRecords, stepInfo, *tally, opts); err != nil {
				return nil, fmt.Errorf("step interrupted: %w", err)
			}
		}
	}

	// Process cable product records (create N individual cables per quantity)
	// Group cables by config group for auto-connect
	cablesByGroup := make(map[string][]*devicetypes.CaniCableType)

	for _, rec := range productRecords {
		*recordNum++
		cableTypeSlug := resolveCableTypeSlug(rec.PartNumber, rec.Description)

		// Create N individual CaniCableType objects
		var createdCables []*devicetypes.CaniCableType
		for i := 0; i < rec.Quantity; i++ {
			label := generateCableLabel(rec.Description, i, rec.Quantity)
			cable := devicetypes.NewCable(cableTypeSlug, label)

			// Parse length from description if present
			if length, unit := parseLengthFromDescription(rec.Description); length > 0 {
				cable.Length = &length
				cable.LengthUnit = unit
			}

			cables[cable.ID] = cable
			inventory.Cables[cable.ID] = cable
			createdCables = append(createdCables, cable)

			// Track by config group for auto-connect
			if rec.ConfigGroup != "" {
				cablesByGroup[rec.ConfigGroup] = append(cablesByGroup[rec.ConfigGroup], cable)
			}
		}

		// Show step output for cable product record
		if stepMode && len(createdCables) > 0 {
			tally.Cables += len(createdCables)
			stepInfo := buildCableProductStepInfo(rec, createdCables)
			if err := visual.PromptTransformStep(*recordNum, totalRecords, stepInfo, *tally, opts); err != nil {
				return nil, fmt.Errorf("step interrupted: %w", err)
			}
		}
	}

	// Auto-connect cables to devices based on config group relationships
	if len(cablesByGroup) > 0 {
		autoConnectCables(inventory, cablesByGroup)
	}

	return cables, nil
}

// generateCableLabel creates a label for a cable instance.
func generateCableLabel(description string, index, total int) string {
	if total == 1 {
		return description
	}
	return fmt.Sprintf("%s-%03d", description, index+1)
}

// parseLengthFromDescription extracts cable length from description (e.g., "3m", "10ft").
// Uses word boundary to avoid matching "M/M" patterns.
func parseLengthFromDescription(description string) (float64, string) {
	// Match number followed by unit at word boundary (not M/M pattern)
	re := regexp.MustCompile(`\b(\d+(?:\.\d+)?)\s*(m|ft|cm|in)\b`)
	matches := re.FindStringSubmatch(strings.ToLower(description))
	if len(matches) < 3 {
		return 0, ""
	}

	length, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, ""
	}

	return length, matches[2]
}

// createCableFromExplicitRecord creates a cable from a record with explicit endpoints.
// Renamed from createCableFromRecord to clarify purpose.
func createCableFromExplicitRecord(
	inventory *devicetypes.Inventory,
	rec import_.CsvRecord,
) (*devicetypes.CaniCableType, error) {
	srcDevice := findDeviceByName(inventory, rec.SourceDevice)
	if srcDevice == nil {
		return nil, fmt.Errorf("source device %q not found", rec.SourceDevice)
	}

	srcIface := srcDevice.GetInterface(rec.SourcePort)
	if srcIface == nil {
		return nil, fmt.Errorf("interface %q not found on device %q", rec.SourcePort, rec.SourceDevice)
	}

	dstDevice := findDeviceByName(inventory, rec.DestDevice)
	if dstDevice == nil {
		return nil, fmt.Errorf("destination device %q not found", rec.DestDevice)
	}

	dstIface := dstDevice.GetInterface(rec.DestPort)
	if dstIface == nil {
		return nil, fmt.Errorf("interface %q not found on device %q", rec.DestPort, rec.DestDevice)
	}

	cableType := rec.CableType
	if cableType == "" {
		cableType = inferCableType(srcIface.Type)
	}

	label := fmt.Sprintf("%s:%s ↔ %s:%s",
		rec.SourceDevice, rec.SourcePort,
		rec.DestDevice, rec.DestPort)

	cable := devicetypes.NewCable(cableType, label)
	cable.SetTerminations(srcIface.ID, dstIface.ID)
	cable.SetDeviceTerminations(srcDevice.ID, dstDevice.ID, rec.SourcePort, rec.DestPort)

	if rec.CableLength != "" {
		length, unit := parseCableLength(rec.CableLength)
		if length > 0 {
			cable.Length = &length
			cable.LengthUnit = unit
		}
	}

	return cable, nil
}

// autoConnectCables connects cables to devices based on config group relationships.
// Per .forge.md: cables in config group 09XX connect to devices in groups 02XX, 03XX.
// Uses hub-spoke topology: switches are hubs, other devices are spokes.
func autoConnectCables(inventory *devicetypes.Inventory, cablesByGroup map[string][]*devicetypes.CaniCableType) {
	// Get devices grouped by config group from provider metadata
	devicesByGroup := groupDevicesByConfigGroup(inventory)

	for cableGroup, cables := range cablesByGroup {
		// Find related device groups (e.g., 0900 cables → 0200, 0300 devices)
		relatedGroups := findRelatedDeviceGroups(cableGroup, devicesByGroup)
		if len(relatedGroups) == 0 {
			continue
		}

		// Collect all devices from related groups
		var devices []*devicetypes.CaniDeviceType
		for _, group := range relatedGroups {
			devices = append(devices, devicesByGroup[group]...)
		}

		if len(devices) == 0 {
			continue
		}

		// Find hub devices (switches) and spoke devices (servers, nodes)
		var hubs, spokes []*devicetypes.CaniDeviceType
		for _, dev := range devices {
			if dev.HardwareType == "switch" || dev.HardwareType == "mgmt-switch" || dev.HardwareType == "hsn-switch" {
				hubs = append(hubs, dev)
			} else {
				spokes = append(spokes, dev)
			}
		}

		// Auto-connect cables between hubs and spokes
		connectCablesHubSpoke(inventory, cables, hubs, spokes)
	}
}

// groupDevicesByConfigGroup groups devices by their config group from provider metadata.
func groupDevicesByConfigGroup(inventory *devicetypes.Inventory) map[string][]*devicetypes.CaniDeviceType {
	result := make(map[string][]*devicetypes.CaniDeviceType)

	for _, device := range inventory.Devices {
		if device == nil || device.ProviderMetadata == nil {
			continue
		}

		// Look for config group in provider metadata
		if exampleMeta, ok := device.ProviderMetadata["example"].(map[string]any); ok {
			if configGroup, ok := exampleMeta["ConfigGroup"].(string); ok && configGroup != "" {
				result[configGroup] = append(result[configGroup], device)
			}
		}
	}

	return result
}

// findRelatedDeviceGroups finds device config groups that should be connected to a cable group.
// Per .forge.md: cables in 09XX connect to devices in 02XX, 03XX, etc.
func findRelatedDeviceGroups(cableGroup string, devicesByGroup map[string][]*devicetypes.CaniDeviceType) []string {
	var related []string

	// Get numeric prefix of cable group (e.g., "0900" → 9)
	if len(cableGroup) < 2 {
		return related
	}

	// Find all non-rack device groups (not 01XX which are racks)
	for group := range devicesByGroup {
		if len(group) >= 2 && group[:2] != "01" {
			// Skip the cable group itself
			if group != cableGroup {
				related = append(related, group)
			}
		}
	}

	return related
}

// connectCablesHubSpoke connects cables between hub (switch) and spoke (server) devices.
func connectCablesHubSpoke(
	inventory *devicetypes.Inventory,
	cables []*devicetypes.CaniCableType,
	hubs, spokes []*devicetypes.CaniDeviceType,
) {
	if len(hubs) == 0 || len(spokes) == 0 {
		return
	}

	cableIdx := 0
	for _, spoke := range spokes {
		if cableIdx >= len(cables) {
			break
		}

		// Find an available interface on the spoke
		spokeIface := findAvailableInterface(inventory, spoke)
		if spokeIface == nil {
			continue
		}

		// Find an available interface on a hub
		var hubIface *devicetypes.InterfaceSpec
		var hub *devicetypes.CaniDeviceType
		for _, h := range hubs {
			hubIface = findAvailableInterface(inventory, h)
			if hubIface != nil {
				hub = h
				break
			}
		}

		if hubIface == nil {
			continue
		}

		// Connect the cable
		cable := cables[cableIdx]
		cable.SetTerminations(spokeIface.ID, hubIface.ID)
		cable.SetDeviceTerminations(spoke.ID, hub.ID, spokeIface.Name, hubIface.Name)
		cable.Label = fmt.Sprintf("%s:%s ↔ %s:%s",
			spoke.Name, spokeIface.Name,
			hub.Name, hubIface.Name)

		// Update interface connection state
		spokeIface.ConnectedCable = &cable.ID
		hubIface.ConnectedCable = &cable.ID

		cableIdx++
	}
}

// findAvailableInterface finds an unconnected interface on a device.
func findAvailableInterface(inventory *devicetypes.Inventory, device *devicetypes.CaniDeviceType) *devicetypes.InterfaceSpec {
	for i := range device.Interfaces {
		iface := &device.Interfaces[i]
		if iface.ConnectedCable == nil {
			return iface
		}
	}
	return nil
}

// findDeviceByName searches inventory for a device with the given name.
func findDeviceByName(inventory *devicetypes.Inventory, name string) *devicetypes.CaniDeviceType {
	for _, device := range inventory.Devices {
		if device.Name == name {
			return device
		}
	}
	return nil
}

// linkInterfacesToCable updates the ConnectedCable pointer on both interfaces.
func linkInterfacesToCable(inventory *devicetypes.Inventory, cable *devicetypes.CaniCableType) error {
	ifaceA, _ := inventory.GetInterfaceByID(cable.TerminationA)
	if ifaceA == nil {
		return fmt.Errorf("interface %s not found", cable.TerminationA)
	}
	ifaceA.ConnectedCable = &cable.ID

	ifaceB, _ := inventory.GetInterfaceByID(cable.TerminationB)
	if ifaceB == nil {
		return fmt.Errorf("interface %s not found", cable.TerminationB)
	}
	ifaceB.ConnectedCable = &cable.ID

	return nil
}

// inferCableType guesses cable type from interface type.
func inferCableType(ifaceType devicetypes.InterfacesElemType) string {
	switch ifaceType {
	case devicetypes.InterfacesElemTypeA1000BaseT:
		return cableTypeCat6
	case devicetypes.InterfacesElemTypeA10GbaseT:
		return cableTypeCat6a
	case devicetypes.InterfacesElemTypeA10GbaseXSfpp,
		devicetypes.InterfacesElemTypeA25GbaseXSfp28,
		devicetypes.InterfacesElemTypeA40GbaseXQsfpp,
		devicetypes.InterfacesElemTypeA100GbaseXQsfp28,
		devicetypes.InterfacesElemTypeA400GbaseXQsfpdd:
		return cableTypeDacPassive
	default:
		return cableTypeOther
	}
}

// parseCableLength extracts numeric length and unit from a string like "3m" or "10ft".
func parseCableLength(s string) (float64, string) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 0, ""
	}

	re := regexp.MustCompile(`^([\d.]+)\s*([a-z]*)$`)
	matches := re.FindStringSubmatch(s)
	if len(matches) < 2 {
		return 0, ""
	}

	length, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, ""
	}

	unit := "m"
	if len(matches) >= 3 && matches[2] != "" {
		unit = matches[2]
	}

	return length, unit
}

// buildCableStepInfo creates step info for an explicit cable connection record.
func buildCableStepInfo(rec import_.CsvRecord, cable *devicetypes.CaniCableType) visual.TransformStepInfo {
	info := visual.TransformStepInfo{
		Quantity: 1,
		HwType:   "cable",
		Mappings: []visual.FieldMapping{
			{
				SourceField: "SourceDevice:Port",
				SourceValue: rec.SourceDevice + ":" + rec.SourcePort,
				TargetType:  "CaniCableType",
				TargetField: "TerminationA",
				TargetValue: cable.TerminationA.String()[:8],
				IsDerived:   true,
			},
			{
				SourceField: "DestDevice:Port",
				SourceValue: rec.DestDevice + ":" + rec.DestPort,
				TargetType:  "CaniCableType",
				TargetField: "TerminationB",
				TargetValue: cable.TerminationB.String()[:8],
				IsDerived:   true,
			},
			{
				SourceField: "CableType",
				SourceValue: rec.CableType,
				TargetType:  "CaniCableType",
				TargetField: "CableTypeSlug",
				TargetValue: cable.Slug,
				IsDerived:   rec.CableType == "",
			},
		},
		CreatedItems: []visual.CreatedItemInfo{
			{
				ID:   cable.ID.String()[:8],
				Name: cable.Label,
			},
		},
	}

	return info
}

// buildCableProductStepInfo creates step info for a cable product record (quantity N).
func buildCableProductStepInfo(rec import_.CsvRecord, cables []*devicetypes.CaniCableType) visual.TransformStepInfo {
	info := visual.TransformStepInfo{
		Quantity: rec.Quantity,
		HwType:   "cable",
		Mappings: []visual.FieldMapping{
			{
				SourceField: "PartNumber",
				SourceValue: rec.PartNumber,
				TargetType:  "CaniCableType",
				TargetField: "CableTypeSlug",
				TargetValue: cables[0].Slug,
				IsDerived:   false,
			},
			{
				SourceField: "Description",
				SourceValue: rec.Description,
				TargetType:  "CaniCableType",
				TargetField: "Label",
				TargetValue: cables[0].Label,
				IsDerived:   true,
			},
		},
	}

	// Add created items
	for _, cable := range cables {
		info.CreatedItems = append(info.CreatedItems, visual.CreatedItemInfo{
			ID:   cable.ID.String()[:8],
			Name: cable.Label,
		})
	}

	return info
}
