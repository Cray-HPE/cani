package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// Client wraps an authenticated *http.Client for CSM API calls.
type Client struct {
	HTTP       *http.Client
	BaseURLSLS string
	BaseURLHSM string
	token      string
}

// NewClient builds a TLS-configured, optionally authenticated Client
// from the given Options.
func NewClient(opts Options) (*Client, error) {
	opts.applyDefaults()

	transport, err := buildTransport(opts)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{Transport: transport}

	token := opts.APIGatewayToken
	if token == "" && !opts.UseSimulation {
		if opts.TokenUsername == "" || opts.TokenPassword == "" {
			return nil, fmt.Errorf("keycloak credentials required when not using simulation mode")
		}
		log.Printf("Fetching auth token from %s", opts.ProviderHost)
		t, err := fetchToken(httpClient, opts.ProviderHost, opts.TokenUsername, opts.TokenPassword)
		if err != nil {
			return nil, fmt.Errorf("authenticating with %s: %w", opts.ProviderHost, err)
		}
		token = t
	}

	return &Client{
		HTTP:       httpClient,
		BaseURLSLS: opts.BaseURLSLS,
		BaseURLHSM: opts.BaseURLHSM,
		token:      token,
	}, nil
}

// Get performs an authenticated GET and returns the response body.
func (c *Client) Get(url string) ([]byte, error) {
	return c.do(http.MethodGet, url, nil)
}

// Put performs an authenticated PUT with a JSON body.
func (c *Client) Put(url string, body []byte) ([]byte, error) {
	return c.do(http.MethodPut, url, body)
}

// Delete performs an authenticated DELETE.
func (c *Client) Delete(url string) ([]byte, error) {
	return c.do(http.MethodDelete, url, nil)
}

// do executes an HTTP request with auth headers and returns the body.
func (c *Client) do(method, url string, body []byte) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, fmt.Errorf("creating %s request for %s: %w", method, url, err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", method, url, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response from %s %s: %w", method, url, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return respBody, fmt.Errorf("%s %s returned HTTP %d: %s", method, url, resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// buildTransport creates an http.Transport with TLS configured from opts.
func buildTransport(opts Options) (*http.Transport, error) {
	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("loading system cert pool: %w", err)
	}

	if opts.CaCertPath != "" {
		pem, err := os.ReadFile(opts.CaCertPath)
		if err != nil {
			return nil, fmt.Errorf("reading CA cert %s: %w", opts.CaCertPath, err)
		}
		if ok := certPool.AppendCertsFromPEM(pem); !ok {
			return nil, fmt.Errorf("failed to append CA cert from %s", opts.CaCertPath)
		}
	}

	if opts.InsecureSkipVerify {
		log.Printf("WARNING: TLS certificate verification is disabled for %s; the connection is vulnerable to man-in-the-middle attacks and credential interception", opts.ProviderHost)
	}

	return &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			RootCAs:            certPool,
			InsecureSkipVerify: opts.InsecureSkipVerify,
		},
	}, nil
}
