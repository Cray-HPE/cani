package devicetypes

func All() (devices map[string]DeviceType) {
	return allDeviceTypes
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

func (inv Inventory) Systems() (systems []CaniDeviceType) {
	return inv.DevicesByType("system")
}

// DevicesByType returns all devices of a specific type from the inventory.
func (inv Inventory) DevicesByType(deviceType string) (devices []CaniDeviceType) {
	for _, device := range inv.Devices {
		if device != nil && device.Type == Type(deviceType) {
			devices = append(devices, *device)
		}
	}
	return devices
}

// Exists checks if a device with the given name exists in the inventory.
func (inv Inventory) Exists(name string) bool {
	for _, device := range inv.Devices {
		if device != nil && device.Name == name {
			return true
		}
	}
	return false
}

// FindName finds a device by its name in the inventory.
func (inv Inventory) FindName(name string) (device *CaniDeviceType, exists bool) {
	for _, device := range inv.Devices {
		if device != nil && device.Name == name {
			return device, true
		}
	}
	return nil, false
}
