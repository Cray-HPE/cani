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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
)

var (
	//go:embed schemas/*
	schemas embed.FS
)

type RawJson interface{}

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

func unmarshalToInterface(bytes []byte) (RawJson, ValidationResult, error) {
	var parsedJson RawJson
	if err := json.Unmarshal(bytes, &parsedJson); err != nil {
		result :=
			ValidationResult{
				CheckID:     SLSSchemaCheck,
				Result:      Fail,
				ComponentID: "SLS State",
				Description: fmt.Sprintf("SLS error unmarshaling json. %s", err)}
		return parsedJson, result, err
	}
	result :=
		ValidationResult{
			CheckID:     SLSSchemaCheck,
			Result:      Pass,
			ComponentID: "SLS State",
			Description: "SLS State is valid json."}
	return parsedJson, result, nil
}

func unmarshalToSlsState(bytes []byte) (*sls_client.SlsState, ValidationResult, error) {
	var slsState sls_client.SlsState
	if err := json.Unmarshal(bytes, &slsState); err != nil {
		result :=
			ValidationResult{
				CheckID:     SLSSchemaCheck,
				Result:      Fail,
				ComponentID: "SLS State",
				Description: fmt.Sprintf("SLS error unmarshaling json to struct. %s", err)}
		return &slsState, result, err
	}
	result :=
		ValidationResult{
			CheckID:     SLSSchemaCheck,
			Result:      Pass,
			ComponentID: "SLS State",
			Description: "SLS State is parseable struct."}

	return &slsState, result, nil
}

func allError(results []ValidationResult) error {
	var allError error
	for _, result := range results {
		if result.Result == Fail {
			e := fmt.Errorf("%s: %s", result.ComponentID, result.Description)
			allError = errors.Join(allError, e)
		}
	}
	return allError
}

// Validate validates the data in the response against the SLS schema.
func ValidateHTTPResponse(slsState *sls_client.SlsState, response *http.Response) ([]ValidationResult, error) {
	results := make([]ValidationResult, 0)

	// Parse HTTP response body to get raw JSON payload
	responseBytes, err := io.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		results = append(results,
			ValidationResult{
				CheckID:     SLSSchemaCheck,
				Result:      Fail,
				ComponentID: "SLS Networks",
				Description: fmt.Sprintf("SLS failed to get raw json dumpstate. %s", err)})
	}

	rawJson, result, err := unmarshalToInterface(responseBytes)
	results = append(results, result)
	if err != nil {
		results = append(results,
			ValidationResult{
				CheckID:     SLSSchemaCheck,
				Result:      Fail,
				ComponentID: "SLS",
				Description: fmt.Sprintf("SLS failed to parse dumpstate. %s", err)})
	}

	return validate(slsState, rawJson, results...)
}

func ValidateString(slsStateBytes []byte) ([]ValidationResult, error) {
	results := make([]ValidationResult, 0)

	rawJson, result, err := unmarshalToInterface(slsStateBytes)
	results = append(results, result)
	if err != nil {
		return results, err
	}

	slsState, result, err := unmarshalToSlsState(slsStateBytes)
	results = append(results, result)
	if err != nil {
		return results, err
	}

	r, err := validate(slsState, rawJson)
	results = append(results, r...)
	return results, err
}

func Validate(slsState *sls_client.SlsState) ([]ValidationResult, error) {
	// If we don't get a raw SLS payload, such as validating an SLS state build inside this tool we need to create the JSON version of the paylpoad
	rawSLSState, err := json.Marshal(*slsState)
	if err != nil {
		return nil, err
	}

	results := make([]ValidationResult, 0)
	rawJson, result, err := unmarshalToInterface(rawSLSState)
	results = append(results, result)
	if err != nil {
		return results, err
	}

	return validate(slsState, rawJson, results...)
}

func validate(slsState *sls_client.SlsState, rawSLSState RawJson, additionalResults ...ValidationResult) ([]ValidationResult, error) {
	results := make([]ValidationResult, 0)
	results = append(results, additionalResults...)

	// id := NewID("Network", "CAN")
	// id2 := id.append(Pair{"fish", "bones"})
	// fmt.Printf("ID: %v\n", id)
	// fmt.Printf("ID2: %v\n", id2)
	// fmt.Printf("ID str: %s\n", id.str())
	// fmt.Printf("ID str pair: %s\n", id.strPair())
	// fmt.Printf("ID2 str: %s\n", id2.str())
	// fmt.Printf("ID2 str pair: %s\n", id2.strPair())
	// fmt.Printf("ID2 str pair: %s\n", id2.strYaml())

	r := validateAgainstSchemas(rawSLSState)
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
					return results, err
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

	return results, allError(results)
}
