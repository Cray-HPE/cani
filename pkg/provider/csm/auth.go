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
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

// newClient returns a new http client and context
func (csm *CSM) newClient() (httpClient *retryablehttp.Client, ctx context.Context, err error) {
	// use the system certificates if possible
	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load system cert pool: %v", err)
	}

	// optionally, use a custom cert
	if csm.Options.CaCertPath != "" {
		var cert []byte
		if csm.Options.CaCertPath != "" {
			cert, err = os.ReadFile(csm.Options.CaCertPath)
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
		InsecureSkipVerify: csm.Options.InsecureSkipVerify,
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

	if csm.Options.APIGatewayToken == "" && !csm.Options.UseSimulation {
		log.Info().Msgf("No API Gateway token provided, getting one from provider %s", csm.Options.ProviderHost)
		// Get the auth token from keycloak
		token, err := csm.getAuthToken(ctx)
		if err != nil {
			return nil, nil, err
		}
		csm.Options.APIGatewayToken = token.AccessToken
	}

	return httpClient, ctx, nil
}

// getAuthToken retrieves an auth token from the auth server using the provided credentials and certificate
func (csm *CSM) getAuthToken(ctx context.Context) (*oauth2.Token, error) {
	// Setup the oauth2 config
	conf := &oauth2.Config{
		ClientID: "shasta",
		Endpoint: oauth2.Endpoint{
			TokenURL: fmt.Sprintf("https://%s/keycloak/realms/shasta/protocol/openid-connect/token", csm.Options.ProviderHost),
		},
		Scopes: []string{"openid"},
	}

	// Get the token
	token, err := conf.PasswordCredentialsToken(ctx, csm.Options.TokenUsername, string(csm.Options.TokenPassword))
	if err != nil {
		log.Error().Msgf("Failed to get token.  Check if the token host is reachable and the credentials are correct. %s", err)
		return nil, err
	}

	return token, nil
}
