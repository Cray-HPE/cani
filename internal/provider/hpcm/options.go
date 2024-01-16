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
	"fmt"
	"os"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func (hpcm *Hpcm) SetProviderOptions(cmd *cobra.Command, args []string) error {
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
		hpcm.Options.Simulation = true
		insecure = true
		token = os.Getenv("token")
	}
	if insecure {
		hpcm.Options.InsecureSkipVerify = true
	}
	hpcm.Options.CmdbHost = host
	hpcm.Options.CmdbUrlBase = baseCmdbUrl
	hpcm.Options.CaCert = cacert
	hpcm.Options.Token = token

	err := hpcm.SetupClient(cmd, args)
	if err != nil {
		return err
	}

	return nil
}

func (hpcm *Hpcm) GetProviderOptions() (interface{}, error) {
	if hpcm.Options == nil {
		return nil, fmt.Errorf("options are nil")
	}
	return hpcm.Options, nil
}

func (hpcm *Hpcm) SetFields(hw *inventory.Hardware, values map[string]string) (result provider.SetFieldsResult, err error) {
	log.Warn().Msgf("SetFields not yet implemented")
	return provider.SetFieldsResult{}, nil
}
