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
package ngsm

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/netbox-community/go-netbox/v3"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// SetProviderOptions provider-specific options, usually passed in via a provider-defined init command
func (ngsm *Ngsm) SetProviderOptions(cmd *cobra.Command, args []string) (err error) {
	if cmd.Flags().Changed("bom") {
		err = ngsm.initNetboxClient(cmd, args)
		if err != nil {
			return err
		}
		// set the netbox options for the cani config file, so they can be set there going forward
		ngsm.Options.NetboxOpts.Scheme = ngsm.NetboxClient.GetConfig().Scheme
		ngsm.Options.NetboxOpts.Host = ngsm.NetboxClient.GetConfig().Host
		ngsm.Options.NetboxOpts.Token = strings.TrimPrefix(ngsm.NetboxClient.GetConfig().DefaultHeader["Authorization"], "Token ")
		ngsm.Options.NetboxOpts.Insecure = ngsm.NetboxClient.GetConfig().HTTPClient.Transport.(*retryablehttp.RoundTripper).Client.HTTPClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify
	}
	return nil
}

// initNetboxClient initializes the Netbox client using the environment variables
func (ngsm *Ngsm) initNetboxClient(cmd *cobra.Command, args []string) (err error) {
	ignoreSsl := os.Getenv("IGNORE_SSL_ERRORS")
	netboxToken := os.Getenv("NETBOX_TOKEN")
	netboxUrl := os.Getenv("NETBOX_URL")
	tlsConfig := &tls.Config{}
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	// Setup our HTTP transport and client
	httpClient := retryablehttp.NewClient()
	httpClient.HTTPClient.Transport = tr
	httpClient.Logger = nil
	c := httpClient.StandardClient()

	switch ignoreSsl {
	case "True", "true", "1", "t", "T":
		tlsConfig.InsecureSkipVerify = true
	case "False", "false", "0", "f", "F":
		tlsConfig.InsecureSkipVerify = false
	default:
		tlsConfig.InsecureSkipVerify = false
	}

	// create the netbox config
	re := regexp.MustCompile(`^(?P<protocol>\w+)://(?P<host>[^/:]+):?(?P<port>\d*)`)
	match := re.FindStringSubmatch(netboxUrl)
	url := make(map[string]string)
	// parse the submatches to the map
	for i, name := range re.SubexpNames() {
		if i > 0 && i <= len(match) {
			url[name] = match[i]
		}
	}
	// if the port is not specified, default to just the hostname
	var nbHost string
	if url["port"] != "" {
		nbHost = url["host"] + ":" + url["port"]
	} else {
		nbHost = url["host"]
	}

	// create the netbox config
	nbcfg := netbox.NewConfiguration()
	nbcfg.Host = nbHost
	nbcfg.HTTPClient = c
	nbcfg.DefaultHeader["Authorization"] = fmt.Sprintf("Token %s", netboxToken)
	if cmd.Root().Flags().Changed("debug") {
		nbcfg.Debug = true
	} else {
		nbcfg.Debug = false
	}
	nbcfg.Scheme = url["protocol"]

	// use the config to create the client
	ngsm.context = context.Background()
	ngsm.NetboxClient = netbox.NewAPIClient(nbcfg)

	return nil
}

// SetProviderOptions provider-specific options, usually passed in via a provider-defined init command
func (ngsm *Ngsm) SetProviderOptionsInterface(interface{}) error {
	log.Warn().Msgf("SetProviderOptionsInterface not yet implemented")
	return nil
}
