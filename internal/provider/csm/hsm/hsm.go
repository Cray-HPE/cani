/*
MIT License

(C) Copyright 2022 Hewlett Packard Enterprise Development LP

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

package hsm

import (
	"crypto/tls"
	_ "image/png"
	"net/http"
	"os"

	hsm_client "github.com/Cray-HPE/cani/pkg/hsm-client"
)

// EnableSimulation is an exported setter function to set the value of the internal variables used in an HSM simulation mode.
func EnableSimulation() *hsm_client.APIClient {
	// Disable TLS verification for the simulator
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr}

	// Create an HSM client
	c := &hsm_client.Configuration{
		BasePath:   "https://localhost:8443/apis/smd/hsm/v2",
		HTTPClient: client,
		UserAgent:  "simulation",
		DefaultHeader: map[string]string{
			"Authorization": "Bearer " + os.Getenv("TOKEN"),
			"Content-Type":  "application/json",
		},
	}
	return hsm_client.NewAPIClient(c)
}

// DisableSimulation is an exported setter function
func DisableSimulation() *hsm_client.APIClient {
	// Enable TLS verification for production
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}
	client := &http.Client{Transport: tr}

	// Create an HSM client
	c := &hsm_client.Configuration{
		BasePath:   "https://api-gw-service-nmn/apis/smd/hsm/v2",
		HTTPClient: client,
		UserAgent:  "production",
		DefaultHeader: map[string]string{
			"Authorization": "Bearer " + os.Getenv("TOKEN"),
			"Content-Type":  "application/json",
		},
	}
	return hsm_client.NewAPIClient(c)
}
