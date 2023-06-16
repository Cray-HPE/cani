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
	"fmt"
	"io"
	"net/http"

	"github.com/Cray-HPE/cani/internal/provider/csm/validate/checks"
	"github.com/Cray-HPE/cani/internal/provider/csm/validate/common"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
)

var (
	//go:embed schemas/*
	schemas embed.FS
)

type RawJson interface{}

func unmarshalToInterface(bytes []byte) (RawJson, common.ValidationResult, error) {
	var parsedJson RawJson
	if err := json.Unmarshal(bytes, &parsedJson); err != nil {
		result :=
			common.ValidationResult{
				CheckID:     common.SLSSchemaCheck,
				Result:      common.Fail,
				ComponentID: "SLS State",
				Description: fmt.Sprintf("SLS error unmarshaling json. %s", err)}
		return parsedJson, result, err
	}
	result :=
		common.ValidationResult{
			CheckID:     common.SLSSchemaCheck,
			Result:      common.Pass,
			ComponentID: "SLS State",
			Description: "SLS State is valid json."}
	return parsedJson, result, nil
}

func unmarshalToSlsState(bytes []byte) (*sls_client.SlsState, common.ValidationResult, error) {
	var slsState sls_client.SlsState
	if err := json.Unmarshal(bytes, &slsState); err != nil {
		result :=
			common.ValidationResult{
				CheckID:     common.SLSSchemaCheck,
				Result:      common.Fail,
				ComponentID: "SLS State",
				Description: fmt.Sprintf("SLS error unmarshaling json to struct. %s", err)}
		return &slsState, result, err
	}
	result :=
		common.ValidationResult{
			CheckID:     common.SLSSchemaCheck,
			Result:      common.Pass,
			ComponentID: "SLS State",
			Description: "SLS State is parseable struct."}

	return &slsState, result, nil
}

// Validate validates the data in the response against the SLS schema.
func ValidateHTTPResponse(slsState *sls_client.SlsState, response *http.Response) ([]common.ValidationResult, error) {
	results := make([]common.ValidationResult, 0)

	// Parse HTTP response body to get raw JSON payload
	responseBytes, err := io.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		results = append(results,
			common.ValidationResult{
				CheckID:     common.SLSSchemaCheck,
				Result:      common.Fail,
				ComponentID: "SLS Networks",
				Description: fmt.Sprintf("SLS failed to get raw json dumpstate. %s", err)})
	}

	rawJson, result, err := unmarshalToInterface(responseBytes)
	results = append(results, result)
	if err != nil {
		results = append(results,
			common.ValidationResult{
				CheckID:     common.SLSSchemaCheck,
				Result:      common.Fail,
				ComponentID: "SLS",
				Description: fmt.Sprintf("SLS failed to parse dumpstate. %s", err)})
	}

	return validate(slsState, rawJson, results...)
}

func ValidateString(slsStateBytes []byte) ([]common.ValidationResult, error) {
	results := make([]common.ValidationResult, 0)

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

func Validate(slsState *sls_client.SlsState) ([]common.ValidationResult, error) {
	// If we don't get a raw SLS payload, such as validating an SLS state build inside this tool we need to create the JSON version of the payload
	rawSLSState, err := json.Marshal(*slsState)
	if err != nil {
		return nil, err
	}

	results := make([]common.ValidationResult, 0)
	rawJson, result, err := unmarshalToInterface(rawSLSState)
	results = append(results, result)
	if err != nil {
		return results, err
	}

	return validate(slsState, rawJson, results...)
}

func validate(slsState *sls_client.SlsState, rawSLSState RawJson, additionalResults ...common.ValidationResult) ([]common.ValidationResult, error) {
	results := common.NewValidationResults()
	results.Add(additionalResults...)

	r := validateAgainstSchemas(rawSLSState)
	results.Add(r...)

	slsStateExtended := common.NewSlsStateExtended(slsState)

	checkers := []common.Checker{
		checks.NewHardwareCabinetCheck(slsState.Hardware),
		checks.NewSwitchIpCheck(slsStateExtended),
		checks.NewHardwareCheck(slsState.Hardware, slsStateExtended.TypeToHardware),
		checks.NewRequiedNetworkCheck(slsState.Networks),
		checks.NewNetworkIpRangeCheck(slsStateExtended),
	}

	for _, checker := range checkers {
		checker.Validate(results)
	}

	return results.GetResults(), results.ToError()
}
