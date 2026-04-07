package devicetypes

import (
	"strings"
)

func All() (devices map[string]CaniDeviceType) {
	return allDeviceTypes
}

// ByPartNumber returns the map of device types indexed by part number
func ByPartNumber() map[string]CaniDeviceType {
	return deviceTypesByPartNum
}

// GetByPartNumber looks up a device type by its part number.
// Returns the device type and true if found, or an empty CaniDeviceType and false if not found.
func GetByPartNumber(partNumber string) (CaniDeviceType, bool) {
	dt, ok := deviceTypesByPartNum[partNumber]
	return dt, ok
}

// GetBySlug looks up a device type by its slug.
// Returns the device type and true if found, or an empty CaniDeviceType and false if not found.
func GetBySlug(slug string) (CaniDeviceType, bool) {
	dt, ok := allDeviceTypes[slug]
	return dt, ok
}

// AllCables returns all loaded cable types.
func AllCables() map[string]CaniCableType {
	return allCableTypes
}

// GetCableTypeByPartNumber looks up a cable type by its part number.
func GetCableTypeByPartNumber(partNumber string) (CaniCableType, bool) {
	ct, ok := cableTypesByPartNum[partNumber]
	return ct, ok
}

// GetCableTypeBySlug looks up a cable type by its slug.
func GetCableTypeBySlug(slug string) (CaniCableType, bool) {
	ct, ok := allCableTypes[slug]
	return ct, ok
}

// AllRackTypes returns all loaded rack types.
func AllRackTypes() map[string]CaniRackType {
	return allRackTypes
}

// GetRackTypeBySlug looks up a rack type by its slug.
func GetRackTypeBySlug(slug string) (CaniRackType, bool) {
	rt, ok := allRackTypes[slug]
	return rt, ok
}

// GetRackTypeByPartNumber looks up a rack type by its part number.
func GetRackTypeByPartNumber(partNumber string) (CaniRackType, bool) {
	rt, ok := rackTypesByPartNum[partNumber]
	return rt, ok
}

// RegisterRackType adds a rack type to the in-memory registry.
func RegisterRackType(rt CaniRackType) {
	allRackTypes[rt.Slug] = rt
	if rt.PartNumber != "" {
		rackTypesByPartNum[rt.PartNumber] = rt
	}
}

// AllFruTypes returns all loaded FRU/inventory-item types.
func AllFruTypes() map[string]CaniFruType {
	return allFruTypes
}

// GetFruTypeBySlug looks up a FRU type by its slug.
func GetFruTypeBySlug(slug string) (CaniFruType, bool) {
	ft, ok := allFruTypes[slug]
	return ft, ok
}

// GetFruTypeByPartNumber looks up a FRU type by its part number.
func GetFruTypeByPartNumber(partNumber string) (CaniFruType, bool) {
	ft, ok := fruTypesByPartNum[partNumber]
	return ft, ok
}

// RegisterFruType adds a FRU type to the in-memory registry.
func RegisterFruType(ft CaniFruType) {
	allFruTypes[ft.Slug] = ft
	if ft.PartNumber != "" {
		fruTypesByPartNum[ft.PartNumber] = ft
	}
}

// RegisterCableType adds a cable type to the in-memory registry.
func RegisterCableType(ct CaniCableType) {
	allCableTypes[ct.Slug] = ct
	if ct.PartNumber != "" {
		cableTypesByPartNum[ct.PartNumber] = ct
	}
}

// AllLocationTypes returns all loaded location type definitions.
func AllLocationTypes() map[string]LocationTypeDefinition {
	return allLocationTypes
}

// GetLocationTypeBySlug looks up a location type definition by its slug.
func GetLocationTypeBySlug(slug string) (LocationTypeDefinition, bool) {
	lt, ok := allLocationTypes[slug]
	return lt, ok
}

// RegisterLocationType adds a location type definition to the in-memory registry.
func RegisterLocationType(lt LocationTypeDefinition) {
	allLocationTypes[lt.Slug] = lt
}

// RegisterDeviceType adds a device type to the in-memory registry.
func RegisterDeviceType(dt CaniDeviceType) {
	allDeviceTypes[dt.Slug] = dt
	if dt.PartNumber != "" {
		deviceTypesByPartNum[dt.PartNumber] = dt
	}
}

// GetByManufacturerModel looks up a device type by manufacturer and model.
// The comparison is case-insensitive and also checks the Identifications array
// for alternate manufacturer/model combinations.
// If model is empty, it will match device types with the same manufacturer and an empty model.
// Returns the device type and true if found, or an empty CaniDeviceType and false if not found.
func GetByManufacturerModel(manufacturer, model string) (CaniDeviceType, bool) {
	manufacturerLower := strings.ToLower(manufacturer)
	modelLower := strings.ToLower(model)

	for _, dt := range allDeviceTypes {
		// Check primary manufacturer and model
		dtManufacturerLower := strings.ToLower(dt.Manufacturer)
		dtModelLower := strings.ToLower(dt.Model)

		// Match manufacturer and model (both must match, including empty model)
		if dtManufacturerLower == manufacturerLower && dtModelLower == modelLower {
			return dt, true
		}

		// Also check Identifications array for alternate manufacturer/model combinations
		for _, id := range dt.Identifications {
			if strings.ToLower(id.Manufacturer) == manufacturerLower &&
				strings.ToLower(id.Model) == modelLower {
				return dt, true
			}
		}
	}
	return CaniDeviceType{}, false
}

// AllModules returns all loaded module types
func AllModules() map[string]CaniModuleType {
	return allModuleTypes
}

// GetModuleBySlug looks up a module type by its slug.
// Returns the module type and true if found, or an empty CaniModuleType and false if not found.
func GetModuleBySlug(slug string) (CaniModuleType, bool) {
	mt, ok := allModuleTypes[slug]
	return mt, ok
}

// GetModuleTypeBySlug is an alias for GetModuleBySlug for consistency with device type naming.
// Returns the module type and true if found, or an empty CaniModuleType and false if not found.
func GetModuleTypeBySlug(slug string) (CaniModuleType, bool) {
	return GetModuleBySlug(slug)
}

// GetModuleTypeByPartNumber looks up a module type by its part number.
// Returns the module type and true if found, or an empty CaniModuleType and false if not found.
func GetModuleTypeByPartNumber(partNumber string) (CaniModuleType, bool) {
	mt, ok := moduleTypesByPartNum[partNumber]
	return mt, ok
}

// GetModuleByManufacturerModel looks up a module type by manufacturer and model.
// The comparison is case-insensitive.
// Returns the module type and true if found, or an empty CaniModuleType and false if not found.
func GetModuleByManufacturerModel(manufacturer, model string) (CaniModuleType, bool) {
	manufacturerLower := strings.ToLower(manufacturer)
	modelLower := strings.ToLower(model)

	for _, mt := range allModuleTypes {
		// Check primary manufacturer and model
		mtManufacturerLower := strings.ToLower(mt.Manufacturer)
		mtModelLower := strings.ToLower(mt.Model)

		// Match manufacturer and model (both must match, including empty model)
		if mtManufacturerLower == manufacturerLower && mtModelLower == modelLower {
			return mt, true
		}
	}
	return CaniModuleType{}, false
}

// RegisterModuleType adds a module type to the in-memory registry
func RegisterModuleType(mt CaniModuleType) {
	allModuleTypes[mt.Slug] = mt
	if mt.PartNumber != "" {
		moduleTypesByPartNum[mt.PartNumber] = mt
	}
}

func AllTypes() []Type {
	result := make([]Type, 0, len(allTypes))
	result = append(result, allTypes...)
	return result
}

func AllTypesString() []string {
	result := make([]string, 0, len(allTypes))
	for _, t := range allTypes {
		result = append(result, string(t))
	}
	return result
}

// DevicesByType returns all devices of a specific type from the inventory.
func (inv *Inventory) DevicesByType(deviceType string) []CaniDeviceType {
	if inv == nil {
		return nil
	}
	var devices []CaniDeviceType
	for _, device := range inv.Devices {
		if device != nil && device.Type == Type(deviceType) {
			devices = append(devices, *device)
		}
	}
	return devices
}

// Exists checks if a device with the given name exists in the inventory.
func (inv *Inventory) Exists(name string) bool {
	if inv == nil {
		return false
	}
	for _, device := range inv.Devices {
		if device != nil && device.Name == name {
			return true
		}
	}
	return false
}

// FindName finds a device by its name in the inventory.
func (inv *Inventory) FindName(name string) (*CaniDeviceType, bool) {
	if inv == nil {
		return nil, false
	}
	for _, device := range inv.Devices {
		if device != nil && device.Name == name {
			return device, true
		}
	}
	return nil, false
}
