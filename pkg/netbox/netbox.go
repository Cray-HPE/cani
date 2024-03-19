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
	nbcfg.Scheme = "https"

	ctx := context.Background()
	client := netbox.NewAPIClient(nbcfg)

	return client, ctx, nil
}
