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

	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/netbox-community/go-netbox/v3"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Ngsm struct {
	hardwareLibrary *hardwaretypes.Library
	Options         *NgsmOpts
	NetboxClient    *netbox.APIClient
	context         context.Context
	Boms            map[string]*queue
}

type NgsmOpts struct {
	NetboxOpts NetboxOpts `yaml:"netbox"`
}

type NetboxOpts struct {
	Scheme   string                `yaml:"scheme"`
	Host     string                `yaml:"host"`
	Token    string                `yaml:"token"`
	Insecure bool                  `yaml:"insecure"`
	Debug    bool                  `yaml:"debug"`
	cfg      *netbox.Configuration `yaml:"-"`
}

// New returns a new ngsm object, which represents the ngsm inventory
// When initializing a new sesssion with 'init,' this function is called.
// ValidateInternal is then called to ensure CANI's inventory structure is ok
// ValidateExternal is then called to ensure the provider inventory is ok
// Import is the last to be called and if successful, a session is activated
func New(cmd *cobra.Command, args []string, hardwareLibrary *hardwaretypes.Library, opts interface{}) (ngsm *Ngsm, err error) {
	ngsm = &Ngsm{
		hardwareLibrary: hardwareLibrary,
		Options: &NgsmOpts{
			NetboxOpts: NetboxOpts{
				cfg: &netbox.Configuration{},
			},
		},
		Boms: make(map[string]*queue, 0),
	}

	if cmd.Parent().Name() == "init" {
		// set the provider options
		err = ngsm.SetProviderOptions(cmd, args)
		if err != nil {
			return ngsm, err
		}
	} else {
		// use the existing options if a new session is not being initialized
		// unmarshal to a CsmOpts object
		optsMarshaled, err := yaml.Marshal(opts)
		if err != nil {
			return nil, err
		}
		NgsmOpts := NgsmOpts{}
		err = yaml.Unmarshal(optsMarshaled, &NgsmOpts)
		if err != nil {
			return ngsm, err
		}

		// set the options to the object
		ngsm.Options = &NgsmOpts
	}

	return ngsm, nil
}

func (f Ngsm) Slug() string {
	return "ngsm"
}
