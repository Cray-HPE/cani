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
package hpengi

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func (hpengi *Hpengi) SetProviderOptions(cmd *cobra.Command, args []string) error {
	useSimulation := cmd.Flags().Changed("use-simulator")
	baseCmdbUrl, _ := cmd.Flags().GetString("cmdb-url")
	insecure := cmd.Flags().Changed("insecure")
	host, _ := cmd.Flags().GetString("host")
	cacert, _ := cmd.Flags().GetString("cacert")
	token, _ := cmd.Flags().GetString("token")
	if useSimulation {
		// if a custom host is not requested, use the simulated address
		if !cmd.Flags().Changed("host") {
			host = "localhost:8888"
		}
		hpengi.Hpcm.Options.Simulation = true
		insecure = true
		token = os.Getenv("token")
	}
	if insecure {
		hpengi.Hpcm.Options.InsecureSkipVerify = true
	}
	hpengi.Hpcm.Options.CmdbHost = host
	hpengi.Hpcm.Options.CmdbUrlBase = baseCmdbUrl
	hpengi.Hpcm.Options.CaCert = cacert
	hpengi.Hpcm.Options.Token = token

	err := hpengi.Hpcm.SetupClient(cmd, args)
	if err != nil {
		return err
	}

	return nil
}

// SetProviderOptions provider-specific options, usually passed in via a provider-defined init command
func (hpengi *Hpengi) SetProviderOptionsInterface(interface{}) error {
	log.Warn().Msgf("SetProviderOptionsInterface not yet implemented")
	return nil
}
