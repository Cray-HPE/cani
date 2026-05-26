package transform

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/devicetypes/connections"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/google/uuid"
)

// TransformSystem converts parsed system CSV data into a TransformResult.
// Uses a 5-pass algorithm: roles → racks → devices → modules → connections.
func TransformSystem(existing devicetypes.Inventory, data *import_.SystemCSV) (*devicetypes.TransformResult, error) {
	initInventoryMaps(&existing)

	result := &devicetypes.TransformResult{
		Locations: make(map[uuid.UUID]*devicetypes.CaniLocationType),
		Racks:     make(map[uuid.UUID]*devicetypes.CaniRackType),
		Devices:   make(map[uuid.UUID]*devicetypes.CaniDeviceType),
		Modules:   make(map[uuid.UUID]*devicetypes.CaniModuleType),
		Cables:    make(map[uuid.UUID]*devicetypes.CaniCableType),
	}

	// Pass 0: Roles
	meta, err := transformRoles(data)
	if err != nil {
		return nil, fmt.Errorf("transformRoles: %w", err)
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

	// Pass 4: Connections
	err = transformSystemConnections(data, result, &existing)
	if err != nil {
		return nil, fmt.Errorf("transformSystemConnections: %w", err)
	}

	log.Printf("System CSV transformed: %d roles, %d locations, %d racks, %d devices, %d modules, %d cables",
		len(data.Roles), len(result.Locations), len(result.Racks), len(result.Devices), len(result.Modules), len(result.Cables))

	return result, nil
}

// transformRoles creates MetadataEntry items from role records.
func transformRoles(data *import_.SystemCSV) (*devicetypes.InventoryMetadata, error) {
	if len(data.Roles) == 0 {
		return nil, nil
	}
	meta := &devicetypes.InventoryMetadata{}
	for _, rec := range data.Roles {
		rec = data.ApplyDefaults(rec)
		if rec.Name == "" {
			return nil, fmt.Errorf("role record missing Name")
		}
		var contentTypes []string
		if rec.ContentTypes != "" {
			for _, ct := range strings.Split(rec.ContentTypes, ",") {
				ct = strings.TrimSpace(ct)
				if ct != "" {
					contentTypes = append(contentTypes, ct)
				}
			}
		}
		meta.Roles = append(meta.Roles, devicetypes.MetadataEntry{
			Name:         rec.Name,
			ContentTypes: contentTypes,
		})
	}
	return meta, nil
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
		if rec.Role == "" {
			return nil, fmt.Errorf("location %q missing Role (location type: dc, level, section)", rec.Name)
		}

		id := uuid.New()
		var contentTypes []string
		if rec.ContentTypes != "" {
			for _, ct := range strings.Split(rec.ContentTypes, ",") {
				ct = strings.TrimSpace(ct)
				if ct != "" {
					contentTypes = append(contentTypes, ct)
				}
			}
		}

		loc := &devicetypes.CaniLocationType{
			ID:           id,
			Name:         rec.Name,
			LocationType: rec.Role,
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
					module.Slug = mt.Slug
					module.Manufacturer = mt.Manufacturer
					module.Model = mt.Model
					module.Type = mt.Type
					if mt.PartNumber != "" {
						module.PartNumber = mt.PartNumber
					}
				} else if mt, ok := devicetypes.GetModuleTypeByPartNumber(rec.PartNumber); ok {
					module.Slug = mt.Slug
					module.Manufacturer = mt.Manufacturer
					module.Model = mt.Model
					module.Type = mt.Type
					if mt.PartNumber != "" {
						module.PartNumber = mt.PartNumber
					}
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
	if module == nil || module.Name != "" || device == nil {
		return ""
	}

	lowerSlug := strings.ToLower(module.Slug)
	if module.Type == devicetypes.TypeGPU || strings.Contains(lowerSlug, "gpu") {
		if module.ModuleBayName == "" {
			return ""
		}
		return fmt.Sprintf("gpu-%s-%s", device.Name, module.ModuleBayName)
	}

	if strings.Contains(lowerSlug, "connectx-6") {
		return fmt.Sprintf("CX6-%s", device.Name)
	}

	return ""
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
