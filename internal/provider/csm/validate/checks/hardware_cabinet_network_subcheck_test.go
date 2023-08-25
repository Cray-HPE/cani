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

func TestValidate(t *testing.T) {
	file := "valid-cabinet.json"
	data := loadTestData(t, file)
	slsState, _ := unmarshalToSlsState(t, file, data)

	checker := NewHardwareCabinetNetworkSubCheck(slsState.Networks)
	for _, h := range slsState.Hardware {
		results := common.NewValidationResults()
		props := getProps(&h)
		checker.Validate(results, &h, props)
		passCount, warnCount, failCount := resultsCount(results.GetResults())
		if failCount != 0 {
			t.Errorf("Expected %d failures using file, %s, while validating %s %s, pass: %d, warn: %d, fail: %d, results:\n%s",
				0, file, h.Xname, h.Class, passCount, warnCount, failCount, results.ToString())
		}
		if isRiver(&h) {
			if passCount != 4 {
				t.Errorf("Expected %d passing results using file, %s, while validating %s %s, pass: %d, warn: %d, fail: %d, results:\n%s",
					4, file, h.Xname, h.Class, passCount, warnCount, failCount, results.ToString())
			}
		} else {
			if passCount != 2 {
				t.Errorf("Expected %d passing results using file, %s, while validating %s %s, pass: %d, warn: %d, fail: %d, results:\n%s",
					2, file, h.Xname, h.Class, passCount, warnCount, failCount, results.ToString())
			}
		}
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
