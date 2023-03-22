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
package csminv

import (
	"encoding/json"
	"fmt"
	"net"
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
		// Extract, transform, load
		_, err := extract(args)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// transformed, _ := transform(extracted)
		// loaded, _ := load(transformed)
		// fmt.Println(extracted)
		// Transform
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
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// extractCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// extractCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initialize will accept a list of files and extract the data from them
// this converts the data into a new CSM Inventory, while retaining the original data in a useable format
func extract(args []string) (Inventory, error) {
	// this will be the struct that holds the extracted data from three or more sources
	//   - csi data
	//   - canu data
	//   - sls data
	extract := Inventory{}

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

	// TODO: Transform the extracted data into a new CSM Inventory via way of .Transform() method
	// This method would do all the business-logic to transform the data into our new inventory
	// The extracted data is a key in the new inventory
	// This object will hold the old data "Extract" and define the new data
	// inv := Inventory{}
	// csmInventory, err := inv.Transform()

	// So for now, we will just use the extracted data as the new inventory
	// Each of the three files we expect will be unmarshalled into the Extract struct
	_, err := uconfig.Classic(&extract.Extract.SlsConfig, files)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, err = uconfig.Classic(&extract.Extract.CanuConfig, files)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, err = uconfig.Classic(&extract.Extract.CsiConfig, files)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Now the data can be used however
	// examples:
	//   get the SLS hardware: extract.Extract.SlsConfig.Hardware
	//   get the canu architecture: extract.Extract.CanuConfig.Architecture
	//   get the csi version: extract.Extract.CsiConfig.Version
	// allHw := []Hardware{}
	var names []string
	for _, s := range extract.Extract.SlsConfig.Hardware {
		// Create a new Hardware struct for each SLS Hardware
		hw := Hardware{}

		// The SLS type becomes the Type in the new Inventory
		hw.Type = s.Type.String()
		names, err = GetCommonNames(s)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		hw.Names = names

		// The SLS brand becomes the Manufacturer in the new Inventory
		brand, err := GetBrand(s)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if brand != "" {
			hw.Manufacturer = brand
		}

		// The SLS IP becomes the IP in the new Inventory
		ip, err := GetIPAddress(s)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if ip != nil {
			hw.IP = ip
		}

		// The SLS IP becomes the IP in the new Inventory
		class, err := GetClass(s)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		hw.Class = class

		// hw.Model = ""
		// The GUID comes from Redfish (or is generated)
		// hw.GUID = ""
		// fmt.Printf("FROM SLS: %s %+v\n", i, hw)
	}
	for _, canu := range extract.Extract.CanuConfig.Topology {
		// Create a new Hardware struct for each SLS Hardware
		hw := Hardware{}
		hw.Architechture = canu.Architecture
		hw.Manufacturer = canu.Vendor
		hw.Vendor = canu.Vendor
		// fmt.Printf("FROM CANU: %d %+v\n", i, hw)
	}

	// Create a new Hardware struct for each SLS Hardware
	hw := Hardware{}
	hw.CsmVersion = extract.Extract.CsiConfig.CsmVersion
	if extract.Extract.CsiConfig.SiteDNS != "" {
		hw.Networking.SiteDNS = []net.IP{net.ParseIP(extract.Extract.CsiConfig.SiteDNS)}
	}
	if extract.Extract.CsiConfig.SiteDomain != "" {
		hw.Networking.SiteDomain = extract.Extract.CsiConfig.SiteDomain
	}

	if extract.Extract.CsiConfig.CanGateway != "" {
		hw.Networking.CanGW = net.ParseIP(extract.Extract.CsiConfig.CanGateway)
	}

	if extract.Extract.CsiConfig.SiteIP != "" {
		hw.Networking.SiteIP = net.ParseIP(extract.Extract.CsiConfig.SiteIP)
	}

	// fmt.Printf("FROM CSI: %+v\n", hw)
	// let's pretty print it as JSON for example:
	configAsJson, err := json.MarshalIndent(extract, "", " ")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Print(string(configAsJson))
	return extract, nil
}
