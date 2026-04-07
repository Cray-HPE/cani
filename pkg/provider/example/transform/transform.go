package transform

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/example/commands"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/Cray-HPE/cani/pkg/visual"
	"github.com/google/uuid"
)

// Hardware type inference patterns (case-insensitive).
// Note: Order matters - cable patterns are checked BEFORE switch patterns
// to correctly classify cables that may contain switch manufacturer names (e.g., "Aruba AOC").
var (
	rackPattern   = regexp.MustCompile(`(?i)\d+u.*rack|cabinet`)
	cablePattern  = regexp.MustCompile(`(?i)cat5e?|cat6a?|dac|direct.?attach|aoc|active.?optical|om[34]|mmf|smf|single.?mode|power.?cord|jumper|rj45|qsfp|osfp|cable`)
	switchPattern = regexp.MustCompile(`(?i)switch|aruba`)
	nodePattern   = regexp.MustCompile(`(?i)proliant|server|dl\d{3}|blade|node`)
)

// providerGetter returns the Example singleton with raw records.
// Set by the parent package to break import cycles.
var providerGetter func() interface {
	GetRecords() []import_.CsvRecord
}

// SetProviderGetter allows the parent package to provide singleton access.
func SetProviderGetter(getter func() interface {
	GetRecords() []import_.CsvRecord
}) {
	providerGetter = getter
}

// classifiedRecords holds records categorized by type.
type classifiedRecords struct {
	racks   []import_.CsvRecord
	devices []import_.CsvRecord
	cables  []import_.CsvRecord
}

// classifyRecords categorizes all records into racks, devices, and cables.
// Checks device type library by part number first, then falls back to pattern matching.
// Returns an error if any record cannot be classified.
func classifyRecords(records []import_.CsvRecord) (*classifiedRecords, error) {
	result := &classifiedRecords{}

	for i, rec := range records {
		// Check for explicit cable connection records first
		if import_.IsCableRecord(rec) {
			result.cables = append(result.cables, rec)
			continue
		}

		// Try device type library lookup by part number first (authoritative)
		var hwType string
		if rec.PartNumber != "" {
			if dt, ok := devicetypes.GetByPartNumber(rec.PartNumber); ok && dt.HardwareType != "" {
				hwType = dt.HardwareType
			}
		}

		// Fall back to pattern matching on description
		if hwType == "" {
			hwType = inferHardwareType(rec.Description)
		}

		switch hwType {
		case "rack":
			result.racks = append(result.racks, rec)
		case "cable":
			result.cables = append(result.cables, rec)
		case "switch", "node", "blade", "chassis", "pdu", "cdu":
			result.devices = append(result.devices, rec)
		case "":
			return nil, fmt.Errorf("record %d: cannot classify hardware type for %q", i+1, rec.Description)
		default:
			// Unknown but non-empty hardware type from library - treat as device
			result.devices = append(result.devices, rec)
		}
	}

	return result, nil
}

// Transform converts raw records into CaniDeviceType, CaniRackType, and CaniCableType objects.
// Uses three passes: racks first, then devices (with parenting), then cables (with linking).
// Returns a TransformResult containing all created items.
func Transform(existing devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	// Reset rack position tracking for fresh import
	resetRackPositionStates()

	if providerGetter == nil {
		return nil, fmt.Errorf("transform: providerGetter not set")
	}

	prov := providerGetter()
	allRecords := prov.GetRecords()

	// Pre-sort records by type
	classified, err := classifyRecords(allRecords)
	if err != nil {
		return nil, fmt.Errorf("classifyRecords: %w", err)
	}

	initInventoryMaps(&existing)

	// Track created items by config group for parenting
	racksByGroup := make(map[string][]uuid.UUID)
	devicesByGroup := make(map[string][]uuid.UUID)

	// Track newly created items for result
	createdRacks := make(map[uuid.UUID]*devicetypes.CaniRackType)
	createdDevices := make(map[uuid.UUID]*devicetypes.CaniDeviceType)

	// Setup step mode options
	stepMode := config.Cfg.StepMode
	opts := visual.ETLOptions{NoColor: config.Cfg.NoColor}
	tally := visual.StepTally{}
	totalRecords := len(classified.racks) + len(classified.devices) + len(classified.cables)
	recordNum := 0

	// Pass 1: Create racks
	for _, rec := range classified.racks {
		recordNum++
		created := createItemsFromRecord(&existing, rec, "rack", racksByGroup, devicesByGroup)

		for _, rack := range created.Racks {
			createdRacks[rack.ID] = rack
		}

		if stepMode && len(created.Racks) > 0 {
			tally.Racks += len(created.Racks)
			stepInfo := buildTransformStepInfo(rec, "rack", created)
			if err := visual.PromptTransformStep(recordNum, totalRecords, stepInfo, tally, opts); err != nil {
				return nil, fmt.Errorf("step interrupted: %w", err)
			}
		}
	}

	// Pass 2: Create devices
	for _, rec := range classified.devices {
		recordNum++

		// Try device type library lookup by part number first (authoritative)
		var hwType string
		if rec.PartNumber != "" {
			if dt, ok := devicetypes.GetByPartNumber(rec.PartNumber); ok && dt.HardwareType != "" {
				hwType = dt.HardwareType
			}
		}

		// Fall back to pattern matching on description
		if hwType == "" {
			hwType = inferHardwareType(rec.Description)
		}

		created := createItemsFromRecord(&existing, rec, hwType, racksByGroup, devicesByGroup)

		for _, device := range created.Devices {
			createdDevices[device.ID] = device
		}

		if stepMode && len(created.Devices) > 0 {
			tally.Devices += len(created.Devices)
			stepInfo := buildTransformStepInfo(rec, hwType, created)
			if err := visual.PromptTransformStep(recordNum, totalRecords, stepInfo, tally, opts); err != nil {
				return nil, fmt.Errorf("step interrupted: %w", err)
			}
		}
	}

	// Establish parent-child relationships via config groups
	assignConfigGroupParenting(&existing, racksByGroup, devicesByGroup)

	// Pass 3: Create cables
	createdCables, err := transformCables(&existing, classified.cables, stepMode, opts, &tally, &recordNum, totalRecords)
	if err != nil {
		return nil, fmt.Errorf("transformCables: %w", err)
	}

	log.Printf("Transformed: %d racks, %d devices, %d cables",
		len(createdRacks), len(createdDevices), len(createdCables))

	return &devicetypes.TransformResult{
		Racks:   createdRacks,
		Devices: createdDevices,
		Cables:  createdCables,
	}, nil
}

// inferHardwareType determines hardware type from product description.
// Note: Cable pattern is checked BEFORE switch pattern to correctly classify
// cables that may contain switch manufacturer names (e.g., "HPE Aruba 100G AOC").
// Returns empty string if no pattern matches, allowing fallback to library lookup.
func inferHardwareType(description string) string {
	switch {
	case rackPattern.MatchString(description):
		return "rack"
	case cablePattern.MatchString(description):
		return "cable"
	case switchPattern.MatchString(description):
		return "switch"
	case nodePattern.MatchString(description):
		return "node"
	default:
		return ""
	}
}

// initInventoryMaps ensures inventory maps are initialized.
func initInventoryMaps(inventory *devicetypes.Inventory) {
	if inventory.Racks == nil {
		inventory.Racks = make(map[uuid.UUID]*devicetypes.CaniRackType)
	}
	if inventory.Devices == nil {
		inventory.Devices = make(map[uuid.UUID]*devicetypes.CaniDeviceType)
	}
	if inventory.Cables == nil {
		inventory.Cables = make(map[uuid.UUID]*devicetypes.CaniCableType)
	}
}

// CreatedItems holds items created from a single CSV record for step display.
type CreatedItems struct {
	Racks   []*devicetypes.CaniRackType
	Devices []*devicetypes.CaniDeviceType
}

// createItemsFromRecord creates inventory items from a single CSV record.
// Returns the created items for step-through display.
func createItemsFromRecord(
	inventory *devicetypes.Inventory,
	rec import_.CsvRecord,
	hwType string,
	racksByGroup map[string][]uuid.UUID,
	devicesByGroup map[string][]uuid.UUID,
) CreatedItems {
	result := CreatedItems{}

	for i := 0; i < rec.Quantity; i++ {
		id := uuid.New()
		name := generateName(rec.Description, i, rec.Quantity)

		if hwType == "rack" {
			rack := createRack(inventory, id, name, rec, racksByGroup)
			result.Racks = append(result.Racks, rack)
		} else {
			device := createDevice(inventory, id, name, rec, hwType, devicesByGroup)
			result.Devices = append(result.Devices, device)
		}
	}

	return result
}

// createRack creates a rack and adds it to inventory.
// Returns the created rack for step-through display.
func createRack(
	inventory *devicetypes.Inventory,
	id uuid.UUID,
	name string,
	rec import_.CsvRecord,
	racksByGroup map[string][]uuid.UUID,
) *devicetypes.CaniRackType {
	uHeight := 48 // default
	slug := slugify(rec.Description)

	// Look up rack type from library by part number
	if dt, ok := devicetypes.GetByPartNumber(rec.PartNumber); ok {
		slug = dt.Slug
		if dt.UHeight > 0 {
			uHeight = dt.UHeight
		}
	}

	rack := &devicetypes.CaniRackType{
		ID:         id,
		Name:       name,
		Slug:       slug,
		UHeight:    uHeight,
		ObjectMeta: devicetypes.ObjectMeta{Status: string(devicetypes.StatusActive), ProviderMetadata: buildProviderMetadata(rec)},
		Devices:    []uuid.UUID{},
	}
	inventory.Racks[id] = rack
	if rec.ConfigGroup != "" {
		racksByGroup[rec.ConfigGroup] = append(racksByGroup[rec.ConfigGroup], id)
	}
	return rack
}

// createDevice creates a device and adds it to inventory.
// Returns the created device for step-through display.
func createDevice(
	inventory *devicetypes.Inventory,
	id uuid.UUID,
	name string,
	rec import_.CsvRecord,
	hwType string,
	devicesByGroup map[string][]uuid.UUID,
) *devicetypes.CaniDeviceType {
	device := buildDeviceFromRecord(id, name, rec, hwType)
	inventory.Devices[id] = device
	if rec.ConfigGroup != "" {
		devicesByGroup[rec.ConfigGroup] = append(devicesByGroup[rec.ConfigGroup], id)
	}
	return device
}

// buildDeviceFromRecord creates a CaniDeviceType from a CSV record.
func buildDeviceFromRecord(id uuid.UUID, name string, rec import_.CsvRecord, hwType string) *devicetypes.CaniDeviceType {
	device := &devicetypes.CaniDeviceType{
		ID:           id,
		Name:         name,
		Slug:         slugify(rec.Description),
		PartNumber:   rec.PartNumber,
		HardwareType: hwType,
		ObjectMeta:   devicetypes.ObjectMeta{Status: string(devicetypes.StatusStaged), ProviderMetadata: buildProviderMetadata(rec)},
	}

	// Look up device type from library by part number
	if dt, ok := devicetypes.GetByPartNumber(rec.PartNumber); ok {
		populateFromDeviceType(device, &dt)
	}

	return device
}

// populateFromDeviceType copies fields from a registry CaniDeviceType to an inventory instance.
func populateFromDeviceType(device *devicetypes.CaniDeviceType, dt *devicetypes.CaniDeviceType) {
	device.Slug = dt.Slug
	device.Manufacturer = dt.Manufacturer
	device.Model = dt.Model
	if dt.HardwareType != "" {
		device.HardwareType = dt.HardwareType
	}
	device.Interfaces = dt.Interfaces
}

// buildProviderMetadata creates provider metadata from a record.
func buildProviderMetadata(rec import_.CsvRecord) map[string]any {
	return map[string]any{
		"example": map[string]any{
			"Source":      commands.CsvFlag,
			"PartNumber":  rec.PartNumber,
			"ConfigGroup": rec.ConfigGroup,
		},
	}
}

// assignConfigGroupParenting links devices to racks based on config group prefixes.
// Devices are sorted by rack placement priority before assignment:
// PDUs/CDUs at the bottom, nodes/blades in the middle, switches at the top.
func assignConfigGroupParenting(
	inventory *devicetypes.Inventory,
	racksByGroup map[string][]uuid.UUID,
	devicesByGroup map[string][]uuid.UUID,
) {
	parentRackIDs := findParentRackIDs(racksByGroup)
	if len(parentRackIDs) == 0 {
		return
	}

	// Collect all devices that should be linked to racks
	var allDeviceIDs []uuid.UUID
	for group, deviceIDs := range devicesByGroup {
		if shouldLinkToRacks(group) {
			allDeviceIDs = append(allDeviceIDs, deviceIDs...)
		}
	}

	// Sort by rack placement priority so positions fill bottom-up correctly:
	// PDUs/CDUs first (bottom), nodes/blades next (middle), switches last (top)
	sortDevicesByRackPriority(inventory, allDeviceIDs)

	distributeDevicesToRacks(inventory, allDeviceIDs, parentRackIDs)
}

// findParentRackIDs finds rack IDs in the 01XX config group.
func findParentRackIDs(racksByGroup map[string][]uuid.UUID) []uuid.UUID {
	for group, rackIDs := range racksByGroup {
		if getConfigGroupPrefix(group) == "01" {
			return rackIDs
		}
	}
	return nil
}

// shouldLinkToRacks returns true if devices in this group should link to racks.
func shouldLinkToRacks(group string) bool {
	prefix := getConfigGroupPrefix(group)
	return prefix != "" && prefix != "01"
}

// getConfigGroupPrefix returns the two-digit prefix of a config group.
func getConfigGroupPrefix(configGroup string) string {
	if len(configGroup) < 2 {
		return ""
	}
	return configGroup[:2]
}

// rackZone classifies where a device type belongs in a rack.
type rackZone int

const (
	zoneBottom rackZone = iota // PDUs, CDUs
	zoneMiddle                 // nodes, blades, chassis
	zoneTop                    // switches
)

// deviceTypePriority returns a sort order for rack placement.
// Lower values are placed first (bottom of rack).
func deviceTypePriority(hwType string) int {
	return int(deviceRackZone(hwType))
}

// deviceRackZone returns the rack zone for a hardware type.
func deviceRackZone(hwType string) rackZone {
	switch hwType {
	case "pdu", "cdu":
		return zoneBottom
	case "switch", "mgmt-switch", "hsn-switch":
		return zoneTop
	default:
		return zoneMiddle
	}
}

// sortDevicesByRackPriority sorts device IDs by hardware type priority.
// PDUs/CDUs sort first (bottom of rack), nodes/blades next, switches last (top).
func sortDevicesByRackPriority(inventory *devicetypes.Inventory, deviceIDs []uuid.UUID) {
	sort.SliceStable(deviceIDs, func(i, j int) bool {
		di := inventory.Devices[deviceIDs[i]]
		dj := inventory.Devices[deviceIDs[j]]
		var pi, pj int
		if di != nil {
			pi = deviceTypePriority(di.HardwareType)
		}
		if dj != nil {
			pj = deviceTypePriority(dj.HardwareType)
		}
		return pi < pj
	})
}

// groupDevicesByZone partitions device IDs into bottom, middle, and top lists.
func groupDevicesByZone(
	inventory *devicetypes.Inventory,
	deviceIDs []uuid.UUID,
) (bottom, middle, top []uuid.UUID) {
	for _, id := range deviceIDs {
		dev := inventory.Devices[id]
		if dev == nil {
			middle = append(middle, id)
			continue
		}
		switch deviceRackZone(dev.HardwareType) {
		case zoneBottom:
			bottom = append(bottom, id)
		case zoneTop:
			top = append(top, id)
		default:
			middle = append(middle, id)
		}
	}
	return bottom, middle, top
}

// rackPosState tracks bottom-up and top-down cursors for a rack.
type rackPosState struct {
	nextBottom int // next free U from bottom (ascending)
	nextTop    int // next free U from top (descending, this is the start U)
}

// rackPositionStates tracks per-rack placement cursors.
var rackPositionStates = make(map[uuid.UUID]*rackPosState)

// resetRackPositionStates clears position tracking for a fresh import.
func resetRackPositionStates() {
	rackPositionStates = make(map[uuid.UUID]*rackPosState)
}

// ensureRackState initialises the position state for a rack if needed.
func ensureRackState(rackID uuid.UUID, uHeight int) *rackPosState {
	if s, ok := rackPositionStates[rackID]; ok {
		return s
	}
	s := &rackPosState{
		nextBottom: 1,
		nextTop:    uHeight, // topmost U in the rack
	}
	rackPositionStates[rackID] = s
	return s
}

// distributeDevicesToRacks assigns devices to racks round-robin.
// Devices are placed in zone order: bottom (PDUs/CDUs), middle (nodes), top (switches).
func distributeDevicesToRacks(
	inventory *devicetypes.Inventory,
	deviceIDs []uuid.UUID,
	rackIDs []uuid.UUID,
) {
	// Ensure state exists for every target rack
	for _, rackID := range rackIDs {
		rack := inventory.Racks[rackID]
		if rack == nil {
			continue
		}
		ensureRackState(rackID, rack.UHeight)
	}

	// Partition devices by zone
	bottom, middle, top := groupDevicesByZone(inventory, deviceIDs)

	// Place zones: top first (switches from ceiling), then middle (nodes downward
	// below switches), then bottom (PDUs/CDUs upward from floor).
	distributeZone(inventory, top, rackIDs, zoneTop)
	distributeZone(inventory, middle, rackIDs, zoneMiddle)
	distributeZone(inventory, bottom, rackIDs, zoneBottom)
}

// distributeZone assigns one zone's devices round-robin across racks.
func distributeZone(
	inventory *devicetypes.Inventory,
	deviceIDs []uuid.UUID,
	rackIDs []uuid.UUID,
	zone rackZone,
) {
	for i, deviceID := range deviceIDs {
		rackIndex := i % len(rackIDs)
		rackID := rackIDs[rackIndex]
		linkDeviceToRack(inventory, deviceID, rackID, zone)
	}
}

// linkDeviceToRack sets the device's parent to the rack and assigns a U position.
func linkDeviceToRack(inventory *devicetypes.Inventory, deviceID, rackID uuid.UUID, zone rackZone) {
	device, ok := inventory.Devices[deviceID]
	if !ok {
		return
	}
	device.Parent = rackID

	rack, ok := inventory.Racks[rackID]
	if !ok {
		return
	}
	rack.Devices = append(rack.Devices, deviceID)

	assignRackPosition(device, rack, zone)
}

// assignRackPosition sets the device's U position within the rack.
// Bottom zone fills upward from U1; middle and top zones fill downward from UHeight.
func assignRackPosition(device *devicetypes.CaniDeviceType, rack *devicetypes.CaniRackType, zone rackZone) {
	uHeight := getDeviceUHeight(device)
	if uHeight <= 0 {
		uHeight = 1
	}

	state := ensureRackState(rack.ID, rack.UHeight)

	var pos int
	switch zone {
	case zoneBottom:
		// Fill upward from floor
		if state.nextBottom+uHeight-1 > state.nextTop {
			return // won't fit
		}
		pos = state.nextBottom
		state.nextBottom = pos + uHeight
	default:
		// Middle and top both fill downward from ceiling
		startU := state.nextTop - uHeight + 1
		if startU < 1 || startU < state.nextBottom {
			return // won't fit
		}
		pos = startU
		state.nextTop = startU - 1
	}

	device.RackPosition = pos
	if device.Face == "" {
		device.Face = "front"
	}
}

// getDeviceUHeight looks up the UHeight for a device from the type library.
func getDeviceUHeight(device *devicetypes.CaniDeviceType) int {
	// Try by part number first
	if device.PartNumber != "" {
		if dt, ok := devicetypes.GetByPartNumber(device.PartNumber); ok && dt.UHeight > 0 {
			return dt.UHeight
		}
	}
	// Try by slug
	if device.Slug != "" {
		if dt, ok := devicetypes.GetBySlug(device.Slug); ok && dt.UHeight > 0 {
			return dt.UHeight
		}
	}
	return 0
}

// generateName creates a device name with optional index suffix.
func generateName(description string, index, total int) string {
	name := description
	if len(name) > 30 {
		name = name[:30]
	}
	if total > 1 {
		return fmt.Sprintf("%s-%03d", name, index+1)
	}
	return name
}

// slugify converts a description to a URL-safe slug.
func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// buildTransformStepInfo creates step info for display during transform.
func buildTransformStepInfo(rec import_.CsvRecord, hwType string, created CreatedItems) visual.TransformStepInfo {
	info := visual.TransformStepInfo{
		Quantity: rec.Quantity,
		HwType:   hwType,
		Mappings: buildFieldMappings(rec, hwType, created),
	}

	// Add created item summaries
	for _, rack := range created.Racks {
		info.CreatedItems = append(info.CreatedItems, visual.CreatedItemInfo{
			ID:   rack.ID.String()[:8],
			Name: rack.Name,
		})
	}
	for _, device := range created.Devices {
		info.CreatedItems = append(info.CreatedItems, visual.CreatedItemInfo{
			ID:   device.ID.String()[:8],
			Name: device.Name,
		})
	}

	return info
}

// buildFieldMappings creates field mappings for the step display.
func buildFieldMappings(rec import_.CsvRecord, hwType string, created CreatedItems) []visual.FieldMapping {
	var mappings []visual.FieldMapping

	// Common mappings from CSV to device/rack
	mappings = append(mappings, visual.FieldMapping{
		SourceField: "PartNumber",
		SourceValue: rec.PartNumber,
		TargetType:  hwType,
		TargetField: "PartNumber",
		TargetValue: rec.PartNumber,
		IsDerived:   false,
	})

	mappings = append(mappings, visual.FieldMapping{
		SourceField: "Description",
		SourceValue: rec.Description,
		TargetType:  hwType,
		TargetField: "Name",
		TargetValue: truncateName(rec.Description),
		IsDerived:   true,
	})

	if rec.ConfigGroup != "" {
		mappings = append(mappings, visual.FieldMapping{
			SourceField: "ConfigGroup",
			SourceValue: rec.ConfigGroup,
			TargetType:  hwType,
			TargetField: "Metadata.ConfigGroup",
			TargetValue: rec.ConfigGroup,
			IsDerived:   false,
		})
	}

	// Add hardware type inference
	mappings = append(mappings, visual.FieldMapping{
		SourceField: "(inferred)",
		SourceValue: rec.Description,
		TargetType:  hwType,
		TargetField: "HardwareType",
		TargetValue: hwType,
		IsDerived:   true,
	})

	return mappings
}

// truncateName returns a truncated name suitable for display.
func truncateName(s string) string {
	if len(s) > 30 {
		return s[:30]
	}
	return s
}

// buildNewItemsSummary creates a summary of only newly created items for display.
// Uses inventory to look up parent rack names for devices.
func buildNewItemsSummary(created CreatedItems, inventory *devicetypes.Inventory) visual.ImportSummary {
	summary := visual.ImportSummary{
		RackNames:     []string{},
		DevicesByRack: make(map[string][]string),
		Cables:        []visual.CableSummary{},
	}

	// Build rack ID to name lookup from full inventory
	// This ensures we can find parent racks for devices even if the rack already existed
	racksByID := make(map[uuid.UUID]string)
	for id, rack := range inventory.Racks {
		racksByID[id] = rack.Name
	}

	// Add only newly created rack names to summary
	for _, rack := range created.Racks {
		summary.RackNames = append(summary.RackNames, rack.Name)
	}

	// Group only newly created devices by their parent rack
	for _, device := range created.Devices {
		rackName := ""
		if device.Parent != uuid.Nil {
			if name, ok := racksByID[device.Parent]; ok {
				rackName = name
			}
		}
		summary.DevicesByRack[rackName] = append(summary.DevicesByRack[rackName], device.Name)
	}

	return summary
}

// BuildTransformSummary creates a summary of transformed items for display.
func BuildTransformSummary(inventory *devicetypes.Inventory) visual.ImportSummary {
	summary := visual.ImportSummary{
		RackNames:     []string{},
		DevicesByRack: make(map[string][]string),
		Cables:        []visual.CableSummary{},
	}

	racksByID := make(map[uuid.UUID]string)
	for id, rack := range inventory.Racks {
		summary.RackNames = append(summary.RackNames, rack.Name)
		racksByID[id] = rack.Name
	}

	for _, device := range inventory.Devices {
		rackName := ""
		if device.Parent != uuid.Nil {
			if name, ok := racksByID[device.Parent]; ok {
				rackName = name
			}
		}
		summary.DevicesByRack[rackName] = append(summary.DevicesByRack[rackName], device.Name)
	}

	return summary
}
