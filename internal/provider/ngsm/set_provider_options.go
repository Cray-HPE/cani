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
	nb "github.com/Cray-HPE/cani/pkg/netbox"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func (ngsm *Ngsm) SetProviderOptions(cmd *cobra.Command, args []string) (err error) {
	// if os.Getenv("NETBOX_URL") == "" {
	// 	log.Fatal().Msgf("NETBOX_URL must be set when using --bom")
	// }

	// if os.Getenv("NETBOX_TOKEN") == "" {
	// 	log.Fatal().Msgf("NETBOX_TOKEN must be set when using --bom")
	// }
	// // IGNORE_SSL_ERRORS is optional

	// ngsm.Options.NetboxOpts.Host = os.Getenv("NETBOX_URL")
	// ngsm.Options.NetboxOpts.Token = os.Getenv("NETBOX_TOKEN")
	// if os.Getenv("IGNORE_SSL_ERRORS") == "True" {
	// 	ngsm.Options.NetboxOpts.Insecure = true
	// } else {
	// 	ngsm.Options.NetboxOpts.Insecure = false
	// }
	if cmd.Flags().Changed("bom") {
		// setup a new client
		ngsm.NetboxClient, ngsm.context, err = nb.NewClient()
		if err != nil {
			return err
		}

		ngsm.NetboxClient.GetConfig().Scheme = "http" // FIXME
	}
	return nil
}

// SetProviderOptions provider-specific options, usually passed in via a provider-defined init command
func (ngsm *Ngsm) SetProviderOptionsInterface(interface{}) error {
	log.Warn().Msgf("SetProviderOptionsInterface not yet implemented")
	return nil
}
