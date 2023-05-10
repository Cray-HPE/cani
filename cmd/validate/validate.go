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

package validate

import (
	"embed"
	"fmt"
	"io/fs"
	"net"
	"strings"

	"github.com/Cray-HPE/cani/cmd/inventory"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

var (
	//go:embed schemas/*
	schemas embed.FS
)

type ValidationResult struct {
	CheckID     ValidationCheck
	Result      Result
	ComponentID string
	Description string
}

type ValidationCheck string

const (
	IPRangeConflictCheck ValidationCheck = "ip-range-conflict"
)

type Result string

const (
	Fail    Result = "fail"
	Warning Result = "warning"
	Pass    Result = "pass"
)

func loadSchemas() (networksSchema *jsonschema.Schema, reservationsSchema *jsonschema.Schema, subnetsSchema *jsonschema.Schema, err error) {

	files, err := fs.ReadDir(schemas, "schemas")
	if err != nil {
		return
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			switch file.Name() {
			case "sls_networks_schema.json":
				if networksSchema != nil {
					continue
				}
			case "sls_reservations_schema.json":
				if reservationsSchema != nil {
					continue
				}
			case "sls_subnets_schema.json":
				if subnetsSchema != nil {
					continue
				}
			}

			filePath := fmt.Sprintf("schemas/%s", file.Name())
			content, err := schemas.ReadFile(filePath)
			if err != nil {
				fmt.Printf("Error reading embeded schema file: %s. %s\n", filePath, err)
				return networksSchema, reservationsSchema, nil, err
			}
			s, err := jsonschema.CompileString(file.Name(), string(content))
			if err != nil {
				fmt.Printf("Error compiling embeded schema. file: %s. %s\n", filePath, err)
				return networksSchema, reservationsSchema, nil, err
			}

			switch file.Name() {
			case "sls_networks_schema.json":
				networksSchema = s
			case "sls_reservations_schema.json":
				reservationsSchema = s
			case "sls_subnets_schema.json":
				subnetsSchema = s
			default:
				fmt.Printf("Unused schema: %s\n", file.Name())
			}
		}
	}
	return
}

func Validate(system *inventory.Hardware) {
	networksSchema, reservationsSchema, subnetsSchema, err := loadSchemas()
	if err != nil {
		return
	}
	fmt.Printf("networks schema: %v\n", networksSchema)
	fmt.Printf("reservations schema: %v\n", reservationsSchema)
	fmt.Printf("subnets schema: %v\n", subnetsSchema)

	// for xname, hardware := range system.Extract.SlsConfig.Hardware {
	// 	fmt.Printf("xname: %s\n%v\n\n", xname, hardware)
	// }
	// ip range to network
	results := make([]ValidationResult, 0)

	passFailResults := make(map[string]bool)
	ipRangeMap := make(map[string]*sls_client.Network)
	ipRanges := make([]string, 0)
	// for name, network := range system.Extract.SlsConfig.Networks {
	for _, network := range system.Extract.SlsConfig.Networks {
		n := network
		for _, r := range network.IPRanges {
			// fmt.Printf("%s, %s\n", name, r)
			ipRangeMap[r] = &n
			ipRanges = append(ipRanges, r)
			passFailResults[r] = true
		}
	}
	// fmt.Println("-----------------------")
	for i, ipRange1 := range ipRanges {
		// fmt.Printf("%d, %s, %s\n", i, ipRange1, ipRangeMap[ipRange1].Name)
		_, net1, _ := net.ParseCIDR(ipRange1)
		name1 := ipRangeMap[ipRange1].Name
		if !net1.IP.IsUnspecified() {
			for j := i + 1; j < len(ipRanges); j++ {
				ipRange2 := ipRanges[j]
				// todo check for error
				_, net2, _ := net.ParseCIDR(ipRange2)
				// fmt.Printf("    %d, %s\n", j, ipRange2)

				// fmt.Printf("%s, %t\n", ipRange2, net2.IP.IsUnspecified())

				// only bican can have 0.0.0.0/0
				// SystemDefaultRoute one of CAN, CHN
				// check for wrong fields
				// schema check on ExtraProperties
				if !net2.IP.IsUnspecified() && (net1.Contains(net2.IP) || net2.Contains(net1.IP)) {
					name2 := ipRangeMap[ipRange2].Name
					results = append(results,
						ValidationResult{
							CheckID:     IPRangeConflictCheck,
							Result:      Fail,
							ComponentID: name1,
							Description: fmt.Sprintf("%s %s overlaps with %s %s", name1, ipRange1, name2, ipRange2)})
					results = append(results,
						ValidationResult{
							CheckID:     IPRangeConflictCheck,
							Result:      Fail,
							ComponentID: name2,
							Description: fmt.Sprintf("%s %s overlaps with %s %s", name2, ipRange2, name1, ipRange1)})
				}
			}
		}
		// fmt.Printf("%d, %s\n", network.Name, iprange)
	}
	for ipRange, pass := range passFailResults {
		name := ipRangeMap[ipRange].Name
		if pass {
			results = append(results,
				ValidationResult{
					CheckID:     IPRangeConflictCheck,
					Result:      Pass,
					ComponentID: name,
					Description: fmt.Sprintf("IP range for %s %s does not overlap with any other network", name, ipRange)})
		}
	}

	for _, result := range results {
		fmt.Printf("%v\n", result)
	}
}
