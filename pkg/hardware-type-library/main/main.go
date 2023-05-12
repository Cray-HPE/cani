package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	hardware_type_library "github.com/Cray-HPE/cani/pkg/hardware-type-library"
)

// type HardwarePath struct {
// 	DeviceBay string

// }

type HardwareBuildOut struct {
	// HardwareType hardware_type_library.HardwareType
	DeviceTypeString string
	DeviceType       hardware_type_library.DeviceType
	Path             []string
	Ordinal          int
	OrdinalPath      []int
	HardwareTypePath []hardware_type_library.HardwareType
}

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

	// results := []HardwareBuildOut{}

	// deviceTypes := []string{"hpe-crayex-ex420-compute-blade"}
	// for len(deviceTypes) != 0 {
	// 	currentDeviceType := deviceTypes[0]
	// 	deviceTypes = deviceTypes[1:]

	// 	fmt.Println(currentDeviceType)

	// 	deviceType, ok := library.DeviceTypes[currentDeviceType]
	// 	if !ok {
	// 		panic(fmt.Sprint("Device type does not exist", currentDeviceType))
	// 	}

	// 	for _, deviceBay := range deviceType.DeviceBays {
	// 		fmt.Println("  Device bay:", deviceBay.Name)
	// 		if deviceBay.Default != nil {
	// 			fmt.Println("    Default:", deviceBay.Default.Slug)
	// 			deviceTypes = append(deviceTypes, deviceBay.Default.Slug)
	// 		}
	// 	}
	// }

	results := []HardwareBuildOut{}
	queue := []HardwareBuildOut{
		{
			DeviceTypeString: "hpe-crayex-ex420-compute-blade",
			Path:             []string{}, // This is the root of the path
			Ordinal:          -1,
		},
		// {
		// 	DeviceTypeString: "hpe-ex4000",
		// 	Path:             []string{}, // This is the root of the path
		// 	Ordinal:          -1,
		// },
	}
	for len(queue) != 0 {
		current := queue[0]
		queue = queue[1:]

		fmt.Println("Visiting: ", current.DeviceTypeString)
		currentDeviceType, ok := library.DeviceTypes[current.DeviceTypeString]
		if !ok {
			panic(fmt.Sprint("Device type does not exist", current.DeviceType))
		}

		// Retrieve the hardware type at this point in time, so we only lookup in the map once
		current.DeviceType = currentDeviceType
		current.HardwareTypePath = append(current.HardwareTypePath, current.DeviceType.HardwareType)

		for _, deviceBay := range currentDeviceType.DeviceBays {
			fmt.Println("  Device bay:", deviceBay.Name)
			if deviceBay.Default != nil {
				fmt.Println("    Default:", deviceBay.Default.Slug)

				// Extract the ordinal
				// This is one way of going about, but it assumes that each name has a number
				// There are two other ways to consider:
				// - Embed an actual ordinal number in the yaml files
				// - Get all of the device base with that type, and then sort them lexicographically. This is how HSM does it, but assumes the names can be sorted in a predictable order
				r := regexp.MustCompile(`\d+`)
				match := r.FindString(deviceBay.Name)
				fmt.Printf("%s|%s\n", deviceBay.Name, match)

				var ordinal int
				if match != "" {
					ordinal, err = strconv.Atoi(match)
					if err != nil {
						panic(err)
					}
				}

				queue = append(queue, HardwareBuildOut{
					// Hardware type is defered until when it is processed
					DeviceTypeString: deviceBay.Default.Slug,
					Path:             append(current.Path, deviceBay.Name),
					Ordinal:          ordinal,
					OrdinalPath:      append(current.OrdinalPath, ordinal),
					HardwareTypePath: current.HardwareTypePath,
				})
			}
		}

		results = append(results, current)
	}

	seperator := fmt.Sprintf("|%s|%s|%s|%s|%s|%s|%s|", strings.Repeat("-", 40+1), strings.Repeat("-", 20+1), strings.Repeat("-", 60+1), strings.Repeat("-", 40+1), strings.Repeat("-", 40+1), strings.Repeat("-", 40+1), strings.Repeat("-", 40+1))
	fmt.Println(seperator)
	fmt.Printf("| %-40s| %-20s| %-60s| %-40s| %-40s| %-40s| %-40s|\n", "Hardware Type", "Manufacturer", "Model", "Ordinal", "Path", "Ordinal Path", "Hardware Type Path")
	fmt.Println(seperator)
	for _, result := range results {
		fmt.Printf("| %-40s| %-20s| %-60s| %-40d| %-40s| %-40s| %-40s|\n", result.DeviceType.HardwareType, result.DeviceType.Manufacturer, result.DeviceType.Model, result.Ordinal, strings.Join(result.Path, "->"), joinInts(result.OrdinalPath, "->"), joinHardwareTypes(result.HardwareTypePath, "->"))
	}
	fmt.Println(seperator)

	// Lets now say we know the cabinet, chassis, and slot that these devices are going into
	// cabinet := 1001
	// chassis := 1
	// slot := 7

	// mySlot := hardware_type_library.Ch

}
