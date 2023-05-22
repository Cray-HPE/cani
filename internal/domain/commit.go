package domain

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"github.com/Cray-HPE/cani/internal/inventory"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/antihax/optional"
)

func (d *Domain) Commit(inv inventory.Inventory) error {
	// Disable TLS verification for the simulator
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr}
	c := &sls_client.Configuration{
		BasePath:   "https://localhost:8443/apis/sls/v1",
		HTTPClient: client,
		UserAgent:  "simulation",
		DefaultHeader: map[string]string{
			"Authorization": "Bearer " + os.Getenv("TOKEN"),
			"Content-Type":  "application/json",
		},
	}
	// Create a new SLS client
	sls := sls_client.NewAPIClient(c)

	// Get the existing hardware inventory from SLS
	slshw, resp, err := sls.HardwareApi.HardwareGet(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to get SLS dumpstate of hardware: %s %s", err, resp.Status)
	}

	// Reconcile the two inventories
	reconciled, err := d.externalInventoryProvider.Reconcile(inv, slshw)
	if err != nil {
		return fmt.Errorf("failed to reconcile inventory: %s", err)
	}

	// Post the reconciled inventory to SLS
	op := &sls_client.HardwareApiHardwarePostOpts{
		Body: optional.NewInterface(reconciled),
	}
	_, resp, err = sls.HardwareApi.HardwarePost(context.TODO(), op)
	if err != nil {
		return fmt.Errorf("failed to post reconciled inventory: %s %s", err, resp.Status)
	}

	return nil
}
