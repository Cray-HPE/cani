/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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

	hpcm_client "github.com/Cray-HPE/cani/pkg/hpcm-client"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

// dumpCmdb GETs endpoints from the CMDB and saves the results in the *Hpcm.Cmdb
// this is typically run during ValidateExternal so known-good data is added
// TODO: add further validation checks
func (hpcm *Hpcm) dumpCmdb(cmd *cobra.Command, args []string) (err error) {
	// alertOpts := &hpcm_client.AlertOperationsApiGetAllOpts{}
	// alerts, _, err := hpcm.client.AlertOperationsApi.GetAll(hpcm.Options.context, alertOpts)
	// if err != nil {
	// 	return err
	// }

	controllerOpts := &hpcm_client.ControllerOperationsApiGetAllOpts{}
	hpcm.Cmdb.Controllers, _, err = hpcm.client.ControllerOperationsApi.GetAll(hpcm.Options.context, controllerOpts)
	if err != nil {
		return err
	}

	customGroupsOpts := &hpcm_client.CustomGroupOperationsApiGetAllOpts{}
	hpcm.Cmdb.CustomGroups, _, err = hpcm.client.CustomGroupOperationsApi.GetAll(hpcm.Options.context, customGroupsOpts)
	if err != nil {
		return err
	}

	// eventHooksOpts := &hpcm_client.EventHookOperationsApiGetAllOpts{}
	// eventhooks, _, err := hpcm.client.EventHookOperationsApi.GetAll(hpcm.Options.context, eventHooksOpts)
	// if err != nil {
	// 	return err
	// }

	// eventsOpts := &hpcm_client.EventsOperationsApiGetAllOpts{}
	// events, _, err := hpcm.client.EventsOperationsApi.GetAll(hpcm.Options.context, eventsOpts)
	// if err != nil {
	// 	return err
	// }

	imageGroupsOpts := &hpcm_client.ImageGroupOperationsApiGetAllOpts{}
	hpcm.Cmdb.ImageGroups, _, err = hpcm.client.ImageGroupOperationsApi.GetAll(hpcm.Options.context, imageGroupsOpts)
	if err != nil {
		return err
	}

	mgmtCardOpts := &hpcm_client.ManagementCardOperationsApiGetAllOpts{}
	hpcm.Cmdb.ManagementCards, _, err = hpcm.client.ManagementCardOperationsApi.GetAll(hpcm.Options.context, mgmtCardOpts)
	if err != nil {
		return err
	}

	// metricsOpts := &hpcm_client.MetricOperationsApiGetAllOpts{}
	// metrics, _, err := hpcm.client.MetricOperationsApi.GetAll(hpcm.Options.context, metricsOpts)
	// if err != nil {
	// 	return err
	// }

	networkGroupsOpts := &hpcm_client.NetworkGroupOperationsApiGetAllOpts{}
	hpcm.Cmdb.NetworkGroups, _, err = hpcm.client.NetworkGroupOperationsApi.GetAll(hpcm.Options.context, networkGroupsOpts)
	if err != nil {
		return err
	}

	networksOpts := &hpcm_client.NetworkOperationsApiGetAllOpts{}
	hpcm.Cmdb.Networks, _, err = hpcm.client.NetworkOperationsApi.GetAll(hpcm.Options.context, networksOpts)
	if err != nil {
		return err
	}

	nicsOpts := &hpcm_client.NicOperationsApiGetAllOpts{}
	hpcm.Cmdb.Nics, _, err = hpcm.client.NicOperationsApi.GetAll(hpcm.Options.context, nicsOpts)
	if err != nil {
		return err
	}

	nodeTemplateOpts := &hpcm_client.NodeTemplateOperationsApiGetAllOpts{}
	hpcm.Cmdb.NodeTemplates, _, err = hpcm.client.NodeTemplateOperationsApi.GetAll(hpcm.Options.context, nodeTemplateOpts)
	if err != nil {
		return err
	}

	nodeOpts := &hpcm_client.NodeOperationsApiGetAllOpts{}
	hpcm.Cmdb.Nodes, _, err = hpcm.client.NodeOperationsApi.GetAll(hpcm.Options.context, nodeOpts)
	if err != nil {
		return err
	}

	// sessionsOpts := &hpcm_client.SessionOperationsApiGetAllOpts{}
	// _, err = hpcm.client.SessionOperationsApi.GetAll(hpcm.Options.context, sessionsOpts)
	// if err != nil {
	// 	return err
	// }

	// cabinetsWhere := optional.NewString("name=admin")
	// cabinetsFields := optional.NewInterface("name")
	systemGroupOpts := &hpcm_client.SystemGroupOperationsApiGetAllOpts{
		// Fields: cabinetsFields,
		// Where: cabinetsWhere,
	}
	hpcm.Cmdb.SystemGroups, _, err = hpcm.client.SystemGroupOperationsApi.GetAll(hpcm.Options.context, systemGroupOpts)
	if err != nil {
		return err
	}

	// tasksOpts := &hpcm_client.TasksOperationsApiGetAllOpts{}
	// tasks, _, err := hpcm.client.TasksOperationsApi.GetAll(hpcm.Options.context, tasksOpts)
	// if err != nil {
	// 	return err
	// }

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
	}

	return httpClient, ctx, nil
}
