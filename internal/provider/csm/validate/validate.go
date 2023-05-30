/*
MIT License

(C) Copyright 2023 Hewlett Packard Enterprise Development LP

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
	"errors"
	"fmt"
	"net"
	"net/http"

	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
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
	SLSSchemaCheck       ValidationCheck = "sls-schema-validation"
)

type Result string

const (
	Fail    Result = "fail"
	Warning Result = "warning"
	Pass    Result = "pass"
)

// Validate validates the data in the response against the SLS schema.
func Validate(slsState *sls_client.SlsState, response *http.Response) error {
	results := make([]ValidationResult, 0)

	// id := NewID("Network", "CAN")
	// id2 := id.append(Pair{"fish", "bones"})
	// fmt.Printf("ID: %v\n", id)
	// fmt.Printf("ID2: %v\n", id2)
	// fmt.Printf("ID str: %s\n", id.str())
	// fmt.Printf("ID str pair: %s\n", id.strPair())
	// fmt.Printf("ID2 str: %s\n", id2.str())
	// fmt.Printf("ID2 str pair: %s\n", id2.strPair())
	// fmt.Printf("ID2 str pair: %s\n", id2.strYaml())

	r := validateAgainstSchemas(response)
	results = append(results, r...)

	passFailResults := make(map[string]bool)
	ipRangeMap := make(map[string]*sls_client.Network)
	ipRanges := make([]string, 0)
	// for name, network := range system.Extract.SlsConfig.Networks {
	for _, network := range slsState.Networks {
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
				_, net2, err := net.ParseCIDR(ipRange2)
				if err != nil {
					return err
				}
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

	var allError error
	for _, result := range results {
		if result.Result == Fail {
			e := fmt.Errorf("%s: %s", result.ComponentID, result.Description)
			allError = errors.Join(allError, e)
		}
	}

	if allError != nil {
		return allError
	}
	return nil
}
