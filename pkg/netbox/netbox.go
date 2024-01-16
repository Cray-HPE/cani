/*
 *
 *  MIT License
 *
 *  (C) Copyright 2024 Hewlett Packard Enterprise Development LP
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
package netbox

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/netbox-community/go-netbox/v3"
)

// NewClient creates a new netbox client using the environment variables
// This maintains parity/compatiblity with Device-Type-Librery-Import repo
// by using the same environment variables
func NewClient() (*netbox.APIClient, context.Context, error) {
	// use the certificates in the http client
	tlsConfig := &tls.Config{}
	if os.Getenv("IGNORE_SSL_ERRORS") == "True" {
		tlsConfig.InsecureSkipVerify = true
	} else {
		tlsConfig.InsecureSkipVerify = false
	}

	// use TLS config in transport
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	// Setup our HTTP transport and client
	httpClient := retryablehttp.NewClient()
	httpClient.HTTPClient.Transport = tr
	httpClient.Logger = nil
	c := httpClient.StandardClient()

	// create the netbox config
	token := os.Getenv("NETBOX_TOKEN")
	host := stripProtocolsAndSpecialChars(os.Getenv("NETBOX_URL"))
	nbcfg := netbox.NewConfiguration()
	nbcfg.Host = host
	nbcfg.HTTPClient = c
	nbcfg.DefaultHeader["Authorization"] = fmt.Sprintf("Token %s", token)
	nbcfg.Debug = false
	nbcfg.Scheme = "http" // FIXME

	ctx := context.Background()
	client := netbox.NewAPIClient(nbcfg)

	return client, ctx, nil
}
