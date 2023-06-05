package csm

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/oauth2"
)

// newClient returns a new http client and context
func (opts *NewOpts) newClient() (*retryablehttp.Client, context.Context, error) {
	// use the system certificates if possible
	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load system cert pool: %v", err)
	}

	// optionally, use a custom cert
	if opts.CaCertPath != "" {
		var cert []byte
		if opts.CaCertPath != "" {
			cert, err = ioutil.ReadFile(opts.CaCertPath)
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
		InsecureSkipVerify: opts.InsecureSkipVerify,
	}

	// use TLS config in transport
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	// Setup our HTTP transport and client
	httpClient := retryablehttp.NewClient()
	httpClient.HTTPClient.Transport = tr

	// Set the context with the http client
	ctx := context.Background()
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient.StandardClient())

	return httpClient, ctx, nil
}

// getAuthToken retrieves an auth token from the auth server using the provided credentials and certificate
func (opts *NewOpts) getAuthToken(ctx context.Context, client *retryablehttp.Client) (string, error) {
	// use the system certificates if possible
	certPool, err := x509.SystemCertPool()
	if err != nil {
		return "", fmt.Errorf("failed to load system cert pool: %v", err)
	}

	// optionally, use a custom cert
	if opts.CaCertPath != "" {
		var cert []byte
		if opts.CaCertPath != "" {
			cert, err = ioutil.ReadFile(opts.CaCertPath)
			if err != nil {
				return "", err
			}
		}

		// append custom cert to cert pool
		if ok := certPool.AppendCertsFromPEM(cert); !ok {
			return "", fmt.Errorf("failed to append certificate %v", err)
		}
	}

	// use the certificates in the http client
	tlsConfig := &tls.Config{
		RootCAs:            certPool,
		InsecureSkipVerify: opts.InsecureSkipVerify,
	}

	// use TLS config in transport
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	// Setup our HTTP transport and client
	httpClient := retryablehttp.NewClient()
	httpClient.HTTPClient.Transport = tr

	// Set the context with the http client
	ctx = context.Background()
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient.StandardClient())

	// Setup the oauth2 config
	conf := &oauth2.Config{
		ClientID: "shasta",
		Endpoint: oauth2.Endpoint{
			TokenURL: fmt.Sprintf("%s/keycloak/realms/shasta/protocol/openid-connect/token", opts.TokenHost),
		},
		Scopes: []string{"openid"},
	}

	// Get the token
	token, err := conf.PasswordCredentialsToken(ctx, opts.TokenUsername, string(opts.TokenPassword))
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}
