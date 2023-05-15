package main

import (
	"fmt"
	"reflect"
	"strings"

	hardware_type_library "github.com/Cray-HPE/cani/pkg/hardware-type-library"
	"github.com/Cray-HPE/hms-xname/xnames"
)

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

// TODO this is something that should exist in the SLS provider
func buildXname(hardwareTypePath []hardware_type_library.HardwareType, locationPath []int) xnames.Xname {
	// TODO check that the length of hardware typePath and location path are the same
	fmt.Println(hardwareTypePath, locationPath)
	type typeConverter struct {
		hardwareTypePath []hardware_type_library.HardwareType
		convert          func() xnames.Xname
	}

	// TODO this could probably be auto generated, assuming a type mapping table exists
	typeConverters := []typeConverter{
		{
			// Cabinet
			hardwareTypePath: []hardware_type_library.HardwareType{
				hardware_type_library.HardwareTypeCabinet,
			},

			convert: func() xnames.Xname {
				return xnames.Cabinet{
					Cabinet: locationPath[0],
				}
			},
		},
		{
			// CEC
			hardwareTypePath: []hardware_type_library.HardwareType{
				hardware_type_library.HardwareTypeCabinet,
				hardware_type_library.HardwareTypeCabinetEnvironmentalController,
			},

			convert: func() xnames.Xname {
				return xnames.CEC{
					Cabinet: locationPath[0],
					CEC:     locationPath[1],
				}
			},
		},

		{
			// Chassis
			hardwareTypePath: []hardware_type_library.HardwareType{
				hardware_type_library.HardwareTypeCabinet,
				hardware_type_library.HardwareTypeChassis,
			},

			convert: func() xnames.Xname {
				return xnames.ComputeModule{
					Cabinet: locationPath[0],
					Chassis: locationPath[1],
				}
			},
		},
		{
			// Chassis BMC
			hardwareTypePath: []hardware_type_library.HardwareType{
				hardware_type_library.HardwareTypeCabinet,
				hardware_type_library.HardwareTypeChassis,
				hardware_type_library.HardwareTypeChassisManagementModule,
			},

			convert: func() xnames.Xname {
				return xnames.ChassisBMC{
					Cabinet:    locationPath[0],
					Chassis:    locationPath[1],
					ChassisBMC: locationPath[2],
				}
			},
		},

		{
			// Slot/Node Blade
			hardwareTypePath: []hardware_type_library.HardwareType{
				hardware_type_library.HardwareTypeCabinet,
				hardware_type_library.HardwareTypeChassis,
				hardware_type_library.HardwareTypeNodeBlade,
			},

			convert: func() xnames.Xname {
				return xnames.ComputeModule{
					Cabinet:       locationPath[0],
					Chassis:       locationPath[1],
					ComputeModule: locationPath[2],
				}
			},
		},

		{
			// NodeBMC
			hardwareTypePath: []hardware_type_library.HardwareType{
				hardware_type_library.HardwareTypeCabinet,
				hardware_type_library.HardwareTypeChassis,
				hardware_type_library.HardwareTypeNodeBlade,
				hardware_type_library.HardwareTypeNodeCard,
			},

			convert: func() xnames.Xname {
				return xnames.NodeBMC{
					Cabinet:       locationPath[0],
					Chassis:       locationPath[1],
					ComputeModule: locationPath[2],
					NodeBMC:       locationPath[3],
				}
			},
		},
		{
			// Node
			hardwareTypePath: []hardware_type_library.HardwareType{
				hardware_type_library.HardwareTypeCabinet,
				hardware_type_library.HardwareTypeChassis,
				hardware_type_library.HardwareTypeNodeBlade,
				hardware_type_library.HardwareTypeNodeCard,
				hardware_type_library.HardwareTypeNode,
			},

			convert: func() xnames.Xname {
				return xnames.Node{
					Cabinet:       locationPath[0],
					Chassis:       locationPath[1],
					ComputeModule: locationPath[2],
					NodeBMC:       locationPath[3],
					Node:          locationPath[4],
				}
			},
		},
	}

	for _, typeConverter := range typeConverters {
		// fmt.Println("Want: ", typeConverter.hardwareTypePath)
		// fmt.Println("Have: ", hardwareTypePath)
		// fmt.Println("Equal:", reflect.DeepEqual(typeConverter.hardwareTypePath, hardwareTypePath))
		if !reflect.DeepEqual(typeConverter.hardwareTypePath, hardwareTypePath) {
			continue
		}

		return typeConverter.convert()
	}

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

	cabinetExample(library)
	nodeBladeExample(library)
}

func cabinetExample(library *hardware_type_library.Library) {
	// Lets now say we know the cabinet, chassis, and slot that these devices are going into
	// cabinet := 1001
	// chassis := 1
	// slot := 7

	// Lets now say we know the cabinet, chassis, and slot that these devices are going into
	cabinet := 1001

	deviceTypeSlug := "hpe-ex4000"

	// TODO Interact with the inventory
	// - Check to see if the cabinet exists

	// TODO Ask the inventory for the paths to the system in the tree
	// TODO right now the system part of the path is not being considered
	locationPath := []int{}
	deviceTypePath := []hardware_type_library.HardwareType{}

	commonLogic(library, deviceTypeSlug, cabinet, deviceTypePath, locationPath)
}

func nodeBladeExample(library *hardware_type_library.Library) {
	// Lets now say we know the cabinet, chassis, and slot that these devices are going into
	// cabinet := 1001
	// chassis := 1
	// slot := 7

	// Lets now say we know the cabinet, chassis, and slot that these devices are going into
	cabinet := 1001
	chassis := 1
	slot := 7

	deviceTypeSlug := "hpe-crayex-ex420-compute-blade"

	// TODO Interact with the inventory
	// - Check to see if the cabinet exists
	// - Check to see if the chassis exists
	// - Check to see if the slot is empty
	//	- TODO also should check to see if the slot is within bounds
	// TODO Ask the inventory to see if node blade exists at cabinet: 1001, chassis: 1, slot: 7

	// TODO Ask the inventory for the paths to the chassis in the tree
	// TODO right now the system part of the path is not being considered
	locationPath := []int{cabinet, chassis}
	deviceTypePath := []hardware_type_library.HardwareType{
		hardware_type_library.HardwareTypeCabinet,
		hardware_type_library.HardwareTypeChassis,
	}

	commonLogic(library, deviceTypeSlug, slot, deviceTypePath, locationPath)
}

func commonLogic(library *hardware_type_library.Library, deviceTypeSlug string, deviceOrdinal int, deviceTypePath []hardware_type_library.HardwareType, locationPath []int) {
	// Check to see if the device type exists
	if _, err := library.GetDeviceType(deviceTypeSlug); err != nil {
		panic(err)
	}

	// Build out the hardware for a hardware type
	allChildHardware, err := library.GetDefaultHardwareBuildOut(deviceTypeSlug, deviceOrdinal)
	if err != nil {
		panic(err)
	}

	hardwareXnames := []string{}
	for _, childHardware := range allChildHardware {
		// Get full hardware type path
		childHardwareTypePath := append(deviceTypePath, childHardware.HardwareTypePath...)

		// Get full location/ordinal path
		childLocationPath := append(locationPath, childHardware.OrdinalPath...)

		xname := buildXname(childHardwareTypePath, childLocationPath)
		hardwareXnames = append(hardwareXnames, xname.String())
	}

	seperator := fmt.Sprintf("|%s|%s|%s|%s|%s|%s|%s|%s|", strings.Repeat("-", 40+1), strings.Repeat("-", 20+1), strings.Repeat("-", 60+1), strings.Repeat("-", 10+1), strings.Repeat("-", 40+1), strings.Repeat("-", 40+1), strings.Repeat("-", 45+1), strings.Repeat("-", 15+1))
	fmt.Println(seperator)
	fmt.Printf("| %-40s| %-20s| %-60s| %-10s| %-40s| %-40s| %-45s| %-15s|\n", "Hardware Type", "Manufacturer", "Model", "Ordinal", "Path", "Ordinal Path", "Hardware Type Path", "Xname")
	fmt.Println(seperator)
	for i, result := range allChildHardware {
		fmt.Printf("| %-40s| %-20s| %-60s| %-10d| %-40s| %-40s| %-45s| %-15s|\n", result.DeviceType.HardwareType, result.DeviceType.Manufacturer, result.DeviceType.Model, result.Ordinal, strings.Join(result.Path, "->"), joinInts(result.OrdinalPath, "->"), joinHardwareTypes(result.HardwareTypePath, "->"), hardwareXnames[i])
	}
	fmt.Println(seperator)

	// TODO Now at this point i think there is enough information to put data into inventory, or at least at more questions for more information
}
