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

package checks

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/Cray-HPE/cani/internal/provider/csm/validate/common"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
)

const (
	TestDataDir = "../../../../../testdata/fixtures/sls-fragments"
)

func TestValidateCabinetNetwork(t *testing.T) {
	file := "valid-cabinet.json"
	data := loadTestData(t, file)
	slsState, _ := unmarshalToSlsState(t, file, data)

	checker := NewHardwareCabinetNetworkSubCheck(slsState.Networks)
	for _, h := range slsState.Hardware {
		results := common.NewValidationResults()
		props := getProps(&h)
		checker.Validate(results, &h, props)
		if isRiver(&h) {
			reportResults(t, &h, file, results, 4, 0)
		} else {
			reportResults(t, &h, file, results, 2, 0)
		}
	}
}

func TestValidateWithInvalidData(t *testing.T) {
	file := "invalid-cabinet-networks.json"
	data := loadTestData(t, file)
	slsState, _ := unmarshalToSlsState(t, file, data)

	checker := NewHardwareCabinetNetworkSubCheck(slsState.Networks)

	// Test bad CIDR
	h := slsState.Hardware["x1000"]
	if isRiver(&h) {
		t.Errorf("Failure of test assumption. Expected %s %s to be a Mountain or Hill cabinet", h.Xname, h.Class)
	}
	results := common.NewValidationResults()
	props := getProps(&h)
	checker.Validate(results, &h, props)
	reportResults(t, &h, file, results, 1, 1)
	// expected results ^^^
	// fail: /Hardware/x1000: The cabinet HMN CIDR 10.104.0.1/22 was not found in /Networks/HMN_MTN
	// pass: /Hardware/x1000: The cabinet network cn with the CIDR 10.100.0.0/22 matched the network /Networks/NMN_MTN

	// Test bad gateway and vlan
	h = slsState.Hardware["x3000"]
	if !isRiver(&h) {
		t.Errorf("Failure of test assumption. Expected %s %s to be a River cabinet", h.Xname, h.Class)
	}
	results = common.NewValidationResults()
	props = getProps(&h)
	checker.Validate(results, &h, props)
	reportResults(t, &h, file, results, 2, 2)
	// expected results ^^^
	// pass: /Hardware/x3000: The cabinet network cn with the CIDR 10.107.0.0/22 matched the network /Networks/HMN_RVR
	// pass: /Hardware/x3000: The cabinet network cn with the CIDR 10.106.0.0/22 matched the network /Networks/NMN_RVR
	// fail: /Hardware/x3000: The cabinet HMN Gateway 10.107.0.2 for CIDR 10.107.0.0/22 did not match the gateway 10.107.0.1 in /Networks/HMN_RVR
	// fail: /Hardware/x3000: The cabinet NMN vlan 177 for CIDR 10.106.0.0/22 did not match the vlan 1770 in /Networks/NMN_RVR

	// Test missing networks
	h = slsState.Hardware["x3001"]
	if !isRiver(&h) {
		t.Errorf("Failure of test assumption. Expected %s %s to be a River cabinet", h.Xname, h.Class)
	}
	results = common.NewValidationResults()
	props = getProps(&h)
	checker.Validate(results, &h, props)
	reportResults(t, &h, file, results, 1, 2)
	// expected results ^^^
	// pass: /Hardware/x3001: The cabinet network cn with the CIDR 10.107.4.0/22 matched the network /Networks/HMN_RVR
	// fail: /Hardware/x3001: x3001 Cabinet missing the NMN network
	// fail: /Hardware/x3001: x3001 Cabinet missing the ncn network
}

func reportResults(t *testing.T, h *sls_client.Hardware, testFile string, results *common.ValidationResults, expectedPass int, expectedFail int) {
	passCount, warnCount, failCount := resultsCount(results.GetResults())
	if failCount != expectedFail {
		t.Errorf("Expected %d failures using file, %s, while validating %s %s, pass: %d, warn: %d, fail: %d, results:\n%s",
			expectedFail, testFile, h.Xname, h.Class, passCount, warnCount, failCount, results.ToString())
	}
	if passCount != expectedPass {
		t.Errorf("Expected %d passing results using file, %s, while validating %s %s, pass: %d, warn: %d, fail: %d, results:\n%s",
			expectedPass, testFile, h.Xname, h.Class, passCount, warnCount, failCount, results.ToString())
	}
}

func loadTestData(t *testing.T, name string) []byte {
	content, err := os.ReadFile(TestDataDir + "/" + name)
	if err != nil {
		t.Fatalf("Failed to load file %s. error: %v", name, err)
	}
	return content
}

func unmarshalToSlsState(t *testing.T, name string, bytes []byte) (*sls_client.SlsState, error) {
	var slsState sls_client.SlsState
	err := json.Unmarshal(bytes, &slsState)
	if err != nil {
		t.Fatalf("Failed to unmarshal %s to an interface", name)
	}
	return &slsState, err
}
