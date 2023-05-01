/*
MIT License

(C) Copyright 2022 Hewlett Packard Enterprise Development LP

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/
package cani

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/omeid/uconfig"
	"github.com/omeid/uconfig/plugins/file"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:   "extract [...FILES]",
	Short: "Extract data from legacy files.",
	Long:  `Extract data from legacy files.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Extract and transform existing data
		inv, err := extractAndTransform(args)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// TODO: load the transformed data somewhere
		// loaded, _ := load(transformed)
		// for now, just pretty print it as JSON as an example
		transformed, err := json.MarshalIndent(inv, "", " ")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Print(string(transformed))
	},
}

func init() {
	rootCmd.AddCommand(extractCmd)
	extractCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		// Make an empty uconfig.Config the usage can be printed (this overrides the default cobra --help behavior)
		// This allows for users to feed the system in multiple ways:
		//   - via flags
		//   - via environment variables
		//   - via default values
		// So this just shows the user every possible way to interact with the system
		u, err := uconfig.Classic(&Extract{}, uconfig.Files{{"", yaml.Unmarshal}})

		// In order to show this we have to override the default cobra help behavior
		// and show the usage from uconfig by setting initCmd.SetHelpFunc
		u.Usage()

		if err != nil {
			os.Exit(1)
		}
	})
}

// initialize will accept a list of files and extract the data from them
// this converts the data into a new CSM Inventory, while retaining the original data in a useable format
func extractAndTransform(args []string) (Inventory, error) {
	// this will be the struct that holds the extracted data from three or more sources
	//   - csi data
	//   - canu data
	//   - sls data
	newInventory := Inventory{}

	// Since we will be extracting data from multiple files, we need to create a slice of files
	var files = file.Files{}
	// and define how they should be unmarshalled
	var um file.Unmarshal

	// For each file passed in, determine the file type and add it to the slice of files
	for _, s := range args {
		// For now, this is a simple file-extension check, but further input-validation is still required
		// Just because a file has a .yaml extension, doesn't mean it is a valid yaml file or data we want
		extension := filepath.Ext(s)
		switch extension {
		case ".yaml":
			um = yaml.Unmarshal
		case ".yml":
			um = yaml.Unmarshal
		case ".json":
			um = json.Unmarshal
		default:
			fmt.Println("Unknown file type:", s)
			os.Exit(1)
		}
		// Create the struct so it can be added to the slice
		f := struct {
			Path      string
			Unmarshal file.Unmarshal
		}{
			Path:      s,
			Unmarshal: um,
		}

		// Add the file and unmarshal directive to the slice
		files = append(files, f)

	}

	// So for now, we will just use the extracted data as the new inventory
	// Each of the three files we expect will be unmarshalled into the Extract struct
	_, err := uconfig.Classic(&newInventory.Extract.SlsConfig, files)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, err = uconfig.Classic(&newInventory.Extract.CanuConfig, files)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, err = uconfig.Classic(&newInventory.Extract.CsiConfig, files)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Now the data can be used however
	// examples:
	//   get the SLS hardware: extract.Extract.SlsConfig.Hardware
	//   get the canu architecture: extract.Extract.CanuConfig.Architecture
	//   get the csi version: extract.Extract.CsiConfig.Version

	err = newInventory.TransformSlsExtract()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = newInventory.TransformCanuExtract()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = newInventory.TransformCsiExtract()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return newInventory, nil
}
