package main

import (
	"fmt"

	hardware_type_library "github.com/Cray-HPE/cani/pkg/hardware-type-library"
)

func main() {
	// Create the library
	library, err := hardware_type_library.NewEmbeddedLibrary()
	if err != nil {
		panic(err)
	}

	fmt.Println(library)
}
