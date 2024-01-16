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

	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	hpcm_client "github.com/Cray-HPE/cani/pkg/hpcm-client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func New(cmd *cobra.Command, args []string, hardwareLibrary *hardwaretypes.Library, opts interface{}) (hpcm *Hpcm, err error) {
	hpcm = &Hpcm{
		hardwareLibrary: hardwareLibrary,
		CmConfig:        HpcmConfig{},
		client:          &hpcm_client.APIClient{},
		Options:         &HpcmOpts{},
		Cmdb:            Cmdb{},
	}

	if cmd.Parent().Name() == "init" {
		err = hpcm.SetProviderOptions(cmd, args)
		if err != nil {
			return hpcm, err
		}
	} else {
		// use the existing options if a new session is not being initialized
		optsMarshaled, err := yaml.Marshal(opts)
		if err != nil {
			return hpcm, err
		}
		hpcmOpts := HpcmOpts{}
		err = yaml.Unmarshal(optsMarshaled, &hpcmOpts)
		if err != nil {
			return hpcm, err
		}

		return hpcm, nil
	}

	return hpcm, nil
}

func (hpcm *Hpcm) Slug() string {
	return "hpcm"
}

func (hpcm *Hpcm) SetupClient(cmd *cobra.Command, args []string) error {
	// generate a new client to use with the API
	client, ctx, err := hpcm.newClient()
	if err != nil {
		return err
	}

	hpcm.Options.context = ctx

	// craft the url string
	baseUrl := fmt.Sprintf("https://%s/%s", hpcm.Options.CmdbHost, hpcm.Options.CmdbUrlBase)

	// create a hpcm client config using the client created earlier
	cfg := &hpcm_client.Configuration{
		BasePath:      baseUrl,
		Host:          hpcm.Options.CmdbHost,
		UserAgent:     taxonomy.App,
		Scheme:        "https",
		DefaultHeader: make(map[string]string),
		HTTPClient:    client.StandardClient(),
	}

	// use the token for auth
	cfg.AddDefaultHeader("X-Auth-Token", hpcm.Options.Token)

	// create the API client
	c := hpcm_client.NewAPIClient(cfg)
	hpcm.client = c
	return nil
}
