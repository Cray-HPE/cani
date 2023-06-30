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
package csm

import (
	"fmt"

	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"

	bss_client "github.com/Cray-HPE/cani/pkg/bss-client"
	hsm_client "github.com/Cray-HPE/cani/pkg/hsm-client"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
)

type NewOpts struct {
	UseSimulation      bool
	InsecureSkipVerify bool
	APIGatewayToken    string
	BaseUrlSLS         string
	BaseUrlHSM         string
	BaseUrlBSS         string
	SecretName         string
	KubeConfig         string
	ClientID           string `json:"-" yaml:"-"` // omit credentials from cani.yml
	ClientSecret       string `json:"-" yaml:"-"` // omit credentials from cani.yml
	TokenHost          string
	TokenUsername      string `json:"-" yaml:"-"` // omit credentials from cani.yml
	TokenPassword      string `json:"-" yaml:"-"` // omit credentials from cani.yml
	CaCertPath         string
	ValidRoles         []string
	ValidSubRoles      []string
}

var DefaultValidRoles = []string{
	"Compute",
	"Service",
	"System",
	"Application",
	"Storage",
	"Management",
}
var DefaultValidSubRolesRoles = []string{
	"Worker",
	"Master",
	"Storage",
	"UAN",
	"Gateway",
	"LNETRouter",
	"Visualization",
	"UserDefined",
}

type CSM struct {
	// Clients
	slsClient *sls_client.APIClient
	hsmClient *hsm_client.APIClient
	bssClient *bss_client.BSSClient
	// System Configuration data
	ValidRoles    []string
	ValidSubRoles []string

	hardwareLibrary *hardwaretypes.Library
}

func New(opts *NewOpts, hardwareLibrary *hardwaretypes.Library) (*CSM, error) {
	csm := &CSM{
		hardwareLibrary: hardwareLibrary,
	}

	// Setup HTTP client and context using csm options
	httpClient, _, err := opts.newClient()
	if err != nil {
		return nil, err
	}

	if opts.UseSimulation {
		if opts.BaseUrlSLS == "" {
			opts.BaseUrlSLS = "https://localhost:8443/apis/sls/v1"
		}
		if opts.BaseUrlHSM == "" {
			opts.BaseUrlHSM = "https://localhost:8443/apis/smd/hsm/v2"
		}
		if opts.BaseUrlBSS == "" {
			opts.BaseUrlHSM = "https://localhost:8443/apis/bss/boot/v1"
		}
	}

	slsClientConfiguration := &sls_client.Configuration{
		BasePath:   opts.BaseUrlSLS,
		HTTPClient: httpClient.StandardClient(),
		UserAgent:  taxonomy.App,
		DefaultHeader: map[string]string{
			"Content-Type": "application/json",
		},
	}

	hsmClientConfiguration := &hsm_client.Configuration{
		BasePath:   opts.BaseUrlHSM,
		HTTPClient: httpClient.StandardClient(),
		UserAgent:  taxonomy.App,
		DefaultHeader: map[string]string{
			"Content-Type": "application/json",
		},
	}

	if opts.APIGatewayToken != "" {
		// Set the token for use in the clients
		slsClientConfiguration.DefaultHeader["Authorization"] = fmt.Sprintf("Bearer %s", opts.APIGatewayToken)
		hsmClientConfiguration.DefaultHeader["Authorization"] = fmt.Sprintf("Bearer %s", opts.APIGatewayToken)
	}

	// Set the clients
	csm.slsClient = sls_client.NewAPIClient(slsClientConfiguration)
	csm.hsmClient = hsm_client.NewAPIClient(hsmClientConfiguration)

	// Create the BSS client, since its not swagger generated need to do it differently
	csm.bssClient = bss_client.NewBSSClient(opts.BaseUrlBSS, httpClient.StandardClient(), opts.APIGatewayToken)

	// Load system specific config data
	csm.ValidRoles = opts.ValidRoles
	csm.ValidSubRoles = opts.ValidSubRoles
	return csm, nil
}
