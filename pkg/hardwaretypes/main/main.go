// MIT License
//
// (C) Copyright [2023] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"fmt"
	"strings"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider/csm"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
)

func joinInts(ints []int, sep string) string {
	intsStr := []string{}
	for _, value := range ints {
		intsStr = append(intsStr, fmt.Sprint(value))
	}

	return strings.Join(intsStr, sep)
}

func joinHardwareTypes(in []hardwaretypes.HardwareType, sep string) string {
	out := []string{}
	for _, value := range in {
		out = append(out, string(value))
	}

	return strings.Join(out, sep)
}

func main() {
	// Create the library
	library, err := hardwaretypes.NewEmbeddedLibrary()
	if err != nil {
		panic(err)
	}

	// List cabinets
	fmt.Println()
	fmt.Println("Cabinets")
	cabinetDeviceTypes := library.GetDeviceTypesByHardwareType(hardwaretypes.HardwareTypeCabinet)
	for _, cabinetDeviceType := range cabinetDeviceTypes {
		fmt.Println(cabinetDeviceType.Slug)
	}

	fmt.Println()
	fmt.Println("Node Blade")
	nodeBladeDeviceTypes := library.GetDeviceTypesByHardwareType(hardwaretypes.HardwareTypeNodeBlade)
	for _, nodeBladeDeviceType := range nodeBladeDeviceTypes {
		fmt.Println(nodeBladeDeviceType.Slug)
	}

	cabinetExample(library)
	nodeBladeExample(library)
}

func cabinetExample(library *hardwaretypes.Library) {
	// Lets now say we know the cabinet that these devices are going into
	// cabinet := 1001

	// Lets now say we know the cabinet that these devices are going into
	cabinet := 1001

	deviceTypeSlug := "hpe-ex4000"

	// TODO Interact with the inventory
	// - Check to see if the cabinet exists

	// TODO Ask the inventory for the paths to the system in the tree
	// TODO right now the system part of the path is not being considered
	locationPath := []int{}
	deviceTypePath := []hardwaretypes.HardwareType{}

	commonLogic(library, deviceTypeSlug, cabinet, deviceTypePath, locationPath)
}

func nodeBladeExample(library *hardwaretypes.Library) {
	// Lets now say we know the cabinet, chassis, and slot that these devices are going into
	// cabinet := 1001
	// chassis := 1
	// slot := 7

	// Lets now say we know the cabinet, chassis, and slot that these devices are going into
	cabinet := 1001
	chassis := 1
	slot := 7

	deviceTypeSlug := "hpe-crayex-ex235a-compute-blade"

	// TODO Interact with the inventory
	// - Check to see if the cabinet exists
	// - Check to see if the chassis exists
	// - Check to see if the slot is empty
	//	- TODO also should check to see if the slot is within bounds
	// TODO Ask the inventory to see if node blade exists at cabinet: 1001, chassis: 1, slot: 7

	// TODO Ask the inventory for the paths to the chassis in the tree
	// TODO right now the system part of the path is not being considered
	locationPath := []int{cabinet, chassis}
	deviceTypePath := []hardwaretypes.HardwareType{
		hardwaretypes.HardwareTypeCabinet,
		hardwaretypes.HardwareTypeChassis,
	}

	commonLogic(library, deviceTypeSlug, slot, deviceTypePath, locationPath)
}

func commonLogic(library *hardwaretypes.Library, deviceTypeSlug string, deviceOrdinal int, deviceTypePath []hardwaretypes.HardwareType, locationPath []int) {
	// Check to see if the device type exists
	if _, err := library.GetDeviceType(deviceTypeSlug); err != nil {
		panic(err)
	}

	// Build out the hardware for a hardware type
	allChildHardware, err := library.GetDefaultHardwareBuildOut(deviceTypeSlug, deviceOrdinal, uuid.New())
	if err != nil {
		panic(err)
	}

	hardwareXnames := []string{}
	for _, childHardware := range allChildHardware {
		// Get full hardware type path
		childHardwareTypePath := append(deviceTypePath, childHardware.HardwareTypePath...)

		// Get full location/ordinal path
		childLocationOrdinalPath := append(locationPath, childHardware.OrdinalPath...)

		childHardwareLocation := inventory.LocationPath{}
		for i := range childHardwareTypePath {
			childHardwareLocation = append(childHardwareLocation, inventory.LocationToken{
				HardwareType: childHardwareTypePath[i],
				Ordinal:      childLocationOrdinalPath[i],
			})
		}

		xname, _ := csm.BuildXname(inventory.Hardware{}, childHardwareLocation)
		// if err != nil {
		// 	panic(err)
		// }

		if xname != nil {
			hardwareXnames = append(hardwareXnames, xname.String())
		} else {
			hardwareXnames = append(hardwareXnames, "n/a")
		}
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
