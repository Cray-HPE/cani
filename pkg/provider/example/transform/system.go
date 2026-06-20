package transform

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/devicetypes/connections"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/google/uuid"
)

// TransformSystem converts parsed system CSV data into a TransformResult.
// Uses a 6-pass algorithm: roles → racks → devices → modules → interfaces → connections.
func TransformSystem(existing devicetypes.Inventory, data *import_.SystemCSV) (*devicetypes.TransformResult, error) {
	initInventoryMaps(&existing)

	result := &devicetypes.TransformResult{
		Locations: make(map[uuid.UUID]*devicetypes.CaniLocationType),
		Racks:     make(map[uuid.UUID]*devicetypes.CaniRackType),
		Devices:   make(map[uuid.UUID]*devicetypes.CaniDeviceType),
		Modules:   make(map[uuid.UUID]*devicetypes.CaniModuleType),
		Cables:    make(map[uuid.UUID]*devicetypes.CaniCableType),
	}

	// Pass 0: Roles and statuses (metadata catalog)
	meta, err := transformMetadata(data)
	if err != nil {
		return nil, fmt.Errorf("transformMetadata: %w", err)
	}
	result.Metadata = meta

	// Pass 0b: Locations
	locationsByName, err := transformSystemLocations(data, result, &existing)
	if err != nil {
		return nil, fmt.Errorf("transformSystemLocations: %w", err)
	}

	// Pass 1: Racks
	racksByName, err := transformSystemRacks(data, result, &existing, locationsByName)
	if err != nil {
		return nil, fmt.Errorf("transformSystemRacks: %w", err)
	}

	// Pass 2: Devices
	err = transformSystemDevices(data, result, &existing, racksByName)
	if err != nil {
		return nil, fmt.Errorf("transformSystemDevices: %w", err)
	}

	// Pass 3: Modules
	err = transformSystemModules(data, result, &existing)
	if err != nil {
		return nil, fmt.Errorf("transformSystemModules: %w", err)
	}

	// Pass 3b: Interfaces (per-interface metadata such as MAC addresses)
	if err = transformSystemInterfaces(data, result, &existing); err != nil {
		return nil, fmt.Errorf("transformSystemInterfaces: %w", err)
	}

	// Pass 4: Connections
	err = transformSystemConnections(data, result, &existing)
	if err != nil {
		return nil, fmt.Errorf("transformSystemConnections: %w", err)
	}

	// Pass 5: Validate catalog references
	if err = validateReferences(result); err != nil {
		return nil, err
	}

	log.Printf("System CSV transformed: %d roles, %d locations, %d racks, %d devices, %d modules, %d cables",
		len(data.Roles), len(result.Locations), len(result.Racks), len(result.Devices), len(result.Modules), len(result.Cables))

	return result, nil
}

// validateReferences checks that device, rack, and location role and status
// references resolve against the metadata catalog defined by the same import.
// Unknown references are logged as warnings; under strict mode any unresolved
// reference aborts the import. A catalog that is empty is treated as
// "not validated" so imports relying on export-time auto-creation still work.
func validateReferences(result *devicetypes.TransformResult) error {
	if result.Metadata == nil {
		return nil
	}
	roles := metadataNameSet(result.Metadata.Roles)
	statuses := metadataNameSet(result.Metadata.Statuses)

	var problems []string
	for _, dev := range result.Devices {
		if dev == nil {
			continue
		}
		problems = appendUnknownRef(problems, roles, dev.Role, "device", dev.Name, "role")
		problems = appendUnknownRef(problems, statuses, dev.Status, "device", dev.Name, "status")
	}
	for _, rack := range result.Racks {
		if rack != nil {
			problems = appendUnknownRef(problems, statuses, rack.Status, "rack", rack.Name, "status")
		}
	}
	for _, loc := range result.Locations {
		if loc != nil {
			problems = appendUnknownRef(problems, statuses, loc.Status, "location", loc.Name, "status")
		}
	}

	for _, p := range problems {
		log.Printf("WARN: %s", p)
	}
	if len(problems) > 0 && config.Cfg != nil && config.Cfg.Strict {
		return fmt.Errorf("system CSV has %d unresolved catalog reference(s)", len(problems))
	}
	return nil
}

// metadataNameSet builds a set of catalog entry names for membership checks.
func metadataNameSet(entries []devicetypes.MetadataEntry) map[string]bool {
	set := make(map[string]bool, len(entries))
	for _, e := range entries {
		set[e.Name] = true
	}
	return set
}

// appendUnknownRef records a problem when ref is non-empty, the catalog is
// non-empty, and ref is absent from it; an empty catalog allows any reference.
func appendUnknownRef(problems []string, catalog map[string]bool, ref, kind, name, field string) []string {
	if ref == "" || len(catalog) == 0 || catalog[ref] {
		return problems
	}
	return append(problems, fmt.Sprintf("%s %q references unknown %s %q", kind, name, field, ref))
}

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

// transformSystemLocations creates locations from system CSV location records.
// Returns a map of location name → UUID for rack parenting.
func transformSystemLocations(data *import_.SystemCSV, result *devicetypes.TransformResult, inv *devicetypes.Inventory) (map[string]uuid.UUID, error) {
	locationsByName := make(map[string]uuid.UUID)

	for _, rec := range data.Locations {
		rec = data.ApplyDefaults(rec)
		if rec.Name == "" {
			return nil, fmt.Errorf("location record missing Name")
		}
		locType := rec.LocationType
		if locType == "" {
			locType = rec.Role
		}
		if locType == "" {
			return nil, fmt.Errorf("location %q missing LocationType (e.g. dc, level, section)", rec.Name)
		}

		id := uuid.New()
		contentTypes := parseContentTypes(rec.ContentTypes)

		loc := &devicetypes.CaniLocationType{
			ID:           id,
			Name:         rec.Name,
			LocationType: locType,
			ContentTypes: contentTypes,
			ObjectMeta:   devicetypes.ObjectMeta{Status: rec.Status},
		}

		// Resolve parent by name
		if rec.Location != "" {
			parentID, ok := locationsByName[rec.Location]
			if !ok {
				parentID, ok = findLocationByName(inv, rec.Location)
			}
			if !ok {
				return nil, fmt.Errorf("location %q references unknown parent %q", rec.Name, rec.Location)
			}
			loc.Parent = parentID
		}

		result.Locations[id] = loc
		inv.Locations[id] = loc
		locationsByName[rec.Name] = id
	}

	return locationsByName, nil
}

// findLocationByName searches existing inventory for a location by name.
func findLocationByName(inv *devicetypes.Inventory, name string) (uuid.UUID, bool) {
	for id, loc := range inv.Locations {
		if loc.Name == name {
			return id, true
		}
	}
	return uuid.Nil, false
}

// transformSystemRacks creates racks from system CSV rack records.
// Returns a map of rack name → UUID for device parenting.
func transformSystemRacks(data *import_.SystemCSV, result *devicetypes.TransformResult, inv *devicetypes.Inventory, locationsByName map[string]uuid.UUID) (map[string]uuid.UUID, error) {
	racksByName := make(map[string]uuid.UUID)

	for _, rec := range data.Racks {
		rec = data.ApplyDefaults(rec)
		if rec.PartNumber == "" {
			return nil, fmt.Errorf("rack %q missing PartNumber (slug or part number)", rec.Name)
		}

		// Lookup rack type from library
		slug := rec.PartNumber
		uHeight := 48
		var partNumber, manufacturer, model string

		if rt, ok := devicetypes.GetRackTypeBySlug(rec.PartNumber); ok {
			slug = rt.Slug
			uHeight = rt.UHeight
			partNumber = rt.PartNumber
			manufacturer = rt.Manufacturer
			model = rt.Model
		} else if rt, ok := devicetypes.GetRackTypeByPartNumber(rec.PartNumber); ok {
			slug = rt.Slug
			uHeight = rt.UHeight
			partNumber = rt.PartNumber
			manufacturer = rt.Manufacturer
			model = rt.Model
		}

		for i := 0; i < rec.Qty; i++ {
			id := uuid.New()
			name := rec.Name
			if rec.Qty > 1 && name != "" {
				name = fmt.Sprintf("%s-%d", name, i+1)
			}

			rack := &devicetypes.CaniRackType{
				ID:           id,
				Name:         name,
				Slug:         slug,
				PartNumber:   partNumber,
				Manufacturer: manufacturer,
				Model:        model,
				UHeight:      uHeight,
				ObjectMeta:   devicetypes.ObjectMeta{Status: rec.Status},
				Devices:      []uuid.UUID{},
			}

			// Assign location
			if rec.Location != "" {
				locID, ok := locationsByName[rec.Location]
				if !ok {
					locID, ok = findLocationByName(inv, rec.Location)
				}
				if ok {
					rack.Location = locID
				}
			}

			result.Racks[id] = rack
			inv.Racks[id] = rack
			if name != "" {
				racksByName[name] = id
			}
		}
	}

	return racksByName, nil
}

// transformSystemDevices creates devices from system CSV device records.
func transformSystemDevices(data *import_.SystemCSV, result *devicetypes.TransformResult, inv *devicetypes.Inventory, racksByName map[string]uuid.UUID) error {
	for _, rec := range data.Devices {
		rec = data.ApplyDefaults(rec)
		if rec.PartNumber == "" {
			return fmt.Errorf("device %q missing PartNumber (slug or part number)", rec.Name)
		}

		for i := 0; i < rec.Qty; i++ {
			id := uuid.New()
			name := rec.Name
			if rec.Qty > 1 && name != "" {
				name = fmt.Sprintf("%s-%d", name, i+1)
			}

			device := &devicetypes.CaniDeviceType{
				ID:         id,
				Name:       name,
				Slug:       rec.PartNumber,
				PartNumber: rec.PartNumber,
				Serial:     rec.Serial,
				ObjectMeta: devicetypes.ObjectMeta{
					Status: rec.Status,
					Role:   rec.Role,
				},
			}

			// Populate from device type library
			if dt, ok := devicetypes.GetBySlug(rec.PartNumber); ok {
				populateFromDeviceType(device, &dt)
			} else if dt, ok := devicetypes.GetByPartNumber(rec.PartNumber); ok {
				populateFromDeviceType(device, &dt)
			}

			// Place in rack
			if rec.Rack != "" {
				rackID, ok := racksByName[rec.Rack]
				if !ok {
					// Try existing inventory
					rackID, ok = findRackByName(inv, rec.Rack)
				}
				if !ok {
					return fmt.Errorf("device %q references unknown rack %q", name, rec.Rack)
				}
				device.Parent = rackID
				device.Rack = rackID

				if rec.Position > 0 {
					device.RackPosition = rec.Position
					face := rec.Face
					if face == "" {
						face = devicetypes.FaceFront
					}
					device.Face = face

					if rack, ok := inv.Racks[rackID]; ok {
						height := device.UHeight
						if height < 1 {
							height = 1
						}
						if !rack.PlaceDevice(id, rec.Position, height, face, device.IsFullDepth) {
							log.Printf("WARN: device %q cannot fit at U%d in rack %q", name, rec.Position, rec.Rack)
						}
					}
				}
			}

			result.Devices[id] = device
			inv.Devices[id] = device
		}
	}

	return nil
}

// transformSystemModules creates modules from system CSV module records.
func transformSystemModules(data *import_.SystemCSV, result *devicetypes.TransformResult, inv *devicetypes.Inventory) error {
	for _, rec := range data.Modules {
		rec = data.ApplyDefaults(rec)

		for i := 0; i < rec.Qty; i++ {
			id := uuid.New()
			var parentDevice *devicetypes.CaniDeviceType

			module := &devicetypes.CaniModuleType{
				ID:         id,
				Name:       rec.Name,
				Slug:       rec.PartNumber,
				PartNumber: rec.PartNumber,
				Serial:     rec.Serial,
				ObjectMeta: devicetypes.ObjectMeta{Status: rec.Status},
			}

			// Populate from module type library
			if rec.PartNumber != "" {
				if mt, ok := devicetypes.GetModuleBySlug(rec.PartNumber); ok {
					populateModuleFromType(module, &mt)
				} else if mt, ok := devicetypes.GetModuleTypeByPartNumber(rec.PartNumber); ok {
					populateModuleFromType(module, &mt)
				}
			}

			// Parent to device
			if rec.Device != "" {
				dev := inv.FindDeviceByNameOrID(rec.Device)
				if dev == nil {
					return fmt.Errorf("module %q references unknown device %q", rec.Name, rec.Device)
				}
				parentDevice = dev
				module.ParentDevice = dev.ID
			}

			if rec.Bay != "" {
				module.ModuleBayName = resolveSystemModuleBayName(parentDevice, rec.Bay)
			}

			if module.Name == "" {
				module.Name = synthesizeSystemModuleName(module, parentDevice)
			}

			result.Modules[id] = module
		}
	}

	return nil
}

// populateModuleFromType copies identity and hardware specs from a module
// type library entry, including a fresh copy of the interface specs so that
// per-interface details (e.g. MAC addresses) can be set on the module.
func populateModuleFromType(module *devicetypes.CaniModuleType, mt *devicetypes.CaniModuleType) {
	module.Slug = mt.Slug
	module.Manufacturer = mt.Manufacturer
	module.Model = mt.Model
	module.Type = mt.Type
	if mt.PartNumber != "" {
		module.PartNumber = mt.PartNumber
	}
	module.Interfaces = append([]devicetypes.InterfaceSpec(nil), mt.Interfaces...)
}

func resolveSystemModuleBayName(device *devicetypes.CaniDeviceType, requestedBay string) string {
	if device == nil || requestedBay == "" {
		return requestedBay
	}

	for _, bay := range device.ModuleBays {
		if bay.Name == requestedBay || bay.Position == requestedBay {
			return bay.Name
		}
	}

	return requestedBay
}

func synthesizeSystemModuleName(module *devicetypes.CaniModuleType, device *devicetypes.CaniDeviceType) string {
	if module == nil || module.Name != "" {
		return ""
	}

	lowerSlug := strings.ToLower(module.Slug)
	if device != nil && module.ModuleBayName != "" && (module.Type == devicetypes.TypeGPU || strings.Contains(lowerSlug, "gpu")) {
		return fmt.Sprintf("gpu-%s-%s", device.Name, module.ModuleBayName)
	}

	if device != nil && strings.Contains(lowerSlug, "connectx-6") {
		return fmt.Sprintf("CX6-%s", device.Name)
	}

	return fallbackSystemModuleName(module, device)
}

// fallbackSystemModuleName builds a deterministic name from the module slug,
// parent device, and bay so a module is never left unnamed. Unnamed modules are
// dropped from datastore summaries and collide when a row sets Qty > 1.
func fallbackSystemModuleName(module *devicetypes.CaniModuleType, device *devicetypes.CaniDeviceType) string {
	base := module.Slug
	if base == "" {
		base = "module"
	}
	if device != nil {
		base = fmt.Sprintf("%s-%s", base, device.Name)
	}
	if module.ModuleBayName != "" {
		base = fmt.Sprintf("%s-%s", base, module.ModuleBayName)
	}
	return base
}

// transformSystemInterfaces applies per-interface metadata (currently MAC
// addresses) from the system CSV's `interface` rows onto the matching device
// or module interface specs. An interface row uses Device as the owner name
// and Name as the interface name. The owner is resolved first among objects
// created in this import and then among devices and modules already present in
// the inventory, so an interface row can annotate hardware imported in an
// earlier run. Rows referencing an unknown owner or interface are logged and
// skipped so a single bad row does not abort import.
func transformSystemInterfaces(data *import_.SystemCSV, result *devicetypes.TransformResult, inv *devicetypes.Inventory) error {
	for _, rec := range data.Interfaces {
		rec = data.ApplyDefaults(rec)

		if rec.Device == "" {
			log.Printf("WARN: system CSV interface row for %q has no Device, skipping", rec.Name)
			continue
		}
		if rec.Name == "" {
			log.Printf("WARN: system CSV interface row for device %q has no Name, skipping", rec.Device)
			continue
		}

		spec := findResultInterfaceSpec(result, rec.Device, rec.Name)
		if spec == nil {
			spec = findInventoryInterfaceSpec(inv, rec.Device, rec.Name)
		}
		if spec == nil {
			log.Printf("WARN: system CSV interface %q not found on device/module %q, skipping", rec.Name, rec.Device)
			continue
		}

		if rec.MacAddress != "" {
			mac, err := devicetypes.NormalizeMAC(rec.MacAddress)
			if err != nil {
				log.Printf("WARN: system CSV interface %q on %q: %v, skipping", rec.Name, rec.Device, err)
				continue
			}
			spec.MacAddress = mac
		}
	}
	return nil
}

// findResultInterfaceSpec locates an InterfaceSpec by interface name on the
// device or module with the given name within the transform result.
func findResultInterfaceSpec(result *devicetypes.TransformResult, ownerName, ifaceName string) *devicetypes.InterfaceSpec {
	if spec := findInterfaceSpecInDevices(result.Devices, ownerName, ifaceName); spec != nil {
		return spec
	}
	return findInterfaceSpecInModules(result.Modules, ownerName, ifaceName)
}

// findInventoryInterfaceSpec locates an InterfaceSpec by interface name on a
// device or module already present in the inventory, allowing an interface row
// to annotate hardware imported in an earlier run.
func findInventoryInterfaceSpec(inv *devicetypes.Inventory, ownerName, ifaceName string) *devicetypes.InterfaceSpec {
	if inv == nil {
		return nil
	}
	if spec := findInterfaceSpecInDevices(inv.Devices, ownerName, ifaceName); spec != nil {
		return spec
	}
	return findInterfaceSpecInModules(inv.Modules, ownerName, ifaceName)
}

// findInterfaceSpecInDevices returns the named interface spec on the device with
// the given name, or nil when no device or interface matches.
func findInterfaceSpecInDevices(devices map[uuid.UUID]*devicetypes.CaniDeviceType, ownerName, ifaceName string) *devicetypes.InterfaceSpec {
	for _, dev := range devices {
		if dev == nil || dev.Name != ownerName {
			continue
		}
		for i := range dev.Interfaces {
			if dev.Interfaces[i].Name == ifaceName {
				return &dev.Interfaces[i]
			}
		}
	}
	return nil
}

// findInterfaceSpecInModules returns the named interface spec on the module with
// the given name, or nil when no module or interface matches.
func findInterfaceSpecInModules(modules map[uuid.UUID]*devicetypes.CaniModuleType, ownerName, ifaceName string) *devicetypes.InterfaceSpec {
	for _, mod := range modules {
		if mod == nil || mod.Name != ownerName {
			continue
		}
		for i := range mod.Interfaces {
			if mod.Interfaces[i].Name == ifaceName {
				return &mod.Interfaces[i]
			}
		}
	}
	return nil
}

// transformSystemConnections resolves connection records into cables.
func transformSystemConnections(data *import_.SystemCSV, result *devicetypes.TransformResult, inv *devicetypes.Inventory) error {
	if len(data.Connections) == 0 {
		return nil
	}

	// Build a ConnectionMap from system CSV connection records
	cm := &connections.ConnectionMap{
		Version: "v1",
	}

	// Determine defaults from _defaults and connection_defaults
	globalDefaults := data.Defaults
	sectionDefaults, hasSectionDefaults := data.SectionDefaults["connection"]
	if globalDefaults.Status != "" || hasSectionDefaults {
		cd := &connections.CableDefaults{}
		if globalDefaults.Status != "" {
			cd.Status = globalDefaults.Status
		}
		if hasSectionDefaults {
			if sectionDefaults.Status != "" {
				cd.Status = sectionDefaults.Status
			}
			if sectionDefaults.Color != "" {
				cd.Color = sectionDefaults.Color
			}
			if sectionDefaults.LengthUnit != "" {
				cd.LengthUnit = sectionDefaults.LengthUnit
			}
		}
		cm.CableDefaults = cd
	}

	for _, rec := range data.Connections {
		rec = data.ApplyDefaults(rec)

		entry := connections.ConnectionEntry{
			A: connections.Endpoint{Device: rec.ADevice, Port: rec.APort},
			B: connections.Endpoint{Device: rec.BDevice, Port: rec.BPort},
		}

		cable := &connections.CableProps{}
		hasCableProps := false

		if rec.PartNumber != "" {
			cable.Type = rec.PartNumber
			hasCableProps = true
		}
		if rec.Color != "" {
			cable.Color = rec.Color
			hasCableProps = true
		}
		if rec.Length != "" {
			if l, err := strconv.ParseFloat(rec.Length, 64); err == nil {
				cable.Length = &l
				hasCableProps = true
			}
		}
		if rec.LengthUnit != "" {
			cable.LengthUnit = rec.LengthUnit
			hasCableProps = true
		}
		if rec.Status != "" {
			cable.Status = rec.Status
			hasCableProps = true
		}

		if hasCableProps {
			entry.Cable = cable
		}

		cm.Connections = append(cm.Connections, entry)
	}

	// Resolve patterns and device names
	resolved, errs := connections.ResolveConnectionMap(cm, inv)
	if len(errs) > 0 {
		for _, e := range errs {
			log.Printf("WARN: connection resolution: %v", e)
		}
		if len(resolved) == 0 {
			return fmt.Errorf("no connections resolved; %d errors", len(errs))
		}
	}

	// Create cable objects
	for _, conn := range resolved {
		cable := devicetypes.NewCable(conn.Cable.Type, conn.Cable.Label)
		cable.TerminationADevice = conn.ADevice
		cable.TerminationAPort = conn.APort
		cable.TerminationBDevice = conn.BDevice
		cable.TerminationBPort = conn.BPort

		if conn.Cable.Color != "" {
			cable.Color = conn.Cable.Color
		}
		if conn.Cable.Length != nil {
			cable.Length = conn.Cable.Length
		}
		if conn.Cable.LengthUnit != "" {
			cable.LengthUnit = conn.Cable.LengthUnit
		}
		if conn.Cable.Status != "" {
			cable.Status = conn.Cable.Status
		}

		result.Cables[cable.ID] = cable
	}

	return nil
}

// findRackByName searches the inventory for a rack with the given name.
func findRackByName(inv *devicetypes.Inventory, name string) (uuid.UUID, bool) {
	for id, rack := range inv.Racks {
		if rack.Name == name {
			return id, true
		}
	}
	return uuid.Nil, false
}
