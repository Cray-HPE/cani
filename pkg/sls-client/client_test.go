// MIT License
//
// (C) Copyright 2022 Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package sls_client

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	sls_common "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/stretchr/testify/suite"
)

type SLSClientTestSuite struct {
	suite.Suite
}

func (suite *SLSClientTestSuite) TestGetAllHardware() {
	expectedUserAgent := "sls-client"

	allHardware := []sls_common.GenericHardware{
		sls_common.NewGenericHardware("x1000c0", sls_common.ClassMountain, nil),
		sls_common.NewGenericHardware("x1000c0b0", sls_common.ClassMountain, nil),
		sls_common.NewGenericHardware("x1000c0s0b0n0", sls_common.ClassMountain, nil),
		sls_common.NewGenericHardware("x1000c0s0b0n1", sls_common.ClassMountain, nil),
	}

	var requestCount int
	var expectedUserAgentPresent bool
	var expectedAPITokenPresent bool

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/hardware" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Update our book keeping
		requestCount++
		expectedUserAgentPresent = r.Header.Get("User-Agent") == expectedUserAgent
		expectedAPITokenPresent = reflect.DeepEqual(r.Header["Authorization"], []string{"Bearer api_token"})

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(allHardware)

	}))

	// Create the SLS Client
	slsClient := NewSLSClient(ts.URL, ts.Client(), expectedUserAgent).WithAPIToken("api_token")

	returnedHardware, err := slsClient.GetAllHardware(context.TODO())
	suite.NoError(err)
	suite.Len(returnedHardware, 4)
	suite.Equal(allHardware, returnedHardware)

	suite.Equal(1, requestCount, "Requests made to /v1/hardware")
	suite.True(expectedUserAgentPresent, "User Agent present")
	suite.True(expectedAPITokenPresent, "API Token present")
}

func (suite *SLSClientTestSuite) TestPutHardware_Existing() {
	expectedUserAgent := "sls-client"

	expectedHardwareRaw := `{"Parent":"x1000c0","Xname":"x1000c0b0","Type":"comptype_chassis_bmc","Class":"Mountain","TypeString":"ChassisBMC"}`

	var requestCount int
	var expectedUserAgentPresent bool
	var expectedRequestBodyProvided bool
	var expectedAPITokenPresent bool

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/hardware/x1000c0b0" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		defer r.Body.Close()

		// Update our book keeping
		requestCount++
		expectedUserAgentPresent = r.Header.Get("User-Agent") == expectedUserAgent
		expectedRequestBodyProvided = expectedHardwareRaw == string(body)
		expectedAPITokenPresent = reflect.DeepEqual(r.Header["Authorization"], []string{"Bearer api_token"})

		w.WriteHeader(http.StatusOK)
	}))

	// Create the SLS Client
	slsClient := NewSLSClient(ts.URL, ts.Client(), expectedUserAgent).WithAPIToken("api_token")

	hardware := sls_common.NewGenericHardware("x1000c0b0", sls_common.ClassMountain, nil)
	err := slsClient.PutHardware(context.TODO(), hardware)
	suite.NoError(err)

	suite.Equal(1, requestCount, "Requests made to /v1/hardware/x1000c0b0")
	suite.True(expectedUserAgentPresent, "User Agent present")
	suite.True(expectedRequestBodyProvided, "Expected Request body provided")
	suite.True(expectedAPITokenPresent, "API Token present")
}

func (suite *SLSClientTestSuite) TestPutHardware_New() {
	expectedUserAgent := "sls-client"

	expectedHardwareRaw := `{"Parent":"x1000c0","Xname":"x1000c0b0","Type":"comptype_chassis_bmc","Class":"Mountain","TypeString":"ChassisBMC"}`

	var requestCount int
	var expectedUserAgentPresent bool
	var expectedRequestBodyProvided bool
	var expectedAPITokenPresent bool

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/hardware/x1000c0b0" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		defer r.Body.Close()

		// Update our book keeping
		requestCount++
		expectedUserAgentPresent = r.Header.Get("User-Agent") == expectedUserAgent
		expectedRequestBodyProvided = expectedHardwareRaw == string(body)
		expectedAPITokenPresent = reflect.DeepEqual(r.Header["Authorization"], []string{"Bearer api_token"})

		w.WriteHeader(http.StatusCreated)
	}))

	// Create the SLS Client
	slsClient := NewSLSClient(ts.URL, ts.Client(), expectedUserAgent).WithAPIToken("api_token")

	hardware := sls_common.NewGenericHardware("x1000c0b0", sls_common.ClassMountain, nil)
	err := slsClient.PutHardware(context.TODO(), hardware)
	suite.NoError(err)

	suite.Equal(1, requestCount, "Requests made to /v1/hardware/x1000c0b0")
	suite.True(expectedUserAgentPresent, "User Agent present")
	suite.True(expectedRequestBodyProvided, "Expected Request body provided")
	suite.True(expectedAPITokenPresent, "API Token present")
}

func (suite *SLSClientTestSuite) TestGetDumpState() {
	expectedUserAgent := "sls-client"

	expectedSLSState := sls_common.SLSState{
		Hardware: map[string]sls_common.GenericHardware{
			"x1000c0":       sls_common.NewGenericHardware("x1000c0", sls_common.ClassMountain, nil),
			"x1000c0b0":     sls_common.NewGenericHardware("x1000c0b0", sls_common.ClassMountain, nil),
			"x1000c0s0b0n0": sls_common.NewGenericHardware("x1000c0s0b0n0", sls_common.ClassMountain, nil),
			"x1000c0s0b0n1": sls_common.NewGenericHardware("x1000c0s0b0n1", sls_common.ClassMountain, nil),
		},
		Networks: map[string]sls_common.Network{
			"HMN": sls_common.Network{
				Name: "HMN",
			},
			"NMN": sls_common.Network{
				Name: "HMN",
			},
		},
	}

	var requestCount int
	var expectedUserAgentPresent bool
	var expectedAPITokenPresent bool

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/dumpstate" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Update our book keeping
		requestCount++
		expectedUserAgentPresent = r.Header.Get("User-Agent") == expectedUserAgent
		expectedAPITokenPresent = reflect.DeepEqual(r.Header["Authorization"], []string{"Bearer api_token"})

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedSLSState)
	}))

	// Create the SLS Client
	slsClient := NewSLSClient(ts.URL, ts.Client(), expectedUserAgent).WithAPIToken("api_token")

	slsState, err := slsClient.GetDumpState(context.TODO())
	suite.NoError(err)
	suite.Len(slsState.Hardware, 4)
	suite.Len(slsState.Networks, 2)
	suite.Equal(expectedSLSState, slsState)

	suite.Equal(1, requestCount, "Requests made to /v1/dumpstate")
	suite.True(expectedUserAgentPresent, "User Agent present")
	suite.True(expectedAPITokenPresent, "API Token present")
}

func (suite *SLSClientTestSuite) PutNetwork_New() {
	expectedUserAgent := "sls-client"

	expectedNetworkRaw := `{"Name":"HMN_RVR","FullName":"River Hardware Management Network","IPRanges":["10.107.0.0/17"],"Type":"ethernet","LastUpdated":1655153192,"LastUpdatedTime":"2022-06-13 20:46:32.255491 +0000 +0000","ExtraProperties":{"CIDR":"10.107.0.0/17","MTU":9000,"Subnets":[{"CIDR":"10.107.0.0/22","DHCPEnd":"10.107.3.254","DHCPStart":"10.107.0.10","FullName":"","Gateway":"10.107.0.1","Name":"cabinet_3000","VlanID":1513}],"VlanRange":[1513]}}`

	var requestCount int
	var expectedUserAgentPresent bool
	var expectedRequestBodyProvided bool
	var expectedAPITokenPresent bool

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/networks/HMN_RVR" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		defer r.Body.Close()

		// Update our book keeping
		requestCount++
		expectedUserAgentPresent = r.Header.Get("User-Agent") == expectedUserAgent
		expectedRequestBodyProvided = expectedNetworkRaw == string(body)
		expectedAPITokenPresent = reflect.DeepEqual(r.Header["Authorization"], []string{"Bearer api_token"})

		w.WriteHeader(http.StatusCreated)
	}))

	// Create the SLS Client
	slsClient := NewSLSClient(ts.URL, ts.Client(), expectedUserAgent).WithAPIToken("api_token")

	network := sls_common.Network{
		Name:     "HMN_RVR",
		FullName: "River Hardware Management Network",
		IPRanges: []string{
			"10.107.0.0/17",
		},
		Type: sls_common.NetworkTypeEthernet,
		ExtraPropertiesRaw: sls_common.NetworkExtraProperties{
			CIDR: "10.107.0.0/17",
			MTU:  9000,
			Subnets: []sls_common.IPV4Subnet{
				sls_common.IPV4Subnet{
					CIDR:      "10.107.0.0/22",
					DHCPEnd:   net.ParseIP("10.107.3.254"),
					DHCPStart: net.ParseIP("10.107.0.10"),
					FullName:  "",
					Gateway:   net.ParseIP("10.107.0.1"),
					Name:      "cabinet_3000",
					VlanID:    1513,
				},
			},
			VlanRange: []int16{1513},
		},
	}

	err := slsClient.PutNetwork(context.TODO(), network)
	suite.NoError(err)

	suite.Equal(1, requestCount, "Requests made to /v1/networks/HMN_RVR")
	suite.True(expectedUserAgentPresent, "User Agent present")
	suite.True(expectedRequestBodyProvided, "Expected Request body provided")
	suite.True(expectedAPITokenPresent, "API Token present")
}

func (suite *SLSClientTestSuite) PutNetwork_Existing() {
	expectedUserAgent := "sls-client"

	expectedNetworkRaw := `{"Name":"HMN_RVR","FullName":"River Hardware Management Network","IPRanges":["10.107.0.0/17"],"Type":"ethernet","LastUpdated":1655153192,"LastUpdatedTime":"2022-06-13 20:46:32.255491 +0000 +0000","ExtraProperties":{"CIDR":"10.107.0.0/17","MTU":9000,"Subnets":[{"CIDR":"10.107.0.0/22","DHCPEnd":"10.107.3.254","DHCPStart":"10.107.0.10","FullName":"","Gateway":"10.107.0.1","Name":"cabinet_3000","VlanID":1513}],"VlanRange":[1513]}}`

	var requestCount int
	var expectedUserAgentPresent bool
	var expectedRequestBodyProvided bool
	var expectedAPITokenPresent bool

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/networks/HMN_RVR" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		defer r.Body.Close()

		// Update our book keeping
		requestCount++
		expectedUserAgentPresent = r.Header.Get("User-Agent") == expectedUserAgent
		expectedRequestBodyProvided = expectedNetworkRaw == string(body)
		expectedAPITokenPresent = reflect.DeepEqual(r.Header["Authorization"], []string{"Bearer api_token"})

		w.WriteHeader(http.StatusOK)
	}))

	// Create the SLS Client
	slsClient := NewSLSClient(ts.URL, ts.Client(), expectedUserAgent).WithAPIToken("api_token")

	network := sls_common.Network{
		Name:     "HMN_RVR",
		FullName: "River Hardware Management Network",
		IPRanges: []string{
			"10.107.0.0/17",
		},
		Type: sls_common.NetworkTypeEthernet,
		ExtraPropertiesRaw: sls_common.NetworkExtraProperties{
			CIDR: "10.107.0.0/17",
			MTU:  9000,
			Subnets: []sls_common.IPV4Subnet{
				sls_common.IPV4Subnet{
					CIDR:      "10.107.0.0/22",
					DHCPEnd:   net.ParseIP("10.107.3.254"),
					DHCPStart: net.ParseIP("10.107.0.10"),
					FullName:  "",
					Gateway:   net.ParseIP("10.107.0.1"),
					Name:      "cabinet_3000",
					VlanID:    1513,
				},
			},
			VlanRange: []int16{1513},
		},
	}

	err := slsClient.PutNetwork(context.TODO(), network)
	suite.NoError(err)

	suite.Equal(1, requestCount, "Requests made to /v1/networks/HMN_RVR")
	suite.True(expectedUserAgentPresent, "User Agent present")
	suite.True(expectedRequestBodyProvided, "Expected Request body provided")
	suite.True(expectedAPITokenPresent, "API Token present")
}

func TestSLSClientTestSuite(t *testing.T) {
	suite.Run(t, new(SLSClientTestSuite))
}
