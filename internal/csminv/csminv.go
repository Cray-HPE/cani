package csminv

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	hsm_client "github.com/Cray-HPE/csminv/pkg/hsm-client"
	sls_client "github.com/Cray-HPE/hms-sls/v2/pkg/sls-client"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/spf13/viper"
)

type CSMInventory struct {
	Ctx context.Context

	SLSClient *sls_client.SLSClient
	HSMClient *hsm_client.HSMClient
}

func New(v *viper.Viper) *CSMInventory {
	ci := &CSMInventory{}

	//
	// Setup clients
	//

	ci.Ctx = SetupContext()

	// Setup HTTP client
	httpClient := retryablehttp.NewClient()
	httpClient.HTTPClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Setup SLS client
	slsURL := v.GetString("sls-url")
	ci.SLSClient = sls_client.NewSLSClient(slsURL, httpClient.StandardClient(), "")
	if v.IsSet("token") {
		ci.SLSClient.WithAPIToken(v.GetString("token"))
	}

	// Setup HSM client
	hsmURL := v.GetString("hsm-url")
	ci.HSMClient = hsm_client.NewHSMClient(hsmURL, httpClient.StandardClient(), "")
	if v.IsSet("token") {
		ci.HSMClient.WithAPIToken(v.GetString("token"))
	}

	return ci
}

func SetupContext() context.Context {
	var cancel context.CancelFunc
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-c

		// Cancel the context to cancel any in progress HTTP requests.
		cancel()
	}()

	return ctx
}
