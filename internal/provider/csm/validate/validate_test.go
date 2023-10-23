/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */

package validate

import (
	"os"
	"testing"

	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/internal/provider/csm/validate/common"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
)

func loadTestData(t *testing.T, name string) []byte {
	content, err := os.ReadFile(TestDataDir + "/" + name)
	if err != nil {
		t.Fatalf("Failed to load file %s. error: %v", name, err)
	}
	return content
}

func GetConfigOptions() ToBeValidated {
	return ToBeValidated{
		ValidRoles:      []string{"Service", "System", "Application", "Storage", "Management", "Compute"},
		ValidSubRoles:   []string{"LNETRouter", "UserDefined", "Master", "Worker", "Storage", "Gateway", "UAN", "Visualization"},
		K8sPodsCidr:     "10.32.0.0/12",
		K8sServicesCidr: "10.16.0.0/12",
	}
}

func loadTestObjects(t *testing.T, filename string) (slsState *sls_client.SlsState, rawSLSState RawJson) {
	fileContent := loadTestData(t, filename)

	raw, result, err := unmarshalToInterface(fileContent)
	if err != nil {
		t.Fatalf("Failed to unmarshal %s. error: %s", filename, err)
	}

	if result.Result != common.Pass {
		t.Fatalf("Failed to unmarshal %s. result: %v, error: %s", filename, result, err)
	}

	if raw == nil {
		t.Fatalf("Failed to unmarshal %s. the returned interface{} is nil", filename)
	}

	slsState, result, err = unmarshalToSlsState(fileContent)
	if err != nil {
		t.Fatalf("failed to unmarshal %s. error: %s", filename, err)
	}

	if result.Result != common.Pass {
		t.Fatalf("Failed to unmarshal %s. result: %v, error: %s", filename, result, err)
	}

	return slsState, raw
}

func TestUnmarshalToString(t *testing.T) {
	datafile := "invalid-mug.json"
	content := loadTestData(t, datafile)

	raw, result, err := unmarshalToInterface(content)
	if err != nil {
		t.Fatalf("Failed to unmarshal %s. error: %s", datafile, err)
	}

	if result.Result != common.Pass {
		t.Fatalf("Failed to unmarshal %s. result: %v, error: %s", datafile, result, err)
	}

	if raw == nil {
		t.Fatalf("Failed to unmarshal %s. the returned interface{} is nil", datafile)
	}
}

func TestUnmarshalToSlsState(t *testing.T) {
	datafile := "invalid-mug.json"
	content := loadTestData(t, datafile)
	slsState, result, err := unmarshalToSlsState(content)

	if err != nil {
		t.Fatalf("failed to unmarshal %s. error: %s", datafile, err)
	}

	if result.Result != common.Pass {
		t.Fatalf("Failed to unmarshal %s. result: %v, error: %s", datafile, result, err)
	}

	if slsState == nil {
		t.Fatalf("Failed to unmarshal %s. the returned slsState is nil", datafile)
	}

	if len(slsState.Hardware) == 0 {
		t.Errorf("Failed to unmarshal %s. Found zero hardware", datafile)
	}

	if len(slsState.Networks) == 0 {
		t.Errorf("Failed to unmarshal %s. Found zero networks", datafile)
	}
}

func TestValidateValid(t *testing.T) {
	datafile := "valid-mug.json"
	slsState, rawSLSState := loadTestObjects(t, datafile)
	configOptions := GetConfigOptions()
	results, err := configOptions.validate(provider.ConfigOptions{}, slsState, rawSLSState)
	passCount, warnCount, failCount := resultsCount(results)
	logResults(t, results)
	if err != nil {
		t.Logf("Validation Error: \n%s", err)
	}

	if err != nil {
		t.Errorf("Expected vaildation to pass and not return an error. pass: %d, warn: %d, fail: %d\n%s", passCount, warnCount, failCount, err)
	}

	expectedFailures := 0
	if failCount != expectedFailures {
		t.Errorf("Expected %d failures. pass: %d, warn: %d, fail: %d", expectedFailures, passCount, warnCount, failCount)
	}
}

func TestValidateInvalid(t *testing.T) {
	// Expected validation errors
	// /Hardware/x3000c0s13b0n0: x3000c0s13b0n0 Node has an invalid Role: BadRole, Valid Roles are: [Service System Application Storage Management Compute]
	// /Hardware/x3000c0s15b0n0: x3000c0s15b0n0 Node has an invalid SubRole: BadSubRole, Valid SubRoles are: [LNETRouter UserDefined Master Worker Storage Gateway UAN Visualization]
	// /Hardware/x3000c0s19b2n0: The NID 2 for x3000c0s19b2n0 is not unique. It conflicts with x3000c0s31b0n0.
	// /Hardware/x3000c0s19b1n0: The NID 1 for x3000c0s19b1n0 is not unique. It conflicts with x3000c0s30b0n0.
	// /Networks: Required network HMN_MTN is missing.
	// /Networks: Required network NMN_MTN is missing.
	datafile := "invalid-mug.json"
	slsState, rawSLSState := loadTestObjects(t, datafile)
	configOptions := GetConfigOptions()
	results, err := configOptions.validate(provider.ConfigOptions{}, slsState, rawSLSState)
	passCount, warnCount, failCount := resultsCount(results)
	logResults(t, results)
	if err != nil {
		t.Logf("Validation Error: \n%s", err)
	}

	if err == nil {
		t.Errorf("There was no error when one was expected. pass: %d, warn: %d, fail: %d", passCount, warnCount, failCount)
	}

	expectedFailures := 6
	if failCount != expectedFailures {
		t.Errorf("Expected %d failures. pass: %d, warn: %d, fail: %d", expectedFailures, passCount, warnCount, failCount)
	}
}

func TestValidateValidUsingOnlySlsState(t *testing.T) {
	datafile := "valid-mug.json"
	slsState, _ := loadTestObjects(t, datafile)
	configOptions := GetConfigOptions()
	results, err := configOptions.Validate(provider.ConfigOptions{}, slsState)
	passCount, warnCount, failCount := resultsCount(results)
	logResults(t, results)
	if err != nil {
		t.Logf("Validation Error: \n%s", err)
	}

	if err != nil {
		t.Errorf("Expected vaildation to pass and not return an error. pass: %d, warn: %d, fail: %d\n%s", passCount, warnCount, failCount, err)
	}

	expectedFailures := 0
	if failCount != expectedFailures {
		t.Errorf("Expected %d failures. pass: %d, warn: %d, fail: %d", expectedFailures, passCount, warnCount, failCount)
	}
}
