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
	"fmt"
	"testing"

	"github.com/Cray-HPE/cani/internal/provider/csm/validate/common"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
)

const (
	k8sPodsCidr     = "10.32.0.0/12"
	k8sServicesCidr = "10.16.0.0/12"
)

var ranges = []string{
	"0.0.0.0/0",
	"10.102.97.128/25",
	"10.102.97.0/25",
	"10.254.0.0/17",
	"10.94.100.0/24",
	"10.104.0.0/17",
	"10.107.0.0/17",
	"10.253.0.0/16",
	"10.1.1.0/16",
	"10.252.0.0/17",
	"10.92.100.0/24",
	"10.100.0.0/17",
	"10.106.0.0/17"}

func TestIpRanges(t *testing.T) {
	slsStateExtended := createSlsStateExtended(ranges...)
	checker := NewNetworkIpRangeCheck(slsStateExtended, k8sPodsCidr, k8sServicesCidr)
	results := common.NewValidationResults()
	checker.Validate(results)
	if len(results.GetResults()) != len(ranges) {
		t.Errorf("Expected %d results for good ranges: %v, results:\n%s", len(ranges), ranges, results.ToString())
	}
	passCount, warnCount, failCount := resultsCount(results.GetResults())
	if passCount != len(ranges) {
		t.Errorf("Expected %d passing results for good ranges: %v, passCount: %d, warnCount: %d, failCount: %d, results:\n%v",
			len(ranges), ranges, passCount, warnCount, failCount, results.ToString())
	}
}

func TestIpRangesOverlapping(t *testing.T) {
	slsStateExtended := createSlsStateExtended(ranges[1], ranges[2], ranges[2])
	checker := NewNetworkIpRangeCheck(slsStateExtended, k8sPodsCidr, k8sServicesCidr)
	results := common.NewValidationResults()
	checker.Validate(results)
	if len(results.GetResults()) != 3 {
		t.Errorf("Expected %d results: %v, resultCount: %d, results:\n%s",
			3, ranges, len(results.GetResults()), results.ToString())
	}
	passCount, warnCount, failCount := resultsCount(results.GetResults())
	if passCount != 1 {
		t.Errorf("Expected %d passing results: %v, passCount: %d, warnCount: %d, failCount: %d, results:\n%v",
			1, ranges, passCount, warnCount, failCount, results.ToString())

	}
	if failCount != 2 {
		t.Errorf("Expected %d failing results: %v, passCount: %d, warnCount: %d, failCount: %d, results:\n%v",
			2, ranges, passCount, warnCount, failCount, results.ToString())
	}
}

func TestIpRangesOverlappingWithK8s(t *testing.T) {
	slsStateExtended := createSlsStateExtended(k8sPodsCidr, k8sServicesCidr, ranges[0], ranges[1])
	checker := NewNetworkIpRangeCheck(slsStateExtended, k8sPodsCidr, k8sServicesCidr)
	results := common.NewValidationResults()
	checker.Validate(results)
	if len(results.GetResults()) != 6 {
		t.Errorf("Expected %d results: %v, resultCount: %d, results:\n%s",
			6, ranges, len(results.GetResults()), results.ToString())
	}
	passCount, warnCount, failCount := resultsCount(results.GetResults())
	if passCount != 4 {
		t.Errorf("Expected %d passing results: %v, passCount: %d, warnCount: %d, failCount: %d, results:\n%v",
			4, ranges, passCount, warnCount, failCount, results.ToString())

	}
	if failCount != 2 {
		t.Errorf("Expected %d failing results: %v, passCount: %d, warnCount: %d, failCount: %d, results:\n%v",
			2, ranges, passCount, warnCount, failCount, results.ToString())
	}
}

func TestIpRangesBadK8sCidrs(t *testing.T) {
	slsStateExtended := createSlsStateExtended(ranges...)
	checker := NewNetworkIpRangeCheck(slsStateExtended, "junk", "junk2")
	results := common.NewValidationResults()
	checker.Validate(results)
	if len(results.GetResults()) != len(ranges)+2 {
		t.Errorf("Expected %d results: %v, resultCount: %d, results:\n%s",
			len(ranges)+2, ranges, len(results.GetResults()), results.ToString())
	}
	passCount, warnCount, failCount := resultsCount(results.GetResults())
	if passCount != len(ranges) {
		t.Errorf("Expected %d passing results: %v, passCount: %d, warnCount: %d, failCount: %d, results:\n%v",
			len(ranges), ranges, passCount, warnCount, failCount, results.ToString())

	}
	if failCount != 2 {
		t.Errorf("Expected %d failing results: %v, passCount: %d, warnCount: %d, failCount: %d, results:\n%v",
			2, ranges, passCount, warnCount, failCount, results.ToString())
	}
}

func createSlsStateExtended(cidrs ...string) *common.SlsStateExtended {
	networks := make(map[string]sls_client.Network)
	slsState := &sls_client.SlsState{Networks: networks}
	for i, cidr := range cidrs {
		name := fmt.Sprintf("network_%d", i)
		network := sls_client.Network{Name: name}
		network.IPRanges = []string{cidr}
		slsState.Networks[name] = network
	}
	slsStateExtended := &common.SlsStateExtended{SlsState: slsState}
	return slsStateExtended
}

func resultsCount(results []common.ValidationResult) (passCount int, warnCount int, failCount int) {
	for _, r := range results {
		switch r.Result {
		case common.Pass:
			passCount++
		case common.Warning:
			warnCount++
		case common.Fail:
			failCount++
		}
	}

	return passCount, warnCount, failCount
}
