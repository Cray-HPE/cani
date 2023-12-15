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
package hpcm

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/Cray-HPE/cani/cmd/taxonomy"
	hpcm_client "github.com/Cray-HPE/cani/pkg/hpcm-client"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

func (hpcm *Hpcm) ValidateExternal(cmd *cobra.Command, args []string) (err error) {
	// generate a new client to use with the API
	client, _, err := hpcm.newClient()
	if err != nil {
		return err
	}

	// craft the url string
	baseUrl := fmt.Sprintf("https://%s/%s", hpcm.Options.CmdbHost, hpcm.Options.CmdbUrlBase)

	// create a hpcm client config using the client created earlier
	cfg := &hpcm_client.Configuration{
		BasePath:      baseUrl,
		Host:          hpcm.Options.CmdbHost,
		UserAgent:     taxonomy.App,
		Scheme:        "https",
		DefaultHeader: make(map[string]string),
		HTTPClient:    client.StandardClient(),
	}

	// use the token for auth
	cfg.AddDefaultHeader("X-Auth-Token", hpcm.Options.Token)

	// create the API client
	c := hpcm_client.NewAPIClient(cfg)

	// get all the nodes
	err = hpcm.getNodes(c, cmd.Context())
	if err != nil {
		return err
	}

	// get all the management cards
	err = hpcm.getMgmtCards(c, cmd.Context())
	if err != nil {
		return err
	}

	return nil
}

// getNodes gets all the nodes from the CMDB
func (hpcm *Hpcm) getNodes(c *hpcm_client.APIClient, ctx context.Context) (err error) {
	opts := &hpcm_client.NodeOperationsApiGetAllOpts{}
	nodes, _, err := c.NodeOperationsApi.GetAll(ctx, opts)
	if err != nil {
		return err
	}

	hpcm.Nodes = nodes

	return nil
}

// getMgmtCards gets all the management cards from the CMDB
func (hpcm *Hpcm) getMgmtCards(c *hpcm_client.APIClient, ctx context.Context) (err error) {
	opts := &hpcm_client.ManagementCardOperationsApiGetAllOpts{}
	mgmtcards, _, err := c.ManagementCardOperationsApi.GetAll(ctx, opts)
	if err != nil {
		return err
	}

	hpcm.MgmtCards = mgmtcards

	return nil
}

// newClient returns a new http client and context
func (hpcm *Hpcm) newClient() (httpClient *retryablehttp.Client, ctx context.Context, err error) {
	// use the system certificates if possible
	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load system cert pool: %v", err)
	}

	// optionally, use a custom cert
	if hpcm.Options.CaCert != "" {
		var cert []byte
		if hpcm.Options.CaCert != "" {
			cert, err = os.ReadFile(hpcm.Options.CaCert)
			if err != nil {
				return nil, nil, err
			}
		}

		// append custom cert to cert pool
		if ok := certPool.AppendCertsFromPEM(cert); !ok {
			return nil, nil, fmt.Errorf("failed to append certificate %v", err)
		}
	}

	// use the certificates in the http client
	tlsConfig := &tls.Config{
		RootCAs:            certPool,
		InsecureSkipVerify: hpcm.Options.InsecureSkipVerify,
	}

	// use TLS config in transport
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	// Setup our HTTP transport and client
	httpClient = retryablehttp.NewClient()
	httpClient.HTTPClient.Transport = tr

	// Set the context with the http client
	ctx = context.Background()
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient.StandardClient())

	if hpcm.Options.Token == "" && !hpcm.Options.Simulation {
		log.Warn().Msgf("No token provided")
		// Get the auth token from keycloak
		// token, err := csm.getAuthToken(ctx)
		// if err != nil {
		// 	return nil, nil, err
		// }
		// hpcm.Options.APIGatewayToken = token.AccessToken
	}

	return httpClient, ctx, nil
}
