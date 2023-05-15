package main

import (
	"fmt"
	"strings"

	hardware_type_library "github.com/Cray-HPE/cani/pkg/hardware-type-library"
	"github.com/Cray-HPE/hms-xname/xnames"
)

// type HardwarePath struct {
// 	DeviceBay string

// }

func joinInts(ints []int, sep string) string {
	intsStr := []string{}
	for _, value := range ints {
		intsStr = append(intsStr, fmt.Sprint(value))
	}

	return strings.Join(intsStr, sep)
}

func joinHardwareTypes(in []hardware_type_library.HardwareType, sep string) string {
	out := []string{}
	for _, value := range in {
		out = append(out, string(value))
	}

	return strings.Join(out, sep)
}

func buildChassisPath(cabinet, chassis int) hardware_type_library.HardwareBuildOut {
	return hardware_type_library.HardwareBuildOut{
		OrdinalPath: []int{cabinet, chassis},
		HardwareTypePath: []hardware_type_library.HardwareType{
			hardware_type_library.HardwareTypeCabinet,
			hardware_type_library.HardwareTypeChassis,
		},
	}
}

func buildXname(hardwareTypePath []hardware_type_library.HardwareType, locationPath []int) xnames.Xname {
	type typeConverter struct {
		hardwareTypePath []hardware_type_library.HardwareType
		convert          func() xnames.Xname
	}

	// TODO this could probably be auto generated, assuming a type mapping table exists
	// typeConverters := []typeConverter{
	// 	{
	// 		// Slot/Node Blade
	// 		hardwareTypePath: []hardware_type_library.HardwareType{
	// 			hardware_type_library.HardwareTypeCabinet,
	// 			hardware_type_library.HardwareTypeChassis,
	// 			hardware_type_library.HardwareTypeNodeBlade,
	// 		},

	// 		convert: func() xnames.Xname {
	// 			return xnames.ComputeModule{
	// 				Cabinet:       locationPath[0],
	// 				Chassis:       locationPath[1],
	// 				ComputeModule: locationPath[2],
	// 			}
	// 		},
	// 	},
	// }

	// TODO work in progress
	return nil

}

func main() {
	// Create the library
	library, err := hardware_type_library.NewEmbeddedLibrary()
	if err != nil {
		panic(err)
	}

	// List cabinets
	fmt.Println()
	fmt.Println("Cabinets")
	cabinetDeviceTypes := library.GetDeviceTypesByHardwareType(hardware_type_library.HardwareTypeCabinet)
	for _, cabinetDeviceType := range cabinetDeviceTypes {
		fmt.Println(cabinetDeviceType.Slug)
	}

	fmt.Println()
	fmt.Println("Node Blade")
	nodeBladeDeviceTypes := library.GetDeviceTypesByHardwareType(hardware_type_library.HardwareTypeNodeBlade)
	for _, nodeBladeDeviceType := range nodeBladeDeviceTypes {
		fmt.Println(nodeBladeDeviceType.Slug)
	}

	allChildHardware, err := library.GetDefaultChildHardwareBuildOut("hpe-crayex-ex420-compute-blade")
	if err != nil {
		panic(err)
	}

	seperator := fmt.Sprintf("|%s|%s|%s|%s|%s|%s|%s|", strings.Repeat("-", 40+1), strings.Repeat("-", 20+1), strings.Repeat("-", 60+1), strings.Repeat("-", 40+1), strings.Repeat("-", 40+1), strings.Repeat("-", 40+1), strings.Repeat("-", 40+1))
	fmt.Println(seperator)
	fmt.Printf("| %-40s| %-20s| %-60s| %-40s| %-40s| %-40s| %-40s|\n", "Hardware Type", "Manufacturer", "Model", "Ordinal", "Path", "Ordinal Path", "Hardware Type Path")
	fmt.Println(seperator)
	for _, result := range allChildHardware {
		fmt.Printf("| %-40s| %-20s| %-60s| %-40d| %-40s| %-40s| %-40s|\n", result.DeviceType.HardwareType, result.DeviceType.Manufacturer, result.DeviceType.Model, result.Ordinal, strings.Join(result.Path, "->"), joinInts(result.OrdinalPath, "->"), joinHardwareTypes(result.HardwareTypePath, "->"))
	}
	fmt.Println(seperator)

	// Lets now say we know the cabinet, chassis, and slot that these devices are going into
	// cabinet := 1001
	// chassis := 1
	// slot := 7

	// TODO Interact with the inventory
	// - Check to see if the cabinet exists
	// - Check to see if the chassis exists
	// - Check to see if the slot is empty
	//	- TODO also should check to see if the slot is within bounds
	// TODO Ask the inventory to see if node blade exists at cabinet: 1001, chassis: 1, slot: 7

	// TODO Ask the inventory for the paths to the chassis in the tree
	// locationPath := []int{1001, 1, 7}
	// deviceTypePath := []hardware_type_library.HardwareType{
	// 	hardware_type_library.HardwareTypeCabinet,
	// 	hardware_type_library.HardwareTypeChassis,
	// 	hardware_type_library.HardwareTypeNodeBlade,
	// }

	// for _, childHardware := range allChildHardware {
	// 	// Get full hardware type path
	// 	childHardwareTypePath := append(deviceTypePath, childHardware.HardwareTypePath...)

	// 	// Get full location/ordinal path
	// 	childLocationPath := append(locationPath, childHardware.OrdinalPath...)
	// }
}
