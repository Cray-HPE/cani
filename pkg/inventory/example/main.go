package main

import "github.com/Cray-HPE/cani/pkg/inventory"

func main() {
	myInventory := inventory.Inventory{}

	myInventory.Add(inventory.NewHardware(inventory.TypeSystem, "", inventory.SystemProperties{
		Name: ,
	}))
}
