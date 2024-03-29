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

	"github.com/Cray-HPE/cani/internal/provider/csm/validate/common"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
)

const (
	requiredNetworkCheck common.ValidationCheck = "required-networks"
)

type RequiredNetworkCheck struct {
	networks map[string]sls_client.Network
}

func NewRequiedNetworkCheck(networks map[string]sls_client.Network) *RequiredNetworkCheck {
	networkCheck := RequiredNetworkCheck{
		networks: networks,
	}
	return &networkCheck
}

func (n *RequiredNetworkCheck) Validate(results *common.ValidationResults) {
	const (
		networksEndpoint      = "/Networks"
		fixMissingNetworksUrl = "https://cray-hpe.github.io/cani/latest/troubleshooting/known_errors/#required-network-hmn_mtn-is-missing-or-required-network-nmn_mtn-is-missing"
	)
	foundHmnMtnNetwork := false
	foundNmnMtnNetwork := false
	for _, network := range n.networks {
		switch network.Name {
		case "HMN_MTN":
			foundHmnMtnNetwork = true
		case "NMN_MTN":
			foundNmnMtnNetwork = true
		}
	}

	if foundHmnMtnNetwork {
		results.Pass(
			requiredNetworkCheck,
			networksEndpoint,
			"Required network HMN_MTN is present")
	} else {
		results.Fail(
			requiredNetworkCheck,
			networksEndpoint,
			fmt.Sprintf("Required network HMN_MTN is missing.  The procedure to fix this can be found here: %s", fixMissingNetworksUrl))

	}

	if foundNmnMtnNetwork {
		results.Pass(
			requiredNetworkCheck,
			networksEndpoint,
			"Required network NMN_MTN is present")
	} else {
		results.Fail(
			requiredNetworkCheck,
			networksEndpoint,
			fmt.Sprintf("Required network NMN_MTN is missing.  The procedure to fix this can be found here: %s", fixMissingNetworksUrl))
	}
}
