package export

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// Export prints the inventory in a hierarchical visual format.
func Export(inv devicetypes.Inventory) error {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    CANI Inventory Export                     ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Print locations hierarchy
	if len(inv.Locations) > 0 {
		for _, loc := range inv.Locations {
			printLocation(loc, &inv, 0)
		}
	} else if len(inv.Racks) > 0 {
		// No locations, print racks directly
		for _, rack := range inv.Racks {
			printRack(rack, &inv, 0)
		}
	} else if len(inv.Devices) > 0 {
		// No locations or racks, print devices directly
		fmt.Println("Devices:")
		for _, device := range inv.Devices {
			printDevice(device, &inv, 1)
		}
	}

	// Summary
	fmt.Println()
	fmt.Println("────────────────────────────────────────────────────────────────")
	fmt.Printf("Summary: %d locations, %d racks, %d devices, %d modules, %d cables\n",
		len(inv.Locations), len(inv.Racks), len(inv.Devices), len(inv.Modules), len(inv.Cables))

	return nil
}

// printLocation prints a location and its children.
func printLocation(loc *devicetypes.CaniLocationType, inv *devicetypes.Inventory, indent int) {
	if loc == nil {
		return
	}
	prefix := strings.Repeat("  ", indent)

	fmt.Printf("%s📍 %s (%s)\n", prefix, loc.Name, loc.LocationType)

	// Print child locations
	for _, childID := range loc.Children {
		if child, ok := inv.Locations[childID]; ok {
			printLocation(child, inv, indent+1)
		}
	}

	// Print racks at this location
	for _, rackID := range loc.Racks {
		if rack, ok := inv.Racks[rackID]; ok {
			printRack(rack, inv, indent+1)
		}
	}
}

// printRack prints a rack and its devices sorted by U-position.
func printRack(rack *devicetypes.CaniRackType, inv *devicetypes.Inventory, indent int) {
	if rack == nil {
		return
	}
	prefix := strings.Repeat("  ", indent)

	fmt.Printf("%s🗄️  %s [%dU]\n", prefix, rack.Name, rack.UHeight)

	// Build a map of devices in this rack by their starting U-position
	type devicePos struct {
		device *devicetypes.CaniDeviceType
		startU int
	}
	var devicesInRack []devicePos

	for _, deviceID := range rack.Devices {
		if device, ok := inv.Devices[deviceID]; ok {
			startU := rack.GetDeviceStartU(deviceID)
			if startU == 0 {
				startU = device.RackPosition
			}
			devicesInRack = append(devicesInRack, devicePos{device: device, startU: startU})
		}
	}

	// Sort by U-position descending (top to bottom)
	sort.Slice(devicesInRack, func(i, j int) bool {
		return devicesInRack[i].startU > devicesInRack[j].startU
	})

	// Print devices in rack order
	fmt.Printf("%s  ┌─────────────────────────────────────────────────────────┐\n", prefix)
	for _, dp := range devicesInRack {
		uHeight := dp.device.GetUHeight()
		endU := dp.startU + uHeight - 1
		uRange := fmt.Sprintf("U%02d-U%02d", dp.startU, endU)
		if uHeight == 1 {
			uRange = fmt.Sprintf("U%02d    ", dp.startU)
		}
		hwType := string(dp.device.Type)
		if hwType == "" {
			hwType = "device"
		}
		fmt.Printf("%s  │ %s │ %-12s │ %-20s │\n", prefix, uRange, dp.device.Name, hwType)
	}
	fmt.Printf("%s  └─────────────────────────────────────────────────────────┘\n", prefix)

	// Print cables for this rack
	printRackCables(rack, inv, indent+1)
}

// printRackCables prints cable connections for devices in this rack.
func printRackCables(rack *devicetypes.CaniRackType, inv *devicetypes.Inventory, indent int) {
	if len(inv.Cables) == 0 {
		return
	}

	prefix := strings.Repeat("  ", indent)

	// Build set of device IDs in this rack
	rackDevices := make(map[uuid.UUID]bool)
	for _, deviceID := range rack.Devices {
		rackDevices[deviceID] = true
	}

	// Find cables where at least one endpoint is in this rack
	var rackCables []*devicetypes.CaniCableType
	for _, cable := range inv.Cables {
		_, deviceA := inv.GetInterfaceByID(cable.TerminationA)
		_, deviceB := inv.GetInterfaceByID(cable.TerminationB)
		deviceAInRack := deviceA != nil && rackDevices[deviceA.ID]
		deviceBInRack := deviceB != nil && rackDevices[deviceB.ID]
		if deviceAInRack || deviceBInRack {
			rackCables = append(rackCables, cable)
		}
	}

	if len(rackCables) == 0 {
		return
	}

	fmt.Printf("%sCables:\n", prefix)
	for _, cable := range rackCables {
		ifaceA, deviceA := inv.GetInterfaceByID(cable.TerminationA)
		ifaceB, deviceB := inv.GetInterfaceByID(cable.TerminationB)
		deviceAName := "unknown"
		deviceBName := "unknown"
		portAName := "?"
		portBName := "?"
		if deviceA != nil {
			deviceAName = deviceA.Name
		}
		if deviceB != nil {
			deviceBName = deviceB.Name
		}
		if ifaceA != nil {
			portAName = ifaceA.Name
		}
		if ifaceB != nil {
			portBName = ifaceB.Name
		}
		fmt.Printf("%s  ⚡ [%s] %s:%s ══ %s:%s\n",
			prefix,
			cable.Slug,
			deviceAName, portAName,
			deviceBName, portBName)
	}
}

// printDevice prints a device and its children/modules.
func printDevice(device *devicetypes.CaniDeviceType, inv *devicetypes.Inventory, indent int) {
	if device == nil {
		return
	}
	prefix := strings.Repeat("  ", indent)

	hwType := string(device.Type)
	if hwType == "" {
		hwType = "device"
	}
	fmt.Printf("%s🖥️  %s (%s) - %s\n", prefix, device.Name, hwType, device.Model)

	// Print child devices
	for _, childID := range device.Children {
		if child, ok := inv.Devices[childID]; ok {
			printDevice(child, inv, indent+1)
		}
	}

	// Print modules in this device
	for _, module := range inv.Modules {
		if module != nil && module.ParentDevice == device.ID {
			printModule(module, inv, indent+1)
		}
	}
}

// printModule prints a module.
func printModule(module *devicetypes.CaniModuleType, inv *devicetypes.Inventory, indent int) {
	if module == nil {
		return
	}
	prefix := strings.Repeat("  ", indent)
	fmt.Printf("%s📦 %s [%s] - %s\n", prefix, module.Name, module.ModuleBayName, module.Slug)
}

// getDeviceName looks up a device name by ID.
func getDeviceName(inv *devicetypes.Inventory, deviceID uuid.UUID) string {
	if device, ok := inv.Devices[deviceID]; ok {
		return device.Name
	}
	return deviceID.String()[:8]
}
